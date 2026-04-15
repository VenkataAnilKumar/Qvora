import { Webhook } from "svix";

const GO_API_BASE_URL = process.env.GO_API_URL ?? "http://localhost:8080";

const readString = (record: Record<string, unknown>, key: string): string | null => {
  const value = record[key];
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : null;
};

const readNestedString = (record: Record<string, unknown>, path: string[]): string | null => {
  let current: unknown = record;
  for (const segment of path) {
    if (typeof current !== "object" || current === null) {
      return null;
    }

    current = (current as Record<string, unknown>)[segment];
  }

  if (typeof current !== "string") {
    return null;
  }

  const trimmed = current.trim();
  return trimmed.length > 0 ? trimmed : null;
};

async function forwardWorkspaceLifecycle(args: {
  orgId: string;
  event: "organization.created" | "organization.deleted";
}) {
  const headers = new Headers();
  headers.set("Content-Type", "application/json");
  headers.set("Accept", "application/json");
  headers.set("X-User-Id", "clerk-webhook");
  headers.set("X-Org-Id", args.orgId);
  headers.set("X-Org-Role", "admin");

  const internalApiKey = process.env.INTERNAL_API_KEY;
  if (internalApiKey) {
    headers.set("X-Internal-Api-Key", internalApiKey);
  }

  const response = await fetch(
    `${GO_API_BASE_URL}/api/v1/workspaces/${encodeURIComponent(args.orgId)}/lifecycle`,
    {
      method: "PATCH",
      headers,
      body: JSON.stringify({ event: args.event }),
      cache: "no-store",
    },
  );

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`workspace_lifecycle_forward_failed: ${response.status} ${body}`);
  }
}

async function forwardWorkspaceMembership(args: {
  orgId: string;
  event: "organizationMembership.created" | "organizationMembership.deleted";
  membershipId: string;
  userId?: string;
  role?: string;
}) {
  const headers = new Headers();
  headers.set("Content-Type", "application/json");
  headers.set("Accept", "application/json");
  headers.set("X-User-Id", "clerk-webhook");
  headers.set("X-Org-Id", args.orgId);
  headers.set("X-Org-Role", "admin");

  const internalApiKey = process.env.INTERNAL_API_KEY;
  if (internalApiKey) {
    headers.set("X-Internal-Api-Key", internalApiKey);
  }

  const response = await fetch(
    `${GO_API_BASE_URL}/api/v1/workspaces/${encodeURIComponent(args.orgId)}/memberships/sync`,
    {
      method: "PATCH",
      headers,
      body: JSON.stringify({
        event: args.event,
        membership_id: args.membershipId,
        user_id: args.userId,
        role: args.role,
      }),
      cache: "no-store",
    },
  );

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`workspace_membership_forward_failed: ${response.status} ${body}`);
  }
}

export async function POST(req: Request) {
  const body = await req.text();
  const svix_id = req.headers.get("svix-id");
  const svix_timestamp = req.headers.get("svix-timestamp");
  const svix_signature = req.headers.get("svix-signature");
  const clerkWebhookSecret = process.env.CLERK_WEBHOOK_SECRET;

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Missing svix headers", { status: 400 });
  }
  if (!clerkWebhookSecret) {
    return new Response("CLERK_WEBHOOK_SECRET is not configured", { status: 500 });
  }

  const wh = new Webhook(clerkWebhookSecret);
  let event: { type: string; data: Record<string, unknown> };

  try {
    event = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as typeof event;
  } catch {
    return new Response("Webhook verification failed", { status: 400 });
  }

  switch (event.type) {
    case "organization.created": {
      const orgId = readString(event.data, "id");
      if (!orgId) {
        console.warn("organization.created missing org id");
        break;
      }
      await forwardWorkspaceLifecycle({ orgId, event: "organization.created" });
      break;
    }
    case "organization.deleted": {
      const orgId = readString(event.data, "id");
      if (!orgId) {
        console.warn("organization.deleted missing org id");
        break;
      }
      await forwardWorkspaceLifecycle({ orgId, event: "organization.deleted" });
      break;
    }
    case "organizationMembership.created":
    case "organizationMembership.deleted": {
      const orgId =
        readString(event.data, "organization_id") ??
        readNestedString(event.data, ["organization", "id"]);
      const membershipId = readString(event.data, "id");
      const userId =
        readString(event.data, "public_user_data.user_id") ??
        readNestedString(event.data, ["public_user_data", "user_id"]) ??
        readString(event.data, "public_user_data.userId") ??
        readNestedString(event.data, ["public_user_data", "userId"]) ??
        readString(event.data, "user_id") ??
        readNestedString(event.data, ["user", "id"]);
      const role = readString(event.data, "role");

      if (!orgId || !membershipId) {
        console.warn(`${event.type} missing org id or membership id`);
        break;
      }

      await forwardWorkspaceMembership({
        orgId,
        event: event.type,
        membershipId,
        userId: userId ?? undefined,
        role: role ?? undefined,
      });
      break;
    }
    default:
      break;
  }

  return new Response(JSON.stringify({ received: true }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}
