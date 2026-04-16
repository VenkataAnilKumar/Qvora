# Qvora — Implementation Checklist
**Last Updated:** Apr 16, 2026 | **Total:** 8 Phases (11 weeks) | **V1 Target: Launch**

---

## Progress Summary

| Phase | Name | Duration | Status |
|---|---|---|---|
| Phase 0 | Foundation & Infrastructure | Week 1 | ✅ Complete |
| Phase 1 | Core Data Layer | Week 2 | ✅ Complete |
| Phase 2 | URL Ingestion & Brief Engine | Weeks 3–4 | ✅ Complete |
| Phase 3 | Video Generation Pipeline | Weeks 5–7 | ✅ Complete |
| Phase 4 | Brand Kit & Export | Week 8 | ⚠️ Partial |
| Phase 5 | Asset Library & Team | Week 9 | ❓ Not Validated |
| Phase 6 | Platform, Billing & Trial | Week 10 | ❓ Not Validated |
| Phase 7 | Polish, Observability & Launch | Week 11 | ❓ Not Validated |

**Complete: 4 / 8 — Partial: 1 / 8 — Not Started: 3 / 8**

---

## Phase 0 — Foundation & Infrastructure ✅ Complete

### Monorepo & Tooling
- [x] Turborepo with `src/apps/`, `src/packages/`, `src/services/`
- [x] pnpm workspaces configured
- [x] `biome.json` (replaces ESLint + Prettier)
- [x] `lefthook.yml` (replaces Husky)
- [x] `.nvmrc` (Node 22 LTS)
- [x] `.editorconfig`, `.gitattributes`
- [x] `.env.example` with all required vars

### GitHub Setup
- [x] `ci.yml` — Turbo lint + typecheck on every PR
- [x] Path-filtered deploy workflows (web, api, worker, postprocess, db)
- [x] `CODEOWNERS`
- [x] PR template + issue templates
- [x] Branch protection on `main`
- [x] Dependabot config

### Infrastructure Provisioning
- [x] Vercel project (apps/web, preview deploys on PR)
- [x] Railway project — api, worker, postprocess, Railway Redis (TCP)
- [x] Supabase project (PostgreSQL, RLS enabled)
- [x] Upstash Redis (HTTP — cache + rate-limit only)
- [x] Cloudflare R2 bucket (zero egress)
- [x] Mux environment (signed playback enabled)
- [x] Doppler (dev / stg / prd)
- [x] Clerk application (Organizations enabled)

### Local Dev
- [x] `docker-compose.yml` — Postgres + Redis ×2 + all services
- [x] Go module init (`src/services/api`, `src/services/worker`)
- [x] Rust project init (`src/services/postprocess`) — Axum health route
- [x] Next.js 15 scaffold (`src/apps/web`) — App Router, Tailwind v4, `@theme {}`
- [x] `src/packages/ui` — shadcn/ui components
- [x] `src/packages/types` — shared TypeScript types
- [x] `src/packages/config` — Biome + TS base configs

### Gate
- [x] `turbo dev` starts all services without errors
- [x] CI passes on blank PR
- [x] Vercel preview URL live
- [x] Railway services respond on `/health`
- [x] All secrets in Doppler, not in repo

---

## Phase 1 — Core Data Layer ✅ Complete

### Database Schema
- [x] Migrations in `supabase/migrations/` only (not `services/api/db/`)
- [x] Tables: workspaces, users, brands, briefs, brief_angles, brief_hooks, jobs, variants, asset_tags, exports
- [x] `jobs.status` CHECK: `queued, scraping, generating, postprocessing, complete, failed` (no 'briefing')
- [x] `plan_tier` CHECK: `starter, growth, agency` (no 'scale')
- [x] RLS policies on all tables (workspace isolation pattern)
- [x] Indexes on workspace_id, status, job_id, brief_id
- [x] Additional migrations: `002_postprocess_callbacks.sql`, `003_mux_webhook_events_and_reconcile.sql`

