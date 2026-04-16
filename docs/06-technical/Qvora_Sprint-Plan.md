# Qvora Sprint Plan

**Product:** Qvora — AI Ad Creative Agent  
**Planning Horizon:** Sprint 0–4 (11 weeks to MVP)  
**Methodology:** 2-week sprints, async-first team  
**Definition of Done:** Feature merged to `main`, passing tests, deployed to staging, acceptance criteria verified

> **Status:** Legacy V1 sprint artifact. Canonical execution roadmap is maintained in `../07-implementation/Qvora_Implementation-Phases.md` and `../07-implementation/IMPLEMENTATION_CHECKLIST.md`.

---

## Table of Contents

1. [MVP Definition](#mvp-definition)
2. [Build Order Rationale](#build-order-rationale)
3. [Sprint 0 — Foundation (Weeks 1–2)](#sprint-0--foundation-weeks-12)
4. [Sprint 1 — Brief Pipeline (Weeks 3–4)](#sprint-1--brief-pipeline-weeks-34)
5. [Sprint 2 — Video Generation (Weeks 5–6)](#sprint-2--video-generation-weeks-56)
6. [Sprint 3 — Workspace & Export (Weeks 7–8)](#sprint-3--workspace--export-weeks-78)
7. [Sprint 4 — Polish & Launch Prep (Weeks 9–10)](#sprint-4--polish--launch-prep-weeks-910)
8. [Post-MVP: V2 Backlog](#post-mvp-v2-backlog)
9. [Risks & Mitigations](#risks--mitigations)
10. [Team & Capacity](#team--capacity)

---

## MVP Definition

**MVP ships when a media buyer can:**
1. Paste a product URL
2. Receive a structured creative brief with 3 ad angles in <30s
3. Generate one 15s vertical video from a brief angle in <3 min
4. Download a structured ZIP export with the video + brief PDF
5. Do all of this within a multi-tenant org workspace with Clerk auth

**Out of scope for MVP (V2):**
- Performance learning loop (fatigue detection, A/B data)
- Ad platform integrations (Meta/TikTok API)
- Image→Video mode (V2V / Voice→Video is V1)
- White-label / custom domain
- Advanced analytics dashboard

---

## Build Order Rationale

Dependencies flow in one direction:

```
Auth + DB schema → Brief pipeline → Video generation → Export → Workspace UX
```

- **Auth first**: every API call needs `org_id` from JWT; nothing else can be built without it
- **Schema first**: Brief pipeline writes to `briefs` via Go API — the table must exist
- **Brief before video**: video generation reads `brief.content.angles` to build the prompt
- **Video before export**: export packages need at least one ready asset
- **Export before launch**: without export, users have no delivery mechanism

The SSE streaming layer is built alongside video generation in Sprint 2 — it's the UX glue that makes async jobs feel real-time.

---

## Sprint 0 — Foundation (Weeks 1–2)

**Goal:** Repo wired, auth working, DB live, dev environment reproducible.

### Stories

| ID | Story | Points | Owner |
|---|---|---|---|
| S0-01 | Monorepo setup: `/src/apps/web`, `/src/services/api`, `/src/services/worker` with shared root `package.json` | 3 | Backend |
| S0-02 | Clerk org-mode auth: JWT with `org_id`, `org_role`, `plan` claims | 3 | Backend |
| S0-03 | Go Echo server scaffolding: health check, CORS, Clerk JWT middleware | 2 | Backend |
| S0-04 | Supabase project: apply full schema DDL, RLS policies, seed data | 3 | Backend |
| S0-05 | Cloudflare R2 bucket setup: `assets`, `exports`, `thumbnails` with CORS config | 2 | Infra |
| S0-06 | Next.js 15 app scaffold: App Router, Tailwind, shadcn/ui, Clerk provider | 2 | Frontend |
| S0-07 | tRPC BFF setup: server + client, Clerk session → JWT forwarding | 3 | Frontend |
| S0-08 | CI pipeline: lint, typecheck, Go test, next build on PR | 3 | DevOps |
| S0-09 | Docker Compose: local dev stack (Postgres, Redis, Go API, Python agent) | 2 | DevOps |
| S0-10 | Environment variable management: `.env.example`, secret naming conventions | 1 | DevOps |
| S0-11 | Vercel project + staging environment wired to main branch | 1 | DevOps |

**Total: 25 points**

### Acceptance Criteria

- [ ] `GET /health` returns `200` from Go server
- [ ] Clerk sign-in creates user + org, JWT contains `org_id`
- [ ] All 12 DB tables created with RLS active (verified via Supabase dashboard)
- [ ] R2 bucket accessible with signed URL generation
- [ ] `npm run dev` starts all services from Docker Compose in <2 min
- [ ] PR to `main` triggers CI and fails on lint error

### Exit Gate

Dev environment setup time for new engineer: <20 minutes.

---

## Sprint 1 — Brief Pipeline (Weeks 3–4)

**Goal:** A user can paste a URL and receive a structured creative brief.

### Stories

| ID | Story | Points | Owner |
|---|---|---|---|
| S1-01 | `POST /briefs` endpoint: URL validation, job creation, asynq task enqueue | 3 | Backend |
| S1-02 | Brief pipeline: Vercel AI SDK generateObject chain (scrape→parse→angles→hooks→validate) | 4 | AI |
| S1-03 | URL scraper: Playwright headless + fallback to `readability-js` for JS-heavy pages | 3 | AI |
| S1-04 | `GET /briefs` + `GET /briefs/{id}` endpoints with RLS-scoped queries | 2 | Backend |
| S1-05 | `PATCH /briefs/{id}`: inline angle editing, optimistic update in tRPC | 2 | Backend |
| S1-06 | SSE endpoint: `GET /stream/briefs/{id}` with Go `http.Flusher` | 3 | Backend |
| S1-07 | Brief creation UI: URL input screen (S-02 wireframe), loading state with SSE progress | 3 | Frontend |
| S1-08 | Brief detail view: angle cards with edit-in-place (C-08 component) | 3 | Frontend |
| S1-09 | Projects CRUD: `POST /projects`, list sidebar, project switcher | 2 | Backend + Frontend |
| S1-10 | Brief list view: dashboard state (S-01 wireframe), empty state | 2 | Frontend |
| S1-11 | Error handling: scrape failed, generation failed UI states | 2 | Frontend |
| S1-12 | Brief retry logic: Zod validation failure → retry generateObject call, max 2 retries | 1 | AI |

**Total: 32 points**

### Acceptance Criteria

- [ ] Paste `https://allbirds.com/products/mens-tree-runners` → brief arrives in <30s with 3 angles
- [ ] SSE stream shows 5 distinct progress stages in the UI
- [ ] Angle fields (headline, hook, body, CTA) are editable inline and saved on blur
- [ ] Brief list shows correct data for org; user B cannot see org A's briefs
- [ ] Failed brief (scrape error) shows actionable error state, not blank screen
- [ ] Brief pipeline produces valid structured output for 95% of product pages in test set (20 URLs)

### Exit Gate

An agency user can generate 5 briefs in one session without hitting an error.

---

## Sprint 2 — Video Generation (Weeks 5–6)

**Goal:** A user can generate a 15s vertical video from a brief angle.

### Stories

| ID | Story | Points | Owner |
|---|---|---|---|
| S2-01 | FAL.AI client: async submit → `task_id` → webhook callback pattern (Go) | 3 | Backend |
| S2-02 | `POST /assets/generate` endpoint: T2V mode, FAL.AI routing logic (auto-select model) | 3 | Backend |
| S2-03 | asynq worker: process `video_generation` tasks, FAL.AI submit, error handling | 3 | Backend |
| S2-04 | FAL.AI webhook: `POST /internal/webhooks/falai`, HMAC verify, R2 upload, thumbnail | 4 | Backend |
| S2-05 | Mux video ingestion: upload R2-stored video to Mux for HLS streaming | 2 | Backend |
| S2-06 | SSE endpoint: `GET /stream/assets/{id}` with stage-by-stage progress | 3 | Backend |
| S2-07 | Prompt builder: brief angle → FAL.AI prompt template (Go service) | 2 | Backend |
| S2-08 | `GET /assets` + `GET /assets/{id}` endpoints | 2 | Backend |
| S2-09 | Generation settings UI (S-06 wireframe): T2V mode selector, style, format, duration | 3 | Frontend |
| S2-10 | Generation progress card (C-19): SSE-driven progress bar with stage labels | 3 | Frontend |
| S2-11 | Video player: Mux HLS player, thumbnail fallback, aspect-ratio wrapper | 2 | Frontend |
| S2-12 | Asset card (C-12): play, download, retry actions | 2 | Frontend |
| S2-13 | Asset list view (S-07 wireframe): grid layout, filter by format/status | 2 | Frontend |
| S2-14 | Quota enforcement: check `org_subscriptions.ads_used < ads_limit` before generation | 2 | Backend |
| S2-15 | `POST /assets/{id}/retry` endpoint for failed jobs | 1 | Backend |

**Total: 37 points**

### Acceptance Criteria

- [ ] Click "Generate" on an angle → video appears in asset library in <3 min
- [ ] SSE stream updates in real-time: queued → submitted → rendering → complete
- [ ] Video plays in-app via Mux HLS player at correct aspect ratio
- [ ] Starter plan user blocked after 20 ads/month with upgrade prompt
- [ ] FAL.AI webhook verified with HMAC; reject unverified requests with `403`
- [ ] Asset `file_name` follows convention: `{brand_slug}_{angle_type}_{format}_{hook_variant}_{YYYYMMDD}_v{n}.mp4`
- [ ] Failed generation shows retry button; retry re-queues successfully

### Exit Gate

Generate 3 different-format videos from one brief with no manual intervention.

---

## Sprint 3 — Workspace & Export (Weeks 7–8)

**Goal:** Users can manage assets, organize by project, and download structured exports.

### Stories

| ID | Story | Points | Owner |
|---|---|---|---|
| S3-01 | `POST /exports` endpoint: ZIP packaging with ffmpeg thumbnail extraction, brief PDF | 4 | Backend |
| S3-02 | Export worker: asynq task, JSZip in Go, R2 upload, signed download URL (1h TTL) | 3 | Backend |
| S3-03 | `GET /exports/{id}` + `GET /exports` endpoints | 2 | Backend |
| S3-04 | Export creation UI (S-09 wireframe): asset multi-select, format checklist, naming | 3 | Frontend |
| S3-05 | Export detail view: manifest table, download button, expiry countdown | 2 | Frontend |
| S3-06 | Exports list view | 1 | Frontend |
| S3-07 | Brands CRUD: auto-create on brief generation, manual edit, brand profile card | 3 | Backend + Frontend |
| S3-08 | Project sidebar: project switcher with brief/asset counts, color picker | 2 | Frontend |
| S3-09 | Command palette (⌘K): brief search, quick actions (new brief, export) | 3 | Frontend |
| S3-10 | Org member management: invite flow, member list, role badge (admin-only) | 3 | Backend + Frontend |
| S3-11 | Settings screen (S-11 wireframe): org profile, billing portal link, usage bar | 2 | Frontend |
| S3-12 | Brief PDF generation: Go html→PDF via `chromedp`, brand-styled layout | 3 | Backend |
| S3-13 | `DELETE` endpoints: briefs, assets, exports, projects (soft delete) | 2 | Backend |
| S3-14 | Keyboard shortcuts: ⌘N (new brief), ⌘E (export), ⌘/ (sidebar) | 1 | Frontend |

**Total: 34 points**

### Acceptance Criteria

- [ ] Select 4 assets → Create export → download ZIP with videos + brief PDF in <45s
- [ ] Export ZIP filename structure matches: `qvora_export_{brand}_{date}.zip`
- [ ] Brief PDF is human-readable with correct brand name, angles, and hooks
- [ ] Org admin can invite member via email; invited user joins correct org
- [ ] Command palette (⌘K) surfaces briefs and actions in <100ms
- [ ] Soft-deleted briefs not visible in list; assets still accessible via direct link for 7 days

### Exit Gate

A complete "URL → Brief → Video → Export" workflow completed end-to-end with no backend errors.

---

## Sprint 4 — Polish & Launch Prep (Weeks 9–10)

**Goal:** Production-ready, observable, with billing live and launch checklist complete.

### Stories

| ID | Story | Points | Owner |
|---|---|---|---|
| S4-01 | Stripe billing integration: plan-gated features, usage reporting, portal redirect | 4 | Backend |
| S4-02 | `GET /org` usage endpoint: ads_used, briefs, storage, seats | 2 | Backend |
| S4-03 | Usage bar in sidebar + settings (ads remaining, upgrade CTA at 80%) | 2 | Frontend |
| S4-04 | Onboarding flow: sign-up → org create → first brief wizard (S-12 wireframe) | 3 | Frontend |
| S4-05 | Empty states: dashboard, asset library, exports (actionable, not blank) | 2 | Frontend |
| S4-06 | Error boundary: global React error boundary with "Report bug" link | 1 | Frontend |
| S4-07 | Posthog analytics: page views, brief_created, asset_generated, export_downloaded | 2 | Frontend |
| S4-08 | Sentry: error tracking in both Go and Next.js | 2 | DevOps |
| S4-09 | Structured logging: Go `zerolog` with `request_id`, `org_id`, `user_id` on all logs | 2 | Backend |
| S4-10 | Load test: k6 script for brief generation under 50 concurrent users | 3 | QA |
| S4-11 | Security audit: OWASP top 10 checklist, RLS penetration test, auth bypass check | 3 | Security |
| S4-12 | Mobile responsive: all screens responsive at 375px (iOS SE breakpoint) | 2 | Frontend |
| S4-13 | `robots.txt`, `og:image`, meta tags, favicon | 1 | Frontend |
| S4-14 | Production deploy: Vercel (web), Railway (Go API), Fly.io (Python agent), Supabase prod | 3 | DevOps |
| S4-15 | Launch checklist: DNS, SSL, Stripe live keys, Clerk prod, FAL.AI prod key | 1 | DevOps |

**Total: 33 points**

### Acceptance Criteria

- [ ] Stripe checkout creates subscription; plan reflected in JWT `plan` claim within 60s
- [ ] Growth plan user blocked from Agency features (5+ angle count) with correct error
- [ ] Posthog dashboard shows funnel: sign_up → brief_created → asset_generated → exported
- [ ] 95th percentile brief generation time <35s under 50 concurrent users (k6 test)
- [ ] RLS test: authenticated user for Org A receives `403` on any Org B resource
- [ ] All pages pass Lighthouse accessibility score ≥ 90
- [ ] Production deployment smoke test: full workflow end-to-end in prod

### Exit Gate

Founder can demo the full workflow live to 5 beta users without hitting a blocker.

---

## Sprint Summary

| Sprint | Focus | Stories | Points | Weeks |
|---|---|---|---|---|
| Sprint 0 | Foundation | 11 | 25 | 1–2 |
| Sprint 1 | Brief Pipeline | 12 | 32 | 3–4 |
| Sprint 2 | Video Generation | 15 | 37 | 5–6 |
| Sprint 3 | Workspace & Export | 14 | 34 | 7–8 |
| Sprint 4 | Polish & Launch | 15 | 33 | 9–10 |
| **Total** | | **67** | **161** | **10 weeks** |

---

## Post-MVP: V2 Backlog

V2 work not in MVP scope, ordered by customer value:

| Priority | Feature | Why Deferred |
|---|---|---|
| P1 | Image→Video (i2v) mode | FAL.AI integration pattern same as T2V; deferred to reduce Sprint 2 scope |
| P1 | Voice→Video (v2v) mode | HeyGen Avatar v3 integration adds complexity; worth a dedicated sprint |
| P1 | Performance learning loop | Requires ad_accounts + asset_metrics data; needs V1 to collect baseline data first |
| P2 | Meta Ads / TikTok Ads API integration | Requires OAuth flows + partner API access approval |
| P2 | Ad fatigue detection | Needs 60+ days of performance data per account |
| P2 | Multi-language support (UI) | i18n infrastructure adds setup cost; English-first is fine for initial ICP |
| P3 | White-label / custom domain | Agency upsell; requires multi-domain Clerk config |
| P3 | Advanced analytics dashboard | Agency plan retention driver |
| P3 | AI-generated thumbnails | Nice-to-have for social preview; not blocking export |

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| FAL.AI API downtime during demo | Medium | High | Implement mock mode with pre-recorded video response for demos |
| Brief quality too low for agency use | Medium | High | Build eval harness in Sprint 1 (20-URL test set), iterate on Langfuse prompt versions before Sprint 2 |
| Playwright scraper blocked by bot detection | High | Medium | Fallback chain: Playwright → Jina Reader API → raw fetch → graceful error |
| Stripe webhook delays causing plan mismatch | Low | Medium | Optimistic plan grant on checkout session, confirm on webhook arrival |
| Supabase RLS misconfiguration | Low | Critical | Automated RLS penetration test in Sprint 4: Org A JWT + Org B resource ID → must return `403` |
| HeyGen Avatar v3 not available for V1 | Medium | Low | V2V is post-MVP; no impact on Sprint 0–4 |
| Sprint 2 overload (15 stories, 37 points) | Medium | Medium | Split: backend (S2-01→S2-08) done by W5 day 4; frontend (S2-09→S2-13) can start W5 day 3 with mock API |

---

## Team & Capacity

**Assumed team for 10-week sprint:**

| Role | Capacity | Sprints |
|---|---|---|
| Full-stack engineer (Backend-heavy) | 100% | Sprint 0–4 |
| Full-stack engineer (Frontend-heavy) | 100% | Sprint 0–4 |
| AI/ML engineer | 100% | Sprint 1–2 (brief + video pipeline) |
| DevOps / Infra | 50% | Sprint 0, 4 |

**Solo founder mode:** If building solo, sequence: Backend tracks → AI tracks → Frontend tracks within each sprint. Sprint 2 is the highest-risk sprint solo (37 points); plan to drop S2-11 (Mux player) and use basic `<video>` tag for MVP if behind.

---

## Definition of Done (Per Story)

A story is done when:
1. Code merged to `main` via PR (no open review comments)
2. Unit tests written and passing (min 70% coverage for new files)
3. Deployed to staging environment
4. Acceptance criteria in this doc manually verified on staging
5. No new Sentry errors introduced (from Sprint 4 onward)
6. Type-safety maintained: no `any` types added, tRPC procedures fully typed
