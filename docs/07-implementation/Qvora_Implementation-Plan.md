# Qvora — Implementation Plan
**Version:** 1.0 | **Date:** April 2026 | **Status:** Active

> This document is the engineering execution plan for Qvora V1. It translates the Feature Spec, User Stories, and Architecture Stack into time-boxed phases with explicit deliverables, dependencies, and acceptance gates. Read `.github/CONTEXT.md` and `docs/06-technical/Qvora_Architecture-Stack.md` before this document.

---

## Plan Overview

| Phase | Name | Duration | Focus |
|---|---|---|---|
| **Phase 0** | Foundation & Infrastructure | 1 week | Repo, CI/CD, infra, auth plumbing |
| **Phase 1** | Core Data Layer | 1 week | DB schema, sqlc, RLS, API skeleton |
| **Phase 2** | URL Ingestion & Brief Engine | 2 weeks | Playwright scrape + GPT-4o brief generation |
| **Phase 3** | Video Generation Pipeline | 3 weeks | FAL.AI + ElevenLabs + HeyGen + asynq workers |
| **Phase 4** | Brand Kit & Export | 1 week | Brand kit system + Rust postprocessor + Mux |
| **Phase 5** | Asset Library & Team | 1 week | Library UI, collaboration, roles |
| **Phase 6** | Platform, Billing & Trial | 1 week | Stripe, tier enforcement, trial flow |
| **Phase 7** | Polish, Observability & Launch | 1 week | Sentry, PostHog, Langfuse, QA, deploy |

**Total V1 duration: 11 weeks**
**V2 scope (Signal + Ad Connector): planned post-launch**

---

## Phase 0 — Foundation & Infrastructure
**Duration:** Week 1  
**Goal:** Every engineer can run the full stack locally. CI is green. Prod environment exists.

### 0.1 — Monorepo Scaffold

| Task | Owner Layer | Output |
|---|---|---|
| Initialize Turborepo with `src/apps/`, `src/packages/`, `src/services/` | Frontend | Repo root + `turbo.json` |
| Configure `pnpm` workspaces | Frontend | `pnpm-workspace.yaml` |
| Add `biome.json` (lint + format) | Frontend | Replaces ESLint + Prettier |
| Add `lefthook.yml` (pre-commit hooks) | Frontend | Biome format on commit |
| Add `.nvmrc` (Node 22 LTS) | Frontend | Version pinning |
| Add `.editorconfig`, `.gitattributes` | Frontend | Cross-OS consistency |
| Add `.env.example` with all required vars | All | Environment documentation |

### 0.2 — GitHub Repository Setup

| Task | Output |
|---|---|
| Create `.github/workflows/ci.yml` | Turbo lint + typecheck on every PR |
| Create path-filtered deploy workflows (web, api, worker, postprocess) | Independent service deploys |
| Create `.github/CODEOWNERS` | Per-service review ownership |
| Add PR template + issue templates | Contribution standards |
| Configure branch protection on `main` | Require CI pass + review |
| Add Dependabot config | Auto-update deps |

### 0.3 — Vercel + Railway Provisioning

| Task | Service | Notes |
|---|---|---|
| Create Vercel project (connect GitHub) | `src/apps/web` | Preview deployments on PR |
| Create Railway project + 4 services | api, worker, postprocess, redis | TCP Railway Redis required for asynq |
| Provision Supabase project | PostgreSQL | Enable RLS immediately |
| Provision Upstash Redis (HTTP) | Cache + rate-limit | Separate from Railway Redis |
| Provision Cloudflare R2 bucket | Object storage | Zero egress cost |
| Create Mux environment | Video streaming | Signed playback enabled |
| Configure Doppler (dev / stg / prd) | All secrets | No `.env` files in repo |
| Create Clerk application | Auth | Enable Organizations for workspaces |

### 0.4 — Local Dev Environment

