import type { NextConfig } from "next";
import { withSentryConfig } from "@sentry/nextjs";

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
          {
            key: "Content-Security-Policy",
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://clerk.qvora.com https://*.clerk.accounts.dev",
              "style-src 'self' 'unsafe-inline'",
              "img-src 'self' data: blob: https://assets.qvora.com https://image.mux.com",
              "media-src 'self' blob: https://stream.mux.com",
              // Sentry (error ingest) and PostHog (product analytics) both make
              // their own fetch/beacon calls from the browser — without these,
              // CSP would silently drop every event client-side.
              "connect-src 'self' https://api.qvora.com https://*.clerk.accounts.dev https://clerk.qvora.com https://*.sentry.io https://*.posthog.com",
              "font-src 'self' data:",
              "frame-ancestors 'none'",
            ].join("; "),
          },
          {
            key: "Strict-Transport-Security",
            value: "max-age=31536000; includeSubDomains",
          },
        ],
      },
    ];
  },
};

// withSentryConfig auto-discovers sentry.client/server/edge.config.ts and
// injects source-map upload during build. No-ops cleanly if SENTRY_ORG/
// SENTRY_PROJECT/SENTRY_AUTH_TOKEN aren't set (e.g. local dev).
export default withSentryConfig(nextConfig, {
  org: process.env.SENTRY_ORG,
  project: process.env.SENTRY_PROJECT,
  silent: true,
  widenClientFileUpload: true,
  disableLogger: true,
});
