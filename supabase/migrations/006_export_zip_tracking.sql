-- =============================================================================
-- Migration 006 — Export ZIP tracking
-- Adds the columns the Export Engine needs to serve bulk-ZIP downloads:
-- the R2 key of the generated ZIP, and a download counter incremented each
-- time GET /api/v1/exports/:id/download is hit.
-- =============================================================================

ALTER TABLE exports ADD COLUMN IF NOT EXISTS r2_key TEXT;
ALTER TABLE exports ADD COLUMN IF NOT EXISTS download_count INT NOT NULL DEFAULT 0;
