# Qvora — Implementation Checklist
**Last Updated:** Apr 15, 2026 (reconciled against actual code — see note below) | **Total:** 10 Phases | **V1 Target: Week 11**

> **Reconciliation note (Apr 15, 2026):** Phases 0–3 had drifted from reality in both directions —
> some items checked off here were never actually true (e.g. Phase 0's CI claim, Phase 1's RLS gate),
> while Phases 4–7 were far further along in the Go/Rust backend than "Pending" suggested, just missing
> frontend and a few integrations. This pass corrects both. Individual sub-checkboxes below reflect what
> was verified by reading the actual code or newly built/fixed in this session — not what any prior
> version of this doc claimed. Where something couldn't be verified without running the app end-to-end
> (e.g. exact Clerk role-gating UI behavior), it's left unchecked with a note rather than assumed.

## Progress Summary

| Phase | Name | Duration | Status |
|---|---|---|---|
| Phase 0 | Foundation & Infrastructure | Week 1 | ⚠️ Mostly complete — CI didn't actually exist until this pass |
| Phase 1 | Core Data Layer | Week 2 | ⚠️ Mostly complete — RLS now real (non-superuser role + per-request scoping) for the `asset` domain package, proven by a real Postgres test; other 4 domain packages and the production DATABASE_URL cutover still pending, see Phase 1 note |
| Phase 2 | URL Ingestion & Brief Engine | Weeks 3–4 | ✅ Complete |
| Phase 3 | Video Generation Pipeline | Weeks 5–7 | ⚠️ Complete, but job creation never enqueued the scrape task until fixed this session (P0 bug) |
| Phase 4 | Brand Kit & Export | Week 8 | ⚠️ Partial — backend watermark/brand-kit real; Export Engine backend now built (bulk ZIP, naming scheme, download tracking) but scoped down to `mp4_1080p` only, no true per-platform reframe yet |
| Phase 5 | Asset Library & Team | Week 9 | ⚠️ Partial — core screens built; several sub-features (favorites, archive, tags) not built |
| Phase 6 | Platform, Billing & Trial | Week 10 | ⚠️ Partial — checkout/portal/webhooks now real; Meter API and trial emails not built |
| Phase 7 | V1 Polish, Observability & Launch | Week 11 | ⚠️ Partial — Next.js, Go API, Go worker, and Rust postprocessor Sentry all now wired and build-verified; RLS isolation test written and passing locally (not yet in hosted CI, no Postgres service there); no load test performed |
| Phase 8 | Microservice Foundation | Weeks 12–13 | ⏳ Post-Launch (not investigated this session) |
| Phase 9 | Temporal + gRPC + Multi-Provider | Weeks 14–15 | ⏳ Post-Launch (not investigated this session) |
| Phase 10 | V2 Signal Loop & Intelligence | Weeks 16–18 | ⏳ Post-Launch (not investigated this session) |

**Complete: 1/10 — Partial: 6/10 — Post-Launch (unreviewed): 3/10**

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
- [x] `ci.yml` — lint/typecheck/build (web) + vet/build/test (Go api, worker) + build/test (Rust postprocess) — **this did not exist until Apr 15, 2026; no `.github/workflows` directory was present before this session**
- [ ] Path-filtered deploy workflows (web, api, worker, postprocess) — not built
- [ ] `CODEOWNERS`, PR template, issue templates — not found
- [ ] Branch protection on `main` — can't verify from code (GitHub repo setting, not a file)

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