### sqlc
- [x] `sqlc.yaml` — schema points to `../../supabase/migrations/`, pgx/v5 driver
- [x] SQL queries for all core operations (`src/services/api/db/queries/queries.sql`)
- [x] `sqlc generate` run — `src/services/api/internal/db/` generated
- [x] Queries: CreateJob, UpdateJobStatus, GetJobByID, ListJobsByWorkspace
- [x] Queries: CreateVariant, UpdateVariantFalRequestID, UpdateVariantMuxByID
- [x] Queries: CreateBrief, CreateBriefAngle, CreateBriefHook, ListBriefsByWorkspace

### Go API Skeleton
- [x] Echo v4 server with graceful shutdown (`src/services/api/cmd/api/main.go`)
- [x] Middleware: request ID, logger, CORS, recover, rate-limiter
- [x] Clerk JWT middleware — extracts `org_id` + `org_role` (NOT `app_metadata` path)
- [x] Tier limit middleware — keys: `starter/growth/agency` (NOT `scale`)
- [x] `GET /health` → `{"status":"ok"}`
- [x] Route groups: `/v1/jobs`, `/v1/briefs`, `/v1/workspaces`, `/v1/assets`, `/v1/exports`, `/v1/variants`

### tRPC + Frontend Auth
- [x] `@trpc/server`, `@trpc/client`, `@trpc/react-query`, `@tanstack/react-query` installed
- [x] `initTRPC.create()` with Clerk context (`userId`, `orgId`, `orgRole`)
- [x] `appRouter` routers: briefs, assets, exports, projects, brands, jobs, org (no 'campaigns')
- [x] tRPC Route Handler: `app/api/trpc/[trpc]/route.ts`
- [x] `ClerkProvider` + `TRPCProvider` + `QueryClientProvider` in root `layout.tsx`
- [x] `clerkMiddleware()` in `middleware.ts`
- [x] No `tailwind.config.ts` — all tokens in `globals.css` `@theme {}`

### Gate
- [x] Migrations run cleanly on local Postgres
- [x] RLS: user A cannot read user B's briefs
- [x] sqlc types generated and imported into Go handlers
- [x] `GET /health` returns 200 from Railway
- [x] Clerk sign-in → org_id in JWT verified by Go middleware
- [x] tRPC query returns data through full Next.js → Go stack

---

## Phase 2 — URL Ingestion & Brief Engine ✅ Complete

### Modal Playwright Scraper
- [x] Modal Python function for Playwright scrape (`POST /scrape`)
- [x] JS SPA headless render (Chromium)
- [x] Extracts: name, category, price, features, proof points, CTA, image URLs
- [x] Confidence score (0–100 per field)
- [x] 24-hour Upstash cache by URL hash
- [x] Returns `ProductExtraction` JSON

### Orchestration
- [x] `POST /v1/briefs` Go handler → enqueues asynq `scrape_url` task
- [x] `scrape_url` task → calls Modal, stores ProductExtraction, publishes status
- [x] Job transitions: `queued → scraping → generating` (no 'briefing' step)
- [x] No LLM call in Go worker (brief generation is Next.js BFF only)

### Brief Generation (Next.js BFF — NOT Go worker)
- [x] `@ai-sdk/openai` + `@ai-sdk/anthropic` installed
- [x] `generateObject()` ×2 in `src/apps/web/src/server/trpc/routers/briefs.ts`
  - [x] GPT-4o → `productExtractionSchema` (structured product from scraped data)
  - [x] Claude Sonnet 4.6 → `anglesGenerationSchema` (3–5 angles + hooks)
- [x] `OPENAI_API_KEY` + `ANTHROPIC_API_KEY` guards
- [x] Prompt file at `src/ai/prompts/angles-gen.prompt.ts` (imported via `@qvora/prompts/*` alias)
- [x] `@qvora/prompts/*` path alias in `src/apps/web/tsconfig.json`
- [x] Brief + angles + hooks persisted to Go API after generation

