# QVORA
## Architecture & Tech Stack
**Version:** 2.1 | **Date:** April 16, 2026 | **Status:** Transitional (V1 Runtime Active, Phase 8+ Target)
**Reference:** `Qvora_Microservice-Architecture.md` (full service catalogue)

> Current runtime status and migration sequencing are tracked in `docs/07-implementation/Qvora_Implementation-Phases.md`.

---

## Architecture Pattern

Qvora is a **polyglot microservice system** decomposed by bounded domain context:

```
CLIENT (Next.js 15)
    │  HTTPS
API GATEWAY (Go · Echo v4)
    │  gRPC (internal)          │  NATS JetStream (async)
DOMAIN SERVICES                 MESSAGE BUS
  identity-svc  (Go)            ingestion.*, brief.*, media.*
  ingestion-svc (Go)            asset.*, signal.*
  brief-svc     (Go)
  asset-svc     (Go)
  signal-svc    (Go)   V2       TEMPORAL WORKFLOWS
  scoring-svc   (Python) V2     VideoCreationWorkflow (durable pipeline)
    │
MEDIA PIPELINE
  media-orchestrator (Go + Temporal worker)
  media-postprocessor (Rust + gRPC)
    │
DATA LAYER
  PostgreSQL (Supabase + RLS + Realtime)
  Cloudflare R2 · Mux · Upstash Redis (rate-limit only)
```

---

## Stack by Layer

### Frontend
| Concern | Choice |
|---|---|
| Framework | Next.js 15 (App Router) |
| UI | shadcn/ui + Radix + Tailwind CSS v4 |
| State | Zustand + TanStack Query v5 |
| Real-time | **Supabase Realtime** (job progress — Postgres Changes via WebSocket) |
| Animation | Framer Motion |
| Video preview | Mux Player (signed HLS) |
| Forms | React Hook Form + Zod |
| Type-safe API | tRPC v11 (BFF layer) |
| Monorepo | Turborepo |

**Changed from v1.0:** SSE Route Handler replaced by Supabase Realtime. No custom polling loop.

---

### API Gateway
| Concern | Choice |
|---|---|
| HTTP server | **Go · Echo v4** |
| Auth validation | Clerk JWT (org_id, role, plan extracted per request) |
| Rate limiting | Upstash Redis sliding window (per-org, per-endpoint) |
| Idempotency | UUID idempotency key on all job creation mutations |
| Billing events | Stripe Meter API (`meter_event` per billable operation) |
| Internal routing | gRPC fan-out to domain services |

---

### Domain Services
| Service | Runtime | Responsibility |
|---|---|---|
| `identity-svc` | Go | Auth claims, quota checks, tier enforcement, Stripe metering |
| `ingestion-svc` | Go + Modal | URL scraping (Playwright), product data extraction, R2 storage |
| `brief-svc` | Go + Vercel AI SDK | LLM orchestration (GPT-4o + Claude), brief/angles/hooks storage |
| `asset-svc` | Go | Asset metadata, brand kits, export ZIP assembly, Mux playback URLs |
| `signal-svc` | Go | Ad platform sync, performance event ingestion, GDPR cleanup (V2) |
| `scoring-svc` | **Python · FastAPI** | Predictive creative scoring (rules → scikit-learn → PyTorch CNN) (V2) |

---

### Media Pipeline
| Concern | Choice |
|---|---|
| Workflow orchestration | **Temporal.io** (durable execution, saga, retry, visibility) |
| Video generation | FAL.AI — `fal.queue.submit()` only (never sync) |
| Provider interface | Model-agnostic Go interface: Veo 3.1 / Kling 3.0 / Runway Gen-4.5 |
| Avatar lip-sync (default) | **HeyGen Avatar API v3** (`developers.heygen.com`) |
| Avatar lip-sync (secondary) | **Tavus v2** (fallback, cost-optimized at volume) |
| TTS / Voiceover | ElevenLabs (`eleven_v3` quality / `eleven_flash_v2_5` preview) |
| Video post-processing | **Rust · Axum + ffmpeg-next** (gRPC interface, CPU-bound) |
| Internal comms | **gRPC** (Go → Rust, Go → Go hot paths) |

