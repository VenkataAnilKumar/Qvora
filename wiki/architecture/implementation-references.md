---
title: Implementation References
category: architecture
tags: [sdk-docs, install-commands, api-references, links, development]
sources: [Qvora_Implementation-References]
updated: 2026-04-15
---

# Implementation References

## TL;DR
Single-lookup doc for every SDK, API, and tool in the Qvora stack. Contains official doc URLs, install commands, Qvora-specific usage notes, and critical gotchas. Check here before building any new module.

---

## Quick Stack Lookup

| Layer | Technology | Host | Doc |
|---|---|---|---|
| Frontend | Next.js 15 App Router | Vercel | https://nextjs.org/docs/app |
| UI Components | shadcn/ui | — | https://ui.shadcn.com/docs |
| Styling | Tailwind CSS v4 | — | https://tailwindcss.com/docs |
| State | Zustand | — | https://zustand.docs.pmnd.rs |
| Server State | TanStack Query v5 | — | https://tanstack.com/query/latest |
| Forms | React Hook Form + Zod | — | https://react-hook-form.com |
| Animation | Framer Motion | — | https://www.framer.com/motion |
| Video | Mux Player | Mux | https://docs.mux.com/guides/mux-player |
| BFF | tRPC | Vercel | https://trpc.io/docs |
| AI SDK | Vercel AI SDK v6 | — | https://sdk.vercel.ai/docs |
| T2V/I2V | FAL.AI | FAL.AI | https://fal.ai/docs |
| Avatar | HeyGen API **v3** | HeyGen | https://developers.heygen.com |
| TTS | ElevenLabs | ElevenLabs | https://elevenlabs.io/docs |
| Scraping | Modal + Playwright | Modal | https://modal.com/docs |
| LLM Obs | Langfuse | Cloud/Self | https://langfuse.com/docs |
| Auth | Clerk | Clerk | https://clerk.com/docs |
| DB | PostgreSQL / Supabase | Supabase | https://supabase.com/docs |
| SQL Codegen | sqlc | — | https://docs.sqlc.dev |
| API | Go Echo v4 | Railway | https://echo.labstack.com |
| Job Queue | asynq v0.26.0 | Railway | https://github.com/hibiken/asynq |
| Cache Redis | Upstash (HTTP) | Upstash | https://docs.upstash.com/redis |
| Queue Redis | Railway Redis (TCP) | Railway | — |
| Storage | Cloudflare R2 | Cloudflare | https://developers.cloudflare.com/r2 |
| CDN | Cloudflare | Cloudflare | — |
| Payments | Stripe | Stripe | https://stripe.com/docs |
| Secrets | Doppler | Doppler | https://docs.doppler.com |
| Monorepo | Turborepo | — | https://turbo.build/repo/docs |
| Error | Sentry | Sentry | https://docs.sentry.io |
| Uptime/Logs | Better Stack | Better Stack | https://betterstack.com/docs |
| Analytics | PostHog | PostHog | https://posthog.com/docs |

---

## Frontend

### Next.js 15 App Router

```bash
# Install
pnpm create next-app@latest --app --ts --tailwind --eslint
```

**Qvora notes:**
- App Router only. **No Pages Router.**
- SSE generation at `/api/generation/[jobId]/stream` — standalone Route Handler with `ReadableStream`.
- `clerkMiddleware()` wraps the default middleware export in `middleware.ts`.

### shadcn/ui

```bash
# Init in monorepo
pnpm dlx shadcn@latest init -t next --monorepo

# Add components
pnpm dlx shadcn@latest add button card input badge skeleton
pnpm dlx shadcn@latest add dialog sheet tabs dropdown-menu

# Add to specific workspace
pnpm dlx shadcn@latest add card -c src/apps/web
```

**Qvora notes:** Components are copied (not imported) — fully customizable. Uses `@workspace/ui/components/[component]` path alias.

### Tailwind CSS v4

**Qvora rule: No `tailwind.config.ts`.** All tokens in `@theme {}` block in `globals.css`.

```css
/* globals.css */
@import "tailwindcss";

@theme {
  --color-volt: #7B2FFF;
  --color-convert-green: #00E87A;
  --color-signal-red: #FF3D3D;
  --color-data-blue: #2E9CFF;
  --color-qvora-black: #0A0A0F;
  --color-qvora-white: #F5F5F7;
}
```

---

## AI Layer

### Vercel AI SDK v6

```bash
pnpm add ai
pnpm add @ai-sdk/openai @ai-sdk/anthropic
```

| Use Case | Function | Model |
|---|---|---|
| Product data extraction | `generateObject` | GPT-4o (JSON strict) |
| Angle/hook generation | `streamText` | Claude Sonnet 4.6 |
| UI streaming | `useObject` hook | — |
| Inline regen | `streamText` | Claude Sonnet 4.6 |

