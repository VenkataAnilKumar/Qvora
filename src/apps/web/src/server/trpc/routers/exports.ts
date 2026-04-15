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
});
