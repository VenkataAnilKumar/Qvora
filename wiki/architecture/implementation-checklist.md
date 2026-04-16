---
title: Implementation Checklist
category: architecture
tags: [implementation, checklist, phases, progress, status]
sources: [IMPLEMENTATION_CHECKLIST, Qvora_Implementation-Phases]
updated: 2026-04-16
---

# Implementation Checklist

## TL;DR
8 phases, 11 weeks. Phases 0–3 complete. Phase 4 partial. Phases 5–7 not validated. Key confirmed constraints: `plan_tier` is `starter/growth/agency` (no 'scale'), `jobs.status` CHECK does not include 'briefing'.

---

## Phase Progress

| Phase | Name | Duration | Status |
|---|---|---|---|
| Phase 0 | Foundation & Infrastructure | Week 1 | ✅ Complete |
| Phase 1 | Core Data Layer | Week 2 | ✅ Complete |
| Phase 2 | URL Ingestion & Brief Engine | Weeks 3–4 | ✅ Complete |
| Phase 3 | Video Generation Pipeline | Weeks 5–7 | ✅ Complete |
| Phase 4 | Brand Kit & Export | Week 8 | ⚠️ Partial |
| Phase 5 | Asset Library & Team | Week 9 | ❓ Not Validated |
| Phase 6 | Platform, Billing & Trial | Week 10 | ❓ Not Validated |
| Phase 7 | Polish, Observability & Launch | Week 11 | ❓ Not Validated |

**Complete: 4/8 · Partial: 1/8 · Not Started: 3/8** *(as of Apr 16, 2026)*

---

## Phase 0 — Foundation & Infrastructure ✅

### Confirmed complete:
- Turborepo monorepo with `src/apps/`, `src/packages/`, `src/services/`
- pnpm workspaces
- `biome.json` (replaces ESLint + Prettier)
- `lefthook.yml` (replaces Husky)
- `.env.example` with all required vars
- GitHub Actions: CI + path-filtered deploy workflows (web, api, worker, postprocess, db)
- `CODEOWNERS`, PR template, branch protection on `main`, Dependabot
- Vercel + Railway + Supabase + Upstash Redis + Cloudflare R2 + Mux + Doppler + Clerk provisioned
- Docker Compose: Postgres + Redis ×2 + all services
- `src/apps/web` — Next.js 15, App Router, Tailwind v4, `@theme {}`
- `src/packages/ui` (shadcn/ui), `src/packages/types`, `src/packages/config`
- Go modules, Rust project, Axum health route

**Gate:** `turbo dev` starts without errors; CI passes; all Railway services on `/health`.

---

## Phase 1 — Core Data Layer ✅

### Key constraints confirmed in this phase:
- Migrations in `supabase/migrations/` ONLY (not `services/api/db/`)
- Tables: workspaces, users, brands, briefs, brief_angles, brief_hooks, jobs, variants, asset_tags, exports
- `jobs.status` CHECK: `queued, scraping, generating, postprocessing, complete, failed` — **no 'briefing'**
- `plan_tier` CHECK: `starter, growth, agency` — **no 'scale'**
- RLS policies on all tables (workspace isolation)
- Additional migrations: `002_postprocess_callbacks.sql`, `003_mux_webhook_events_and_reconcile.sql`
- sqlc queries generated, pgx/v5 driver
- Clerk JWT middleware in Go: validates `org_id` + `plan_tier` on every request
- Upstash Redis: rate-limit middleware (HTTP, not TCP)
- R2 presigned URL generation working

---

## Phase 2 — URL Ingestion & Brief Engine ✅

### Confirmed complete:
- Modal Playwright scraping (pay-per-second serverless)
- GPT-4o `generateObject` → Zod-validated structured product JSON
- Claude Sonnet 4.6 angles + hooks generation
- Brief stored to DB with fail-fast on angle/hook insert errors (**BRIEF-08 resolved**)
- `ai/prompts/angles-gen.prompt.ts` for prompt management
- tRPC: `briefs.create`, `briefs.byId`, `briefs.list`
- Frontend: URL input → brief preview with angles + hooks + inline edit

> **Note (BRIEF-09):** Per-angle/per-hook regenerate was referenced as "in progress" in Apr 15 snapshot. Check current state.

---

## Phase 3 — Video Generation Pipeline ✅

### Confirmed complete:
- asynq worker: `generation:video` task → Railway Redis TCP
- FAL.AI: `fal.queue.submit()` (always async, never `fal.subscribe()`)
- FAL.AI webhook callback handler at `/v1/webhooks/fal`
- ElevenLabs TTS (`eleven_v3` quality / `eleven_flash_v2_5` fast)
- Rust postprocessor: watermark + captions + transcode + reframe (CPU-bound)
- R2 upload + Mux ingest → signed playback
- SSE Route Handler: `/api/generation/[jobId]/stream/route.ts`
- Frontend: generation settings page + SSE progress cards + Mux player

---

## Phase 4 — Brand Kit & Export ⚠️ Partial

### Items likely complete:
- Brand kit CRUD (Go API + Supabase)
- Logo upload to R2 with presigned PUT
- Brand switcher in topbar

### Items potentially incomplete (validate):
- Export ZIP bundle assembly (generation:export asynq task)
- Export download with pre-signed URL TTL
- Brief PDF generation in ZIP export

---

## Phase 5 — Asset Library & Team ❓

*Not yet validated. Planned:*
- Asset library with filter (by brief, status, platform)
- Asset tagging
- Team invite + role assignment (admin/creator/reviewer)
- Seat limit enforcement per tier

---

## Phase 6 — Platform, Billing & Trial ❓

*Not yet validated. Planned:*
- Stripe webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- Tier limit middleware (variant count, brand count) in Go
- Trial countdown UI: day 6 banner + day 8 generation lock
- Conversion email triggers via Stripe events

---

## Phase 7 — Polish, Observability & Launch ❓

*Not yet validated. Planned:*
- Sentry integration (frontend + Go + Rust)
- Better Stack uptime + log drain
- PostHog event tracking (activation, generation, export)
- Langfuse prompt versioning + cost attribution per org
- Load testing: 50 concurrent generation jobs
- End-to-end QA run

---

## Open Questions
- [ ] BRIEF-09 (per-angle regenerate) — completed or still in progress?
- [ ] Phase 4 export ZIP — is `generation:export` worker task implemented?
- [ ] HeyGen V2V lip-sync — which phase? Not explicitly listed in checklist.

## Related Pages
- [[sprint-plan]] — planned sprint structure
- [[roadmap]] — phase status from product perspective
- [[system-architecture]] — services involved in each phase
