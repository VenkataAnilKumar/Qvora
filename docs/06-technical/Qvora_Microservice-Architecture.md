# QVORA
## Microservice Architecture
**Version:** 2.1 | **Date:** April 16, 2026 | **Status:** Target (Phase 8+ Post-Launch Migration)

---

## Overview

Qvora target architecture is decomposed into **8 domain services** across 3 runtimes (Go, Rust, Python), orchestrated by NATS JetStream for async messaging and Temporal for durable video pipeline workflows.

This document is the canonical reference for service boundaries, communication protocols, data ownership, and deployment topology for the Phase 8+ migration roadmap.

---

## Architecture Diagram

```
┌───────────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                      │
│              Next.js 15 (Vercel Edge + Serverless Functions)              │
│    shadcn/ui · tRPC BFF · Supabase Realtime (live job updates)           │
│                         Mux Player                                        │
└──────────────────────────────┬────────────────────────────────────────────┘
                               │ HTTP/REST (internal shared token)
┌──────────────────────────────▼────────────────────────────────────────────┐
│                      API GATEWAY (Go · Echo v4)                           │
│  Clerk JWT validation · Per-org rate limiting (Upstash) · Idempotency    │
│  Stripe meter_event emission · gRPC fan-out to domain services            │
└───┬──────────┬──────────┬──────────┬──────────┬──────────┬───────────────┘
    │          │          │          │          │          │
    ▼          ▼          ▼          ▼          ▼          ▼
[identity] [ingestion] [brief]  [asset]  [signal]  [scoring]
  -svc       -svc       -svc     -svc     -svc      -svc
   Go         Go         Go       Go       Go       Python
                         │
              ┌──────────┴────────────────────┐
              ▼                               ▼
      TEMPORAL WORKFLOWS              NATS JetStream
   (video pipeline — durable)    (simple background jobs)
              │
   ┌──────────┴──────────┐
   ▼                     ▼
[media-orchestrator]  [media-postprocessor]
      Go                  Rust
      │                   (gRPC ← from orchestrator)
      ▼
[fal.ai · HeyGen v3 · Tavus]
   (provider interface)
```

---

## Service Catalogue

### 1. `identity-svc` — Go

**Port:** 8001 (internal gRPC)
**DB ownership:** `organizations`, `users`
**Triggered by:** API Gateway on every authenticated request

**Responsibilities:**
- Clerk JWT validation and org claim extraction
- Tier enforcement: `starter` → `growth` → `agency`
- Quota check before all expensive operations (LLM calls, video generation)
- Stripe `meter_event` emission per billable operation
- Trial state management (14-day window, grace period, Day-8 lock)
- Per-org feature flag resolution (custom voice, 4K export, custom avatar)

**Internal API (gRPC):**
```protobuf
service IdentityService {
  rpc CheckQuota(QuotaRequest) returns (QuotaResponse);
  rpc EmitMeterEvent(MeterEventRequest) returns (google.protobuf.Empty);
  rpc GetOrgPlan(OrgPlanRequest) returns (OrgPlanResponse);
  rpc ValidateClaims(ClaimsRequest) returns (ClaimsResponse);
}
```

**Key pattern:** Double-validation — Next.js validates JWT signature at edge, Go validates claims per request. Prevents CVE-2024-10976 (RLS bypass via stale session context).

---

### 2. `ingestion-svc` — Go + Modal (Playwright)

**Port:** 8002 (internal HTTP)
**DB ownership:** None (stateless). Writes `ProductData` JSON to Cloudflare R2.
**Triggered by:** NATS subject `ingestion.scrape`

**Responsibilities:**
- Accept URL → dispatch Playwright scrape via Modal serverless
- Extract: product copy, images, pricing, USPs, OG metadata, page title
- Return structured `ProductData` JSON
- 24-hour cache by URL SHA256 hash (Upstash Redis)
- Confidence scoring per extracted field

**NATS subjects:**
```
ingestion.scrape     → input: { job_id, url, org_id }
ingestion.complete   → output: { job_id, product_data_r2_key }
ingestion.failed     → output: { job_id, error, retryable }
```

**Rate limit:** 100 scrapes/org/hour via Upstash sliding window.

---

### 3. `brief-svc` — Go + Vercel AI SDK

