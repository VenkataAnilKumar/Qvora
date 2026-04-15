# QVORA Implementation Checklist

**Last Updated:** Apr 15, 2026 | **Status:** Phase 1/2 Complete, Phase 3 In Planning

---

## Phase 1/2: Core Platform (✅ COMPLETE)

### Data Layer
- [x] PostgreSQL schema (Supabase migration)
- [x] Table naming (jobs, variants, briefs, brief_angles, brief_hooks, asset_tags, exports)
- [x] RLS policies (per-workspace)
- [x] Indexes (workspace_id, status, job_id, brief_id)
- [x] sqlc codegen to `internal/db`
- [x] Docker Compose migration bootstrap

### Auth & Multi-Tenancy
- [x] Clerk integration (web)
- [x] Clerk integration (API & middleware)
- [x] Organization → workspace mapping
- [x] Row-level security (all tables)
- [x] Middleware auth guards

### Pricing & Tier Enforcement
- [x] Plan tier definitions (Starter/Growth/Agency)
- [x] Variant limit checks (3/10/unlimited)
- [x] Tier enforcement in job submit
- [x] Plan tier sourced from workspace state (not headers)
- [x] Stripe Entitlements schema prep

### Brief Generation (Web-Side AI SDK)
- [x] tRPC briefs router (create/list/get)
- [x] AI SDK integration (generateObject ×2: GPT-4o product extraction + Claude Sonnet 4.6 angles/hooks)
- [x] Structural schema (Zod)
- [x] Prompt engineering
- [x] Environment guards (OPENAI_API_KEY)
- [x] Response streaming

### URL Scraping
- [x] Modal Playwright integration
- [x] Task enqueue (scrape → generate, no brief queue task)
- [x] ProductExtraction struct
- [x] Job status transitions (scraping → generating)
- [x] Error handling & retry

### Video Generation (FAL.AI)
- [x] FAL async queue submit (not subscribe)
- [x] Model selection (Veo/Kling/Runway/Sora)
- [x] Request ID tracking
- [x] 9:16 aspect ratio enforcement
- [x] Job status transitions (generating → postprocessing)
- [x] Error handling

### Job Orchestration (Asynq)
- [x] Railway Redis TCP setup
- [x] Task serialization (JSON)
- [x] Queue priorities (critical/default/low)
- [x] Task handlers (TypeScrape, TypeGenerate, TypePostprocess)
- [x] Status callbacks to API
- [ ] Orphan/stale task cleanup

### API & tRPC
- [x] Go Echo API server
- [x] POST /api/v1/jobs (submit)
- [x] GET /api/v1/jobs (list)
- [x] GET /api/v1/jobs/:id (detail)
- [x] PATCH /api/v1/jobs/:id/status (worker-only update)
- [x] GET /api/v1/workspaces/:orgId
- [x] tRPC briefs router
- [ ] SSE stream endpoint (wire frontend to `app/api/generation/[jobId]/stream/route.ts`)

### Frontend Dashboard
- [x] Root layout (TRPCProvider, ClerkProvider, QueryClientProvider)
- [x] Briefs list page (/briefs)
- [x] Briefs detail page (/briefs/[id])
- [ ] Job/variant list display UI
- [ ] Real-time status via SSE (EventSource → /api/generation/[jobId]/stream)
- [ ] Export modal

### Docker & Deployment
- [x] Docker Compose (local dev)
- [x] API Dockerfile
- [x] Worker Dockerfile
- [x] Postprocessor Dockerfile
- [x] Environment templates (.env.example)

### Testing & Validation
- [x] Web typecheck (tsc)
- [x] API build & tests
- [x] Worker build & tests
- [ ] E2E job submission (manual)
- [ ] Integration: scrape → generate

---

## Phase 3: Video Postprocessing & Mux (🟡 PLANNED)

### Rust Postprocessor
- [ ] Axum HTTP server (scaffold: ✅)
- [ ] FAL output download from R2
- [ ] ffmpeg watermark overlay
- [ ] ffmpeg caption burn-in
- [ ] ffmpeg 9:16 reframe/letterbox
- [ ] ffmpeg H.264 transcode
- [ ] Upload processed output to R2
- [ ] Asynq queue integration (job:postprocess)

### Mux Integration
- [ ] Mux API client (HLS upload)
- [ ] Asset ID + playback ID storage (variants table)
- [ ] Signed playback tokens (workspace scope)
- [ ] Mux Player SDK integration
- [ ] Webhook: upload complete → job done

### Job Status & Callbacks
- [ ] FAL webhook receiver (optional)
- [ ] Postprocess enqueue from worker
- [ ] Mux upload webhook handler
- [ ] Final job completion (all variants done)
- [ ] User notification (in-app + email)

