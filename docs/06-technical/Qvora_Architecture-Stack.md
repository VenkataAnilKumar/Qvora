# QVORA
## Architecture & Tech Stack
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Proposal

---

## Architecture Pattern

Qvora runs three distinct workload profiles:
- **Interactive** — brief generation, edits, real-time UI (seconds)
- **Async compute** — video generation, rendering, export (minutes, GPU-bound)
- **Scheduled / event-driven** — performance signal ingestion, learning loop (V2)

```
CLIENT (Next.js 15)
    │  REST + SSE
API GATEWAY / BFF (tRPC + Next.js API Routes)
    │  Sync                    │  Enqueue jobs
CORE API (Go · Echo)      JOB QUEUE (asynq + Redis)
    │                          │
    └──────── AI ORCHESTRATION LAYER ──────────
              Brief Engine      Video Pipeline     Signal (V2)
              Vercel AI SDK v6   FAL.AI             Temporal
              GPT-4o / Claude   ElevenLabs
                                HeyGen Avatar API
    │
DATA LAYER
    PostgreSQL (Supabase) · Redis (Upstash) · R2 (Cloudflare)
```

---

## Stack by Layer

### Frontend
| Concern | Choice |
|---|---|
| Framework | Next.js 15 (App Router) |
| UI | shadcn/ui + Radix + Tailwind CSS v4 |
| State | Zustand + TanStack Query v5 |
| Real-time | Server-Sent Events (generation progress — see note below) |
| Animation | Framer Motion |
| Video preview | Mux Player |
| Forms | React Hook Form + Zod |
| Monorepo | Turborepo |

### Backend API
| Concern | Choice |
|---|---|
| BFF | tRPC (end-to-end type safety) |
| Core API server | **Go · Echo v4** |
| Job workers | **Go · asynq** (Redis-backed, BullMQ equivalent) |
| Video post-processor | **Rust · Axum + ffmpeg-sys** (CPU-bound only) |
| Validation | Zod (frontend) · Go structs + sqlc (backend) |
| Auth | Clerk (multi-tenant, JWT, SSO) |
| Rate limiting | Upstash Redis (sliding window, per-user quota — HTTP/REST; safe for serverless) |
| URL extraction | Go orchestrator → Playwright on Modal (isolated containers) |

### AI Orchestration
| Concern | Choice |
|---|---|
| LLM calls | Vercel AI SDK v6 (streaming + structured outputs) |
| Brief model | GPT-4o (structured output / JSON strict mode) |
| Regeneration | Claude Sonnet 4.6 (creative quality, lower cost) |
| Video generation | FAL.AI (gateway to Veo 3.1 / Kling 3.0 / Runway Gen-4.5) |
| TTS | ElevenLabs API (voice cloning, 175+ languages) |
| Avatar lip-sync | HeyGen Avatar API v3 |
| Prompt management | Langfuse (versioning, A/B, cost tracking) |
| Observability | OpenTelemetry + Langfuse |

### Data Layer
| Concern | Choice |
|---|---|
| Primary DB | PostgreSQL via Supabase (RLS for multi-tenancy) |
| ORM / queries | sqlc (type-safe SQL → Go codegen) |
| Cache + queue | Upstash Redis (HTTP/REST — caching, session, rate limiting only. **Not used for asynq job queues** — see note below) |
| File / video storage | Cloudflare R2 (zero egress cost) |
| CDN | Cloudflare CDN |
| Video streaming | Mux (adaptive HLS, preview analytics) |
| Embeddings (V2) | pgvector on Supabase |

### Infrastructure
| Concern | Choice |
|---|---|
| Frontend + BFF | Vercel |
| API + workers | Railway (autoscale, Go binaries) |
| Scraping containers | Modal (serverless, pay-per-second) |
| Secrets | Doppler |
| CI/CD | GitHub Actions + Turborepo remote cache |
| Error tracking | Sentry |
| Logs + uptime | Better Stack |
| Product analytics | PostHog |
| Payments | Stripe (subscription billing + usage metering) |

---

## Go vs Rust Decision

| Service | Language | Reason |
|---|---|---|
| API server | Go | I/O-bound, goroutine concurrency, fast iteration |
| Job workers | Go | Goroutine-per-job, asynq Redis queues |
| URL extractor orchestrator | Go | Orchestrates Playwright container calls |
| SSE progress streaming | Next.js Route Handler | `ReadableStream` at `/api/generation/[jobId]/stream` — polls Redis for job status |
| **Video post-processor** | **Rust** | CPU-bound (watermark, captions, transcode, reframe) |

Go is chosen over Rust for the API and workers because Qvora's backend is **overwhelmingly I/O-bound** — latency is owned by FAL.AI, OpenAI, and HeyGen, not the application server. Rust's borrow checker adds friction during rapid schema iteration with no meaningful throughput benefit on network waits.

Rust is justified **only** for the video post-processing microservice (logo overlay, safe-zone compliance, caption burn, format reframing) where work is genuinely CPU-bound and zero-copy memory matters.

---

## Architecture Notes

### Redis — Two Separate Instances

`asynq` (the job queue for Go workers) requires a **persistent TCP connection** to Redis. Upstash Redis is HTTP/REST-based (serverless) and does not support persistent TCP connections on standard plans — using Upstash for asynq will fail in production.

**Two Redis instances are required:**

| Instance | Provider | Purpose |
|---|---|---|
| **Redis A — Job queue** | Railway Redis container | asynq job queues (persistent TCP, long-lived connections) |
| **Redis B — App cache** | Upstash Redis | Rate limiting, session cache, brief cache (HTTP/REST, serverless-safe) |

Both are reflected in the infrastructure cost estimate.

### SSE — Standalone Route Handler, Not tRPC

Generation progress is delivered via **Server-Sent Events (SSE)**. tRPC subscriptions use WebSockets; they cannot be used for SSE. The progress stream must be implemented as a standalone Next.js Route Handler:

```
GET /api/generation/[jobId]/stream
  → ReadableStream (text/event-stream)
  → polls asynq job status from Redis A
  → emits: { status, percent, step } events
  → client: EventSource('/api/generation/[jobId]/stream')
```

tRPC handles all standard API calls. The SSE endpoint is the only non-tRPC surface in the BFF.

---

## V1 → V2 Upgrade Path

| Component | V1 | V2 |
|---|---|---|
| Workflows | asynq simple queues | Temporal (durable, multi-step) |
| Feedback loop | None | Meta / TikTok Ads API → Qvora Signal DB |
| Brief generation | Static LLM prompts | Performance-weighted prompt routing |
| Model routing | FAL.AI single call | Multi-model routing by cost / quality SLA |
| Search | None | pgvector semantic brief + asset search |

---

## Estimated Infrastructure Cost (MVP · ~500 active users)

| Service | Est. /mo |
|---|---|
| Vercel Pro | $20 |
| Supabase Pro | $25 |
| Upstash Redis | $10 |
| Railway (2× workers) | $40 |
| Cloudflare R2 + CDN | $15 |
| Mux | $30–50 |
| Sentry + Better Stack | $20 |
| PostHog | $0–20 |
| **Total infra (excl. AI APIs)** | **~$160–180/mo** |

> FAL.AI video generation is the dominant cost driver. GEN-14 (profitable unit economics at $99/mo Starter) requires credit metering from day one.
