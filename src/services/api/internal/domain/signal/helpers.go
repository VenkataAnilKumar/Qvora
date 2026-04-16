package signal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/store"
)

// ensureSignalTables creates all signal tables and indexes if they don't exist.
func ensureSignalTables(c echo.Context) error {
	if store.Pool() == nil {
		return errors.New("database_not_initialized")
	}
	ctx := c.Request().Context()
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS signal_connections (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			platform TEXT NOT NULL CHECK (platform IN ('meta', 'tiktok')),
			status TEXT NOT NULL DEFAULT 'connected' CHECK (status IN ('connected', 'disconnected', 'token_expired')),
			account_id TEXT NOT NULL,
			account_name TEXT,
			error_reason TEXT,
			token_expires_at TIMESTAMPTZ,
			last_synced_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, platform, account_id)
		)`,
		`ALTER TABLE signal_connections ADD COLUMN IF NOT EXISTS error_reason TEXT`,
		`CREATE TABLE IF NOT EXISTS signal_metrics_daily (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			variant_id UUID REFERENCES variants(id) ON DELETE SET NULL,
			platform TEXT NOT NULL CHECK (platform IN ('meta', 'tiktok')),
			metric_date DATE NOT NULL,
			impressions BIGINT NOT NULL DEFAULT 0,
			clicks BIGINT NOT NULL DEFAULT 0,
			spend_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
			conversions BIGINT NOT NULL DEFAULT 0,
			revenue_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
			hold_25 DOUBLE PRECISION,
			hold_50 DOUBLE PRECISION,
			hold_75 DOUBLE PRECISION,
			hold_100 DOUBLE PRECISION,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, platform, variant_id, metric_date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_metrics_daily_workspace_date ON signal_metrics_daily(workspace_id, metric_date DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_connections_workspace_platform ON signal_connections(workspace_id, platform)`,
		`CREATE TABLE IF NOT EXISTS signal_fatigue_events (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			variant_id UUID NOT NULL REFERENCES variants(id) ON DELETE CASCADE,
			detected_on DATE NOT NULL,
			current_ctr DOUBLE PRECISION NOT NULL,
			peak_ctr DOUBLE PRECISION NOT NULL,
			drop_pct DOUBLE PRECISION NOT NULL,
			sustained_days INTEGER NOT NULL DEFAULT 3,
			suggested_action TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved')),
			first_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, variant_id, detected_on)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_fatigue_events_workspace_status ON signal_fatigue_events(workspace_id, status, last_detected_at DESC)`,
		`CREATE TABLE IF NOT EXISTS signal_brief_recommendations (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			angle TEXT NOT NULL,
			suggested_hook TEXT NOT NULL,
			rationale TEXT NOT NULL,
			confidence_score DOUBLE PRECISION NOT NULL,
			impression_volume BIGINT NOT NULL DEFAULT 0,
			window_days INTEGER NOT NULL DEFAULT 90,
			first_generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, angle)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_recommendations_workspace_generated ON signal_brief_recommendations(workspace_id, last_generated_at DESC)`,
		`CREATE TABLE IF NOT EXISTS signal_recommendation_feedback (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			angle TEXT NOT NULL,
			action TEXT NOT NULL CHECK (action IN ('accept', 'ignore')),
			source TEXT NOT NULL DEFAULT 'brief_create_panel',
			created_by TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_signal_recommendation_feedback_workspace_created ON signal_recommendation_feedback(workspace_id, created_at DESC)`,
	} {
		if _, err := store.Pool().Exec(ctx, stmt); err != nil {
			return fmt.Errorf("ensureSignalTables: %w", err)
		}
	}
	return nil
}

// parseDays parses the days query param, clamped to [1, 90].
func parseDays(raw string) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return 30
	}
	if v > 90 {
		return 90
	}
	return v
}

// parsePositiveFloat parses a positive float64 from a string.
func parsePositiveFloat(raw string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

// parseBool parses a bool from a query string.
func parseBool(raw string, fallback bool) bool {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return fallback
	}
	if trimmed == "1" || trimmed == "true" || trimmed == "yes" {
		return true
	}
	if trimmed == "0" || trimmed == "false" || trimmed == "no" {
		return false
	}
	return fallback
}
