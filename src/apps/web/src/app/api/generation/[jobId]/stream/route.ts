import { auth } from "@clerk/nextjs/server";
import type { SSEEvent } from "@qvora/types";

// =============================================================================
// SSE Generation Stream — Standalone Route Handler
// NOT a tRPC subscription. Uses text/event-stream + ReadableStream.
// Polls Go API and forwards job status events to the client.
// =============================================================================

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(_req: Request, { params }: { params: Promise<{ jobId: string }> }) {
  const { userId, orgId } = await auth();

  if (!userId || !orgId) {
    return new Response("Unauthorized", { status: 401 });
  }

  const { jobId } = await params;

  if (!jobId || !/^[0-9a-f-]{36}$/.test(jobId)) {
    return new Response("Invalid jobId", { status: 400 });
  }

  const goApiUrl = process.env.GO_API_URL ?? "http://localhost:8080";

  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder();

      const send = (event: SSEEvent) => {
        const data = `data: ${JSON.stringify(event)}\n\n`;
        controller.enqueue(encoder.encode(data));
      };

      // Forward SSE from Go API to client
      // The Go API exposes its own SSE stream; we proxy it here with auth validation
      try {
        const upstream = await fetch(`${goApiUrl}/api/v1/jobs/${jobId}/stream?org_id=${orgId}`, {
          headers: {
            Accept: "text/event-stream",
            "X-User-Id": userId,
            "X-Org-Id": orgId,
          },
        });

        if (!upstream.ok || !upstream.body) {
          send({
            type: "job_failed",
            jobId,
            data: { error: "upstream_unavailable" },
            timestamp: new Date().toISOString(),
          });
          controller.close();
          return;
        }

        const reader = upstream.body.getReader();
        const decoder = new TextDecoder();

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          // Forward raw SSE bytes — already in `data: {...}\n\n` format
          controller.enqueue(value);
          // Log forwarded chunk length for observability
          void decoder.decode(value);
        }
      } catch (_err) {
        send({
          type: "job_failed",
          jobId,
          data: { error: "stream_error" },
          timestamp: new Date().toISOString(),
        });
      } finally {
        controller.close();
      }
    },
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no", // Disable Nginx buffering (Railway / Vercel)
    },
  });
}
