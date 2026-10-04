import {
  LLMError,
  type ChatMessage,
  type CompletionOptions,
  type LoadProgress,
  type LLMProvider,
  type LLMProviderName,
} from "./types";

type OllamaChatResponse = {
  message?: { content?: string };
  error?: string;
  done?: boolean;
};

/**
 * OllamaProvider talks to a local Ollama server over its OpenAI-compatible
 * streaming API. This is the recommended path during local development, where a
 * warm qwen3:8b avoids downloading multi-gigabyte weights into the browser.
 */
export class OllamaProvider implements LLMProvider {
  readonly name: LLMProviderName = "ollama";
  readonly model: string;

  private readonly baseUrl: string;
  private ready = false;

  constructor(baseUrl: string, model: string) {
    this.baseUrl = baseUrl;
    this.model = model;
  }

  isReady(): boolean {
    return this.ready;
  }

  async load(onProgress: (progress: LoadProgress) => void): Promise<void> {
    onProgress({ progress: null, text: `Connecting to Ollama at ${this.baseUrl}…` });

    let tags: { models?: Array<{ name?: string; model?: string }> };
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      tags = await response.json();
    } catch (error) {
      throw new LLMError(
        `Cannot reach Ollama at ${this.baseUrl}. Start it with "ollama serve" or set NEXT_PUBLIC_LLM_PROVIDER=webllm.`,
        error,
      );
    }

    const installed = (tags.models ?? []).map((entry) => entry.name ?? entry.model ?? "");
    const found = installed.some(
      (name) => name === this.model || name.startsWith(`${this.model}:`),
    );
    if (installed.length > 0 && !found) {
      throw new LLMError(
        `Model "${this.model}" is not pulled on this Ollama server. Run "ollama pull ${this.model}".`,
      );
    }

    this.ready = true;
    onProgress({ progress: 1, text: `Ollama ready — ${this.model}` });
  }

  async *stream(
    messages: ChatMessage[],
    options: CompletionOptions = {},
  ): AsyncGenerator<string, void, unknown> {
    if (!this.ready) {
      throw new LLMError("Ollama provider is not connected yet.");
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/api/chat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          model: this.model,
          messages,
          stream: true,
          options: {
            temperature: options.temperature ?? 0.7,
            top_p: options.topP ?? 0.95,
            num_predict: options.maxTokens ?? 1024,
          },
        }),
      });
    } catch (error) {
      throw new LLMError(`Request to Ollama failed.`, error);
    }

    if (!response.ok || !response.body) {
      const detail = await response.text().catch(() => "");
      throw new LLMError(`Ollama responded with HTTP ${response.status}. ${detail}`.trim());
    }

    for await (const line of readNdjson(response.body)) {
      if (!line) continue;

      let parsed: OllamaChatResponse;
      try {
        parsed = JSON.parse(line) as OllamaChatResponse;
      } catch {
        continue;
      }

      if (parsed.error) {
        throw new LLMError(parsed.error);
      }

      const delta = parsed.message?.content;
      if (delta) yield delta;
    }
  }

  async dispose(): Promise<void> {
    this.ready = false;
  }
}

/**
 * readNdjson decodes a streaming body into whole lines. Ollama emits
 * newline-delimited JSON objects, and the final record may not be terminated.
 */
async function* readNdjson(body: ReadableStream<Uint8Array>): AsyncGenerator<string, void, unknown> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      let newlineIndex = buffer.indexOf("\n");
      while (newlineIndex !== -1) {
        yield buffer.slice(0, newlineIndex).trim();
        buffer = buffer.slice(newlineIndex + 1);
        newlineIndex = buffer.indexOf("\n");
      }
    }

    const tail = decoder.decode().trim();
    if (tail) yield tail;
  } finally {
    reader.releaseLock();
  }
}