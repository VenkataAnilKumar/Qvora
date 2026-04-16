# Qvora — Repository Structure

> **Decision:** Turborepo Monorepo  
> **Status:** Locked  
> **Last validated:** April 2026 (against Turborepo 2.x, Next.js 15, Go 1.24, Rust Axum 0.7+)

---

## Structure Decision

### Options Evaluated

| Structure | Description |
|---|---|
| **Monorepo** | All services + packages in one repo, one CI system |
| **Polyrepo** | Each service in its own repository |
| **Hybrid** | Frontend monorepo + separate backend repos |

### Decision: Monorepo (Turborepo)

**Rationale:**

1. **Shared TypeScript types are a hard requirement.** `src/packages/types` is consumed by `src/apps/web` and tRPC routers. In a polyrepo this requires versioned npm publishes on every change. In a monorepo it is a zero-config workspace import.

2. **Cross-service changes ship atomically.** A single generation feature touches `web → api → worker → postprocess`. One PR. One review. One deploy sequence. Polyrepo means 4 PRs with coordination overhead that kills pre-launch velocity.

3. **Turborepo path filtering gives polyrepo-level deploy isolation anyway.** Each service deploys independently via GitHub Actions path filters — no redeployment triggered unless that service's code changed.

4. **Go + Rust coexist cleanly.** `src/services/` is completely independent of npm workspaces. Go uses `go.mod`; Rust uses `Cargo.toml`. Turborepo only orchestrates the TypeScript layer. No toolchain conflicts.

5. **Stage of company.** Polyrepo overhead (cross-repo PRs, dep bumping, access management) only pays off with dedicated per-service teams. At pre-launch this is pure friction with zero benefit.

### When to Revisit (Not V1)

| Signal | Stage |
|---|---|
| Dedicated team per service with separate deploy cadences | Series A+ |
| Security requirement to restrict repo access per service | Enterprise |
| Repo clone/CI time becomes a bottleneck | 500k+ LOC |

---

## Full Repository Structure

