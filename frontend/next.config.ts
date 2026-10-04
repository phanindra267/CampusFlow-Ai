import path from "node:path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Pin the Turbopack root to this app. Without it the resolver walks up the
  // filesystem and can pick up an unrelated package.json outside the repo.
  turbopack: {
    root: path.resolve(import.meta.dirname),
  },
};

export default nextConfig;