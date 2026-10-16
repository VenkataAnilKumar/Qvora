"use client";

import { useEffect } from "react";
import posthog from "posthog-js";

// posthog-js's client is a module-level singleton — components elsewhere call
// `posthog.capture(...)` directly via the same `import posthog from "posthog-js"`,
// no Context/hook needed. This component's only job is to initialize it once.
export function PostHogProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    const key = process.env.NEXT_PUBLIC_POSTHOG_KEY;
    if (!key || posthog.__loaded) return;

    posthog.init(key, {
      api_host: process.env.NEXT_PUBLIC_POSTHOG_HOST ?? "https://us.i.posthog.com",
      person_profiles: "identified_only",
      capture_pageview: true,
    });
  }, []);

  return <>{children}</>;
}
