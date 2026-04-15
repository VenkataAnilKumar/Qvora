# QVORA
## Database Schema
**Version:** 1.0 | **Date:** April 14, 2026 | **Status:** Draft
**Database:** PostgreSQL 16 via Supabase · ORM: sqlc (Go codegen)

---

## Design Principles

1. **Org-level isolation** — every table has `org_id` (Clerk org); RLS enforced at DB layer
2. **V2-ready schema** — Signal tables and FK anchors defined in V1 even if unused until V2
3. **Immutable assets** — generated assets are never mutated; new versions create new rows
4. **Audit trail** — `created_at` / `updated_at` on all tables; soft-delete via `archived_at`
5. **Tag-first design** — all creative metadata stored as queryable columns, not JSON blobs

---

## Entity Relationship Overview

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

## Schema DDL

### Extensions

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";          -- V2: pgvector for semantic search
```

---

### Table: organizations

```sql
CREATE TABLE organizations (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  clerk_org_id    TEXT UNIQUE NOT NULL,           -- Clerk organization ID
  name            TEXT NOT NULL,
  plan            TEXT NOT NULL DEFAULT 'trial'
                  CHECK (plan IN ('trial','starter','growth','scale','enterprise')),
  plan_started_at TIMESTAMPTZ,
  trial_ends_at   TIMESTAMPTZ,
  ads_used_month  INTEGER NOT NULL DEFAULT 0,     -- rolling monthly counter
  ads_limit_month INTEGER NOT NULL DEFAULT 3,     -- 3 during trial; 20/100/∞ on plan
  billing_anchor  DATE,                           -- day of month billing resets
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orgs_clerk_id ON organizations(clerk_org_id);
```

---

### Table: users

```sql
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  clerk_user_id   TEXT UNIQUE NOT NULL,
  email           TEXT NOT NULL,
  display_name    TEXT,
  role            TEXT NOT NULL DEFAULT 'creator'
                  CHECK (role IN ('admin','creator','reviewer')),
  onboarding_role TEXT,                           -- media_buyer / creative_director / etc.
  ad_spend_range  TEXT,                           -- < 10k / 10k-100k / etc.
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_org    ON users(org_id);
CREATE INDEX idx_users_clerk  ON users(clerk_user_id);
```

---

### Table: brands

```sql
CREATE TABLE brands (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  slug            TEXT NOT NULL,                  -- url-safe, used in export naming
  primary_color   TEXT,                           -- hex e.g. "#7B2FFF"
  secondary_color TEXT,
  logo_url        TEXT,                           -- R2 URL
  font_url        TEXT,                           -- R2 URL (TTF/OTF)
  intro_url       TEXT,                           -- R2 URL (MP4 bumper)
  outro_url       TEXT,                           -- R2 URL (MP4 bumper)
  tone_notes      TEXT,                           -- free text brand voice notes
  default_avatar  TEXT,                           -- HeyGen avatar ID
  default_voice   TEXT,                           -- ElevenLabs voice ID
  voice_clone_id  TEXT,                           -- ElevenLabs cloned voice ID (Growth+)
  archived_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(org_id, slug)
);

CREATE INDEX idx_brands_org ON brands(org_id);

-- RLS
ALTER TABLE brands ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON brands
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: briefs

```sql
CREATE TABLE briefs (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id            UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  brand_id          UUID NOT NULL REFERENCES brands(id),
  created_by        UUID NOT NULL REFERENCES users(id),
  product_url       TEXT,
  product_summary   TEXT,
  product_data      JSONB,                        -- raw extraction result
  extraction_score  SMALLINT,                     -- 0-100 confidence
  template_used     TEXT,                         -- dtc / mobile_app / saas / etc.
  status            TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','extracting','ready','approved','archived')),
  human_reviewed    BOOLEAN NOT NULL DEFAULT FALSE,
  shared_token      TEXT UNIQUE,                  -- read-only share link token
  shared_expires_at TIMESTAMPTZ,
  archived_at       TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_briefs_org    ON briefs(org_id);
CREATE INDEX idx_briefs_brand  ON briefs(brand_id);
CREATE INDEX idx_briefs_status ON briefs(org_id, status);

ALTER TABLE briefs ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON briefs
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: brief_angles

```sql
CREATE TABLE brief_angles (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  brief_id     UUID NOT NULL REFERENCES briefs(id) ON DELETE CASCADE,
  org_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  angle_name   TEXT NOT NULL,
  rationale    TEXT,
  emotion      TEXT CHECK (emotion IN
               ('aspiration','urgency','fear','humor','trust','curiosity','social_proof')),
  funnel_stage TEXT CHECK (funnel_stage IN
               ('awareness','consideration','conversion','retention')),
  sort_order   SMALLINT NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_angles_brief ON brief_angles(brief_id);
CREATE INDEX idx_angles_org   ON brief_angles(org_id);

ALTER TABLE brief_angles ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON brief_angles
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: brief_hooks

```sql
CREATE TABLE brief_hooks (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  angle_id      UUID NOT NULL REFERENCES brief_angles(id) ON DELETE CASCADE,
  org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  hook_type     TEXT NOT NULL CHECK (hook_type IN
                ('problem','desire','social_proof','shock','curiosity')),
  opening_line  TEXT NOT NULL,
  variant_index SMALLINT NOT NULL DEFAULT 1,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hooks_angle ON brief_hooks(angle_id);
CREATE INDEX idx_hooks_org   ON brief_hooks(org_id);

ALTER TABLE brief_hooks ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON brief_hooks
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: generation_jobs

```sql
CREATE TABLE generation_jobs (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id            UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  brand_id          UUID NOT NULL REFERENCES brands(id),
  brief_id          UUID REFERENCES briefs(id),
  created_by        UUID NOT NULL REFERENCES users(id),

  -- Job config
  generation_mode   TEXT NOT NULL CHECK (generation_mode IN
                    ('ugc','demo','text_motion','text_to_video',
                     'image_to_video','voice_to_video')),
  model_used        TEXT,                         -- veo-3.1 / kling-3.0 / runway / heygen
  platform          TEXT CHECK (platform IN ('meta_feed','tiktok','youtube_short','stories')),
  aspect_ratio      TEXT CHECK (aspect_ratio IN ('9:16','1:1','16:9')),
  duration_sec      SMALLINT,

  -- Input refs
  angle_id          UUID REFERENCES brief_angles(id),
  hook_id           UUID REFERENCES brief_hooks(id),
  prompt_text       TEXT,
  source_image_url  TEXT,                         -- I2V: uploaded product image
  voice_source      TEXT CHECK (voice_source IN ('upload','elevenlabs','clone')),
  voice_asset_url   TEXT,                         -- V2V: uploaded audio URL
  avatar_id         TEXT,                         -- HeyGen avatar ID

  -- Job tracking
  status            TEXT NOT NULL DEFAULT 'queued'
                    CHECK (status IN ('queued','processing','post_processing',
                                      'uploading','complete','failed','cancelled')),
  fal_request_id    TEXT,                         -- FAL.AI async queue ID
  heygen_job_id     TEXT,                         -- HeyGen job ID
  error_message     TEXT,
  progress_pct      SMALLINT DEFAULT 0,

  -- Timing
  queued_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  started_at        TIMESTAMPTZ,
  completed_at      TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_org     ON generation_jobs(org_id);
CREATE INDEX idx_jobs_status  ON generation_jobs(org_id, status);
CREATE INDEX idx_jobs_fal     ON generation_jobs(fal_request_id) WHERE fal_request_id IS NOT NULL;

ALTER TABLE generation_jobs ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON generation_jobs
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: assets

```sql
-- Core creative metadata — the V2 learning foundation
-- All tag columns are queryable; no JSONB tags blob

CREATE TABLE assets (
  id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id              UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  brand_id            UUID NOT NULL REFERENCES brands(id),
  job_id              UUID NOT NULL REFERENCES generation_jobs(id),
  brief_id            UUID REFERENCES briefs(id),

  -- File refs
  r2_key              TEXT NOT NULL,              -- Cloudflare R2 object key
  r2_url              TEXT NOT NULL,              -- public CDN URL
  mux_asset_id        TEXT,                       -- Mux video asset ID
  mux_playback_id     TEXT,                       -- Mux playback ID (HLS)
  filename            TEXT NOT NULL,              -- structured export name
  file_size_bytes     INTEGER,
  duration_sec        SMALLINT,
  resolution          TEXT,                       -- "1080p" / "4k"
  aspect_ratio        TEXT,                       -- "9:16" / "1:1" / "16:9"

  -- Creative metadata tags (queryable columns — V2 Signal foundation)
  generation_mode     TEXT NOT NULL,              -- ugc / t2v / i2v / v2v / demo / text_motion
  angle_type          TEXT CHECK (angle_type IN
                      ('awareness','consideration','conversion','retention')),
  hook_type           TEXT CHECK (hook_type IN
                      ('problem','desire','social_proof','shock','curiosity')),
  format              TEXT,                       -- ugc / avatar / product_demo / text_motion
  emotion             TEXT,                       -- aspiration / urgency / fear / humor / trust
  platform            TEXT,                       -- meta_feed / tiktok / youtube_short / stories
  model_used          TEXT,                       -- veo-3.1 / kling-3.0 / etc.
  variant_index       SMALLINT NOT NULL DEFAULT 1,

  -- Status
  status              TEXT NOT NULL DEFAULT 'ready'
                      CHECK (status IN ('processing','ready','exported','archived')),
  human_reviewed      BOOLEAN NOT NULL DEFAULT FALSE,
  archived_at         TIMESTAMPTZ,

  -- V2: Ad account FK anchors (schema reserved in V1)
  ad_id               TEXT,                       -- Meta / TikTok ad creative ID
  campaign_id         TEXT,                       -- Ad platform campaign ID
  fatigue_detected_at TIMESTAMPTZ,                -- Set by Signal engine when fatigue found

  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assets_org         ON assets(org_id);
CREATE INDEX idx_assets_brand       ON assets(brand_id);
CREATE INDEX idx_assets_brief       ON assets(brief_id);
CREATE INDEX idx_assets_status      ON assets(org_id, status);
CREATE INDEX idx_assets_angle_type  ON assets(org_id, angle_type);
CREATE INDEX idx_assets_hook_type   ON assets(org_id, hook_type);
CREATE INDEX idx_assets_fatigue     ON assets(fatigue_detected_at) WHERE fatigue_detected_at IS NOT NULL;

ALTER TABLE assets ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON assets
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: exports

```sql
CREATE TABLE exports (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  brand_id      UUID NOT NULL REFERENCES brands(id),
  created_by    UUID NOT NULL REFERENCES users(id),
  asset_count   SMALLINT NOT NULL,
  format        TEXT NOT NULL DEFAULT 'mp4_h264',
  resolution    TEXT NOT NULL DEFAULT '1080p',
  zip_r2_key    TEXT,
  zip_url       TEXT,                             -- signed CDN URL (48hr expiry)
  zip_size_bytes BIGINT,
  manifest_url  TEXT,                             -- manifest.csv URL
  status        TEXT NOT NULL DEFAULT 'queued'
                CHECK (status IN ('queued','building','ready','expired')),
  expires_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE export_items (
  export_id UUID NOT NULL REFERENCES exports(id) ON DELETE CASCADE,
  asset_id  UUID NOT NULL REFERENCES assets(id),
  PRIMARY KEY (export_id, asset_id)
);

CREATE INDEX idx_exports_org ON exports(org_id);

ALTER TABLE exports ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON exports
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: ad_accounts (V2)

```sql
CREATE TABLE ad_accounts (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  brand_id        UUID NOT NULL REFERENCES brands(id),
  platform        TEXT NOT NULL CHECK (platform IN ('meta','tiktok','google')),
  account_id      TEXT NOT NULL,                  -- Platform ad account ID
  account_name    TEXT,
  access_token    TEXT,                           -- encrypted
  refresh_token   TEXT,                           -- encrypted
  token_expires_at TIMESTAMPTZ,
  last_synced_at  TIMESTAMPTZ,
  sync_status     TEXT DEFAULT 'pending'
                  CHECK (sync_status IN ('pending','active','error','disconnected')),
  error_message   TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(org_id, platform, account_id)
);

CREATE INDEX idx_ad_accounts_org ON ad_accounts(org_id);

ALTER TABLE ad_accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON ad_accounts
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: asset_metrics (V2)

```sql
CREATE TABLE asset_metrics (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  asset_id        UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  ad_account_id   UUID NOT NULL REFERENCES ad_accounts(id),
  platform        TEXT NOT NULL,
  date            DATE NOT NULL,

  -- Performance metrics
  impressions     INTEGER NOT NULL DEFAULT 0,
  clicks          INTEGER NOT NULL DEFAULT 0,
  spend_cents     INTEGER NOT NULL DEFAULT 0,     -- in cents to avoid float
  conversions     INTEGER NOT NULL DEFAULT 0,
  ctr             NUMERIC(6,4),                   -- computed: clicks/impressions
  cpa_cents       INTEGER,                        -- spend/conversions
  roas            NUMERIC(8,4),                   -- revenue/spend
  hold_rate_25    NUMERIC(5,2),                   -- % watched 25%
  hold_rate_50    NUMERIC(5,2),
  hold_rate_75    NUMERIC(5,2),
  hold_rate_100   NUMERIC(5,2),
  completion_rate NUMERIC(5,2),

  synced_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(asset_id, ad_account_id, date)
);

CREATE INDEX idx_metrics_asset   ON asset_metrics(asset_id);
CREATE INDEX idx_metrics_org     ON asset_metrics(org_id);
CREATE INDEX idx_metrics_date    ON asset_metrics(org_id, date DESC);

ALTER TABLE asset_metrics ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON asset_metrics
  USING (org_id = (auth.jwt() ->> 'org_id')::UUID);
```

---

### Table: team_invites

```sql
CREATE TABLE team_invites (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  invited_by  UUID NOT NULL REFERENCES users(id),
  email       TEXT NOT NULL,
  role        TEXT NOT NULL DEFAULT 'creator'
              CHECK (role IN ('admin','creator','reviewer')),
  token       TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex'),
  accepted_at TIMESTAMPTZ,
  expires_at  TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '48 hours',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  UNIQUE(org_id, email)
);

CREATE INDEX idx_invites_token ON team_invites(token);
CREATE INDEX idx_invites_org   ON team_invites(org_id);
```

---

## Triggers

```sql
-- Auto-update updated_at on all tables
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
DO $$
DECLARE t TEXT;
BEGIN
  FOREACH t IN ARRAY ARRAY[
    'organizations','users','brands','briefs',
    'brief_angles','brief_hooks','assets'
  ] LOOP
    EXECUTE format(
      'CREATE TRIGGER trg_updated_at BEFORE UPDATE ON %I
       FOR EACH ROW EXECUTE FUNCTION update_updated_at()', t
    );
  END LOOP;
END $$;

-- Auto-increment org ads_used_month when asset status → ready
CREATE OR REPLACE FUNCTION increment_ads_used()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status = 'ready' AND OLD.status != 'ready' THEN
    UPDATE organizations
    SET ads_used_month = ads_used_month + 1
    WHERE id = NEW.org_id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ads_used
AFTER UPDATE ON assets
FOR EACH ROW EXECUTE FUNCTION increment_ads_used();
```

---

## Indexes Summary (RLS performance)

All columns used in RLS policies are indexed — Supabase best practice for multi-tenant query performance:

```sql
-- Ensure all org_id columns have indexes (RLS policy columns)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_briefs_org_id       ON briefs(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_assets_org_id       ON assets(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_brands_org_id       ON brands(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_jobs_org_id         ON generation_jobs(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exports_org_id      ON exports(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ad_accounts_org_id  ON ad_accounts(org_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_metrics_org_id      ON asset_metrics(org_id);
```

---

## V1 vs V2 Schema Boundaries

| Table | V1 | V2 activates |
|---|---|---|
| `organizations` | ✅ Full | — |
| `users` | ✅ Full | — |
| `brands` | ✅ Full (voice_clone_id nullable) | voice_clone_id populated |
| `briefs` | ✅ Full | — |
| `brief_angles` | ✅ Full | — |
| `brief_hooks` | ✅ Full | — |
| `generation_jobs` | ✅ Full | — |
| `assets` | ✅ Full (ad_id, campaign_id, fatigue nullable) | ad_id populated via connector |
| `exports` | ✅ Full | — |
| `ad_accounts` | ✅ Table created, empty | Populated when user connects account |
| `asset_metrics` | ✅ Table created, empty | Populated by Signal sync worker |
| `team_invites` | ✅ Full | — |

---

*Database Schema v1.0 — Qvora*
*April 14, 2026 — Confidential*

---

**Sources:**
- [Supabase RLS Best Practices — Makerkit](https://makerkit.dev/blog/tutorials/supabase-rls-best-practices)
- [Multi-Tenant RLS on Supabase — Antstack](https://www.antstack.com/blog/multi-tenant-applications-with-rls-on-supabase-postgress/)
- [Supabase RLS Docs](https://supabase.com/docs/guides/database/postgres/row-level-security)
