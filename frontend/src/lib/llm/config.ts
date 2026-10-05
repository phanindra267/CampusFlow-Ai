/**
 * Configuration for the on-device language model.
 *
 * WebLLM on WebGPU is the only runtime, so there is no provider resolution and
 * no fallback chain. Everything tunable is data:
 *
 *   NEXT_PUBLIC_WEBLLM_MODEL          prebuilt model id (default below)
 *   NEXT_PUBLIC_WEBLLM_CONTEXT_WINDOW context window size in tokens
 *   NEXT_PUBLIC_WEBLLM_TEMPERATURE    sampling temperature
 *   NEXT_PUBLIC_WEBLLM_TOP_P          nucleus sampling cutoff
 *   NEXT_PUBLIC_WEBLLM_MAX_TOKENS     default cap on generated tokens
 *
 * The default is a small quantised Qwen model. Optimising for low latency and
 * low memory matters more here than parameter count: the model has to download
 * once and run inside a browser tab.
 */

/**
 * Small, quantised, Qwen-family. Roughly a quarter the memory of the 8B
 * default that this project used previously, which is the difference between
 * "loads on a laptop" and "loads on a phone" for browser inference.
 */
export const DEFAULT_WEBLLM_MODEL = "Qwen2.5-0.5B-Instruct-q4f16_1-MLC";

/** Fallback context window when the model id is unknown or unparseable. */
const DEFAULT_CONTEXT_WINDOW = 4096;

export type LLMSettings = {
  model: string;
  contextWindowSize: number;
  temperature: number;
  topP: number;
  maxTokens: number;
};

function readString(key: string): string | undefined {
  const value = process.env[key]?.trim();
  return value ? value : undefined;
}

/**
 * readNumber clamps to a sane range so a typo in .env cannot make the model
 * unusably slow or nonsensical.
 */
function readNumber(key: string, fallback: number, min: number, max: number): number {
  const raw = readString(key);
  if (raw === undefined) return fallback;

  const parsed = Number(raw);
  if (!Number.isFinite(parsed)) return fallback;

  return Math.min(max, Math.max(min, parsed));
}

export function getLLMSettings(): LLMSettings {
  const contextWindowSize = readNumber(
    "NEXT_PUBLIC_WEBLLM_CONTEXT_WINDOW",
    DEFAULT_CONTEXT_WINDOW,
    512,
    32768,
  );

  return {
    model: readString("NEXT_PUBLIC_WEBLLM_MODEL") ?? DEFAULT_WEBLLM_MODEL,
    contextWindowSize,
    temperature: readNumber("NEXT_PUBLIC_WEBLLM_TEMPERATURE", 0.7, 0, 2),
    topP: readNumber("NEXT_PUBLIC_WEBLLM_TOP_P", 0.95, 0, 1),
    // Never let generation exceed the window, leaving room for the prompt.
    maxTokens: Math.min(
      readNumber("NEXT_PUBLIC_WEBLLM_MAX_TOKENS", 512, 32, 8192),
      contextWindowSize,
    ),
  };
}

/** True when the browser exposes a WebGPU adapter factory. */
export function hasWebGPUSupport(): boolean {
  return typeof navigator !== "undefined" && "gpu" in navigator;
}