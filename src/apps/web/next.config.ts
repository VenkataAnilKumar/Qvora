import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Turborepo workspace packages
  transpilePackages: ["@qvora/ui", "@qvora/types"],

  images: {
    remotePatterns: [
      // Cloudflare R2 assets
      {
        protocol: "https",
        hostname: "assets.qvora.com",
      },
      // Mux thumbnails
      {
        protocol: "https",
        hostname: "image.mux.com",
      },
    ],
  },

  // Disable x-powered-by header
  poweredByHeader: false,

  // Strict mode
  reactStrictMode: true,

  // Security headers
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
