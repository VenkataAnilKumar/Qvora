---
title: Database Schema
category: architecture
tags: [database, schema, postgresql, supabase, sqlc, rls, tables]
sources: [Qvora_Database-Schema]
updated: 2026-04-15
---

# Database Schema

## TL;DR
PostgreSQL 16 via Supabase. sqlc for Go codegen (no ORM). Every table has `org_id` (RLS isolation). Immutable assets — new version = new row. V2 anchor tables (ad_accounts, asset_metrics) defined in V1 schema but unused until V2.

---

## Design Principles

1. **Org-level isolation** — every table has `org_id` (Clerk org); RLS enforced at DB layer, always `(SELECT auth.uid())` wrapper
2. **V2-ready schema** — Signal tables and FK anchors defined in V1 even if unused
3. **Immutable assets** — generated assets never mutated; new versions = new rows
4. **Audit trail** — `created_at` / `updated_at` on all tables; soft-delete via `archived_at`
5. **Tag-first design** — all creative metadata stored as queryable columns, not JSON blobs

---

## Entity Relationship

```
organizations ──< brands ──< briefs ──< brief_angles ──< brief_hooks
               │
               └──< generation_jobs ──< assets ──< asset_metrics (V2)
                                          │
                                          └──< export_items >── exports

organizations ──< users (via Clerk)
organizations ──< ad_accounts (V2)
```

---

## Key Tables

### `organizations`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `clerk_org_id` | TEXT UNIQUE | Clerk organization ID |
| `name` | TEXT | — |
| `plan` | TEXT | CHECK: `trial | starter | growth | scale | enterprise` |
| `plan_started_at` | TIMESTAMPTZ | — |
| `trial_ends_at` | TIMESTAMPTZ | 7 days from creation |
| `ads_used_month` | INTEGER | Rolling monthly counter |
| `ads_limit_month` | INTEGER | 3 (trial) → tier limits on upgrade |
| `billing_anchor` | DATE | Day of month billing resets |

> ⚠️ **Discrepancy vs. Implementation Checklist:** Schema DDL shows `plan` CHECK includes `'scale'`. Checklist states `plan_tier CHECK: starter, growth, agency` (no 'scale'). **Checklist is likely more recent.** Verify against `supabase/migrations/001_initial_schema.sql`.

### `users`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK → organizations | — |
| `clerk_user_id` | TEXT UNIQUE | — |
| `email` | TEXT | — |
| `role` | TEXT | CHECK: `admin | creator | reviewer` |
| `onboarding_role` | TEXT | media_buyer / creative_director / etc. |
| `ad_spend_range` | TEXT | Collected at signup |

### `brands`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK | — |
| `name` | TEXT | — |
| `slug` | TEXT | URL-safe, used in export naming |
| `primary_color` | TEXT | Hex e.g. `#7B2FFF` |
| `secondary_color` | TEXT | — |
| `logo_r2_key` | TEXT | Cloudflare R2 key |
| `tone_of_voice` | TEXT | Free-text brand voice notes |
| `archived_at` | TIMESTAMPTZ | Soft-delete |

### `briefs`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK | — |
| `brand_id` | UUID FK → brands | — |
| `source_url` | TEXT | Original product URL |
| `product_name` | TEXT | Extracted from page |
| `product_description` | TEXT | Extracted/summarized |
| `status` | TEXT | `pending | extracting | ready | failed` |
| `created_by` | UUID FK → users | — |

### `brief_angles`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `brief_id` | UUID FK | — |
| `org_id` | UUID FK | — |
| `angle_name` | TEXT | e.g. "Social Proof", "Problem-Solution" |
| `rationale` | TEXT | Why this angle for this product |
| `recommended_format` | TEXT | UGC / Spokesperson / Demo / Voiceover |
| `visual_direction` | TEXT | Scene/style description |
| `position` | INTEGER | Order in brief (1–5) |

### `brief_hooks`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `angle_id` | UUID FK → brief_angles | — |
| `org_id` | UUID FK | — |
| `hook_text` | TEXT | The hook copy |
| `hook_type` | TEXT | question / statement / bold_claim / stat |
| `position` | INTEGER | Order within angle (1–3) |

### `generation_jobs`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK | — |
| `brief_id` | UUID FK | — |
| `angle_id` | UUID FK | — |
| `hook_id` | UUID FK | — |
| `status` | TEXT | CHECK: `queued | scraping | generating | postprocessing | complete | failed` |
| `model` | TEXT | veo-3.1 / kling-3.0 / runway-gen-4.5 / sora-2 |
| `fal_request_id` | TEXT | From FAL.AI queue submit response |
| `error_message` | TEXT | Last error if failed |

### `assets`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK | — |
| `job_id` | UUID FK → generation_jobs | — |
| `r2_key` | TEXT | Cloudflare R2 key |
| `mux_asset_id` | TEXT | Mux asset ID |
| `mux_playback_id` | TEXT | Mux signed playback ID |
| `duration_seconds` | INTEGER | — |
| `format` | TEXT | UGC / Spokesperson / Demo / Voiceover |
| `platform` | TEXT | tiktok / reels / meta / youtube_shorts |
| `watermarked` | BOOLEAN | — |
| `archived_at` | TIMESTAMPTZ | Soft-delete |

### `exports`

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | — |
| `org_id` | UUID FK | — |
| `status` | TEXT | `queued | building | ready | failed` |
| `r2_key` | TEXT | ZIP bundle key |
| `download_url` | TEXT | Pre-signed R2 URL |
| `expires_at` | TIMESTAMPTZ | Download link TTL |

---

## RLS Pattern

All tables use this pattern:
```sql
-- Enable RLS
ALTER TABLE tablename ENABLE ROW LEVEL SECURITY;

-- Read policy
CREATE POLICY "org_read" ON tablename
  FOR SELECT TO authenticated
  USING (org_id = (SELECT auth.uid()::text::uuid));

-- Write policy  
CREATE POLICY "org_write" ON tablename
  FOR ALL TO authenticated
  USING (org_id = (SELECT auth.uid()::text::uuid));
```

> **Rule:** Always use `(SELECT auth.uid())` wrapper — not bare `auth.uid()` — for performance (prevents per-row evaluation).

---

## Index Strategy

All tables have at minimum:
- Index on `org_id` (isolation queries)
- Index on `status` (job polling)
- Index on `created_at DESC` (list queries)
- Composite index on `(org_id, status)` for filtered lists

---

## Open Questions
- [ ] `plan` CHECK: is 'scale' still valid or is it `starter | growth | agency` only? (discrepancy between schema doc and checklist)
- [ ] V2: `asset_metrics` table structure for Signal data ingestion?
- [ ] V2: pgvector column placement — on `assets` or `briefs` or separate embeddings table?

## Related Pages
- [[data-layer]] — Redis, R2, Mux alongside PostgreSQL
- [[system-architecture]] — how Go API interacts with this schema
- [[implementation-checklist]] — which tables are confirmed built