| Task | Output |
|---|---|
| `docker-compose.yml` for local dev | Postgres, Redis ×2 (ports 6379 + 6380), mock services |
| Go module init (`src/services/api`, `src/services/worker`) | `go.mod` for each |
| Rust project init (`src/services/postprocess`) | `Cargo.toml`, basic Axum health route |
| Next.js 15 app scaffold (`src/apps/web`) | App Router, Tailwind v4, `@theme {}` in `globals.css` |
| `src/packages/ui` — shadcn/ui init | Base components copied |
| `src/packages/types` — TypeScript types scaffold | Empty `src/index.ts` |
| `src/packages/config` — Biome + TS base configs | `biome/base.json` + `typescript/base.json` shared via workspace refs |

### Phase 0 Gate
- [ ] `turbo dev` starts all services without errors
- [ ] CI workflow passes on a blank PR
- [ ] Vercel preview URL live for `src/apps/web`
- [ ] Railway services deploy (even if they return 200 on `/health` only)
- [ ] All secrets in Doppler, not in repo

---

## Phase 1 — Core Data Layer
**Duration:** Week 2  
**Goal:** Complete schema in Postgres, type-safe Go DB layer via sqlc, RLS policies enforced, tRPC + Clerk auth wired.

### 1.1 — Database Schema

All tables created via migration files in `supabase/migrations/` (Supabase CLI — `supabase db push`). Do **not** add migration files to `src/services/api/db/` — that directory holds sqlc query definitions only.

**Core tables to create:**

```
workspaces        — Clerk org_id, plan_tier, stripe_customer_id
users             — Clerk user_id, workspace_id, role (admin/member/viewer)
brands            — workspace_id, name, primary_color, logo_r2_key, tone_notes
briefs            — brand_id, product_url, product_summary, template_used, human_reviewed
brief_angles      — brief_id, angle_name, rationale, emotion, funnel_stage, sort_order
brief_hooks       — angle_id, hook_type, opening_line, variant_index
jobs              — workspace_id, brief_id, status, type, created_by, created_at
variants          — job_id, angle_id, hook_id, format, mux_asset_id, r2_key, metadata
asset_tags        — variant_id, tag_key, tag_value (angle_type, hook_type, format, emotion, platform)
exports           — variant_id, workspace_id, format, filename, r2_key, created_at
```

**RLS policy pattern for every table:**
```sql
-- Always use subquery to avoid re-evaluation
CREATE POLICY "workspace_isolation" ON briefs
  FOR ALL TO authenticated
  USING (workspace_id IN (
    SELECT workspace_id FROM users WHERE id = (SELECT auth.uid())
  ));

-- Index required to prevent sequential scans under RLS
CREATE INDEX ON briefs (workspace_id);
```

### 1.2 — sqlc Code Generation

| Task | Output |
|---|---|
| Write `sqlc.yaml` with pgx/v5 driver | Config validated |
| Write SQL queries for all tables | `src/services/api/db/queries/*.sql` |
| Run `sqlc generate` | Type-safe Go structs + queriers in `src/services/api/internal/db/` |
| Add `make generate` target to Makefile | Developer shortcut |

### 1.3 — Go API Skeleton

| Task | Output |
|---|---|
| Echo v4 server setup with graceful shutdown | `src/services/api/main.go` |
| Middleware stack: request ID, logger, CORS, recover | `internal/middleware/` |
| Clerk JWT verification middleware | Validates `org_id` + `org_role` from JWT |
| Tier limit middleware (Starter/Growth/Agency) | `internal/middleware/tier.go` — server-side enforcement |
| Health endpoint `GET /health` | Returns `{"status":"ok"}` |
| Route group structure: `/v1/workspaces`, `/v1/brands`, `/v1/briefs`, `/v1/jobs`, `/v1/variants` | Route file |

### 1.4 — tRPC + Clerk Auth (Frontend)