**Port:** 8003 (internal HTTP)
**DB ownership:** `briefs`, `brief_angles`, `brief_hooks`
**Triggered by:** NATS subject `brief.generate` (after ingestion completes)

**Responsibilities:**
- Consume `ProductData` from R2 → generate 3–5 creative angles (Claude Sonnet 4.6)
- Generate hooks and copy variations per angle
- GPT-4o for structured JSON parsing (strict mode)
- Langfuse trace on every LLM call with `org_id` and `job_id` tags
- V2: Inject top-performing hooks from `signal-svc` as few-shot LLM context

**LLM routing:**
| Stage | Model | Reason |
|---|---|---|
| Product extraction | `gpt-4o` | JSON strict mode, reliable structured output |
| Angle generation | `claude-sonnet-4-6` | Creative quality, tone-aware output |
| Hook generation | `claude-sonnet-4-6` | Variant diversity, brand voice matching |
| Manual regen | `claude-sonnet-4-6` | Per-angle/hook regeneration on user request |

**NATS subjects:**
```
brief.generate      → input: { job_id, org_id, product_data_r2_key }
brief.complete      → output: { job_id, brief_id, angles_count }
brief.failed        → output: { job_id, error }
```

**OTel attributes (all LLM spans):**
```
gen_ai.system = "anthropic" | "openai"
gen_ai.request.model = "claude-sonnet-4-6" | "gpt-4o"
gen_ai.usage.input_tokens = N
gen_ai.usage.output_tokens = N
qvora.org_id = <org_id>
qvora.job_id = <job_id>
qvora.brief_id = <brief_id>
```

---

### 4. `media-orchestrator` — Go + Temporal

**Port:** 8004 (Temporal worker, no HTTP port)
**DB ownership:** `generation_jobs` (outer wrapper over fal `request_id`)
**Triggered by:** Temporal workflow trigger from API Gateway

**Core pattern — Outer Job Wrapper:**
Every `generation_job` in Supabase wraps a fal.ai `request_id`. This decouples Qvora job lifecycle from fal lifecycle and enables:
- Provider swap (Veo → Kling → Runway) on failure without resetting the job
- Cost attribution per org across providers
- User-facing status normalization (Qvora status, not fal status)
- Retry at the specific failed step — not from scraping

**Temporal workflow:**
```go
func VideoCreationWorkflow(ctx workflow.Context, input VideoJobInput) error {
    // Step 1: Select provider based on format + tier
    provider := workflow.ExecuteActivity(ctx, SelectVideoProvider, input)

    // Step 2: Submit to fal.ai async queue
    falRequestID := workflow.ExecuteActivity(ctx, SubmitToFal, provider, input,
        workflow.WithActivityOptions(workflow.ActivityOptions{
            StartToCloseTimeout: 10 * time.Minute,
            RetryPolicy: &temporal.RetryPolicy{MaxAttempts: 3},
        }))

    // Step 3: Wait for fal webhook (signal) or timeout + poll
    result := workflow.ExecuteActivity(ctx, WaitForFalCompletion, falRequestID,
        workflow.WithActivityOptions(workflow.ActivityOptions{
            StartToCloseTimeout: 30 * time.Minute,
        }))

    // Step 4: Trigger postprocessor via gRPC
    processed := workflow.ExecuteActivity(ctx, PostProcessVideo, result,
        workflow.WithActivityOptions(workflow.ActivityOptions{
            StartToCloseTimeout: 5 * time.Minute,
        }))

    // Step 5: Ingest to Mux
    asset := workflow.ExecuteActivity(ctx, IngestToMux, processed)

    // Step 6: Update Supabase → triggers Realtime push to frontend
    return workflow.ExecuteActivity(ctx, MarkJobComplete, asset)
}
```

**Provider interface (model-agnostic):**
```go
type VideoProvider interface {
    Submit(ctx context.Context, req VideoRequest) (requestID string, err error)
    GetStatus(ctx context.Context, requestID string) (VideoResult, error)
}

// Registered providers:
// "veo_3_1"        → FalVeo31Provider{}
// "kling_3_0"      → FalKling30Provider{}
// "runway_gen4_5"  → FalRunwayGen45Provider{}
```

