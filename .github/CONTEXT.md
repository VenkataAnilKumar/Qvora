# Qvora — Project Context

> Load this file first when starting any AI-assisted work on Qvora.
> Architecture decisions are now in `docs/06-technical/Qvora_Microservice-Architecture.md`.

---

## What Is Qvora?

**Qvora** is an AI-powered performance creative SaaS that converts a product URL into multiple high-converting short-form video ads, then continuously learns from real ad performance data to generate better variants over time.

- **Tagline:** *"Born to Convert."*
- **One-liner:** *"Paste a URL. Get 10 video ads. Know which ones win — automatically."*
- **Category:** AI creative intelligence + production system for paid social advertising

---

## Core Workflow

```
Product URL
    ↓  (Playwright scrape via Modal → structured ProductData)
Creative Strategy Brief
    ↓  (GPT-4o structured extraction → Claude Sonnet 4.6 angles + hooks)
Video Ad Variants
    ↓  (FAL.AI T2V/I2V + ElevenLabs TTS + HeyGen v3 lip-sync)
    ↓  (Rust ffmpeg: reframe, watermark, captions, transcode)
Export (9:16, platform-ready, Mux HLS delivery)
    ↓  [V2] Ad platform connector (Meta / TikTok / Google)
Performance Signal
    ↓  [V2] scoring-svc prediction + feedback loop → better next briefs
```

---

## ICP

| Priority | Persona | Context |
|---|---|---|
| **V1 Primary** | Agency Media Buyer | $50K–$500K/mo ad spend, 3–15 brand clients |
| **V1 Primary** | Agency Creative Director | Owns brand voice + creative strategy quality |
| **V1 Reviewer** | Agency Account Manager | Read-only. No generation. |
| **Phase 2 only** | DTC Brand Manager | Not built in V1. |

---

## Pricing

| Tier | Price | Variants/Angle |
|---|---|---|
| Starter | $99/mo | Max 3 |
| Growth | $149/mo | Max 10 |
| Agency | $399/mo | Unlimited |

- 14-day free trial, no credit card
- Day 15: generation locked, 30-day data retention
- Conversion emails: Day 5 / Day 10 / Day 15

---

## Tech Stack (Quick Reference)

| Layer | Technology |
|---|---|
| Frontend | Next.js 15 App Router + shadcn/ui + Tailwind v4 |
| BFF | tRPC v11 (TypeScript ≥ 5.7.2) |
| Real-time | **Supabase Realtime** (Postgres Changes → browser WebSocket) |
| API Gateway | Go Echo v4 (Railway) |
| Message Bus | **NATS JetStream** (replaces asynq + Railway Redis for queuing) |
| Workflow Engine | **Temporal.io** (video generation pipeline — durable execution) |
| Domain Services | Go (identity, ingestion, brief, asset, signal) |
| Video Post-Processing | Rust Axum + ffmpeg-next (Railway) |
| Internal Comms | **gRPC** (hot paths: Go → Rust, Go → Go) |
| AI Orchestration | Vercel AI SDK v6 |
| Text-to-Video | FAL.AI — `fal.queue.submit()` only (Veo 3.1 / Kling 3.0 / Runway Gen-4.5) |
| TTS | ElevenLabs (`eleven_v3` / `eleven_flash_v2_5`) |
| Avatar Lip-Sync | HeyGen Avatar API **v3** (default) + Tavus v2 (secondary) |
| LLM Strategy | GPT-4o (structured JSON) + Claude Sonnet 4.6 (creative) |
| LLM Observability | Langfuse + OpenLLMetry |
| Auth | Clerk (Organizations = workspaces) |
| Primary DB | PostgreSQL 16 via Supabase + RLS + Realtime |
| SQL Codegen | sqlc |
| Cache / Rate-limit | Upstash Redis (HTTP/REST only — no queuing) |
| Object Storage | Cloudflare R2 (zero egress) |
| Video Hosting | Mux (HLS + analytics) |
| Payments | Stripe Subscriptions + Meter API |
| Secrets | Doppler |
| Monorepo | Turborepo |
| CI/CD | GitHub Actions |
| Observability | Sentry + Better Stack + PostHog + Grafana Tempo (OTel) |
| Creative Scoring | Python FastAPI (V2 — scikit-learn → PyTorch) |

---

## Architecture Rules

### Rules unchanged from V1

1. **Tailwind v4 is CSS-only.** All tokens live in `@theme {}` in `globals.css`. There is no `tailwind.config.ts`.

