# QVORA
## Implementation Phases Tracker
**Version:** 1.0 | **Date:** April 15, 2026 | **Status:** Phase 1/2 Complete, Phase 3 In Planning

---

## Executive Summary

Qvora is implemented in three phases:
- **Phase 1/2 (COMPLETE):** Core API, brief generation (web-side AI SDK), video generation pipeline (FAL.AI async queue), auth (Clerk), tier enforcement, dashboard scaffold.
- **Phase 3 (PLANNED):** Postprocessing (Rust ffmpeg), Mux integration, performance signal loop, expanded admin/analytics.
- **Phase 4+ (ICEBOX):** V2 features (Signal learning loop, Ad account connector).

---

## Phase 1/2: Core Platform & Brief/Video Pipeline
**Status:** ✅ **COMPLETE** (Apr 15, 2026)

### Completed Deliverables

#### 1. Data Layer & Schema
| Item | Status | Notes |
|---|---|---|
| PostgreSQL schema (Supabase) | ✅ | Canonical migration at `supabase/migrations/001_initial_schema.sql` |
| Table naming (spec-aligned) | ✅ | jobs, variants, briefs, brief_angles, brief_hooks, asset_tags, exports |
| RLS policies (per-workspace) | ✅ | All tables use workspace_id as partition key |
| Indexes (query perf) | ✅ | Created on workspace_id, status, job_id, brief_id |
| sqlc codegen (Go) | ✅ | `internal/db` package auto-generated from `db/queries/queries.sql` |
| Migration bootstrap (Docker Compose) | ✅ | Mounts `./supabase/migrations` to Postgres init |

**Code Location:** [supabase/migrations/001_initial_schema.sql](../../supabase/migrations/001_initial_schema.sql)

#### 2. Auth & Multi-Tenancy
| Item | Status | Notes |
|---|---|---|
| Clerk integration (web) | ✅ | `ClerkProvider`, `clerkMiddleware()`, JWT org_id + org_role |
| Clerk integration (API) | ✅ | Echo middleware extracts org_id from JWT claims |
| Organization workspace mapping | ✅ | One Clerk Org = one Qvora workspace |
| Row-level security | ✅ | All DB policies validate org_id context |
| Middleware auth guards | ✅ | API endpoints reject missing/invalid claims |

**Code Location:** [src/services/api/internal/middleware/auth.go](../../src/services/api/internal/middleware/auth.go)

#### 3. Pricing & Tier Enforcement
| Item | Status | Notes |
|---|---|---|
| Plan tiers (Starter/Growth/Agency) | ✅ | Starter: 3 variants/angle, Growth: 10, Agency: unlimited |
| Variant limit enforcement | ✅ | Server-side check in job submit handler |
| Plan tier sourcing | ✅ | Now from workspace state (not request headers) |
| Stripe Entitlements (future) | 🟡 | Schema ready; webhook integration pending Phase 3 |

**Code Location:** [src/services/api/internal/middleware/tier.go](../../src/services/api/internal/middleware/tier.go), [src/services/api/internal/handler/job.go](../../src/services/api/internal/handler/job.go)

#### 4. Brief Generation (Web-Side AI SDK)
| Item | Status | Notes |
|---|---|---|
| tRPC briefs router | ✅ | `create`, `list`, `get` endpoints |
| AI SDK integration | ✅ | `generateObject` ×2: GPT-4o (product extraction) + Claude Sonnet 4.6 (angles/hooks) |
| Structural schema | ✅ | Zod schema for angles + hooks output |
| Prompt engineering | ✅ | Dedicated prompt file with performance-marketing tone |
| Environment guards | ✅ | Validates OPENAI_API_KEY before calling AI |
| Response streaming | ✅ | Structured output returned in create response |

**Code Location:** [src/apps/web/src/server/trpc/routers/briefs.ts](../../src/apps/web/src/server/trpc/routers/briefs.ts), [src/ai/prompts/angles-gen.prompt.ts](../../src/ai/prompts/angles-gen.prompt.ts)

**Architecture Decision:** Brief generation moved entirely to web tRPC (AI SDK) to avoid worker latency bottleneck. Worker-side brief task (`task.HandleBrief`) removed from active queue pipeline.

#### 5. URL Scraping & Extraction
| Item | Status | Notes |
|---|---|---|
| Modal Playwright endpoint integration | ✅ | Worker calls serverless scraper, handles Modal response |
| Task enqueue (scrape → generate) | ✅ | Scrape task now enqueues generation tasks directly |
| ProductExtraction struct | ✅ | Captures name, category, price, features, proof_points, images, description |
| Job status transitions | ✅ | scraping → generating (no briefing step) |
| Error handling + retry | ✅ | Failed scrapes transition job to "failed" status |