| Task | Output |
|---|---|
| Install: `@trpc/server`, `@trpc/client`, `@trpc/react-query`, `@tanstack/react-query` | Package deps |
| `initTRPC.create()` with Clerk context | `src/apps/web/trpc/server.ts` |
| `appRouter` with stub routers: `briefs`, `assets`, `exports`, `projects`, `brands`, `jobs`, `org` | `src/apps/web/trpc/routers/` |
| tRPC Route Handler: `app/api/trpc/[trpc]/route.ts` | HTTP adapter |
| `ClerkProvider` + `TRPCProvider` + `QueryClientProvider` in root `layout.tsx` | Provider tree |
| `clerkMiddleware()` in `middleware.ts` | Protect all `(dashboard)` routes |

### Phase 1 Gate
- [ ] All migrations run cleanly on local Postgres
- [ ] RLS tested: user A cannot read user B's briefs
- [ ] sqlc types generated and imported into Go handlers
- [ ] `GET /health` returns 200 from Railway
- [ ] Sign in via Clerk; dashboard route protected; org_id in JWT verified by Go middleware
- [ ] tRPC hello query returns data through full Next.js → Go stack

---

## Phase 2 — URL Ingestion & Brief Engine
**Duration:** Weeks 3–4  
**Goal:** Paste a URL → receive an AI-generated creative brief. The core activation moment (US-03).

### 2.1 — Modal Playwright Scraper

| Task | Output |
|---|---|
| Create Modal Python function for Playwright scrape | `POST /scrape` webhook endpoint |
| Render JS SPAs (headless Chromium) | Full page render before extraction |
| Extract fields: name, category, price, features, proof points, CTA, image URLs | Structured JSON output |
| Add extraction confidence score (0–100 per field) | Surface to user if < 60 |
| 24-hour cache in Upstash by URL hash | Avoid repeat scraping |
| Return structured `ProductExtraction` JSON | Matches `src/packages/types/generation.ts` |

**Extraction timeout strategy:**
```
Page render timeout:  15 seconds → return partial extraction
Field extraction:     10 seconds → log and return what was extracted
Confidence < 40:      Prompt user to verify or switch to manual input
```

### 2.2 — Orchestration: Scrape (Go) + Brief Generation (Next.js BFF)

**Architecture:** Scraping is async (Go worker + Modal). Brief generation runs in the Next.js BFF tRPC mutation using Vercel AI SDK `generateObject()` — NOT in the Go worker.

| Task | Output |
|---|---|
| `POST /v1/briefs` Go handler → enqueue asynq `scrape_url` task | `src/services/worker/internal/tasks/scrape_url.go` |
| `scrape_url` asynq task → call Modal webhook, store `ProductExtraction` JSON, publish `scraped` status | Worker task complete; no LLM call here |
| tRPC `briefs.create` mutation (Next.js BFF) → on scrape complete: call `generateObject()` ×2 → persist brief | `src/apps/web/trpc/routers/briefs.ts` |
| GPT-4o `generateObject()` → `ProductSchema` structured extraction | Vercel AI SDK v6 in Next.js BFF |
| Claude Sonnet 4.6 `generateObject()` → `AnglesSchema` (3–5 angles + hooks) | Vercel AI SDK v6 in Next.js BFF |
| Status updates via Upstash Redis → SSE stream consumer | Job status: `queued → scraping → scraped → generating → complete → failed` |

```typescript
// src/apps/web/trpc/routers/briefs.ts — tRPC mutation (Next.js BFF)
import { generateObject } from 'ai';
import { openai } from '@ai-sdk/openai';
import { anthropic } from '@ai-sdk/anthropic';

const { object: product } = await generateObject({
  model: openai('gpt-4o'),
  schema: ProductSchema,
  prompt: `Extract product information from: ${extraction}`,
});
const { object: brief } = await generateObject({
  model: anthropic('claude-sonnet-4-6'),
  schema: AnglesSchema,
  prompt: buildAnglesPrompt(product, url),
});
// Persist brief to DB via Go API
```

### 2.3 — SSE Generation Stream

**Path:** `src/apps/web/app/api/generation/[jobId]/stream/route.ts`

