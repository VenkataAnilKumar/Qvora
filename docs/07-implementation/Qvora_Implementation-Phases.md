# Qvora — Implementation Phases
**Version:** 3.1 | **Updated:** April 16, 2026 | **Status:** Phases 0–3 Baseline Complete; Hardening In Progress

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
| Phase 7 | V1 Polish, Observability & Launch | Week 11 | ⏳ Pending |
| Phase 8 | Microservice Foundation | Weeks 12–13 | ⏳ Post-Launch |
| Phase 9 | Temporal + gRPC + Avatar Multi-Provider | Weeks 14–15 | ⏳ Post-Launch |
| Phase 10 | V2 Signal Loop & Intelligence | Weeks 16–18 | ⏳ Post-Launch |

**V1 Launch: Week 11 | Microservice Migration: Weeks 12–15 | V2 Signal Loop: Weeks 16–18**

---

## Phases 0–3 — Complete

Baseline delivery for Phases 0–3 is complete. See `IMPLEMENTATION_CHECKLIST.md` for current hardening status under **Phase 0–3 Fixes — In Progress**.

---

## Phase 4 — Brand Kit & Export

**Goal:** Per-workspace brand identity applied to all generated videos; platform-ready exports downloadable.

**Duration:** 1 week

### Deliverables

**Brand Kit System:**
- Brand creation wizard (`(dashboard)/brand/new`)
- Logo upload → R2 presigned PUT (PNG/SVG, max 5MB)
- Brand color picker + CSS preview
- Intro/outro bumper upload (MP4/MOV, max 5s, stored in R2)
- Custom font upload (TTF/OTF → stored in R2)
- Tone of voice notes (300 chars → injected into Claude prompt context)
- Multi-brand selector in dashboard sidebar
- Brand kit auto-applied on generation (logo + colors passed to Rust postprocessor via job payload)

**Export Engine:**
- `POST /v1/exports` → named package: `[Brand]_[Angle]_[Hook]_[Platform]_V[n]`
- Formats: MP4 1080p (all tiers), MP4 4K (Agency only), GIF preview (Growth+)
- Platform exports: Meta (9:16 + 1:1), TikTok (9:16), YouTube Shorts (9:16)
- Bulk ZIP download (Go server-side, R2 presigned URL, 48h expiry)
- Export history in `exports` table with R2 key + download count
- Platform compliance check: safe zones, minimum text size, duration limits

**Gate:**
- [ ] Brand logo watermark appears on all generated variants
- [ ] Export downloads as correctly named MP4
- [ ] Bulk ZIP works for 10+ variants with correct naming
- [ ] Platform compliance check rejects out-of-spec asset with clear error

---

## Phase 5 — Asset Library & Team

**Goal:** Browsable asset library; team roles enforced end-to-end.

**Duration:** 1 week

### Deliverables

**Asset Library:**
- Variants grid view (`(dashboard)/library`) — filter: brand / angle / format / date / status
- Search by tag metadata (`asset_tags` table — full-text)
- Variant detail page (`(dashboard)/library/[variantId]`) — metadata, playback, download
- Favorites / starring (`user_variant_stars` table)
- Archive vs Active (soft delete — `archived_at` timestamp)
- Storage usage indicator per workspace (sum R2 bytes)

**Team & Collaboration:**
- Invite team member by email via Clerk org invitation
- Role assignment on invite: Admin / Member / Viewer
- Viewer role: read-only — `403` at Go API + UI controls hidden
- Remove member from workspace (Clerk org member removal)
- Pending invite list with resend / revoke
- Seat count display + seat limits per tier (`(dashboard)/settings/team`)

**Gate:**
- [ ] Library shows all variants with correct metadata filters working
- [ ] Viewer role cannot trigger generation (blocked at API and UI layer)
- [ ] Team invite flow: invite → email → accept → workspace access
- [ ] Archive removes from default view, accessible via filter

---

## Phase 6 — Platform, Billing & Trial

**Goal:** Stripe checkout live; trial enforced; all tier limits activated.

**Duration:** 1 week

### Deliverables

