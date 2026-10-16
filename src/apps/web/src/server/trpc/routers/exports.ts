import { z } from "zod";
import { TRPCError } from "@trpc/server";
import { createTRPCRouter, workspaceProcedure } from "../init";

const GO_API_BASE_URL = process.env.GO_API_URL ?? "http://localhost:8080";

const buildInternalHeaders = (ctx: {
  userId: string;
  orgId: string;
  orgRole: string | null | undefined;
  headers: Headers;
}) => {
  const headers = new Headers();
  headers.set("Content-Type", "application/json");
  headers.set("Accept", "application/json");
  headers.set("X-User-Id", ctx.userId);
  headers.set("X-Org-Id", ctx.orgId);
  headers.set("X-Org-Role", ctx.orgRole ?? "member");

  const authHeader = ctx.headers.get("authorization");
  if (authHeader) {
    headers.set("Authorization", authHeader);
  }

  const internalApiKey = process.env.INTERNAL_API_KEY;
  if (internalApiKey) {
    headers.set("X-Internal-Api-Key", internalApiKey);
  }

  return headers;
};

export const exportsRouter = createTRPCRouter({
  list: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/exports`, {
      method: "GET",
      headers: buildInternalHeaders({
        userId: ctx.userId,
        orgId: ctx.orgId,
        orgRole: ctx.orgRole,
        headers: ctx.headers,
      }),
      cache: "no-store",
    });

    if (!response.ok) {
      throw new Error("export_list_failed");
    }

    const payload = (await response.json()) as {
      org_id: string;
      exports: Array<{
        id: string;
        variant_id: string;
        destination: string;
        status: string;
        created_at?: string;
        updated_at?: string;
      }>;
    };

    return {
      orgId: payload.org_id,
      exports: payload.exports,
    };
  }),

  // Bundles one or more complete variants into a single ZIP (built server-side
  // by the Go API from their already-processed R2 assets) and returns an API
  // path to download it through — not a raw R2 URL, so download_count tracking
  // in GET /api/v1/exports/:id/download always runs.
  create: workspaceProcedure
    .input(
      z.object({
        variantIds: z.array(z.string().uuid()).min(1).max(100),
        format: z.enum(["mp4_1080p"]).default("mp4_1080p"),
        platform: z.enum(["meta", "tiktok", "youtube_shorts"]).default("tiktok"),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(`${GO_API_BASE_URL}/api/v1/exports`, {
        method: "POST",
        headers: buildInternalHeaders({
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        }),
        body: JSON.stringify({
          variant_ids: input.variantIds,
          format: input.format,
          platform: input.platform,
        }),
        cache: "no-store",
      });

      if (!response.ok) {
        const body = (await response.json().catch(() => ({}))) as {
          error?: string;
          message?: string;
        };
        throw new TRPCError({
          code:
            response.status === 400
              ? "BAD_REQUEST"
              : response.status === 404
                ? "NOT_FOUND"
                : response.status === 409
                  ? "CONFLICT"
                  : "INTERNAL_SERVER_ERROR",
          message: body.message ?? body.error ?? "export_create_failed",
        });
      }

      const payload = (await response.json()) as {
        export_ids: string[];
        variant_count: number;
        format: string;
        platform: string;
      };

      const firstId = payload.export_ids[0];
      if (!firstId) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "export_create_returned_no_id",
        });
      }

      return {
        exportIds: payload.export_ids,
        variantCount: payload.variant_count,
        format: payload.format,
        platform: payload.platform,
        // Next.js-owned proxy route (app/api/exports/[id]/download) — NOT the
        // Go API's own path, which the browser can't call directly (no auth
        // headers, and the Go response is itself a redirect to R2).
        downloadUrl: `/api/exports/${firstId}/download`,
      };
    }),
});
