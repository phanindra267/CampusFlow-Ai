/**
 * Shared types for the on-device language-model layer.
 *
 * WebLLM running on WebGPU is the only generative runtime in this application.
 * There is no server-side LLM and no third-party hosted model API.
 */

export type ChatRole = "system" | "user" | "assistant";

export type ChatMessage = {
  role: ChatRole;
  content: string;
};

export type LoadProgress = {
  /** 0..1 overall progress, or null when progress cannot be estimated. */
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
 * AIUnavailableError signals that generative AI cannot run on this device.
 *
 * It is deliberately distinct from LLMError: the rest of CampusCare (routing,
 * structured queries and hybrid search) keeps working when it is raised.
 */
export class AIUnavailableError extends Error {
  readonly code: "no-webgpu" | "worker-failed" | "engine-failed";

  constructor(message: string, code: AIUnavailableError["code"]) {
    super(message);
    this.name = "AIUnavailableError";
    this.code = code;
  }
}

/**
 * CapabilityReport summarises what this browser can actually do. It is computed
 * without loading any model, so it is safe to call during render.
 */
export type CapabilityReport = {
  /** `navigator.gpu` exists. */
  hasWebGPU: boolean;
  /** A Web Worker can be created, which is where inference runs. */
  hasWorkerSupport: boolean;
  /** Generic engine availability: false when generative AI is impossible. */
  canGenerate: boolean;
  /** Member-facing explanation, suitable for display. */
  reason: string | null;
};

/**
 * LLMSettings is the whole configurable surface of the generative layer.
 * Model selection is data, not code, so swapping models never requires a
 * rewrite of the application.
 */
export type LLMSettings = {
  /** A `prebuiltAppConfig.model_list` id from @mlc-ai/web-llm. */
  model: string;
  contextWindowSize: number;
  temperature: number;
  topP: number;
  maxTokens: number;
};

/**
 * CompletionRequest is the payload handed to the worker for one generation.
 */
export type CompletionRequest = {
  requestId: string;
  messages: ChatMessage[];
  options?: CompletionOptions;
};