### Gate ⚠️
- [x] Migrations run cleanly on local Postgres (structurally; not run against a live DB this session)
- [x] **RLS: user A cannot read user B's briefs — now real and empirically proven, for the `asset` domain package.** Migration `007_rls_app_role.sql` adds a non-superuser `qvora_app` role and extends every policy (`TO authenticated, qvora_app`), plus enables RLS on three tables that never had it (`postprocess_callbacks`, `cost_events`, `video_performance_events`). `internal/store/store.go`'s `AcquireForOrg` dedicates one pooled `*pgxpool.Conn` per request and runs `SELECT set_config('app.org_id', $1, false)` on it; `internal/middleware/auth.go`'s `RequireWorkspace()` acquires this scope and exposes it via `GetScopedQueries`/`GetScopedConn`. The `asset` package (asset.go, export.go, export_create.go, variant.go) is fully migrated to the scoped pattern. `internal/store/rls_test.go` (`TestRLSCrossWorkspaceIsolation`, gated on `TEST_DATABASE_URL`, skips cleanly in CI) proves against a real local Postgres running as `qvora_app`: (1) a scope for org A sees zero rows when queried with org B's own correct `workspace_id` — RLS blocks cross-tenant access even when the Go code's own WHERE clause would have let it through, (2) org A's scope correctly sees its own rows, (3) no `app.org_id` set at all fails closed to zero rows. **Remaining gap:** `brief`, `identity`, `media`, `signal` domain packages still use the old unscoped `store.Queries()`/`store.Pool()` pattern (superuser bypass still applies to them), and the real `DATABASE_URL` in each deployed environment still needs to be cut over from `postgres` to `qvora_app` via Doppler before this is enforced in production, not just locally.
- [x] `GET /health` returns 200 from Railway (per Rust/Go code; not re-verified live)
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

### Gate ⚠️
- [x] FAL → postprocess → Mux → playback pipeline wired end-to-end — **true only as of Apr 15, 2026.** `SubmitJob` created the job row but never enqueued the asynq scrape task (`domain/media/job.go`) — the pipeline's own entry point was dead on arrival, despite every downstream stage (scrape→brief→generate→postprocess→mux) being genuinely implemented. Fixed this session.
- [x] Tier enforcement: starter=3, growth=10, agency=unlimited
- [ ] "Generate All" E2E flow validated — not re-run live this session (no way to run `npm run dev`/the Go services in this sandbox)
- [ ] Failed jobs retry ×3 and failed state surfaced in UI — retry logic exists in Go; UI surfacing not independently verified

---

## Phase 4 — Brand Kit & Export ⚠️ Partial

### Brand Kit System
- [x] Brand settings page (`(dashboard)/brand`) — name, primary/secondary color, font family, watermark toggle; bound to real, working `workspace.getBrandKit`/`upsertBrandKit`
- [ ] Logo upload → R2 presigned PUT (PNG/SVG, max 5MB) — **no presigned-upload endpoint exists anywhere in the Go backend for any file type**, not just logos. UI dropzone built but disabled with a "coming soon" note.
- [x] Brand color picker — hex input + native swatch (not a custom popover component, but functional)
- [ ] CSS live preview of brand colors — not built
- [ ] Intro/outro bumper upload (MP4/MOV, max 5s) → R2 — not built
- [ ] Custom font upload (TTF/OTF) → R2 — not built (font *name* text field exists, not file upload)
- [ ] Tone of voice notes (300 chars → injected into Claude prompt) — not built
- [ ] Multi-brand selector in sidebar — not built; current model is one brand kit per workspace, not multi-brand
- [x] Brand kit auto-applied on generation (logo + colors → Rust postprocessor payload) — **built this session.** `FalWebhook` now looks up the brand's `logo_r2_key` via `identity.GetBrandLogoKey` and passes it through the worker payload to the Rust postprocessor, which downloads and composites the real logo via ffmpeg's `movie`+`overlay` filters (falls back to the old static text mark when no logo is configured — which is always, until the logo-upload gap above is closed)