**fal.ai concurrency guard:**
```go
// Redis semaphore per org (fal.ai hard limit: 2 concurrent/user)
// Key: fal:semaphore:<org_id>  →  max 2 via SETNX with expiry
// Checked by SelectVideoProvider activity before submission
// Released on WaitForFalCompletion completion or timeout
```

**Cost circuit breaker:**
```go
// Redis key: fal:cost:<YYYYMMDD>:<HH>
// Incremented by estimated cost per model invocation
// Checked before every submission
// Limits: starter=$2/hr, growth=$8/hr, agency=$20/hr
// Reset via scheduled NATS message at top of each hour
```

**NATS subjects (for simple non-workflow tasks):**
```
media.webhook.fal   → fal.ai webhook received, forward to Temporal signal
media.webhook.mux   → Mux asset.ready, forward to Temporal signal
```

---

### 5. `media-postprocessor` — Rust + Axum

**Port:** 8005 (internal gRPC — called by media-orchestrator)
**DB ownership:** None (stateless)
**Triggered by:** gRPC from `media-orchestrator` Temporal activity

**Pipeline (Tokio async):**
```
1. Download raw video from fal.ai CDN (streaming, not buffered)
2. tokio::process::Command → ffmpeg:
   ├── 9:16 reframe (scale + crop to 1080×1920)
   ├── Watermark overlay (org brand kit logo from R2)
   ├── Caption burn-in (SRT from ElevenLabs transcript)
   ├── H.264 transcode (libx264, CRF 23, ultrafast, +faststart)
   └── MP4 output (Mux-ready)
3. Upload to Cloudflare R2 (presigned PUT URL)
4. Return R2 object URL to media-orchestrator via gRPC response
```

**gRPC interface:**
```protobuf
service PostProcessService {
  rpc Process(ProcessRequest) returns (stream ProcessProgress);
}

message ProcessRequest {
  string raw_video_url = 1;
  string brand_logo_r2_key = 2;
  string captions_srt = 3;
  string output_spec = 4;      // "1080p" | "4k"
  string org_id = 5;
  string job_id = 6;
}

message ProcessProgress {
  string stage = 1;            // "downloading" | "transcoding" | "uploading"
  int32 percent = 2;
  string output_r2_url = 3;   // populated on final message
}
```

**Scale:** CPU-bound. No GPU required. Spot instances at 60–80% discount. `preStop` lifecycle hook waits for in-flight ffmpeg completion before pod termination.

---

### 6. `asset-svc` — Go

**Port:** 8006 (internal HTTP + gRPC)
**DB ownership:** `assets` (metadata), `exports`, `brands`

**Responsibilities:**
- Asset metadata CRUD (`mux_asset_id`, `mux_playback_id`, `r2_url`, `duration`, `model_used`)
- Mux signed playback URL generation (per-org HS256 JWT, 1-hour expiry)
- Brand kit management (logo R2 key, hex palette, voice settings per workspace)
- Export ZIP bundle assembly (NATS-triggered background job)
- Export naming convention: `[Brand]_[Angle]_[Hook]_[Platform]_V[n]`

**NATS subjects:**
```
asset.export.create → input: { export_id, asset_ids[], org_id, format }
asset.export.ready  → output: { export_id, zip_r2_url, signed_url }
```

**Export bundle layout:**
```
export-<export_id>/
  ├── videos/
  │   ├── [Brand]_ProblemSolution_Hook1_TikTok_V1.mp4
  │   └── [Brand]_SocialProof_Hook2_Meta_V2.mp4
  ├── copy/
  │   ├── hooks.json          (headlines, CTAs per video)
  │   └── captions.srt
  └── metadata.json           (model, dimensions, duration, predicted_score)
```

---

### 7. `signal-svc` — Go (V2)

**Port:** 8007 (internal HTTP)
**DB ownership:** `ad_accounts`, `video_performance_events`
**Triggered by:** NATS schedule subject `signal.sync` (every 6 hours)

**Core architectural decision — Event Sourcing:**
Performance data is never updated in-place. Every metric is an append-only event. Scores are derived via Supabase materialized views refreshed hourly.

