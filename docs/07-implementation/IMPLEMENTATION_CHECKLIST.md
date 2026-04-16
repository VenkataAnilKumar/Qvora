# Qvora — Implementation Checklist
**Last Updated:** Apr 16, 2026 | **Total:** 10 Phases | **V1 Target: Week 11**

---

## Progress Summary

| Phase | Name | Duration | Status |
|---|---|---|---|
| Phase 0 | Foundation & Infrastructure | Week 1 | ✅ Complete |
| Phase 1 | Core Data Layer | Week 2 | ✅ Complete |
| Phase 2 | URL Ingestion & Brief Engine | Weeks 3–4 | ✅ Complete |
| Phase 3 | Video Generation Pipeline | Weeks 5–7 | ✅ Complete |
| Phase 4 | Brand Kit & Export | Week 8 | ⏳ Pending |
| Phase 5 | Asset Library & Team | Week 9 | ⏳ Pending |
| Phase 6 | Platform, Billing & Trial | Week 10 | ⏳ Pending |
| Phase 7 | V1 Polish, Observability & Launch | Week 11 | ⏳ Pending |
| Phase 8 | Microservice Foundation | Weeks 12–13 | ⏳ Post-Launch |
| Phase 9 | Temporal + gRPC + Multi-Provider | Weeks 14–15 | ⏳ Post-Launch |
| Phase 10 | V2 Signal Loop & Intelligence | Weeks 16–18 | ⏳ Post-Launch |

**Complete: 4/10 — Pending V1: 4/10 — Post-Launch: 3/10**

---

## Phase 0 — Foundation & Infrastructure ✅ Complete

### Monorepo & Tooling
- [x] Turborepo with `src/apps/`, `src/packages/`, `src/services/`, `src/ai/`
- [x] pnpm workspaces configured
- [x] `biome.json` (replaces ESLint + Prettier)
- [x] `lefthook.yml` (replaces Husky)
- [x] `.nvmrc` (Node 22 LTS)
- [x] `.env.example` with all required vars

### GitHub Setup
- [x] `ci.yml` — Turbo lint + typecheck on every PR
- [x] Path-filtered deploy workflows (web, api, worker, postprocess)
- [x] `CODEOWNERS`, PR template, issue templates
- [x] Branch protection on `main`
- [x] Dependabot config

### Infrastructure Provisioning
- [x] Vercel project (preview deploys on PR)
- [x] Railway project (api, worker, postprocess, Railway Redis TCP)
- [x] Supabase project (PostgreSQL 16, RLS enabled)
- [x] Upstash Redis (HTTP — cache + rate-limit only)
- [x] Cloudflare R2 bucket (zero egress)
- [x] Mux environment (signed playback enabled)
- [x] Doppler (dev / stg / prd)
- [x] Clerk application (Organizations enabled)

### Local Dev
- [x] `docker-compose.yml` — Postgres + Redis ×2 + all services
- [x] Go modules: `src/services/api`, `src/services/worker`
- [x] Rust project: `src/services/postprocess` (Axum health route)
- [x] Next.js 15 scaffold — App Router, Tailwind v4, `@theme {}`
- [x] `src/packages/ui` — shadcn/ui, `src/packages/types`, `src/packages/config`

### Gate ✅
- [x] `turbo dev` starts all services without errors
- [x] CI passes on blank PR
- [x] Vercel preview URL live
- [x] Railway services respond on `/health`
- [x] All secrets in Doppler, not in repo

---

## Phase 1 — Core Data Layer ✅ Complete

### Database Schema
- [x] Migrations in `supabase/migrations/` only
- [x] Tables: workspaces, users, brands, briefs, brief_angles, brief_hooks, jobs, variants, asset_tags, exports
- [x] `jobs.status` CHECK: `queued, scraping, generating, postprocessing, complete, failed`
- [x] `plan_tier` CHECK: `starter, growth, agency`
- [x] RLS policies on all tables (workspace isolation)
- [x] Indexes on workspace_id, status, job_id, brief_id
- [x] Migrations: `002_postprocess_callbacks.sql`, `003_mux_webhook_events_and_reconcile.sql`