### SSE Stream
- [x] Standalone Route Handler at `src/apps/web/src/app/api/generation/[jobId]/stream/route.ts`
- [x] Uses `ReadableStream` — NOT tRPC subscription
- [x] Proxies to Go API `GET /api/v1/jobs/:id/stream`
- [x] Frontend `EventSource` connects to `/api/generation/{jobId}/stream`
- [x] React Query polling fallback

### Frontend Routes
- [x] `(dashboard)/briefs/page.tsx` — brief list
- [x] `(dashboard)/briefs/[id]/page.tsx` — brief detail
- [x] No `(dashboard)/campaigns/` routes

### Gate
- [x] Shopify URL → brief generated < 25 seconds
- [x] SSE stream updates visible in browser
- [x] Brief persisted to DB (angles + hooks saved)
- [x] Inline editing persisted to DB (BRIEF-08, P0)
- [x] Per-angle / per-hook regeneration button (BRIEF-09, P0)
- [x] Regenerate single angle < 10 seconds
- [x] Langfuse traces per brief

---

## Phase 3 — Video Generation Pipeline ✅ Complete

### FAL.AI Video Generation
- [x] `@fal-ai/client` installed
- [x] `fal.queue.submit()` — NEVER `fal.subscribe()`
- [x] Models: Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2
- [x] `fal_request_id` stored in variants table (`PATCH /api/v1/variants/:id/fal-request`)
- [x] 9:16 aspect ratio enforced in FAL payload

### ElevenLabs Voiceover
- [x] `elevenlabs` SDK installed
- [x] `eleven_v3` (quality) + `eleven_flash_v2_5` (preview ~75ms)
- [x] Script from brief angle (hook + body + CTA)
- [x] Audio MP3 → R2

### HeyGen V2V Lip-Sync
- [x] HeyGen Avatar API **v3** only (`developers.heygen.com`) — no v4
- [x] V2V: video clip + audio → lip-sync job (async)
- [x] Poll job status every 15s
- [x] Output → R2

### asynq Worker Pipeline
- [x] Railway Redis TCP (never Upstash) for asynq
- [x] Task types: TypeScrape, TypeGenerate, TypePostprocess
- [x] Queue priorities: critical / default / low
- [x] Retry: 3 attempts, exponential backoff
- [x] Dead letter queue for unrecoverable failures
- [x] Status callbacks to Go API via `patchJobStatus()`

### Rust Video Postprocessor
- [x] Axum HTTP server — `POST /process`, `GET /health`
- [x] ffmpeg-next bindings (NOT `Command::new("ffmpeg")`)
- [x] 9:16 reframe + letterbox (1080×1920)
- [x] H.264 transcode (libx264, 2.5 Mbps)
- [x] Watermark overlay (drawtext filter)
- [x] Caption burn-in (drawtext, multiline, escaped)
- [x] Input/output via R2 presigned URLs (stateless)
- [x] Callback to Go API on completion/failure
- [x] Multi-stage Dockerfile with ffmpeg runtime libs

### Mux Integration
- [x] Mux asset upload from R2 presigned URL
- [x] `mux_asset_id` + `mux_playback_id` stored in variants
- [x] Signed playback tokens (workspace-scoped JWT, 1-hour expiry)
- [x] `POST /webhooks/mux` — HMAC-SHA256 verified, handles `video.asset.ready`
- [x] `GET /api/v1/variants/:id/playback-url` → signed token
- [x] `<MuxPlayer>` in dashboard with signed token

### Webhooks & Callbacks
- [x] `POST /webhooks/fal` — parses completion, enqueues `job:postprocess` to critical queue
- [x] `POST /webhooks/mux` — updates variant mux IDs, marks complete
- [x] `POST /internal/postprocess/callback` — postprocessor success/failure callback
- [x] `POST /internal/jobs/reconcile-stuck` — stale job recovery

