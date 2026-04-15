# Qvora — Implementation Reference Resources

> Authoritative doc links, SDK identifiers, and Qvora-specific implementation notes for every layer of the stack. Use this as the single lookup document when building any module.

---

## Quick Stack Snapshot

| Layer | Technology | Host |
|---|---|---|
| Frontend | Next.js 15 App Router + shadcn/ui + Tailwind v4 | Vercel |
| BFF | tRPC (TypeScript ≥5.7.2) | Vercel (Edge/Node) |
| Generation Stream | SSE Route Handler `/api/generation/[jobId]/stream` | Vercel |
| API | Go Echo v4 | Railway |
| Job Workers | Go + asynq | Railway |
| Video Post-Processing | Rust Axum + ffmpeg-sys | Railway |
| AI Orchestration | Vercel AI SDK v6 | — |
| Text-to-Video | FAL.AI (Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2) | FAL.AI |
| Avatar Lip-Sync | HeyGen Avatar API v3 | HeyGen |
| TTS / Voice Clone | ElevenLabs API | ElevenLabs |
| URL Scraping | Modal + Playwright | Modal |
| LLM Observability | Langfuse | Cloud / Self-host |
| Auth | Clerk (multi-tenant) | Clerk |
| Primary DB | PostgreSQL via Supabase (RLS) | Supabase |
| SQL Codegen | sqlc | — |
| Cache / Rate-limit Redis | Upstash Redis (HTTP) | Upstash |
| Job Queue Redis | Railway Redis (TCP) | Railway |
| Object Storage | Cloudflare R2 (zero egress) | Cloudflare |
| CDN | Cloudflare | Cloudflare |
| Video Hosting | Mux (HLS + analytics) | Mux |
| Payments | Stripe Billing + usage metering | Stripe |
| Secrets | Doppler | Doppler |
| Monorepo | Turborepo | — |
| CI/CD | GitHub Actions | GitHub |
| Error Tracking | Sentry | Sentry |
| Uptime / Logs | Better Stack | Better Stack |
| Product Analytics | PostHog | PostHog |

---

## 1. Frontend

### 1.1 Next.js 15 App Router

| Item | Value |
|---|---|
| Docs | https://nextjs.org/docs/app |
| Getting Started | https://nextjs.org/docs/getting-started/installation |
| Streaming UI | https://nextjs.org/docs/app/building-your-application/rendering/server-components#streaming |
| Route Handlers (SSE) | https://nextjs.org/docs/app/building-your-application/routing/route-handlers |
| Middleware | https://nextjs.org/docs/app/building-your-application/routing/middleware |

**Qvora notes:**
- App Router only. No Pages Router.
- SSE generation progress uses a standalone Route Handler at `/api/generation/[jobId]/stream` with `ReadableStream` — **not** tRPC.
- `clerkMiddleware()` wraps the default middleware export.

---

### 1.2 shadcn/ui

| Item | Value |
|---|---|
| Docs | https://ui.shadcn.com/docs |
| Next.js Install | https://ui.shadcn.com/docs/installation/next |
| Monorepo Install | https://ui.shadcn.com/docs/installation/next (use `--monorepo` flag) |
| Components | https://ui.shadcn.com/docs/components |
| Theming | https://ui.shadcn.com/docs/theming |
| CLI | https://ui.shadcn.com/docs/cli |

**Key commands:**
```bash
# New monorepo project
pnpm dlx shadcn@latest init -t next --monorepo

# Add a component
pnpm dlx shadcn@latest add card
pnpm dlx shadcn@latest add button

# Add to specific workspace in monorepo
pnpm dlx shadcn@latest add card -c apps/web
```

**Qvora notes:**
- Components are copied (not imported) into the repo — customizable.
- Uses `@workspace/ui/components/[component]` path in Turborepo setup.
- Tailwind v4 CSS-first config (`@theme {}`) — no `tailwind.config.ts`.

---

### 1.3 Tailwind CSS v4

| Item | Value |
|---|---|
| Docs | https://tailwindcss.com/docs |
| v4 Migration | https://tailwindcss.com/docs/upgrade-guide |
| CSS Variables / @theme | https://tailwindcss.com/docs/theme |

**Qvora notes:**
- Config lives in CSS (`@theme {}` block in `globals.css`) — no JS config file.
- Custom tokens `--success`, `--data` registered in `@theme {}` for `bg-success`, `text-data` utilities.

