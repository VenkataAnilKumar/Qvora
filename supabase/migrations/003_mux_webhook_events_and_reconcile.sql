-- =============================================================================
-- Phase 3 Slice: Mux webhook dedup + reconciliation support
-- =============================================================================

CREATE TABLE IF NOT EXISTS mux_webhook_events (
    id            BIGSERIAL PRIMARY KEY,
    event_id      TEXT NOT NULL UNIQUE,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mux_webhook_events_created_at
    ON mux_webhook_events(created_at DESC);
