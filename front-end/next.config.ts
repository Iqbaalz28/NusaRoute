import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Emit a minimal, self-contained server bundle (.next/standalone) for a small
  // production Docker image. See node_modules/next/dist/docs .../17-deploying.md.
  output: "standalone",
};

export default nextConfig;