```sql
-- Append-only events (never UPDATE a row)
CREATE TABLE video_performance_events (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  video_id      UUID NOT NULL REFERENCES assets(id),
  org_id        UUID NOT NULL,
  platform      TEXT NOT NULL,     -- 'meta' | 'tiktok' | 'google'
  metric_type   TEXT NOT NULL,     -- 'ctr' | 'vtr' | 'roas' | 'impressions' | 'spend'
  value         NUMERIC NOT NULL,
  recorded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Materialized view for scoring-svc to read
CREATE MATERIALIZED VIEW creative_scores AS
  SELECT
    video_id,
    org_id,
    AVG(CASE WHEN metric_type = 'ctr' THEN value END) AS avg_ctr,
    AVG(CASE WHEN metric_type = 'vtr' THEN value END) AS avg_vtr,
    AVG(CASE WHEN metric_type = 'roas' THEN value END) AS avg_roas,
    SUM(CASE WHEN metric_type = 'spend' THEN value END) AS total_spend
  FROM video_performance_events
  WHERE recorded_at > NOW() - INTERVAL '7 days'
  GROUP BY video_id, org_id;
```

**NATS subjects:**
```
signal.sync         → trigger ad platform sync per org
signal.recommend    → score creatives, flag fatigue, suggest regen
signal.gdpr.cleanup → delete events older than retention window
```

**V2 feedback loop into brief-svc:**
```
Top-performing (hook, angle) pairs → stored as exemplars per industry vertical
brief-svc fetches exemplars for org on brief creation
Injected as few-shot context: "This hook achieved 4.2% CTR on TikTok for DTC skincare"
```

---

### 8. `scoring-svc` — Python + FastAPI (V2)

**Port:** 8008 (internal HTTP)
**DB ownership:** Reads `creative_scores` view (read-only)
**Triggered by:** `asset-svc` after video is marked READY

**Responsibilities:**
- Pre-publish creative scoring (before ad spend)
- V1: Rule-based heuristics (hook clarity, CTA presence, pacing)
- V2: Scikit-learn model trained on org's own performance data
- V3: PyTorch CNN on video frame features (visual quality, text legibility)
- Returns `{ predicted_ctr, predicted_vtr, confidence, reasoning }` per video

**API:**
```python
POST /score
{
  "video_id": "uuid",
  "org_id": "uuid",
  "hook_text": "...",
  "angle_type": "problem_solution",
  "platform": "tiktok",
  "thumbnail_r2_key": "..."
}

→ {
  "predicted_ctr": 0.042,
  "predicted_vtr": 0.68,
  "confidence": 0.71,
  "reasoning": "Strong pattern interrupt hook, clear CTA, fast pacing",
  "flags": ["text_overlap_safe_zone"]
}
```

**Technology:**
- FastAPI + uvicorn (async)
- scikit-learn (feature-based scoring)
- Pydantic v2 (validation)
- httpx (async HTTP to Supabase REST API)

---

## Message Bus — NATS JetStream

**Replaces:** Railway Redis (TCP) for job queuing.
**Railway Redis is decommissioned** once NATS migration is complete.
**Upstash Redis is retained** for: rate-limit counters, URL scrape cache, SSE fallback.

### Stream topology:

```
Stream: QVORA_PIPELINE
  Subjects: ingestion.*, brief.*, media.*, asset.*
  Retention: WorkQueue (consumed once)
  Replicas: 3 (Railway cluster)
  MaxAge: 24h
  AckPolicy: Explicit (at-least-once)

Stream: QVORA_SIGNALS (V2)
  Subjects: signal.*
  Retention: Limits (keep last 1000 per subject)
  Replicas: 3

Stream: QVORA_DLQ
  Subjects: dlq.*
  Retention: Limits (7 days)
  Replicas: 1
```

### Full subject map:

