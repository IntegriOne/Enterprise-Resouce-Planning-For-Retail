import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  // Nonaktifkan header X-Powered-By: Next.js pada respons HTTP.
  poweredByHeader: false,
  cacheComponents: true,
  partialPrefetching: true,
  reactCompiler: true,
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
