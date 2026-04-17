# Qvora

> **Born to Convert.**

**Paste a URL. Get 10 video ads. Know which ones win — automatically.**

Qvora is an AI-powered performance creative platform for agencies. It turns any product URL into a full batch of short-form video ads (9:16), then uses real performance signals to generate better-performing variants in the next round. No briefs to write. No editors to brief. No guesswork.

---

## What It Does

```
Product URL  →  AI Brief  →  Video Variants  →  Performance Data  →  Better Variants
```

1. **Scrape** — Playwright extracts product copy, visuals, and USPs from any URL
2. **Brief** — GPT-4o generates 3–5 creative angles, each with hooks and scripts
3. **Generate** — FAL.AI (Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2) produces video variants; ElevenLabs adds voiceover; HeyGen v3 handles avatar lip-sync
4. **Post-process** — Rust + ffmpeg-sys handles watermarking, captions, reframing, and transcode
5. **Deliver** — Mux HLS streaming with signed playback; Cloudflare R2 for zero-egress asset storage
6. **Learn** *(V2)* — Meta & TikTok signal ingestion feeds back into the next generation cycle

---

## Who It's For (V1)

| Role | Access |
|---|---|
| Agency Media Buyer | Full generation + performance dashboard |
| Agency Creative Director | Full generation + brief editing |
| Agency Account Manager | Read-only reviewer |

> DTC Brand Manager is Phase 2 — not built in V1.

---

## Pricing

| Tier | Price | Variants per Angle |
|---|---|---|
| Starter | $99 / mo | Up to 3 |
| Growth | $149 / mo | Up to 10 |
| Agency | $399 / mo | Unlimited |

7-day free trial · No credit card required · Day 8 locks generation · 30-day data retention on cancel

---

## Build Status (April 2026)

| Phase | Name | Status |
|---|---|---|
| 0 | Foundation & Infrastructure | ✅ Complete |
| 1 | Core Data Layer | ✅ Complete |
| 2 | URL Ingestion & Brief Engine | ✅ Complete |
| 3 | Video Generation Pipeline | ✅ Complete |
| 4 | Brand Kit & Export | ⚠️ Partial |
| 5 | Asset Library & Team | ❓ Pending validation |
| 6 | Platform, Billing & Trial | ❓ Pending validation |
| 7 | Polish, Observability & Launch | ❓ Pending validation |
| 8 | Microservices Domain Extraction | ✅ Complete |

Full detail: [Implementation checklist](docs/07-implementation/IMPLEMENTATION_CHECKLIST.md)

---

## Architecture

### Services

| Service | Language | Responsibility |
|---|---|---|
| `apps/web` | Next.js 15 + tRPC | Frontend + BFF |
| `services/api` | Go Echo v4 | REST API, business logic domain packages |
| `services/worker` | Go + asynq v0.26 | Async job processing (video, TTS, lip-sync) |
| `services/postprocess` | Rust Axum + ffmpeg-sys | CPU-bound video post-processing only |

### API Domain Layer (Phase 8)

All business logic lives in typed domain packages — handlers are thin single-line delegates:

```
services/api/internal/
  domain/
    identity/    ←  workspace, subscription, trial, seats
    brief/       ←  scrape ingestion, angle/hook generation, inline edits
    asset/       ←  asset library CRUD, tagging, search
    media/       ←  Mux lifecycle, R2 uploads, webhooks
    signal/      ←  ad platform connections, metrics, fatigue, recommendations, oauth
  handler/       ←  Echo route handlers (delegates only)
  middleware/    ←  Clerk JWT, rate-limit, tier enforcement
  store/         ←  pgxpool + sqlc abstraction
  util/          ←  UUID, timestamp, Postgres helpers
```

### Key Infrastructure

| Concern | Provider | Notes |
|---|---|---|
| Auth + multi-tenancy | Clerk Organizations | One org = one workspace |
| Database | Supabase Postgres + RLS | sqlc codegen, pgx/v5 driver |
| Cache + rate-limiting | Upstash Redis (HTTP) | **Never** used for queues |
| Job queues | Railway Redis (TCP) | asynq — BLPOP requires persistent TCP |
| Object storage | Cloudflare R2 | Zero egress; presigned PUT uploads |
| Video streaming | Mux | HLS + signed playback + analytics |
| AI models | Vercel AI SDK v6 | `generateObject` (GPT-4o), `streamText` (Claude) |
| Video generation | FAL.AI | `fal.queue.submit()` async — never `fal.subscribe()` |
| TTS | ElevenLabs | `eleven_v3` (quality) / `eleven_flash_v2_5` (speed) |
| Avatar lip-sync | HeyGen **v3** | V2V only; v4 does not exist |
| Payments | Stripe | Subscriptions + Entitlements API |
| Scraping | Modal + Playwright | Serverless, pay-per-second |
| Secrets | Doppler | dev / stg / prd environments |
| Observability | Sentry + Better Stack + PostHog + Langfuse | |
| Hosting | Vercel (web) · Railway (Go + Rust) | Turborepo CI/CD via GitHub Actions |