---

### 1.4 Framer Motion

| Item | Value |
|---|---|
| Docs | https://motion.dev/docs/react-quick-start |
| Animation | https://motion.dev/docs/react-animation |
| Layout | https://motion.dev/docs/react-layout-animations |

---

### 1.5 Zustand + TanStack Query v5

| Item | Value |
|---|---|
| Zustand Docs | https://zustand.docs.pmnd.rs/ |
| TanStack Query | https://tanstack.com/query/latest/docs/framework/react/overview |
| TanStack Query SSR (Next.js) | https://tanstack.com/query/latest/docs/framework/react/guides/advanced-ssr |

---

### 1.6 React Hook Form + Zod

| Item | Value |
|---|---|
| React Hook Form | https://react-hook-form.com/docs |
| Zod | https://zod.dev/ |
| RHF + Zod resolver | https://react-hook-form.com/docs/useform#resolver |

---

### 1.7 Mux Player (frontend)

| Item | Value |
|---|---|
| Mux Player React | https://www.mux.com/player |
| Docs | https://docs.mux.com/guides/mux-player |
| `<MuxPlayer>` props | https://docs.mux.com/guides/mux-player-web-guide |

---

## 2. BFF — tRPC

| Item | Value |
|---|---|
| Docs | https://trpc.io/docs |
| Quickstart | https://trpc.io/docs/quickstart |
| Next.js App Router adapter | https://trpc.io/docs/client/nextjs/app-dir |
| React Query integration | https://trpc.io/docs/client/react |
| Input validation (Zod) | https://trpc.io/docs/server/validators |

**Key patterns:**
```typescript
// Router setup
import { initTRPC } from '@trpc/server';
import { z } from 'zod';

const t = initTRPC.create();
const router = t.router;
const publicProcedure = t.procedure;

export const appRouter = router({
  getBriefs: publicProcedure
    .input(z.object({ workspaceId: z.string() }))
    .query(async ({ input }) => { /* ... */ }),
  createGeneration: publicProcedure
    .input(z.object({ briefId: z.string() }))
    .mutation(async ({ input }) => { /* ... */ }),
});
```

**Client setup:**
```typescript
import { createTRPCClient, httpBatchLink } from '@trpc/client';

const client = createTRPCClient<AppRouter>({
  links: [httpBatchLink({ url: '/api/trpc' })],
});
```

**Qvora notes:**
- TypeScript ≥5.7.2 required.
- SSE generation progress stream is a **standalone Route Handler** — not a tRPC subscription.
- Go API calls tRPC endpoints from server-side, or Go API is called directly via REST from tRPC procedures.

---

## 3. Backend API — Go Echo v4

| Item | Value |
|---|---|
| Echo v4 Docs | https://echo.labstack.com/docs |
| Quickstart | https://echo.labstack.com/docs/quick-start |
| Middleware | https://echo.labstack.com/docs/middleware |
| JWT Middleware | https://echo.labstack.com/docs/middleware/jwt |
| Rate Limiter | https://echo.labstack.com/docs/middleware/rate-limiter |

**Key patterns:**
```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.GET("/health", func(c echo.Context) error {
    return c.JSON(200, map[string]string{"status": "ok"})
})
e.Logger.Fatal(e.Start(":8080"))
```

---

## 4. Job Workers — asynq

| Item | Value |
|---|---|
| GitHub | https://github.com/hibiken/asynq |
| pkg.go.dev | https://pkg.go.dev/github.com/hibiken/asynq |
| Getting Started Wiki | https://github.com/hibiken/asynq/wiki/Getting-Started |
| Queue Priority | https://github.com/hibiken/asynq/wiki/Queue-Priority |
| Periodic Tasks | https://github.com/hibiken/asynq/wiki/Periodic-Tasks |
| Task Retry | https://github.com/hibiken/asynq/wiki/Task-Retry |
| Asynqmon (Web UI) | https://github.com/hibiken/asynqmon |

**Install:**
```bash
go get -u github.com/hibiken/asynq
```

**Key patterns:**

