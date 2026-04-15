# Qvora — Agent Instructions

> This file is read by agentic AI tools (Claude Code, OpenAI Codex, Gemini CLI, etc.).
> Load this before taking any action in this repository.

---

## Project Overview

**Qvora** is an AI-powered performance creative SaaS.
**Workflow:** Product URL → AI creative brief → short-form video ads → performance signal → better next variants.
**Tagline:** *"Born to Convert."*
**Status:** Pre-build. All planning docs complete. Implementation not started.

---

## Primary Reference Files

Read these files before working on any module:

| File | When to Read |
|---|---|
| `.github/CONTEXT.md` | Always — product, stack, rules quick-reference |
| `.github/MEMORY.md` | Always — locked decisions, doc fix history, research data |
| `docs/06-technical/Qvora_Architecture-Stack.md` | Before any backend, infra, or API work |
| `docs/04-specs/Qvora_Feature-Spec.md` | Before implementing any feature module |
| `docs/04-specs/Qvora_User-Stories.md` | Before implementing any user-facing flow |
| `docs/05-design/Qvora_Design-System.md` | Before writing any UI component or style |
| `docs/07-implementation/Qvora_Implementation-References.md` | Before integrating any third-party SDK |

---

## ICP (Who You Are Building For)

- **V1:** Agency Media Buyers + Agency Creative Directors only
- **Account Managers (P3):** Reviewer role — no generation features
- **DTC Brand Managers (P4):** Phase 2 only — build nothing for DTC in V1

---

## Architecture Rules (Hard Constraints)

### Redis — Two Instances, Never Substitutable

| Variable | Provider | Protocol | Use |
|---|---|---|---|
| `UPSTASH_REDIS_REST_URL` | Upstash | HTTP/REST | Cache, rate-limit, session |
| `RAILWAY_REDIS_URL` | Railway | TCP | asynq job queues |

`asynq` uses `BLPOP`/`BRPOP` (blocking pop) which requires a persistent TCP connection. Upstash HTTP proxy does not support this. Substituting Upstash for the job queue silently fails in production.

### SSE Generation Stream

- **Path:** `src/apps/web/app/api/generation/[jobId]/stream/route.ts`
- **Transport:** `text/event-stream` via `ReadableStream`
- **Rule:** This is a standalone Next.js Route Handler — NOT a tRPC subscription.

### Tailwind v4

- **No `tailwind.config.ts`** — the file must not exist.
- All tokens defined in `@theme {}` block inside `src/apps/web/app/globals.css`.

### HeyGen Version

- Active platform: **v3** at `developers.heygen.com`
- V2V lip-sync is v3-only.
- Any reference to "HeyGen v4" in this codebase is incorrect — treat as v3.
- v2 supported until Oct 31, 2026; migrate to v3 before production.

### FAL.AI Async

```typescript
// Always async queue — never blocking subscribe
const { request_id } = await fal.queue.submit("fal-ai/veo3", { input });
// Store request_id in asynq task; poll from Go worker
```

### Language Boundaries

| Layer | Language | Reason |
|---|---|---|
| API server | Go (Echo v4) | I/O-bound; latency from FAL/OpenAI/HeyGen |
| Job workers | Go (asynq) | Goroutine-per-job model |
| URL orchestrator | Go | Calls Modal Playwright webhooks |
| Video postprocessor | Rust (Axum + ffmpeg-sys) | CPU-bound (watermark, transcode, reframe) |

Do not expand Rust beyond the postprocessor service.

---

## Pricing Tiers (Enforce in Go API Middleware)

| Tier | Monthly | Variant Limit |
|---|---|---|
| Starter | $99 | 3 per angle |
| Growth | $149 | 10 per angle |
| Agency | $399 | Unlimited |

Tier limits are enforced server-side in Go API middleware — never trust the client.

```go
maxVariants := map[string]int{"starter": 3, "growth": 10, "agency": -1}[workspace.PlanTier]
if maxVariants != -1 && count >= maxVariants {
    return c.JSON(402, map[string]string{"error": "variant_limit_exceeded"})
}
```

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

## Feature Module Scope (V1 vs V2)

| Module | ID | V1 | V2 |
|---|---|---|---|
| URL Ingestion & Extraction | EXT | ✅ | |
| Creative Strategy Engine | BRIEF | ✅ | |
| Video Generation (T2V / I2V / V2V) | GEN | ✅ | |
| Voiceover & Caption Engine | VOICE | ✅ | |
| Brand Kit System | BRAND | ✅ | |
| Export & Naming | EXPORT | ✅ | |
| Asset Library | LIB | ✅ | |
| Team & Collaboration | TEAM | ✅ | |
| Ad Account Connector | CONN | | ✅ |
| Performance Learning Engine (Signal) | SIGNAL | | ✅ |
| Platform & Billing | PLAT | ✅ | |

---

## Key Package Installs

```bash
# Frontend
npm install ai @fal-ai/client elevenlabs @clerk/nextjs
npm install @trpc/server @trpc/client @trpc/react-query
npm install @upstash/redis @upstash/ratelimit
npm install @mux/mux-node @mux/mux-player-react stripe
npm install zustand @tanstack/react-query framer-motion
npm install react-hook-form zod posthog-js @sentry/nextjs

# Backend (Go)
go get github.com/labstack/echo/v4
go get github.com/hibiken/asynq
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

---

## Do Not

- Do not create `tailwind.config.ts` or `tailwind.config.js`
- Do not use tRPC for the SSE generation stream
- Do not use Upstash Redis for asynq job queues
- Do not build any feature tagged "Phase 2" or "V2" in the current sprint
- Do not reference HeyGen API v4 — use v3 (`developers.heygen.com`)
- Do not use `fal.subscribe()` for video generation (blocking — breaks under load)
- Do not add Rust code outside `src/services/postprocess/`
- Do not enforce tier limits in client-side code
