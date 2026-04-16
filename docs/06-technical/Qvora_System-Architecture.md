# QVORA
## System Architecture
**Version:** 2.0 | **Date:** April 16, 2026 | **Status:** Active
**Reference:** `Qvora_Microservice-Architecture.md` (canonical service catalogue)
**Stack:** `Qvora_Architecture-Stack.md`

---

## Workload Profiles

| Profile | Latency | Compute | Services |
|---|---|---|---|
| **Interactive** | < 2s | CPU / LLM API | identity-svc, brief-svc, asset-svc |
| **Async pipeline** | 60–180s | GPU via fal.ai | media-orchestrator, media-postprocessor |
| **Background** | Seconds | CPU | ingestion-svc, NATS consumers |
| **Scheduled** | Hourly/daily | CPU | signal-svc, scoring-svc (V2) |

---

## High-Level Component Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│  CLIENT LAYER                                                            │
│  Next.js 15 (Vercel Edge)                                                │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │  App Router · shadcn/ui · TanStack Query · Framer Motion         │   │
│  │  tRPC client · Supabase Realtime (job progress — no polling)     │   │
│  │  Mux Player (signed HLS playback)                                │   │
│  └─────────────────────────┬────────────────────────────────────────┘   │
└───────────────────────────┬┘────────────────────────────────────────────┘
                            │ HTTPS (internal shared token)
┌───────────────────────────▼──────────────────────────────────────────────┐
│  API GATEWAY (Go · Echo v4)                                              │
│  Clerk JWT validation · Idempotency keys · gRPC fan-out                 │
│  Per-org rate limiting (Upstash) · Stripe meter_event emission          │
└───┬──────────┬──────────┬──────────┬──────────┬──────────┬─────────────┘
    │          │          │          │          │          │
    ▼          ▼          ▼          ▼          ▼          ▼
[identity] [ingestion] [brief]  [asset]  [signal]  [scoring]
  -svc       -svc       -svc     -svc     -svc      -svc
   Go         Go         Go       Go       Go       Python
              │          │                          (V2)
              └──────────┴──────────────────┐
                                            ▼
                                    NATS JetStream
                                  (async messaging bus)
                                            │
                            ┌───────────────┤
                            ▼               ▼
                    TEMPORAL WORKFLOWS   Simple consumers
                   (video pipeline)    (export, signals)
                            │
               ┌────────────┴────────────┐
               ▼                         ▼
       [media-orchestrator]     [media-postprocessor]
              Go                      Rust + Axum
              │                    (gRPC streaming)
              │
   ┌──────────┼─────────────┐
   ▼          ▼             ▼
[fal.ai]  [HeyGen v3]  [Tavus v2]
 T2V/I2V   Avatar       Avatar
           (default)   (fallback)
              │
              ▼
           [Mux]
          HLS delivery
              │
              ▼
      Supabase Realtime → Next.js browser (live status)
```

---

## Sequence Diagrams

### Flow 1 — URL to Creative Brief

```
Browser      Next.js BFF    API Gateway    ingestion-svc   brief-svc    Supabase
   │               │               │               │             │           │
   │ POST /briefs  │               │               │             │           │
   │ { url, org }  ├─ tRPC ───────►               │             │           │
   │               │               │ INSERT job    │             │           │
   │               │               │ status=CREATED────────────────────────►│
   │               │               │◄─ job_id ─────────────────────────────┤│
   │◄─ { job_id }──┤◄──────────────┤               │             │           │
   │               │               │               │             │           │
   │ [Realtime sub]│ subscribe:     │               │             │           │
   │ generation_jobs where id=job_id               │             │           │
   │               │               │               │             │           │
   │               │               │ NATS publish  │             │           │
   │               │               │ ingestion.scrape────────────►            │
   │               │               │               │ Modal/Playwright         │
   │               │               │               │ scrape URL  │           │
   │               │               │               │             │           │
   │               │               │ UPDATE job    │             │           │
   │◄─ Realtime:   │               │ status=SCRAPING─────────────────────────►
   │  { status: 'SCRAPING', pct:15}│               │             │           │
   │               │               │               │             │           │
   │               │               │               │ NATS publish│           │
   │               │               │               │ brief.generate──────────►
   │               │               │               │             │           │
   │               │               │               │    GPT-4o → product     │
   │               │               │               │    Claude → angles      │
   │               │               │               │    Claude → hooks       │
   │               │               │               │             │           │
   │               │               │ UPDATE job    │             │           │
   │◄─ Realtime:   │               │ status=READY ──────────────────────────►
   │  { status: 'READY', pct:100 } │               │             │           │
