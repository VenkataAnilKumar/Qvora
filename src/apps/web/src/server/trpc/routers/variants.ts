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

const parseApiError = async (response: Response) => {
  try {
    const payload = (await response.json()) as { error?: string };
    return payload.error ?? "upstream_request_failed";
  } catch {
    return "upstream_request_failed";
  }
};

export const variantsRouter = createTRPCRouter({
  getPlaybackUrl: workspaceProcedure
    .input(z.object({ variantId: z.string().uuid() }))
    .query(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/variants/${input.variantId}/playback-url`,
        {
          method: "GET",
          headers: buildInternalHeaders({
            userId: ctx.userId,
            orgId: ctx.orgId,
            orgRole: ctx.orgRole,
            headers: ctx.headers,
          }),
          cache: "no-store",
        },
      );

      if (!response.ok) {
        const code =
          response.status === 404
            ? "NOT_FOUND"
            : response.status === 409
              ? "CONFLICT"
              : response.status === 400
                ? "BAD_REQUEST"
                : "INTERNAL_SERVER_ERROR";

        throw new TRPCError({
          code,
          message: await parseApiError(response),
        });
      }

      const payload = (await response.json()) as {
        variant_id: string;
        playback_id: string;
        playback_url: string;
        token: string;
        token_expires: string;
      };

      return {
        variantId: payload.variant_id,
        playbackId: payload.playback_id,
        playbackUrl: payload.playback_url,
        token: payload.token,
        tokenExpires: payload.token_expires,
      };
    }),
});
