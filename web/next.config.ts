import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  output: "export",
  images: { unoptimized: true },
  ...(isDev
    ? {
        async rewrites() {
          return [
            {
              source: "/api/v1/:path*",
              destination: "http://127.0.0.1:8080/api/v1/:path*",
            },
          ];
        },
      }
    : {}),
};

export default nextConfig;
