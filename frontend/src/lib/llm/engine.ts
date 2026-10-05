/**
 * Main-thread half of the WebLLM layer.
 *
 * Responsibilities are deliberately narrow: own the worker, expose streaming
 * tokens and cancellation, and keep exactly one engine alive for the lifetime
 * of the page. All model logic lives in webllm.worker.ts.
 *
 * Nothing is imported from @mlc-ai/web-llm here, so the multi-megabyte runtime
 * bundle is only fetched when a generation is actually requested.
 */

import {
  AIUnavailableError,
  type CapabilityReport,
  type ChatMessage,
  type LLMSettings,
} from "./types";
import type { WorkerRequest, WorkerResponse } from "./protocol";
import { getLLMSettings, hasWebGPUSupport } from "./config";

export const WEBGPU_UNAVAILABLE_MESSAGE =
  "Generative AI requires a WebGPU-compatible browser on this device. Search and campus services remain available.";

type GenerationHandlers = {
  onDelta: (delta: string) => void;
  onDone: (text: string) => void;
  onError: (message: string) => void;
  onCancelled: () => void;
};

type PendingGeneration = GenerationHandlers & {
  resolve: (text: string) => void;
  reject: (error: Error) => void;
  text: string;
};

let worker: Worker | null = null;
/** Shared engine handle; concurrent initialisation collapses onto one promise. */
let initPromise: Promise<string> | null = null;
let readyModel: string | null = null;
let initError: string | null = null;
let progressListener: ((text: string, progress: number | null) => void) | null = null;

const pending = new Map<string, PendingGeneration>();
/** Resolvers waiting on the current load, so `ready`/`init-error` can settle them. */
let initWaiters: Array<{
  resolve: (model: string) => void;
  reject: (error: AIUnavailableError) => void;
}> = [];
let requestCounter = 0;

function settleInitSuccess(model: string): void {
  const waiters = initWaiters;
  initWaiters = [];
  for (const waiter of waiters) waiter.resolve(model);
}

function settleInitFailure(message: string): void {
  const waiters = initWaiters;
  initWaiters = [];
  for (const waiter of waiters) {
    waiter.reject(new AIUnavailableError(message, "engine-failed"));
  }
}

/**
 * detectCapabilities inspects the browser without touching the network or
 * loading any weights. Safe to call on every render.
 */
export function detectCapabilities(): CapabilityReport {
  if (typeof window === "undefined") {
    return {
      hasWebGPU: false,
      hasWorkerSupport: false,
      canGenerate: false,
      reason: "The on-device model only runs in a browser.",
    };
  }

  const hasWorkerSupport = typeof Worker !== "undefined";
  const hasWebGPU = hasWebGPUSupport();

  let reason: string | null = null;
  if (!hasWebGPU) {
    reason = WEBGPU_UNAVAILABLE_MESSAGE;
  } else if (!hasWorkerSupport) {
    reason = "This browser cannot run the model in a background worker.";
  }

  return {
    hasWebGPU,
    hasWorkerSupport,
    canGenerate: hasWebGPU && hasWorkerSupport,
    reason,
  };
}

export function getLoadedModel(): string | null {
  return readyModel;
}

/** Subscribe to weight-download progress. Returns an unsubscribe function. */
export function onLoadProgress(
  listener: (text: string, progress: number | null) => void,
): () => void {
  progressListener = listener;
  return () => {
    if (progressListener === listener) progressListener = null;
  };
}

function ensureWorker(): Worker {
  if (worker) return worker;

  try {
    worker = new Worker(new URL("./webllm.worker.ts", import.meta.url), { type: "module" });
  } catch {
    // The original error can name an absolute path, so it is not surfaced.
    throw new AIUnavailableError(
      "Could not start the on-device model worker.",
      "worker-failed",
    );
  }

  worker.onmessage = (event: MessageEvent<WorkerResponse>) => {
    handleWorkerMessage(event.data);
  };

  worker.onerror = () => {
    const failure = new AIUnavailableError(
      "The on-device model worker stopped unexpectedly.",
      "worker-failed",
    );

    for (const generation of pending.values()) {
      generation.reject(failure);
      generation.onError(failure.message);
    }
    pending.clear();

    initPromise = null;
    initError = failure.message;
    readyModel = null;
    settleInitFailure(failure.message);
  };

  return worker;
}