```typescript
// generateObject — structured extraction
import { generateObject } from 'ai';
import { openai } from '@ai-sdk/openai';
const { object } = await generateObject({
  model: openai('gpt-4o'),
  schema: productSchema,   // Zod schema
  prompt: 'Extract product data from: ...',
});

// streamText — creative generation
import { streamText } from 'ai';
import { anthropic } from '@ai-sdk/anthropic';
const { textStream } = await streamText({
  model: anthropic('claude-sonnet-4-5'),
  prompt: '...',
});
```

### FAL.AI

```bash
pnpm add @fal-ai/client
```

> ⚠️ **Always `fal.queue.submit()`. Never `fal.subscribe()` — it blocks the worker thread.**

```typescript
import * as fal from '@fal-ai/client';

// Submit async job
const { request_id } = await fal.queue.submit('fal-ai/veo3.1', {
  input: { prompt: '...' },
  webhookUrl: 'https://api.qvora.com/v1/webhooks/fal',
});

// Models available:
// fal-ai/veo3.1        — Veo 3.1 (T2V premium)
// fal-ai/kling-video   — Kling 3.0 (T2V/I2V)
// fal-ai/runway-gen4   — Runway Gen-4.5
// fal-ai/sora          — Sora 2
```

### HeyGen Avatar API v3

> ⚠️ **HeyGen = v3 only.** `developers.heygen.com`. "v4" references in docs are errors — v4 does not exist. V2V lip-sync is v3-only.

```bash
# No npm package — use REST API directly
curl https://api.heygen.com/v3/...
# API Key header: X-Api-Key: {key}
```

### ElevenLabs

```bash
pnpm add elevenlabs
```

| Model | Use |
|---|---|
| `eleven_v3` | Quality TTS (default) |
| `eleven_flash_v2_5` | ~75ms latency for real-time voice |

### Langfuse

```bash
pnpm add langfuse
pnpm add @langfuse/langchain  # if using LangChain
```

Used for: LLM observability, prompt versioning, cost attribution per workspace (`org_id`).

---

## Backend — Go

### Go Echo v4

```bash
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
```

### asynq v0.26.0

```bash
go get github.com/hibiken/asynq@v0.26.0
```

> ⚠️ **asynq requires Railway Redis (TCP).** Upstash HTTP does not support `BLPOP` — asynq will not work with Upstash.

```go
// Server (worker process)
srv := asynq.NewServer(
    asynq.RedisClientOpt{Addr: os.Getenv("RAILWAY_REDIS_URL")},
    asynq.Config{Concurrency: 10},
)

// Client (enqueue from Go API)
client := asynq.NewClient(
    asynq.RedisClientOpt{Addr: os.Getenv("RAILWAY_REDIS_URL")},
)
task := asynq.NewTask("generation:video", payload)
client.Enqueue(task)
```

### sqlc

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate   # reads sqlc.yaml → generates Go types from SQL queries
```

Config in `src/services/api/sqlc.yaml`. Driver: pgx/v5.

---

## Data Layer

### Supabase + PostgreSQL

```bash
pnpm add @supabase/supabase-js
# OR for Go backend:
go get github.com/jackc/pgx/v5
```

**RLS rule:** Always `(SELECT auth.uid())` wrapper — never bare `auth.uid()`.

### Upstash Redis (HTTP — cache/rate-limit only)

```bash
pnpm add @upstash/redis
```

```typescript
import { Redis } from '@upstash/redis';
const redis = new Redis({
  url: process.env.UPSTASH_REDIS_REST_URL!,
  token: process.env.UPSTASH_REDIS_REST_TOKEN!,
});
```

> ⚠️ **Upstash is HTTP only.** Do NOT use for asynq job queues.

### Cloudflare R2

```bash
pnpm add @aws-sdk/client-s3 @aws-sdk/s3-request-presigner
```

R2 is S3-compatible. Use presigned PUT URLs for direct browser uploads. Zero egress cost.

### Mux

```bash
pnpm add @mux/mux-node         # Go backend — use REST API
pnpm add @mux/mux-player-react # Frontend player
```

---

## Auth — Clerk

```bash
pnpm add @clerk/nextjs
```

```typescript
// middleware.ts
import { clerkMiddleware } from '@clerk/nextjs/server';
export default clerkMiddleware();
```

**Qvora auth flow:** Clerk Organization = Qvora workspace. JWT carries `org_id` + `org_role`. Go API middleware validates JWT on every request; tier limits enforced in Go — never client-side.

---

## Payments — Stripe

```bash
pnpm add stripe @stripe/stripe-js
```

**Key webhooks to handle:**
- `invoice.paid` — activate subscription
- `customer.subscription.updated` — tier change
- `customer.subscription.deleted` — cancel → lock generation on day 8

---

## Open Questions
- [ ] Exact FAL.AI model IDs for Veo 3.1 and Sora 2 — confirm from FAL.AI docs before building.
- [ ] Langfuse self-hosted vs. cloud for production? (cost tradeoff)

## Related Pages
- [[stack-overview]] — high-level stack summary
- [[ai-layer]] — AI model usage patterns
- [[data-layer]] — Redis, R2, Mux in depth
- [[system-architecture]] — how all these pieces connect