### Gate
- [x] FAL → postprocess → Mux → playback pipeline wired end-to-end
- [x] Tier enforcement: starter=3, growth=10, agency=unlimited
- [x] `jobs.status` enum matches DB constraint (no 'briefing')
- [x] "Generate All" E2E flow wired and validated in app stack
- [x] Failed jobs retry configured to ×3 and failed state is surfaced

---

## Phase 4 — Brand Kit & Export ⚠️ Partial

### Brand Kit System
- [ ] Brand creation wizard (`(dashboard)/brand/new`)
- [ ] Logo upload → R2 presigned PUT
- [ ] Brand color picker + preview
- [ ] Intro/outro bumper upload (MP4/MOV, max 5s)
- [ ] Custom font upload (TTF/OTF)
- [ ] Tone of voice notes (300 chars → fed into LLM prompt)
- [ ] Multi-brand selector in sidebar
- [ ] Brand kit auto-applied on generation (passed to Rust postprocessor)

### Export Engine
- [ ] `POST /v1/exports` → named package: `[Brand]_[Angle]_[Hook]_[Platform]_V[n]`
- [ ] Formats: MP4 1080p, MP4 4K (Agency+), GIF preview
- [ ] Platform exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16)
- [ ] Bulk ZIP download (Go server-side, R2 presigned URL)
- [ ] Export history in DB + R2 key
- [ ] Platform compliance check (safe zones, text size, duration)

### Gate
- [ ] Brand logo watermark on all generated videos
- [ ] Export downloads as correctly named MP4
- [ ] Bulk ZIP works for 10+ variants
- [ ] Platform compliance check rejects out-of-spec asset

---

## Phase 5 — Asset Library & Team ❓ Not Validated

### Asset Library
- [ ] Variants grid view (`(dashboard)/library`) — filter by brand / angle / format / date
- [ ] Search by tag metadata (`asset_tags` table)
- [ ] Variant detail page (`(dashboard)/library/[variantId]`)
- [ ] Favorites / starring (`user_variant_stars` table)
- [ ] Archive vs Active (soft delete)
- [ ] Storage usage indicator per workspace

### Team & Collaboration
- [ ] Invite team member by email (Clerk org invitation)
- [ ] Role assignment: Admin / Member / Viewer
- [ ] Viewer role: read-only — cannot generate (enforced at API + tRPC)
- [ ] Remove member (Clerk org)
- [ ] Pending invite list
- [ ] Seat count display (`(dashboard)/settings/team`)

### Gate
- [ ] Library shows all variants with correct metadata filters
- [ ] Viewer role cannot trigger generation (blocked at API + UI)
- [ ] Team invite flow: invite → email → accept → workspace access
- [ ] Archive removes from default view

---

## Phase 6 — Platform, Billing & Trial ❓ Not Validated

### Stripe Integration
- [ ] Stripe products + prices: Starter $99 / Growth $149 / Agency $399
- [ ] `POST /v1/billing/checkout` → Stripe checkout session
- [ ] `POST /v1/billing/portal` → Stripe customer portal
- [ ] Webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] Webhook updates `workspaces.plan_tier` + `workspaces.stripe_status`
- [ ] Stripe Entitlements API for feature flags
- [ ] Idempotent webhook handlers (Stripe-Signature HMAC-SHA256 verified)

### Trial Flow
- [ ] 7-day trial on workspace creation (`trial_ends_at = created_at + 7 days`)
- [ ] Trial badge: "Trial — X days left"
- [ ] Day 5 in-app urgency banner
- [ ] Day 7 login modal
- [ ] Day 8: generation blocked (402 from Go API)
- [ ] 30-day data retention post-expiry (asynq cleanup job)
- [ ] Conversion emails: Day 3 / Day 6 / Day 8

### Tier Enforcement (activation)
- [ ] Starter blocked at 4th variant (clear upgrade CTA)
- [ ] Growth blocked at 11th variant
- [ ] Agency: unlimited
- [ ] Custom voice: Growth+ only
- [ ] 4K export: Agency only
- [ ] Custom avatar: Agency only

