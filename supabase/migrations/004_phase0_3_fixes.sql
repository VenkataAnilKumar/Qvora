-- =============================================================================
-- Migration 004 — Phase 0-3 Fixes
-- Adds: video_performance_events, cost_events, idempotency columns,
--       fal_request_id unique constraint, avatar_provider column
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. video_performance_events
--    Append-only performance metrics per variant processing stage.
--    Never UPDATE rows — only INSERT.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS video_performance_events (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id     UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  variant_id       UUID        REFERENCES variants(id) ON DELETE SET NULL,
  job_id           UUID        REFERENCES jobs(id) ON DELETE SET NULL,

  -- Stage: 'scrape' | 'brief_gen' | 'fal_queue' | 'fal_generate' |
  --        'postprocess_download' | 'postprocess_transcode' |
  --        'postprocess_upload' | 'mux_ingest' | 'total'
  stage            TEXT        NOT NULL,
  duration_ms      INT         NOT NULL CHECK (duration_ms >= 0),

  -- Model used (for cost attribution)
  model            TEXT,
  -- fal.ai request_id for correlation
  fal_request_id   TEXT,
  -- Error info (NULL if success)
  error_type       TEXT,
  error_msg        TEXT,

  recorded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vpe_workspace_recorded
  ON video_performance_events (workspace_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_vpe_variant
  ON video_performance_events (variant_id, stage);
CREATE INDEX IF NOT EXISTS idx_vpe_job
  ON video_performance_events (job_id);

-- RLS
ALTER TABLE video_performance_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON video_performance_events
  USING (workspace_id IN (
    SELECT id FROM workspaces WHERE org_id = current_setting('app.org_id', true)
  ));

-- -----------------------------------------------------------------------------
-- 2. cost_events
--    Per-workspace credit consumption log. Used by circuit breaker.
--    Each video generation + postprocess emits one row.
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cost_events (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id     UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  variant_id       UUID        REFERENCES variants(id) ON DELETE SET NULL,
  job_id           UUID        REFERENCES jobs(id) ON DELETE SET NULL,

  -- Source: 'fal_generate' | 'elevenlabs_tts' | 'heygen_avatar' | 'modal_scrape'
  source           TEXT        NOT NULL,
  model            TEXT,

  -- Cost in USD (estimated at dispatch time, actual at callback)
  estimated_usd    NUMERIC(10, 6) NOT NULL DEFAULT 0,
  actual_usd       NUMERIC(10, 6),

  -- Credits consumed (maps to Stripe metered billing)
  credits          INT         NOT NULL DEFAULT 1,

  recorded_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cost_workspace_recorded
  ON cost_events (workspace_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_cost_workspace_month
  ON cost_events (workspace_id, recorded_at);

-- RLS
ALTER TABLE cost_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "org_isolation" ON cost_events
  USING (workspace_id IN (
    SELECT id FROM workspaces WHERE org_id = current_setting('app.org_id', true)
  ));

-- -----------------------------------------------------------------------------
-- 3. Idempotency key on jobs
--    Client sends X-Idempotency-Key header → stored here.
--    ON CONFLICT (workspace_id, idempotency_key) DO NOTHING.
-- -----------------------------------------------------------------------------
ALTER TABLE jobs
  ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_jobs_idempotency
  ON jobs (workspace_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

-- -----------------------------------------------------------------------------
-- 4. fal_request_id UNIQUE constraint on variants
--    Prevents duplicate postprocess on duplicate FAL webhook delivery.
-- -----------------------------------------------------------------------------
ALTER TABLE variants
  ADD COLUMN IF NOT EXISTS avatar_provider TEXT DEFAULT 'heygen_v3';

ALTER TABLE variants
  ADD COLUMN IF NOT EXISTS avatar_job_id TEXT;

-- Add unique constraint on fal_request_id to prevent duplicate callbacks
CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_fal_request_id
  ON variants (fal_request_id)
  WHERE fal_request_id IS NOT NULL;

-- -----------------------------------------------------------------------------
-- 5. Workspace monthly cost limit
--    Configurable per tier. Circuit breaker checks this.
-- -----------------------------------------------------------------------------
ALTER TABLE workspaces
  ADD COLUMN IF NOT EXISTS monthly_cost_limit_usd  NUMERIC(10, 2) DEFAULT 50.00,
  ADD COLUMN IF NOT EXISTS current_month_cost_usd  NUMERIC(10, 2) DEFAULT 0.00,
  ADD COLUMN IF NOT EXISTS cost_reset_at           TIMESTAMPTZ    DEFAULT date_trunc('month', NOW()) + INTERVAL '1 month';

-- Function to reset monthly cost at start of billing period
CREATE OR REPLACE FUNCTION reset_monthly_cost()
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  UPDATE workspaces
  SET current_month_cost_usd = 0,
      cost_reset_at = date_trunc('month', NOW()) + INTERVAL '1 month'
  WHERE cost_reset_at <= NOW();
END;
$$;

-- -----------------------------------------------------------------------------
-- 6. Materialized view — creative_scores (foundation for V2 signal loop)
--    Created now so the schema exists; populated when signal-svc goes live.
-- -----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS creative_scores AS
  SELECT
    vpe.variant_id,
    vpe.workspace_id,
    COUNT(*)                                                         AS event_count,
    AVG(CASE WHEN vpe.stage = 'total' THEN vpe.duration_ms END)    AS avg_total_ms,
    MIN(CASE WHEN vpe.stage = 'total' THEN vpe.duration_ms END)    AS min_total_ms,
    MAX(CASE WHEN vpe.stage = 'total' THEN vpe.duration_ms END)    AS max_total_ms,
    MAX(vpe.recorded_at)                                            AS last_recorded_at
  FROM video_performance_events vpe
  WHERE vpe.recorded_at > NOW() - INTERVAL '7 days'
  GROUP BY vpe.variant_id, vpe.workspace_id
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_creative_scores_variant
  ON creative_scores (variant_id);