```go
// Enqueue (from Go API)
client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
defer client.Close()

task := asynq.NewTask("video:generate", payload, asynq.MaxRetry(3), asynq.Timeout(20*time.Minute))
info, err := client.Enqueue(task, asynq.Queue("critical"))

// Process (worker binary)
srv := asynq.NewServer(
    asynq.RedisClientOpt{Addr: redisAddr},
    asynq.Config{
        Concurrency: 5,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
    },
)

mux := asynq.NewServeMux()
mux.HandleFunc("video:generate", handleVideoGenerate)
srv.Run(mux)
```

**Qvora notes:**
- **Requires TCP Redis** (blocking BLPOP). Incompatible with Upstash HTTP Redis.
- Use **Railway Redis** (container) for asynq job queues.
- **Upstash Redis** (HTTP via `@upstash/redis`) is for caching/rate limiting only.
- Task types for Qvora: `video:generate`, `brief:analyze`, `signal:process`, `video:postprocess`.
- v0.26.0 is the current stable release (MIT license, 13k+ stars).
- Use `asynq.Unique(ttl)` option to deduplicate repeated generation requests.

---

## 5. AI Layer

### 5.1 Vercel AI SDK v6

| Item | Value |
|---|---|
| Docs | https://ai-sdk.dev/docs/introduction |
| Next.js App Router guide | https://ai-sdk.dev/docs/getting-started/nextjs-app-router |
| AI SDK Core | https://ai-sdk.dev/docs/ai-sdk-core/overview |
| AI SDK UI (hooks) | https://ai-sdk.dev/docs/ai-sdk-ui/overview |
| Generating structured data | https://ai-sdk.dev/docs/ai-sdk-core/generating-structured-data |
| Streaming | https://ai-sdk.dev/docs/foundations/streaming |
| Object generation (useObject) | https://ai-sdk.dev/docs/ai-sdk-ui/object-generation |
| Providers index | https://ai-sdk.dev/providers |
| OpenAI provider | https://ai-sdk.dev/providers/ai-sdk-providers/openai |
| Anthropic provider | https://ai-sdk.dev/providers/ai-sdk-providers/anthropic |
| FAL AI provider | https://ai-sdk.dev/providers/ai-sdk-providers/fal |
| llms.txt (for agents/IDE) | https://ai-sdk.dev/llms.txt |

**Qvora usage:**
- `generateObject()` with `schema: z.object(...)` for structured brief/strategy JSON (GPT-4o)
- `streamText()` for creative copy regen streaming (Claude Sonnet 4.6)
- FAL AI provider for video generation dispatch (T2V/I2V)
- `useObject()` hook for streaming structured generation progress in the UI

---

### 5.2 FAL.AI

| Item | Value |
|---|---|
| Docs | https://fal.ai/docs |
| Models index | https://fal.ai/models |
| JavaScript client | https://www.npmjs.com/package/@fal-ai/client |
| Python client | https://pypi.org/project/fal-client/ |
| Serverless deploy | https://fal.ai/docs/deploy |
| Queue API | https://fal.ai/docs/model-endpoints/queue |
| Veo 3.1 | https://fal.ai/models/fal-ai/veo3 |
| Kling 3.0 | https://fal.ai/models/fal-ai/kling-video |
| Runway Gen-4.5 | https://fal.ai/models/fal-ai/runway-gen4-turbo |

**Key pattern:**
```typescript
import * as fal from "@fal-ai/client";

// Async queue (long-running video gen)
const { request_id } = await fal.queue.submit("fal-ai/veo3", {
  input: { prompt, aspect_ratio: "16:9" },
});

// Poll or subscribe to stream
const result = await fal.queue.result("fal-ai/veo3", { requestId: request_id });
```

**Qvora notes:**
- Use `fal.queue.submit()` (async) for video generation — not `fal.subscribe()` (sync).
- Store `request_id` in asynq task payload; poll from worker via `fal.queue.status()`.
- GPU scaling: `min_concurrency`, `max_concurrency`, `concurrency_buffer` params in fal deploy config.

---

### 5.3 ElevenLabs (TTS / Voice Clone)

| Item | Value |
|---|---|
| Docs | https://elevenlabs.io/docs/overview |
| API Reference | https://elevenlabs.io/docs/api-reference/text-to-speech |
| Node.js SDK | https://www.npmjs.com/package/elevenlabs |
| Python SDK | https://pypi.org/project/elevenlabs/ |
| Voice IDs | https://elevenlabs.io/docs/voices/voices-list |
| Models | https://elevenlabs.io/docs/models |