**Changed from v1.0:**
- asynq task chains → Temporal workflows (durable, versioned, observable)
- HTTP between services → gRPC for hot paths
- HeyGen-only → provider interface (HeyGen default + Tavus secondary)

---

### AI Orchestration
| Concern | Choice |
|---|---|
| LLM SDK | Vercel AI SDK v6 (`generateObject` with Zod schemas) |
| Product extraction | GPT-4o (JSON strict mode — reliable structured output) |
| Creative generation | Claude Sonnet 4.6 (angles, hooks, regeneration — quality + diversity) |
| URL extraction | Go orchestrator → Modal Playwright (serverless, pay-per-second) |
| LLM observability | Langfuse (prompt versions, cost per org, latency per call) |
| LLM instrumentation | OpenLLMetry (OTel extension — token metrics, gen_ai.* attributes) |

---

### Message Bus
| Concern | Choice |
|---|---|
| Async messaging | **NATS JetStream** (replaces asynq + Railway Redis for queuing) |
| Delivery guarantee | At-least-once (AckExplicit, MaxDeliver=3) |
| Dead-letter | NATS DLQ stream (`dlq.*` subjects) |
| Stream replay | Native NATS JetStream (reprocess failed jobs from any point) |
| Simple background jobs | NATS consumers (export assembly, signal sync) |
| Complex multi-step pipelines | Temporal workflows (video creation, provider failover) |

**Changed from v1.0:** Railway Redis (TCP) decommissioned for queuing. NATS is the single message bus.

---

### Data Layer
| Concern | Choice |
|---|---|
| Primary DB | PostgreSQL 16 via **Supabase** (RLS + Realtime) |
| SQL codegen | sqlc (type-safe SQL → Go structs) |
| Real-time push | **Supabase Realtime** (Postgres Changes → browser WebSocket) |
| Rate-limit counters | Upstash Redis HTTP (only remaining Redis usage) |
| URL scrape cache | Upstash Redis HTTP (24h TTL by URL hash) |
| Object storage | Cloudflare R2 (zero egress cost) |
| Video streaming | Mux (adaptive HLS, signed playback, in-app analytics) |
| Embeddings (V2) | pgvector on Supabase |

**Changed from v1.0:** Redis split (Railway TCP + Upstash HTTP) simplified. Railway Redis removed. Upstash retained for rate-limit counters and cache only.

---

### Infrastructure
| Concern | Choice |
|---|---|
| Frontend + BFF | Vercel (edge + serverless functions) |
| All backend services | **Railway** (api, 6 domain services, Temporal, NATS, postprocessor) |
| Scraping | Modal (serverless Playwright, pay-per-second) |
| Workflow engine | Temporal OSS on Railway (or Temporal Cloud for managed) |
| Message broker | NATS JetStream on Railway (3-node cluster) |
| Secrets | Doppler (dev / stg / prd environments) |
| CI/CD | GitHub Actions + Turborepo remote cache |
| Error tracking | Sentry (Next.js + Go + Rust + Python) |
| Logs | Better Stack (structured JSON log drain from Railway) |
| Traces | Grafana Tempo (OTel collector → Tempo) |
| Metrics | Prometheus + Grafana dashboards |
| Product analytics | PostHog |
| Payments | Stripe (subscriptions + Meter API for usage billing) |

---

## Language Allocation

| Service | Language | Reason |
|---|---|---|
| API Gateway | Go | I/O-bound, goroutine concurrency, Echo v4 middleware chain |
| identity-svc | Go | Simple query/response, fast startup, gRPC server |
| ingestion-svc | Go | Orchestrates Modal HTTP calls; no CPU work |
| brief-svc | Go | LLM SDK via Vercel AI SDK (TypeScript) or Go HTTP client |
| asset-svc | Go | Mux API, R2 upload, ZIP assembly — all I/O |
| signal-svc | Go | Ad platform HTTP polling; event insertion; I/O-bound |
| media-orchestrator | Go | Temporal Go SDK; fal.ai HTTP; all I/O-bound |
| **media-postprocessor** | **Rust** | CPU-bound: ffmpeg transcode, watermark, caption burn |
| **scoring-svc** | **Python** | ML ecosystem (scikit-learn, PyTorch) — no viable Go alternative |

