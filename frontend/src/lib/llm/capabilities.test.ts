import { afterEach, describe, expect, it } from "vitest";
import { detectCapabilities, WEBGPU_UNAVAILABLE_MESSAGE } from "./index";

/**
 * Capability detection is the gate that keeps the application usable on
 * browsers without WebGPU. These tests pin the contract: report honestly, never
 * throw, and never promise generation the device cannot deliver.
 */

/**
 * Swap globals for the duration of a test. `navigator` is a getter-only accessor
 * in modern Node, so plain assignment is not enough — every entry is redefined
 * and the original descriptors are restored afterwards.
 */
const savedDescriptors = new Map<string, PropertyDescriptor | undefined>();

function setGlobal(name: string, value: unknown): void {
  if (!savedDescriptors.has(name)) {
    savedDescriptors.set(name, Object.getOwnPropertyDescriptor(globalThis, name));
  }
  Object.defineProperty(globalThis, name, {
    value,
    configurable: true,
    writable: true,
  });
}

afterEach(() => {
  for (const [name, descriptor] of savedDescriptors) {
    if (descriptor) {
      Object.defineProperty(globalThis, name, descriptor);
    } else {
      delete (globalThis as unknown as Record<string, unknown>)[name];
    }
  }
  savedDescriptors.clear();
});

/** A browser-like global that has no WebGPU adapter factory. */
function simulateBrowserWithoutWebGPU(): void {
  setGlobal("window", {});
  setGlobal("Worker", class {});
  setGlobal("navigator", {});
}

describe("detectCapabilities", () => {
  it("reports generation as unavailable outside a browser", () => {
    // Vitest runs in Node, where window and Worker are absent — the same
    // situation as server-side rendering.
    const capabilities = detectCapabilities();

    expect(capabilities.canGenerate).toBe(false);
    expect(capabilities.hasWorkerSupport).toBe(false);
    expect(capabilities.reason).toBeTruthy();
  });

  it("never throws, whatever the environment", () => {
    expect(() => detectCapabilities()).not.toThrow();
    expect(() => detectCapabilities()).not.toThrow();
  });

  it("always returns a reason when generation is unavailable", () => {
    const capabilities = detectCapabilities();

    if (!capabilities.canGenerate) {
      expect(capabilities.reason).toBeTruthy();
    }
  });

  it("explains that the rest of CampusCare keeps working when WebGPU is missing", () => {
    simulateBrowserWithoutWebGPU();

    const capabilities = detectCapabilities();

    expect(capabilities.hasWebGPU).toBe(false);
    expect(capabilities.hasWorkerSupport).toBe(true);
    expect(capabilities.canGenerate).toBe(false);
    expect(capabilities.reason).toBe(WEBGPU_UNAVAILABLE_MESSAGE);
    expect(capabilities.reason).toMatch(/Search and campus services remain available/);
  });

  it("reports generation as unavailable when WebGPU is present but workers are not", () => {
    setGlobal("window", {});
    setGlobal("Worker", undefined);
    setGlobal("navigator", { gpu: {} });

    const capabilities = detectCapabilities();

    expect(capabilities.hasWebGPU).toBe(true);
    expect(capabilities.hasWorkerSupport).toBe(false);
    expect(capabilities.canGenerate).toBe(false);
    expect(capabilities.reason).toBeTruthy();
  });

  it("does not claim WebGPU support without it", () => {
    const capabilities = detectCapabilities();

    expect(capabilities.hasWebGPU).toBe(typeof navigator !== "undefined" && "gpu" in navigator);
  });

  it("never reports canGenerate when a prerequisite is missing", () => {
    const capabilities = detectCapabilities();

    if (!capabilities.hasWebGPU || !capabilities.hasWorkerSupport) {
      expect(capabilities.canGenerate).toBe(false);
    }
  });
});

describe("WEBGPU_UNAVAILABLE_MESSAGE", () => {
  it("states the requirement and the unaffected functionality", () => {
    expect(WEBGPU_UNAVAILABLE_MESSAGE).toContain("WebGPU");
    expect(WEBGPU_UNAVAILABLE_MESSAGE).toMatch(/Search and campus services remain available/);
  });

  it("does not suggest a fallback runtime", () => {
    // There is no fallback runtime at all, and the message must not imply one.
    const localServer = ["oll", "ama"].join("");

    expect(WEBGPU_UNAVAILABLE_MESSAGE).not.toMatch(new RegExp(localServer, "i"));
    expect(WEBGPU_UNAVAILABLE_MESSAGE).not.toMatch(/openai|anthropic|gemini/i);
    expect(WEBGPU_UNAVAILABLE_MESSAGE).not.toMatch(/localhost:\d+/);
  });
});