---
title: Architecture & Tech Stack Overview
category: architecture
tags: [stack, non-negotiables, go, rust, next.js, redis, railway, vercel]
sources: [Qvora_Architecture-Stack, copilot-instructions]
updated: 2026-04-15
---

# Architecture & Tech Stack Overview

## TL;DR
Three workload profiles: interactive (Next.js/tRPC), async compute (Go + asynq), CPU-bound video (Rust + ffmpeg). Seven non-negotiable rules. Two Redis instances that must never be swapped.

---

## Workload Profiles

| Profile | Concern | Stack Component |
|---|---|---|
| Interactive | Brief generation, edits, real-time UI | Next.js 15 + tRPC + Go API |
| Async compute | Video generation, rendering, export | Go asynq + FAL.AI + ElevenLabs + HeyGen |
| CPU-bound | Watermark, captions, transcode, reframe | Rust Axum + ffmpeg-sys |

---

## Stack by Layer

### Frontend (Vercel)

| Concern | Choice |
|---|---|
| Framework | Next.js 15 App Router only (no Pages Router) |
| UI | shadcn/ui + Radix + Tailwind CSS v4 |
| **Tailwind v4 rule** | CSS-only config. All tokens in `@theme {}` in `globals.css`. **No `tailwind.config.ts`.** |
| State | Zustand + TanStack Query v5 |
| Forms | React Hook Form + Zod |
| Animation | Framer Motion |
| Video | Mux Player (`<MuxPlayer>`) |
| Monorepo | Turborepo |

### BFF

| Concern | Choice |
|---|---|
| Type-safe API | tRPC (TypeScript ≥5.7.2): `initTRPC.create()`, `httpBatchLink` |
| **Generation progress stream** | **Standalone Route Handler** at `/api/generation/[jobId]/stream/route.ts` — **NOT a tRPC subscription** |

### Backend (Railway)

| Concern | Choice |
|---|---|
| API server | Go Echo v4 |
| Job workers | Go + asynq v0.26.0 (Redis-backed) |
| Video postprocessor | Rust Axum + ffmpeg-sys (CPU-bound only: watermark, captions, transcode, reframe) |

### Auth

| Concern | Choice |
|---|---|
| Provider | Clerk (`@clerk/nextjs`, `clerkMiddleware()`, `ClerkProvider`) |
| Multi-tenant | One Clerk Organization = one Qvora workspace |
| JWT | Carries `org_id` (workspace) + `org_role`; verified by Go API middleware |

---

## Hosting

| Service | Component |
|---|---|
| Vercel | Next.js web app + tRPC BFF |
| Railway | Go API, Go workers, Railway Redis, Rust postprocessor |
| Modal | Playwright scraping (serverless, pay-per-second) |
| Doppler | Secrets management (dev/stg/prd environments) |
| GitHub Actions + Turborepo | CI/CD |

---

## Observability

| Tool | Purpose |
|---|---|
| Sentry | Error tracking |
| Better Stack | Logs and uptime |
| PostHog | Product analytics |
| Langfuse | LLM observability, prompt versioning, cost attribution per workspace |

---

## The 7 Non-Negotiable Rules

> These rules are architectural constraints, not preferences. Violating them breaks the system.

1. **Two Redis** — Upstash = HTTP cache only. Railway = asynq TCP only. Never swap. BLPOP requires persistent TCP; Upstash HTTP does not support it.
2. **SSE is not tRPC** — Generation progress uses a standalone Route Handler. Never implement as a tRPC subscription.
3. **Tailwind v4 = CSS-only** — No `tailwind.config.ts`. All tokens in `@theme {}`.
4. **HeyGen = v3** — `developers.heygen.com`. V2V lip-sync is v3-only. "v4" does not exist.
5. **Agency = V1 ICP** — DTC (P4) is Phase 2. Never build V1 features for DTC.
6. **Go = I/O bound; Rust = CPU bound** — Do not expand Rust beyond the video postprocessor.
7. **FAL.AI = async queue** — Always `fal.queue.submit()`, never `fal.subscribe()` (blocks).

---

## Monorepo Structure

```
src/
  apps/
    web/          → Next.js 15 App Router (frontend + tRPC BFF)
  packages/
    ui/           → shadcn/ui components
    types/        → Shared TypeScript types
    config/       → Shared ESLint / TS / Tailwind configs
  services/
    api/          → Go Echo v4 API
    worker/       → Go asynq workers
    postprocess/  → Rust Axum + ffmpeg-sys video processor
```

---

## Open Questions
- [ ] Is the Rust postprocessor deployed as a sidecar or a separate Railway service?
- [ ] What is the asynq queue retry policy for failed FAL.AI jobs?

## Related Pages
- [[ai-layer]] — AI model choices and API integrations
- [[data-layer]] — databases, Redis, storage
- [[architecture-decisions]] — why these choices were made
- [[roadmap]] — implementation status by phase
