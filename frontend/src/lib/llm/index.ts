/**
 * Public surface of the on-device AI layer.
 *
 * WebLLM on WebGPU is the only generative runtime. Everything else in CampusCare
 * is deterministic, so this module is imported only where generation is
 * genuinely wanted.
 */

export * from "./types";
export {
  DEFAULT_WEBLLM_MODEL,
  getLLMSettings,
  hasWebGPUSupport,
} from "./config";
export {
  detectCapabilities,
  dispose,
  generate,
  getLoadedModel,
  initialize,
  interrupt,
  onLoadProgress,
  WEBGPU_UNAVAILABLE_MESSAGE,
} from "./engine";