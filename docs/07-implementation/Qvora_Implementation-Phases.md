# Qvora — Implementation Phases
**Version:** 2.1 | **Updated:** April 16, 2026 | **Status:** Phase 0, 1, 2, 3 Complete

---

## Phase Overview

| Phase | Name | Duration | Status |
|---|---|---|---|
| Phase 0 | Foundation & Infrastructure | Week 1 | ✅ Complete |
| Phase 1 | Core Data Layer | Week 2 | ✅ Complete |
| Phase 2 | URL Ingestion & Brief Engine | Weeks 3–4 | ✅ Complete |
| Phase 3 | Video Generation Pipeline | Weeks 5–7 | ✅ Complete |
| Phase 4 | Brand Kit & Export | Week 8 | ⏳ Pending |
| Phase 5 | Asset Library & Team | Week 9 | ⏳ Pending |
| Phase 6 | Platform, Billing & Trial | Week 10 | ⏳ Pending |
| Phase 7 | Polish, Observability & Launch | Week 11 | ⏳ Pending |

**V1 Total: 11 weeks | V2 (Signal loop, Ad Connector): post-launch icebox**

---

## Phase 0 — Foundation & Infrastructure ✅ Complete

**Goal:** Deployable skeleton — every service starts, CI passes, secrets managed.

### Deliverables
- Turborepo monorepo: `src/apps/`, `src/packages/`, `src/services/`, `src/ai/`
- pnpm workspaces, Biome (lint+format), Lefthook (pre-commit hooks), `.nvmrc` Node 22 LTS
- GitHub Actions: `ci.yml` (Turbo lint + typecheck), path-filtered deploy workflows per service
- Infrastructure provisioned: Vercel, Railway (api/worker/postprocess/Redis), Supabase, Upstash, Cloudflare R2, Mux, Clerk, Doppler
- `docker-compose.yml` — Postgres + Redis ×2 + all services for local dev
- Service scaffolds: Next.js 15 App Router, Go Echo v4, Go asynq worker, Rust Axum postprocessor
- `src/packages/ui` (shadcn), `src/packages/types`, `src/packages/config`

### Key Decisions
| Decision | Rationale |
|---|---|
| Two Redis instances | Railway Redis (TCP) = asynq BLPOP; Upstash (HTTP) = cache/rate-limit — never substitutable |
| Biome over ESLint+Prettier | Single tool, faster, zero config drift |
| Doppler for secrets | All env vars in Doppler; never committed to repo |

---

## Phase 1 — Core Data Layer ✅ Complete

**Goal:** Schema, auth, and API skeleton wired end-to-end.

### Deliverables
- Supabase migrations in `supabase/migrations/` only (not `services/api/db/`)
- Tables: workspaces, users, brands, briefs, brief_angles, brief_hooks, jobs, variants, asset_tags, exports
- `jobs.status` CHECK: `queued, scraping, generating, postprocessing, complete, failed`
- `plan_tier` CHECK: `starter, growth, agency`
- RLS policies on all tables — workspace isolation via `app.org_id` session var
- sqlc codegen (`src/services/api/internal/db/`) from `src/services/api/db/queries/queries.sql`
- Echo v4 API: request ID, logger, CORS, recover, rate-limiter middleware
- Clerk JWT middleware — extracts `org_id` + `org_role` from top-level JWT claims (not `app_metadata`)
- Tier enforcement middleware — keys: `starter/growth/agency`
- Route groups: `/v1/jobs`, `/v1/briefs`, `/v1/workspaces`, `/v1/assets`, `/v1/exports`, `/v1/variants`
- tRPC: `initTRPC.create()` with Clerk context; `appRouter` routers: briefs, assets, exports, projects, brands, jobs, org
- `ClerkProvider` + `TRPCProvider` + `QueryClientProvider` in root `layout.tsx`

### Key Decisions
| Decision | Rationale |
|---|---|
| RLS via `app.org_id` session var | Data-layer multi-tenancy; eliminates app-layer auth bugs |
| Migrations in `supabase/migrations/` only | `services/api/db/` holds sqlc query definitions only — no migrations |
| `org_id` from top-level JWT | Clerk puts org claims at top-level; `app_metadata` path is wrong |
| Plan tiers: starter/growth/agency | Agency = unlimited variants; no "scale" tier |

---

## Phase 2 — URL Ingestion & Brief Engine ✅ Complete

**Goal:** URL in → structured brief out → persisted to DB.