**Key models:**
| Model ID | Notes |
|---|---|
| `eleven_v3` | Highest quality, 70+ languages |
| `eleven_flash_v2_5` | ~75ms latency, production streaming |
| `eleven_multilingual_v2` | Best non-English quality |

**Billing:** 1 credit = 1 character.

**Qvora notes:**
- Use `eleven_flash_v2_5` for real-time TTS in avatar generation pipeline.
- Store voice IDs mapped to brand personas in Supabase config table.
- Voice clone workflow: upload audio samples → create custom voice → store `voice_id`.

---

### 5.4 HeyGen Avatar API v3

| Item | Value |
|---|---|
| v3 Docs (active) | https://developers.heygen.com/ |
| Endpoint comparison | https://developers.heygen.com/endpoint-version-comparison |
| Legacy docs (v2, until Oct 2026) | https://docs.heygen.com/docs/quick-start |
| Avatar Video Generation | https://developers.heygen.com |
| Lip-sync (V2V) | Available in v3 under "Lipsyncs" |
| Webhooks | https://docs.heygen.com/docs/using-heygens-webhook-events |

**Key v3 features over v2:**
- Video Agent with styles + references + interactive chat
- Lipsyncs (V2V lip-sync — critical for Qvora)
- Voice Design
- Agentic CLI / MCP / Skills

**Qvora notes:**
- Architecture doc says "HeyGen Avatar API v4" — update reference to v3 (v4 does not exist; v3 is the latest active platform as of 2025).
- **Migrate to v3** before production: v2 sunset October 31, 2026.
- V2V lip-sync is a v3-only feature (not in v2).

---

### 5.5 Langfuse (LLM Observability)

| Item | Value |
|---|---|
| Docs | https://langfuse.com/docs |
| Self-host | https://langfuse.com/docs/deployment/self-host |
| Vercel AI SDK integration | https://langfuse.com/docs/integrations/vercelaisdk |
| Prompt management | https://langfuse.com/docs/prompts/get-started |
| Evaluations | https://langfuse.com/docs/scores/overview |
| OpenTelemetry | https://langfuse.com/docs/opentelemetry/introduction |

**Qvora notes:**
- Use for versioned prompt management (brief analysis prompts, strategy generation prompts).
- A/B test prompt versions via Langfuse prompt labels: `production`, `staging`, `experiment-v2`.
- Trace every GPT-4o / Claude call with `userId = workspace_id` for per-workspace cost attribution.

---

## 6. Auth — Clerk

| Item | Value |
|---|---|
| Docs | https://clerk.com/docs |
| Next.js App Router quickstart | https://clerk.com/docs/quickstarts/nextjs |
| Multi-tenancy (Organizations) | https://clerk.com/docs/organizations/overview |
| SSO / SAML | https://clerk.com/docs/authentication/enterprise-connections |
| JWT Templates | https://clerk.com/docs/backend-requests/making/jwt-templates |
| Webhooks | https://clerk.com/docs/integrations/webhooks |

**Packages:**
```bash
npm install @clerk/nextjs
```

**Key setup:**
```typescript
// middleware.ts
import { clerkMiddleware } from '@clerk/nextjs/server';
export default clerkMiddleware();
export const config = { matcher: ['/((?!.+\\.[\\w]+$)|/).*', '/', '/(api|trpc)(.*)'] };

// layout.tsx
import { ClerkProvider, SignInButton, UserButton } from '@clerk/nextjs';
export default function RootLayout({ children }) {
  return (
    <ClerkProvider>
      <html><body>
        <SignInButton /><UserButton />
        {children}
      </body></html>
    </ClerkProvider>
  );
}
```

**Qvora notes:**
- Multi-tenant: one Clerk Organization = one Qvora workspace.
- JWT template injects `org_id` (workspace ID) and `org_role` into JWT — validated by Go API middleware.
- Pass Clerk JWT as `Authorization: Bearer <token>` to Go API; verify with Clerk public JWKS.

---

## 7. Database

### 7.1 Supabase PostgreSQL + RLS

| Item | Value |
|---|---|
| Docs | https://supabase.com/docs |
| Row Level Security | https://supabase.com/docs/guides/database/postgres/row-level-security |
| Auth helpers | https://supabase.com/docs/guides/auth |
| pgvector (V2 feature) | https://supabase.com/docs/guides/database/extensions/pgvector |
| Connecting to Postgres | https://supabase.com/docs/guides/database/connecting-to-postgres |
| RLS Performance guide | https://github.com/GaryAustin1/RLS-Performance |

