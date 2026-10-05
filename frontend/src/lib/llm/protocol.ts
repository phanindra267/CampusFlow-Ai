/**
 * Worker protocol between the main thread and the WebLLM worker.
 *
 * Kept in its own module so both sides import the same discriminated union
 * without the worker dragging React or the main-thread engine in.
 */

import type { ChatMessage, CompletionOptions, LoadProgress } from "./types";

export type WorkerRequest =
  | {
      type: "init";
      model: string;
      contextWindowSize: number;
    }
  | {
      type: "generate";
      requestId: string;
      messages: ChatMessage[];
      options?: CompletionOptions;
    }
  | {
      type: "interrupt";
      requestId: string;
    }
  | {
      type: "dispose";
    };

export type WorkerResponse =
  | { type: "init-progress"; progress: LoadProgress }
  | { type: "ready"; model: string }
  | { type: "init-error"; message: string }
  | { type: "chunk"; requestId: string; delta: string }
  | { type: "complete"; requestId: string; text: string }
  | { type: "error"; requestId: string; message: string }
  | { type: "cancelled"; requestId: string };