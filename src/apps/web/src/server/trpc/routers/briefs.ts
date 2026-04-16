import { z } from "zod";
import { generateObject } from "ai";
import { openai } from "@ai-sdk/openai";
import { anthropic } from "@ai-sdk/anthropic";
import { TRPCError } from "@trpc/server";
import {
  anglesGenerationSchema,
  buildAnglesGenerationPrompt,
  productExtractionSchema,
  buildProductExtractionPrompt,
} from "@qvora/prompts/angles-gen.prompt";
import { createTRPCRouter, workspaceProcedure } from "../init";
import { recordBriefTrace } from "../langfuse";

const singleAngleSchema = z.object({
  headline: z.string().min(3),
  script: z.string().min(10),
  cta: z.string().min(2),
  voiceTone: z.string().min(2),
});

const singleHookSchema = z.object({
  hook: z.string().min(3).max(140),
});

type ApiBrief = {
  brief_id: string;
  scrape_job_id?: string;
  org_id: string;
  product_url: string;
  model?: string;
  template?: string;
  status: string;
  created_at?: string;
};

type ApiBriefAngle = {
  angle: string;
  headline: string;
  script: string;
  cta: string;
  voice_tone?: string | null;
};

type ApiBriefDetail = {
  brief_id: string;
  org_id: string;
  product_url: string;
  model: string;
  status: string;
  angles: ApiBriefAngle[];
  hooks: string[];
  created_at?: string;
  updated_at?: string;
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

type InternalHeadersCtx = {
  userId: string;
  orgId: string;
  orgRole: string | null | undefined;
  headers: Headers;
};

const buildInternalHeaders = (ctx: InternalHeadersCtx) => {
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

const fetchBriefDetailFromApi = async (briefId: string, ctx: InternalHeadersCtx) => {
  const response = await fetch(`${GO_API_BASE_URL}/api/v1/briefs/${encodeURIComponent(briefId)}`, {
    method: "GET",
    headers: buildInternalHeaders(ctx),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new TRPCError({
      code: response.status === 404 ? "NOT_FOUND" : "INTERNAL_SERVER_ERROR",
      message: await parseApiError(response),
    });
  }

  return (await response.json()) as ApiBriefDetail;
};

const persistBriefContentToApi = async (
  briefId: string,
  payload: {
    angles: Array<{
      angle: string;
      headline: string;
      script: string;
      cta: string;
      voice_tone?: string;
    }>;
    hooks: string[];
  },
  ctx: InternalHeadersCtx,
) => {
  const response = await fetch(
    `${GO_API_BASE_URL}/api/v1/briefs/${encodeURIComponent(briefId)}/content`,
    {
      method: "PUT",
      headers: buildInternalHeaders(ctx),
      body: JSON.stringify(payload),
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
    brief_id: string;
    updated_at: string;
    angle_count: number;
    hook_count: number;
  };
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
      if (!process.env.OPENAI_API_KEY) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "OPENAI_API_KEY is required for brief generation",
        });
      }
      if (!process.env.ANTHROPIC_API_KEY) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "ANTHROPIC_API_KEY is required for brief generation",
        });
      }
      const scrapeEndpoint = process.env.MODAL_SCRAPER_ENDPOINT;
      if (!scrapeEndpoint) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "MODAL_SCRAPER_ENDPOINT is required for brief generation",
        });
      }

      // Step 1: Scrape product URL via Modal Playwright
      const scrapeResponse = await fetch(scrapeEndpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          product_url: input.productUrl,
          workspace_id: ctx.orgId,
          user_id: ctx.userId,
        }),
        cache: "no-store",
      });
      if (!scrapeResponse.ok) {
        throw new TRPCError({ code: "BAD_REQUEST", message: "scrape_failed" });
      }
      const scraped = (await scrapeResponse.json()) as ScrapePayload;

      // Step 2: GPT-4o — structure raw scraped data into clean product summary
      let product: z.infer<typeof productExtractionSchema>;
      try {
        const result = await generateObject({
          model: openai("gpt-4o"),
          schema: productExtractionSchema,
          prompt: buildProductExtractionPrompt({ productUrl: input.productUrl, scraped }),
        });
        product = result.object;
      } catch {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "product_extraction_failed",
        });
      }

      // Step 3: Claude Sonnet 4.6 — generate creative angles and hooks from product
      let brief: z.infer<typeof anglesGenerationSchema>;
      try {
        const result = await generateObject({
          model: anthropic("claude-sonnet-4-6"),
          schema: anglesGenerationSchema,
          prompt: buildAnglesGenerationPrompt({
            productUrl: input.productUrl,
            template: input.template,
            product,
          }),
        });
        brief = result.object;
      } catch {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "brief_generation_failed",
        });
      }

      // Step 4: Persist brief + angles + hooks to Go API
      const persistResponse = await fetch(`${GO_API_BASE_URL}/api/v1/briefs`, {
        method: "POST",
        headers: buildInternalHeaders({
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        }),
        body: JSON.stringify({
          product_url: input.productUrl,
          template: input.template ?? null,
          model: input.model,
          angles: brief.angles,
          hooks: brief.hooks,
        }),
        cache: "no-store",
      });
      if (!persistResponse.ok) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: await parseApiError(persistResponse),
        });
      }

      const persisted = (await persistResponse.json()) as {
        brief_id: string;
        created_at: string;
        status: string;
      };

      await recordBriefTrace({
        traceName: "brief.create",
        userId: ctx.userId,
        orgId: ctx.orgId,
        briefId: persisted.brief_id,
        model: input.model,
        input: { productUrl: input.productUrl, template: input.template, scraped },
        output: { product, brief },
        metadata: {
          angleCount: brief.angles.length,
          hookCount: brief.hooks.length,
        },
      });

      return {
        briefId: persisted.brief_id,
        scrapeJobId: persisted.brief_id,
        orgId: ctx.orgId,
        status: persisted.status,
        productUrl: input.productUrl,
        model: input.model,
        createdAt: persisted.created_at,
        product,
        angles: brief.angles,
        hooks: brief.hooks,
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

  get: workspaceProcedure
    .input(
      z.object({
        briefId: z.string().uuid(),
      }),
    )
    .query(async ({ input, ctx }) => {
      const detail = await fetchBriefDetailFromApi(input.briefId, {
        userId: ctx.userId,
        orgId: ctx.orgId,
        orgRole: ctx.orgRole,
        headers: ctx.headers,
      });

      return {
        briefId: detail.brief_id,
        orgId: detail.org_id,
        productUrl: detail.product_url,
        model: detail.model,
        status: detail.status,
        angles: detail.angles.map((angle) => ({
          angle: angle.angle,
          headline: angle.headline,
          script: angle.script,
          cta: angle.cta,
          voiceTone: angle.voice_tone ?? "",
        })),
        hooks: detail.hooks,
        createdAt: detail.created_at ?? new Date().toISOString(),
        updatedAt: detail.updated_at ?? new Date().toISOString(),
      };
    }),

  updateContent: workspaceProcedure
    .input(
      z.object({
        briefId: z.string().uuid(),
        angles: z
          .array(
            z.object({
              angle: z.string().min(1),
              headline: z.string().min(1),
              script: z.string().min(1),
              cta: z.string().min(1),
              voiceTone: z.string().optional(),
            }),
          )
          .min(1),
        hooks: z.array(z.string().min(1)).min(1),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const updated = await persistBriefContentToApi(
        input.briefId,
        {
          angles: input.angles.map((angle) => ({
            angle: angle.angle,
            headline: angle.headline,
            script: angle.script,
            cta: angle.cta,
            voice_tone: angle.voiceTone,
          })),
          hooks: input.hooks,
        },
        {
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        },
      );

      await recordBriefTrace({
        traceName: "brief.updateContent",
        userId: ctx.userId,
        orgId: ctx.orgId,
        briefId: input.briefId,
        model: "manual-edit",
        input: { angles: input.angles, hooks: input.hooks },
        output: updated,
        metadata: {
          angleCount: input.angles.length,
          hookCount: input.hooks.length,
        },
      });

      return {
        briefId: updated.brief_id,
        updatedAt: updated.updated_at,
        angleCount: updated.angle_count,
        hookCount: updated.hook_count,
      };
    }),

  regenerateAngle: workspaceProcedure
    .input(
      z.object({
        briefId: z.string().uuid(),
        angleIndex: z.number().min(0),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const started = Date.now();
      const detail = await fetchBriefDetailFromApi(input.briefId, {
        userId: ctx.userId,
        orgId: ctx.orgId,
        orgRole: ctx.orgRole,
        headers: ctx.headers,
      });

      if (input.angleIndex >= detail.angles.length) {
        throw new TRPCError({ code: "BAD_REQUEST", message: "invalid_angle_index" });
      }

      const current = detail.angles[input.angleIndex];
      if (!current) {
        throw new TRPCError({ code: "BAD_REQUEST", message: "invalid_angle_index" });
      }
      const regenerated = await generateObject({
        model: openai("gpt-4.1-mini"),
        schema: singleAngleSchema,
        prompt: [
          "You are a performance creative strategist for short-form paid social ads.",
          `Regenerate only this angle while preserving the angle type: ${current.angle}`,
          `Product URL: ${detail.product_url}`,
          `Current angle content: ${JSON.stringify(current)}`,
          "Return concise, punchy copy suitable for 15-30 second ad scripts.",
        ].join("\n"),
      });

      const nextAngles = detail.angles.map((angle, index) => {
        if (index !== input.angleIndex) {
          return {
            angle: angle.angle,
            headline: angle.headline,
            script: angle.script,
            cta: angle.cta,
            voiceTone: angle.voice_tone ?? "",
          };
        }

        return {
          angle: angle.angle,
          headline: regenerated.object.headline,
          script: regenerated.object.script,
          cta: regenerated.object.cta,
          voiceTone: regenerated.object.voiceTone,
        };
      });

      await persistBriefContentToApi(
        input.briefId,
        {
          angles: nextAngles.map((angle) => ({
            angle: angle.angle,
            headline: angle.headline,
            script: angle.script,
            cta: angle.cta,
            voice_tone: angle.voiceTone,
          })),
          hooks: detail.hooks,
        },
        {
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        },
      );

      const elapsedMs = Date.now() - started;
      await recordBriefTrace({
        traceName: "brief.regenerateAngle",
        userId: ctx.userId,
        orgId: ctx.orgId,
        briefId: input.briefId,
        model: "gpt-4.1-mini",
        input: { angleIndex: input.angleIndex, current },
        output: regenerated.object,
        metadata: { elapsedMs },
      });

      const regeneratedAngle = nextAngles[input.angleIndex];
      if (!regeneratedAngle) {
        throw new TRPCError({
          code: "INTERNAL_SERVER_ERROR",
          message: "regenerated_angle_missing",
        });
      }

      return {
        angleIndex: input.angleIndex,
        angle: regeneratedAngle,
        elapsedMs,
        under10Seconds: elapsedMs < 10_000,
      };
    }),

  regenerateHook: workspaceProcedure
    .input(
      z.object({
        briefId: z.string().uuid(),
        hookIndex: z.number().min(0),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const started = Date.now();
      const detail = await fetchBriefDetailFromApi(input.briefId, {
        userId: ctx.userId,
        orgId: ctx.orgId,
        orgRole: ctx.orgRole,
        headers: ctx.headers,
      });

      if (input.hookIndex >= detail.hooks.length) {
        throw new TRPCError({ code: "BAD_REQUEST", message: "invalid_hook_index" });
      }

      const currentHook = detail.hooks[input.hookIndex];
      const regenerated = await generateObject({
        model: openai("gpt-4.1-mini"),
        schema: singleHookSchema,
        prompt: [
          "Generate one high-performing ad hook line for short-form paid social.",
          `Product URL: ${detail.product_url}`,
          `Current hook: ${currentHook}`,
          "Keep it concise and scroll-stopping.",
        ].join("\n"),
      });

      const nextHooks = detail.hooks.map((hook, index) =>
        index === input.hookIndex ? regenerated.object.hook : hook,
      );

      await persistBriefContentToApi(
        input.briefId,
        {
          angles: detail.angles.map((angle) => ({
            angle: angle.angle,
            headline: angle.headline,
            script: angle.script,
            cta: angle.cta,
            voice_tone: angle.voice_tone ?? undefined,
          })),
          hooks: nextHooks,
        },
        {
          userId: ctx.userId,
          orgId: ctx.orgId,
          orgRole: ctx.orgRole,
          headers: ctx.headers,
        },
      );

      const elapsedMs = Date.now() - started;
      await recordBriefTrace({
        traceName: "brief.regenerateHook",
        userId: ctx.userId,
        orgId: ctx.orgId,
        briefId: input.briefId,
        model: "gpt-4.1-mini",
        input: { hookIndex: input.hookIndex, currentHook },
        output: regenerated.object,
        metadata: { elapsedMs },
      });

      return {
        hookIndex: input.hookIndex,
        hook: regenerated.object.hook,
        hooks: nextHooks,
        elapsedMs,
        under10Seconds: elapsedMs < 10_000,
      };
    }),

  batchGenerate: workspaceProcedure
    .input(
      z.object({
        briefId: z.string().uuid(),
        variantsPerSpec: z.number().min(1).max(10).default(1),
        specs: z
          .array(
            z.object({
              angle: z.string().min(1),
              hook: z.string().optional(),
              model: z.enum(["veo3", "kling3", "runway4", "sora2"]).default("veo3"),
            }),
          )
          .min(1)
          .max(50),
      }),
    )
    .mutation(async ({ input, ctx }) => {
      const response = await fetch(
        `${GO_API_BASE_URL}/api/v1/briefs/${encodeURIComponent(input.briefId)}/batch-generate`,
        {
          method: "POST",
          headers: buildInternalHeaders({
            userId: ctx.userId,
            orgId: ctx.orgId,
            orgRole: ctx.orgRole,
            headers: ctx.headers,
          }),
          body: JSON.stringify({
            variants_per_spec: input.variantsPerSpec,
            specs: input.specs.map((spec) => ({
              angle: spec.angle,
              hook: spec.hook,
              model: spec.model,
            })),
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
                : response.status === 402
                  ? "PAYMENT_REQUIRED"
                  : "INTERNAL_SERVER_ERROR",
          message: await parseApiError(response),
        });
      }

      return (await response.json()) as {
        org_id: string;
        workspace_id: string;
        brief_id: string;
        plan_tier: "starter" | "growth" | "agency";
        total_requested: number;
        approved_per_spec: number;
        specs_count: number;
        jobs: Array<{
          job_id: string;
          angle: string;
          hook: string;
          model: "veo3" | "kling3" | "runway4" | "sora2";
          variants_per_spec: number;
          status: string;
        }>;
        message: string;
      };
    }),
});