**Stripe Integration:**
- Stripe products + prices: Starter $99 / Growth $149 / Agency $399
- `POST /v1/billing/checkout` → Stripe Checkout Session (hosted page)
- `POST /v1/billing/portal` → Stripe Customer Portal (plan changes, invoices)
- Webhook handler: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- Stripe Meter API: `meter_event` per video generation (usage-based billing foundation)
- Idempotent webhook handlers (Stripe-Signature HMAC-SHA256 verified)
- `workspaces.plan_tier` + `workspaces.stripe_status` updated on webhook

**Trial Flow:**
- 14-day trial on workspace creation (`trial_ends_at = created_at + 14 days`)
- Trial badge: "Trial — X days left" in sidebar
- Day 10 in-app urgency banner
- Day 13 modal gate (dismiss once)
- Day 15: generation blocked (402 from Go API, clear upgrade CTA)
- 30-day data retention post-expiry (asynq cleanup job)
- Conversion emails: Day 5 / Day 10 / Day 15 (via Resend / Postmark)

**Tier Enforcement (activation):**
- Starter: blocked at 4th variant per angle (clear upgrade CTA)
- Growth: blocked at 11th variant per angle
- Agency: unlimited
- Custom voice (ElevenLabs voice clone): Growth+ only
- 4K export: Agency only
- Custom avatar (HeyGen V2V): Agency only
- Extra seats: Growth 5 seats, Agency unlimited

**Gate:**
- [ ] Stripe checkout E2E (test mode): trial → paid plan → features unlock
- [ ] Webhook updates plan tier in DB within 5 seconds
- [ ] Day-15 generation lock activates automatically
- [ ] Stripe webhook rejects unsigned requests (HMAC failure → 400)
- [ ] All tier limits enforced in staging across all 3 tiers

---

## Phase 7 — V1 Polish, Observability & Launch

**Goal:** All services instrumented; E2E QA passed; production deployed; V1 launched.

**Duration:** 1 week

### Deliverables

**Observability (OTel foundation):**
- Sentry: `@sentry/nextjs` + Go SDK + Rust SDK across all services
- Better Stack: structured log drain from Railway (JSON, always: `job_id`, `org_id`, `trace_id`)
- PostHog: activation funnel + trial conversion events
- Langfuse: LLM cost per workspace, prompt versions, latency per call
- W3C `traceparent` header propagated across Next.js → Go API → Go workers → Rust

**Note:** Full OTel Collector (Grafana Tempo + Prometheus) configured in Phase 8. Phase 7 establishes Sentry + Better Stack as the minimum bar for V1 launch.

**PostHog Events:**
- `user_signed_up` (role, plan, referrer)
- `brief_generated` (latency_ms, template, angle_count)
- `video_generation_started` (format, angle_type, tier, model)
- `video_generation_complete` (duration_s, model, success)
- `export_downloaded` (format, platform, variant_count)
- `trial_to_paid` (days_to_convert, plan, source)
- `variant_limit_hit` (tier, upgrade_shown)

**Security Hardening:**
- Upstash rate limiting on all public endpoints (1000 req/hr per workspace)
- Input validation on all Go API handlers (no raw user strings in SQL)
- R2 presigned URLs expire in 15 minutes (down from current)
- CORS locked to verified origins in prod (no `*`)
- RLS cross-workspace isolation test in CI (user A cannot read user B data)

**QA Sign-Off:**
- E2E: sign up → brand setup → URL → brief → video generated → exported (< 15 min)
- URL scrape: Shopify PDP, WooCommerce, App Store listing, custom landing page
- All 3 tier limits enforced in staging
- Trial flow: Day 5/10/15 emails triggered, Day-15 lock verified
- Load test: 10 concurrent generation jobs without failure or data cross-contamination

**Launch:**
- Staging environment fully green
- Production deploy (Vercel + Railway + Supabase migrations)
- SLO defined: 99.5% uptime, < 25s brief P95, < 180s video P95
- On-call runbook written (incident triggers, escalation, rollback steps)
- Launch comms ready (Product Hunt, agency communities)

