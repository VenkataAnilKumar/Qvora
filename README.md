# Qvora

## Born to Convert.

Qvora is an AI-powered performance creative SaaS that transforms a single product URL into short-form ad variants (9:16), then improves the next generation using performance signals.

Paste a URL. Get 10 video ads. Know which ones win - automatically.

## Why Teams Share Qvora

- Faster creative cycles from idea to launch-ready variants.
- Performance-first angle generation instead of generic ad copy.
- Built for agency workflows where speed, consistency, and output quality all matter.

## Product Snapshot (2026)

- Primary ICP (V1): Agency Media Buyer, Agency Creative Director
- Reviewer-only role (V1): Agency Account Manager
- Phase 2 only: DTC Brand Manager

### Pricing (Planned)

| Tier | Price | Variant Limit |
|---|---:|---|
| Starter | $99/mo | Max 3 variants per angle |
| Growth | $149/mo | Max 10 variants per angle |
| Agency | $399/mo | Unlimited |

## Current Build Status

Implementation is active and tracked in phase gates.

- Phase 0: Complete
- Phase 1: Complete
- Phase 2: Complete
- Phase 3: Complete
- Phase 4: Partial
- Phase 5-7: Pending validation

For the exact live status, see [Implementation checklist](docs/07-implementation/IMPLEMENTATION_CHECKLIST.md).

## Core Workflow

1. URL ingestion and product extraction (Playwright)
2. AI creative brief generation (3-5 angles + hooks)
3. Variant generation pipeline (T2V/I2V/V2V)
4. Voiceover, captions, and post-processing
5. Streaming playback and export-ready assets
6. V2 signal loop for better next variants

## Architecture at a Glance

- Frontend: Next.js 15 App Router, shadcn/ui, Tailwind CSS v4
- BFF: tRPC
- API: Go Echo v4
- Workers: Go + asynq
- Post-processing: Rust Axum + ffmpeg-sys
- AI stack: Vercel AI SDK v6, FAL.AI, ElevenLabs, HeyGen v3
- Data: Supabase Postgres + RLS + sqlc
- Auth: Clerk Organizations
- Caching/Queues: Upstash Redis (HTTP) + Railway Redis (TCP)
- Storage/Delivery: Cloudflare R2 + Mux
- Infra: Vercel, Railway, Modal, Doppler, GitHub Actions

## Non-Negotiables

1. Redis split is mandatory: Upstash for HTTP cache/rate limits, Railway Redis for asynq TCP queues.
2. Generation stream is SSE via standalone Route Handler, not a tRPC subscription.
3. Tailwind v4 is CSS-only theme config in globals.css (no tailwind.config.ts).
4. HeyGen integration is v3 only.
5. V1 remains agency-first.
6. Go is for I/O-bound services, Rust is limited to CPU-bound video post-processing.

## Monorepo Layout

```text
src/
  apps/
    web/          -> Next.js app + tRPC BFF
  packages/
    config/       -> Shared tooling configs
    types/        -> Shared TypeScript types
    ui/           -> Shared UI components
  services/
    api/          -> Go API
    worker/       -> Go asynq workers
    postprocess/  -> Rust video processor
```

## Quick Start

```bash
npm install
npm run dev
```

Useful scripts:

- `npm run typecheck`
- `npm run lint`
- `npm run build`

## Documentation Map

- [Project context](.github/CONTEXT.md)
- [Decision memory](.github/MEMORY.md)
- [Copilot instructions](.github/copilot-instructions.md)
- [Product definition](docs/02-product/Qvora_Product-Definition.md)
- [Feature specification](docs/04-specs/Qvora_Feature-Spec.md)
- [User stories](docs/04-specs/Qvora_User-Stories.md)
- [Architecture stack](docs/06-technical/Qvora_Architecture-Stack.md)
- [System architecture](docs/06-technical/Qvora_System-Architecture.md)
- [Implementation phases](docs/07-implementation/Qvora_Implementation-Phases.md)
- [Implementation checklist](docs/07-implementation/IMPLEMENTATION_CHECKLIST.md)
- [Implementation references](docs/07-implementation/Qvora_Implementation-References.md)

## Repository Visibility

This repository is private and internal.

- Planning, architecture, and implementation details are confidential.
- Do not publish code, docs, or architecture diagrams externally without explicit approval.
- Use private issue and PR workflows for all project discussions.

## License

Proprietary. All rights reserved.
