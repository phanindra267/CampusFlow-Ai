import { getLLMSettings, isOllamaReachable, type LLMSettings } from "./config";
import { OllamaProvider } from "./ollama-provider";
import type { LLMProvider } from "./types";
import { WebLLMProvider } from "./webllm-provider";

export * from "./types";
export { getLLMSettings, hasWebGPUSupport, DEFAULT_OLLAMA_MODEL, DEFAULT_WEBLLM_MODEL } from "./config";

let activeProvider: LLMProvider | null = null;

/**
 * Resolve the provider to use for this session.
 *
 * An explicit `webllm` / `ollama` preference is honoured as configured. With
 * `auto`, a reachable Ollama server wins (the usual local-development setup)
 * and the in-browser WebLLM runtime is used otherwise.
 */
export async function resolveProvider(settings: LLMSettings = getLLMSettings()): Promise<LLMProvider> {
  if (settings.preference === "webllm") {
    return new WebLLMProvider(settings.webllmModel);
  }

  if (settings.preference === "ollama") {
    return new OllamaProvider(settings.ollamaUrl, settings.ollamaModel);
  }

  if (typeof window !== "undefined" && (await isOllamaReachable(settings.ollamaUrl))) {
    return new OllamaProvider(settings.ollamaUrl, settings.ollamaModel);
  }

  return new WebLLMProvider(settings.webllmModel);
}

/**
 * getProvider returns a process-wide provider instance, creating it on first
 * use. Model weights are expensive to initialise, so the runtime is shared by
 * every consumer on the page.
 */
export async function getProvider(): Promise<LLMProvider> {
  if (activeProvider) return activeProvider;
  activeProvider = await resolveProvider();
  return activeProvider;
}

export async function disposeProvider(): Promise<void> {
  const provider = activeProvider;
  activeProvider = null;
  if (!provider) return;

  try {
    await provider.dispose();
  } catch (error) {
    console.warn("Failed to dispose the LLM provider cleanly:", error);
  }
}

export { OllamaProvider } from "./ollama-provider";
export { WebLLMProvider } from "./webllm-provider";
export { LLMError } from "./types";