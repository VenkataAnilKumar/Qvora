# Qvora — Long Memory

> Persistent knowledge base for AI agents and developers working on Qvora. Covers all decisions made, patterns locked, issues resolved, and conventions enforced across all docs and design work.

---

## Product Decisions (Locked)

| Decision | Rationale | Locked In |
|---|---|---|
| Agency-first V1 ICP (not DTC) | DTC has lower willingness to pay; agencies have 10x creator leverage and proven tooling budgets | Product-Definition, User-Stories |
| 7-day free trial (no credit card) | Reduces friction for agency evaluation; benchmark from UserPilot 2026 (no-card trials convert 2x better) | Feature-Spec PLAT-08 |
| Performance feedback loop is V2 | Reduces V1 complexity; agency buyers accept V1 purely on creative production speed | Feature-Spec, User-Stories |
| Ad account connector (CONN) is V2 | Requires OAuth per-platform; not MVP blocker — export alone validates core value | Feature-Spec |
| Variant limits enforced at Go API middleware level | Tier limits are business logic, not UI logic — enforced server-side, not client-side | Architecture-Stack |
| HeyGen V2V lip-sync promoted to V1 | Competitive differentiator vs Creatify; V2V adds UGC-style avatar ads without real actors | Feature-Spec GEN-07 |

---

## Architecture Decisions (Locked)

### Two Redis Instances

**Decision:** Two separate Redis deployments are mandatory.

| Instance | Provider | Protocol | Purpose |
|---|---|---|---|
| `UPSTASH_REDIS_REST_URL` | Upstash | HTTP/REST | Cache, rate-limiting, session tokens |
| `RAILWAY_REDIS_URL` | Railway | TCP persistent | asynq job queues (`BLPOP`/`BRPOP`) |

**Why:** `asynq` uses Lua scripts and blocking pop commands that require a persistent TCP connection. Upstash Redis HTTP proxy does not support these. Using Upstash for asynq silently fails in production.

**Rule:** Never use Upstash as the asynq Redis. Never use Railway Redis for HTTP-compatible workloads. Keep env vars distinct.

---

### SSE Is Not tRPC

**Decision:** Generation progress streaming uses a standalone Next.js Route Handler, not a tRPC subscription.

- **Path:** `apps/web/app/api/generation/[jobId]/stream/route.ts`
- **Transport:** `text/event-stream` via `ReadableStream`
- **Why:** tRPC subscriptions add unnecessary adapter complexity. The SSE endpoint is simple, stateless, and framework-agnostic.

```typescript
export async function GET(req, { params }) {
  const stream = new ReadableStream({ async start(controller) { /* poll → emit */ } });
  return new Response(stream, {
    headers: { "Content-Type": "text/event-stream", "Cache-Control": "no-cache" },
  });
}
```

---

### Tailwind v4 — CSS-Only Config

**Decision:** No `tailwind.config.ts`. All tokens live in `globals.css` using `@theme {}`.

```css
/* apps/web/app/globals.css */
@import "tailwindcss";

@theme {
  --color-volt: #7B2FFF;
  --color-convert-green: #00E87A;
  --color-signal-red: #FF3D3D;
  --color-data-blue: #2E9CFF;
  --color-success: #00E87A;
  --color-data: #2E9CFF;
  --font-display: "Clash Display", "Space Grotesk", sans-serif;
  --font-mono: "JetBrains Mono", monospace;
}
```

---

### Go vs Rust Split

| Service | Language | Reason |
|---|---|---|
| API server | Go (Echo v4) | I/O-bound; latency owned by FAL/OpenAI/HeyGen |
| Job workers | Go (asynq) | Goroutine-per-job; Redis queues |
| URL orchestrator | Go | Orchestrates Modal Playwright calls |
| Video postprocessor | Rust (Axum + ffmpeg-sys) | CPU-bound: watermark, caption burn, transcode, reframe |

**Rule:** Do not introduce Rust outside the postprocessor service. Go borrow-checker friction adds cost with no throughput gain on network-wait workloads.

---

### Clerk Organizations = Workspaces

**Decision:** One Clerk Organization = one Qvora workspace. Multi-tenant isolation achieved via Clerk JWT `org_id` claim, not a custom user_id column.

- JWT template injects `org_id` and `org_role` into every token
- Go API middleware verifies `org_id` against workspace record
- Supabase RLS policies use `(auth.jwt() ->> 'org_id')::UUID` — Clerk JWT puts `org_id` at the top-level claim, not inside `app_metadata`

---

### FAL.AI — Use Async Queue, Not Sync

**Decision:** Always use `fal.queue.submit()` for video generation, never `fal.subscribe()` (which blocks).

```typescript
// ✅ Correct
const { request_id } = await fal.queue.submit("fal-ai/veo3", { input: { prompt } });
// Store request_id in asynq task payload; poll from Go worker

// ❌ Wrong — blocks until completion, unusable for 30–120s operations
const result = await fal.subscribe("fal-ai/veo3", { input: { prompt } });
```

---

## Document Fixes Applied (Full History)

### Qvora_Architecture-Stack.md
- Added "Architecture Notes" section documenting Two-Redis rule
- Corrected SSE note: "standalone Route Handler, NOT tRPC"
- Fixed Stripe label (was blank/generic)

