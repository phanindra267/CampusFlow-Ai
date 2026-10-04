export type LLMProviderName = "webllm" | "ollama";

export type ChatRole = "system" | "user" | "assistant";

export type ChatMessage = {
  role: ChatRole;
  content: string;
};

export type LoadProgress = {
  /** 0..1 overall progress, or null when the provider cannot estimate it. */
  progress: number | null;
  text: string;
};

export type CompletionOptions = {
  temperature?: number;
  maxTokens?: number;
  topP?: number;
};

export type LLMStatus = "idle" | "loading" | "ready" | "error";

export type Citation = {
  id: string;
  source: string;
};

/**
 * LLMProvider is the contract every backend must satisfy. Both the in-browser
 * WebLLM runtime and a local Ollama server implement it, so the rest of the
 * app never needs to know which one is active.
 */
export interface LLMProvider {
  readonly name: LLMProviderName;
  readonly model: string;

  /** Load weights and warm the runtime. Safe to call repeatedly. */
  load(onProgress: (progress: LoadProgress) => void): Promise<void>;

  /** True once the provider can serve completions. */
  isReady(): boolean;

  /** Stream a completion, yielding incremental text. */
  stream(
    messages: ChatMessage[],
    options?: CompletionOptions,
  ): AsyncGenerator<string, void, unknown>;

  /** Drop the runtime and release GPU memory. */
  dispose(): Promise<void>;
}

export class LLMError extends Error {
  readonly cause?: unknown;

  constructor(message: string, cause?: unknown) {
    super(message);
    this.name = "LLMError";
    this.cause = cause;
  }
}