### sqlc + Go API
- [x] `sqlc.yaml` — schema points to `../../supabase/migrations/`, pgx/v5 driver
- [x] SQL queries for all core operations
- [x] Echo v4 API with all middleware (request ID, logger, CORS, recover, rate-limiter)
- [x] Clerk JWT middleware — extracts `org_id` + `org_role`
- [x] Tier enforcement middleware — keys: `starter/growth/agency`
- [x] Route groups: `/v1/jobs`, `/v1/briefs`, `/v1/workspaces`, `/v1/assets`, `/v1/exports`, `/v1/variants`

### tRPC + Frontend Auth
- [x] `initTRPC.create()` with Clerk context
- [x] `appRouter` routers: briefs, assets, exports, projects, brands, jobs, org
- [x] `ClerkProvider` + `TRPCProvider` + `QueryClientProvider` in root `layout.tsx`

### Gate ✅
- [x] Migrations run cleanly on local Postgres
- [x] RLS: user A cannot read user B's briefs
- [x] `GET /health` returns 200 from Railway
- [x] Clerk sign-in → org_id in JWT verified by Go middleware
- [x] tRPC query returns data through full Next.js → Go stack

---

## Phase 2 — URL Ingestion & Brief Engine ✅ Complete

### Modal Playwright Scraper
- [x] Modal Python function for Playwright scrape
- [x] Extracts: name, category, price, features, proof points, CTA, image URLs
- [x] Confidence score per field; 24-hour Upstash cache by URL hash

### Orchestration + Brief Generation
- [x] `POST /v1/briefs` → enqueues asynq `scrape_url` task
- [x] Job transitions: `queued → scraping → generating`
- [x] Brief generation in Next.js BFF tRPC (not Go worker)
- [x] `generateObject()` ×2: GPT-4o (product) + Claude Sonnet 4.6 (angles + hooks)
- [x] Brief + angles + hooks persisted to Go API

### SSE Stream (V1)
- [x] Standalone Route Handler at `src/apps/web/src/app/api/generation/[jobId]/stream/route.ts`
- [x] Uses `ReadableStream` — NOT tRPC subscription
- [x] Frontend `EventSource` connects to `/api/generation/{jobId}/stream`

### Gate ✅
- [x] Shopify URL → brief generated < 25 seconds
- [x] SSE stream updates visible in browser
- [x] Brief persisted to DB (angles + hooks saved)
- [x] Inline editing persisted (BRIEF-08)
- [x] Per-angle/per-hook regeneration (BRIEF-09)
- [x] Langfuse traces per brief

---

## Phase 3 — Video Generation Pipeline ✅ Complete

### FAL.AI + ElevenLabs + HeyGen
- [x] `fal.queue.submit()` only — models: Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2
- [x] `fal_request_id` stored in variants table
- [x] ElevenLabs voiceover: `eleven_v3` / `eleven_flash_v2_5`, audio → R2
- [x] HeyGen Avatar API **v3** only — V2V lip-sync, output → R2

### asynq Worker Pipeline
- [x] Railway Redis TCP (never Upstash) for asynq
- [x] Task types: TypeScrape, TypeGenerate, TypePostprocess
- [x] Queue priorities: critical / default / low
- [x] Retry: 3 attempts, exponential backoff; dead letter queue

### Rust Postprocessor
- [x] Axum HTTP server — `POST /process`, `GET /health`
- [x] ffmpeg-next bindings (NOT `Command::new("ffmpeg")`)
- [x] 9:16 reframe, H.264 transcode, watermark, caption burn-in
- [x] Input/output via R2 presigned URLs
- [x] Callback to Go API on completion/failure

