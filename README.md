# Qvora

Born to Convert.

Qvora is an AI-powered performance creative SaaS that turns a product URL into short-form video ads (9:16), then improves future variants using performance signals.

## One-liner
Paste a URL. Get 10 video ads. Know which ones win - automatically.

## Status
Planning complete. Implementation in progress.

- Product definition, specs, and architecture docs are available.
- Monorepo and phased implementation plan are defined.
- V1 focuses on agency users (media buyers and creative directors).

## Repository Visibility
This is a private internal repository.

- Internal planning, architecture, and implementation details are confidential.
- Do not publish code, docs, or architecture diagrams externally without explicit approval.
- Use private issue and PR workflows for all project discussions.

## Core Workflow
1. Product URL ingestion (Playwright extraction)
2. AI creative brief generation (3-5 angles)
3. Video variant generation (T2V/I2V/V2V)
4. Voiceover + captions + post-processing
5. Export platform-ready assets
6. V2: performance signal loop for better next variants

## V1 ICP
- Agency Media Buyer (primary)
- Agency Creative Director (primary)
- Agency Account Manager (reviewer only)

DTC Brand Manager is Phase 2 only.

## Pricing (Planned)
- Starter: $99/month (max 3 variants per angle)
- Growth: $149/month (max 10 variants per angle)
- Agency: $399/month (unlimited)

## Architecture Snapshot
- Frontend: Next.js 15 App Router, shadcn/ui, Tailwind CSS v4
- BFF: tRPC
- API: Go Echo v4
- Workers: Go + asynq
- Video post-processing: Rust Axum + ffmpeg-sys
- AI layer: Vercel AI SDK v6, FAL.AI, ElevenLabs, HeyGen v3
- Auth: Clerk Organizations
- Data: Supabase Postgres + RLS, sqlc
- Redis (mandatory split): Upstash (HTTP cache) + Railway Redis (TCP queue)
- Storage/streaming: Cloudflare R2 + Mux
- Billing: Stripe subscriptions + entitlements
- Infra: Vercel, Railway, Modal, Doppler, GitHub Actions

## Non-Negotiable Rules
1. Upstash Redis is for cache/rate limits only. Railway Redis is for asynq only.
2. Generation SSE stream is a standalone Route Handler, not a tRPC subscription.
3. Tailwind v4 uses CSS-only theme config in globals.css. No tailwind.config.ts.
4. HeyGen API version is v3.
5. V1 is agency-first.
6. Go handles I/O-bound services. Rust is only for CPU-bound video post-processing.

## Repository Layout (Planned)
```text
apps/
  web/            -> Next.js app + tRPC BFF
packages/
  ui/             -> shared UI components
  types/          -> shared TypeScript types
  config/         -> shared lint/TS configs
src/
  apps/
    web/          -> Next.js app
  packages/
    config/       -> Shared tooling config
    types/        -> Shared TypeScript types
    ui/           -> Shared UI components
  services/
    api/          -> Go API
    worker/       -> Go asynq workers
    postprocess/  -> Rust video service
```

## Key Documents
- [Project context](.github/CONTEXT.md)
- [Decision memory](.github/MEMORY.md)
- [Agent instructions](.github/AGENTS.md)
- [Product definition](docs/02-product/Qvora_Product-Definition.md)
- [Feature specification](docs/04-specs/Qvora_Feature-Spec.md)
- [User stories](docs/04-specs/Qvora_User-Stories.md)
- [Architecture stack](docs/05-technical/Qvora_Architecture-Stack.md)
- [Repository structure](docs/05-technical/Qvora_Repo-Structure.md)
- [Implementation references](docs/06-implementation/Qvora_Implementation-References.md)
- [Implementation plan](docs/06-implementation/Qvora_Implementation-Plan.md)

## Getting Started
Implementation scaffolding is being built. Until code bootstrap lands, use docs-first workflow:

1. Read [Project context](.github/CONTEXT.md)
2. Read [Architecture stack](docs/05-technical/Qvora_Architecture-Stack.md)
3. Follow [Implementation plan](docs/06-implementation/Qvora_Implementation-Plan.md)

## License
Proprietary. All rights reserved.