| Subject | Publisher | Subscriber | Description |
|---|---|---|---|
| `ingestion.scrape` | API Gateway | ingestion-svc | Trigger URL scrape |
| `ingestion.complete` | ingestion-svc | brief-svc | Product data ready |
| `ingestion.failed` | ingestion-svc | API Gateway | Scrape failed, notify user |
| `brief.generate` | ingestion-svc | brief-svc | Start brief generation |
| `brief.complete` | brief-svc | media-orchestrator | Brief ready, start videos |
| `brief.failed` | brief-svc | API Gateway | Brief generation failed |
| `media.webhook.fal` | Go webhook handler | Temporal signal | fal.ai callback |
| `media.webhook.mux` | Go webhook handler | Temporal signal | Mux asset ready |
| `asset.export.create` | asset-svc | asset-svc worker | Start ZIP export |
| `asset.export.ready` | asset-svc worker | API Gateway | Export ready, notify user |
| `signal.sync` | NATS scheduler | signal-svc | Trigger ad platform sync |
| `signal.recommend` | signal-svc | signal-svc | Run fatigue detection |
| `signal.gdpr.cleanup` | NATS scheduler | signal-svc | GDPR data cleanup |
| `dlq.*` | Any service | ops team | Dead-letter queue |

### Consumer configuration (Go example):
```go
js, _ := nc.JetStream()

js.Subscribe("ingestion.scrape", func(msg *nats.Msg) {
    var payload IngestPayload
    json.Unmarshal(msg.Data, &payload)

    if err := handleScrape(ctx, payload); err != nil {
        // NATS will redeliver up to MaxDeliver times
        msg.Nak()
        return
    }
    msg.Ack()
}, nats.Durable("ingestion-scrape-worker"),
   nats.AckExplicit(),
   nats.MaxDeliver(3),
   nats.BackOff([]time.Duration{5*time.Second, 30*time.Second, 5*time.Minute}))
```

---

## Real-Time Updates — Supabase Realtime

**Replaces:** Custom SSE Route Handler (`/api/generation/[jobId]/stream`).

Workers update `generation_jobs.status` in Supabase. Supabase Realtime propagates Postgres changes to the browser via WebSocket. No polling, no Redis key writes, no custom SSE route.

### How it works:
```
1. Temporal activity updates Supabase:
   UPDATE generation_jobs SET status='VIDEO_GENERATING', progress=60 WHERE id=<job_id>

2. Supabase Realtime detects WAL change → broadcasts to channel subscribers

3. Frontend subscription:
   const channel = supabase
     .channel(`job-${jobId}`)
     .on('postgres_changes', {
       event: 'UPDATE',
       schema: 'public',
       table: 'generation_jobs',
       filter: `id=eq.${jobId}`,
     }, (payload) => {
       updateProgressUI(payload.new.status, payload.new.progress)
     })
     .subscribe()

4. RLS on Realtime channel:
   Supabase checks org_id in JWT claim before broadcasting
   → Zero cross-org data leakage
```

### Job status progression:
```
CREATED (5%)  → SCRAPING (15%) → BRIEF_GENERATING (30%) →
VIDEO_QUEUED (40%) → VIDEO_GENERATING (60%) →
POST_PROCESSING (80%) → MUX_INGESTING (90%) →
READY (100%) | FAILED
```

---

## Service Communication Matrix

| From | To | Protocol | Auth |
|---|---|---|---|
| Next.js | Go API Gateway | HTTPS REST | Internal shared token |
| Go API Gateway | identity-svc | gRPC (internal) | mTLS |
| Go API Gateway | brief-svc | gRPC (internal) | mTLS |
| Go API Gateway | asset-svc | gRPC (internal) | mTLS |
| Go API Gateway | NATS JetStream | NATS protocol | NATS credentials |
| media-orchestrator | media-postprocessor | gRPC streaming | mTLS |
| media-orchestrator | fal.ai | HTTPS REST | fal.ai API key |
| media-orchestrator | HeyGen v3 | HTTPS REST | HeyGen API key |
| media-orchestrator | Tavus | HTTPS REST | Tavus API key |
| media-orchestrator | ElevenLabs | HTTPS REST | ElevenLabs API key |
| fal.ai | Go webhook handler | HTTPS (webhook) | SHA256 sig verify |
| Mux | Go webhook handler | HTTPS (webhook) | Mux signing secret |
| asset-svc | Mux API | HTTPS REST | Mux token |
| signal-svc | Meta/TikTok/Google | HTTPS REST | OAuth tokens |
| scoring-svc | Supabase | HTTPS REST | Service role key |

**All internal gRPC uses mTLS. All external APIs use service-owned API keys stored in Doppler.**

---

## Avatar Provider Strategy

**Pattern:** Provider interface with HeyGen v3 as default, Tavus as secondary.