### Architecture Rules (Non-negotiable)

1. **Two Redis instances** — Upstash HTTP = cache only. Railway TCP = asynq only. Never swap.
2. **SSE is not tRPC** — Generation progress uses a standalone Next.js Route Handler.
3. **Tailwind v4 = CSS-only** — All tokens in `@theme {}` in `globals.css`. No `tailwind.config.ts`.
4. **HeyGen = v3** — `developers.heygen.com`. Any "v4" reference in docs is wrong.
5. **FAL.AI = async queue** — Always `fal.queue.submit()`, never `fal.subscribe()`.
6. **Go = I/O bound. Rust = CPU bound.** Rust scope is limited to the video postprocessor.
7. **Agency-first V1** — DTC features are Phase 2.

---

## Monorepo Layout

```
src/
  apps/
    web/              → Next.js 15 App Router + tRPC BFF
  packages/
    ui/               → shadcn/ui component library
    types/            → Shared TypeScript types
    config/           → Shared ESLint / TS / Tailwind configs
  services/
    api/              → Go Echo v4 API (domain layer architecture)
    worker/           → Go asynq workers (VideoProvider interface)
    postprocess/      → Rust Axum + ffmpeg-sys
docs/
  01-brand/           → Brand identity
  02-product/         → Product definition + overview
  03-market/          → Competitive analysis
  04-specs/           → Feature spec + user stories
  05-design/          → Design system + UI spec + wireframes
  06-technical/       → Architecture + API design + schema
  07-implementation/  → Phases + checklist + references
wiki/                 → Living architecture wiki (canonical source)
supabase/migrations/  → Database migrations (source of truth)
```

---

## Quick Start

```bash
# Install all workspace dependencies
npm install

# Start all services in development
npm run dev

# Type-check
npm run typecheck

# Lint
npm run lint

# Build all packages
npm run build
```

Go services:
```bash
cd src/services/api && go build ./...
cd src/services/worker && go build ./...
```

---

## Documentation

| Document | Location |
|---|---|
| Product definition | [docs/02-product/Qvora_Product-Definition.md](docs/02-product/Qvora_Product-Definition.md) |
| Feature specification | [docs/04-specs/Qvora_Feature-Spec.md](docs/04-specs/Qvora_Feature-Spec.md) |
| User stories | [docs/04-specs/Qvora_User-Stories.md](docs/04-specs/Qvora_User-Stories.md) |
| Architecture stack | [docs/06-technical/Qvora_Architecture-Stack.md](docs/06-technical/Qvora_Architecture-Stack.md) |
| System architecture | [docs/06-technical/Qvora_System-Architecture.md](docs/06-technical/Qvora_System-Architecture.md) |
| Database schema | [docs/06-technical/Qvora_Database-Schema.md](docs/06-technical/Qvora_Database-Schema.md) |
| Implementation phases | [docs/07-implementation/Qvora_Implementation-Phases.md](docs/07-implementation/Qvora_Implementation-Phases.md) |
| Implementation checklist | [docs/07-implementation/IMPLEMENTATION_CHECKLIST.md](docs/07-implementation/IMPLEMENTATION_CHECKLIST.md) |
| Implementation references | [docs/07-implementation/Qvora_Implementation-References.md](docs/07-implementation/Qvora_Implementation-References.md) |
| Living wiki | [wiki/](wiki/index.md) |
| Copilot instructions | [.github/copilot-instructions.md](.github/copilot-instructions.md) |

---

## Brand

| Token | Value | Use |
|---|---|---|
| Volt | `#7B2FFF` | Primary CTAs, brand accent |
| Convert Green | `#00E87A` | Success states, performance wins |
| Signal Red | `#FF3D3D` | Fatigue alerts, warnings |
| Data Blue | `#2E9CFF` | Analytics, charts |
| Qvora Black | `#0A0A0F` | Primary background |
| Qvora White | `#F5F5F7` | Text, surfaces |

**Fonts:** Clash Display (hero) · Space Grotesk Bold (display) · Inter (UI) · JetBrains Mono (data)

---

## Repository

This repository is private and internal. All planning, architecture, and implementation details are confidential. Do not publish externally without explicit approval.

**License:** Proprietary. All rights reserved.

- Badge test: Pair flow (2026-04-17)