### Deliverables
- Modal Playwright scraper (`POST /scrape`): JS SPA headless render, extracts product fields, 24-hour Upstash cache by URL hash
- `POST /v1/briefs` Go handler → enqueues asynq `scrape_url` task
- Job transitions: `queued → scraping → generating` (no 'briefing' step)
- Brief generation in **Next.js BFF tRPC only** (not Go worker):
  - `generateObject()` ×2 in `src/apps/web/src/server/trpc/routers/briefs.ts`
  - GPT-4o → `productExtractionSchema` (structured product data)
  - Claude Sonnet 4.6 → `anglesGenerationSchema` (3–5 angles + hooks)
  - `OPENAI_API_KEY` + `ANTHROPIC_API_KEY` guards
- Prompt file: `src/ai/prompts/angles-gen.prompt.ts` — imported via `@qvora/prompts/*` tsconfig alias
- Brief + angles + hooks persisted to Go API (`briefs`, `brief_angles`, `brief_hooks` tables)
- SSE: standalone Route Handler `src/apps/web/src/app/api/generation/[jobId]/stream/route.ts`
  - `ReadableStream` — NOT tRPC subscription
  - Proxies to Go API `GET /api/v1/jobs/:id/stream`
- Frontend routes: `(dashboard)/briefs/page.tsx`, `(dashboard)/briefs/[id]/page.tsx`

### Completed Gate Closures
- **BRIEF-08** — Inline brief editing persisted to DB via `PUT /api/v1/briefs/:id/content`
- **BRIEF-09** — Per-angle and per-hook regenerate actions in brief detail UI
- Single-angle regenerate latency path instrumented with elapsed-time return for <10s target tracking
- Langfuse tracing integrated per brief create/update/regenerate action

### Key Decisions
| Decision | Rationale |
|---|---|
| Brief generation in web tRPC, not Go worker | Direct Vercel AI SDK call; avoids queue latency; no OPENAI_API_KEY in worker |
| Two `generateObject()` calls | GPT-4o for structured product extraction; Claude Sonnet 4.6 for creative angle generation |
| SSE as Route Handler (not tRPC) | tRPC subscriptions require WebSocket server; standalone Route Handler is simpler and Vercel-compatible |
| `@qvora/prompts/*` alias | Shared prompt files under `src/ai/prompts/` — imported across apps without relative path hacks |

---

## Phase 3 — Video Generation Pipeline ✅ Complete

**Goal:** Brief → video → Mux playback, fully async.

### Deliverables
- FAL.AI async queue: `fal.queue.submit()` (never `fal.subscribe()`) — models: Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2
- `fal_request_id` tracked: `generate.go` → `patchVariantFalRequestID()` → `PATCH /api/v1/variants/:id/fal-request` → DB
- ElevenLabs voiceover: `eleven_v3` (quality) / `eleven_flash_v2_5` (preview); script from brief hook; audio → R2
- HeyGen Avatar API **v3 only** (`developers.heygen.com`): V2V lip-sync (video + audio → lip-sync job); output → R2
- asynq pipeline: TypeScrape / TypeGenerate / TypePostprocess; queues: critical / default / low; retry ×3 exponential backoff
- Rust postprocessor (Axum `POST /process`): ffmpeg-next bindings (not `Command::new("ffmpeg")`), 9:16 reframe, H.264 transcode, watermark overlay, caption burn-in; input/output via R2 presigned URLs
- Mux: HLS upload from R2 presigned URL; `mux_asset_id` + `mux_playback_id` in variants; signed playback tokens (workspace-scoped HS256 JWT, 1-hour, `sub` = workspaceID)
- Webhooks: `POST /webhooks/fal` (completion → enqueue postprocess to critical), `POST /webhooks/mux` (asset ready → mark complete)
- Callbacks: `POST /internal/postprocess/callback`, `POST /internal/jobs/reconcile-stuck`
- `<MuxPlayer>` in dashboard with signed token
- Tier enforcement active: starter=3 variants/angle, growth=10, agency=unlimited

### Key Decisions
| Decision | Rationale |
|---|---|
| `fal.queue.submit()` only | Non-blocking; `fal.subscribe()` blocks worker threads under load |
| FAL request ID stored in DB | Enables webhook matching and postprocess enqueue by variant |
| Rust for postprocessing | ffmpeg-next bindings = no subprocess; CPU-bound work isolated on its own Railway service |
| Mux signed tokens scoped to workspace | Security: each workspace gets its own HS256 JWT, 1-hour expiry |
| HeyGen v3 only | v3 is the only version with V2V lip-sync; v4 does not exist |
| Railway Redis TCP for asynq | Upstash HTTP doesn't support BLPOP; asynq requires persistent TCP connection |

---

## Phase 4 — Brand Kit & Export ⏳ Pending

**Goal:** Per-workspace brand applied to all videos; exports named and downloadable.

