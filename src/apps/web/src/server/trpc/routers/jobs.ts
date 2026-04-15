import { z } from "zod";
import { TRPCError } from "@trpc/server";
import type { GenerationJob, VideoModel } from "@qvora/types";
import { createTRPCRouter, workspaceProcedure } from "../init";

type ApiJob = {
  job_id: string;
  org_id: string;
  status: GenerationJob["status"];
  product_url: string;
  model?: VideoModel;
  created_at?: string;
};

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

export const jobsRouter = createTRPCRouter({
  get: workspaceProcedure
    .input(z.object({ jobId: z.string().uuid() }))
    .query(async ({ input, ctx }) => {
      const response = await fetch(`${GO_API_BASE_URL}/api/v1/jobs/${input.jobId}`, {
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
        throw new TRPCError({
          code: response.status === 404 ? "NOT_FOUND" : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      const payload = (await response.json()) as ApiJob;

      return {
        jobId: payload.job_id,
        orgId: payload.org_id,
        productUrl: payload.product_url,
        status: payload.status,
        model: payload.model ?? "veo3",
        createdAt: payload.created_at ?? new Date().toISOString(),
      };
    }),

  list: workspaceProcedure
    .input(
      z.object({
        limit: z.number().min(1).max(50).default(20),
      }),
    )
    .query(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/jobs?limit=${input.limit.toString()}`,
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
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      const payload = (await response.json()) as {
        org_id: string;
        jobs: ApiJob[];
      };

      return {
        orgId: payload.org_id,
        jobs: payload.jobs.map((job) => ({
          jobId: job.job_id,
          orgId: job.org_id,
          productUrl: job.product_url,
          status: job.status,
          model: job.model ?? "veo3",
          createdAt: job.created_at ?? new Date().toISOString(),
        })),
      };
    }),
});
