import { auth } from "@clerk/nextjs/server";

// =============================================================================
// Export Download — Standalone Route Handler
// The Go API's GET /api/v1/exports/:id/download requires server-to-server auth
// headers the browser doesn't have, and responds with a 302 to a presigned R2
// URL. This proxies the auth + redirect: Next.js calls the Go API with
// redirect: "manual" to read the Location header itself, then redirects the
// browser straight to R2 (not through this server), rather than streaming the
// ZIP bytes through Next.js.
// =============================================================================

export const dynamic = "force-dynamic";

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { userId, orgId } = await auth();
  if (!userId || !orgId) {
    return new Response("Unauthorized", { status: 401 });
  }

  const { id } = await params;
  if (!id || !/^[0-9a-f-]{36}$/i.test(id)) {
    return new Response("Invalid export id", { status: 400 });
  }

  const goApiUrl = process.env.GO_API_URL ?? "http://localhost:8080";

  const upstream = await fetch(`${goApiUrl}/api/v1/exports/${id}/download`, {
    headers: {
      "X-User-Id": userId,
      "X-Org-Id": orgId,
    },
    redirect: "manual",
    cache: "no-store",
  });

  if (upstream.status === 302 || upstream.status === 301) {
    const location = upstream.headers.get("location");
    if (location) {
      return Response.redirect(location, 302);
    }
  }

  if (upstream.status === 404) {
    return new Response("Export not found", { status: 404 });
  }
  if (upstream.status === 409) {
    return new Response("Export not ready", { status: 409 });
  }

  return new Response("Download failed", { status: 502 });
}
