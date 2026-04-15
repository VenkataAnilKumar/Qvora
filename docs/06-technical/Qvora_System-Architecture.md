# QVORA
## System Architecture
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft
**Stack reference:** `Qvora_Architecture-Stack.md`

---

## Architecture Overview

Qvora runs three distinct workload profiles, each with different latency and compute requirements:

| Profile | Latency | Compute | Examples |
|---|---|---|---|
| **Interactive** | < 2s | CPU / LLM API | Brief generation, UI updates, auth |
| **Async compute** | 1–5 min | GPU (via FAL.AI) | Video generation, rendering, export |
| **Scheduled** | Hourly/daily | CPU | Signal ingestion, fatigue detection (V2) |

---

## High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CLIENT LAYER                                                               │
│  Next.js 15 (Vercel)                                                        │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  App Router pages · shadcn/ui · TanStack Query · Framer Motion      │   │
│  │  tRPC client · SSE listener (generation progress)                   │   │
│  └───────────────────────────┬─────────────────────────────────────────┘   │
└──────────────────────────────┼──────────────────────────────────────────────┘
                               │ tRPC (internal) / REST (public API)
┌──────────────────────────────▼──────────────────────────────────────────────┐
│  BFF LAYER                                                                  │
│  Next.js API Routes + tRPC Router (Vercel)                                  │
│  Auth: Clerk JWT validation · Rate limit: Upstash Redis                     │
└────────┬───────────────────────────────────────────────────────────────┬────┘
         │ REST (Go HTTP)                                                 │ SSE stream
┌────────▼────────────────────────────────────────────────────────────┐  │
│  CORE API                                                           │  │
│  Go · Echo v4 (Railway)                                             │  │
│  ├── /briefs          Brief CRUD, angle/hook management             │  │
│  ├── /generations     Job submission, status polling                │◄─┘
│  ├── /assets          Asset library, metadata                       │
│  ├── /exports         Export bundle creation                        │
│  ├── /brands          Brand kit management                          │
│  ├── /team            Invite, roles, seats                          │
│  └── /signal          Ad account, metrics (V2)                      │
│                                                                     │
│  sqlc → PostgreSQL (Supabase)                                       │
└────────┬────────────────────────────────────────────────────────────┘
         │ Enqueue job (asynq)
┌────────▼────────────────────────────────────────────────────────────┐
│  JOB QUEUE                                                          │
│  asynq + Upstash Redis                                              │
│  ├── brief:extract     Playwright scraping job                      │
│  ├── generation:video  FAL.AI video generation job                  │
│  ├── generation:export ZIP bundle assembly job                      │
│  └── signal:sync       Ad account metrics pull (V2)                 │
└────────┬────────────────────────────────────────────────────────────┘
         │
┌────────▼────────────────────────────────────────────────────────────┐
│  AI ORCHESTRATION LAYER                                             │
│                                                                     │
│  ┌─────────────────────────┐  ┌─────────────────────────────────┐  │
│  │  BRIEF ENGINE           │  │  VIDEO PIPELINE                 │  │
│  │  Vercel AI SDK v6        │  │  Go workers (Railway)           │  │
│  │  (TypeScript / Next.js) │  │                                 │  │
│  │                         │  │  FAL.AI API                     │  │
│  │  1. URL Extract (Modal) │  │  ├── Veo 3.1 (T2V premium)      │  │
│  │  2. Parse — GPT-4o      │  │  ├── Kling 3.0 (T2V/I2V)        │  │
│  │  3. Angles — Claude 4.6 │  │  └── Runway Gen-4.5 (control)   │  │
│  │  4. Hooks — Claude 4.6  │  │                                 │  │
│  │  5. Format rec          │  │  HeyGen Avatar v3 API (V2V)     │  │
│  │  6. Zod validate+retry  │  │  ElevenLabs API (TTS/clone)     │  │
│  └─────────────────────────┘  │                                 │  │
│                                │  Rust · ffmpeg-sys              │  │
│  Parse: GPT-4o (JSON strict)   │  (watermark, captions,          │  │
│  Creative: Claude Sonnet 4.6   │   transcode, reframe)           │  │
│  Prompt mgmt: Langfuse         └─────────────────────────────────┘  │
└────────┬────────────────────────────────────────────────────────────┘
         │
┌────────▼────────────────────────────────────────────────────────────┐
│  DATA LAYER                                                         │
│  PostgreSQL (Supabase) · Upstash Redis · Cloudflare R2 · Mux       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Sequence Diagrams

### Flow 1 — Brief Generation (URL → Structured Brief)