**Code Location:** [src/services/worker/internal/task/scrape.go](../../src/services/worker/internal/task/scrape.go)

#### 6. Video Generation (FAL.AI Async Queue)
| Item | Status | Notes |
|---|---|---|
| FAL queue submit (not subscribe) | ✅ | Real async queue submit, never blocking fal.subscribe() |
| Model selection | ✅ | Veo 3.1, Kling 3.0, Runway Gen-4.5, Sora 2 supported |
| Request ID tracking | ✅ | Parses FAL request_id for polling/webhook callbacks |
| 9:16 aspect ratio enforcement | ✅ | Hardcoded in FAL payload |
| Job status transitions | ✅ | generating → postprocessing (on submit success) |
| Error handling (timeout, API errors) | ✅ | Failed submits mark job as "failed" |

**Code Location:** [src/services/worker/internal/task/generate.go](../../src/services/worker/internal/task/generate.go)

**Architecture Note:** Workers always use FAL queue submit via HTTP REST gateway (`https://queue.fal.run/`), never fal.subscribe() (blocks under load). Request IDs are stored for polling-based status checks or webhook integration in Phase 3.

#### 7. Job Orchestration (Asynq Queue)
| Item | Status | Notes |
|---|---|---|
| Redis queue setup | ✅ | Railway Redis TCP (required for BLPOP; never Upstash HTTP) |
| Task serialization | ✅ | JSON payload marshaling/unmarshaling |
| Queue priorities | ✅ | critical (postprocess, webhooks) / default (gen) / low (cleanup) |
| Task handlers | ✅ | TypeScrape, TypeGenerate, TypePostprocess registered |
| Status callbacks to API | ✅ | Workers patch job status via PATCH /api/v1/jobs/:id/status |
| Orphan / stale task cleanup | 🟡 | TODO for Phase 3 |

**Code Location:** [src/services/worker/cmd/worker/main.go](../../src/services/worker/cmd/worker/main.go), [src/services/worker/internal/task/types.go](../../src/services/worker/internal/task/types.go), [src/services/worker/internal/task/api.go](../../src/services/worker/internal/task/api.go)

**Key Decision:** Removed legacy worker-side brief task (queue type `job:brief`) from active queue. Brief generation now only runs in web tRPC layer. Scrape task enqueues generation tasks directly.

#### 8. API & tRPC BFF
| Item | Status | Notes |
|---|---|---|
| Go Echo API server | ✅ | `/api/v1/jobs`, `/api/v1/briefs`, `/api/v1/workspaces` routes |
| Endpoint: POST /api/v1/jobs | ✅ | Validates URL, model, tier limits; returns job_id + status |
| Endpoint: GET /api/v1/jobs | ✅ | List jobs with pagination (limit 20–50) |
| Endpoint: GET /api/v1/jobs/:id | ✅ | Fetch single job details + status |
| Endpoint: PATCH /api/v1/jobs/:id/status | ✅ | Worker-only status updates (internal API key auth) |
| Endpoint: GET /api/v1/workspaces/:orgId | ✅ | Return workspace tier + subscription status |
| tRPC briefs router | ✅ | End-to-end type-safe briefs create/list/get |
| SSE stream handler | 🟡 | Standalone Route Handler at `app/api/generation/[jobId]/stream/route.ts`; frontend polling not yet wired |

**Code Location:** [src/services/api/cmd/api/main.go](../../src/services/api/cmd/api/main.go), [src/apps/web/src/server/trpc/routers/briefs.ts](../../src/apps/web/src/server/trpc/routers/briefs.ts)

#### 9. Frontend Dashboard
| Item | Status | Notes |
|---|---|---|
| Layout at root (layout.tsx) | ✅ | TRPCProvider, ClerkProvider, QueryClientProvider |
| Briefs list page | ✅ | Dashboard route `/briefs` |
| Briefs detail page | ✅ | Dashboard route `/briefs/[id]` |
| Job/variant list display | 🟡 | Scaffold ready; UI iteration pending |
| Real-time status polling | 🟡 | TODO (will use React Query polling on job GET) |
| Export modal | 🟡 | Scaffold ready; export destinations pending Phase 3 |

