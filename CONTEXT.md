# Qvora — Project Context

> Load this file first when starting any AI-assisted work on Qvora.

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
    ↓  (Playwright scrape → structured extraction)
Creative Strategy Brief
    ↓  (GPT-4o structured output → 3–5 creative angles)
Video Ad Variants
    ↓  (FAL.AI T2V/I2V + ElevenLabs TTS + HeyGen lip-sync)
Export (9:16, platform-ready)
    ↓  [V2] Ad platform connector
Performance Signal
    ↓  [V2] Learning loop → better next variants
```

---

## ICP

| Priority | Persona | Context |
|---|---|---|
| **V1 Primary** | Agency Media Buyer | $50K–$500K/mo ad spend, 3–15 brand clients |
| **V1 Primary** | Agency Creative Director | Owns brand voice + creative strategy quality |
| **V1 Reviewer** | Agency Account Manager | Read-only. No generation stories. |
| **Phase 2 only** | DTC Brand Manager | Not built in V1. Growth tier. |

---

## Pricing

| Tier | Price | Variants/Angle |
|---|---|---|
| Starter | $99/mo | Max 3 |
| Growth | $149/mo | Max 10 |
| Agency | $399/mo | Unlimited |

- 7-day free trial, no credit card
- Day 8: generation locked, 30-day data retention
- Conversion emails: Day 3 / Day 6 / Day 8

---

## Tech Stack (Quick Reference)

| Layer | Technology |
|---|---|
| Frontend | Next.js 15 App Router + shadcn/ui + Tailwind v4 |
| BFF | tRPC (TypeScript ≥5.7.2) |
| Generation Stream | SSE Route Handler `/api/generation/[jobId]/stream` |
| API | Go Echo v4 (Railway) |
| Job Workers | Go asynq v0.26.0 (Railway) |
| Video Post-Processing | Rust Axum + ffmpeg-sys (Railway) |
| AI Orchestration | Vercel AI SDK v6 |
| Text-to-Video | FAL.AI (Veo 3.1 / Kling 3.0 / Runway Gen-4.5 / Sora 2) |
| TTS | ElevenLabs (`eleven_v3` / `eleven_flash_v2_5`) |
| Avatar Lip-Sync | HeyGen Avatar API **v3** (`developers.heygen.com`) |
| LLM Strategy | GPT-4o (structured JSON) + Claude Sonnet 4.6 (regen) |
| LLM Observability | Langfuse |
| Auth | Clerk (Organizations = workspaces) |
| Primary DB | PostgreSQL via Supabase + RLS |
| SQL Codegen | sqlc |
| Cache/Rate-limit Redis | Upstash (HTTP/REST) |
| Job Queue Redis | Railway Redis (TCP — required for asynq) |
| Object Storage | Cloudflare R2 (zero egress) |
| Video Hosting | Mux (HLS + analytics) |
| Payments | Stripe Subscriptions + Entitlements API |
| Secrets | Doppler |
| Monorepo | Turborepo |
| CI/CD | GitHub Actions |
| Observability | Sentry + Better Stack + PostHog |

---

## Non-Negotiable Architecture Rules

1. **Two Redis instances are mandatory and non-interchangeable:**
   - `Upstash` (HTTP/REST) → cache, rate-limiting, session store only
   - `Railway Redis` (TCP) → asynq job queues only (BLPOP requires persistent TCP)

2. **SSE is not tRPC.** Generation progress uses a standalone Next.js Route Handler — not a tRPC subscription.

3. **Tailwind v4 is CSS-only.** All tokens live in `@theme {}` in `globals.css`. There is no `tailwind.config.ts`.

4. **HeyGen is v3.** Active platform: `developers.heygen.com`. V2V lip-sync is v3-only. Any reference to "v4" in docs is incorrect.

5. **Agency = V1 ICP.** DTC (P4) is Phase 2 only. All V1 features, flows, and limits are agency-first.

6. **Go = I/O-bound; Rust = CPU-bound.** Go for API + workers (latency owned by FAL/OpenAI/HeyGen). Rust only for video postprocessor (watermark, caption burn, transcode, reframe).

---

## Monorepo Structure

```
apps/
  web/          → Next.js 15 (frontend + tRPC BFF)
packages/
  ui/           → shadcn/ui component library
  types/        → Shared TypeScript types
  config/       → ESLint, TS, Tailwind shared configs
services/
  api/          → Go Echo v4 REST API
  worker/       → Go asynq workers
  postprocess/  → Rust Axum + ffmpeg-sys
```

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

---

## Docs

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
| `docs/06-technical/Qvora_Architecture-Stack.md` | Full stack decisions + architecture notes |
| `docs/06-technical/Qvora_System-Architecture.md` | Component diagram, sequence flows, code patterns |
| `docs/06-technical/Qvora_Database-Schema.md` | Full PostgreSQL DDL, RLS policies, 12 tables |
| `docs/06-technical/Qvora_API-Design.md` | REST endpoints, request/response schemas, tRPC procedures |
| `docs/06-technical/Qvora_Sprint-Plan.md` | Sprint 0–4 story map, acceptance criteria, build order |
| `docs/07-implementation/Qvora_Implementation-References.md` | SDK docs, install commands, code patterns |
