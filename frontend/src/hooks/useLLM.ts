"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  disposeProvider,
  getProvider,
  type ChatMessage,
  type CompletionOptions,
  type LLMProvider,
  type LLMStatus,
  type LoadProgress,
} from "@/lib/llm";

export type UseLLMResult = {
  status: LLMStatus;
  providerName: string | null;
  model: string | null;
  progress: LoadProgress | null;
  error: string | null;
  isStreaming: boolean;
  /** Load (or reuse) the runtime. Resolves to true when ready. */
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
  return "The language model failed to respond.";
}

export function useLLM(): UseLLMResult {
  const [status, setStatus] = useState<LLMStatus>("idle");
  const [providerName, setProviderName] = useState<string | null>(null);
  const [model, setModel] = useState<string | null>(null);
  const [progress, setProgress] = useState<LoadProgress | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);

  const providerRef = useRef<LLMProvider | null>(null);
  const abortRef = useRef(false);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const initialize = useCallback(async (): Promise<boolean> => {
    if (providerRef.current?.isReady()) return true;

    setStatus("loading");
    setError(null);
    setProgress({ progress: null, text: "Starting the language model runtime…" });

    try {
      const provider = await getProvider();
      providerRef.current = provider;
      setProviderName(provider.name);
      setModel(provider.model);

      await provider.load((update) => {
        if (mountedRef.current) setProgress(update);
      });

      if (!mountedRef.current) return false;

      setProgress(null);
      setStatus("ready");
      return true;
    } catch (cause) {
      if (!mountedRef.current) return false;
      providerRef.current = null;
      setStatus("error");
      setProgress(null);
      setError(describe(cause));
      return false;
    }
  }, []);

  const generate = useCallback(
    async (
      messages: ChatMessage[],
      onDelta: (delta: string) => void,
      options?: CompletionOptions,
    ): Promise<string> => {
      const provider = providerRef.current;
      if (!provider?.isReady()) {
        throw new Error("The language model is not ready yet.");
      }

      abortRef.current = false;
      setIsStreaming(true);
      setError(null);

      let full = "";
      try {
        for await (const delta of provider.stream(messages, options)) {
          if (abortRef.current) break;
          full += delta;
          onDelta(delta);
        }
      } catch (cause) {
        if (mountedRef.current) setError(describe(cause));
        throw cause;
      } finally {
        if (mountedRef.current) setIsStreaming(false);
      }

      return full;
    },
    [],
  );

  const stop = useCallback(() => {
    abortRef.current = true;
    setIsStreaming(false);
  }, []);

  useEffect(() => {
    return () => {
      abortRef.current = true;
      void disposeProvider();
    };
  }, []);

  return {
    status,
    providerName,
    model,
    progress,
    error,
    isStreaming,
    initialize,
    generate,
    stop,
  };
}