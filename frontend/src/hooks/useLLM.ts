"use client";

import { useCallback, useEffect, useMemo, useRef, useState, useSyncExternalStore } from "react";
import {
  dispose,
  detectCapabilities,
  generate,
  getLLMSettings,
  getLoadedModel,
  initialize,
  interrupt,
  onLoadProgress,
  type CapabilityReport,
  type ChatMessage,
  type CompletionOptions,
  type LLMStatus,
  type LoadProgress,
} from "@/lib/llm";

export type UseLLMResult = {
  status: LLMStatus;
  /** The loaded model id, or null while nothing is loaded. */
  model: string | null;
  progress: LoadProgress | null;
  error: string | null;
  isStreaming: boolean;
  /** What this browser can do, computed without loading any weights. */
  capabilities: CapabilityReport;
  /** Load (or reuse) the model. Resolves to true when ready. */
  initialize: () => Promise<boolean>;
  /** Stream a completion, invoking onDelta for each token. */
  generate: (
    messages: ChatMessage[],
    onDelta: (delta: string) => void,
    options?: CompletionOptions,
  ) => Promise<string>;
  /** Abort an in-flight generation. */
  stop: () => void;
};

function describe(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "The on-device model failed to respond.";
}

/**
 * Capabilities are a property of the environment, not of React state. Reading
 * them through an external store keeps them out of an effect — no cascading
 * render, and no hydration mismatch, because the server snapshot is always the
 * "checking" placeholder.
 */
const CHECKING_CAPABILITIES: CapabilityReport = {
  hasWebGPU: false,
  hasWorkerSupport: false,
  canGenerate: false,
  reason: "Checking browser capabilities…",
};

let capabilitySnapshot: CapabilityReport | null = null;

function readCapabilities(): CapabilityReport {
  capabilitySnapshot ??= detectCapabilities();
  return capabilitySnapshot;
}

function subscribeToCapabilities(onChange: () => void): () => void {
  window.addEventListener("webgpuadapterchange", onChange);
  return () => window.removeEventListener("webgpuadapterchange", onChange);
}

function serverCapabilities(): CapabilityReport {
  return CHECKING_CAPABILITIES;
}

export function useLLM(): UseLLMResult {
  const capabilities = useSyncExternalStore(
    subscribeToCapabilities,
    readCapabilities,
    serverCapabilities,
  );

  const [status, setStatus] = useState<LLMStatus>("idle");
  const [progress, setProgress] = useState<LoadProgress | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);

  const requestIdRef = useRef<string | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const initializeEngine = useCallback(async (): Promise<boolean> => {
    // The engine is shared process-wide; a warm load resolves immediately.
    const loaded = getLoadedModel();
    if (loaded) {
      setStatus("ready");
      return true;
    }

    setStatus("loading");
    setError(null);
    setProgress({ progress: null, text: "Starting the on-device model runtime…" });

    const unsubscribe = onLoadProgress((text, value) => {
      if (mountedRef.current) setProgress({ text, progress: value });
    });

    try {
      const settings = getLLMSettings();
      const loadedModel = await initialize(settings);

      if (!mountedRef.current) return false;

      unsubscribe();
      setProgress(null);
      setStatus("ready");
      return loadedModel !== null;
    } catch (cause) {
      if (!mountedRef.current) return false;

      unsubscribe();
      setStatus("error");
      setProgress(null);
      setError(describe(cause));
      return false;
    }
  }, []);

  const runGenerate = useCallback(
    async (
      messages: ChatMessage[],
      onDelta: (delta: string) => void,
      options?: CompletionOptions,
    ): Promise<string> => {
      if (!getLoadedModel()) {
        throw new Error("The on-device model is not loaded yet.");
      }

      setIsStreaming(true);
      setError(null);

      const { requestId, promise } = generate(
        messages,
        {
          onDelta,
          onDone: () => undefined,
          onCancelled: () => undefined,
          onError: (message) => {
            if (mountedRef.current) setError(message);
          },
        },
        { ...getLLMSettings(), ...options },
      );

      requestIdRef.current = requestId;

      try {
        return await promise;
      } finally {
        requestIdRef.current = null;
        if (mountedRef.current) setIsStreaming(false);
      }
    },
    [],
  );

  const stop = useCallback(() => {
    if (requestIdRef.current) {
      interrupt(requestIdRef.current);
    }
    setIsStreaming(false);
  }, []);

  useEffect(() => {
    return () => {
      void dispose();
    };
  }, []);

  return useMemo(
    () => ({
      status,
      model: getLoadedModel(),
      progress,
      error,
      isStreaming,
      capabilities,
      initialize: initializeEngine,
      generate: runGenerate,
      stop,
    }),
    [capabilities, error, initializeEngine, isStreaming, progress, runGenerate, status, stop],
  );
}