### Mux + Webhooks
- [x] Mux asset upload from R2 presigned URL
- [x] `mux_asset_id` + `mux_playback_id` in variants
- [x] Signed playback tokens (workspace-scoped HS256 JWT, 1-hour expiry)
- [x] `POST /webhooks/fal` — SHA256 verified → enqueue postprocess to critical
- [x] `POST /webhooks/mux` — updates variant Mux IDs, marks complete
- [x] `POST /internal/jobs/reconcile-stuck` — stale job recovery
- [x] `<MuxPlayer>` with signed token in dashboard

### Gate ✅
- [x] FAL → postprocess → Mux → playback pipeline wired end-to-end
- [x] Tier enforcement: starter=3, growth=10, agency=unlimited
- [x] "Generate All" E2E flow validated
- [x] Failed jobs retry ×3 and failed state surfaced in UI

---

## Phase 4 — Brand Kit & Export ⏳ Pending

### Brand Kit System
- [ ] Brand creation wizard (`(dashboard)/brand/new`)
- [ ] Logo upload → R2 presigned PUT (PNG/SVG, max 5MB)
- [ ] Brand color picker + CSS preview
- [ ] Intro/outro bumper upload (MP4/MOV, max 5s) → R2
- [ ] Custom font upload (TTF/OTF) → R2
- [ ] Tone of voice notes (300 chars → injected into Claude prompt)
- [ ] Multi-brand selector in sidebar
- [ ] Brand kit auto-applied on generation (logo + colors → Rust postprocessor payload)

### Export Engine
- [ ] `POST /v1/exports` → naming: `[Brand]_[Angle]_[Hook]_[Platform]_V[n]`
- [ ] Formats: MP4 1080p (all), MP4 4K (Agency+), GIF preview (Growth+)
- [ ] Platform exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16)
- [ ] Bulk ZIP download — Go server-side, R2 presigned URL (48h expiry)
- [ ] Export history in `exports` table with R2 key + download count
- [ ] Platform compliance check (safe zones, text size, duration)

### Gate
- [ ] Brand logo watermark on all generated videos
- [ ] Export downloads as correctly named MP4
- [ ] Bulk ZIP works for 10+ variants
- [ ] Platform compliance check rejects out-of-spec asset

---

## Phase 5 — Asset Library & Team ⏳ Pending

### Asset Library
- [ ] Variants grid view (`(dashboard)/library`) — filter: brand / angle / format / date
- [ ] Search by tag metadata (`asset_tags` table)
- [ ] Variant detail page (`(dashboard)/library/[variantId]`)
- [ ] Favorites / starring (`user_variant_stars` table)
- [ ] Archive vs Active (`archived_at` soft delete)
- [ ] Storage usage indicator per workspace

### Team & Collaboration
- [ ] Invite team member by email (Clerk org invitation)
- [ ] Role assignment: Admin / Member / Viewer
- [ ] Viewer role: read-only (`403` at API + UI controls hidden)
- [ ] Remove member (Clerk org member removal)
- [ ] Pending invite list with resend / revoke
- [ ] Seat count display (`(dashboard)/settings/team`)

### Gate
- [ ] Library shows all variants with correct metadata filters
- [ ] Viewer role cannot trigger generation (blocked at API + UI)
- [ ] Team invite flow: invite → email → accept → workspace access
- [ ] Archive removes from default view

---

## Phase 6 — Platform, Billing & Trial ⏳ Pending

### Stripe Integration
- [ ] Stripe products + prices: Starter $99 / Growth $149 / Agency $399
- [ ] `POST /v1/billing/checkout` → Stripe Checkout Session
- [ ] `POST /v1/billing/portal` → Stripe Customer Portal
- [ ] Webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] Stripe Meter API: `meter_event` per video generation
- [ ] Webhook updates `workspaces.plan_tier` + `workspaces.stripe_status`
- [ ] Idempotent webhook handlers (Stripe-Signature HMAC-SHA256 verified)

