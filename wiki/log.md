## [2026-04-15] bootstrap | Initial wiki creation from docs/

**Operation:** Bootstrap ingest  
**Sources processed:** 8 docs (Product-Definition, Brand-Identity, Competitive-Analysis, Feature-Spec, Architecture-Stack, Database-Schema, Design-System, Implementation-Phases)  
**Pages created:** 14  
**Pages updated:** 0 (initial ingest)

### What was created
- `product/qvora-overview.md` — product one-liner, core flow, strategic objective
- `product/features.md` — all feature modules, V1/V2 scope
- `product/personas.md` — ICP, three personas (Media Buyer, Creative Director, Account Manager)
- `product/pricing.md` — Starter/Growth/Agency tiers, trial logic, entitlements
- `product/brand.md` — brand identity, tokens, fonts, taglines
- `product/roadmap.md` — phases, current implementation status
- `market/market-context.md` — CPM data, creative velocity problem
- `market/competitive-landscape.md` — comparison table, Qvora positioning
- `market/creatify.md` — deep-dive on primary competitor
- `market/arcads.md` — Arcads profile
- `market/heygen.md` — HeyGen profile (competitor + Qvora vendor)
- `architecture/stack-overview.md` — full stack, non-negotiables
- `architecture/ai-layer.md` — AI model choices and rationale
- `architecture/data-layer.md` — data stores, Redis duality, Mux

### Open questions flagged
- What is Arcads pricing? (not public as of April 2026)
- What is the TAM for mid-market performance creative SaaS?
- Personal section is empty — user to seed with first entry

---

## [2026-04-15] ingest | Full docs/ ingest — all remaining sources

**Operation:** Complete ingest of all remaining `docs/` source files  
**Sources processed:** User-Stories, User-Journey, Design-System, UI-Spec, API-Design, System-Architecture, Repo-Structure, Database-Schema, Sprint-Plan, Implementation-References, IMPLEMENTATION_CHECKLIST, Product-Overview  
**Pages created:** 11  
**Pages updated:** 0

### What was created

- `product/user-stories.md` — 8 epics, story IDs, acceptance criteria summaries; EPIC 7 (Signal loop) = V2 only; P3 = reviewer-only, no generation stories
- `product/user-journey.md` — 7-phase journey map; Phase 0 acquisition channels; Phase 2 activation = first brief in < 60s; design principle: "every screen must earn its place"
- `design/design-system.md` — Full CSS variable mapping, dark-first color philosophy, Tailwind v4 `@theme {}` rule, type scale, 4px spacing, 150–250ms motion
- `design/ui-components.md` — 25-component spec (C-01 to C-25): Button (6 variants + Volt glow), Badge, topbar, Mux player, generation status card (SSE-driven)
- `architecture/api-design.md` — REST routes, tRPC procedure list, SSE Route Handler spec (NOT tRPC), webhook endpoint, JWT payload structure, rate limits
- `architecture/system-architecture.md` — 5-layer component diagram, 3 workload profiles, sequence diagrams (brief gen, video gen), SSE stream mechanism, infra map
- `architecture/repo-structure.md` — Turborepo rationale (5 reasons), full directory tree, App Router route groups `(auth)/(onboarding)/(dashboard)`, Rust scope locked
- `architecture/database-schema.md` — ERD overview, 5 design principles, 9+ tables, RLS pattern code, index strategy, V2 anchor tables (video_variants, performance_signals)
- `architecture/sprint-plan.md` — Sprint 0–4 (11 weeks), MVP criteria, build order rationale, exit gates, post-MVP V2 backlog
- `architecture/implementation-checklist.md` — Phase 0–7 status table; confirmed: `plan_tier = starter/growth/agency` (no 'scale'), `jobs.status` has no 'briefing' value
- `architecture/implementation-references.md` — SDK docs, install commands, API URLs, critical gotchas for every tool in the stack

### New directory created
- `wiki/design/` — was not in original scaffold; auto-created with design-system.md

### Key discrepancies / findings
- **`plan` tier discrepancy resolved:** Schema doc said `trial|starter|growth|scale|enterprise` — superseded by Implementation Checklist which confirms `starter|growth|agency` (3 tiers only). Implementation Checklist is authoritative.
- **`jobs.status` clarified:** Does NOT include 'briefing' — values are `queued, scraping, generating, postprocessing, complete, failed`

---

