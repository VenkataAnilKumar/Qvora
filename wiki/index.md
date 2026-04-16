# Qvora Wiki — Index

> Content-oriented catalog of all wiki pages. Updated on every ingest. The LLM reads this first when answering queries.
> 
> Last updated: 2026-04-15 | Full ingest — all docs/ sources

---

## Product

| Page | Summary | Updated |
|---|---|---|
| [qvora-overview](product/qvora-overview.md) | Product one-liner, core flow (URL→brief→video→export), strategic objective, why now | 2026-04-15 |
| [features](product/features.md) | All 14 feature modules, V1/V2 scope, key acceptance criteria per module | 2026-04-15 |
| [personas](product/personas.md) | ICP: Agency Media Buyer (P1), Creative Director (P2), Account Manager (P3 read-only), DTC (P4 Phase 2) | 2026-04-15 |
| [pricing](product/pricing.md) | Starter $99 / Growth $149 / Agency $399; trial logic; Stripe webhook handling; entitlements | 2026-04-15 |
| [brand](product/brand.md) | Brand tokens (Volt #7B2FFF, Green #00E87A, etc.), fonts (Clash Display, Space Grotesk, Inter, JetBrains Mono), taglines | 2026-04-15 |
| [roadmap](product/roadmap.md) | Phase 0–5 status: Phase 0+1+3 complete, Phase 2 partial (BRIEF-08/09 in progress), Phase 4+ pending | 2026-04-15 |
| [user-stories](product/user-stories.md) | 8 epics (EPIC 1–8); key stories, acceptance criteria; EPIC 7 (Signal) V2 only | 2026-04-15 |
| [user-journey](product/user-journey.md) | 7 journey phases from awareness to Signal; activation moment = first brief in 60s; screen-level flow | 2026-04-15 |

---

## Market

| Page | Summary | Updated |
|---|---|---|
| [market-context](market/market-context.md) | CPM trends (Instagram $9.46, TikTok $4–9), creative velocity gap, PureGym 5.6x case study, structural tailwinds | 2026-04-15 |
| [competitive-landscape](market/competitive-landscape.md) | Full comparison table; Qvora's differentiated position (brief engine + closed loop) | 2026-04-15 |
| [creatify](market/creatify.md) | Primary competitor: $650M ad spend, 15K brands, $24M raised; no strategy layer = Qvora's wedge | 2026-04-15 |
| [arcads](market/arcads.md) | Realistic AI actors for UGC; actor-centric, not system-centric; pricing not public | 2026-04-15 |
| [heygen](market/heygen.md) | Avatar quality leader; Qvora uses HeyGen as vendor (Avatar API v3); competitor for explainer video market | 2026-04-15 |
| [adcreative](market/adcreative.md) | Image-first ad tool; minimal video; low overlap with Qvora | 2026-04-15 |

---

## Architecture

| Page | Summary | Updated |
|---|---|---|
| [stack-overview](architecture/stack-overview.md) | Full stack by layer + 7 non-negotiable rules; Vercel/Railway/Modal/Doppler deployment | 2026-04-15 |
| [ai-layer](architecture/ai-layer.md) | GPT-4o brief, Claude regen, FAL.AI T2V queue, ElevenLabs TTS, HeyGen v3 lip-sync, Langfuse obs | 2026-04-15 |
| [data-layer](architecture/data-layer.md) | PostgreSQL+sqlc, two Redis instances (Upstash HTTP ≠ Railway TCP), R2 storage, Mux HLS | 2026-04-15 |
| [system-architecture](architecture/system-architecture.md) | Component diagram, workload profiles, key sequence flows (brief gen + video gen + SSE) | 2026-04-15 |
| [api-design](architecture/api-design.md) | REST endpoints, tRPC procedures, SSE stream, JWT auth, error codes, rate limits | 2026-04-15 |
| [database-schema](architecture/database-schema.md) | Full schema DDL overview, RLS pattern, key tables, V2 anchor tables | 2026-04-15 |
| [repo-structure](architecture/repo-structure.md) | Monorepo layout (Turborepo), App Router file structure, Go/Rust service dirs, decision rationale | 2026-04-15 |
| [sprint-plan](architecture/sprint-plan.md) | Sprint 0–4 (11 weeks), MVP definition, build order rationale, V2 backlog | 2026-04-15 |
| [implementation-checklist](architecture/implementation-checklist.md) | Phase completion status (Phases 0–7), confirmed constraints, items to validate | 2026-04-16 |
| [implementation-references](architecture/implementation-references.md) | SDK docs, install commands, API URLs, Qvora-specific usage notes for every tool | 2026-04-15 |

---

## Design

| Page | Summary | Updated |
|---|---|---|
| [design-system](design/design-system.md) | Dark-first design philosophy, color tokens (shadcn mapping), typography scale, spacing, motion | 2026-04-15 |
| [ui-components](design/ui-components.md) | 25 components (C-01 to C-25), variants, states, Volt glow, Mux player, SSE generation card | 2026-04-15 |

---

## Personal

| Page | Summary | Updated |
|---|---|---|
| [README](personal/README.md) | Instructions for using the personal section; suggested files to create | 2026-04-15 |

---

## Synthesis

*(Empty — populated as you ask questions and file answers)*

---

## Source Documents Ingested

| Source | Domain | Ingested |
|---|---|---|
| `docs/02-product/Qvora_Product-Definition.md` | Product | 2026-04-15 |
| `docs/01-brand/Qvora_Brand-Identity.md` | Product | 2026-04-15 |
| `docs/04-specs/Qvora_Feature-Spec.md` | Product | 2026-04-15 |
| `docs/07-implementation/Qvora_Implementation-Phases.md` | Product | 2026-04-15 |
| `docs/03-market/Qvora_Competitive-Analysis.md` | Market | 2026-04-15 |
| `docs/06-technical/Qvora_Architecture-Stack.md` | Architecture | 2026-04-15 |
| `docs/06-technical/Qvora_Database-Schema.md` | Architecture | 2026-04-15 |
| `docs/06-technical/Qvora_API-Design.md` | Architecture | 2026-04-15 |
| `docs/06-technical/Qvora_System-Architecture.md` | Architecture | 2026-04-15 |
| `docs/06-technical/Qvora_Repo-Structure.md` | Architecture | 2026-04-15 |
| `docs/06-technical/Qvora_Sprint-Plan.md` | Architecture | 2026-04-15 |
| `docs/07-implementation/IMPLEMENTATION_CHECKLIST.md` | Architecture | 2026-04-16 |
| `docs/07-implementation/Qvora_Implementation-References.md` | Architecture | 2026-04-15 |
| `docs/04-specs/Qvora_User-Stories.md` | Product | 2026-04-15 |
| `docs/04-specs/Qvora_User-Journey.md` | Product | 2026-04-15 |
| `docs/05-design/Qvora_Design-System.md` | Design | 2026-04-15 |
| `docs/05-design/Qvora_UI-Spec.md` | Design | 2026-04-15 |
| `docs/02-product/Qvora_Product-Overview.md` | Product | 2026-04-15 |
| `.github/copilot-instructions.md` | Architecture (non-negotiables) | 2026-04-15 |

---

## Not Yet Ingested

| Source | Notes |
|---|---|
| `docs/05-design/Qvora_Wireframes.md` | Wireframes reference; ingest when working on screen-by-screen layouts |