### Trial Flow
- [ ] 14-day trial on workspace creation (`trial_ends_at = created_at + 14 days`)
- [ ] Trial badge: "Trial — X days left"
- [ ] Day 10 in-app urgency banner
- [ ] Day 13 modal gate (dismiss once)
- [ ] Day 15: generation blocked (402 from Go API)
- [ ] 30-day data retention post-expiry
- [ ] Conversion emails: Day 5 / Day 10 / Day 15

### Tier Enforcement
- [ ] Starter blocked at 4th variant (upgrade CTA shown)
- [ ] Growth blocked at 11th variant
- [ ] Agency: unlimited
- [ ] Custom voice: Growth+ only
- [ ] 4K export: Agency only
- [ ] Custom avatar (V2V): Agency only

### Gate
- [ ] Stripe checkout E2E (test mode): trial → paid
- [ ] Webhook updates plan tier in DB within 5 seconds
- [ ] Day-15 lock activates automatically
- [ ] Stripe webhook rejects unsigned requests

---

## Phase 7 — V1 Polish, Observability & Launch ⏳ Pending

### Observability (V1 minimum)
- [ ] Sentry: `@sentry/nextjs` + Go SDK + Rust SDK
- [ ] Better Stack: log drain from Railway (structured JSON)
- [ ] PostHog: `posthog-js` with activation funnel events
- [ ] Langfuse: LLM cost per workspace, latency traces
- [ ] `traceparent` header propagated: Next.js → Go API → Go workers

### PostHog Events
- [ ] `user_signed_up` (role, plan, referrer)
- [ ] `brief_generated` (latency_ms, angle_count)
- [ ] `video_generation_started` (format, angle_type, tier, model)
- [ ] `video_generation_complete` (duration_s, model, success)
- [ ] `export_downloaded` (format, platform, variant_count)
- [ ] `trial_to_paid` (days_to_convert, plan)
- [ ] `variant_limit_hit` (tier, upgrade_shown)

### Security Hardening
- [ ] Upstash rate limiting: 1000 req/hr per workspace on public endpoints
- [ ] Input validation on all Go API handlers
- [ ] R2 presigned URLs expire in 15 minutes
- [ ] CORS locked to verified origins (no `*` in prod)
- [ ] RLS cross-workspace isolation test in CI

### QA Sign-Off
- [ ] E2E: sign up → brand → brief → video → export (< 15 min)
- [ ] URL scrape: Shopify, WooCommerce, App Store, custom landing page
- [ ] Tier limits enforced in staging at all 3 tiers
- [ ] Trial flow: Day 5/10/15 emails + Day-15 lock verified
- [ ] Load test: 10 concurrent generation jobs without failure

### Launch
- [ ] Staging fully green
- [ ] Production deploy checklist complete
- [ ] SLO defined (99.5% uptime, < 25s brief, < 180s video)
- [ ] On-call runbook written
- [ ] Launch comms ready

### Gate
- [ ] All QA sign-off items passing in staging
- [ ] Zero P0 bugs open
- [ ] Sentry < 1% error rate in staging
- [ ] Production deploy successful

---

## Phase 8 — Microservice Foundation ⏳ Post-Launch

### 8A — NATS JetStream Setup
- [ ] NATS JetStream 3-node cluster provisioned on Railway
- [ ] Streams created: `QVORA_PIPELINE`, `QVORA_SIGNALS`, `QVORA_DLQ`
- [ ] Consumer config: `AckExplicit`, `MaxDeliver=3`, exponential backoff
- [ ] `NATS_URL` + `NATS_CREDENTIALS` in Doppler
- [ ] NATS added to `docker-compose.yml` for local dev
- [ ] `pkg/messaging/` Go package with typed publish/subscribe helpers
- [ ] DLQ forwarding on `MaxDeliver` exhaustion
- [ ] NATS `/health` check in all consuming services
- [ ] NATS Surveyor dashboard deployed (consumer lag visibility)

