# Qvora — GitHub Copilot Instructions

You are working on **Qvora**, an AI-powered performance creative SaaS.
Read this file fully before answering any question or making any change.

---

## Product

- **What it is:** Converts a product URL into short-form video ads (9:16), then learns from ad performance data to generate better variants.
- **Tagline:** *"Born to Convert."*
- **One-liner:** *"Paste a URL. Get 10 video ads. Know which ones win — automatically."*
- **Core flow:** URL → Playwright scrape → GPT-4o creative brief (3–5 angles) → FAL.AI video generation + ElevenLabs TTS + HeyGen lip-sync → Export → [V2] Performance Signal loop

---

## ICP

| Priority | Persona | Notes |
|---|---|---|
| V1 Primary | Agency Media Buyer | $50K–$500K/mo ad spend, 3–15 brand clients |
| V1 Primary | Agency Creative Director | Owns brand voice + creative strategy |
| V1 Reviewer only | Agency Account Manager | No generation features — read-only |
| **Phase 2 only** | DTC Brand Manager | **Not built in V1** |

---

## Pricing Tiers

| Tier | Price | Variant Limit |
|---|---|---|
| Starter | $99/mo | Max 3 variants/angle |
| Growth | $149/mo | Max 10 variants/angle |
| Agency | $399/mo | Unlimited |

- 7-day free trial, no credit card required
- Day 8: generation locked; 30-day data retention
- Conversion emails: Day 3 / Day 6 / Day 8

---

## Tech Stack

### Frontend — Vercel
- **Framework:** Next.js 15 App Router only (no Pages Router)
- **UI:** shadcn/ui + Radix + Tailwind CSS v4
- **Tailwind v4 rule:** CSS-only config. All tokens in `@theme {}` in `globals.css`. **No `tailwind.config.ts`.**
- **State:** Zustand + TanStack Query v5
- **Forms:** React Hook Form + Zod
- **Animation:** Framer Motion
- **Video:** Mux Player (`<MuxPlayer>`)
- **Monorepo:** Turborepo

### BFF
- **tRPC** (TypeScript ≥5.7.2): `initTRPC.create()`, `httpBatchLink`
- **SSE generation stream:** Standalone Route Handler at `src/apps/web/app/api/generation/[jobId]/stream/route.ts` — **NOT a tRPC subscription**

### Backend — Railway
- **API:** Go Echo v4
- **Job workers:** Go + asynq v0.26.0 (Redis-backed)
- **Video postprocessor:** Rust Axum + ffmpeg-sys (CPU-bound only; watermark, captions, transcode, reframe)

### AI Layer
- **Vercel AI SDK v6** (`ai` package): `generateObject` (GPT-4o structured), `streamText` (Claude regen), `useObject` hook (streaming UI)
- **FAL.AI:** T2V/I2V via `fal.queue.submit()` async — never `fal.subscribe()` (blocks)
  - Models: Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2
- **ElevenLabs:** `eleven_v3` (quality) / `eleven_flash_v2_5` (~75ms latency)
- **HeyGen:** Avatar API **v3** — active platform at `developers.heygen.com`. V2V lip-sync is v3-only. **"v4" does not exist.**
- **Langfuse:** LLM observability, prompt versioning, cost attribution per workspace

### Auth
- **Clerk:** `@clerk/nextjs`, `clerkMiddleware()`, `ClerkProvider`
- **Multi-tenant:** One Clerk Organization = one Qvora workspace
- JWT carries `org_id` (workspace) + `org_role`; verified by Go API middleware

### Data
- **PostgreSQL** via Supabase with Row-Level Security
  - Always: `(SELECT auth.uid())` wrapper, `TO authenticated`, index on `user_id`
- **sqlc:** Write SQL → `sqlc generate` → type-safe Go (`sqlc.yaml`, pgx/v5 driver)
- **Redis — TWO MANDATORY INSTANCES (never substitute one for the other):**
  - `Upstash` (HTTP/REST) → cache, rate-limiting, session store
  - `Railway Redis` (TCP) → asynq job queues (`BLPOP` requires persistent TCP — Upstash HTTP does not work here)
- **Cloudflare R2:** Object storage (zero egress). Use presigned PUT URLs for direct uploads.
- **Mux:** HLS video streaming + analytics. Signed playback for workspace-scoped access.

### Payments
- **Stripe** subscriptions + Entitlements API
- Statuses to handle: `trialing` → `active` → `past_due` → `canceled`
- Key webhooks: `invoice.paid`, `customer.subscription.updated`, `customer.subscription.deleted`
- Tier limits enforced in Go API middleware (server-side), not client-side

### Infrastructure
- **Vercel** — src/apps/web (Next.js + tRPC)
- **Railway** — Go API, Go workers, Railway Redis, Rust postprocessor
- **Modal** — Playwright scraping (serverless, pay-per-second)
- **Doppler** — secrets management (dev/stg/prd environments)
- **GitHub Actions + Turborepo** — CI/CD
- **Sentry, Better Stack, PostHog** — error tracking, logs, product analytics

---

## Non-Negotiable Rules

1. **Two Redis rule:** Upstash = HTTP cache only. Railway = asynq TCP only. Never swap.
2. **SSE is not tRPC.** Generation progress uses a standalone Route Handler.
3. **Tailwind v4 = CSS-only.** No `tailwind.config.ts`. All tokens in `@theme {}`.
4. **HeyGen = v3.** Any reference to "v4" in docs is an error.
5. **Agency = V1 ICP.** DTC (P4) is Phase 2. Never build V1 features for DTC.
6. **Go = I/O bound; Rust = CPU bound.** Do not expand Rust beyond the video postprocessor.
7. **FAL.AI = async queue.** Always `fal.queue.submit()`, never `fal.subscribe()`.

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
    postprocess/  → Rust Axum + ffmpeg-sys
```

---

## Brand Tokens

```css
--color-volt: #7B2FFF;          /* Primary brand, CTAs */
--color-convert-green: #00E87A; /* Success, performance wins */
--color-signal-red: #FF3D3D;    /* Fatigue alerts, warnings */
--color-data-blue: #2E9CFF;     /* Analytics, charts */
--color-qvora-black: #0A0A0F;   /* Primary background */
--color-qvora-white: #F5F5F7;   /* Text, surfaces */
```

**Fonts:** Clash Display (hero) · Space Grotesk Bold (display) · Inter (UI) · JetBrains Mono (data)

---

## Key Docs

| File | Contents |
|---|---|
| `.github/CONTEXT.md` | Quick-reference product + stack summary |
| `.github/MEMORY.md` | Decision log, all doc fixes applied, research numbers |
| `docs/02-product/Qvora_Product-Definition.md` | Full product definition, personas, KPIs |
| `docs/04-specs/Qvora_Feature-Spec.md` | Functional requirements + acceptance criteria |
| `docs/04-specs/Qvora_User-Stories.md` | All epics + user stories |
| `docs/06-technical/Qvora_Architecture-Stack.md` | Full stack + architecture decisions |
| `docs/07-implementation/Qvora_Implementation-References.md` | SDK docs, install commands, code patterns |