**Code Location:** [src/apps/web/src/app/layout.tsx](../../src/apps/web/src/app/layout.tsx), [src/apps/web/src/app/(dashboard)/briefs/](../../src/apps/web/src/app/(dashboard)/briefs/)

#### 10. Docker & Deployment
| Item | Status | Notes |
|---|---|---|
| Docker Compose (local dev) | ✅ | Postgres + Redis + web + api + worker services |
| API Dockerfile | ✅ | Multi-stage Go compile, alpine base |
| Worker Dockerfile | ✅ | Multi-stage Go compile, scratch base |
| Postprocessor Dockerfile | ✅ | Multi-stage Rust compile, ffmpeg dev libs |
| Environment file templates | ✅ | `.env.example` for local dev |

**Code Location:** [docker-compose.yml](../../docker-compose.yml), [src/services/api/Dockerfile](../../src/services/api/Dockerfile), [src/services/worker/Dockerfile](../../src/services/worker/Dockerfile)

### Implementation Decisions (Phase 1/2)

| Decision | Rationale | Status |
|---|---|---|
| Brief generation in web tRPC (not worker) | Reduces latency (direct API call vs queue wait); leverages Vercel AI SDK v6 | ✅ Implemented |
| Scrape → Generate pipeline (no brief queue task) | Simplifies job state machine; brief strategy now just example scripts | ✅ Implemented |
| Tier enforcement from workspace state | Replaces request header fallback; prepares for Stripe Entitlements integration | ✅ Implemented |
| FAL async queue (not subscribe) | Prevents worker blocking under load; HTTP REST is safe for containerized workers | ✅ Implemented |
| Railway Redis for asynq (not Upstash) | Upstash HTTP doesn't support BLPOP; TCP required for persistent worker connections | ✅ Architecture locked in |
| PostgreSQL RLS over app-layer auth | Enforces multi-tenancy at data layer; reduces bug surface | ✅ Implemented |

---

## Phase 3: Video Postprocessing & Mux Integration
**Status:** 🟡 **PLANNED** (Target: May 2026)

### Deliverables

#### 1. Rust Postprocessor Service
| Item | Status | Notes |
|---|---|---|
| Axum HTTP server | 🟡 | Scaffold ready; handler stubs in place |
| FAL output download from R2 | 🟡 | TODO |
| ffmpeg watermark overlay | 🟡 | TODO |
| ffmpeg caption burn-in (if script provided) | 🟡 | TODO |
| ffmpeg 9:16 reframe / letterbox | 🟡 | TODO |
| ffmpeg H.264 transcode | 🟡 | TODO |
| Upload processed output to R2 | 🟡 | TODO |
| Queue integration (job:postprocess) | 🟡 | Task struct ready; worker handler ready |

**Code Location:** [src/services/postprocess/](../../src/services/postprocess/) (scaffold complete)

**Scope Note:** Rust handles only CPU-bound ffmpeg work. Go workers orchestrate async queue submission to Rust. In production, postprocessor runs on isolated Railway service with guaranteed CPU allocation.

#### 2. Mux Integration
| Item | Status | Notes |
|---|---|---|
| Mux API client (upload HLS) | 🟡 | TODO |
| Asset ID + playback ID storage | 🟡 | TODO (variants table: mux_asset_id, mux_playback_id) |
| Signed playback tokens (workspace scope) | 🟡 | TODO |
| Video preview player (Mux Player SDK) | 🟡 | TODO |
| Webhook: upload complete → job done | 🟡 | TODO |

**Code Location:** TBD (Phase 3 sprint)

#### 3. Job Status & Callback Flow
| Item | Status | Notes |
|---|---|---|
| FAL webhook receiver (optional) | 🟡 | Alternative to polling; reduces latency |
| Postprocess enqueue from worker | 🟡 | TODO (once ffmpeg ready) |
| Mux upload webhook handler | 🟡 | TODO (on success, mark variant "complete") |
| Final job completion (all variants done) | 🟡 | TODO (async job completion check) |
| User notification (in-app + email opt-in) | 🟡 | TODO (Loops.io or Sendgrid) |

### Phase 3 Acceptance Criteria

- [ ] Rust service compiles and deploys to Railway
- [ ] End-to-end test: FAL output → postprocess → Mux upload → playback link (10 min latency)
- [ ] Video plays in dashboard preview with correct 9:16 aspect ratio
- [ ] Watermark + captions visible in output
- [ ] All variants for a job complete, job marked "complete" in API

---

## Phase 4: Signal Learning Loop & V2 Features
**Status:** 🟡 **ICEBOX** (Target: Q3 2026)

