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

export const signalRouter = createTRPCRouter({
  initiateOAuth: workspaceProcedure
    .input(
      z.object({
        platform: z.enum(["meta", "tiktok"]),
      }),
    )
    .query(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/oauth/${encodeURIComponent(input.platform)}/initiate`,
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
          code: response.status === 400 ? "BAD_REQUEST" : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        platform: "meta" | "tiktok";
        state: string;
        redirect_uri: string;
        url: string;
      };
    }),

  listConnections: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/signal/connections`, {
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
        code: "INTERNAL_SERVER_ERROR",
        message: await parseApiError(response),
      });
    }

    const payload = (await response.json()) as {
      org_id: string;
      workspace_id: string;
      connections: Array<{
        platform: "meta" | "tiktok";
        status: "connected" | "disconnected" | "token_expired";
        account_id: string;
        account_name?: string;
        token_expires_at?: string;
        last_synced_at?: string;
      }>;
    };

    return {
      orgId: payload.org_id,
      workspaceId: payload.workspace_id,
      connections: payload.connections,
    };
  }),

  upsertConnection: workspaceProcedure
    .input(
      z.object({
        platform: z.enum(["meta", "tiktok"]),
        accountId: z.string().min(1),
        accountName: z.string().optional(),
        status: z.enum(["connected", "disconnected", "token_expired"]).default("connected"),
        tokenExpiresAt: z.string().datetime().optional(),
        lastSyncedAt: z.string().datetime().optional(),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/connections/${encodeURIComponent(input.platform)}`,
        {
          method: "PUT",
          headers: buildInternalHeaders({
            userId: ctx.userId,
            orgId: ctx.orgId,
            orgRole: ctx.orgRole,
            headers: ctx.headers,
          }),
          body: JSON.stringify({
            account_id: input.accountId,
            account_name: input.accountName,
            status: input.status,
            token_expires_at: input.tokenExpiresAt,
            last_synced_at: input.lastSyncedAt,
          }),
          cache: "no-store",
        },
      );

      if (!response.ok) {
        throw new TRPCError({
          code: response.status === 400 ? "BAD_REQUEST" : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        platform: "meta" | "tiktok";
        account_id: string;
        status: "connected" | "disconnected" | "token_expired";
        updated: boolean;
      };
    }),

  patchConnectionHealth: workspaceProcedure
    .input(
      z.object({
        platform: z.enum(["meta", "tiktok"]),
        accountId: z.string().min(1),
        status: z.enum(["connected", "disconnected", "token_expired"]).optional(),
        errorReason: z.string().optional(),
        tokenExpiresAt: z.string().datetime().optional(),
        lastSyncedAt: z.string().datetime().optional(),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/connections/${encodeURIComponent(input.platform)}/${encodeURIComponent(input.accountId)}/health`,
        {
          method: "PATCH",
          headers: buildInternalHeaders({
            userId: ctx.userId,
            orgId: ctx.orgId,
            orgRole: ctx.orgRole,
            headers: ctx.headers,
          }),
          body: JSON.stringify({
            status: input.status,
            error_reason: input.errorReason,
            token_expires_at: input.tokenExpiresAt,
            last_synced_at: input.lastSyncedAt,
          }),
          cache: "no-store",
        },
      );

      if (!response.ok) {
        throw new TRPCError({
          code:
            response.status === 404
              ? "NOT_FOUND"
              : response.status === 400
                ? "BAD_REQUEST"
                : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        platform: "meta" | "tiktok";
        account_id: string;
        updated: boolean;
      };
    }),

  getDashboard: workspaceProcedure
    .input(
      z.object({
        days: z.number().min(1).max(90).default(30),
      }),
    )
    .query(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/dashboard?days=${input.days}`,
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

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        days: number;
        totals: {
          impressions: number;
          clicks: number;
          spend_usd: number;
          conversions: number;
          revenue_usd: number;
          ctr: number;
          cpa: number;
          roas: number;
        };
        recommendation_feedback: {
          total: number;
          accepted: number;
          acceptance_rate: number;
          current_7d_rate: number;
          previous_7d_rate: number;
          acceptance_delta_pct_points: number;
          trend: Array<{
            date: string;
            total: number;
            accepted: number;
            acceptance_rate: number;
          }>;
        };
        by_angle: Array<{
          angle: string;
          impressions: number;
          clicks: number;
          spend_usd: number;
          conversions: number;
          revenue_usd: number;
          ctr: number;
        }>;
        by_platform: Array<{
          platform: string;
          impressions: number;
          clicks: number;
          spend_usd: number;
          conversions: number;
          revenue_usd: number;
          ctr: number;
        }>;
      };
    }),

  detectFatigue: workspaceProcedure
    .input(
      z.object({
        days: z.number().min(1).max(90).default(30),
        dropPct: z.number().min(1).max(95).default(30),
        sustainedDays: z.number().min(3).max(7).default(3),
        minPeakCtr: z.number().min(0.0001).default(0.01),
        minImpressions: z.number().min(1).default(1000),
        persist: z.boolean().default(true),
      }),
    )
    .query(async ({ input, ctx }) => {
      const params = new URLSearchParams({
        days: String(input.days),
        drop_pct: String(input.dropPct),
        sustained_days: String(input.sustainedDays),
        min_peak_ctr: String(input.minPeakCtr),
        min_impressions: String(input.minImpressions),
        persist: String(input.persist),
      });

      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/fatigue?${params.toString()}`,
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

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        days: number;
        drop_pct: number;
        sustained_days: number;
        persisted: boolean;
        alerts: Array<{
          variant_id: string;
          detected_on: string;
          angle: string;
          current_ctr: number;
          peak_ctr: number;
          drop_pct: number;
          sustained_days: number;
          suggested_action: string;
          refresh_link: string;
        }>;
      };
    }),

  listFatigueEvents: workspaceProcedure
    .input(
      z.object({
        limit: z.number().min(1).max(100).default(20),
        status: z.enum(["active", "resolved", "all"]).default("active"),
      }),
    )
    .query(async ({ input, ctx }) => {
      const params = new URLSearchParams({
        limit: String(input.limit),
        status: input.status,
      });

      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/fatigue/events?${params.toString()}`,
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

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        status: "active" | "resolved" | "all";
        events: Array<{
          variant_id: string;
          detected_on: string;
          angle: string;
          current_ctr: number;
          peak_ctr: number;
          drop_pct: number;
          sustained_days: number;
          suggested_action: string;
          status: "active" | "resolved";
          first_detected_at: string;
          last_detected_at: string;
          resolved_at?: string;
        }>;
      };
    }),

  getRecommendations: workspaceProcedure
    .input(
      z.object({
        days: z.number().min(1).max(90).default(90),
        refresh: z.boolean().default(false),
      }),
    )
    .query(async ({ input, ctx }) => {
      const params = new URLSearchParams({
        days: String(input.days),
        refresh: String(input.refresh),
      });

      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/recommendations?${params.toString()}`,
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

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        refreshed: boolean;
        recommendations: Array<{
          angle: string;
          suggested_hook: string;
          rationale: string;
          confidence_score: number;
          impression_volume: number;
          window_days: number;
          last_generated_at: string;
        }>;
      };
    }),

  submitRecommendationFeedback: workspaceProcedure
    .input(
      z.object({
        angle: z.string().min(1),
        action: z.enum(["accept", "ignore"]),
        source: z.string().min(1).default("brief_create_panel"),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(`${GO_API_BASE_URL}/api/v1/signal/recommendations/feedback`, {
        method: "POST",
        headers: buildInternalHeaders({
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        }),
        body: JSON.stringify({
          angle: input.angle,
          action: input.action,
          source: input.source,
        }),
        cache: "no-store",
      });

      if (!response.ok) {
        throw new TRPCError({
          code: response.status === 400 ? "BAD_REQUEST" : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        angle: string;
        action: "accept" | "ignore";
        source: string;
        tracked: boolean;
      };
    }),

  listRecommendationFeedbackByAngle: workspaceProcedure
    .input(
      z.object({
        days: z.number().min(1).max(90).default(30),
        limit: z.number().min(1).max(50).default(10),
      }),
    )
    .query(async ({ input, ctx }) => {
      const params = new URLSearchParams({
        days: String(input.days),
        limit: String(input.limit),
      });

      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/signal/recommendations/feedback?${params.toString()}`,
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

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        days: number;
        feedback_by_angle: Array<{
          angle: string;
          total: number;
          accepted: number;
          ignored: number;
          acceptance_rate: number;
        }>;
      };
    }),

  triggerMetricsSync: workspaceProcedure.mutation(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/internal/signal/metrics/sync-all`, {
      method: "POST",
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
        code: "INTERNAL_SERVER_ERROR",
        message: await parseApiError(response),
      });
    }

    return (await response.json()) as {
      synced_connections: number;
      synced_rows: number;
      days: number;
    };
  }),
});