```
User          BFF (Next.js)      Core API (Go)    Brief Engine (Vercel AI SDK)   DB
 │                │                   │                      │                   │
 │ POST /briefs   │                   │                      │                   │
 │ {url, brand}  ├─ tRPC mutation ───►                       │                   │
 │               │  createBrief()     │ INSERT brief          │                   │
 │               │                   │ status=pending ───────────────────────────►
 │               │                   │◄── brief_id ──────────────────────────────┤
 │◄── {brief_id} ┤◄── brief_id ──────┤                      │                   │
 │               │                   │                      │                   │
 │ [SSE connect] │ GET /api/generation│                      │                   │
 │ ──────────────►  /[briefId]/stream │                      │                   │
 │               │                   │                      │                   │
 │               │  Modal webhook ──────────────────────────►                   │
 │               │                   │                      │ Playwright →       │
 │               │                   │                      │ scrape URL         │
 │               │                   │                      │                   │
 │◄─ SSE 15% ────┤ emit: extracting  │                      │ generateObject()  │
 │               │                   │                      │ GPT-4o → product  │
 │◄─ SSE 35% ────┤ emit: parsing     │                      │                   │
 │               │                   │                      │ generateObject()  │
 │               │                   │                      │ Claude 4.6 → angles│
 │◄─ SSE 65% ────┤ emit: generating  │                      │                   │
 │               │                   │                      │ generateObject()  │
 │               │                   │                      │ Claude 4.6 → hooks│
 │◄─ SSE 85% ────┤ emit: validating  │                      │ Zod validate      │
 │               │                   │                      │                   │
 │               │  REST → Core API ─►                       │                   │
 │               │                   │ UPDATE brief          │                   │── bulk insert
 │               │                   │ INSERT angles+hooks   │                   │── bulk insert
 │               │                   │ status=ready ─────────────────────────────►
 │◄─ SSE 100% ───┤ emit: complete    │                      │                   │
```

**Latency budget:**
- Playwright scrape (Modal): 3–8s
- Vercel AI SDK pipeline (3 LLM calls, sequential): 6–14s
- Total user-visible: 10–22s (matches < 30s target)

---

### Flow 2 — Video Generation (Brief → Generated Asset)

```
User        BFF       Core API (Go)    asynq Queue    Go Worker    FAL.AI    HeyGen    Rust ffmpeg    R2 / Mux
 │           │             │               │              │           │          │           │           │
 │ POST      │             │               │              │           │          │           │           │
 │ /generations            │               │              │           │          │           │           │
 │ {brief_id,├─────────────►               │              │           │          │           │           │
 │  format,  │             │ INSERT job    │              │           │          │           │           │
 │  angles}  │             │ status=queued │              │           │          │           │           │
 │           │             ├── Enqueue ───►               │           │          │           │           │
 │           │             │               │ Dequeue      │           │          │           │           │
 │◄─ job_ids ┤◄── job_ids ─┤               ├── Run ──────►            │          │           │           │
 │           │             │               │              │           │          │           │           │
 │ [SSE]     │             │               │              │ Script gen│          │           │           │
 │           │             │               │              │ (from hook│          │           │           │
 │           │             │               │              │  + angle) │          │           │           │
 │           │             │               │              │           │          │           │           │
 │           │             │               │   ┌── UGC/Demo/T2V ──────►          │           │           │
 │           │             │               │   │          │  POST     │          │           │           │
 │           │             │               │   │          │  /queue   │          │           │           │
 │           │             │               │   │          │◄─task_id ─┤          │           │           │
 │◄─ SSE 20% ┤             │               │   │          │           │          │           │           │
 │           │             │               │   └── V2V ──────────────────────────►           │           │
 │           │             │               │              │  HeyGen   │          │           │           │
 │           │             │               │              │  webhook  │          │           │           │
 │◄─ SSE 60% ┤             │               │              │           │          │           │           │
 │           │             │               │              │◄─ video URL (raw) ───┤           │           │
 │           │             │               │              │                      │           │           │
 │           │             │               │              │ Post-process ────────────────────►           │
 │           │             │               │              │ (logo, captions,     │           │           │
 │           │             │               │              │  reframe, transcode) │           │           │
 │           │             │               │              │◄─ processed.mp4 ─────────────────┤           │
 │           │             │               │              │                      │           │           │
 │           │             │               │              │ Upload ──────────────────────────────────────►
 │           │             │               │              │◄─ r2_url + mux_id ──────────────────────────┤
 │           │             │               │              │                      │           │           │
 │           │             │ UPDATE asset  │◄── complete ─┤                      │           │           │
 │           │             │ status=ready  │              │                      │           │           │
 │◄─ SSE 100%┤◄─ SSE push ─┤               │              │                      │           │           │
 │  complete │ asset:ready │               │              │                      │           │           │
```