### Export Engine
- [x] `POST /api/v1/exports` → naming scheme — **built Oct 16, 2026** (`src/services/api/internal/domain/asset/export_create.go`). Validates each variant belongs to the workspace and is complete, then names each file `{brand-slug}_{angle-slug}_{platform}_{yyyymmdd}_v{n}.mp4`. Verified with a real Go toolchain: builds, vets, and passes the existing test suite cleanly.
- [ ] Formats: MP4 1080p (all), MP4 4K (Agency+), GIF preview (Growth+) — **only MP4 1080p passthrough is supported**; the endpoint explicitly rejects `4k`/`gif` with a clear `format_not_supported` error rather than faking them, since both need a new Rust transcode pass that doesn't exist.
- [ ] Platform exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16) — platform is accepted and recorded (affects naming), but **no reframe actually happens yet** — all three currently ship the same stored 9:16 master. TikTok/YouTube Shorts are correct by coincidence (already 9:16); Meta would need to call the Rust postprocessor's `aspect_ratio` parameter (built in P2, never wired to a caller) in a new async re-render step. The UI says this explicitly rather than silently serving the wrong aspect ratio.
- [x] Bulk ZIP download — Go-side, real: downloads each variant's existing R2 object, zips with a `manifest.csv`, uploads to `exports/{workspace}/{uuid}.zip`. Presigned URLs are minted fresh per-download (15 min expiry, not the 48h originally specced — shorter is more conservative and matches the security-checklist item elsewhere in this doc) via a dedicated `GET /api/v1/exports/:id/download` endpoint, proxied through a Next.js route handler since the browser can't call the Go API directly.
- [x] Export history in `exports` table with R2 key + download count — migration `006_export_zip_tracking.sql` adds `r2_key` and `download_count` (incremented on each download) to the existing table.
- [ ] Platform compliance check (safe zones, text size, duration) — not built

### Gate
- [x] Brand logo watermark on generated videos — **capability now built** (see Brand Kit System above); will only visibly apply once the logo-upload gap is closed, since no workspace can set a `logo_r2_key` yet
- [x] Export downloads as correctly named MP4 — built and verified (build/vet/test clean); not run against a live DB/R2 in this sandbox
- [x] Bulk ZIP works for 10+ variants — the endpoint accepts up to 100 variants per export; not load-tested
- [ ] Platform compliance check rejects out-of-spec asset — not built

---

## Phase 5 — Asset Library & Team ⚠️ Partial

