import { z } from "zod";
import type { PlanTier, SubscriptionStatus } from "@qvora/types";
import { createTRPCRouter, workspaceProcedure } from "../init";

export const workspaceRouter = createTRPCRouter({
  // Get current workspace info
  get: workspaceProcedure.query(async ({ ctx }) => {
    const planTier: PlanTier = "starter";
    const subscriptionStatus: SubscriptionStatus = "trialing";

    // TODO: GET from Go API /api/v1/workspaces/:orgId
    return {
      orgId: ctx.orgId,
      planTier,
      subscriptionStatus,
    };
  }),

  // Get workspace brand kit
  getBrandKit: workspaceProcedure.query(async ({ ctx }) => {
    // TODO: GET from Go API /api/v1/workspaces/:orgId/brand-kit
    return null;
  }),

  // Upsert brand kit
  upsertBrandKit: workspaceProcedure
    .input(
      z.object({
        name: z.string().min(1).max(100),
        primaryColor: z.string().regex(/^#[0-9a-f]{6}$/i, "Must be a hex color"),
        secondaryColor: z
          .string()
          .regex(/^#[0-9a-f]{6}$/i)
          .optional(),
        fontFamily: z.string().optional(),
        watermarkEnabled: z.boolean().default(true),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      // TODO: PUT to Go API /api/v1/workspaces/:orgId/brand-kit
      return { orgId: ctx.orgId, ...input };
    }),
});