**Gate:**
- [ ] All QA sign-off items passing in staging
- [ ] Zero P0 bugs open
- [ ] Sentry showing < 1% error rate in staging
- [ ] Production deploy successful
- [ ] On-call runbook reviewed

---

## Phase 8 — Microservice Foundation

**Goal:** Decompose monolithic Go worker; replace Railway Redis queue with NATS JetStream; replace SSE Route Handler with Supabase Realtime.

**Duration:** 2 weeks | **Runs:** Post-V1 launch

### 8A — NATS JetStream Setup (Week 1)

**Infrastructure:**
- Provision NATS JetStream 3-node cluster on Railway
- Create streams: `QVORA_PIPELINE`, `QVORA_SIGNALS`, `QVORA_DLQ`
- Configure consumers with `AckExplicit`, `MaxDeliver=3`, exponential backoff
- Doppler: `NATS_URL`, `NATS_CREDENTIALS`
- Local dev: add NATS to `docker-compose.yml`

**Go SDK Integration:**
- Add `github.com/nats-io/nats.go` to Go modules
- Create `pkg/messaging/` package with typed publish/subscribe helpers
- Implement DLQ forwarding on `MaxDeliver` exhaustion
- Add NATS connection health check to all service `/health` endpoints

**Gate:**
- [ ] NATS cluster healthy on Railway (3 nodes, quorum)
- [ ] Test consumer processes message, acks, and DLQ receives on 3rd failure
- [ ] NATS Surveyor dashboard shows consumer lag metrics

### 8B — Service Decomposition (Week 1–2)

**Extract `ingestion-svc`:**
- Move `scrape_url` asynq handler → standalone `src/services/ingestion/` Go service
- Subscribe to `ingestion.scrape` NATS subject
- Publish `ingestion.complete` / `ingestion.failed`
- Deploy as separate Railway service
- Remove from monolithic worker binary

**Extract `brief-svc`:**
- Move brief generation LLM pipeline → standalone `src/services/brief/` Go service
- Subscribe to `brief.generate` NATS subject
- Publish `brief.complete` / `brief.failed`
- Preserve Vercel AI SDK calls (keep TypeScript or port to Go HTTP client)
- Deploy as separate Railway service

**Extract `asset-svc`:**
- Move export assembly, brand CRUD, Mux URL generation → `src/services/asset/` Go service
- Subscribe to `asset.export.create` NATS subject
- Publish `asset.export.ready`
- Deploy as separate Railway service

**Extract `identity-svc`:**
- Move Clerk JWT validation, quota check, Stripe meter → `src/services/identity/` Go service
- Expose gRPC interface (`CheckQuota`, `EmitMeterEvent`, `GetOrgPlan`)
- API Gateway calls identity-svc gRPC instead of inline middleware
- Deploy as separate Railway service

**Gate:**
- [ ] All 4 extracted services healthy on Railway
- [ ] Full E2E (URL → brief → video → export) passes through extracted services
- [ ] No regressions in brief generation latency (< 25s P95)
- [ ] Monolithic worker binary no longer handles ingestion or brief tasks

### 8C — Supabase Realtime Migration (Week 2)

**Replace SSE Route Handler:**
- Remove `src/apps/web/src/app/api/generation/[jobId]/stream/route.ts`
- Remove Go worker Upstash Redis writes (`job:{jobId}` status keys)
- All Temporal activities update `generation_jobs` status in Supabase directly
- Frontend: replace `EventSource` with Supabase Realtime subscription

```typescript
// Replace EventSource with:
const channel = supabase
  .channel(`job-${jobId}`)
  .on('postgres_changes', {
    event: 'UPDATE',
    schema: 'public',
    table: 'generation_jobs',
    filter: `id=eq.${jobId}`,
  }, (payload) => updateProgressUI(payload.new))
  .subscribe()
```

- Verify Supabase Realtime RLS: `org_id` claim on channel prevents cross-org leakage
- Supabase Realtime enabled in project settings (already on Pro plan)

