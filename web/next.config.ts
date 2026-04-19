import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

/** Go API base for dev rewrites (must match EventSource target in web/lib/sseUrl.ts). */
const devBackendBase = (
  process.env.DLM_BACKEND_ORIGIN?.trim() ||
  process.env.NEXT_PUBLIC_DLM_API_ORIGIN?.trim() ||
  "http://127.0.0.1:8080"
).replace(/\/$/, "");

const nextConfig: NextConfig = {
  output: "export",
  images: { unoptimized: true },
  ...(isDev
    ? {
        async rewrites() {
          return [
            {
              source: "/api/v1/:path*",
              destination: `${devBackendBase}/api/v1/:path*`,
            },
          ];
        },
      }
    : {}),
};

export default nextConfig;