### V2 Features (Out of Scope Phase 1/2)

| Feature | V1 Status | V2 Plan | Notes |
|---|---|---|---|
| Performance Signal (Qvora Signal) | ❌ | 🟡 | Ad account connector → variant performance metrics → LLM learns best angles |
| Ad Account Connector (Meta/TikTok) | ❌ | 🟡 | OAuth, spend sync, impression/click/conversion tracking |
| Temporal workflow (Signal loop) | ❌ | 🟡 | Scheduled tasks for performance ingestion + model retraining |
| Variant scoring + ranking | ❌ | 🟡 | ML: which angles perform best in each industry |
| Smart exports (TV/Pinterest/YouTube Ads) | ❌ | 🟡 | Auto-format variants for platform-specific specs |
| Team collaboration + roles | ❌ | 🟡 | Reviewer/editor/viewer permissions beyond current Clerk org |
| Brand kit advanced options | ❌ | 🟡 | Font upload, logo animation, custom color palettes |

---

## Current Architecture & Key Services

### Frontend (Next.js 15, Vercel)
```
src/apps/web/
├── src/
│   ├── app/layout.tsx                        [TRPCProvider, ClerkProvider]
│   ├── app/(dashboard)/                      [Dashboard routes]
│   │   ├── briefs/page.tsx                   [List briefs]
│   │   └── briefs/[id]/page.tsx              [Brief detail + video variants]
│   ├── server/
│   │   └── trpc/routers/briefs.ts            [Brief create/list/get]
src/ai/prompts/
└── angles-gen.prompt.ts                      [Prompt + Zod schema — shared AI layer under src/]
└── package.json
```

### Backend API (Go + Echo, Railway)
```
src/services/api/
├── cmd/api/main.go                           [Echo server setup]
├── internal/
│   ├── db/                                   [sqlc generated]
│   ├── handler/
│   │   ├── job.go                            [Job submit, GET, list, status update]
│   │   └── workspace.go                      [Workspace tier + subscription]
│   └── middleware/
│       ├── auth.go                           [Clerk JWT validation]
│       └── tier.go                           [Plan limit enforcement]
├── db/queries/queries.sql                    [sqlc SQL definitions]
├── sqlc.yaml                                 [sqlc config → `./supabase/migrations/`]
└── Dockerfile
```

### Worker (Go + Asynq, Railway)
```
src/services/worker/
├── cmd/worker/main.go                        [asynq server + mux setup]
└── internal/task/
    ├── types.go                              [Task type constants]
    ├── api.go                                [patchJobStatus helper]
    ├── scrape.go                             [TypeScrape → Modal → Generate]
    ├── generate.go                           [TypeGenerate → FAL queue]
    └── postprocess.go                        [TypePostprocess → Rust (Phase 3)]
```

### Postprocessor (Rust + Axum, Railway)
```
src/services/postprocess/
├── src/
│   ├── main.rs                               [Axum server setup]
│   ├── handler.rs                            [POST /process]
│   ├── model.rs                              [ProcessRequest/Response]
│   ├── error.rs                              [Error handling]
│   └── processor.rs                          [ffmpeg placeholder]
├── Cargo.toml                                [Rust deps: axum, ffmpeg-sys]
└── Dockerfile
```

### Data Layer
- **PostgreSQL (Supabase):** [supabase/migrations/001_initial_schema.sql](../../supabase/migrations/001_initial_schema.sql)
- **Redis (Railway):** Asynq job queue (TCP only)
- **Redis (Upstash):** Cache + rate limiting (HTTP REST only)
- **R2 (Cloudflare):** Raw FAL output, processed videos

---

## Testing & Validation

### Current Test Coverage
| Component | Status | Notes |
|---|---|---|
| Web typecheck (tsc) | ✅ | Passes without errors |
| API build + unit tests | ✅ | All packages compile; no test files yet (Phase 3) |
| Worker build + unit tests | ✅ | All packages compile; no test files yet (Phase 3) |
| E2E job submission (manual) | 🟡 | TODO (requires live FAL account) |
| Integration: scrape → generate | 🟡 | TODO (once Modal endpoint live) |

### Phase 3 Testing Plan
- [ ] Unit tests for postprocessor handlers (tokio test framework)
- [ ] Integration test: FAL output → ffmpeg → R2
- [ ] E2E: submit job → scrape → generate → postprocess → Mux → playback

---

## Environment & Configuration

### Local Development (.env variables)

