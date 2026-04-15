-- =============================================================================
-- Qvora Database Schema — PostgreSQL 16 via Supabase
-- Canonical migration location: supabase/migrations/
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Workspaces (one per Clerk Organization)
CREATE TABLE IF NOT EXISTS workspaces (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          TEXT NOT NULL UNIQUE,
    plan_tier       TEXT NOT NULL DEFAULT 'starter' CHECK (plan_tier IN ('starter', 'growth', 'agency')),
    sub_status      TEXT NOT NULL DEFAULT 'trialing' CHECK (sub_status IN ('trialing', 'active', 'past_due', 'canceled')),
    stripe_sub_id   TEXT,
    trial_ends_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Brands (renamed from brand_kits)
CREATE TABLE IF NOT EXISTS brands (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id      UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    logo_r2_key       TEXT,
    primary_color     TEXT NOT NULL DEFAULT '#7B2FFF',
    secondary_color   TEXT,
    font_family       TEXT,
    watermark_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Generation jobs (renamed from generation_jobs)
CREATE TABLE IF NOT EXISTS jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    product_url     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'queued'
                        CHECK (status IN ('queued','scraping','generating','postprocessing','complete','failed')),
    model           TEXT NOT NULL DEFAULT 'veo3'
                        CHECK (model IN ('veo3','kling3','runway4','sora2')),
    brief_json      JSONB,
    error_msg       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Variant rows (renamed from video_variants)
CREATE TABLE IF NOT EXISTS variants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    angle           TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'queued'
                        CHECK (status IN ('queued','generating','postprocessing','complete','failed')),
    fal_request_id  TEXT,
    mux_asset_id    TEXT,
    mux_playback_id TEXT,
    r2_key          TEXT,
    duration_secs   NUMERIC(6,2),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Brief parent rows
CREATE TABLE IF NOT EXISTS briefs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    scrape_job_id   UUID REFERENCES jobs(id) ON DELETE SET NULL,
    product_url     TEXT NOT NULL,
    model           TEXT NOT NULL DEFAULT 'gpt-4o',
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'generated', 'approved')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS brief_angles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id        UUID NOT NULL REFERENCES briefs(id) ON DELETE CASCADE,
    angle           TEXT NOT NULL,
    headline        TEXT NOT NULL,
    script          TEXT NOT NULL,
    cta             TEXT NOT NULL,
    voice_tone      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS brief_hooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brief_id        UUID NOT NULL REFERENCES briefs(id) ON DELETE CASCADE,
    hook            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS asset_tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id      UUID NOT NULL REFERENCES variants(id) ON DELETE CASCADE,
    tag             TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (variant_id, tag)
);

CREATE TABLE IF NOT EXISTS exports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES variants(id) ON DELETE CASCADE,
    destination     TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'complete', 'failed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_jobs_workspace_id ON jobs(workspace_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_variants_job_id ON variants(job_id);
CREATE INDEX IF NOT EXISTS idx_variants_workspace_id ON variants(workspace_id);
CREATE INDEX IF NOT EXISTS idx_briefs_workspace_id ON briefs(workspace_id);
CREATE INDEX IF NOT EXISTS idx_brief_angles_brief_id ON brief_angles(brief_id);
CREATE INDEX IF NOT EXISTS idx_brief_hooks_brief_id ON brief_hooks(brief_id);
CREATE INDEX IF NOT EXISTS idx_exports_workspace_id ON exports(workspace_id);

-- =============================================================================
-- Row-Level Security
-- =============================================================================

ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE brands ENABLE ROW LEVEL SECURITY;
ALTER TABLE jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE variants ENABLE ROW LEVEL SECURITY;
ALTER TABLE briefs ENABLE ROW LEVEL SECURITY;
ALTER TABLE brief_angles ENABLE ROW LEVEL SECURITY;
ALTER TABLE brief_hooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE asset_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE exports ENABLE ROW LEVEL SECURITY;

CREATE POLICY "workspace_members_only" ON workspaces
    FOR ALL TO authenticated
    USING (org_id = (SELECT current_setting('app.org_id', TRUE)));

CREATE POLICY "workspace_brands" ON brands
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

CREATE POLICY "workspace_jobs" ON jobs
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

CREATE POLICY "workspace_variants" ON variants
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

CREATE POLICY "workspace_briefs" ON briefs
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

CREATE POLICY "workspace_brief_angles" ON brief_angles
    FOR ALL TO authenticated
    USING (brief_id IN (
        SELECT id FROM briefs
        WHERE workspace_id IN (
            SELECT id FROM workspaces
            WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
        )
    ));

CREATE POLICY "workspace_brief_hooks" ON brief_hooks
    FOR ALL TO authenticated
    USING (brief_id IN (
        SELECT id FROM briefs
        WHERE workspace_id IN (
            SELECT id FROM workspaces
            WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
        )
    ));

CREATE POLICY "workspace_asset_tags" ON asset_tags
    FOR ALL TO authenticated
    USING (variant_id IN (
        SELECT id FROM variants
        WHERE workspace_id IN (
            SELECT id FROM workspaces
            WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
        )
    ));

CREATE POLICY "workspace_exports" ON exports
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));