```

**Latency budget:**
- Modal Playwright scrape: 3–8s
- GPT-4o product extraction: 2–5s
- Claude angle generation: 4–9s
- Claude hook generation: 2–5s
- **Total target: < 25s P95**

---

### Flow 2 — Brief to Video (Full Pipeline)

```
Browser    API GW   NATS   Temporal   media-orch   fal.ai   media-post   Mux   Supabase
   │          │       │        │           │           │          │        │        │
   │ POST     │       │        │           │           │          │        │        │
   │ /gen     ├──────►│        │           │           │          │        │        │
   │          │       │ trigger│           │           │          │        │        │
   │          │       │ workflow──────────►│           │          │        │        │
   │          │       │        │  INSERT   │           │          │        │        │
   │          │       │        │  job=QUEUED────────────────────────────────────────►
   │◄─ job_id ┤       │        │           │           │          │        │        │
   │          │       │        │           │           │          │        │        │
   │◄─Realtime: QUEUED (5%)    │           │           │          │        │        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │  Check fal│           │          │        │        │
   │          │       │        │  semaphore│           │          │        │        │
   │          │       │        │  (max 2/org)          │          │        │        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │  fal.queue│           │          │        │        │
   │          │       │        │  .submit()────────────►           │        │        │
   │          │       │        │◄─ request_id ─────────┤          │        │        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │  UPDATE job=VIDEO_GEN ─────────────────────────────►
   │◄─Realtime: VIDEO_GENERATING (60%)     │           │          │        │        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │           │◄─ webhook ┤          │        │        │
   │          │       │        │◄─ signal  │           │          │        │        │
   │          │       │        │           │  gRPC     │          │        │        │
   │          │       │        │           │  Process()─────────► │        │        │
   │          │       │        │           │           │ ffmpeg   │        │        │
   │          │       │        │           │           │ transcode│        │        │
   │          │       │        │           │           │ watermark│        │        │
   │          │       │        │           │           │ captions │        │        │
   │          │       │        │           │◄─ r2_url ─────────── │        │        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │           │  Mux asset────────────────────►        │
   │          │       │        │           │◄─ asset_id / playback_id ──────┤        │
   │          │       │        │           │           │          │        │        │
   │          │       │        │  UPDATE job=READY ─────────────────────────────────►
   │◄─Realtime: READY (100%) + playback_url           │          │        │        │
   │ [Mux Player loads]        │           │           │          │        │        │
```

**Latency budget per video:**
- fal.ai queue + generation: 45–150s (Veo 3.1 fast: 45–70s)
- Rust postprocessing: 5–15s
- Mux ingest: 3–8s
- **Total target: 60–180s**

---

### Flow 3 — Avatar Lip-Sync (V2V, HeyGen v3)

```
media-orchestrator     HeyGen v3 API       media-postprocessor    Mux
       │                     │                      │               │
       │ POST /v1/avatar/    │                      │               │
       │ video-translate     │                      │               │
       │ { video_url,        │                      │               │
       │   audio_url }──────►│                      │               │
       │◄─ { video_id }──────┤                      │               │
       │                     │                      │               │
       │ [poll every 15s]    │                      │               │
       │ GET /v1/avatar/     │                      │               │
       │ video-translate/:id─►                      │               │
       │◄─ { status:         │                      │               │
       │   'completed',      │                      │               │
       │   video_url }───────┤                      │               │
       │                     │                      │               │
       │ gRPC Process()──────────────────────────►  │               │
       │◄─ r2_url ───────────────────────────────── │               │
       │                     │                      │               │
       │ Mux ingest ─────────────────────────────────────────────►  │
       │◄─ playback_id ──────────────────────────────────────────── │
```

**HeyGen v3 polling replaced by Temporal WaitForSignal activity in production:**
```go
// Temporal activity registers HeyGen video_id
// Webhook from HeyGen → Go handler → NATS publish → Temporal signal
// No polling in Temporal workflow
```

---

### Flow 4 — Export Bundle Creation

```
Browser    API GW   asset-svc   NATS   asset worker   R2 / CDN   Supabase
   │          │        │          │          │              │          │
   │ POST     │        │          │          │              │          │
   │ /exports ├────────►           │          │              │          │
   │          │        │ Validate  │          │              │          │
   │          │        │ ownership │          │              │          │
   │          │        │ INSERT    │          │              │          │
   │          │        │ export ───────────────────────────────────────►
   │          │        │ NATS pub  │          │              │          │
   │          │        │ asset.export.create ─►              │          │
   │◄─ export_id ──────┤           │          │              │          │
   │          │        │           │ consume  │              │          │
   │          │        │           ├──────────►              │          │
   │          │        │           │          │ Fetch assets─►          │
   │          │        │           │          │◄─ MP4s ──────┤          │
   │          │        │           │          │ ZIP + naming │          │
   │          │        │           │          │ Upload ──────►          │
   │          │        │           │          │◄─ signed URL ┤          │
   │          │        │           │          │ UPDATE export─────────► │
   │          │        │ NATS pub  │          │              │          │
   │          │        │ asset.export.ready ◄─┤              │          │
   │◄─Realtime: export ready + signed URL    │              │          │