### 8B — Service Extraction
- [ ] `src/services/ingestion/` scaffolded (Go module, Dockerfile, Railway deploy)
- [ ] `scrape_url` asynq handler migrated to `ingestion.scrape` NATS consumer
- [ ] ingestion-svc publishes `ingestion.complete` / `ingestion.failed`
- [ ] ingestion-svc deployed and E2E tested independently
- [ ] `src/services/brief/` scaffolded (Go module, Dockerfile, Railway deploy)
- [ ] Brief generation pipeline migrated to `brief.generate` NATS consumer
- [ ] brief-svc publishes `brief.complete` / `brief.failed`
- [ ] brief-svc deployed and E2E tested independently
- [ ] `src/services/asset/` scaffolded (Go module, Dockerfile, Railway deploy)
- [ ] Export assembly, brand CRUD, Mux URL generation migrated to asset-svc
- [ ] asset-svc publishes `asset.export.ready`
- [ ] asset-svc deployed and E2E tested independently
- [ ] `src/services/identity/` scaffolded (Go module, Dockerfile, Railway deploy)
- [ ] Clerk JWT validation + quota + Stripe meter migrated to identity-svc
- [ ] identity-svc exposes gRPC interface (pre-mTLS, plain text initially)
- [ ] API Gateway calls identity-svc gRPC instead of inline middleware
- [ ] Monolithic `src/services/worker/` no longer handles ingestion / brief / asset tasks

### 8C — Supabase Realtime Migration
- [ ] Supabase Realtime enabled in project settings
- [ ] `generation_jobs` table published to Realtime (WAL change feed)
- [ ] RLS policy on Realtime channel (org_id claim isolation)
- [ ] Frontend: Supabase Realtime subscription replaces `EventSource`
- [ ] All Temporal/worker activities update `generation_jobs` status in Supabase
- [ ] Go workers no longer write `job:{jobId}` keys to Upstash Redis
- [ ] `src/apps/web/src/app/api/generation/[jobId]/stream/route.ts` deleted
- [ ] Cross-org Realtime isolation verified in staging

### Gate (Phase 8)
- [ ] All 4 extracted services healthy on Railway
- [ ] Full E2E through extracted services (no regressions)
- [ ] NATS DLQ receives messages after 3 failed deliveries
- [ ] Supabase Realtime updates browser within 500ms of DB change
- [ ] Cross-org Realtime isolation confirmed
- [ ] Railway Redis TCP decommissioned (asynq fully migrated to NATS)

---

## Phase 9 — Temporal + gRPC + Multi-Provider Avatar ⏳ Post-Launch

### 9A — Temporal Setup
- [ ] Temporal OSS provisioned on Railway (or Temporal Cloud)
- [ ] Temporal schema migration applied to Supabase (or dedicated PG)
- [ ] Temporal Web UI accessible at internal URL
- [ ] `go.temporal.io/sdk` added to `media-orchestrator` module
- [ ] `VideoCreationWorkflow` implemented with all 6 activities
- [ ] Activity: `SelectVideoProvider` (fal semaphore check, model selection)
- [ ] Activity: `SubmitToFal` (fal.queue.submit, request_id → Redis)
- [ ] Activity: `WaitForFalCompletion` (waits for Temporal signal from webhook)
- [ ] Activity: `PostProcessVideo` (gRPC call to Rust postprocessor)
- [ ] Activity: `IngestToMux` (Mux Assets API, returns playback_id)
- [ ] Activity: `MarkJobComplete` (UPDATE generation_jobs status=READY)
- [ ] fal.ai webhook → NATS → Temporal signal bridge
- [ ] Mux webhook → NATS → Temporal signal bridge
- [ ] Parallel workflow: one workflow instance per angle per brief
- [ ] asynq `generation:video` tasks migrated to Temporal workflow triggers
- [ ] Temporal Worker deployed as standalone Railway service (media-orchestrator)