**Rule (unchanged):** Go for I/O-bound services. Rust for CPU-bound video processing. Python only for ML inference. No language introduced outside these domains.

---

## Service Communication Rules

| Path | Protocol | Reason |
|---|---|---|
| Next.js → Go Gateway | HTTPS REST | Public-facing; TLS required |
| Go Gateway → domain services | gRPC (mTLS) | Type safety; 7× faster than REST; bi-di streaming |
| media-orchestrator → postprocessor | gRPC streaming | Stream ffmpeg progress back to workflow |
| Services → NATS | NATS protocol | Async messaging; no direct HTTP between async services |
| Services → fal.ai, HeyGen, ElevenLabs | HTTPS REST | External APIs — HTTP only |
| fal.ai → Go webhook handler | HTTPS (inbound) | SHA256 sig verified; < 500ms response |
| Mux → Go webhook handler | HTTPS (inbound) | Mux signing secret verified |

---

## Go vs Rust vs Python Decision Matrix

| Service | Language | CPU/IO | Decision |
|---|---|---|---|
| API Gateway | Go | I/O | HTTP + gRPC fan-out, no compute |
| LLM services | Go | I/O | Latency owned by OpenAI/Anthropic |
| Job orchestration | Go | I/O | Latency owned by fal.ai/HeyGen |
| **Video postprocessor** | **Rust** | **CPU** | ffmpeg transcode, zero-copy memory, no subprocess |
| **Creative scoring** | **Python** | **CPU (ML)** | scikit-learn/PyTorch — no viable Go ecosystem |

---

## Observability Stack

| Signal | Tool | Source |
|---|---|---|
| Traces | Grafana Tempo (via OTel Collector) | All services (Go + Rust + Python + Next.js) |
| LLM traces | Langfuse | brief-svc (all `generateObject` calls) |
| Metrics | Prometheus → Grafana | All services (NATS lag, fal concurrency, p95 latency) |
| Logs | Better Stack | Railway log drain (structured JSON, always include job_id + org_id) |
| Errors | Sentry | All 4+ services |
| Product events | PostHog | Next.js client |
| Workflow visibility | Temporal Web UI | Temporal server |
| Message bus | NATS Surveyor | NATS cluster |

**Instrumentation libraries:**
- Go: `go.opentelemetry.io/otel`
- Rust: `opentelemetry-rust` + `tracing-opentelemetry`
- Python: `opentelemetry-sdk`
- LLM: `openllmetry` (all services making LLM calls)

---

## V1 → V2 Upgrade Path

| Component | V1 (Current) | V2 (Signal Loop) |
|---|---|---|
| Job orchestration | Temporal (video pipeline) | Temporal + NATS (signal ingestion) |
| Performance signals | Empty `video_performance_events` table | Live ingestion from Meta/TikTok/Google |
| Brief generation | Static LLM prompts | Performance-weighted few-shot injection |
| Creative scoring | None | scoring-svc (Python, rule-based → ML) |
| Model routing | Single call per job | Multi-model routing by cost/quality SLA |
| Avatar | HeyGen v3 only | HeyGen v3 + Tavus v2 (active) |
| Search | None | pgvector semantic brief + asset search |
| Multi-tenancy | Shared schema + RLS | Schema-per-org for Agency+ tier |

---

## Estimated Infrastructure Cost (V1 · ~500 active users)

| Service | Est. /mo |
|---|---|
| Vercel Pro | $20 |
| Supabase Pro | $25 |
| Railway (api + 6 services + Temporal + NATS) | $120 |
| Upstash Redis (rate-limit + cache only) | $10 |
| Cloudflare R2 + CDN | $15 |
| Mux | $30–50 |
| Sentry + Better Stack + Grafana | $30 |
| PostHog | $0–20 |
| **Total infra (excl. AI APIs)** | **~$250–290/mo** |

> fal.ai video generation remains the dominant variable cost. Starter tier profitability (GEN-14) requires credit metering from day one via Stripe Meter API.

---

*Architecture Stack v2.0 — Qvora*
*April 16, 2026 — Confidential*
