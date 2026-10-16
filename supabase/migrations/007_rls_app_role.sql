-- =============================================================================
-- Migration 007 — Make RLS real: a non-superuser application role
--
-- Context: every RLS policy in 001_initial_schema.sql is declared `TO authenticated`
-- — the role Supabase's PostgREST layer assumes per-request when a client carries
-- a valid JWT. This Go API does NOT use PostgREST; it connects directly via pgx,
-- and DATABASE_URL has always pointed at the `postgres` superuser. Postgres
-- superusers unconditionally bypass RLS — it's a hard rule, not a config flag —
-- so every one of those policies has been inert for this application specifically
-- since day one, regardless of anything set at the session level.
--
-- This migration creates the role the Go API should actually connect as, and
-- extends the existing policies to also apply to it (keeping `authenticated` on
-- each policy too, in case PostgREST/Supabase client access is ever added later).
--
-- Deploy note: this migration creates the role WITHOUT a password — set one via
-- `ALTER ROLE qvora_app PASSWORD '...'` through Doppler/secrets management, never
-- committed here. DATABASE_URL must then be updated to connect as qvora_app, not
-- postgres, for RLS to actually take effect.
-- =============================================================================

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'qvora_app') THEN
    CREATE ROLE qvora_app LOGIN;
  END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO qvora_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO qvora_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO qvora_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO qvora_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO qvora_app;

ALTER POLICY "workspace_members_only" ON workspaces TO authenticated, qvora_app;
ALTER POLICY "workspace_brands" ON brands TO authenticated, qvora_app;
ALTER POLICY "workspace_jobs" ON jobs TO authenticated, qvora_app;
ALTER POLICY "workspace_variants" ON variants TO authenticated, qvora_app;
ALTER POLICY "workspace_briefs" ON briefs TO authenticated, qvora_app;
ALTER POLICY "workspace_brief_angles" ON brief_angles TO authenticated, qvora_app;
ALTER POLICY "workspace_brief_hooks" ON brief_hooks TO authenticated, qvora_app;
ALTER POLICY "workspace_asset_tags" ON asset_tags TO authenticated, qvora_app;
ALTER POLICY "workspace_exports" ON exports TO authenticated, qvora_app;

-- Tables added after 001 (002-006) never had RLS enabled at all — enabling it
-- here with the same app.org_id pattern, scoped through their variant_id/
-- workspace_id foreign keys the same way the brief_angles/asset_tags policies
-- above already do.
ALTER TABLE postprocess_callbacks ENABLE ROW LEVEL SECURITY;
CREATE POLICY "workspace_postprocess_callbacks" ON postprocess_callbacks
    FOR ALL TO authenticated, qvora_app
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

ALTER TABLE cost_events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "org_isolation" ON cost_events;
CREATE POLICY "workspace_cost_events" ON cost_events
    FOR ALL TO authenticated, qvora_app
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

ALTER TABLE video_performance_events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "org_isolation" ON video_performance_events;
CREATE POLICY "workspace_video_performance_events" ON video_performance_events
    FOR ALL TO authenticated, qvora_app
    USING (workspace_id IN (
        SELECT id FROM workspaces
        WHERE org_id = (SELECT current_setting('app.org_id', TRUE))
    ));

-- mux_webhook_events has no workspace_id — it's a dedup table keyed on Mux's
-- own event_id, inspected only by internal webhook handling, not per-tenant
-- user-facing reads. Intentionally left without RLS.