**Latency budget per video:**
- Script generation: 2–5s
- FAL.AI queue + generation: 30–150s (model-dependent)
- HeyGen lip-sync (V2V): 60–180s
- Rust post-processing: 5–15s
- Upload to R2: 3–8s
- **Total: 60–180s** (matches spec)

---

### Flow 3 — Export (Assets → ZIP Download)

```
User        BFF       Core API (Go)    asynq Queue    Go Worker    R2/CDN
 │           │             │               │              │           │
 │ POST      │             │               │              │           │
 │ /exports  ├─────────────►               │              │           │
 │ {asset_ids│             │ Validate      │              │           │
 │  format}  │             │ asset ownership              │           │
 │           │             ├── Enqueue ───►               │           │
 │           │             │   export job  │              │           │
 │◄─ export_id┤◄─ export_id ┤               │ Dequeue      │           │
 │           │             │               ├── Run ──────►            │
 │           │             │               │              │ Fetch     │
 │           │             │               │              │ assets ──►│
 │           │             │               │              │◄─ MP4s ───┤
 │           │             │               │              │           │
 │           │             │               │              │ Apply     │
 │           │             │               │              │ naming    │
 │           │             │               │              │ convention│
 │           │             │               │              │ Generate  │
 │           │             │               │              │ manifest  │
 │           │             │               │              │ .csv      │
 │           │             │               │              │           │
 │           │             │               │              │ ZIP ──────►
 │           │             │               │              │◄─ signed URL (48hr)
 │           │             │ UPDATE export │◄── complete ─┤           │
 │           │             │ status=ready  │              │           │
 │◄─ signed URL┤◄─ URL ────┤               │              │           │
 │ [auto-download]         │               │              │           │
```

---

### Flow 4 — Signal Ingestion (V2)

```
Scheduler     Core API     asynq     Go Worker     Meta API     DB (signal tables)
    │              │          │           │              │               │
    │ Cron (6hr)  │          │           │              │               │
    ├─────────────►           │           │              │               │
    │              │ Enqueue  │           │              │               │
    │              ├──────────►           │              │               │
    │              │          │ Dequeue   │              │               │
    │              │          ├───────────►              │               │
    │              │          │           │ OAuth token  │               │
    │              │          │           │ refresh ─────►               │
    │              │          │           │◄─ access_token               │
    │              │          │           │              │               │
    │              │          │           │ GET insights │               │
    │              │          │           │ per ad_id ───►               │
    │              │          │           │◄─ {ctr,cpa,  │               │
    │              │          │           │   roas,hold} │               │
    │              │          │           │              │               │
    │              │          │           │ Match ad_id → asset.ad_id    │
    │              │          │           │ UPSERT asset_metrics ────────►
    │              │          │           │              │               │
    │              │          │           │ Run fatigue  │               │
    │              │          │           │ detection    │               │
    │              │          │           │ algorithm    │               │
    │              │          │           │ UPDATE assets│               │
    │              │          │           │ fatigue_detected_at ─────────►
```

---

## Brief Engine — Vercel AI SDK Pipeline

The brief pipeline runs as a tRPC mutation in the Next.js BFF. No separate Python service. Structured output is handled via `generateObject()` with Zod schemas; retries are a simple async loop.

