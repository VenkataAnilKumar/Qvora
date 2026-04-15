import { Webhook } from "svix";

export async function POST(req: Request) {
  const body = await req.text();
  const svix_id = req.headers.get("svix-id");
  const svix_timestamp = req.headers.get("svix-timestamp");
  const svix_signature = req.headers.get("svix-signature");

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Missing svix headers", { status: 400 });
  }

  const wh = new Webhook(process.env.CLERK_WEBHOOK_SECRET!);
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
      // TODO: forward to Go API to create workspace record
      console.log("Org created:", event.data.id);
      break;
    }
    case "organization.deleted": {
      // TODO: forward to Go API to soft-delete workspace
      console.log("Org deleted:", event.data.id);
      break;
    }
    case "organizationMembership.created":
    case "organizationMembership.deleted": {
      // TODO: forward to Go API to sync team members
      console.log(`Membership ${event.type}:`, event.data.id);
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