```go
type AvatarProvider interface {
    CreateLipSync(ctx context.Context, req AvatarRequest) (jobID string, err error)
    GetStatus(ctx context.Context, jobID string) (AvatarResult, error)
    GetWebhookPayload(body []byte) (AvatarResult, error)
}

// Registry:
var avatarProviders = map[string]AvatarProvider{
    "heygen_v3": &HeyGenV3Provider{},   // default
    "tavus_v2":  &TavusProvider{},      // fallback / price-optimized
}
```

**Provider selection logic:**
- `agency` tier + avatar enabled → HeyGen v3 (quality)
- High volume / cost-optimized path → Tavus v2
- HeyGen v3 is still the active platform (`developers.heygen.com`)
- Tavus integration added as runnable secondary from Phase 9

---

## Multi-Tenancy Isolation

### Three-tier model:

| Tier | Plan | Mechanism |
|---|---|---|
| Shared schema + RLS | Starter / Growth | `org_id` column + Supabase RLS policies on all tables |
| Schema-per-org | Agency+ | Separate Postgres schema, provisioned on org creation |
| Separate DB | Enterprise (V2) | Isolated Supabase project for data residency requirements |

### RLS policy (all tables):
```sql
ALTER TABLE generation_jobs ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation ON generation_jobs
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

### Per-org rate limits (Upstash Redis):
```
ratelimit:<org_id>:video_gen   → 2 concurrent max (fal.ai hard limit)
ratelimit:<org_id>:brief_gen   → 50 requests/hour
ratelimit:<org_id>:scrape      → 100 requests/hour
ratelimit:<org_id>:api         → 1000 requests/hour
```

---

## Observability

**Standard:** OpenTelemetry → single instrumentation layer, pluggable backends.

### Instrumentation by runtime:

| Runtime | Library | Signals |
|---|---|---|
| Go | `go.opentelemetry.io/otel` | Traces, metrics, logs |
| Rust | `opentelemetry-rust` + `tracing-opentelemetry` | Traces, metrics |
| Python | `opentelemetry-sdk` | Traces, metrics |
| Next.js | Vercel built-in OTel | Traces |
| LLM calls (all) | OpenLLMetry | Traces + token metrics |

### Backend routing via OTel Collector:
```
Traces   → Grafana Tempo
Metrics  → Prometheus → Grafana dashboards
Logs     → Better Stack (structured JSON, always include: job_id, org_id, trace_id)
Errors   → Sentry (all services)
LLM      → Langfuse (cost, latency, prompt versions per org)
Product  → PostHog (activation funnel, trial conversion)
```

### Sampling:
- 100% of errors and failures
- 10% of successful traces (tail-based in OTel Collector)

### W3C `traceparent` propagation:
```
Next.js BFF → injects traceparent header
Go API Gateway → reads + continues span, passes to services via gRPC metadata
media-orchestrator → passes to Temporal workflow as baggage
media-postprocessor → reads from gRPC metadata, attaches ffmpeg spans as children
All webhook handlers → inject trace context from payload metadata field
```

### Critical metrics to alert on:
```
NATS_consumer_lag{stream, consumer}          → queue depth scaling signal
fal_concurrent_requests{org_id}              → approach 2/org limit
llm_cost_usd_total{model, org_id}           → cost attribution
video_generation_success_rate{provider}      → Veo vs Kling vs Runway health
p95_e2e_latency_seconds                      → scrape → READY target < 180s
temporal_workflow_failure_rate               → pipeline reliability
```

---

## Deployment Topology

```
┌─── VERCEL ──────────────────────────────┐
│  Next.js 15 (SSR + serverless fns)      │
│  Supabase Realtime client (browser)     │
└─────────────────────────────────────────┘

┌─── RAILWAY ─────────────────────────────────────────────────────────┐
│  api-gateway          (Go Echo v4, 2 replicas min)                  │
│  identity-svc         (Go, 1 replica, autoscale CPU)               │
│  ingestion-svc        (Go, scale-to-zero, NATS consumer)           │
│  brief-svc            (Go, autoscale CPU/LLM-concurrency)          │
│  media-orchestrator   (Go Temporal worker, autoscale NATS lag)     │
│  media-postprocessor  (Rust Axum, autoscale CPU, spot-compatible)  │
│  asset-svc            (Go, 1 replica, autoscale CPU)               │
│  signal-svc           (Go, 1 replica, V2)                          │
│  scoring-svc          (Python FastAPI, 1 replica, V2)              │
│  temporal-server      (Temporal OSS or Temporal Cloud)             │
│  nats-cluster         (NATS JetStream, 3-node cluster)             │
└─────────────────────────────────────────────────────────────────────┘

