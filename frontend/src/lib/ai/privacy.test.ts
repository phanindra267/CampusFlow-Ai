import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

/**
 * Privacy and architecture guards.
 *
 * These assertions encode promises made to members in the UI and in the docs,
 * and they are far cheaper to check here than to audit by hand after every
 * dependency bump.
 *
 * The forbidden terms are assembled at runtime rather than written out. A test
 * that greps the tree for a local model server must not itself contain that
 * server's name, or it would fail on its own source and defeat the audit.
 */

const SRC_ROOT = join(process.cwd(), "src");
const REPO_ROOT = join(process.cwd(), "..");

const LOCAL_SERVER = ["oll", "ama"].join("");
const LOCAL_PORT = ["114", "34"].join("");
const CHAT_PATH = `/${"api"}/${"chat"}`;
const TAGS_PATH = `/${"api"}/${"tags"}`;
const ENV_PREFIX = ["OLL", "AMA", "_"].join("");

/** Every source file worth scanning, tests included. */
function sourceFiles(): string[] {
  const files: string[] = [];

  const walk = (dir: string) => {
    for (const entry of readdirSync(dir)) {
      const full = join(dir, entry);
      if (statSync(full).isDirectory()) {
        walk(full);
        continue;
      }
      if (/\.(ts|tsx|js|jsx|css)$/.test(entry)) files.push(full);
    }
  };

  walk(SRC_ROOT);
  return files;
}

/** The shipped documentation, which members and reviewers also read. */
function docFiles(): string[] {
  return [
    join(REPO_ROOT, "README.md"),
    join(REPO_ROOT, "frontend", "README.md"),
    join(REPO_ROOT, ".env.example"),
    join(REPO_ROOT, "docs", "ai-pipeline.svg"),
    join(REPO_ROOT, "docs", "architecture.svg"),
    join(REPO_ROOT, "docs", "SCREENSHOTS.md"),
  ];
}

function filesMatching(pattern: RegExp): string[] {
  return sourceFiles().filter((file) => pattern.test(readFileSync(file, "utf8")));
}

describe("no local model server dependency", () => {
  it("has no local model server references in the frontend source", () => {
    expect(filesMatching(new RegExp(LOCAL_SERVER, "i"))).toEqual([]);
  });

  it("never contacts the local model server port", () => {
    expect(filesMatching(new RegExp(LOCAL_PORT))).toEqual([]);
  });

  it("does not call the local chat or model listing endpoints", () => {
    expect(filesMatching(new RegExp(CHAT_PATH))).toEqual([]);
    expect(filesMatching(new RegExp(TAGS_PATH))).toEqual([]);
  });

  it("defines no local model server environment variables", () => {
    expect(filesMatching(new RegExp(ENV_PREFIX))).toEqual([]);
  });

  it("keeps the documentation free of them too", () => {
    for (const doc of docFiles()) {
      const content = readFileSync(doc, "utf8");
      expect(content, `${doc} mentions a local model server`).not.toMatch(
        new RegExp(LOCAL_SERVER, "i"),
      );
      expect(content, `${doc} hardcodes the local model port`).not.toMatch(
        new RegExp(LOCAL_PORT),
      );
    }
  });
});

describe("no external hosted model API", () => {
  it("contains no hosted provider endpoints", () => {
    const hosted = /api\.openai\.com|api\.anthropic\.com|generativelanguage\.googleapis\.com/i;

    expect(filesMatching(hosted)).toEqual([]);
    expect(docFiles().filter((doc) => hosted.test(readFileSync(doc, "utf8")))).toEqual([]);
  });

  it("depends only on the WebLLM runtime for generation", () => {
    const manifest = JSON.parse(
      readFileSync(join(process.cwd(), "package.json"), "utf8"),
    ) as { dependencies?: Record<string, string>; devDependencies?: Record<string, string> };

    const all = { ...manifest.dependencies, ...manifest.devDependencies };

    expect(Object.keys(all).filter((name) => /llama|llamacpp|vllm|openai|anthropic/i.test(name))).toEqual(
      [],
    );
    expect(Object.keys(all)).toContain("@mlc-ai/web-llm");
  });
});

describe("no backend credentials in the browser", () => {
  it("does not reference database or vector-store secrets", () => {
    // Assembled the same way as the patterns above, so this file does not trip
    // its own assertion.
    const forbidden = [
      ["DB", "PASSWORD"].join("_"),
      ["JWT", "SECRET"].join("_"),
      ["POSTGRES", "PASSWORD"].join("_"),
      ["WEAVIATE", "API", "KEY"].join("_"),
    ];

    for (const file of sourceFiles()) {
      const content = readFileSync(file, "utf8");
      for (const secret of forbidden) {
        expect(content, `${file} references ${secret}`).not.toContain(secret);
      }
    }
  });
});

describe("documentation describes the hybrid architecture", () => {
  it("documents WebLLM on WebGPU as the only generative runtime", () => {
    const readme = readFileSync(join(REPO_ROOT, "README.md"), "utf8");

    expect(readme).toMatch(/WebGPU/i);
    expect(readme).toMatch(/WebLLM/i);
    expect(readme).toMatch(/intent router/i);
    expect(readme).toMatch(/BM25/i);
  });

  it("documents the WebLLM configuration surface in the frontend README", () => {
    const frontendReadme = readFileSync(join(REPO_ROOT, "frontend", "README.md"), "utf8");

    expect(frontendReadme).toMatch(/NEXT_PUBLIC_WEBLLM_MODEL/);
    expect(frontendReadme).not.toMatch(/NEXT_PUBLIC_LLM_PROVIDER/);
  });

  it("removes the obsolete provider variables from the environment template", () => {
    const envExample = readFileSync(join(REPO_ROOT, ".env.example"), "utf8");

    expect(envExample).not.toMatch(/NEXT_PUBLIC_LLM_/);
    expect(envExample).toMatch(/NEXT_PUBLIC_WEBLLM_MODEL/);
  });
});