**Web (src/apps/web/.env.local)**
```
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=...
CLERK_SECRET_KEY=...
OPENAI_API_KEY=...
```

**API (src/services/api/.env.local)**
```
DATABASE_URL=postgres://...
OPENAI_API_KEY=...
INTERNAL_API_KEY=... (for worker → API auth)
```

**Worker (src/services/worker/.env.local)**
```
DATABASE_URL=postgres://...
RAILWAY_REDIS_URL=redis://...
FAL_KEY=... (for FAL queue submit)
MODAL_SCRAPER_ENDPOINT=https://...
RUST_POSTPROCESS_URL=... (Phase 3)
INTERNAL_API_KEY=... (for auth with API)
```

**Postprocessor (src/services/postprocess/.env.local)**
```
AWS_REGION=auto (for Cloudflare R2)
R2_BUCKET=...
R2_ENDPOINT=https://...
```

---

## Known Limitations & Caveats

| Item | Status | Mitigation |
|---|---|---|
| Plan tier sourced from in-memory workspace state (not DB) | 🟡 | Replaces header fallback; real workspace table lookup pending Stripe integration (Phase 3) |
| Generated brief angles/hooks not persisted to DB briefs tables | 🟡 | Returned in web API response only; Phase 3 will auto-persist to `briefs`, `brief_angles`, `brief_hooks` |
| FAL request IDs not stored in variants table | 🟡 | Prepared schema field; Phase 3 will add async polling → store mux_asset_id |
| Postprocessor ffmpeg logic not implemented | 🟡 | Scaffold + types ready; Phase 3 deliverable |
| No Mux integration yet | 🟡 | Video streams to R2 only; Phase 3 adds HLS upload + playback |
| No export destinations (TikTok/Instagram/YouTube) | 🟡 | Schema ready; Phase 3+ feature |
| Team collaboration limited to Clerk org members | 🟡 | Fine for V1 (Agency tier); Phase 3 adds granular RBAC |

---

## Deployment & DevOps

### Current Deployment Targets
| Service | Platform | Status | Config |
|---|---|---|---|
| Web | Vercel | ✅ | Standard Next.js deployment |
| API | Railway | ✅ | Dockerfile + PORT env var |
| Worker | Railway | ✅ | Dockerfile + asynq config |
| Postprocessor | Railway | 🟡 | Dockerfile ready; not deployed yet |
| DB | Supabase (PostgreSQL) | ✅ | RLS policies enabled |
| Cache + Queue (asynq) | Railway Redis | ✅ | TCP for asynq |
| HTTP Cache + Rate Limit | Upstash Redis | ✅ | HTTP REST |
| File Storage | Cloudflare R2 | ✅ | S3-compatible API |

### Startup Instructions (Local)
```bash
# Clone repo
git clone https://github.com/VenkataAnilKumar/Qvora.git
cd Qvora

# Install + start services
docker-compose up -d
cd src/apps/web && npm run dev   # Starts Next.js on :3000
cd ../../services/api && go run ./cmd/api   # Starts Echo on :8080
cd ../worker && go run ./cmd/worker   # Starts Asynq worker
```

---

## Next Steps & Roadmap

### Immediate (Weeks 1–2)
- [ ] Live FAL account + key provisioning
- [ ] Modal Playwright scraper endpoint integration test
- [ ] Manual E2E: URL → scrape → brief → generate

### Phase 3 (Weeks 3–6)
- [ ] Complete Rust postprocessor (ffmpeg watermark, caption, transcode)
- [ ] Mux HLS upload integration + signed playback tokens
- [ ] Database persistence: briefs/brief_angles/brief_hooks auto-save
- [ ] Webhook: FAL completion → postprocess enqueue

### Phase 4 (Q3 2026)
- [ ] Performance Signal: Meta/TikTok ad account connector
- [ ] Temporal workflow: variant performance ingestion loop
- [ ] Export destinations: native TV, Pinterest, YouTube Ads formats

---

## Document History

| Version | Date | Author | Changes |
|---|---|---|---|
| 1.0 | Apr 15, 2026 | AI Agent | Initial Phase 1/2 completion snapshot; Phase 3+ planning |

---

## References

- [Product Definition](../02-product/Qvora_Product-Definition.md)
- [Feature Specification](../04-specs/Qvora_Feature-Spec.md)
- [User Stories](../04-specs/Qvora_User-Stories.md)
- [Architecture & Stack](../06-technical/Qvora_Architecture-Stack.md)
- [Brand Identity](../01-brand/Qvora_Brand-Identity.md)