```typescript
// Standalone Route Handler — NOT tRPC
export async function GET(req: Request, { params }: { params: { jobId: string } }) {
  const stream = new ReadableStream({
    start(controller) {
      // Subscribe to Redis pub/sub for job status updates
      // Emit: { event: "status", data: { status, progress, message } }
      // Close stream on "complete" or "failed"
    }
  });
  return new Response(stream, {
    headers: { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' }
  });
}
```

### 2.4 — GPT-4o Brief Generation

| Task | Output |
|---|---|
| Vercel AI SDK `generateObject()` with Zod schema | Structured brief JSON |
| System prompt: senior performance creative strategist role | `ai/prompts/angles-gen.prompt.ts` (Langfuse-versioned) |
| Output schema: `BriefOutput` (angles + hooks + formats + platform recs) | Validated via Zod / Go struct validation |
| Temperature 0.7 initial / 0.9 for regeneration | Configured per generation type |
| Validation + retry (up to 3 attempts on schema failure) | Quality gate before DB write |
| Langfuse trace on every GPT-4o call | Cost attribution per workspace |
| Industry template bias (7 templates) | DTC Fashion, Mobile App, SaaS, Health, Beauty, F&B, FinServ |

### 2.5 — Brief UI (Frontend)

| Task | Route | Notes |
|---|---|---|
| URL input + submission form | `(dashboard)/briefs/[id]/generate` | React Hook Form + Zod |
| SSE progress indicator: scraping → generating → ready | Same route | `useEffect` + `EventSource` |
| Brief preview card: angles + hooks + formats | Same route | Inline-editable |
| Inline editing (click to edit any field) | Same route | Optimistic update via tRPC |
| "Regenerate section" per angle / hook | Same route | tRPC mutation → new asynq task |
| Version history drawer (last 3 versions) | Same route | tRPC query |
| Manual brief input fallback | `(dashboard)/briefs/[id]/generate?mode=manual` | Toggle in UI |
| Industry template selector | Pre-brief modal | 7 templates + auto-detect |

### Phase 2 Gate
- [ ] Paste a Shopify URL → brief generated in < 25 seconds end-to-end
- [ ] JS SPA (headless render) extraction working
- [ ] SSE stream updates visible in browser (no polling)
- [ ] Brief editable inline; changes persisted to DB
- [ ] Regenerate single angle returns new variant < 10 seconds
- [ ] Manual input mode produces equivalent brief quality
- [ ] Langfuse traces showing per-brief GPT-4o cost

---

## Phase 3 — Video Generation Pipeline
**Duration:** Weeks 5–7  
**Goal:** Approved brief → video variants generated asynchronously via FAL.AI + ElevenLabs + HeyGen, rendered and stored.

### 3.1 — FAL.AI Video Generation (T2V / I2V)

| Task | Output |
|---|---|
| Install `@fal-ai/client` | Package dep |
| `fal.queue.submit()` for T2V (Veo 3.1 default) | Never `fal.subscribe()` |
| Model selection logic: Veo 3.1 (quality) / Kling 3.0 (speed) / Runway Gen-4.5 (style) / Sora 2 | Brief format determines model |
| Store `request_id` in asynq task payload | `src/services/worker/internal/tasks/generate_video.go` |
| Poll `fal.queue.status(request_id)` from worker every 10s | Long-running asynq task |
| On completion: download output, upload to R2 | `src/services/worker/internal/processor/fal.go` |

**I2V flow:** Product images from extraction → FAL.AI I2V → video clip with motion applied to product shots.

### 3.2 — ElevenLabs Voiceover

| Task | Output |
|---|---|
| Install `elevenlabs` SDK | Package dep |
| `eleven_v3` for final quality renders | Quality model |
| `eleven_flash_v2_5` for preview renders (~75ms) | Speed model for UI previews |
| Script generation: hook + body + CTA from brief angle | `src/services/worker/internal/processor/voice.go` |
| Audio output: MP3 → upload to R2 | Store R2 key on variant |
| Voice selection: default library + workspace custom voice (Growth+) | Tier-gated feature |

### 3.3 — HeyGen Avatar Lip-Sync (V2V)

