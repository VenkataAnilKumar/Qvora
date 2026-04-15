import { TRPCError } from "@trpc/server";
import { createTRPCRouter, publicProcedure } from "../init";

const GO_API_BASE_URL = process.env.GO_API_URL ?? "http://localhost:8080";

export const systemRouter = createTRPCRouter({
  hello: publicProcedure.query(async () => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/health`, {
      method: "GET",
      cache: "no-store",
    });

    if (!response.ok) {
      throw new TRPCError({
        code: "INTERNAL_SERVER_ERROR",
        message: "go_api_unreachable",
      });
    }

    const health = (await response.json()) as { status?: string; ok?: boolean };
    const goApiStatus = health.status ?? (health.ok ? "ok" : "unknown");

    return {
      message: "hello from trpc",
      goApi: goApiStatus,
      timestamp: new Date().toISOString(),
    };
  }),
});