```
qvora/                                   ← Turborepo root
│
├── .github/
│   ├── copilot-instructions.md          ← Auto-loaded by GitHub Copilot
│   ├── workflows/
│   │   ├── ci.yml                       ← Lint + typecheck + test (all packages)
│   │   ├── deploy-web.yml               ← Trigger: src/apps/web/**, src/packages/**
│   │   ├── deploy-api.yml               ← Trigger: src/services/api/**
│   │   ├── deploy-worker.yml            ← Trigger: src/services/worker/**
│   │   ├── deploy-postprocess.yml       ← Trigger: src/services/postprocess/**
│   │   ├── deploy-db.yml                ← Trigger: supabase/migrations/** → supabase db push
│   │   └── security.yml                 ← CodeQL scan
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yaml
│   │   ├── feature_request.yaml
│   │   └── question.yaml
│   ├── CODEOWNERS                       ← Per-service review ownership
│   ├── SECURITY.md                      ← Vulnerability reporting policy
│   └── pull_request_template.md
│
├── src/
│   ├── apps/
│   │   └── web/                         ← Next.js 15 App Router (→ Vercel)
│   │       ├── app/
│       │   ├── layout.tsx               ← Root layout + ClerkProvider
│       │   ├── page.tsx                 ← Marketing / home
│       │   ├── (auth)/                  ← Route group — no URL segment
│       │   │   ├── layout.tsx
│       │   │   ├── sign-in/[[...sign-in]]/page.tsx
│       │   │   └── sign-up/[[...sign-up]]/page.tsx
│       │   ├── (onboarding)/            ← Route group — post-signup wizard
│       │   │   ├── page.tsx             ← Step 1: org + workspace setup (S-12)
│       │   │   └── brand/page.tsx       ← Step 2: brand kit setup
│       │   ├── (dashboard)/             ← Route group — authenticated shell
│       │   │   ├── layout.tsx
│       │   │   ├── page.tsx             ← Redirect → /briefs
│       │   │   ├── briefs/
│       │   │   │   ├── page.tsx         ← Brief list (S-01)
│       │   │   │   └── [id]/
│       │   │   │       ├── page.tsx     ← Brief detail + angles (S-03/04)
│       │   │   │       └── generate/page.tsx ← Generation settings (S-06)
│       │   │   ├── assets/page.tsx      ← Asset library (S-07)
│       │   │   ├── exports/page.tsx     ← Exports list (S-09)
│       │   │   ├── projects/page.tsx    ← Projects list
│       │   │   ├── brand/page.tsx       ← Brand kit (S-10)
│       │   │   └── settings/page.tsx    ← Org + billing (S-11)
│       │   ├── api/
│       │   │   ├── trpc/[trpc]/route.ts          ← tRPC HTTP handler
│       │   │   └── generation/[jobId]/
│       │   │       └── stream/route.ts            ← SSE stream (NOT tRPC)
│       │   ├── globals.css              ← @theme {} Tailwind v4 tokens only
│       │   ├── error.tsx
│       │   └── not-found.tsx
│       ├── _components/                 ← Private components (not routes)
│       ├── _lib/                        ← Private utilities (not routes)
│       ├── trpc/
│       │   ├── client.ts
│       │   ├── server.ts
│       │   └── routers/
│       │       ├── index.ts             ← appRouter (merges all routers)
│       │       ├── briefs.ts            ← create, list, get, patch, delete, regenerate
│       │       ├── assets.ts            ← generate, list, get, patch, retry, delete
│       │       ├── exports.ts           ← create, get, list, delete
│       │       ├── projects.ts          ← CRUD
│       │       ├── brands.ts            ← CRUD
│       │       ├── jobs.ts              ← get, list (job status polling pre-SSE connect)
│       │       └── org.ts               ← profile, members, billing
│       ├── stores/                      ← Zustand stores
│       ├── middleware.ts                ← clerkMiddleware()
│       ├── next.config.ts
│       ├── package.json
│       └── tsconfig.json
│
│   ├── packages/
│   │   ├── ui/                          ← shadcn/ui components (copied, not imported)
│   │   ├── src/components/
│   │   ├── package.json
│   │   └── tsconfig.json
│   │   ├── types/                       ← Shared TypeScript types
│   │   ├── src/
│   │   │   ├── api.ts                   ← API request/response types
│   │   │   ├── generation.ts            ← Job, variant, angle, brief types
│   │   │   └── index.ts
│   │   ├── package.json
│   │   └── tsconfig.json
│   │   └── config/                      ← Shared tooling configs
│       ├── biome/base.json              ← Shared Biome rules (extends from root biome.json)
│       ├── typescript/base.json
│       └── package.json
│
│   ├── services/
│   │   ├── go.work                      ← Go workspace (Go 1.18+) — links api + worker modules locally
│   │   ├── api/                         ← Go Echo v4 (→ Railway)
│   │   ├── main.go                      ← Entry point
│   │   ├── internal/
│   │   │   ├── handler/                 ← HTTP handlers (Echo routes)
│   │   │   ├── middleware/              ← JWT verify + tier limit enforcement
│   │   │   ├── db/                      ← sqlc generated code (querier + models)
│   │   │   ├── service/                 ← Business logic
│   │   │   └── config/                  ← Env / configuration
│   │   ├── db/
│   │   │   └── queries/                 ← sqlc query definitions (.sql files → Go codegen)
│   │   ├── sqlc.yaml
│   │   ├── go.mod
│   │   ├── Dockerfile                   ← Multi-stage: golang:alpine → scratch
│   │   └── .dockerignore
│   │
│   ├── worker/                          ← Go asynq (→ Railway)
│   │   ├── main.go
│   │   ├── internal/
│   │   │   ├── tasks/                   ← Task definitions (generate_video.go, etc.)
│   │   │   ├── processor/               ← FAL poll, ElevenLabs, HeyGen handlers
│   │   │   └── config/
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── .dockerignore
│   │
│   └── postprocess/                     ← Rust Axum + ffmpeg-sys (→ Railway)
│       ├── src/
│       │   ├── main.rs                  ← Entry point + router setup
│       │   ├── handlers/
│       │   │   ├── mod.rs
│       │   │   ├── transcode.rs
│       │   │   ├── watermark.rs
│       │   │   └── caption.rs
│       │   ├── ffmpeg/
│       │   │   ├── mod.rs
│       │   │   └── command.rs           ← ffmpeg-sys bindings
│       │   ├── error.rs                 ← Custom error types + IntoResponse
│       │   └── health.rs
│       ├── Cargo.toml
│       ├── Dockerfile                   ← Multi-stage: rust:alpine → alpine+ffmpeg
│       └── .dockerignore
│
├── supabase/                            ← Supabase CLI project root
│   ├── migrations/                      ← SQL migration files (supabase db push)
│   │   └── 20260414000000_init.sql      ← Initial schema (full DDL from Database-Schema.md)
│   ├── seed.sql                         ← Dev seed data
│   └── config.toml                      ← Supabase project config
│
├── ai/                                  ← AI layer: prompts + evals (Langfuse-versioned)
│   ├── prompts/
│   │   ├── brief-parse.prompt.ts        ← GPT-4o product extraction prompt
│   │   ├── angles-gen.prompt.ts         ← Claude 4.6 creative angles prompt
│   │   └── hooks-gen.prompt.ts          ← Claude 4.6 hook variants prompt
│   └── evals/
│       ├── brief-quality.eval.ts        ← Output quality test suite (20-URL set)
│       └── angle-diversity.eval.ts      ← Angle variation scoring
│
├── docs/
│   ├── 01-brand/                        ← Brand identity
│   ├── 02-product/                      ← Product definition + overview
│   ├── 03-market/                       ← Competitive analysis
│   ├── 04-specs/                        ← Feature spec, user stories, user journey
│   ├── 05-design/                       ← Design system, wireframes, UI spec
│   ├── 06-technical/                    ← Architecture, DB schema, API design, sprint plan, repo structure
│   └── 07-implementation/               ← Phases tracker, task checklist, SDK references (3 docs)
│
├── .env.example                         ← All env vars documented, no secrets
├── .gitignore
├── .gitattributes                       ← text=auto eol=lf (LF enforced cross-OS)
├── .nvmrc                               ← Node 22.x LTS (required for TS ≥5.7.2)
├── .editorconfig                        ← 2-space indent, LF, trim whitespace
├── biome.json                           ← Format + lint (replaces Prettier + ESLint)
├── lefthook.yml                         ← Pre-commit hooks (replaces Husky)
├── docker-compose.yml                   ← Local dev: Postgres, Redis ×2, all services
├── turbo.json                           ← Pipeline: build → lint → test
├── package.json                         ← workspaces: [src/apps/*, src/packages/*]
├── tsconfig.json                        ← Root TS config (references)
│   ├── AGENTS.md                        ← Agentic AI tool context
│   ├── CONTEXT.md                       ← Quick-reference product + stack
│   ├── MEMORY.md                        ← Decision log
└── README.md
```

