---
title: Repository Structure
category: architecture
tags: [monorepo, turborepo, repo-structure, file-layout, routing]
sources: [Qvora_Repo-Structure]
updated: 2026-04-15
---

# Repository Structure

## TL;DR
Turborepo monorepo. All services in one repo with path-filtered CI/CD for independent deploys. `src/packages/types` is consumed everywhere — this is the primary reason for monorepo over polyrepo.

---

## Decision: Monorepo (Turborepo)

**Why monorepo over polyrepo:**

1. **Shared TypeScript types are required.** `src/packages/types` consumed by web + tRPC routers. Polyrepo would need versioned npm publishes on every change.
2. **Cross-service changes ship atomically.** A generation feature touches web → api → worker → postprocess. One PR, one review.
3. **Turborepo path filtering gives polyrepo-level deploy isolation.** Each service deploys independently via GitHub Actions path filters.
4. **Go + Rust coexist cleanly.** `src/services/` is independent of npm workspaces. No toolchain conflicts.
5. **Stage of company.** Polyrepo overhead only pays off with dedicated per-service teams.

**When to revisit (not V1):** Series A+ with dedicated teams per service, or security requirements to restrict repo access per service.

---

## Top-Level Structure

```
qvora/
├── .github/
│   ├── copilot-instructions.md       ← Auto-loaded by GitHub Copilot (THIS FILE)
│   ├── workflows/
│   │   ├── ci.yml                    ← Lint + typecheck + test on every PR
│   │   ├── deploy-web.yml            ← Trigger: src/apps/web/**, src/packages/**
│   │   ├── deploy-api.yml            ← Trigger: src/services/api/**
│   │   ├── deploy-worker.yml         ← Trigger: src/services/worker/**
│   │   ├── deploy-postprocess.yml    ← Trigger: src/services/postprocess/**
│   │   └── deploy-db.yml             ← Trigger: supabase/migrations/** → supabase db push
│   ├── CODEOWNERS
│   └── pull_request_template.md
│
├── src/                              ← All source code
│   ├── apps/
│   │   └── web/                      ← Next.js 15 App Router → Vercel
│   ├── packages/
│   │   ├── types/                    ← Shared TypeScript types
│   │   ├── ui/                       ← shadcn/ui component copies
│   │   └── config/                   ← Biome + TS base configs
│   └── services/
│       ├── api/                      ← Go Echo v4 → Railway
│       ├── worker/                   ← Go asynq → Railway
│       └── postprocess/              ← Rust Axum + ffmpeg-sys → Railway
│
├── supabase/
│   └── migrations/
│       ├── 001_initial_schema.sql
│       ├── 002_postprocess_callbacks.sql
│       └── 003_mux_webhook_events_and_reconcile.sql
│
├── wiki/                             ← LLM wiki (this system)
├── docs/                             ← Source documents ingested into wiki
├── turbo.json
├── biome.json
├── lefthook.yml
└── package.json (pnpm workspaces)
```

---

## `src/apps/web` — Next.js App Router Layout

```
src/apps/web/
├── app/
│   ├── layout.tsx                    ← Root layout + ClerkProvider
│   ├── page.tsx                      ← Marketing / home
│   ├── globals.css                   ← @theme {} Tailwind v4 tokens ONLY
│   ├── (auth)/                       ← Route group (no URL segment)
│   │   ├── sign-in/[[...sign-in]]/
│   │   └── sign-up/[[...sign-up]]/
│   ├── (onboarding)/                 ← Post-signup wizard
│   │   ├── page.tsx                  ← Step 1: org + workspace
│   │   └── brand/page.tsx            ← Step 2: brand kit
│   ├── (dashboard)/                  ← Authenticated shell
│   │   ├── layout.tsx
│   │   ├── briefs/
│   │   │   ├── page.tsx              ← Brief list (S-01)
│   │   │   └── [id]/
│   │   │       ├── page.tsx          ← Brief detail + angles (S-03/04)
│   │   │       └── generate/page.tsx ← Generation settings (S-06)
│   │   ├── assets/page.tsx           ← Asset library (S-07)
│   │   ├── exports/page.tsx          ← Exports list (S-09)
│   │   ├── brand/page.tsx            ← Brand kit (S-10)
│   │   └── settings/page.tsx         ← Org + billing (S-11)
│   └── api/
│       ├── trpc/[trpc]/route.ts      ← tRPC HTTP handler
│       └── generation/[jobId]/stream/route.ts  ← SSE stream (NOT tRPC)
├── middleware.ts                     ← clerkMiddleware() export
└── next.config.ts
```

---

## `src/services/api` — Go Echo v4

```
src/services/api/
├── cmd/api/main.go
├── internal/
│   ├── handler/          ← Echo route handlers
│   ├── middleware/        ← Clerk JWT, rate-limit, tier enforcement
│   └── db/               ← sqlc-generated Go types
├── db/
│   ├── queries/          ← .sql files → sqlc input
│   └── migrations/       ← NOT used (supabase/migrations/ is source of truth)
├── go.mod
└── sqlc.yaml
```

---

## `src/services/worker` — Go asynq

```
src/services/worker/
├── cmd/worker/main.go
└── internal/task/        ← Task handlers: brief:extract, generation:video, etc.
```

---

## `src/services/postprocess` — Rust Axum

```
src/services/postprocess/
├── src/
│   ├── main.rs
│   ├── handler.rs       ← Axum route handlers
│   ├── processor.rs     ← ffmpeg-sys operations
│   ├── model.rs
│   ├── mux.rs
│   └── error.rs
└── Cargo.toml
```

**Scope is locked to:** watermark, captions, transcode, reframe. Do not expand Rust beyond CPU-bound video processing.

---

## `src/packages/types` — Shared TypeScript Types

Consumed by `src/apps/web` and tRPC routers. Primary reason for monorepo. Contains:
- Generation job status enums
- Brief, angle, hook, asset type definitions
- Stripe tier enum
- Zod schemas shared between frontend and BFF

---

## Open Questions
- [ ] Is there a `src/ai/` directory for Vercel AI SDK prompt templates separate from `src/apps/web`?
- [ ] Where does the Playwright scraping Modal code live? (Not yet in repo structure)

## Related Pages
- [[system-architecture]] — runtime topology for these services
- [[stack-overview]] — tooling for each directory
- [[sprint-plan]] — which pieces were built in which sprint
