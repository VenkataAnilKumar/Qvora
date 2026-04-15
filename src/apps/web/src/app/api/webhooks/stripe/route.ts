import { headers } from "next/headers";
import Stripe from "stripe";

const GO_API_BASE_URL = process.env.GO_API_URL ?? "http://localhost:8080";

type PlanTier = "starter" | "growth" | "agency";

function derivePlanTier(subscription: Stripe.Subscription): PlanTier | null {
  const metadataPlan = subscription.metadata?.plan_tier?.toLowerCase();
  if (metadataPlan === "starter" || metadataPlan === "growth" || metadataPlan === "agency") {
    return metadataPlan;
  }

  const lookupKey = subscription.items.data.at(0)?.price.lookup_key?.toLowerCase();
  if (lookupKey?.includes("agency")) return "agency";
  if (lookupKey?.includes("growth")) return "growth";
  if (lookupKey?.includes("starter")) return "starter";

  return null;
}

function deriveOrgId(subscription: Stripe.Subscription): string | null {
  const candidates = [
    subscription.metadata?.org_id,
    subscription.metadata?.workspace_org_id,
    subscription.metadata?.clerk_org_id,
  ];

  for (const candidate of candidates) {
    const value = candidate?.trim();
    if (value) {
      return value;
    }
  }

  return null;
}

function deriveOrgIdFromInvoice(invoice: Stripe.Invoice): string | null {
  const candidates = [
    invoice.metadata?.org_id,
    invoice.metadata?.workspace_org_id,
    invoice.metadata?.clerk_org_id,
  ];

  for (const candidate of candidates) {
    const value = candidate?.trim();
    if (value) {
      return value;
    }
  }

  return null;
}

async function forwardSubscriptionUpdate(args: {
  orgId: string;
  planTier: PlanTier;
  subscriptionStatus: "trialing" | "active" | "past_due" | "canceled";
  stripeSubscriptionId: string;
}) {
  const internalHeaders = new Headers();
  internalHeaders.set("Content-Type", "application/json");
  internalHeaders.set("Accept", "application/json");
  internalHeaders.set("X-User-Id", "stripe-webhook");
  internalHeaders.set("X-Org-Id", args.orgId);
  internalHeaders.set("X-Org-Role", "admin");

  const internalApiKey = process.env.INTERNAL_API_KEY;
  if (internalApiKey) {
    internalHeaders.set("X-Internal-Api-Key", internalApiKey);
  }

  const response = await fetch(
    `${GO_API_BASE_URL}/api/v1/workspaces/${encodeURIComponent(args.orgId)}/subscription`,
    {
      method: "PATCH",
      headers: internalHeaders,
      body: JSON.stringify({
        plan_tier: args.planTier,
        subscription_status: args.subscriptionStatus,
        stripe_subscription_id: args.stripeSubscriptionId,
      }),
      cache: "no-store",
    },
  );

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`workspace_subscription_update_failed: ${response.status} ${body}`);
  }
}

async function forwardInvoicePaid(args: {
  orgId: string;
  invoiceId: string;
  stripeSubscriptionId?: string;
}) {
  const internalHeaders = new Headers();
  internalHeaders.set("Content-Type", "application/json");
  internalHeaders.set("Accept", "application/json");
  internalHeaders.set("X-User-Id", "stripe-webhook");
  internalHeaders.set("X-Org-Id", args.orgId);
  internalHeaders.set("X-Org-Role", "admin");

  const internalApiKey = process.env.INTERNAL_API_KEY;
  if (internalApiKey) {
    internalHeaders.set("X-Internal-Api-Key", internalApiKey);
  }

  const response = await fetch(
    `${GO_API_BASE_URL}/api/v1/workspaces/${encodeURIComponent(args.orgId)}/usage/reset`,
    {
      method: "PATCH",
      headers: internalHeaders,
      body: JSON.stringify({
        invoice_id: args.invoiceId,
        stripe_subscription_id: args.stripeSubscriptionId,
      }),
      cache: "no-store",
    },
  );

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`workspace_usage_reset_failed: ${response.status} ${body}`);
  }
}

function getStripeClient() {
  const secretKey = process.env.STRIPE_SECRET_KEY;
  if (!secretKey) {
    throw new Error("STRIPE_SECRET_KEY is not configured");
  }

  return new Stripe(secretKey, {
    apiVersion: "2025-02-24.acacia",
  });
}

export async function POST(req: Request) {
  let stripe: Stripe;

  try {
    stripe = getStripeClient();
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: 500 });
  }

  const body = await req.text();
  const headersList = await headers();
  const signature = headersList.get("stripe-signature");
  const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET;

  if (!signature) {
    return new Response("Missing stripe-signature header", { status: 400 });
  }
  if (!webhookSecret) {
    return new Response("STRIPE_WEBHOOK_SECRET is not configured", { status: 500 });
  }

  let event: Stripe.Event;

  try {
    event = stripe.webhooks.constructEvent(body, signature, webhookSecret);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(`Webhook signature verification failed: ${message}`, { status: 400 });
  }

  switch (event.type) {
    case "customer.subscription.updated":
    case "customer.subscription.deleted": {
      const subscription = event.data.object as Stripe.Subscription;
      const orgId = deriveOrgId(subscription);
      const planTier = derivePlanTier(subscription) ?? "starter";
      const subscriptionStatus =
        event.type === "customer.subscription.deleted"
          ? "canceled"
          : (subscription.status as "trialing" | "active" | "past_due" | "canceled");

      if (!orgId) {
        console.warn(`Missing org_id metadata for subscription ${subscription.id}`);
        break;
      }

      await forwardSubscriptionUpdate({
        orgId,
        planTier,
        subscriptionStatus,
        stripeSubscriptionId: subscription.id,
      });
      break;
    }
    case "invoice.paid": {
      const invoice = event.data.object as Stripe.Invoice;
      const orgId = deriveOrgIdFromInvoice(invoice);
      const stripeSubscriptionId =
        typeof invoice.subscription === "string" ? invoice.subscription : invoice.subscription?.id;

      if (!orgId) {
        console.warn(`Missing org_id metadata for invoice ${invoice.id}`);
        break;
      }

      await forwardInvoicePaid({
        orgId,
        invoiceId: invoice.id,
        stripeSubscriptionId: stripeSubscriptionId ?? undefined,
      });
      break;
    }
    default:
      // Unhandled event type — ignore
      break;
  }

  return new Response(JSON.stringify({ received: true }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