### 9B — gRPC Internal Communication
- [ ] `proto/` directory in `src/packages/proto/`
- [ ] `proto/identity/v1/identity.proto` — IdentityService defined
- [ ] `proto/postprocess/v1/postprocess.proto` — PostProcessService with streaming RPC
- [ ] Go stubs generated: `protoc --go_out=. --go-grpc_out=.`
- [ ] Rust stubs generated: `tonic-build` in `build.rs`
- [ ] API Gateway → identity-svc: gRPC quota check replaces inline middleware
- [ ] media-orchestrator → media-postprocessor: gRPC streaming replaces HTTP `POST /process`
- [ ] mTLS certificates generated (Doppler: `GRPC_CERT`, `GRPC_KEY`, `GRPC_CA`)
- [ ] mTLS configured in gRPC dial options for all services
- [ ] gRPC connection latency measured (target: < 10ms P99 internal)

### 9C — Multi-Provider Avatar
- [ ] `AvatarProvider` interface defined in Go
- [ ] `HeyGenV3Provider{}` refactored into interface (existing logic)
- [ ] `TavusProvider{}` implemented (Tavus v2 API)
- [ ] Provider registry: `map[string]AvatarProvider` with factory
- [ ] Selection logic: tier + volume → HeyGen (quality) or Tavus (cost)
- [ ] Temporal activity: `CreateAvatarVideo` uses provider registry
- [ ] Automatic fallback: HeyGen 429 → retry with Tavus
- [ ] `generation_jobs.avatar_provider` column added (migration)
- [ ] Tavus v2 end-to-end test: lip-sync video produced in staging

### Gate (Phase 9)
- [ ] Temporal workflow visible in Web UI for every generation job
- [ ] Workflow survives Go worker restart mid-execution (durability test)
- [ ] Workflow retries correctly on simulated fal.ai timeout
- [ ] gRPC quota check working (API Gateway → identity-svc)
- [ ] gRPC streaming postprocess working with progress events
- [ ] mTLS connections rejected without correct certificate
- [ ] HeyGen v3 still works through new provider interface
- [ ] Tavus v2 produces lip-sync video end-to-end

---

## Phase 10 — V2 Signal Loop & Intelligence ⏳ Post-Launch

### 10A — Signal Ingestion
- [ ] Migration: `ad_accounts` table
- [ ] Migration: `video_performance_events` table (append-only, no UPDATE ever)
- [ ] Migration: `creative_scores` materialized view (7-day rolling window)
- [ ] Migration: `assets.platform_ad_id` column (Meta / TikTok / Google ad ID)
- [ ] `signal-svc` deployed as standalone Railway service
- [ ] Meta Ads API OAuth flow (per workspace, PKCE)
- [ ] TikTok Ads API OAuth flow (per workspace)
- [ ] Google Ads API OAuth flow (per workspace)
- [ ] NATS consumer: `signal.sync` → fetch metrics → INSERT events
- [ ] NATS consumer: `signal.recommend` → run fatigue detection
- [ ] NATS consumer: `signal.gdpr.cleanup` → DELETE events older than 90 days
- [ ] NATS scheduler: publish `signal.sync` every 6 hours
- [ ] `creative_scores` view refresh: scheduled NATS message hourly
- [ ] `assets.predicted_ctr` + `assets.predicted_vtr` columns added

### 10B — Creative Scoring Service
- [ ] `src/services/scoring/` scaffolded (Python 3.12, FastAPI, uvicorn, Pydantic v2)
- [ ] Dockerfile: `python:3.12-slim`, multi-stage, Railway deploy
- [ ] `POST /score` endpoint implemented (rule-based V1)
- [ ] Rules: hook clarity score, CTA presence, pacing, safe zone compliance
- [ ] Reads `creative_scores` view via Supabase REST API (httpx async)
- [ ] asset-svc calls scoring-svc after video marked READY
- [ ] Score stored on `assets.predicted_ctr`, `assets.predicted_vtr`
- [ ] `assets.score_reasoning` TEXT column added (human-readable explanation)
- [ ] scikit-learn model training pipeline (when org >= 50 events)
- [ ] Model stored in R2 per org (`scoring-models/<org_id>/latest.pkl`)
- [ ] A/B: rule-based vs ML for orgs with < 50 events
- [ ] OTel instrumentation: `opentelemetry-sdk` + `openllmetry`