---

## CI/CD Pipeline

### Workflow Map

| Workflow | Trigger (path filter) | Action | Target |
|---|---|---|---|
| `ci.yml` | All PRs | Turbo lint + typecheck + test | — |
| `deploy-web.yml` | Push to `main` → `src/apps/web/**` or `src/packages/**` | `turbo build --filter=web` | Vercel |
| `deploy-api.yml` | Push to `main` → `src/services/api/**` | Docker build + push | Railway |
| `deploy-worker.yml` | Push to `main` → `src/services/worker/**` | Docker build + push | Railway |
| `deploy-postprocess.yml` | Push to `main` → `src/services/postprocess/**` | Docker build + push | Railway |
| `deploy-db.yml` | Push to `main` → `supabase/migrations/**` | `supabase db push` → staging | Supabase |
| `security.yml` | Weekly schedule + PRs | CodeQL scan | — |

### Deploy Flow

```
PR opened
    ↓
ci.yml runs: turbo lint + typecheck + test (Turborepo cache-aware)
    ↓ (approved + merged to main)
Path change detected by GitHub Actions:
    src/apps/web/** or src/packages/** → deploy-web.yml     →  Vercel
    src/services/api/**                → deploy-api.yml     →  Railway (Docker)
    src/services/worker/**             → deploy-worker.yml  →  Railway (Docker)
    src/services/postprocess/**        → deploy-postprocess →  Railway (Docker)
    supabase/migrations/**            →  deploy-db.yml      →  Supabase (supabase db push)
```

