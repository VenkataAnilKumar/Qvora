import { z } from "zod";

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

type ScrapedPayload = {
  name?: string;
  category?: string;
  price?: string;
  features?: string[];
  proof_points?: string[];
  image_urls?: string[];
  description?: string;
  confidence?: number;
};

export type AnglesGenerationOutput = z.infer<typeof anglesGenerationSchema>;

export function buildAnglesGenerationPrompt(input: {
  productUrl: string;
  template?: string;
  scraped: ScrapedPayload;
}): string {
  const features = (input.scraped.features ?? []).slice(0, 10).join("; ");
  const proofPoints = (input.scraped.proof_points ?? []).slice(0, 8).join("; ");

  return [
    "You are Qvora, a performance creative strategist for agency teams.",
    "Generate high-conversion short-form video ad angles for 9:16 placements.",
    "Return practical output for paid social testing, not generic brand copy.",
    "",
    `Product URL: ${input.productUrl}`,
    `Template hint: ${input.template ?? "none"}`,
    `Name: ${input.scraped.name ?? "unknown"}`,
    `Category: ${input.scraped.category ?? "unknown"}`,
    `Price: ${input.scraped.price ?? "unknown"}`,
    `Description: ${input.scraped.description ?? "none"}`,
    `Top features: ${features || "none"}`,
    `Proof points: ${proofPoints || "none"}`,
    `Scrape confidence: ${String(input.scraped.confidence ?? "unknown")}`,
    "",
    "Constraints:",
    "1) Produce 3-5 unique angles, each with a different angle type when possible.",
    "2) Keep headlines punchy and hooks optimized for the first 3 seconds.",
    "3) Scripts should sound natural for spoken delivery in ~15 seconds.",
    "4) CTA must be direct and conversion-oriented.",
    "5) Prefer agency performance language over fluffy branding.",
  ].join("\n");
}