**Core RLS patterns for Qvora:**
```sql
-- Enable RLS on all tables
ALTER TABLE org_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE briefs ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets ENABLE ROW LEVEL SECURITY;

-- Org isolation (org_id from JWT — matches Database Schema column name)
CREATE POLICY "org_isolation" ON briefs
  TO authenticated
  USING (
    org_id = (SELECT auth.jwt() ->> 'org_id')::UUID
  );
```

**Performance rules (always apply):**
1. Wrap `auth.uid()` in `(SELECT auth.uid())` — enables initPlan caching.
2. Add `TO authenticated` role specifier on every policy.
3. Index columns used in policy USING clauses.
4. Add explicit filters in queries; don't rely solely on RLS.

---

### 7.2 sqlc

| Item | Value |
|---|---|
| Docs | https://docs.sqlc.dev/en/latest/ |
| Install | https://docs.sqlc.dev/en/latest/overview/install.html |
| PostgreSQL quickstart | https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html |
| Using Go + pgx | https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html |
| Configuration reference | https://docs.sqlc.dev/en/latest/reference/config.html |
| Query annotations | https://docs.sqlc.dev/en/latest/reference/query-annotations.html |

**Workflow:**
1. Write SQL queries with annotations in `.sql` files.
2. Run `sqlc generate` → type-safe Go structs + methods.
3. Use generated `Queries` struct in Go API handlers.

**Example:**
```sql
-- name: GetBriefsByWorkspace :many
SELECT * FROM briefs
WHERE workspace_id = $1
ORDER BY created_at DESC;
```
→ Generates `GetBriefsByWorkspace(ctx, workspaceID uuid.UUID) ([]Brief, error)`

**Qvora notes:**
- Use `pgx/v5` driver (fastest).
- `sqlc.yaml` config: `engine: postgresql`, `gen: go`, `out: internal/db`.

---

### 7.3 Upstash Redis (Cache / Rate-limit)

| Item | Value |
|---|---|
| Docs | https://upstash.com/docs/redis/overall/getstarted |
| `@upstash/redis` SDK | https://upstash.com/docs/redis/sdks/ts/overview |
| Rate Limiting | https://upstash.com/docs/oss/sdks/ts/ratelimit/overview |
| Next.js integration | https://upstash.com/docs/redis/integrations/nextjs |

**Install:**
```bash
npm install @upstash/redis @upstash/ratelimit
```

**Qvora notes:**
- HTTP/REST-based — no persistent TCP connection. Works on Vercel Edge.
- Use for: generation result caching, generation-per-workspace rate limiting (tier enforcement), session token cache.
- **Cannot** be used with asynq (requires TCP BLPOP).

---

### 7.4 Railway Redis (Job Queue)

| Item | Value |
|---|---|
| Railway Docs | https://docs.railway.com |
| Redis service | https://docs.railway.com/databases/redis |
| Environment variables | `RAILWAY_REDIS_URL` |

**Qvora notes:**
- Standard TCP Redis container on Railway.
- Connection string: `redis://[:password@]host:port` — use `asynq.RedisClientOpt{Addr: ..., Password: ...}`.
- Run asynq worker binary on a separate Railway service in the same project.

---

### 7.5 Cloudflare R2

| Item | Value |
|---|---|
| Docs | https://developers.cloudflare.com/r2/ |
| S3 API compatibility | https://developers.cloudflare.com/r2/api/s3/ |
| Workers integration | https://developers.cloudflare.com/r2/api/workers/ |
| Presigned URLs | https://developers.cloudflare.com/r2/api/s3/presigned-urls/ |

**Install (AWS SDK v3 — S3 compatible):**
```bash
npm install @aws-sdk/client-s3 @aws-sdk/s3-request-presigner
```

**Qvora notes:**
- Zero egress cost — generated videos, thumbnails, and uploads stored here.
- Use presigned PUT URLs for direct browser → R2 upload (avoids routing through API).
- Cloudflare CDN serves R2 assets via public bucket or custom domain.

---

## 8. Video — Mux