| Task | Output |
|---|---|
| HeyGen Avatar API **v3** (`developers.heygen.com`) | V2V endpoint only |
| Submit: video clip + audio → lip-sync job | Async submission |
| Poll job status every 15s | `src/services/worker/internal/processor/heygen.go` |
| On completion: download lip-synced video → R2 | Store R2 key |
| Avatar library: default public avatars | Extended avatar upload: Agency tier |

> **Critical:** HeyGen v3 only. Any v4 reference is an error. v2 deprecated Oct 31, 2026.

### 3.4 — asynq Worker Pipeline

**Job state machine:**
```
queued
  → scraping         (Modal Playwright)
  → scraped
  → brief_generating (GPT-4o)
  → brief_ready
  → video_queued     (FAL.AI submitted)
  → voice_queued     (ElevenLabs submitted)
  → lipsync_queued   (HeyGen submitted — V2V only)
  → postprocessing   (Rust Axum)
  → complete
  → failed           (any step, retried up to 3× with backoff)
```

| Task | Output |
|---|---|
| Task type definitions | `src/services/worker/internal/tasks/*.go` |
| Processor implementations | `src/services/worker/internal/processor/` |
| Retry config: 3 attempts, exponential backoff | Per task type |
| Dead letter queue for unrecoverable failures | asynq built-in |
| Status publish to Redis pub/sub on every transition | Consumed by SSE stream |
| Variant limit enforcement before job enqueue | Go middleware check |

### 3.5 — Rust Video Postprocessor

| Task | Route | Notes |
|---|---|---|
| `POST /transcode` | Transcode to 1080p 9:16 | ffmpeg-sys |
| `POST /watermark` | Apply brand logo overlay | R2 presigned URL input |
| `POST /caption` | Burn-in captions from SRT | ffmpeg subtitles filter |
| `POST /reframe` | Smart reframe for 1:1 / 16:9 | Crop + scale |
| Health check `GET /health` | | |
| Input/output via R2 presigned URLs | No file upload to Rust service | Stateless design |

### 3.6 — Generation UI

| Task | Route | Notes |
|---|---|---|
| "Generate All" button → submit all angles | `(dashboard)/briefs/[id]` | Fires one job per angle |
| Per-variant SSE progress bars | Same route | One `EventSource` per jobId |
| Mux `<MuxPlayer>` for video preview | Same route | Signed playback URL |
| Variant grid: thumbnail + status badge | Same route | |
| Light edit panel: swap hook text / voice / CTA | Variant detail sheet | tRPC mutation |
| "Regenerate" with changed params | Variant detail sheet | Starts new job |
| Tier limit UI: upgrade prompt at limit | Same route | Client reads tier from tRPC; enforcement is server-side |

### Phase 3 Gate
- [ ] "Generate All" from approved brief creates one job per angle
- [ ] Each job transitions through all states visibly in SSE stream
- [ ] FAL.AI video returned, postprocessed, uploaded to R2, playable in Mux player
- [ ] Voiceover lip-synced via HeyGen v3 on UGC format
- [ ] Starter tier blocked at 3 variants; Growth at 10; Agency unlimited — enforced by Go middleware
- [ ] Failed jobs retry up to 3× and surface error state to user

---

## Phase 4 — Brand Kit & Export
**Duration:** Week 8  
**Goal:** Every generated asset is on-brand and exportable in platform-ready formats.

### 4.1 — Brand Kit System

| Task | Route | Notes |
|---|---|---|
| Brand creation wizard | `(dashboard)/brand/new` | Name, color, logo, tone |
| Logo upload → R2 presigned PUT | `(dashboard)/brand/[id]` | Direct upload, no proxy |
| Brand color picker + preview | Same route | Hex + contrast checker |
| Intro/outro bumper upload | Same route | MP4/MOV, max 5s |
| Custom font upload | Same route | TTF/OTF |
| Tone of voice notes (free text, 300 chars) | Same route | Fed into LLM system prompt |
| Multi-brand selector | Dashboard sidebar | One workspace = multiple brands |
| Brand kit auto-apply on generation | Worker layer | Passed as metadata to Rust postprocessor |