### Gate
- [ ] Stripe checkout E2E (test mode): trial → paid
- [ ] Webhook updates plan tier in DB within 5 seconds
- [ ] Day 8 lock activates automatically
- [ ] Stripe webhook rejects unsigned requests

---

## Phase 7 — Polish, Observability & Launch ❓ Not Validated

### Observability Stack
- [ ] Sentry — `@sentry/nextjs` + Go SDK + Rust SDK (errors across all 4 services)
- [ ] Better Stack — log drain from Railway (structured logs + trace IDs)
- [ ] PostHog — `posthog-js` (activation funnel, trial conversion, feature adoption)
- [ ] Langfuse — LLM cost per workspace, prompt versions, latency traces

### PostHog Events
- [ ] `user_signed_up` (role, plan)
- [ ] `brief_generated` (method, latency, template)
- [ ] `video_generation_started` (format, angle_type, tier)
- [ ] `video_generation_complete` (duration, model)
- [ ] `export_downloaded` (format, platform)
- [ ] `trial_to_paid` (days_to_convert, plan)
- [ ] `variant_limit_hit` (tier, upgrade_shown)

### Security Hardening
- [ ] Upstash rate limiting on all public endpoints (60 req/min per workspace)
- [ ] Input validation on all Go handlers
- [ ] R2 presigned URLs expire in 15 minutes
- [ ] CORS locked to verified origins (no wildcard in prod)
- [ ] RLS cross-workspace test in CI

### QA Sign-Off
- [ ] E2E: sign up → brand setup → brief → video generated → exported (< 15 min)
- [ ] URL scrape: Shopify PDP, WooCommerce, App Store, custom landing page
- [ ] Tier limits enforced at all 3 tiers in staging
- [ ] Trial flow: Day 3/6/8 emails triggered, Day 8 lock verified
- [ ] Load test: 10 concurrent generation jobs without failure

### Launch
- [ ] Staging environment fully green
- [ ] Production deploy checklist complete
- [ ] Error budget / SLO defined
- [ ] On-call runbook written
- [ ] Launch comms ready

### Gate
- [ ] All QA sign-off items passing in staging
- [ ] Zero P0 bugs open
- [ ] Observability dashboards live
- [ ] Production deploy successful

---

## Environment Variables Reference

### Required — Phases 0–3 (current)
```
# Web
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
CLERK_SECRET_KEY
OPENAI_API_KEY
ANTHROPIC_API_KEY
NEXT_PUBLIC_APP_URL
GO_API_URL
MODAL_SCRAPER_ENDPOINT
INTERNAL_API_KEY

# API
DATABASE_URL
CLERK_SECRET_KEY
INTERNAL_API_KEY
MUX_TOKEN_ID
MUX_TOKEN_SECRET
MUX_WEBHOOK_SECRET
FAL_WEBHOOK_SECRET
RAILWAY_REDIS_URL
UPSTASH_REDIS_REST_URL
UPSTASH_REDIS_REST_TOKEN

# Worker
DATABASE_URL
RAILWAY_REDIS_URL
FAL_KEY
MODAL_SCRAPER_ENDPOINT
RUST_POSTPROCESS_URL
INTERNAL_API_KEY
API_BASE_URL

# Postprocessor
R2_ENDPOINT
R2_BUCKET
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
MUX_ACCESS_TOKEN
MUX_SECRET_TOKEN
API_BASE_URL
INTERNAL_API_KEY
```

### Required — Phase 6 additions
```
STRIPE_SECRET_KEY
STRIPE_WEBHOOK_SECRET
STRIPE_STARTER_PRICE_ID
STRIPE_GROWTH_PRICE_ID
STRIPE_AGENCY_PRICE_ID
```

### Required — Phase 7 additions
```
SENTRY_DSN (web + api + worker + postprocessor)
POSTHOG_API_KEY
LANGFUSE_SECRET_KEY
LANGFUSE_PUBLIC_KEY
BETTERSTACK_SOURCE_TOKEN
```
