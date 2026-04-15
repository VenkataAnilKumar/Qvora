import { z } from "zod";
import { generateObject } from "ai";
import { openai } from "@ai-sdk/openai";
import { TRPCError } from "@trpc/server";
import {
  anglesGenerationSchema,
  buildAnglesGenerationPrompt,
} from "@/ai/prompts/angles-gen.prompt";
import { createTRPCRouter, workspaceProcedure } from "../init";

type ApiBrief = {
  brief_id: string;
  scrape_job_id: string;
  org_id: string;
  product_url: string;
  template?: string;
  status: string;
  created_at?: string;
};

type ScrapePayload = {
  name?: string;
  category?: string;
  price?: string;
  features?: string[];
  proof_points?: string[];
  image_urls?: string[];
  description?: string;
  confidence?: number;
};

const briefModelSchema = z.enum(["gpt-4o", "gpt-4.1-mini"]);

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

export const briefsRouter = createTRPCRouter({
  create: workspaceProcedure
    .input(
      z.object({
        productUrl: z.string().url(),
        template: z.string().optional(),
        model: briefModelSchema.default("gpt-4o"),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const openAIKey = process.env.OPENAI_API_KEY;
      if (!openAIKey) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "OPENAI_API_KEY is required for brief generation",
        });
      }

      const scrapeEndpoint = process.env.MODAL_SCRAPER_ENDPOINT;
      if (!scrapeEndpoint) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "MODAL_SCRAPER_ENDPOINT is required for brief generation",
        });
      }

      const scrapeResponse = await fetch(scrapeEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          product_url: input.productUrl,
          workspace_id: ctx.orgId,
          user_id: ctx.userId,
        }),
        cache: "no-store",
      });

      if (!scrapeResponse.ok) {
        throw new TRPCError({
          code: "BAD_REQUEST",
          message: "scrape_failed",
        });
      }

      const scraped = (await scrapeResponse.json()) as ScrapePayload;

      let object;
      try {
        const result = await generateObject({
          model: openai(input.model),
          schema: anglesGenerationSchema,
          prompt: buildAnglesGenerationPrompt({
            productUrl: input.productUrl,
            template: input.template,
            scraped,
          }),
        });
        object = result.object;
      } catch {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "brief_generation_failed",
        });
      }

      const createdAt = new Date().toISOString();
      const briefId = crypto.randomUUID();

      return {
        briefId,
        scrapeJobId: briefId,
        orgId: ctx.orgId,
        status: "generated",
        productUrl: input.productUrl,
        model: input.model,
        createdAt,
        angles: object.angles,
        hooks: object.hooks,
      };
    }),

  list: workspaceProcedure.query(async ({ ctx }) => {
    const response = await fetch(`${GO_API_BASE_URL}/api/v1/briefs`, {
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
      briefs: ApiBrief[];
    };

    return {
      orgId: payload.org_id,
      briefs: payload.briefs.map((brief) => ({
        briefId: brief.brief_id,
        scrapeJobId: brief.scrape_job_id,
        productUrl: brief.product_url,
        template: brief.template,
        status: brief.status,
        createdAt: brief.created_at ?? new Date().toISOString(),
      })),
    };
  }),
});
