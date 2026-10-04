/**
 * Runtime configuration for the LLM layer.
 *
 * Everything is read from NEXT_PUBLIC_* variables so the provider can be
 * switched without touching application code:
 *
 *   NEXT_PUBLIC_LLM_PROVIDER   auto | webllm | ollama   (default: auto)
 *   NEXT_PUBLIC_LLM_MODEL      model identifier passed to the provider
 *   NEXT_PUBLIC_OLLAMA_URL     base URL of the Ollama server
 *
 * `auto` prefers a reachable Ollama server (useful during local development,
 * where a warm qwen3:8b is already running) and otherwise falls back to the
 * in-browser WebLLM runtime.
 */

export const DEFAULT_WEBLLM_MODEL = "Qwen3-8B-q4f16_1-MLC";
export const DEFAULT_OLLAMA_MODEL = "qwen3:8b";
export const DEFAULT_OLLAMA_URL = "http://localhost:11434";

export type ProviderPreference = "auto" | "webllm" | "ollama";

export type LLMSettings = {
  preference: ProviderPreference;
  webllmModel: string;
  ollamaModel: string;
  ollamaUrl: string;
};

function readPreference(raw: string | undefined): ProviderPreference {
  const value = (raw ?? "auto").trim().toLowerCase();
  if (value === "webllm" || value === "ollama" || value === "auto") return value;
  return "auto";
}

function trimTrailingSlash(url: string): string {
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

export function getLLMSettings(): LLMSettings {
  const ollamaUrl =
    process.env.NEXT_PUBLIC_OLLAMA_URL?.trim() || DEFAULT_OLLAMA_URL;

  return {
    preference: readPreference(process.env.NEXT_PUBLIC_LLM_PROVIDER),
    webllmModel: process.env.NEXT_PUBLIC_LLM_MODEL?.trim() || DEFAULT_WEBLLM_MODEL,
    ollamaModel: process.env.NEXT_PUBLIC_OLLAMA_MODEL?.trim() || DEFAULT_OLLAMA_MODEL,
    ollamaUrl: trimTrailingSlash(ollamaUrl),
  };
}

/** Probe an Ollama server; used by the `auto` provider selection. */
export async function isOllamaReachable(baseUrl: string, signal?: AbortSignal): Promise<boolean> {
  try {
    const response = await fetch(`${baseUrl}/api/tags`, { signal });
    return response.ok;
  } catch {
    return false;
  }
}

export function hasWebGPUSupport(): boolean {
  return typeof navigator !== "undefined" && "gpu" in navigator;
}