2. **HeyGen is v3.** Active platform: `developers.heygen.com`. V2V lip-sync is v3-only. Any reference to "v4" is incorrect. HeyGen v3 is the default provider; Tavus v2 is the secondary.

3. **Agency = V1 ICP.** DTC is Phase 2 only. All V1 features, flows, and limits are agency-first.

4. **Go = I/O-bound; Rust = CPU-bound; Python = ML-only.** Go for API and workers. Rust only for video postprocessor. Python only for scoring-svc. No language introduced outside these domains.

5. **`fal.queue.submit()` only — never `fal.subscribe()`.** Non-blocking; `fal.subscribe()` blocks worker threads under load.

6. **Clerk Organizations = workspaces.** One Clerk Org = one Qvora workspace. Multi-tenant isolation via Clerk JWT `org_id` claim + Supabase RLS.

7. **Migrations in `supabase/migrations/` only.** `services/api/db/` holds sqlc query definitions only — never migrations.

8. **ffmpeg-next bindings only — never `Command::new("ffmpeg")`.** No subprocess; zero-copy memory management.

9. **Mux tokens scoped to workspace.** HS256 JWT, 1-hour expiry, `sub` = workspaceID.

---

### Rules unlocked / changed from V1

10. **NATS JetStream is the message bus — not Railway Redis + asynq.**
    - NATS for all async messaging (`ingestion.*`, `brief.*`, `media.*`, `asset.*`, `signal.*`)
    - asynq is decommissioned for pipeline work
    - Railway Redis TCP instance decommissioned once Phase 8 migration is complete
    - Upstash Redis (HTTP) is retained for rate-limit counters and URL scrape cache only
    - Complex multi-step workflows → Temporal. Simple background jobs → NATS consumers.

11. **Temporal.io orchestrates the video creation pipeline — not asynq task chains.**
    - `VideoCreationWorkflow` is durable: survives process restarts mid-execution
    - Each fal.ai webhook triggers a Temporal signal (not a Redis key write)
    - Workflow visibility via Temporal Web UI
    - Simple non-pipeline jobs (export assembly, signal sync, GDPR cleanup) remain NATS consumers

12. **Supabase Realtime replaces the custom SSE Route Handler.**
    - Workers update `generation_jobs.status` in Supabase
    - Supabase Realtime WAL change feed pushes to the browser (WebSocket, not polling)
    - The SSE Route Handler (`/api/generation/[jobId]/stream`) is removed in Phase 8
    - RLS on Realtime channels enforces per-org isolation automatically

13. **gRPC for internal hot paths — not HTTP REST.**
    - API Gateway → identity-svc: gRPC (quota check, plan lookup)
    - media-orchestrator → media-postprocessor: gRPC streaming (process + progress)
    - mTLS required on all internal gRPC connections (certs via Doppler)
    - External API calls (fal.ai, HeyGen, ElevenLabs) remain HTTPS REST

14. **Avatar providers use a provider interface — not HeyGen-only.**
    - `AvatarProvider` Go interface: `CreateLipSync()`, `GetStatus()`
    - HeyGen v3 = default (quality)
    - Tavus v2 = secondary (cost-optimized at volume, automatic fallback on HeyGen 429)
    - Provider recorded on `generation_jobs.avatar_provider` for cost attribution

15. **Python is allowed for scoring-svc only.**
    - `src/services/scoring/` — FastAPI + scikit-learn + PyTorch
    - No Python in any other service
    - V1: rule-based heuristics. V2: scikit-learn. V3: PyTorch CNN on video frames

16. **Performance events are append-only — never UPDATE.**
    - `video_performance_events` table: insert-only
    - Scores are derived via `creative_scores` materialized view (refreshed hourly)
    - No in-place score updates

---

## Monorepo Structure

### Current (V1)
```
src/
  apps/
    web/              → Next.js 15 (frontend + tRPC BFF)
  packages/
    ui/               → shadcn/ui component library
    types/            → Shared TypeScript types
    config/           → Biome + TS base configs
  services/
    api/              → Go Echo v4 (API Gateway)
    worker/           → Go asynq workers (monolithic — decomposed in Phase 8)
    postprocess/      → Rust Axum + ffmpeg-next
  ai/
    prompts/          → Shared prompt files (imported via @qvora/prompts/* alias)
```