| Item | Value |
|---|---|
| Docs | https://docs.mux.com |
| Video overview | https://docs.mux.com/docs/video |
| Upload a video | https://docs.mux.com/api-reference/video#operation/create-direct-upload |
| Playback | https://docs.mux.com/guides/stream-video-files |
| Signed URLs | https://docs.mux.com/guides/enable-signed-urls-for-your-assets |
| Mux Player | https://docs.mux.com/guides/mux-player |
| Mux Data (analytics SDK) | https://docs.mux.com/guides/data-monitor-video-quality |
| Node.js SDK | https://www.npmjs.com/package/@mux/mux-node |

**Qvora notes:**
- Upload finished video from R2 → Mux via URL ingestion (not re-upload).
- Use `playback_policy: "signed"` for workspace-scoped video access.
- `<MuxPlayer playbackId="..." tokens={{ playback: signedToken }}/>` for secure playback.
- Mux Data SDK tracks watch-through rate per variant — feeds into Performance Signal score.

---

## 9. Payments — Stripe

| Item | Value |
|---|---|
| Docs | https://docs.stripe.com |
| Subscriptions overview | https://docs.stripe.com/billing/subscriptions/overview |
| Subscriptions quickstart | https://docs.stripe.com/billing/quickstart |
| Usage-based billing | https://docs.stripe.com/billing/subscriptions/usage-based |
| Entitlements | https://docs.stripe.com/billing/entitlements |
| Trials | https://docs.stripe.com/billing/subscriptions/trials |
| Webhooks | https://docs.stripe.com/webhooks |
| Portal (self-service) | https://docs.stripe.com/customer-management |
| Node.js SDK | https://www.npmjs.com/package/stripe |

**Subscription statuses:**
| Status | Meaning |
|---|---|
| `trialing` | 7-day trial active — provision access |
| `active` | Paid and current — provision access |
| `past_due` | Payment failed — restrict new generations, show warning |
| `canceled` | Subscription ended — lock to read-only |
| `incomplete` | Initial payment pending (<23h) |
| `paused` | Trial ended, no payment method |

**Key webhooks to handle:**
- `invoice.paid` → activate/continue subscription
- `customer.subscription.trial_will_end` → Day 6 conversion email trigger
- `customer.subscription.updated` → tier change, provision/deprovision features
- `customer.subscription.deleted` → lock workspace

**Qvora tiers:**
| Plan | Price | Variant Limit |
|---|---|---|
| Starter | $99/mo | Max 3 variants/angle |
| Growth | $149/mo | Max 10 variants/angle |
| Agency | $399/mo | Unlimited |

**Qvora notes:**
- `payment_behavior: "default_incomplete"` on subscription create for 3DS safety.
- Store `stripe_customer_id`, `stripe_subscription_id`, `plan_tier` on workspace record.
- Listen to webhook in `/api/webhooks/stripe` — verify with `stripe.webhooks.constructEvent()`.

---

## 10. Infrastructure

### 10.1 Vercel

| Item | Value |
|---|---|
| Docs | https://vercel.com/docs |
| Next.js on Vercel | https://vercel.com/docs/frameworks/nextjs |
| Environment variables | https://vercel.com/docs/projects/environment-variables |
| Edge Runtime | https://vercel.com/docs/functions/edge-functions |
| Turborepo | https://turbo.build/repo/docs |

---

### 10.2 Railway

| Item | Value |
|---|---|
| Docs | https://docs.railway.com |
| Deploy from Dockerfile | https://docs.railway.com/deploy/dockerfiles |
| Services | https://docs.railway.com/reference/services |
| Private networking | https://docs.railway.com/reference/private-networking |
| Redis | https://docs.railway.com/databases/redis |
| Environment variables | https://docs.railway.com/develop/variables |

**Qvora services on Railway:**
1. `qvora-api` — Go Echo v4 API binary
2. `qvora-worker` — Go asynq worker binary
3. `qvora-redis` — Redis TCP container (asynq queues)
4. `qvora-video-proc` — Rust Axum + ffmpeg-sys binary

---

### 10.3 Modal (Playwright Scraping)

| Item | Value |
|---|---|
| Docs | https://modal.com/docs/guide |
| Functions | https://modal.com/docs/guide/functions |
| Containers | https://modal.com/docs/guide/custom-container |
| GPU inference | https://modal.com/docs/guide/gpu |
| Calling from other services | https://modal.com/docs/guide/webhook |