### 4.2 — Export Engine

| Task | Output |
|---|---|
| `POST /v1/exports` → generate export package | Named per convention: `[BrandName]_[AngleType]_[HookType]_[Platform]_V[n]` |
| Format options: MP4 (1080p), MP4 (4K Agency+), GIF (social preview) | Rust Axum transcode |
| Platform-specific exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16) | Pre-validated safe zones |
| Bulk download: ZIP of selected variants | Server-side ZIP in Go, stored R2, presigned download URL |
| Export history in Asset Library | DB record + R2 key |
| Platform compliance auto-check: safe zones, text size, duration | Checked pre-export |

### Phase 4 Gate
- [ ] Brand logo watermark appears on all generated videos
- [ ] Export downloads as correctly named MP4 in selected format
- [ ] Bulk ZIP download works for 10+ variants
- [ ] Platform compliance check rejects an out-of-spec asset

---

## Phase 5 — Asset Library & Team
**Duration:** Week 9  
**Goal:** All generated assets are findable, organized, and sharable within the workspace team.

### 5.1 — Asset Library

| Task | Route | Notes |
|---|---|---|
| All variants grid view | `(dashboard)/library` | Filter by brand / angle / format / date |
| Search by tag metadata | Same route | `asset_tags` table |
| Variant detail: preview + metadata + export | `(dashboard)/library/[variantId]` | Full metadata panel |
| Favorites / starring | Same route | `user_variant_stars` table |
| Archive vs Active status | Same route | Soft delete |
| Storage usage indicator per workspace | `(dashboard)/settings` | Bytes in R2 |

### 5.2 — Team & Collaboration

| Task | Notes |
|---|---|
| Invite team member by email | Clerk invitation via org |
| Role assignment: Admin / Member / Viewer | Viewer = read-only (Account Manager persona) |
| Viewer role: can see all assets, briefs, exports — cannot generate | Server-side tRPC + Go middleware check |
| Remove member | Clerk org member removal |
| Pending invite list | Clerk pending invitations |
| Seat count display | `(dashboard)/settings/team` |

### Phase 5 Gate
- [ ] Library shows all variants with correct metadata filters
- [ ] Viewer role cannot trigger generation (blocked at API + UI)
- [ ] Team invite flow complete: invite → email → accept → workspace access
- [ ] Archive removes from default view; accessible via filter

---

## Phase 6 — Platform, Billing & Trial
**Duration:** Week 10  
**Goal:** Stripe subscriptions live, trial flow complete, tier enforcement locked, conversion emails triggered.

### 6.1 — Stripe Integration

| Task | Output |
|---|---|
| Create Stripe products + prices (Starter $99 / Growth $149 / Agency $399) | Stripe dashboard + IDs in Doppler |
| `POST /v1/billing/checkout` → Stripe checkout session | Go handler |
| `POST /v1/billing/portal` → Stripe customer portal | Go handler |
| Webhook handler: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted` | `src/services/api/internal/handler/webhooks.go` |
| Webhook: update `workspaces.plan_tier` + `workspaces.stripe_status` on every event | DB write via sqlc |
| Stripe Entitlements API for feature flags | Growth+ features gated |
| Idempotency on all webhook handlers (Stripe-Signature header verification) | Security requirement |

**Subscription status machine:**
```
trialing → active → past_due → canceled
                 ↘ canceled
```

### 6.2 — Trial Flow

| Task | Notes |
|---|---|
| 7-day trial starts on first workspace creation | `trial_ends_at = created_at + 7 days` |
| Trial badge in navigation: "Trial — X days left" | Client: days remaining from `trial_ends_at` |
| Day 5 in-app banner: urgency message | PostHog feature flag + DB check |
| Day 7 login modal: "Your trial ends today" | Checked on every `(dashboard)` layout load |
| Day 8: generation blocked — return 402 from Go API | `workspace.stripe_status === 'trial_expired'` |
| 30-day data retention post-expiry | Scheduled cleanup job in asynq |
| Conversion email sequence: Day 3 / Day 6 / Day 8 | Triggered via asynq scheduled tasks → email provider |

### 6.3 — Tier Enforcement

Already plumbed in Phase 1 middleware. Phase 6 activates and tests all limits:

```go
// src/services/api/internal/middleware/tier.go
maxVariants := map[string]int{
  "starter": 3,
  "growth":  10,
  "agency":  -1, // unlimited
}[workspace.PlanTier]

