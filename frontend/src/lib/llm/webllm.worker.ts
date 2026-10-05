/// <reference lib="webworker" />

/**
 * WebLLM runs here, off the main thread.
 *
 * Model download and token generation are both long-running and GPU-bound.
 * Doing them on the main thread would freeze scrolling and input, so all
 * inference is confined to this worker and the UI only ever receives deltas.
 *
 * The engine is created once and reused for every later request; nothing here
 * runs until the main thread sends an explicit `init`.
 */

import { CreateMLCEngine, prebuiltAppConfig, type MLCEngine } from "@mlc-ai/web-llm";
import type { WorkerRequest, WorkerResponse } from "./protocol";
import type { ChatMessage, CompletionOptions } from "./types";

let engine: MLCEngine | null = null;
/** In-flight generations, so an interrupt can cancel the right one. */
const active = new Map<string, AbortController>();

function post(message: WorkerResponse): void {
  self.postMessage(message);
}

function describe(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "The on-device model failed.";
}

async function init(model: string, contextWindowSize: number): Promise<void> {
  // Reuse a warm engine rather than paying the load cost twice.
  if (engine) {
    post({ type: "ready", model });
    return;
  }

  if (typeof navigator !== "undefined" && !("gpu" in navigator)) {
    post({
      type: "init-error",
      message:
        "Generative AI requires a WebGPU-compatible browser on this device. Search and campus services remain available.",
    });
    return;
  }

  const known = prebuiltAppConfig.model_list.some((record) => record.model_id === model);
  if (!known) {
    post({
      type: "init-error",
      message: `"${model}" is not a WebLLM prebuilt model. Set NEXT_PUBLIC_WEBLLM_MODEL to an id from prebuiltAppConfig.model_list.`,
    });
    return;
  }

  try {
    engine = await CreateMLCEngine(
      model,
      {
        initProgressCallback: (report) => {
          post({
            type: "init-progress",
            progress: {
              progress: typeof report.progress === "number" ? report.progress : null,
              text: report.text ?? "Loading model…",
            },
          });
        },
      },
      { context_window_size: contextWindowSize },
    );

    post({ type: "ready", model });
  } catch (error) {
    engine = null;
    post({ type: "init-error", message: describe(error) });
  }
}

async function handleGenerate(
  requestId: string,
  messages: ChatMessage[],
  options: CompletionOptions | undefined,
): Promise<void> {
  const activeEngine = engine;
  if (!activeEngine) {
    post({ type: "error", requestId, message: "The on-device model is not loaded yet." });
    return;
  }

  const controller = new AbortController();
  active.set(requestId, controller);

  let text = "";
  try {
    const stream = await activeEngine.chat.completions.create({
      messages,
      stream: true,
      temperature: options?.temperature ?? 0.7,
      top_p: options?.topP ?? 0.95,
      max_tokens: options?.maxTokens ?? 512,
    });

    for await (const chunk of stream) {
      if (controller.signal.aborted) break;

      const delta = chunk.choices[0]?.delta?.content;
      if (delta) {
        text += delta;
        post({ type: "chunk", requestId, delta });
      }
    }

    post(
      controller.signal.aborted
        ? { type: "cancelled", requestId }
        : { type: "complete", requestId, text },
    );
  } catch (error) {
    post(
      controller.signal.aborted
        ? { type: "cancelled", requestId }
        : { type: "error", requestId, message: describe(error) },
    );
  } finally {
    active.delete(requestId);
  }
}

async function dispose(): Promise<void> {
  for (const controller of active.values()) {
    controller.abort();
  }
  active.clear();

  const activeEngine = engine;
  engine = null;
  if (!activeEngine) return;

  try {
    await activeEngine.interruptGenerate();
  } catch {
    // An interrupted generation is not a disposal failure.
  }
  await activeEngine.unload();
}

self.onmessage = (event: MessageEvent<WorkerRequest>) => {
  const message = event.data;

  switch (message.type) {
    case "init":
      void init(message.model, message.contextWindowSize);
      return;
    case "generate":
      void handleGenerate(message.requestId, message.messages, message.options);
      return;
    case "interrupt":
      active.get(message.requestId)?.abort();
      return;
    case "dispose":
      void dispose();
      return;
  }
};

export {};