import { z } from "zod";
import Stripe from "stripe";
import { TRPCError } from "@trpc/server";
import { createTRPCRouter, workspaceProcedure } from "../init";

const PLAN_PRICE_ENV: Record<"starter" | "growth" | "agency", string> = {
  starter: "STRIPE_STARTER_PRICE_ID",
  growth: "STRIPE_GROWTH_PRICE_ID",
  agency: "STRIPE_AGENCY_PRICE_ID",
};

function getStripeClient(): Stripe {
  const secretKey = process.env.STRIPE_SECRET_KEY;
  if (!secretKey) {
    throw new TRPCError({
      code: "INTERNAL_SERVER_ERROR",
      message: "STRIPE_SECRET_KEY is not configured",
    });
  }
  return new Stripe(secretKey, { apiVersion: "2025-02-24.acacia" });
}

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

export const workspaceRouter = createTRPCRouter({
  // Get current workspace info
  get: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/workspaces/${ctx.orgId}`, {
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

    const payload = (await response.json()) as {
      workspace_id?: string;
      org_id: string;
      plan_tier: "starter" | "growth" | "agency";
      subscription_status: "trialing" | "active" | "past_due" | "canceled";
      trial_ends_at?: string;
      stripe_customer_id?: string | null;
    };

    return {
      workspaceId: payload.workspace_id,
      orgId: payload.org_id,
      planTier: payload.plan_tier,
      subscriptionStatus: payload.subscription_status,
      trialEndsAt: payload.trial_ends_at,
      stripeCustomerId: payload.stripe_customer_id ?? null,
    };
  }),

  // Get workspace usage counters
  getUsage: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/workspaces/${ctx.orgId}/usage`, {
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

    const payload = (await response.json()) as {
      org_id: string;
      workspace_id: string;
      used_variants: number;
      last_invoice_id?: string;
      stripe_subscription_id?: string;
      last_reset_at?: string;
    };

    return {
      orgId: payload.org_id,
      workspaceId: payload.workspace_id,
      usedVariants: payload.used_variants,
      lastInvoiceId: payload.last_invoice_id,
      stripeSubscriptionId: payload.stripe_subscription_id,
      lastResetAt: payload.last_reset_at,
    };
  }),

  // Get workspace brand kit
  getBrandKit: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/workspaces/${ctx.orgId}/brand-kit`, {
      method: "GET",
      headers: buildInternalHeaders({
        userId: ctx.userId,
        orgId: ctx.orgId,
        orgRole: ctx.orgRole,
        headers: ctx.headers,
      }),
      cache: "no-store",
    });

    if (response.status === 404) {
      return null;
    }

    if (!response.ok) {
      throw new TRPCError({
        code: "INTERNAL_SERVER_ERROR",
        message: await parseApiError(response),
      });
    }

    const payload = (await response.json()) as {
      id: string;
      workspace_id: string;
      name: string;
      logo_r2_key?: string;
      primary_color: string;
      secondary_color?: string;
      font_family?: string;
      watermark_enabled: boolean;
      created_at?: string;
      updated_at?: string;
    };

    return {
      id: payload.id,
      workspaceId: payload.workspace_id,
      name: payload.name,
      logoR2Key: payload.logo_r2_key,
      primaryColor: payload.primary_color,
      secondaryColor: payload.secondary_color,
      fontFamily: payload.font_family,
      watermarkEnabled: payload.watermark_enabled,
      createdAt: payload.created_at,
      updatedAt: payload.updated_at,
    };
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
      const response = await fetch(`${GO_API_BASE_URL}/api/v1/workspaces/${ctx.orgId}/brand-kit`, {
        method: "PUT",
        headers: buildInternalHeaders({
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        }),
        body: JSON.stringify({
          name: input.name,
          primary_color: input.primaryColor,
          secondary_color: input.secondaryColor,
          font_family: input.fontFamily,
          watermark_enabled: input.watermarkEnabled,
        }),
        cache: "no-store",
      });

      if (!response.ok) {
        throw new TRPCError({
          code: response.status === 400 ? "BAD_REQUEST" : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      const payload = (await response.json()) as {
        id: string;
        org_id: string;
        workspace_id: string;
        name: string;
        logo_r2_key?: string;
        primary_color: string;
        secondary_color?: string;
        font_family?: string;
        watermark_enabled: boolean;
        created_at?: string;
        updated_at?: string;
      };

      return {
        id: payload.id,
        orgId: payload.org_id,
        workspaceId: payload.workspace_id,
        name: payload.name,
        logoR2Key: payload.logo_r2_key,
        primaryColor: payload.primary_color,
        secondaryColor: payload.secondary_color,
        fontFamily: payload.font_family,
        watermarkEnabled: payload.watermark_enabled,
        createdAt: payload.created_at,
        updatedAt: payload.updated_at,
      };
    }),

  // Creates a Stripe Checkout session for a plan tier. The subscription's
  // metadata.org_id is what the Stripe webhook's deriveOrgId() already reads
  // to route customer.subscription.updated/deleted back to this workspace —
  // no separate customer-id lookup needed for checkout itself.
  createCheckoutSession: workspaceProcedure
    .input(
      z.object({
        planTier: z.enum(["starter", "growth", "agency"]),
        returnPath: z.string().startsWith("/").default("/briefs"),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const priceEnvVar = PLAN_PRICE_ENV[input.planTier];
      const priceId = process.env[priceEnvVar];
      if (!priceId) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: `${priceEnvVar} is not configured`,
        });
      }
      const appUrl = (process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000").replace(
        /\/$/,
        "",
      );
      const stripe = getStripeClient();
      const session = await stripe.checkout.sessions.create({
        mode: "subscription",
        line_items: [{ price: priceId, quantity: 1 }],
        client_reference_id: ctx.orgId,
        subscription_data: {
          metadata: { org_id: ctx.orgId, plan_tier: input.planTier },
        },
        success_url: `${appUrl}${input.returnPath}?checkout=success`,
        cancel_url: `${appUrl}${input.returnPath}?checkout=cancelled`,
      });
      if (!session.url) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "checkout_session_missing_url",
        });
      }
      return { url: session.url };
    }),

  // Opens the Stripe customer portal. Requires stripe_customer_id to already
  // be set — captured from the checkout.session.completed webhook event after
  // a workspace's first successful checkout.
  createPortalSession: workspaceProcedure
    .input(z.object({ returnPath: z.string().startsWith("/").default("/briefs") }))
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(`${GO_API_BASE_URL}/api/v1/workspaces/${ctx.orgId}`, {
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
      const payload = (await response.json()) as { stripe_customer_id?: string | null };
      if (!payload.stripe_customer_id) {
        throw new TRPCError({
          code: "BAD_REQUEST",
          message: "no_stripe_customer_yet — complete a checkout first",
        });
      }
      const appUrl = (process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000").replace(
        /\/$/,
        "",
      );
      const stripe = getStripeClient();
      const session = await stripe.billingPortal.sessions.create({
        customer: payload.stripe_customer_id,
        return_url: `${appUrl}${input.returnPath}`,
      });
      return { url: session.url };
    }),
});