### Qvora_Product-Definition.md
- Reordered personas: Agency (P1/P2) before DTC (P4)
- Added T2V/I2V/V2V feature matrix to video generation section
- Added trial + acquisition motion section (7-day no-card, Day 8 lock)

### Qvora_Product-Overview.md
- Reordered ICP: Agency P1, DTC P2 (Phase 2 note)
- Added trial motion reference pointing to Definition doc

### Qvora_Competitive-Analysis.md
- Added AdCreative.ai §2.4 deep-dive
- Generalized competitor foil language (was HeyGen-specific)
- Expanded sources list to 17 references
- Fixed positioning map agency/DTC axis

### Qvora_Feature-Spec.md
- GEN-07: Variant limits: Starter=max 3/angle, Growth=max 10/angle, Scale=unlimited
- PLAT-08: Day 8 locked state, 30-day data retention, conversion email Day 3/6/8 sequence

### Qvora_User-Stories.md
- P3 (Account Manager): added "Reviewer role only — no generation stories" note
- P4 (DTC): added "Phase 2 ICP — not built for in V1" label
- Added US-04b: Trial state story (generation limit awareness during trial)
- Renumbered US-14b/14c/14d after trial story insertion
- Updated Story Summary table

### Qvora_Design-System.md
- Added `.light {}` CSS block for light theme
- Added Tailwind v4 `@theme {}` config block with all brand tokens
- Registered `--success` and `--data` custom tokens

### Qvora_Wireframes.md
- Fixed progress dots: S-03 `[●○]` (step 1 of 2), S-04 `[●●]` (step 2 of 2)
- Added S-13 Trial Locked wireframe (Day 8 lock screen)
- Added S-14 Plan Limit wireframe (variant limit enforcement UI)
- Former S-13 Signal Dashboard → S-15 (renumbered)
- Updated screen index table

### Qvora_Brand-Identity.md
- Agency-first ICP order in all tables
- Competitive foil generalized
- Summary Card ICP corrected
- Font stack locked: Clash Display / Space Grotesk / Inter / JetBrains Mono
- Onboarding tone includes "7-day no-card trial" language

---

## Key Reference Numbers

| Metric | Value | Source |
|---|---|---|
| Instagram CPM (2026) | $9.46 | Hootsuite 2026 |
| TikTok CPM (2026) | $4–$9 | Hootsuite 2026 |
| YouTube CPM (2026) | $4–$5 | Hootsuite 2026 |
| Creatify managed ad spend | $650M+ | Creatify.ai April 2026 |
| Creatify brand count | 15,000+ | Creatify.ai April 2026 |
| Role-based onboarding retention lift (7-day) | +35% | UserPilot 2026 |
| PureGym Thruplays lift (lo-fi Reels) | 5.6x | Hootsuite / Meta 2026 |

---

## Epics Summary

| Epic | Scope |
|---|---|
| 1 — Onboarding & Activation | Signup, role selection, brand setup, first generation in <15 min |
| 2 — Qvora Brief | URL ingestion, strategy engine, brief editor, angle variants |
| 3 — Qvora Studio | Video generation (T2V/I2V/V2V), variant gallery, preview |
| 4 — Export & Structured Testing | Platform-ready export, naming conventions, test sets |
| 5 — Brand Kit | Multi-brand management, logo/color/voice per brand |
| 6 — Team & Collaboration | Roles (buyer/director/reviewer), workspace sharing |
| 7 — Qvora Signal (V2) | Ad account connector, performance ingestion, fatigue detection |
| 8 — Platform & Admin | Billing, trial lifecycle, tier limits, account management |

---

## Competitors Quick Reference

| Competitor | Core Strength | Critical Weakness | Qvora Advantage |
|---|---|---|---|
| Creatify | Scale, $650M ad spend managed, launch integration | No pre-generation strategy layer | Brief Engine + Signal loop |
| Arcads | Realistic AI actors, UGC feel | Actor-centric workflow, not system | Full creative system: brief → video → signal |
| HeyGen | Best avatar quality, translation | Not ad-native, no brief/performance | Ad-native pipeline, V1 built for buyers |
| AdCreative.ai | Fast ad image + copy gen | Minimal video; no brief engine | End-to-end video pipeline |
| Sora / Runway | Best raw video quality | No ad workflow, no strategy layer | Qvora uses them as models (FAL.AI) |

---

## Implementation Reference Location

Full SDK documentation, install commands, code patterns, and integration notes:
→ `docs/07-implementation/Qvora_Implementation-References.md`

---

## Notes Added After Initial Docs

- **HeyGen version discrepancy:** Architecture-Stack incorrectly says "HeyGen Avatar API v4". Correct version is **v3**. Active platform: `developers.heygen.com`. v4 does not exist. V2V lip-sync is v3-only. Migration from v2 required before production (v2 sunset: Oct 31, 2026).
- **Vercel AI SDK version:** Confirmed v6 (ai-sdk.dev). Two sub-libraries: `ai` (Core: generateText, generateObject, streamText) and `ai` UI hooks (`useChat`, `useObject`, `useCompletion`).
- **asynq version:** v0.26.0 (latest stable, Feb 2026). MIT license, 13.1k GitHub stars.