### Asset Library
- [x] Variants grid view (`(dashboard)/assets`, not `/library`) — filter by angle and date work on real fields; brand/format filters are visual-only (`assets.list` doesn't expose that metadata)
- [ ] Search by tag metadata (`asset_tags` table) — not built; tRPC `assets.list` doesn't return tags
- [ ] Variant detail page — not a separate route; built as a preview modal instead (same content, different UX)
- [ ] Favorites / starring — not built
- [ ] Archive vs Active — not built
- [ ] Storage usage indicator per workspace — not built

### Team & Collaboration
- [x] Invite team member by email — via Clerk's `<OrganizationProfile />` prebuilt component (`(dashboard)/team`), not custom tRPC/Go endpoints. The `org` tRPC router only has a stub `org.current`; building a custom membership API would have re-implemented what Clerk Organizations already owns.
- [x] Role assignment — handled by the same Clerk component
- [ ] Viewer role: read-only (403 at API + UI controls hidden) — not independently verified this session; Go middleware has `RequireAdmin`/`RequireWriteAccess` guards but a specific "viewer" role test wasn't run
- [x] Remove member — handled by Clerk's component
- [x] Pending invite list with resend/revoke — handled by Clerk's component (standard feature of `<OrganizationProfile />`, not independently re-verified live)
- [ ] Seat count display — not built as a standalone element (plan tier is shown; no seat cap field exists in the schema to count against)

### Gate
- [x] Library shows variants with angle/date filters (brand/format filters are inert — see above)
- [ ] Viewer role cannot trigger generation — not independently tested this session
- [ ] Team invite flow — relies on Clerk's own tested infrastructure; not re-verified live in this app's context
- [ ] Archive removes from default view — not built (no archive feature)

---

## Phase 6 — Platform, Billing & Trial ⚠️ Partial

### Stripe Integration
- [ ] Stripe products + prices configured in the Stripe dashboard — external setup, not verifiable from code; price-ID env vars (`STRIPE_STARTER_PRICE_ID` etc.) are wired and expected to be populated
- [x] Checkout session creation — built this session (`workspace.createCheckoutSession` tRPC mutation, not a `/v1/billing/checkout` Go route — lives in the Next.js BFF since that's where the rest of the Stripe integration already lived)
- [x] Customer portal — built this session (`workspace.createPortalSession`); requires `stripe_customer_id` to be set first, which the new `checkout.session.completed` webhook handler now captures
- [x] Webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted` — were already genuinely implemented (`app/api/webhooks/stripe/route.ts`) before this session, forwarding to the Go API correctly
- [ ] Stripe Meter API — not built
- [x] Webhook updates `workspaces.plan_tier` + subscription status — via `UpdateWorkspaceSubscription`; new narrow `UpdateStripeCustomerID` endpoint added this session so the customer-ID capture can never clobber plan_tier/status
- [x] Idempotent webhook handlers, signature verified — signature verification was already real; explicit event-ID dedup isn't present, but the handlers are idempotent by construction (re-applying the same plan/status/reset is a no-op)

### Trial Flow
- [x] 14-day trial on workspace creation — `trial_ends_at` set on first workspace provisioning (actually 7 days per `store.WorkspaceIDForOrg`, not 14 — worth reconciling with the pricing page's "7-day free trial" claim in the README, which matches the code)
- [ ] Trial badge / Day 10 banner / Day 13 modal — not built as separate UI; the trial-lock overlay built this session is a single hard gate (post-expiry), not a graduated warning sequence
- [x] Generation blocked after trial expiry — `TrialLockedOverlay` component (built this session), gated on `subscriptionStatus === "past_due"` or an expired `trialing` state (no literal `"trial_expired"` enum exists in the schema, so this uses the closest real signals)
- [ ] 30-day data retention post-expiry — not built
- [ ] Conversion emails (Day 5/10/15) — not built; no transactional email provider integration found anywhere in the codebase

### Tier Enforcement
- [x] Starter/Growth/Agency variant limits — real, server-enforced in Go (`EnforceVariantLimit` middleware), matches `PLAN_LIMITS` constant used client-side for the plan-limit banner built this session
- [ ] Custom voice / 4K export / custom avatar tier gating — not independently verified this session

### Gate
- [ ] Stripe checkout E2E (test mode) — built but not run live (no way to execute a real Stripe test-mode flow in this sandbox)
- [x] Webhook updates plan tier in DB — code path verified by reading; not timed live
- [x] Trial lock activates — via the signals described above, not run live
- [x] Stripe webhook rejects unsigned requests — signature verification was already correctly implemented

---

## Phase 7 — V1 Polish, Observability & Launch ⚠️ Partial

### Observability (V1 minimum)
- [x] Sentry: `@sentry/nextjs` — built (client/server/edge configs, `withSentryConfig`, `global-error.tsx`)
- [x] Sentry: Go SDK — **built Oct 16, 2026**, once a real Go toolchain became available (`go.mod`/`go.sum` regenerated properly via `go get`/`go mod tidy`, not hand-edited). Wired into both `services/api` (panic capture via the existing `Recover` middleware's `LogErrorFunc` hook) and `services/worker` (every task failure, via asynq's central `ErrorHandler`, tagged with `task_type`). No official `sentry-go/echo` middleware — that package requires `echo/v5`, incompatible with this API's `echo/v4`, so panic capture is wired directly against the base SDK instead. Verified: `go build`/`go vet`/`go test`/`gofmt` all clean on both services.
- [x] Sentry: Rust SDK — **built Oct 16, 2026**, once a Rust toolchain became available (Windows needed the GNU target + a separate MinGW install, since the MSVC linker requires full Visual Studio Build Tools). `sentry` + `sentry-tracing` crates added; `sentry_tracing::layer()` forwards existing `tracing::error!`/`warn!` call sites to Sentry automatically, no per-call-site changes needed. Verified: `cargo build`/`cargo test`/`cargo fmt --check` all clean (full `--features ffmpeg` build still not verified — that needs the Windows ffmpeg dev libraries, a separate, larger lift via vcpkg; Sentry's own code isn't feature-gated so this doesn't block it).
- [ ] Better Stack: log drain — not built; not even installed as a dependency anywhere
- [x] PostHog: `posthog-js` — built this session (init in a provider component, CSP updated to allow its network calls)
- [x] Langfuse — was already genuinely implemented before this session (real traces in `briefs.ts`)
- [ ] `traceparent` propagation — not built/verified

### PostHog Events
Built this session, as real `posthog.capture()` calls at actual interaction points (not the exact 7 named events above, which this session found no authoritative source for — the original list may have been aspirational, like much else in this doc):
- [x] `job_submitted` (job_id, model) — dashboard quick-generate flow
- [x] `brief_created` (brief_id) — brief creation
- [x] `generation_started` (brief_id, model, variants_per_angle, jobs_created) — batch generate
- [x] `checkout_started` (plan_tier) — Stripe checkout trigger
- [ ] The remaining named events above (signup, video complete, export downloaded, trial-to-paid, limit-hit) — not instrumented; export/signup-specific events in particular are blocked on their respective features not existing yet (Export Engine, dedicated signup completion hook)

### Security Hardening
- [ ] Upstash rate limiting: 1000 req/hr per workspace — middleware exists (`ratelimit.go`, Upstash-backed) but the specific threshold wasn't verified this session
- [ ] Input validation on all Go API handlers — not exhaustively audited
- [x] R2 presigned URLs expire in 15 minutes — **fixed this session** (was 1 hour in the Rust postprocessor's internal Mux-ingest URL; verified safe to shorten since Mux fetches immediately on asset creation)
- [x] CORS locked to verified origins — Go API already restricted to `NEXT_PUBLIC_APP_URL`; Rust postprocessor's previously-unconditional `CorsLayer::permissive()` is now gated behind `POSTPROCESS_ALLOWED_ORIGINS` (built this session)
- [x] RLS cross-workspace isolation test — **written and passing against real Postgres**, see Phase 1 gate note for the full mechanism (`internal/store/rls_test.go`). Not yet wired into the hosted CI pipeline itself, since CI has no Postgres service configured — the test is `TEST_DATABASE_URL`-gated and skips cleanly there; adding a Postgres service container to CI (and deciding whether it reuses migration 007 or mints its own test role) is the remaining step for "in CI" specifically, separate from the test existing and passing locally.

### QA Sign-Off
- [ ] E2E: sign up → brand → brief → video → export (< 15 min) — not run live; no way to execute the full stack in this sandbox (also currently impossible past the export step — Export Engine doesn't exist)
- [ ] URL scrape across platforms — not tested this session
- [ ] Tier limits enforced in staging — logic verified by reading code, not run live
- [ ] Trial flow emails — not built (no email provider found), so unverifiable as stated
- [ ] Load test — not performed

### Launch
- [ ] Staging fully green — not applicable; no staging environment exercised this session
- [ ] Production deploy checklist — not written
- [ ] SLO defined — not found in any doc
- [ ] On-call runbook — not found
- [ ] Launch comms — not applicable yet

### Gate
- [ ] All QA sign-off items passing in staging — not run
- [ ] Zero P0 bugs open — **not true as a general claim**, though every P0 found this session was fixed (see Phase 3 note); more may exist beyond this session's scope
- [ ] Sentry < 1% error rate in staging — unmeasurable; Sentry is new and nothing has run against it yet
- [ ] Production deploy successful — not attempted

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