function handleWorkerMessage(message: WorkerResponse): void {
  switch (message.type) {
    case "init-progress":
      progressListener?.(message.progress.text, message.progress.progress);
      return;

    case "ready":
      readyModel = message.model;
      initError = null;
      settleInitSuccess(message.model);
      return;

    case "init-error":
      initError = message.message;
      readyModel = null;
      settleInitFailure(message.message);
      return;

    default: {
      const generation = pending.get(message.requestId);
      if (!generation) return;

      if (message.type === "chunk") {
        generation.text += message.delta;
        generation.onDelta(message.delta);
        return;
      }

      pending.delete(message.requestId);

      if (message.type === "complete") {
        generation.onDone(message.text);
        generation.resolve(message.text);
      } else if (message.type === "cancelled") {
        generation.onCancelled();
        generation.resolve(generation.text);
      } else {
        const failure = new Error(message.message);
        generation.onError(message.message);
        generation.reject(failure);
      }
      return;
    }
  }
}

function send(message: WorkerRequest): void {
  ensureWorker().postMessage(message);
}

/**
 * initialize loads the model once and returns the loaded model id.
 *
 * Calling it repeatedly is cheap: a warm engine resolves immediately and a
 * load already in flight is awaited rather than restarted.
 */
export function initialize(settings: LLMSettings = getLLMSettings()): Promise<string> {
  if (readyModel) return Promise.resolve(readyModel);
  if (initPromise) return initPromise;
  if (initError) return Promise.reject(new AIUnavailableError(initError, "engine-failed"));

  const capabilities = detectCapabilities();
  if (!capabilities.canGenerate) {
    const failure = new AIUnavailableError(
      capabilities.reason ?? WEBGPU_UNAVAILABLE_MESSAGE,
      "no-webgpu",
    );
    initError = failure.message;
    return Promise.reject(failure);
  }

  initPromise = new Promise<string>((resolve, reject) => {
    initWaiters.push({ resolve, reject });

    try {
      send({ type: "init", model: settings.model, contextWindowSize: settings.contextWindowSize });
    } catch (error) {
      initWaiters = initWaiters.filter((waiter) => waiter.resolve !== resolve);
      reject(error);
    }
  }).finally(() => {
    // Allow a later retry to start a fresh load.
    initPromise = null;
  });

  return initPromise;
}

/**
 * generate streams one completion. Resolves with the full text, including the
 * partial output produced before a cancellation.
 */
export function generate(
  messages: ChatMessage[],
  handlers: Pick<GenerationHandlers, "onDelta" | "onDone" | "onError" | "onCancelled">,
  settings: LLMSettings = getLLMSettings(),
): { requestId: string; promise: Promise<string> } {
  requestCounter += 1;
  const requestId = `gen-${requestCounter}`;

  const promise = new Promise<string>((resolve, reject) => {
    pending.set(requestId, {
      onDelta: handlers.onDelta,
      onDone: handlers.onDone,
      onError: handlers.onError,
      onCancelled: handlers.onCancelled,
      resolve,
      reject,
      text: "",
    });
  });

  send({
    type: "generate",
    requestId,
    messages,
    options: {
      temperature: settings.temperature,
      topP: settings.topP,
      maxTokens: settings.maxTokens,
    },
  });

  return { requestId, promise };
}

/** Cancel one in-flight generation. */
export function interrupt(requestId: string): void {
  if (!pending.has(requestId)) return;
  send({ type: "interrupt", requestId });
}

/** Tear the worker down and release GPU memory. */
export async function dispose(): Promise<void> {
  const active = worker;
  worker = null;
  initPromise = null;
  readyModel = null;
  initError = null;
  progressListener = null;
  settleInitFailure("The on-device model was disposed before it finished loading.");

  for (const generation of pending.values()) {
    generation.onCancelled();
  }
  pending.clear();

  if (!active) return;

  active.postMessage({ type: "dispose" } satisfies WorkerRequest);
  active.terminate();
}