---
title: Data Layer — Storage, Redis, & Databases
category: architecture
tags: [postgresql, supabase, redis, upstash, railway, R2, mux, data-layer]
sources: [Qvora_Architecture-Stack, Qvora_Database-Schema]
updated: 2026-04-15
---

# Data Layer — Storage, Redis, & Databases

## TL;DR
PostgreSQL (Supabase + RLS) for primary data. Two Redis instances that must never be swapped. Cloudflare R2 for asset storage. Mux for video streaming + analytics. sqlc for type-safe Go queries.

---

## Primary Database — PostgreSQL via Supabase

**Driver:** pgx/v5 (via sqlc in Go)  
**Multi-tenancy:** Row-Level Security (RLS) with `org_id` scoping

### RLS Rules (Non-negotiable)
- Always use `(SELECT auth.uid())` wrapper (not `auth.uid()` directly — prevents performance regression)
- Grant TO `authenticated` role only
- Every user-scoped table must have an index on `user_id` (and `org_id` where applicable)

### Query Pattern
1. Write SQL in `db/queries/*.sql`
2. Run `sqlc generate` → type-safe Go structs in `internal/db/`
3. Never raw-string SQL in Go handlers

---

## Redis — TWO INSTANCES (Critical)

> ⚠️ **Never substitute one for the other.** They serve fundamentally different roles at the infrastructure level.

| Instance | Provider | Protocol | Used For |
|---|---|---|---|
| **Cache Redis** | Upstash | HTTP/REST | Cache, rate-limiting (sliding window), session store |
| **Queue Redis** | Railway | TCP (persistent) | asynq job queues (`BLPOP` — requires persistent connections) |

### Why Two?
- Upstash uses HTTP — safe for serverless (no persistent connection management), but `BLPOP` (used by asynq) doesn't work over HTTP.
- Railway Redis is a traditional TCP Redis — asynq workers maintain persistent connections to it.
- If you use Upstash for asynq, job workers will fail silently or not pick up jobs at all.

---

## Object Storage — Cloudflare R2

**Zero egress cost** — no outbound bandwidth charges (unlike S3).

**Usage:**
- Raw extracted images from product URLs
- Generated video files (pre-Mux upload)
- Export downloads

**Upload pattern:** Presigned PUT URLs issued by Go API → client uploads directly to R2 (no server proxy).

---

## Video Streaming — Mux

- HLS adaptive streaming for all generated/exported videos
- Signed playback URLs scoped to workspace (`org_id`) — prevents cross-workspace access
- Mux analytics: engagement, play rate, completion rate per video

**Mux webhook:** `mux.video.asset.ready` → Go API updates video asset status in PostgreSQL.

---

## Architecture Decisions

| Decision | Rationale |
|---|---|
| Supabase over raw Postgres | RLS + managed infra; pgvector available for V2 |
| sqlc over ORM (GORM/ent) | Type-safe, no N+1 surprises, explicit SQL = auditable |
| R2 over S3 | Zero egress; Cloudflare CDN collocated |
| Mux over direct video serving | Adaptive HLS, analytics, signed URLs, auto-thumbnail |
| Two Redis over one | Architectural constraint: asynq TCP ≠ Upstash HTTP |

---

## V2 Data

- **pgvector on Supabase** — embedding store for performance signal similarity (find similar ad variants to winning ones)
- Required for the Performance Learning Engine (SIGNAL module)

---

## Open Questions
- [ ] What is the DB schema for `briefs`, `angles`, `hooks`, `variants`? (See `docs/06-technical/Qvora_Database-Schema.md`)
- [ ] What is the Upstash rate-limit key structure? (per user? per org? per endpoint?)
- [ ] What is the retention policy for raw R2 files after Mux ingest?

## Related Pages
- [[stack-overview]] — full architecture overview
- [[ai-layer]] — how AI jobs interact with the queue Redis
- [[features]] — which features produce which data entities