**Key pattern:**
```python
import modal

app = modal.App()
playwright_image = modal.Image.debian_slim().run_commands(
    "pip install playwright",
    "playwright install chromium"
)

@app.function(image=playwright_image, timeout=120)
async def scrape_url(url: str) -> dict:
    from playwright.async_api import async_playwright
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        page = await browser.new_page()
        await page.goto(url)
        content = await page.content()
        await browser.close()
        return {"url": url, "html": content}
```

**Calling from Go (HTTP):**
```go
// Modal exposes functions as HTTP webhooks
resp, err := http.Post(modalWebhookURL, "application/json", body)
```

**Qvora notes:**
- Sub-second cold starts — no persistent container cost.
- Go orchestrator calls Modal webhook URL to kick off Playwright scrape.
- Each scrape call billed per-second of compute.

---

### 10.4 Doppler (Secrets)

| Item | Value |
|---|---|
| Docs | https://docs.doppler.com/docs |
| Getting Started | https://docs.doppler.com/docs/getting-started |
| CLI install | https://docs.doppler.com/docs/install-cli |
| Vercel integration | https://docs.doppler.com/docs/vercel |
| Railway integration | https://docs.doppler.com/docs/railway |
| GitHub Actions | https://docs.doppler.com/docs/github-actions |
| Service tokens (CI/CD) | https://docs.doppler.com/docs/service-tokens |

**CLI usage:**
```bash
# Local development
doppler run -- pnpm dev

# Inject into Docker
doppler run -- docker-compose up
```

**Qvora notes:**
- Project: `qvora`. Environments: `dev`, `stg`, `prd`.
- Vercel: Doppler sync → Vercel project env vars (automated).
- Railway: Doppler CLI in Dockerfile or Railway env var sync.
- GitHub Actions: `DOPPLER_TOKEN` secret → `doppler run -- <command>`.

---

## 11. Observability & Analytics

### 11.1 Sentry

| Item | Value |
|---|---|
| Docs | https://docs.sentry.io |
| Next.js SDK | https://docs.sentry.io/platforms/javascript/guides/nextjs/ |
| Go SDK | https://docs.sentry.io/platforms/go/ |
| Source maps | https://docs.sentry.io/platforms/javascript/guides/nextjs/sourcemaps/ |

---

### 11.2 PostHog

| Item | Value |
|---|---|
| Docs | https://posthog.com/docs |
| Next.js integration | https://posthog.com/docs/libraries/next-js |
| Feature flags | https://posthog.com/docs/feature-flags |
| Session replay | https://posthog.com/docs/session-replay |
| Funnels | https://posthog.com/docs/product-analytics/funnels |

**Qvora key events to track:**
- `brief_submitted`, `generation_started`, `generation_completed`, `variant_previewed`
- `upgrade_clicked`, `trial_started`, `subscription_converted`
- `signal_viewed`, `export_downloaded`

---

### 11.3 Better Stack

| Item | Value |
|---|---|
| Docs | https://betterstack.com/docs |
| Uptime monitoring | https://betterstack.com/docs/uptime |
| Log management | https://betterstack.com/docs/logs |
| Next.js logging | https://betterstack.com/docs/logs/javascript |
| Go logging | https://betterstack.com/docs/logs/go |

---

## 12. Monorepo — Turborepo

| Item | Value |
|---|---|
| Docs | https://turbo.build/repo/docs |
| Getting Started | https://turbo.build/repo/docs/getting-started |
| Workspace configuration | https://turbo.build/repo/docs/crafting-your-repository |
| Remote caching | https://turbo.build/repo/docs/core-concepts/caching |
| GitHub Actions | https://turbo.build/repo/docs/ci/github-actions |

**Qvora monorepo structure:**
```
apps/
  web/          # Next.js 15 App Router (frontend + BFF)
packages/
  ui/           # shadcn/ui components
  types/        # Shared TypeScript types
  config/       # ESLint, TS, Tailwind configs
services/
  api/          # Go Echo API
  worker/       # Go asynq worker
  postprocess/  # Rust Axum video processor
```

---

## 13. Critical Integration Notes

### 13.1 Two Redis Instances

| Instance | Provider | Protocol | Used For |
|---|---|---|---|
| `REDIS_QUEUE_URL` | Railway Redis | TCP | asynq job queues — **required** |
| `UPSTASH_REDIS_REST_URL` | Upstash | HTTP/REST | Caching, rate limiting, session store |