```

---

### Flow 5 — Performance Signal Loop (V2)

```
NATS Scheduler   signal-svc   Meta/TikTok API   Supabase          scoring-svc
      │               │               │               │                 │
      │ signal.sync   │               │               │                 │
      ├───────────────►               │               │                 │
      │               │ OAuth token   │               │                 │
      │               │ refresh ──────►               │                 │
      │               │◄─ access_token┤               │                 │
      │               │               │               │                 │
      │               │ GET insights  │               │                 │
      │               │ per asset_id ─►               │                 │
      │               │◄─ { ctr, vtr, roas, spend }   │                 │
      │               │               │               │                 │
      │               │ INSERT INTO   │               │                 │
      │               │ video_performance_events ──────►                │
      │               │               │               │                 │
      │               │ REFRESH MATERIALIZED VIEW creative_scores ──────►
      │               │               │               │                 │
      │               │               │               │ SELECT from     │
      │               │               │               │ creative_scores ►
      │               │               │               │◄─ scores ───────┤
      │               │               │               │                 │
      │               │               │               │ POST /score ────►
      │               │               │               │◄─ predicted metrics
      │               │               │               │                 │
      │               │               │  UPDATE assets.predicted_score ─►
```

---

## Brief Engine — LLM Pipeline Detail

The brief pipeline runs as a tRPC mutation in Next.js BFF, then persists via Go API.
Go API persists to Supabase. **No LLM calls in Go workers.**

```typescript
// src/apps/web/src/server/trpc/routers/briefs.ts
import { generateObject } from 'ai'
import { anthropic } from '@ai-sdk/anthropic'
import { openai } from '@ai-sdk/openai'

export async function generateBrief(url: string, jobId: string) {
  // Stage 1: Trigger scrape via Go API → NATS → ingestion-svc
  await goApi.post('/v1/briefs/scrape', { url, jobId })
  // Supabase Realtime now handles progress updates to browser

  // Stage 2: Wait for scrape completion signal
  // ingestion-svc publishes brief.generate → brief-svc consumes
  // brief-svc runs LLM pipeline:

  // GPT-4o: structured product extraction
  const { object: product } = await generateObject({
    model: openai('gpt-4o'),
    schema: ProductExtractionSchema,
    prompt: buildExtractionPrompt(scrapedHtml),
  })

  // Claude Sonnet 4.6: creative angle generation
  const { object: angles } = await generateObject({
    model: anthropic('claude-sonnet-4-6'),
    schema: AnglesGenerationSchema,
    prompt: buildAnglesPrompt(product, brandKit),
  })

  // Claude Sonnet 4.6: hook variations per angle
  const { object: hooks } = await generateObject({
    model: anthropic('claude-sonnet-4-6'),
    schema: HooksGenerationSchema,
    prompt: buildHooksPrompt(angles, product),
  })

  // Persist to Supabase via Go API
  await goApi.put(`/v1/briefs/${briefId}`, { product, angles, hooks })
}
```

---

## Video Pipeline — Temporal Activity Implementations

```go
// Activity: SelectVideoProvider
func (a *Activities) SelectVideoProvider(ctx context.Context, req VideoJobInput) (ProviderSelection, error) {
    // Check fal.ai concurrency semaphore (max 2 per org)
    acquired, err := a.redis.SetNX(ctx,
        fmt.Sprintf("fal:semaphore:%s", req.OrgID),
        req.JobID, 30*time.Minute)
    if !acquired {
        return ProviderSelection{}, temporal.NewApplicationError("concurrency limit reached", "CONCURRENCY_LIMIT")
    }
    return ProviderSelection{
        Provider: selectModelForFormat(req.Format, req.Tier),
        Reserved: true,
    }, nil
}

