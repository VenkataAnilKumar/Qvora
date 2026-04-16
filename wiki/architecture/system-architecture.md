---
title: System Architecture
category: architecture
tags: [system-architecture, component-diagram, sequences, workload-profiles, async-jobs]
sources: [Qvora_System-Architecture]
updated: 2026-04-15
---

# System Architecture

## TL;DR
Three workload profiles: interactive (< 2s), async compute (1–5 min GPU), scheduled (hourly/daily). Five layers: Client → BFF → Core API → Job Queue → AI Orchestration → Data. SSE is the real-time glue between async jobs and the UI.

---

## Workload Profiles

| Profile | Latency | Compute | Examples |
|---|---|---|---|
| **Interactive** | < 2s | CPU / LLM API | Brief generation, UI updates, auth |
| **Async compute** | 1–5 min | GPU (via FAL.AI) | Video generation, rendering, export |
| **Scheduled** | Hourly/daily | CPU | Signal ingestion, fatigue detection (V2) |

---

## Layer Map

```
CLIENT LAYER
  Next.js 15 (Vercel)
  App Router · shadcn/ui · TanStack Query · Framer Motion
  tRPC client · SSE listener (generation progress)
         │
         │ tRPC (internal) / REST (public)
         ▼
BFF LAYER
  Next.js API Routes + tRPC Router (Vercel)
  Auth: Clerk JWT validation · Rate limit: Upstash Redis
         │                                          │
         │ REST (Go HTTP)                            │ SSE stream /api/generation/[jobId]/stream
         ▼                                          │
CORE API                                           │
  Go · Echo v4 (Railway)               ◄───────────┘
  /briefs /generations /assets /exports /brands /team /signal(V2)
  sqlc → PostgreSQL (Supabase)
         │
         │ Enqueue job (asynq → Railway Redis TCP)
         ▼
JOB QUEUE
  asynq + Railway Redis (TCP)
  Tasks: brief:extract · generation:video · generation:export · signal:sync(V2)
         │
         ▼
AI ORCHESTRATION LAYER
  ┌─ BRIEF ENGINE (Vercel AI SDK / TypeScript) ─────────────────┐
  │  1. URL Extract → Modal Playwright                          │
  │  2. Parse → GPT-4o (JSON strict / generateObject)          │
  │  3. Angles → Claude Sonnet 4.6                             │
  │  4. Hooks → Claude Sonnet 4.6                              │
  │  5. Format recommendation                                  │
  │  6. Zod validate + retry                                   │
  └─────────────────────────────────────────────────────────────┘
  ┌─ VIDEO PIPELINE (Go workers, Railway) ──────────────────────┐
  │  FAL.AI: Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2    │
  │  HeyGen Avatar API v3 (V2V lip-sync)                       │
  │  ElevenLabs API (TTS / voice clone)                        │
  │  Rust Axum + ffmpeg-sys (watermark, captions, transcode)   │
  └─────────────────────────────────────────────────────────────┘
         │
         ▼
DATA LAYER
  PostgreSQL (Supabase) · Upstash Redis (cache) · R2 (assets) · Mux (HLS)
```

---

## Key Sequence: Brief Generation (URL → Structured Brief)

```
User → BFF (tRPC createBrief) → Go API (INSERT brief status=pending)
  → Enqueue brief:extract (asynq)
  → Worker → Modal Playwright (scrape URL)
  → Worker → GPT-4o (parse product data → structured JSON)
  → Worker → Claude Sonnet 4.6 (generate 3–5 creative angles)
  → Worker → Claude Sonnet 4.6 (generate 3 hooks per angle)
  → Worker → UPDATE brief status=ready + INSERT angles + hooks
  → Go API SSE or polling → BFF → UI renders brief
```

- parse = `generateObject` (GPT-4o, JSON strict, Zod schema)
- angles/hooks = `streamText` (Claude Sonnet 4.6)
- Entire flow: target < 15 seconds

---

## Key Sequence: Video Generation (Angle → Video)

```
User submits generation → BFF (tRPC submit) → Go API (INSERT job status=queued)
  → Enqueue generation:video (asynq)
  → Worker → ElevenLabs (TTS if voiceover)
  → Worker → HeyGen v3 (avatar + lip-sync if V2V)
  → Worker → FAL.AI fal.queue.submit() (T2V async)
  → FAL.AI callback → POST /v1/webhooks/fal
  → Go API → UPDATE job status=postprocessing
  → Enqueue generation:postprocess
  → Worker → Rust Axum (watermark, captions, transcode, reframe)
  → Worker → Upload to R2
  → Worker → Upload to Mux → get playback_id
  → Worker → UPDATE job status=complete + INSERT asset
  → SSE stream pushes complete event to browser
```

> ⚠️ **FAL.AI rule:** Always `fal.queue.submit()`. Never `fal.subscribe()` — it blocks the worker thread.

---

## Key Sequence: SSE Stream

```
Browser opens EventSource to /api/generation/[jobId]/stream
  → Route Handler pulls job status from Upstash Redis (fast read)
  → Pushes SSE events as status transitions occur
  → Events: queued → scraping → generating → postprocessing → complete | failed
```

- **Not tRPC** — standalone Next.js Route Handler with `ReadableStream`
- Uses Upstash Redis for job status cache (HTTP, not TCP)
- Railway Redis (TCP) is for asynq job queues only

---

## Infrastructure Map

| Service | Host | Notes |
|---|---|---|
| Next.js (frontend + BFF) | Vercel | App Router, tRPC, SSE Route Handler |
| Go API | Railway | Echo v4, REST endpoints, JWT middleware |
| Go Workers | Railway | asynq consumers |
| Rust postprocessor | Railway | ffmpeg-sys, CPU-bound only |
| Railway Redis | Railway | TCP, asynq queues ONLY |
| PostgreSQL | Supabase | RLS enabled, sqlc codegen |
| Upstash Redis | Upstash | HTTP, cache + rate-limit ONLY |
| R2 (assets) | Cloudflare | Zero egress, presigned PUT uploads |
| Mux | Mux | HLS + signed playback |
| Playwright scraping | Modal | Serverless, pay-per-second |
| Secrets | Doppler | dev/stg/prd environments |

---

## Critical Non-Negotiables

1. **Two Redis** — Upstash HTTP (cache only) ≠ Railway TCP (asynq only). Never swap.
2. **SSE ≠ tRPC** — generation stream is a standalone Route Handler.
3. **FAL.AI = async queue** — always `fal.queue.submit()`, never `fal.subscribe()`.
4. **Go = I/O bound; Rust = CPU bound** — do not expand Rust beyond postprocessor.

---

## Open Questions
- [ ] Is there a queue priority system? (e.g., Agency tier jobs processed first)
- [ ] What is the retry strategy for failed FAL.AI jobs?
- [ ] How are partial job failures handled (e.g., 3 of 5 videos complete)?

## Related Pages
- [[stack-overview]] — full tech stack reference
- [[api-design]] — endpoint details for this architecture
- [[data-layer]] — data storage in this system
- [[ai-layer]] — AI models and SDK usage
