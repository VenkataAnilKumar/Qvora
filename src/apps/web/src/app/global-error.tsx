"use client";

import * as Sentry from "@sentry/nextjs";
import NextError from "next/error";
import { useEffect } from "react";

export default function GlobalError({ error }: { error: Error & { digest?: string } }) {
  useEffect(() => {
    Sentry.captureException(error);
  }, [error]);

  return (
    <html lang="en">
      <body>
        {/* Next.js's built-in default error UI — keeps this page dependency-free
            beyond Sentry so a broken root layout doesn't also break this page. */}
        <NextError statusCode={0} />
      </body>
    </html>
  );
}