// Activity: SubmitToFal
func (a *Activities) SubmitToFal(ctx context.Context, sel ProviderSelection, req VideoJobInput) (string, error) {
    resp, err := a.falClient.Queue.Submit(ctx, fal.QueueRequest{
        Model:   sel.Provider.FalModelID,
        Input:   buildFalInput(req),
        Webhook: a.webhookBase + "/webhooks/fal",
    })
    if err != nil {
        return "", fmt.Errorf("fal submission: %w", err)
    }
    // Store fal_request_id → job_id mapping in Redis (for webhook routing)
    a.redis.Set(ctx, "fal:req:"+resp.RequestID, req.JobID, 30*time.Minute)
    return resp.RequestID, nil
}

// Activity: WaitForFalCompletion
// Temporal signals replace polling loops
func (a *Activities) WaitForFalCompletion(ctx context.Context, requestID string) (FalResult, error) {
    // Webhook arrives → NATS → Temporal.SignalWorkflow("fal_complete", result)
    // This activity parks until the signal arrives (Temporal manages the wait)
    ch := workflow.GetSignalChannel(ctx, "fal_complete")
    var result FalResult
    ch.Receive(ctx, &result)
    return result, nil
}
```

---

## Infrastructure Diagram

```
┌─── VERCEL ──────────────────────────────────────────────────┐
│  Next.js 15 (SSR + serverless API routes)                   │
│  Supabase Realtime subscription (browser WS)               │
│  Preview deploys on every PR                                │
└──────────────────────────────────────────────────────────────┘

┌─── RAILWAY ──────────────────────────────────────────────────┐
│  api-gateway           (Go, 2 replicas, autoscale)          │
│  identity-svc          (Go, 1 replica)                      │
│  ingestion-svc         (Go, scale-to-zero, NATS-driven)     │
│  brief-svc             (Go, autoscale CPU)                  │
│  media-orchestrator    (Go Temporal worker, autoscale)      │
│  media-postprocessor   (Rust, autoscale CPU, spot OK)       │
│  asset-svc             (Go, 1 replica)                      │
│  signal-svc            (Go, 1 replica — V2)                 │
│  scoring-svc           (Python, 1 replica — V2)             │
│  temporal-server       (Temporal OSS, 3-node)               │
│  nats-cluster          (NATS JetStream, 3-node)             │
└──────────────────────────────────────────────────────────────┘

┌─── SUPABASE ─────────────────────────────────────────────────┐
│  PostgreSQL 16 (primary DB + RLS + Realtime)                │
│  pgvector (V2 semantic search)                              │
└──────────────────────────────────────────────────────────────┘

┌─── CLOUDFLARE ───────────────────────────────────────────────┐
│  R2 (video + image storage, zero egress)                    │
│  CDN (static asset delivery)                                │
└──────────────────────────────────────────────────────────────┘

┌─── EXTERNAL APIs ────────────────────────────────────────────┐
│  fal.ai    → Veo 3.1 / Kling 3.0 / Runway Gen-4.5          │
│  HeyGen v3 → Avatar lip-sync (developers.heygen.com)        │
│  Tavus v2  → Avatar lip-sync (fallback, Phase 9+)           │
│  ElevenLabs → TTS (eleven_v3 / eleven_flash_v2_5)           │
│  Mux       → HLS video delivery + analytics                 │
│  Modal     → Playwright scraping (serverless)               │
│  Upstash   → Rate-limit counters + URL cache (HTTP only)    │
└──────────────────────────────────────────────────────────────┘
```

---

## Non-Functional Requirements

| NFR | Target | Mechanism |
|---|---|---|
| Brief generation | < 25s P95 | Sequential LLM calls in brief-svc; Modal scrape cached 24h |
| Video generation | 60–180s | Temporal workflow; user notified via Supabase Realtime |
| Export download | < 30s | Pre-assembled ZIP on job complete; CDN-accelerated R2 URL |
| API response (sync) | < 200ms P95 | Go Echo; Upstash Redis cache for reads |
| Uptime | 99.5% SLO | Railway health checks; Vercel edge redundancy; Temporal retry |
| Concurrent generations | 50 parallel | Temporal workflow concurrency; fal.ai scales on demand |
| Data isolation | 100% org isolation | Supabase RLS on all tables + Realtime channel auth |
| Message delivery | At-least-once | NATS JetStream AckExplicit + MaxDeliver=3 + DLQ |
| Workflow durability | Crash-safe | Temporal persists workflow state across process restarts |
| GDPR compliance | On account deletion | Cascade delete org data; R2 lifecycle rules; signal retention |

---

*System Architecture v2.0 — Qvora*
*April 16, 2026 — Confidential*