### Database & Persistence
- [ ] Persist generated brief angles to DB
- [ ] Persist generated brief hooks to DB
- [ ] Store FAL request_id in variants
- [ ] Store Mux asset_id + playback_id
- [ ] Auto-save brief changes

### Testing
- [ ] Unit tests for postprocessor
- [ ] Integration: FAL → ffmpeg → R2
- [ ] E2E: submit → scrape → gen → postproc → Mux → playback

---

## Phase 4+: Signal Learning Loop & V2 (🟡 ICEBOX)

### Performance Signal (Qvora Signal)
- [ ] Ad account connector (Meta/TikTok OAuth)
- [ ] Variant performance sync (spend, impressions, clicks, conversions)
- [ ] Temporal workflow (scheduled ingestion)
- [ ] LLM learning loop (best angles per industry)
- [ ] Variant scoring + ranking

### Smart Exports
- [ ] TV (TikTok/Instagram Reels format)
- [ ] Pinterest format
- [ ] YouTube Ads format
- [ ] Native export destinations

### Team Collaboration & RBAC
- [ ] Granular role permissions (reviewer/editor/viewer)
- [ ] Team member management
- [ ] Audit logs

### Brand Kit Advanced
- [ ] Font upload + custom typography
- [ ] Logo animation templates
- [ ] Custom color palette builder
- [ ] Brand asset library management

---

## Known Issues & Mitigations

| Issue | Status | Mitigation |
|---|---|---|
| Plan tier from in-memory state (not DB workspace table) | 🟡 | Replace with DB workspace lookup in Phase 3 |
| Brief angles/hooks not persisted to DB | 🟡 | Auto-persist in Phase 3 |
| FAL request IDs not stored in variants | 🟡 | Add storage + polling in Phase 3 |
| Postprocessor ffmpeg logic not implemented | 🟡 | Phase 3 deliverable |
| No Mux integration | 🟡 | Phase 3 deliverable |
| No export destinations | 🟡 | Phase 4 deliverable |
| Team collab limited to Clerk org | 🟡 | Add RBAC in Phase 4 |

---

## Environment & Configuration

### Required Env Vars (Phase 1/2)
```
# Web
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
CLERK_SECRET_KEY
OPENAI_API_KEY

# API
DATABASE_URL
OPENAI_API_KEY
INTERNAL_API_KEY (for worker auth)

# Worker
DATABASE_URL
RAILWAY_REDIS_URL
FAL_KEY
MODAL_SCRAPER_ENDPOINT
INTERNAL_API_KEY
```

### Phase 3 Additions
```
# Postprocessor
AWS_REGION (auto for Cloudflare R2)
R2_BUCKET
R2_ENDPOINT

# API/Worker
RUST_POSTPROCESS_URL
MUX_TOKEN_ID
MUX_TOKEN_SECRET
```

---

## Quick Stats

| Metric | Phase 1/2 | Phase 3 | Phase 4+ |
|---|---|---|---|
| API endpoints | 7 | +3 (Mux) | +5 (Signal) |
| Database tables | 9 | +2 (signal, metrics) | +3 (learning) |
| Worker task types | 3 | 4 | +2 (Signal jobs) |
| External services | 5 | +1 (Mux) | +2 (Temporal, Ad platforms) |
| Frontend pages | 2 | +4 (variant detail, export, analytics) | +5 (Signal dashboard) |

---

## File Locations

**Core Implementation Files:**
- Data schema: [supabase/migrations/001_initial_schema.sql](../../supabase/migrations/001_initial_schema.sql)
- API: [src/services/api/](../../src/services/api/)
- Worker: [src/services/worker/](../../src/services/worker/)
- Web: [src/apps/web/](../../src/apps/web/)
- Postprocessor: [src/services/postprocess/](../../src/services/postprocess/)
- Docker: [docker-compose.yml](../../docker-compose.yml)

**Reference Docs:**
- Product Definition: [docs/02-product/Qvora_Product-Definition.md](../02-product/Qvora_Product-Definition.md)
- Feature Spec: [docs/04-specs/Qvora_Feature-Spec.md](../04-specs/Qvora_Feature-Spec.md)
- Architecture: [docs/06-technical/Qvora_Architecture-Stack.md](../06-technical/Qvora_Architecture-Stack.md)
- Implementation Phases: [docs/07-implementation/Qvora_Implementation-Phases.md](./Qvora_Implementation-Phases.md)

---

## How to Use This Checklist

1. **Daily standup:** Review Phase 3 checklist items; mark completed items as you go.
2. **Sprint planning:** Use Phase 3 checklist to scope 2-week sprints; flag blockers.
3. **Code review:** Ensure checklist items have passing tests before merging.
4. **Roadmap updates:** Move items between phases as priorities shift; update "Last Updated" date.