**Gate:**
- [ ] Job progress updates appear in browser via Supabase Realtime (no polling)
- [ ] SSE route handler removed from codebase
- [ ] Redis `job:{jobId}` writes removed from Go workers
- [ ] Cross-org realtime isolation verified (org A cannot receive org B's events)

**Key Decisions in Phase 8:**
| Decision | Rationale |
|---|---|
| NATS before Temporal | NATS is simpler to adopt; validate messaging pattern before workflow engine |
| Service extraction one at a time | Reduces blast radius; each extraction fully tested before the next |
| Supabase Realtime last | Least risky change; SSE fallback stays in place until Realtime confirmed |

---

## Phase 9 — Temporal + gRPC + Multi-Provider Avatar

**Goal:** Migrate video generation pipeline to Temporal; add gRPC for Go → Rust; activate Tavus as secondary avatar provider.

**Duration:** 2 weeks | **Runs:** Post-Phase 8

### 9A — Temporal Setup (Week 1)

**Infrastructure:**
- Provision Temporal OSS on Railway (or Temporal Cloud for managed)
- Temporal schema migration on Supabase (or dedicated Temporal PostgreSQL)
- Temporal Web UI accessible at internal URL
- Go SDK: `go.temporal.io/sdk` added to `media-orchestrator` module

**VideoCreationWorkflow implementation:**
```go
// Activities: SelectVideoProvider, SubmitToFal, WaitForFalCompletion,
//             PostProcessVideo, IngestToMux, MarkJobComplete
// Each activity has: StartToCloseTimeout, RetryPolicy, compensation logic
```

- Webhook-to-Temporal signal bridge: fal.ai webhook → Go handler → `temporal.SignalWorkflow`
- Mux webhook → Go handler → `temporal.SignalWorkflow`
- Migrate existing asynq `generation:video` tasks to Temporal workflow triggers
- Parallel workflow execution: one workflow per angle per brief

**Gate:**
- [ ] Temporal Worker connected and polling task queue
- [ ] VideoCreationWorkflow visible in Temporal Web UI
- [ ] Workflow survives Go worker restart mid-execution (durability test)
- [ ] Workflow retries correctly on fal.ai timeout (simulate with mock)
- [ ] Full video pipeline (URL → READY) runs through Temporal end-to-end

### 9B — gRPC Internal Communication (Week 1)

**Protobuf definitions:**
- `proto/identity/v1/identity.proto` — IdentityService (CheckQuota, EmitMeterEvent)
- `proto/postprocess/v1/postprocess.proto` — PostProcessService (Process streaming RPC)
- Generate Go stubs: `protoc --go_out=. --go-grpc_out=.`
- Generate Rust stubs: `tonic-build` in `build.rs`

**Migration:**
- API Gateway → identity-svc: replace inline Clerk middleware with gRPC call
- media-orchestrator → media-postprocessor: replace HTTP `POST /process` with gRPC streaming
- Add mTLS: generate certs via Doppler, configure in gRPC dial options

**Gate:**
- [ ] gRPC quota check working for API Gateway → identity-svc
- [ ] gRPC streaming postprocess working: media-orchestrator calls Rust, receives progress stream
- [ ] mTLS certificates validated (connection rejected without correct cert)
- [ ] No latency regression vs previous HTTP (gRPC should be faster)

### 9C — Multi-Provider Avatar (Week 2)

**Provider interface (Go):**
```go
type AvatarProvider interface {
    CreateLipSync(ctx, req AvatarRequest) (jobID string, err error)
    GetStatus(ctx, jobID string) (AvatarResult, error)
}
```

**Implementations:**
- `HeyGenV3Provider{}` — existing HeyGen v3 logic refactored into interface
- `TavusProvider{}` — new Tavus v2 integration
- Provider registry: `map[string]AvatarProvider`
- Selection logic: Agency tier + avatar enabled → HeyGen (quality); cost-optimized path → Tavus
- Temporal activity: `CreateAvatarVideo` selects provider from registry

**Gate:**
- [ ] HeyGen v3 lip-sync still works through new provider interface
- [ ] Tavus v2 produces lip-sync video end-to-end (test mode)
- [ ] Provider fallback: if HeyGen returns 429, retry with Tavus automatically
- [ ] Avatar provider recorded on `generation_jobs` row for cost attribution

**Key Decisions in Phase 9:**
| Decision | Rationale |
|---|---|
| Temporal OSS first, Temporal Cloud optional | OSS is free; Cloud adds managed ops for $200+/mo — evaluate at scale |
| gRPC mTLS for all internal | Security baseline; Railway internal network is trusted but mTLS adds defense-in-depth |
| Tavus as secondary (not replacing HeyGen) | HeyGen v3 has best V2V quality; Tavus for cost optimization at volume |

---

## Phase 10 — V2 Signal Loop & Intelligence

**Goal:** Connect ad platform performance data; activate creative scoring; close the feedback loop into brief generation.

**Duration:** 3 weeks | **Runs:** Post-Phase 9

### 10A — Signal Ingestion (Week 1)

**`signal-svc` activation:**
- Meta Ads API OAuth connection (per workspace)
- TikTok Ads API OAuth connection (per workspace)
- Google Ads API OAuth connection (per workspace)
- Scheduled NATS message: `signal.sync` every 6 hours
- NATS consumer: fetch ad metrics per connected account → insert into `video_performance_events`
- Ad ID → Qvora asset matching (stored in `assets.platform_ad_id` per platform)
- GDPR cleanup consumer: `signal.gdpr.cleanup` — delete events older than retention window

**Database:**
```sql
-- New tables (migration)
CREATE TABLE ad_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  platform TEXT NOT NULL,
  account_id TEXT NOT NULL,
  oauth_token_encrypted TEXT NOT NULL,
  token_expires_at TIMESTAMPTZ,
  connected_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE video_performance_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  video_id UUID NOT NULL REFERENCES assets(id),
  org_id UUID NOT NULL,
  platform TEXT NOT NULL,
  metric_type TEXT NOT NULL,
  value NUMERIC NOT NULL,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE MATERIALIZED VIEW creative_scores AS
  SELECT video_id, org_id,
    AVG(CASE WHEN metric_type='ctr' THEN value END) AS avg_ctr,
    AVG(CASE WHEN metric_type='vtr' THEN value END) AS avg_vtr,
    AVG(CASE WHEN metric_type='roas' THEN value END) AS avg_roas,
    SUM(CASE WHEN metric_type='spend' THEN value END) AS total_spend
  FROM video_performance_events
  WHERE recorded_at > NOW() - INTERVAL '7 days'
  GROUP BY video_id, org_id;
```

**Gate:**
- [ ] Meta Ads API OAuth connected for test workspace
- [ ] Metrics ingested into `video_performance_events` after first sync
- [ ] `creative_scores` materialized view refreshes on schedule (hourly)
- [ ] GDPR cleanup removes events older than 90 days

### 10B — Creative Scoring Service (Week 2)

**`scoring-svc` (Python FastAPI):**
- Scaffold: `src/services/scoring/` — FastAPI + uvicorn + Pydantic v2
- V1 implementation: rule-based scoring (hook clarity, CTA presence, pacing score)
- Reads `creative_scores` materialized view via Supabase REST API
- Called by `asset-svc` after video is marked READY
- Stores predicted score on `assets.predicted_ctr`, `assets.predicted_vtr`
- Returns `reasoning` field for UI display ("Strong hook, clear CTA, fast pacing")

**V2 implementation (in-sprint iteration):**
- scikit-learn model trained on org's accumulated `video_performance_events`
- Retrain triggered by `signal.sync` completion when N >= 50 events for org
- Model stored in R2 per org (versioned)
- A/B: rule-based vs ML model prediction for orgs with < 50 events

**Gate:**
- [ ] scoring-svc returns score for newly generated video within 2s
- [ ] Score visible on variant detail page in dashboard
- [ ] ML model trains on test data (50+ events) and outperforms rules baseline

### 10C — Feedback Loop into Brief Generation (Week 3)

**brief-svc enhancement:**
- On brief creation, fetch top-performing exemplars from `creative_scores` view
- Filter by: same org, same product category, last 30 days
- Inject as few-shot context into Claude prompt:
  ```
  "For this product category, these hooks performed best:
   1. 'Stop scrolling — this skincare routine changed everything' → 4.2% CTR on TikTok
   2. 'You've been washing your face wrong' → 3.8% CTR on Meta"
  ```
- Configurable injection: only when org has >= 20 events (avoid noise)
- Langfuse tracks: briefs with/without exemplar injection, downstream CTR comparison

**Performance dashboard (`(dashboard)/insights`):**
- Creative score ranking per brief (sort by `predicted_ctr`)
- Historical performance: CTR / VTR / ROAS trend per video
- Top-performing angle type and hook pattern per org
- Fatigue detection: flag assets with declining CTR over 14 days
- Regen suggestion: "This angle is fatiguing — generate 3 new variants?"

**Gate:**
- [ ] Brief generation uses exemplar injection when org has >= 20 events
- [ ] Langfuse shows exemplar-injected briefs vs non-injected (A/B comparison)
- [ ] Insights dashboard renders correct charts for test workspace
- [ ] Fatigue detection flags declining assets with regen CTA

**Key Decisions in Phase 10:**
| Decision | Rationale |
|---|---|
| Append-only events (never UPDATE) | Full performance audit trail; enables rolling window scoring; A/B comparison without data loss |
| Materialized view for scores | Read performance for scoring-svc without repeated aggregation queries |
| scoring-svc starts rule-based | Immediate value; ML layer added as data accumulates |
| Exemplar injection threshold: 20 events | Below this, exemplars add noise not signal |

---

## Architecture Decisions by Phase

| Phase | Key Decision | Rationale |
|---|---|---|
| 0–3 | asynq + Railway Redis | Simplest viable queue for V1; known pattern |
| 4–7 | V1 launch with current stack | Ship V1 without architecture risk |
| 8 | NATS JetStream + service extraction | Purpose-built message bus; service isolation |
| 8 | Supabase Realtime replaces SSE | Eliminate polling; Realtime included in Supabase Pro |
| 9 | Temporal for video pipeline | Durable execution; workflow visibility; crash-safe |
| 9 | gRPC for internal hot paths | Type safety; 7× faster than REST; streaming support |
| 9 | Multi-provider avatar interface | Fallback resilience; cost optimization at volume |
| 10 | Event sourcing for performance data | Immutable audit trail; point-in-time analysis |
| 10 | Python scoring-svc | ML ecosystem requirement (scikit-learn / PyTorch) |

---

## V1 Monorepo Structure (Current)

```
src/
  apps/
    web/                → Next.js 15 (frontend + tRPC BFF)
  packages/
    ui/                 → shadcn/ui components
    types/              → Shared TypeScript types
    config/             → Biome + TS base configs
  services/
    api/                → Go Echo v4 (API Gateway)
    worker/             → Go asynq workers (monolithic)
    postprocess/        → Rust Axum + ffmpeg
  ai/
    prompts/            → Shared prompt files
```

## Target Monorepo Structure (Post-Phase 9)

```
src/
  apps/
    web/                → Next.js 15 (frontend + tRPC BFF)
  packages/
    ui/                 → shadcn/ui components
    types/              → Shared TypeScript types
    config/             → Biome + TS base configs
    proto/              → Protobuf definitions (.proto files)
  services/
    api/                → Go Echo v4 (API Gateway)
    identity/           → Go (auth, quota, billing)
    ingestion/          → Go (URL scraping, Modal)
    brief/              → Go (LLM orchestration)
    media-orchestrator/ → Go (Temporal worker, fal.ai)
    media-postprocessor/→ Rust (ffmpeg, gRPC server)
    asset/              → Go (metadata, exports, brands)
    signal/             → Go (ad platform, events) — V2
    scoring/            → Python (FastAPI, ML scoring) — V2
  ai/
    prompts/            → Shared prompt files
```

---

*Qvora Implementation Phases v3.0 — April 16, 2026 — Confidential*
