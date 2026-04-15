import { z } from "zod";

// ---------------------------------------------------------------------------
// Product Extraction Schema — GPT-4o structures raw scraped data
// ---------------------------------------------------------------------------

export const productExtractionSchema = z.object({
  name: z.string().min(1).max(200),
  category: z.string().min(1).max(100),
  price: z.string().max(50).optional(),
  headline: z.string().max(200).optional(),
  description: z.string().max(1000).optional(),
  features: z.array(z.string().max(200)).max(10),
  proofPoints: z.array(z.string().max(200)).max(8),
  primaryCta: z.string().max(80).optional(),
  targetAudience: z.string().max(200).optional(),
});

export type ProductExtraction = z.infer<typeof productExtractionSchema>;

type RawScrapePayload = {
  name?: string;
  category?: string;
  price?: string;
  features?: string[];
  proof_points?: string[];
  image_urls?: string[];
  description?: string;
  confidence?: number;
};

export function buildProductExtractionPrompt(input: {
  productUrl: string;
  scraped: RawScrapePayload;
}): string {
  return [
    "You are a product analyst. Extract structured product information from the scraped data below.",
    "Fill in missing fields with your best inference from the available data.",
    "Be specific and accurate — this output will drive AI-generated ad creative.",
    "",
    `Product URL: ${input.productUrl}`,
    `Raw name: ${input.scraped.name ?? "unknown"}`,
    `Raw category: ${input.scraped.category ?? "unknown"}`,
    `Raw price: ${input.scraped.price ?? "unknown"}`,
    `Raw description: ${input.scraped.description ?? "none"}`,
    `Raw features: ${(input.scraped.features ?? []).join("; ") || "none"}`,
    `Raw proof points: ${(input.scraped.proof_points ?? []).join("; ") || "none"}`,
    `Scrape confidence: ${String(input.scraped.confidence ?? "unknown")}`,
  ].join("\n");
}

// ---------------------------------------------------------------------------
// Angles Generation Schema — Claude Sonnet 4.6 generates creative strategy
// ---------------------------------------------------------------------------

const angleSchema = z.object({
  angle: z.enum(["problem_solution", "social_proof", "transformation", "urgency", "education"]),
  headline: z.string().min(6).max(120),
  hook: z.string().min(8).max(180),
  script: z.string().min(40).max(800),
  cta: z.string().min(3).max(60),
  voiceTone: z.enum(["energetic", "calm", "authoritative", "conversational", "urgent"]),
});

export const anglesGenerationSchema = z.object({
  angles: z.array(angleSchema).min(3).max(5),
  hooks: z.array(z.string().min(8).max(140)).min(3).max(6),
});

export type AnglesGenerationOutput = z.infer<typeof anglesGenerationSchema>;

export function buildAnglesGenerationPrompt(input: {
  productUrl: string;
  template?: string;
  product: ProductExtraction;
}): string {
  const features = (input.product.features ?? []).slice(0, 10).join("; ");
  const proofPoints = (input.product.proofPoints ?? []).slice(0, 8).join("; ");

  return [
    "You are Qvora, a performance creative strategist for agency teams.",
    "Generate high-conversion short-form video ad angles for 9:16 paid social placements.",
    "Return practical output for performance testing, not generic brand copy.",
    "",
    `Product URL: ${input.productUrl}`,
    `Template hint: ${input.template ?? "none"}`,
    `Name: ${input.product.name}`,
    `Category: ${input.product.category}`,
    `Price: ${input.product.price ?? "unknown"}`,
    `Description: ${input.product.description ?? "none"}`,
    `Top features: ${features || "none"}`,
    `Proof points: ${proofPoints || "none"}`,
    `Primary CTA: ${input.product.primaryCta ?? "none"}`,
    `Target audience: ${input.product.targetAudience ?? "unknown"}`,
    "",
    "Constraints:",
    "1) Produce 3-5 unique angles, each with a different angle type when possible.",
    "2) Keep headlines punchy and hooks optimized for the first 3 seconds.",
    "3) Scripts should sound natural for spoken delivery in ~15 seconds.",
    "4) CTA must be direct and conversion-oriented.",
    "5) Prefer agency performance language over fluffy branding.",
  ].join("\n");
}