if maxVariants != -1 && count >= maxVariants {
  return c.JSON(402, map[string]string{"error": "variant_limit_exceeded"})
}
```

| Gate | Starter | Growth | Agency |
|---|---|---|---|
| Variants per angle | 3 | 10 | Unlimited |
| Custom voice upload | ✗ | ✓ | ✓ |
| 4K export | ✗ | ✗ | ✓ |
| Custom avatar upload | ✗ | ✗ | ✓ |
| Seat count | 3 | 10 | Unlimited |
| API access | ✗ | ✗ | ✓ |

### Phase 6 Gate
- [ ] Stripe checkout flow end-to-end (test mode): free trial → paid plan
- [ ] Webhook updates plan tier in DB within 5 seconds of Stripe event
- [ ] Starter user blocked at 4th variant with clear upgrade CTA
- [ ] Trial badge visible; Day 8 lock activates automatically
- [ ] Canceled subscription blocks generation within 1 billing cycle
- [ ] Stripe webhook signature verified on every event (rejects unsigned requests)

---

## Phase 7 — Polish, Observability & Launch
**Duration:** Week 11  
**Goal:** Production-grade reliability, full observability stack, QA sign-off, and staged launch.

### 7.1 — Observability Stack

| Tool | Integration | What It Tracks |
|---|---|---|
| **Sentry** | `@sentry/nextjs` + Go SDK + Rust SDK | Errors + stack traces across all 4 services |
| **Better Stack** | Log drain from Railway | Structured logs with severity + trace IDs |
| **PostHog** | `posthog-js` in `src/apps/web` | User activation funnel, feature adoption, trial conversion |
| **Langfuse** | Tracing in worker GPT-4o calls | LLM cost per workspace, prompt versions, latency |

**Key PostHog events to instrument:**

```
user_signed_up          → role, plan
brand_created           → workspace_id
brief_generated         → method (url/manual), latency, template_used
video_generation_started → format, angle_type, tier
video_generation_complete → duration, model_used
export_downloaded       → format, platform
trial_to_paid           → days_to_convert, plan_selected
variant_limit_hit       → tier, upgrade_shown
```

### 7.2 — Performance & Security Hardening

| Task | Notes |
|---|---|
| Upstash rate limiting on all public endpoints | `@upstash/ratelimit` — 60 req/min per workspace |
| Input validation on all Go handlers | Echo's built-in validator + custom rules |
| R2 presigned URLs expire in 15 minutes | Prevent link sharing bypass |
| Mux signed playback tokens | Workspace-scoped access only |
| CORS locked to verified origins | No wildcard in production |
| Stripe webhook signature verification | HMAC-SHA256 `Stripe-Signature` header |
| Clerk JWT expiry + refresh handling | Frontend refreshes on 401 |
| RLS tested with cross-workspace user | Automated test in CI |

### 7.3 — QA Sign-Off Checklist

**Core flows must pass end-to-end in staging:**

- [ ] Sign up → brand setup → first brief → video generated → exported (< 15 min, new user)
- [ ] URL scrape: Shopify PDP, WooCommerce, App Store listing, custom landing page
- [ ] Manual brief input → same output quality as URL
- [ ] Brief editing: inline edit + regenerate section + version history
- [ ] All 3 video formats: UGC, Product Demo, Text-Motion
- [ ] All 3 aspect ratios: 9:16, 1:1, 16:9
- [ ] ElevenLabs TTS + HeyGen v3 lip-sync on UGC format
- [ ] SSE stream: progress visible for full generation lifecycle
- [ ] Failed job: retries 3× then surfaces error state
- [ ] Tier limits: Starter (3), Growth (10), Scale (unlimited) enforced
- [ ] Team invite: Admin → invite Viewer → Viewer cannot generate
- [ ] Trial flow: Day 5 banner, Day 7 modal, Day 8 lock
- [ ] Stripe: checkout → plan upgrade → webhook → tier unlocked
- [ ] Stripe: cancellation → webhook → generation blocked
- [ ] Bulk export ZIP: 10 variants
- [ ] Library: filter by brand / angle type / format / date

### 7.4 — Launch Readiness

| Task | Notes |
|---|---|
| Staging environment full smoke test | Mirror of production config |
| Database backup automation (Supabase) | Daily + point-in-time restore |
| Railway service restart policy | `ALWAYS` with health check |
| Doppler `prd` environment locked | Only infra lead can edit |
| Status page (Better Stack) | Public status.qvora.com |
| Incident runbook | P1/P2 severity definitions + on-call |
| README with local dev setup | < 5 commands to run locally |

### Phase 7 Gate
- [ ] All QA checklist items pass in staging
- [ ] Sentry error rate < 0.1% on staging smoke test
- [ ] p95 brief generation latency < 25s
- [ ] p95 video generation < 180s
- [ ] PostHog funnel tracking first 5 events in staging
- [ ] Langfuse showing cost per brief in staging
- [ ] Zero secrets in repo (Doppler audit)

---

## Cross-Phase Dependencies

```
Phase 0 (Infra)
  └─► Phase 1 (Data Layer)
        └─► Phase 2 (Brief Engine)    ─────────────────────────────┐
              └─► Phase 3 (Video Gen)                              │
                    ├─► Phase 4 (Brand Kit + Export)               │
                    ├─► Phase 5 (Library + Team)                   │
                    └─► Phase 6 (Billing + Trial) ◄────────────────┘
                          └─► Phase 7 (Polish + Launch)
