-- =============================================================================
-- Phase 3 Slice 2: Postprocess callback persistence and idempotency
-- =============================================================================

CREATE TABLE IF NOT EXISTS postprocess_callbacks (
    id              BIGSERIAL PRIMARY KEY,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES variants(id) ON DELETE CASCADE,
    request_id      TEXT NOT NULL,
    job_id          UUID REFERENCES jobs(id) ON DELETE SET NULL,
    status          TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (workspace_id, variant_id, request_id)
);

CREATE INDEX IF NOT EXISTS idx_postprocess_callbacks_variant
    ON postprocess_callbacks(variant_id, created_at DESC);