All deploy workflows use:
```yaml
concurrency:
  group: deploy-${{ github.ref }}-${{ github.workflow }}
  cancel-in-progress: true
```

---

## Dockerfile Strategy

### Go Services (api + worker)

```
Build stage:  golang:1.24-alpine   ← compiler + go mod download + build
Final stage:  scratch              ← binary only (~10–20 MB image)
```

- `CGO_ENABLED=0` for static binary
- `GOOS=linux` for cross-compilation on Mac/Windows dev machines
- No shell, no OS, no attack surface in final image

### Rust Service (postprocess)

```
Build stage:  rust:1.87-alpine     ← compiler + ffmpeg-dev + cargo build
Final stage:  alpine:latest        ← ffmpeg runtime + compiled binary (~150–250 MB)
```

- Dependency caching trick: build with dummy `main.rs` first to cache `cargo build` layer
- Alpine final stage (not scratch) because `ffmpeg` requires shared libs at runtime

---

## Tooling Decisions

| Tool | Choice | Replaces | Reason |
|---|---|---|---|
| Monorepo orchestration | Turborepo 2.x | nx, Lerna | Native Vercel integration, fastest cache |
| Linter + formatter | Biome | ESLint + Prettier | 25× faster, single config, same output |
| Git hooks | Lefthook | Husky | No npm dependency, native binary |
| Package manager | pnpm | npm, yarn | Faster installs, strict hoisting |
| Node version | `.nvmrc` (Node 22 LTS) | `.node-version` | Widest tool support |
| Go multi-module | `src/services/go.work` | `replace` directives | Local module linking without version pinning |
| DB migrations | Supabase CLI (`supabase/`) | raw psql | `supabase db push`, `supabase gen types` integration |
| Prompt versioning | `ai/prompts/` + Langfuse | hardcoded strings | Version-controlled source → Langfuse synced at deploy |

---

## Language Boundaries (Hard Rules)

| Service | Language | Why |
|---|---|---|
| `src/apps/web` | TypeScript (Next.js 15) | I/O-bound, React ecosystem |
| `src/services/api` | Go (Echo v4) | I/O-bound, low-latency HTTP + external API calls |
| `src/services/worker` | Go (asynq) | Goroutine-per-job concurrency model |
| `src/services/postprocess` | Rust (Axum + ffmpeg-sys) | CPU-bound video processing |

> **Rule:** Do not expand Rust beyond `src/services/postprocess/`. Do not add a fourth language.

---

## Redis — Two Instances, Never Substitutable

| Instance | Provider | Protocol | Used by |
|---|---|---|---|
| `UPSTASH_REDIS_REST_URL` | Upstash | HTTP/REST | Cache, rate-limiting, session store |
| `RAILWAY_REDIS_URL` | Railway | TCP | asynq job queues (`BLPOP` requires persistent TCP) |

Upstash HTTP proxy does not support `BLPOP`/`BRPOP`. Using Upstash for asynq silently fails in production.

---

## Key Constraints Enforced by This Structure

- **No `tailwind.config.ts`** — Tailwind v4 is CSS-only. All tokens live in `globals.css` `@theme {}`.
- **SSE is not tRPC** — `src/apps/web/app/api/generation/[jobId]/stream/route.ts` is a standalone Route Handler.
- **Tier limits are server-side** — Enforced in `src/services/api/internal/middleware/`, never in `src/apps/web`.
- **DTC features are Phase 2** — Nothing in `(dashboard)/` is built for DTC Brand Managers in V1.
- **HeyGen = v3 only** — Active platform: `developers.heygen.com`. Any reference to "v4" is incorrect. V2V lip-sync is v3-only.
- **FAL.AI = async queue only** — Always `fal.queue.submit()`. Never `fal.subscribe()` (blocks; unusable for 30–120s operations).
- **Migrations = `supabase/migrations/` only** — Do not add migration files to `src/services/api/db/`. That directory holds sqlc query definitions only.