```

**Blocking dependencies:**
- Clerk auth (Phase 0) blocks all authenticated features
- DB schema + RLS (Phase 1) blocks all data writes
- SSE stream (Phase 2) blocks video generation progress UI
- asynq worker pipeline (Phase 3) blocks all async generation
- Stripe webhooks (Phase 6) block tier enforcement + trial lock

---

## V2 Scope (Post-Launch)

These modules are explicitly excluded from V1 and will be planned separately:

| Module | ID | Description |
|---|---|---|
| Ad Account Connector | CONN | Meta + TikTok API integration; push exports directly to ad accounts |
| Performance Learning Engine | SIGNAL | Ingest ROAS / CTR / CPC signal per variant; rank and surface winning patterns |
| DTC Brand Manager features | P4 | Self-serve onboarding optimized for in-house brand teams |
| Bulk URL batch import | EXT-V2 | Process multiple product URLs in one job |
| Real-time brief collaboration | BRIEF-V2 | Multi-user simultaneous editing |
| 4K video output | GEN-V2 | Scale tier 4K export |
| PDF product sheet extraction | EXT-V2 | File upload fallback for pre-launch products |

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| FAL.AI generation latency spikes > 3 min | Medium | High | Model fallback chain (Veo 3.1 → Kling 3.0 → Runway) |
| HeyGen v3 rate limits / availability | Medium | High | Queue with retry backoff; async not blocking |
| Playwright scrape blocked by CAPTCHAs | Medium | Medium | Confidence score fallback → manual input |
| RLS policy misconfiguration → data leak | Low | Critical | Automated cross-workspace test in CI |
| Stripe webhook replay / duplicate processing | Medium | Medium | Idempotency key on every webhook handler |
| Railway Redis TCP connection drop → asynq stall | Low | High | asynq reconnect config; Railway auto-restart |
| GPT-4o cost overrun on free trials | Medium | Medium | Langfuse cost alerts + per-workspace trial limits |
| Modal cold start adds to scrape latency | Medium | Low | Keep-warm ping every 5 min; async UX masks latency |