### Planned Deliverables
- Brand creation wizard (`(dashboard)/brand/new`)
- Logo upload → R2 presigned PUT; brand colors; intro/outro bumper (MP4/MOV ≤5s); custom font (TTF/OTF)
- Tone of voice notes (300 chars → fed into LLM prompt)
- Multi-brand selector in sidebar; brand auto-applied on generation (passed to Rust postprocessor)
- `POST /v1/exports` → named package: `[Brand]_[Angle]_[Hook]_[Platform]_V[n]`
- Formats: MP4 1080p, MP4 4K (Agency+), GIF preview
- Platform exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16)
- Bulk ZIP download (Go server-side, R2 presigned URL)
- Export history in DB + R2 key; platform compliance check (safe zones, text size, duration)

---

## Phase 5 — Asset Library & Team ⏳ Pending

**Goal:** Browsable asset library; team roles enforced end-to-end.

### Planned Deliverables
- Variants grid view (`(dashboard)/library`) — filter by brand / angle / format / date
- Search by tag metadata (`asset_tags` table); favorites (`user_variant_stars` table); soft delete (archive)
- Storage usage indicator per workspace
- Clerk org invites: Admin / Member / Viewer roles
- Viewer role: read-only — blocked at API (`403`) + UI
- Seat count display (`(dashboard)/settings/team`)

---

## Phase 6 — Platform, Billing & Trial ⏳ Pending

**Goal:** Stripe checkout live; trial enforced; tier limits activated.

### Planned Deliverables
- Stripe products: Starter $99 / Growth $149 / Agency $399
- `POST /v1/billing/checkout` → Stripe checkout session; `POST /v1/billing/portal`
- Webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted` — updates `workspaces.plan_tier`
- Stripe-Signature HMAC-SHA256 verification; idempotent handlers
- Trial: 7-day (`trial_ends_at = created_at + 7 days`); badge; Day 5 banner; Day 7 modal; Day 8 generation blocked (402)
- 30-day data retention post-expiry; conversion emails Day 3/6/8
- Tier gates: Custom voice Growth+; 4K export Agency; Custom avatar Agency

---

## Phase 7 — Polish, Observability & Launch ⏳ Pending

**Goal:** All four services instrumented; E2E QA passed; production deployed.

### Planned Deliverables
- Sentry: `@sentry/nextjs` + Go SDK + Rust SDK across all 4 services
- Better Stack: log drain from Railway (structured logs + trace IDs)
- PostHog: activation funnel, trial conversion, feature adoption events
- Langfuse: LLM cost per workspace, prompt versions, latency traces
- PostHog events: `user_signed_up`, `brief_generated`, `video_generation_started/complete`, `export_downloaded`, `trial_to_paid`, `variant_limit_hit`
- Security: Upstash rate limiting 60 req/min per workspace; CORS locked to verified origins; R2 presigned URLs expire 15 min; RLS cross-workspace CI test
- QA sign-off: full E2E < 15 min; all 3 tier limits in staging; trial Day 3/6/8 emails + Day 8 lock
- Load test: 10 concurrent generation jobs
- Production deploy; SLO / error budget defined; on-call runbook

---

## Architecture (Locked — All Phases)

| Concern | Decision |
|---|---|
| Brief generation | Next.js BFF tRPC — `generateObject()` ×2; never Go worker |
| SSE stream | Standalone Route Handler — never tRPC subscription |
| Asynq queue Redis | Railway TCP only — never Upstash (no BLPOP support) |
| Cache / rate-limit Redis | Upstash HTTP only — never Railway Redis |
| FAL submission | `fal.queue.submit()` only — never `fal.subscribe()` |
| HeyGen version | v3 only (`developers.heygen.com`) — v4 does not exist |
| Plan tiers | `starter / growth / agency` — no "scale" tier |
| Prompt files | `src/ai/prompts/` — imported via `@qvora/prompts/*` alias |
| Migrations | `supabase/migrations/` only — never `services/api/db/` |
| Source root | All source under `src/` — no top-level `ai/`, `services/`, `apps/` |
| Postprocessor | ffmpeg-next bindings only — never `Command::new("ffmpeg")` |
| Mux tokens | Workspace-scoped HS256 JWT, 1-hour expiry, `sub` = workspaceID |

---

## V2 — Signal Loop (Icebox, Post-Launch)

| Feature | Notes |
|---|---|
| Qvora Signal | Ad account connector → variant performance metrics → LLM learns best angles |
| Ad Account Connector | Meta + TikTok OAuth, spend/impression/click/conversion sync |
| Temporal workflow | Scheduled signal ingestion loop |
| Variant scoring | ML: which angles perform best per industry vertical |
| Smart exports | Auto-format for TV, Pinterest, YouTube Ads |