### 10C — Feedback Loop into Brief Generation
- [ ] brief-svc: fetch exemplars from `creative_scores` on brief creation
- [ ] Filter: same org, same product_category, last 30 days, CTR > org average
- [ ] Exemplar injection threshold: org >= 20 `video_performance_events`
- [ ] Claude prompt updated: few-shot examples with performance context
- [ ] Langfuse: tag briefs with `exemplar_injected: true/false`
- [ ] Insights dashboard (`(dashboard)/insights`) — new route
- [ ] Chart: creative score ranking per brief (sort by `predicted_ctr`)
- [ ] Chart: historical CTR / VTR / ROAS trend per video (7-day / 30-day)
- [ ] Chart: top angle type and hook pattern per org
- [ ] Fatigue detection: flag assets with declining CTR over 14 days
- [ ] Regen suggestion CTA: "This angle is fatiguing — generate 3 new variants?"

### Gate (Phase 10)
- [ ] Meta Ads OAuth connected for test workspace
- [ ] Metrics ingested into `video_performance_events` after first sync
- [ ] `creative_scores` view refreshes hourly with correct aggregations
- [ ] scoring-svc returns score for new video within 2s
- [ ] Score visible on variant detail page
- [ ] ML model trains on test data (50+ events) and returns predictions
- [ ] Brief with exemplar injection shows better angle quality (manual review)
- [ ] Insights dashboard renders correctly for workspace with > 20 events
- [ ] Fatigue detection flags declining assets with regen CTA

---

## Environment Variables Reference

### Phase 0–3 (Current)
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

# API / Workers
DATABASE_URL
CLERK_SECRET_KEY
INTERNAL_API_KEY
MUX_TOKEN_ID
MUX_TOKEN_SECRET
MUX_WEBHOOK_SECRET
FAL_WEBHOOK_SECRET
RAILWAY_REDIS_URL       # Phase 8: decommissioned
UPSTASH_REDIS_REST_URL
UPSTASH_REDIS_REST_TOKEN
FAL_KEY
ELEVENLABS_API_KEY
HEYGEN_API_KEY
RUST_POSTPROCESS_URL
API_BASE_URL

# Postprocessor
R2_ENDPOINT
R2_BUCKET
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
MUX_ACCESS_TOKEN
MUX_SECRET_TOKEN
```

### Phase 6 additions
```
STRIPE_SECRET_KEY
STRIPE_WEBHOOK_SECRET
STRIPE_STARTER_PRICE_ID
STRIPE_GROWTH_PRICE_ID
STRIPE_AGENCY_PRICE_ID
STRIPE_METER_VIDEO_GEN_ID
```

### Phase 7 additions
```
SENTRY_DSN                  # all services
POSTHOG_API_KEY
LANGFUSE_SECRET_KEY
LANGFUSE_PUBLIC_KEY
BETTERSTACK_SOURCE_TOKEN
```

### Phase 8 additions
```
NATS_URL
NATS_CREDENTIALS            # NATS NKey credentials file
```

### Phase 9 additions
```
TEMPORAL_HOST_URL
TEMPORAL_NAMESPACE
TEMPORAL_TLS_CERT
TEMPORAL_TLS_KEY
TAVUS_API_KEY               # Avatar multi-provider
GRPC_CERT                   # mTLS certificates
GRPC_KEY
GRPC_CA
```

### Phase 10 additions
```
META_APP_ID
META_APP_SECRET
TIKTOK_APP_ID
TIKTOK_APP_SECRET
GOOGLE_ADS_DEVELOPER_TOKEN
GOOGLE_ADS_CLIENT_ID
GOOGLE_ADS_CLIENT_SECRET
```