┌─── SUPABASE ────────────────────────────┐
│  PostgreSQL 16                          │
│  RLS (all tables, org_id isolation)     │
│  Realtime (job status push to browser) │
│  pgvector (V2 — semantic search)       │
└─────────────────────────────────────────┘

┌─── EXTERNAL ─────────────────────────────────────────────────────────┐
│  Cloudflare R2     → Video + image object storage (zero egress)     │
│  Mux               → HLS video delivery (signed playback)           │
│  Upstash Redis     → Rate limits + URL scrape cache (HTTP only)     │
│  Modal             → Playwright scraping (serverless)               │
│  fal.ai            → Veo 3.1 / Kling 3.0 / Runway Gen-4.5         │
│  HeyGen v3         → Avatar lip-sync (default)                      │
│  Tavus v2          → Avatar lip-sync (secondary, Phase 9+)         │
│  ElevenLabs        → TTS (eleven_v3 / eleven_flash_v2_5)           │
│  Clerk             → Auth, org management                           │
│  Stripe            → Subscriptions + Meter API (usage billing)      │
│  Doppler           → Secrets (dev / stg / prd)                     │
│  Langfuse          → LLM prompt observability                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Ownership Summary

| Service | Owns (writes) | Reads |
|---|---|---|
| identity-svc | `organizations`, `users` | — |
| ingestion-svc | R2 (ProductData JSON) | — |
| brief-svc | `briefs`, `brief_angles`, `brief_hooks` | `creative_scores` view (V2) |
| media-orchestrator | `generation_jobs` | `briefs`, `brief_angles` |
| media-postprocessor | R2 (processed MP4) | — |
| asset-svc | `assets`, `exports`, `brands` | `generation_jobs` |
| signal-svc | `ad_accounts`, `video_performance_events` | `assets` |
| scoring-svc | — | `creative_scores` view, R2 thumbnails |

**Rule:** No service reads another service's owned tables directly. Cross-service data is accessed via gRPC calls or NATS messages, never direct SQL cross-schema joins.

---

## Decisions Changed from v1.0

| v1.0 Decision | v2.0 Change | Reason |
|---|---|---|
| Railway Redis (TCP) for asynq | NATS JetStream for pipeline messaging | Purpose-built message bus; at-least-once delivery; stream replay; DLQ native |
| Custom SSE Route Handler | Supabase Realtime | Supabase already deployed; Realtime included; RLS on channels; no polling |
| asynq for video pipeline | Temporal.io workflow | Durable execution; workflow versioning; built-in saga; visibility UI |
| HeyGen v3 locked | Provider interface (HeyGen default + Tavus) | Fallback resilience; cost optimization at volume |
| Go-only workers | Python FastAPI for scoring-svc | ML ecosystem requirement; scikit-learn / PyTorch not viable in Go |
| HTTP REST internal | gRPC for hot paths (Go ↔ Rust) | 7× faster; bi-directional streaming for postprocessor progress |
| Monolithic worker binary | 6 decomposed services | Independent scaling; isolated failure domains |

## Decisions Unchanged from v1.0

| Decision | Still applies |
|---|---|
| Tailwind v4 CSS-only (`@theme {}` in `globals.css`) | Yes |
| HeyGen v3 active platform (`developers.heygen.com`) | Yes — still default provider |
| `fal.queue.submit()` only (never `fal.subscribe()`) | Yes |
| Go = I/O-bound; Rust = CPU-bound | Yes — boundary unchanged |
| Clerk Organizations = workspaces | Yes |
| Agency-first V1 ICP | Yes |
| Supabase RLS on all tables | Yes |
| Doppler for all secrets | Yes |
| ffmpeg-next bindings (never `Command::new("ffmpeg")`) | Yes |
| Migrations in `supabase/migrations/` only | Yes |

---

*Qvora Microservice Architecture v2.0 — April 16, 2026 — Confidential*
