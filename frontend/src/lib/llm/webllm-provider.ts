import {
  LLMError,
  type ChatMessage,
  type CompletionOptions,
  type LoadProgress,
  type LLMProvider,
  type LLMProviderName,
} from "./types";

type WebllmModule = typeof import("@mlc-ai/web-llm");
type MLCEngine = import("@mlc-ai/web-llm").MLCEngine;

/**
 * WebLLMProvider runs the model entirely in the browser on WebGPU.
 *
 * The WebLLM bundle is several megabytes and is browser-only, so it is imported
 * lazily inside load() instead of at module scope.
 */
export class WebLLMProvider implements LLMProvider {
  readonly name: LLMProviderName = "webllm";
  readonly model: string;

  private engine: MLCEngine | null = null;
  private loading: Promise<void> | null = null;

  constructor(model: string) {
    this.model = model;
  }

  isReady(): boolean {
    return this.engine !== null;
  }

  async load(onProgress: (progress: LoadProgress) => void): Promise<void> {
    if (this.engine) return;
    if (this.loading) return this.loading;

    this.loading = this.createEngine(onProgress).finally(() => {
      this.loading = null;
    });

    return this.loading;
  }

  private async createEngine(onProgress: (progress: LoadProgress) => void): Promise<void> {
    if (typeof window === "undefined") {
      throw new LLMError("WebLLM can only run in the browser.");
    }
    if (!("gpu" in navigator)) {
      throw new LLMError(
        "WebGPU is not available in this browser. Enable hardware acceleration, or point NEXT_PUBLIC_LLM_PROVIDER at a local Ollama server.",
      );
    }

    let webllm: WebllmModule;
    try {
      webllm = await import("@mlc-ai/web-llm");
    } catch (error) {
      throw new LLMError("Failed to load the WebLLM runtime.", error);
    }

    const known = webllm.prebuiltAppConfig.model_list.some(
      (record) => record.model_id === this.model,
    );
    if (!known) {
      throw new LLMError(
        `"${this.model}" is not a WebLLM prebuilt model. Pick an id from webllm.prebuiltAppConfig.model_list.`,
      );
    }

    try {
      this.engine = await webllm.CreateMLCEngine(this.model, {
        initProgressCallback: (report) => {
          onProgress({
            progress: typeof report.progress === "number" ? report.progress : null,
            text: report.text ?? "Loading model…",
          });
        },
      });
    } catch (error) {
      this.engine = null;
      throw new LLMError(`Could not initialise WebLLM model "${this.model}".`, error);
    }
  }

  async *stream(
    messages: ChatMessage[],
    options: CompletionOptions = {},
  ): AsyncGenerator<string, void, unknown> {
    const engine = this.engine;
    if (!engine) {
      throw new LLMError("WebLLM model is not loaded yet.");
    }

    const request = {
      messages,
      stream: true as const,
      temperature: options.temperature ?? 0.7,
      top_p: options.topP ?? 0.95,
      max_tokens: options.maxTokens ?? 1024,
    };

    const stream = await engine.chat.completions.create(request);
    for await (const chunk of stream) {
      const delta = chunk.choices[0]?.delta?.content;
      if (delta) yield delta;
    }
  }

  async dispose(): Promise<void> {
    const engine = this.engine;
    this.engine = null;
    if (!engine) return;

    try {
      await engine.interruptGenerate();
    } catch {
      // An interrupted generation is not a disposal failure.
    }
    await engine.unload();
  }
}