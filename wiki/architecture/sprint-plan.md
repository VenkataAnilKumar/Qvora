---
title: Sprint Plan
category: architecture
tags: [sprint-plan, mvp, milestones, stories, build-order]
sources: [Qvora_Sprint-Plan]
updated: 2026-04-15
---

# Sprint Plan

## TL;DR
5 sprints (Sprint 0–4), 11 weeks to MVP. Build order: Auth+DB → Brief pipeline → Video generation → Export → Workspace UX. SSE stream built in Sprint 2 alongside video generation.

---

## MVP Definition

A media buyer can:
1. Paste a product URL
2. Receive a structured creative brief with 3 ad angles in < 30s
3. Generate one 15s vertical video from a brief angle in < 3 min
4. Download a structured ZIP export (video + brief PDF)
5. Do all of this within a multi-tenant org workspace with Clerk auth

**Out of scope for MVP → V2:** Signal/performance loop, Meta/TikTok API integrations, Image-to-Video mode, white-label, advanced analytics.

---

## Build Order Rationale

```
Auth + DB schema → Brief pipeline → Video generation → Export → Workspace UX
```

- Auth first: `org_id` in every JWT; nothing else can be built without it
- Schema first: Brief pipeline writes to `briefs` via Go API; table must exist
- Brief before video: video reads `brief.content.angles` to build the prompt
- Video before export: export needs at least one ready asset
- SSE alongside video: async jobs are meaningless without real-time UI

---

## Sprint 0 — Foundation (Weeks 1–2)

**Goal:** Repo wired, auth working, DB live, dev environment reproducible.

| ID | Story | Points |
|---|---|---|
| S0-01 | Monorepo: web, api, worker, shared package.json | 3 |
| S0-02 | Clerk org-mode auth: JWT with `org_id`, `org_role`, `plan` | 3 |
| S0-03 | Go Echo server: health check, CORS, Clerk JWT middleware | 2 |
| S0-04 | Supabase: apply full schema DDL, RLS, seed data | 3 |
| S0-05 | Cloudflare R2 buckets: assets, exports, thumbnails | 2 |
| S0-06 | Next.js 15 scaffold: App Router, Tailwind v4, shadcn/ui, Clerk | 2 |
| S0-07 | tRPC BFF: server + client, Clerk session → JWT forwarding | 3 |
| S0-08 | CI: lint, typecheck, Go test, next build on PR | 3 |
| S0-09 | Docker Compose: Postgres + Redis + Go API + all services | 2 |
| S0-10 | Env var management: `.env.example`, naming conventions | 1 |
| S0-11 | Vercel project + staging env on main | 1 |

**Exit gate:** New engineer dev setup < 20 minutes.

---

## Sprint 1 — Brief Pipeline (Weeks 3–4)

**Goal:** URL in → structured creative brief out.

Key deliverables:
- Modal Playwright scraping (URL → product data)
- GPT-4o parse → Zod-validated structured product JSON
- Claude Sonnet 4.6 → 3–5 creative angles + 3 hooks/angle
- Brief persist to DB (angles + hooks in `brief_angles` + `brief_hooks`)
- BFF tRPC: `briefs.create`, `briefs.byId`
- UI: URL input → brief display with angles + hooks

---

## Sprint 2 — Video Generation (Weeks 5–6)

**Goal:** Brief angle → rendered video, with real-time progress in UI.

Key deliverables:
- asynq worker: `generation:video` task
- FAL.AI integration: `fal.queue.submit()`, webhook callback handler
- ElevenLabs TTS integration
- Rust postprocessor: watermark + captions + transcode
- Mux upload + signed playback
- **SSE Route Handler:** `/api/generation/[jobId]/stream` (real-time progress)
- UI: Generation settings page, SSE progress cards, video player (Mux)

---

## Sprint 3 — Workspace & Export (Weeks 7–8)

**Goal:** Multi-tenant workspace + organized exports.

Key deliverables:
- Brand kit CRUD (brands table, logo upload to R2)
- Brand switcher in topbar
- Export job: ZIP bundle (videos + brief PDF) → R2
- Export download with pre-signed URL
- Asset library: filterable list with Mux thumbnails
- Team invite + role assignment (admin/creator/reviewer)
- Onboarding wizard (org setup + brand setup screens)

---

## Sprint 4 — Polish & Launch Prep (Weeks 9–10)

**Goal:** Production-ready, trials working, billing gated.

Key deliverables:
- Stripe subscription integration: webhooks (`invoice.paid`, `subscription.updated`, `subscription.deleted`)
- Tier limit enforcement: variant count, brand count (Go middleware)
- Trial countdown UI (day 6, day 8 banners)
- Day 8 generation lock
- Conversion email triggers (Day 3 / Day 6 / Day 8)
- Error states + empty states across all screens
- Sentry, Better Stack, PostHog instrumented
- Langfuse prompt versioning + cost attribution active
- Load testing: 50 concurrent generation jobs

---

## Post-MVP V2 Backlog

- EPIC 7: Qvora Signal (ad account connect, performance data, fatigue detection)
- I2V mode (Image-to-Video)
- pgvector semantic similarity for creative reuse
- Temporal for scheduled Signal sync
- White-label / custom domain
- Advanced analytics dashboard

---

## Open Questions
- [ ] HeyGen V2V (avatar lip-sync) — which sprint? Sprint 2 scope seems tight.
- [ ] Playwright scraping on Modal — is the Modal service built in Sprint 1 or Sprint 0?
- [ ] Point estimates — team size and velocity?

## Related Pages
- [[implementation-checklist]] — actual completion status vs. plan
- [[system-architecture]] — services being built sprint by sprint
- [[features]] — feature module → sprint mapping
