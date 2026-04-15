// Qvora shared types — add domain types here as features are built

// =============================================================================
// Auth / Workspace
// =============================================================================

export type OrgRole = "admin" | "member" | "viewer";

export interface WorkspaceContext {
  orgId: string;
  orgSlug: string;
  role: OrgRole;
  planTier: PlanTier;
}

// =============================================================================
// Billing
// =============================================================================

export type PlanTier = "starter" | "growth" | "agency";
export type SubscriptionStatus = "trialing" | "active" | "past_due" | "canceled";

export interface PlanLimits {
  maxVariantsPerAngle: number | null; // null = unlimited (agency)
}

export const PLAN_LIMITS: Record<PlanTier, PlanLimits> = {
  starter: { maxVariantsPerAngle: 3 },
  growth: { maxVariantsPerAngle: 10 },
  agency: { maxVariantsPerAngle: null },
};

// =============================================================================
// URL Ingestion
// =============================================================================

export interface ScrapedProduct {
  url: string;
  title: string;
  description: string;
  images: string[];
  price?: string;
  features: string[];
  rawHtml?: string;
  scrapedAt: string; // ISO 8601
}

// =============================================================================
// Creative Brief
// =============================================================================

export type AdAngle =
  | "problem_solution"
  | "social_proof"
  | "transformation"
  | "urgency"
  | "education";

export interface CreativeBrief {
  id: string;
  workspaceId: string;
  productUrl: string;
  angles: AngleBrief[];
  createdAt: string;
}

export interface AngleBrief {
  angle: AdAngle;
  headline: string;
  hook: string;
  script: string;
  cta: string;
  voiceTone: string;
}

// =============================================================================
// Video Generation
// =============================================================================

export type GenerationStatus =
  | "queued"
  | "scraping"
  | "briefing"
  | "generating"
  | "postprocessing"
  | "complete"
  | "failed";

export type VideoModel = "veo3" | "kling3" | "runway4" | "sora2";

export interface GenerationJob {
  jobId: string;
  workspaceId: string;
  productUrl: string;
  status: GenerationStatus;
  model: VideoModel;
  variants: VideoVariant[];
  createdAt: string;
  updatedAt: string;
}

export interface VideoVariant {
  id: string;
  jobId: string;
  angle: AdAngle;
  muxAssetId?: string;
  muxPlaybackId?: string;
  r2Key?: string;
  durationSeconds?: number;
  status: GenerationStatus;
}

// =============================================================================
// SSE Generation Stream Events
// =============================================================================

export type SSEEventType =
  | "job_queued"
  | "scraping_started"
  | "scraping_complete"
  | "brief_started"
  | "brief_complete"
  | "generation_started"
  | "generation_progress"
  | "postprocessing_started"
  | "variant_complete"
  | "job_complete"
  | "job_failed";

export interface SSEEvent<T = unknown> {
  type: SSEEventType;
  jobId: string;
  data: T;
  timestamp: string;
}

// =============================================================================
// Brand Kit
// =============================================================================

export interface BrandKit {
  id: string;
  workspaceId: string;
  name: string;
  logoR2Key?: string;
  primaryColor: string;
  secondaryColor?: string;
  fontFamily?: string;
  watermarkEnabled: boolean;
  createdAt: string;
}

// =============================================================================
// API response envelope
// =============================================================================

export interface ApiSuccess<T> {
  ok: true;
  data: T;
}

export interface ApiError {
  ok: false;
  error: string;
  code?: string;
}

export type ApiResponse<T> = ApiSuccess<T> | ApiError;