**Rule:** Never substitute Upstash for the Railway Redis instance. asynq uses `BLPOP`/`BRPOP` which requires a persistent TCP connection that Upstash HTTP does not support.

---

### 13.2 SSE Generation Progress (Not tRPC)

The generation progress stream is a **standalone Next.js Route Handler**, not a tRPC subscription:

```typescript
// app/api/generation/[jobId]/stream/route.ts
export async function GET(
  request: Request,
  { params }: { params: { jobId: string } }
) {
  const encoder = new TextEncoder();
  const stream = new ReadableStream({
    async start(controller) {
      // Poll job status from Go API / Redis
      // Emit SSE events: { stage: "analyzing" | "generating" | "processing" | "done", progress: 0-100 }
      const interval = setInterval(async () => {
        const status = await getJobStatus(params.jobId);
        controller.enqueue(encoder.encode(`data: ${JSON.stringify(status)}\n\n`));
        if (status.stage === "done" || status.stage === "error") {
          clearInterval(interval);
          controller.close();
        }
      }, 1000);
    }
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    },
  });
}
```

---

### 13.3 HeyGen Version Migration

Architecture documents reference "HeyGen API v4" — this is incorrect. The actual versions are:
- **v3** = current active platform (`developers.heygen.com`)
- **v2** = legacy, supported until Oct 31, 2026
- **v4** = does not exist

**Action required:** Migrate lip-sync integration to v3 API. V2V lip-sync is only available in v3.

---

### 13.4 Stripe Tier Enforcement in Go API

```go
// middleware: check generation limits before enqueuing
func checkTierLimits(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        workspace := getWorkspaceFromContext(c)
        variantCount := getVariantCountForAngle(workspace.ID)
        
        maxVariants := map[string]int{
            "starter": 3,
            "growth":  10,
            "scale":   -1, // unlimited
        }[workspace.PlanTier]
        
        if maxVariants != -1 && variantCount >= maxVariants {
            return c.JSON(402, map[string]string{
                "error": "variant_limit_exceeded",
                "tier":  workspace.PlanTier,
            })
        }
        return next(c)
    }
}
```

---

## 14. Package / Module Quick Reference

| Purpose | Package | Install |
|---|---|---|
| AI SDK (TS) | `ai` | `npm install ai` |
| FAL AI client | `@fal-ai/client` | `npm install @fal-ai/client` |
| ElevenLabs TS | `elevenlabs` | `npm install elevenlabs` |
| Clerk Next.js | `@clerk/nextjs` | `npm install @clerk/nextjs` |
| tRPC server | `@trpc/server` | `npm install @trpc/server` |
| tRPC client | `@trpc/client` | `npm install @trpc/client` |
| tRPC React | `@trpc/react-query` | `npm install @trpc/react-query` |
| Upstash Redis | `@upstash/redis` | `npm install @upstash/redis` |
| Upstash Rate Limit | `@upstash/ratelimit` | `npm install @upstash/ratelimit` |
| Mux Node | `@mux/mux-node` | `npm install @mux/mux-node` |
| Mux Player | `@mux/mux-player-react` | `npm install @mux/mux-player-react` |
| Stripe | `stripe` | `npm install stripe` |
| Zod | `zod` | `npm install zod` |
| TanStack Query | `@tanstack/react-query` | `npm install @tanstack/react-query` |
| Zustand | `zustand` | `npm install zustand` |
| React Hook Form | `react-hook-form` | `npm install react-hook-form` |
| Framer Motion | `framer-motion` | `npm install framer-motion` |
| Langfuse JS | `langfuse` | `npm install langfuse` |
| PostHog JS | `posthog-js` | `npm install posthog-js` |
| Sentry Next.js | `@sentry/nextjs` | `npm install @sentry/nextjs` |
| DnD Kit (drag-to-reorder) | `@dnd-kit/core @dnd-kit/sortable` | `npm install @dnd-kit/core @dnd-kit/sortable` |
| cmdk (command palette) | `cmdk` | `npm install cmdk` |
| asynq (Go) | `github.com/hibiken/asynq` | `go get -u github.com/hibiken/asynq` |
| Echo (Go) | `github.com/labstack/echo/v4` | `go get github.com/labstack/echo/v4` |
| sqlc (Go gen tool) | binary | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |

---

*Last updated: compiled from official documentation. Verify version-specific APIs before implementation.*