```typescript
// src/apps/web/src/server/routers/briefs.ts
import { generateObject } from 'ai';
import { anthropic } from '@ai-sdk/anthropic';
import { openai } from '@ai-sdk/openai';
import { z } from 'zod';

const ProductSchema = z.object({
  brand_name: z.string(),
  product_name: z.string(),
  target_audience: z.string(),
  pain_points: z.array(z.string()),
  usp: z.string(),
  tone: z.enum(['bold', 'conversational', 'professional', 'playful']),
  color_palette: z.array(z.string()),
  visual_direction: z.string(),
});

const AnglesSchema = z.object({
  angles: z.array(z.object({
    id: z.string(),
    type: z.enum(['problem_solution', 'social_proof', 'urgency', 'lifestyle', 'feature_focus']),
    headline: z.string(),
    hook: z.string(),
    body: z.string(),
    cta: z.string(),
    recommended_format: z.enum(['vertical_9x16', 'square_1x1', 'horizontal_16x9']),
    recommended_duration: z.number(),
  }))
});

export async function generateBrief(url: string, briefId: string, emit: SSEEmitter) {
  // Stage 1: Scrape URL via Modal → Playwright
  emit({ stage: 'extracting', percent: 10 });
  const html = await scrapeUrl(url); // Modal webhook call

  // Stage 2: Parse product data (GPT-4o — JSON strict mode, reliable extraction)
  emit({ stage: 'parsing', percent: 30 });
  const { object: product } = await generateObject({
    model: openai('gpt-4o'),
    schema: ProductSchema,
    prompt: `Extract product and brand information from this page:\n\n${html}`,
  });

  // Stage 3: Generate creative angles (Claude Sonnet 4.6 — creative quality)
  emit({ stage: 'generating_angles', percent: 55 });
  let brief: z.infer<typeof AnglesSchema> | null = null;
  let retries = 0;

  while (!brief && retries < 2) {
    try {
      const { object } = await generateObject({
        model: anthropic('claude-sonnet-4-6'),
        schema: AnglesSchema,
        prompt: buildAnglesPrompt(product, url),
      });
      brief = object;
    } catch {
      retries++;
    }
  }

  if (!brief) throw new Error('Brief generation failed after 2 retries');

  // Stage 4: Validate + persist via Go API
  emit({ stage: 'validating', percent: 85 });
  await goApi.updateBrief(briefId, { status: 'ready', content: { ...product, ...brief } });

  emit({ stage: 'complete', percent: 100, briefId });
}
```

**LLM routing:**
| Stage | Model | Reason |
|---|---|---|
| Product parse | `gpt-4o` | JSON strict mode, reliable structured extraction |
| Angles + hooks | `claude-sonnet-4-6` | Creative quality, tone-aware output |
| Retry / regen | `claude-sonnet-4-6` | Variant diversity on manual regeneration |

---

## Video Pipeline — FAL.AI Integration

```go
// Go worker — FAL.AI async queue pattern

type VideoJob struct {
    JobID     string
    BriefID   string
    AngleID   string
    Format    string  // "t2v" | "i2v" | "v2v" | "ugc" | "demo"
    Model     string  // "veo-3.1" | "kling-3.0" | "runway-gen4"
    Prompt    string
    AssetURLs []string
}

func processVideoJob(ctx context.Context, job VideoJob) error {
    // 1. Submit to FAL.AI async queue
    resp, err := falClient.Queue.Submit(ctx, fal.QueueRequest{
        Model:   modelMap[job.Model],  // fal model ID
        Input:   buildFALInput(job),
        Webhook: webhookURL + "/fal/callback",
    })

    // 2. Store request_id → job mapping in Redis
    redis.Set(ctx, "fal:"+resp.RequestID, job.JobID, 30*time.Minute)

    // 3. Update job status in DB
    db.UpdateJobStatus(ctx, job.JobID, "processing", resp.RequestID)

    return nil
}

// FAL.AI webhook handler
func handleFALCallback(c echo.Context) error {
    var payload fal.WebhookPayload
    c.Bind(&payload)

    jobID, _ := redis.Get(ctx, "fal:"+payload.RequestID)

    if payload.Status == "OK" {
        // Route to post-processing
        queue.Enqueue("postprocess:video", PostProcessJob{
            JobID:    jobID,
            VideoURL: payload.Output.VideoURL,
        })
    } else {
        db.UpdateJobStatus(ctx, jobID, "failed", payload.Error)
        sse.Push(jobID, "generation:failed")
    }
    return c.NoContent(200)
}
```

---

## Multi-Tenancy Pattern

All data is isolated by `org_id` (Clerk organization ID). Isolation enforced at two layers:

**Layer 1 — API layer (Go):** Every query includes `org_id` extracted from JWT claims.
```go
// Middleware — extract org from Clerk JWT
func OrgMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        claims := c.Get("clerk_claims").(jwt.Claims)
        orgID := claims["org_id"].(string)
        c.Set("org_id", orgID)
        return next(c)
    }
}
```

**Layer 2 — Database (Supabase RLS):** Policies enforce org isolation even if API layer is bypassed.
```sql
-- RLS policy on all tables
CREATE POLICY "org_isolation" ON briefs
  USING (org_id = auth.jwt() ->> 'org_id');
```

---

## SSE — Generation Progress Streaming

SSE runs as a **standalone Next.js Route Handler** (not Go, not tRPC). It polls job status from Redis and pushes events to the browser. This keeps it within the Vercel edge/Node.js environment where the brief pipeline already runs, and avoids an extra HTTP hop to Go for status reads.