### Target (Post-Phase 9)
```
src/
  apps/
    web/              → Next.js 15 (frontend + tRPC BFF)
  packages/
    ui/               → shadcn/ui component library
    types/            → Shared TypeScript types
    config/           → Biome + TS base configs
    proto/            → Protobuf definitions (.proto files)
  services/
    api/              → Go Echo v4 (API Gateway)
    identity/         → Go (auth, quota, Stripe metering)
    ingestion/        → Go (Modal/Playwright orchestration)
    brief/            → Go (LLM pipeline — GPT-4o + Claude)
    media-orchestrator/ → Go (Temporal worker, fal.ai, HeyGen, Tavus)
    media-postprocessor/ → Rust (ffmpeg, gRPC server)
    asset/            → Go (metadata, exports, brands, Mux)
    signal/           → Go (ad platform sync, events) — V2
    scoring/          → Python (FastAPI, scikit-learn, PyTorch) — V2
  ai/
    prompts/          → Shared prompt files
```

---

## Implementation Status

| Phase | Name | Status |
|---|---|---|
| 0 | Foundation & Infrastructure | ✅ Complete |
| 1 | Core Data Layer | ✅ Complete |
| 2 | URL Ingestion & Brief Engine | ✅ Complete |
| 3 | Video Generation Pipeline | ✅ Complete |
| 4 | Brand Kit & Export | ⏳ Pending |
| 5 | Asset Library & Team | ⏳ Pending |
| 6 | Platform, Billing & Trial | ⏳ Pending |
| 7 | V1 Polish, Observability & Launch | ⏳ Pending |
| 8 | Microservice Foundation (NATS + Realtime + service split) | ⏳ Post-Launch |
| 9 | Temporal + gRPC + Multi-Provider Avatar | ⏳ Post-Launch |
| 10 | V2 Signal Loop & Intelligence | ⏳ Post-Launch |

---

## Docs Index

| File | Contents |
|---|---|
| `docs/01-brand/Qvora_Brand-Identity.md` | Colors, typography, voice, logo concept |
| `docs/02-product/Qvora_Product-Definition.md` | Full product definition, personas, KPIs, risks |
| `docs/02-product/Qvora_Product-Overview.md` | Executive overview |
| `docs/03-market/Qvora_Competitive-Analysis.md` | Creatify, Arcads, HeyGen, AdCreative.ai |
| `docs/04-specs/Qvora_Feature-Spec.md` | Functional requirements + acceptance criteria |
| `docs/04-specs/Qvora_User-Stories.md` | Epics 1–8, all user stories |
| `docs/04-specs/Qvora_User-Journey.md` | End-to-end user flows |
| `docs/05-design/Qvora_Design-System.md` | Token system, Tailwind v4 theme, components |
| `docs/05-design/Qvora_Wireframes.md` | Screens S-01 through S-15 |
| `docs/05-design/Qvora_UI-Spec.md` | 25 UI components, variants, TSX snippets |
| `docs/06-technical/Qvora_Microservice-Architecture.md` | **Canonical service catalogue** — boundaries, protocols, data ownership |
| `docs/06-technical/Qvora_Architecture-Stack.md` | Stack decisions + language allocation |
| `docs/06-technical/Qvora_System-Architecture.md` | Component diagrams + sequence flows |
| `docs/06-technical/Qvora_Database-Schema.md` | Full PostgreSQL DDL, RLS policies |
| `docs/06-technical/Qvora_API-Design.md` | REST endpoints, request/response schemas, tRPC procedures |
| `docs/06-technical/Qvora_Sprint-Plan.md` | Sprint 0–4 story map, acceptance criteria |
| `docs/07-implementation/Qvora_Implementation-Phases.md` | Phase goals, deliverables, key decisions (Phases 0–10) |
| `docs/07-implementation/IMPLEMENTATION_CHECKLIST.md` | Detailed task checklist with gates (Phases 0–10) |
| `docs/07-implementation/Qvora_Implementation-References.md` | SDK docs, install commands, code patterns |

---

## Brand

| Token | Hex | Role |
|---|---|---|
| Qvora Black | `#0A0A0F` | Primary background |
| Qvora Volt | `#7B2FFF` | Brand color, CTAs |
| Qvora White | `#F5F5F7` | Text, surfaces |
| Convert Green | `#00E87A` | Success, performance wins |
| Signal Red | `#FF3D3D` | Fatigue alerts, warnings |
| Data Blue | `#2E9CFF` | Analytics, charts |

**Fonts:** Clash Display (hero) · Space Grotesk Bold (display) · Inter (UI/body) · JetBrains Mono (data)