```typescript
// src/apps/web/src/app/api/generation/[jobId]/stream/route.ts
import { redis } from '@/lib/upstash'; // Upstash HTTP Redis

export async function GET(
  _req: Request,
  { params }: { params: { jobId: string } }
) {
  const encoder = new TextEncoder();

  const stream = new ReadableStream({
    async start(controller) {
      const send = (data: object) =>
        controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`));

      const interval = setInterval(async () => {
        const raw = await redis.get<string>(`job:${params.jobId}`);
        if (!raw) return;

        const status = JSON.parse(raw);
        send(status);

        if (status.stage === 'complete' || status.stage === 'error') {
          clearInterval(interval);
          controller.close();
        }
      }, 800);
    },
  });

  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
    },
  });
}

// Job writes progress to Redis (from brief pipeline or Go video worker)
await redis.set(`job:${jobId}`, JSON.stringify({
  stage: 'generating_angles',
  percent: 55,
  message: 'Writing creative angles...',
}), { ex: 300 }); // 5min TTL
```

**Note:** For brief generation (runs in Next.js BFF), the pipeline writes directly to Redis. For video generation (runs in Go worker), the Go worker writes job status to Upstash Redis via HTTP. Both share the same `job:{jobId}` key schema.

---

## Infrastructure Diagram

```
┌─── VERCEL ──────────────────────┐  ┌─── RAILWAY ──────────────────────────┐
│  Next.js 15 (App + API Routes)  │  │  Go Core API (Echo v4)               │
│  tRPC router                    │  │  Go Workers (asynq consumers)        │
│  Static assets via CDN          │  │  Rust post-processor (ffmpeg-sys)    │
└──────────────┬──────────────────┘  └──────────────┬───────────────────────┘
               │                                    │
┌─── SUPABASE ─▼──────────────────┐  ┌─── REDIS (two instances) ───────────┐
│  PostgreSQL (primary DB)        │  │  Railway Redis (TCP)                 │
│  RLS multi-tenancy              │  │  └── asynq job queues                │
│  pgvector (V2 embeddings)       │  │  Upstash Redis (HTTP)                │
└─────────────────────────────────┘  │  └── SSE status · cache · rate-limit │
                                     └──────────────────────────────────────┘
               │                                    │
┌─── CLOUDFLARE ──────────────────┐  ┌─── MODAL ────▼──────────────────────┐
│  R2 object storage (videos)     │  │  Playwright scraping containers      │
│  CDN (asset delivery)           │  │  Serverless, pay-per-second          │
└─────────────────────────────────┘  └──────────────────────────────────────┘
               │
┌─── MUX ─────▼───────────────────┐  ┌─── EXTERNAL AI APIs ────────────────┐
│  Video streaming (HLS)          │  │  FAL.AI  → Veo 3.1 / Kling / Runway │
│  In-app preview analytics       │  │  HeyGen  → Avatar v3 (V2V)          │
└─────────────────────────────────┘  │  ElevenLabs → TTS / Voice clone      │
                                     │  Claude Sonnet 4.6 → Brief gen       │
                                     │  Langfuse → Prompt observability     │
                                     └──────────────────────────────────────┘
```

---

## Non-Functional Requirements

| NFR | Target | Implementation |
|---|---|---|
| Brief generation | < 30s P95 | Vercel AI SDK sequential LLM calls; cached Modal scrape results |
| Video generation | 60–180s | Async queue; user notified via SSE + email |
| Export download | < 30s | CDN-accelerated R2 bundle; pre-built on job complete |
| API response (sync) | < 200ms P95 | Go Echo; Upstash Redis cache for reads |
| Uptime | 99.5% (Growth/Scale) | Railway health checks; Vercel edge redundancy |
| Concurrent generations | 50 parallel | asynq worker concurrency; FAL.AI scales on demand |
| Data isolation | 100% | RLS enforced at DB layer; org_id on all queries |
| GDPR compliance | On account deletion | Cascade delete org data; R2 lifecycle rules |

---

*System Architecture v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources:**
- [Vercel AI SDK — generateObject](https://ai-sdk.dev/docs/ai-sdk-core/generating-structured-data)
- [FAL.AI Queue API Reference](https://docs.fal.ai/model-apis/model-endpoints/queue)
- [Supabase RLS Best Practices — Makerkit](https://makerkit.dev/blog/tutorials/supabase-rls-best-practices)
- [REST vs tRPC 2026 — DEV Community](https://dev.to/pockit_tools/rest-vs-graphql-vs-trpc-vs-grpc-in-2026-the-definitive-guide-to-choosing-your-api-layer-1j8m)
- [HeyGen API v3 — developers.heygen.com](https://developers.heygen.com)
