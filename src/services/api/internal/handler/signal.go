package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

func ensureSignalTables(c echo.Context) error {
	if dbPool == nil {
		return errors.New("database_not_initialized")
	}

	_, err := dbPool.Exec(
		c.Request().Context(),
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
		)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `ALTER TABLE signal_connections ADD COLUMN IF NOT EXISTS error_reason TEXT`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
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
		)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `CREATE INDEX IF NOT EXISTS idx_signal_metrics_daily_workspace_date ON signal_metrics_daily(workspace_id, metric_date DESC)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `CREATE INDEX IF NOT EXISTS idx_signal_connections_workspace_platform ON signal_connections(workspace_id, platform)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
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
		)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `CREATE INDEX IF NOT EXISTS idx_signal_fatigue_events_workspace_status ON signal_fatigue_events(workspace_id, status, last_detected_at DESC)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
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
		)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `CREATE INDEX IF NOT EXISTS idx_signal_recommendations_workspace_generated ON signal_brief_recommendations(workspace_id, last_generated_at DESC)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
		`CREATE TABLE IF NOT EXISTS signal_recommendation_feedback (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			angle TEXT NOT NULL,
			action TEXT NOT NULL CHECK (action IN ('accept', 'ignore')),
			source TEXT NOT NULL DEFAULT 'brief_create_panel',
			created_by TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return err
	}

	_, err = dbPool.Exec(c.Request().Context(), `CREATE INDEX IF NOT EXISTS idx_signal_recommendation_feedback_workspace_created ON signal_recommendation_feedback(workspace_id, created_at DESC)`)
	return err
}

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

func parsePositiveFloat(raw string, fallback float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func parsePositiveInt(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

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

func resolveSuggestedAction(dropPct float64) string {
	if dropPct >= 50 {
		return "refresh_with_new_angle"
	}
	if dropPct >= 40 {
		return "refresh_hook_and_opening"
	}
	return "refresh_hook"
}

func recommendedHookForAngle(angle string) string {
	switch strings.ToLower(strings.TrimSpace(angle)) {
	case "problem_solution":
		return "Call out the pain, then show the fastest path to outcome."
	case "social_proof":
		return "Lead with proof volume and a concrete before/after result."
	case "transformation":
		return "Contrast old state vs new state in the first 3 seconds."
	case "urgency":
		return "Use a time-bound offer and direct CTA in opening line."
	case "education":
		return "Teach one surprising mechanism and tie it to conversion."
	default:
		return "Open with a concrete benefit and immediate CTA."
	}
}

func computeRecommendationConfidence(impressions int64, ctr float64) float64 {
	if impressions <= 0 {
		return 0
	}
	base := math.Min(80, (float64(impressions)/1000.0)*12.0)
	ctrBoost := math.Min(19, ctr*250.0)
	score := base + ctrBoost
	if score > 99 {
		score = 99
	}
	return math.Round(score*10) / 10
}

func shouldRefreshRecommendations(c echo.Context, workspaceID pgtype.UUID, refresh bool) (bool, error) {
	if refresh {
		return true, nil
	}

	var lastGenerated pgtype.Timestamptz
	err := dbPool.QueryRow(
		c.Request().Context(),
		`SELECT MAX(last_generated_at)
		 FROM signal_brief_recommendations
		 WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&lastGenerated)
	if err != nil {
		return false, err
	}

	if !lastGenerated.Valid {
		return true, nil
	}

	return time.Since(lastGenerated.Time.UTC()) >= 7*24*time.Hour, nil
}

func regenerateBriefRecommendations(c echo.Context, workspaceID pgtype.UUID, days int) error {
	if dbPool == nil {
		return errors.New("database_not_initialized")
	}

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT
			COALESCE(v.angle, 'unknown') AS angle,
			COALESCE(SUM(m.impressions), 0) AS impressions,
			COALESCE(SUM(m.clicks), 0) AS clicks,
			CASE WHEN COALESCE(SUM(m.impressions), 0) > 0
				THEN COALESCE(SUM(m.clicks), 0)::float8 / COALESCE(SUM(m.impressions), 0)::float8
				ELSE 0
			END AS ctr
		 FROM signal_metrics_daily m
		 LEFT JOIN variants v ON v.id = m.variant_id
		 WHERE m.workspace_id = $1
		   AND m.metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY COALESCE(v.angle, 'unknown')
		 HAVING COALESCE(SUM(m.impressions), 0) >= 1000
		 ORDER BY ctr DESC, impressions DESC
		 LIMIT 3`,
		workspaceID,
		days,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]struct {
		angle       string
		impressions int64
		ctr         float64
	}, 0)

	for rows.Next() {
		var angle string
		var impressions int64
		var clicks int64
		var ctr float64
		if scanErr := rows.Scan(&angle, &impressions, &clicks, &ctr); scanErr != nil {
			return scanErr
		}
		items = append(items, struct {
			angle       string
			impressions int64
			ctr         float64
		}{
			angle:       angle,
			impressions: impressions,
			ctr:         ctr,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}

	for _, item := range items {
		confidence := computeRecommendationConfidence(item.impressions, item.ctr)
		rationale := "Based on your recent campaigns, " + item.angle + " is outperforming with CTR " +
			fmt.Sprintf("%.2f%%", item.ctr*100) + " across " + strconv.FormatInt(item.impressions, 10) +
			" impressions in the last " + strconv.Itoa(days) + " days."

		_, execErr := dbPool.Exec(
			c.Request().Context(),
			`INSERT INTO signal_brief_recommendations (
				workspace_id, angle, suggested_hook, rationale, confidence_score, impression_volume, window_days, first_generated_at, last_generated_at
			 ) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			 ON CONFLICT (workspace_id, angle)
			 DO UPDATE SET
			   suggested_hook = EXCLUDED.suggested_hook,
			   rationale = EXCLUDED.rationale,
			   confidence_score = EXCLUDED.confidence_score,
			   impression_volume = EXCLUDED.impression_volume,
			   window_days = EXCLUDED.window_days,
			   last_generated_at = NOW(),
			   updated_at = NOW()`,
			workspaceID,
			item.angle,
			recommendedHookForAngle(item.angle),
			rationale,
			confidence,
			item.impressions,
			days,
		)
		if execErr != nil {
			return execErr
		}
	}

	return nil
}

// ListSignalConnections godoc
// GET /api/v1/signal/connections
func ListSignalConnections(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
		`UPDATE signal_connections
		 SET status = 'token_expired',
		     error_reason = COALESCE(error_reason, 'oauth_token_expired'),
		     updated_at = NOW()
		 WHERE workspace_id = $1
		   AND status = 'connected'
		   AND token_expires_at IS NOT NULL
		   AND token_expires_at < NOW()`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT platform, status, account_id, account_name, error_reason, token_expires_at, last_synced_at, created_at, updated_at
		 FROM signal_connections
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC, created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}
	defer rows.Close()

	connections := make([]map[string]any, 0)
	for rows.Next() {
		var platform string
		var status string
		var accountID string
		var accountName *string
		var errorReason *string
		var tokenExpiresAt pgtype.Timestamptz
		var lastSyncedAt pgtype.Timestamptz
		var createdAt pgtype.Timestamptz
		var updatedAt pgtype.Timestamptz
		if scanErr := rows.Scan(&platform, &status, &accountID, &accountName, &errorReason, &tokenExpiresAt, &lastSyncedAt, &createdAt, &updatedAt); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_scan_failed"})
		}

		connections = append(connections, map[string]any{
			"platform":         platform,
			"status":           status,
			"account_id":       accountID,
			"account_name":     accountName,
			"error_reason":     errorReason,
			"token_expires_at": tsTime(tokenExpiresAt),
			"last_synced_at":   tsTime(lastSyncedAt),
			"created_at":       tsTime(createdAt),
			"updated_at":       tsTime(updatedAt),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"connections":  connections,
	})
}

// UpsertSignalConnection godoc
// PUT /api/v1/signal/connections/:platform
func UpsertSignalConnection(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}

	var req struct {
		AccountID      string  `json:"account_id"`
		AccountName    *string `json:"account_name"`
		Status         string  `json:"status"`
		ErrorReason    *string `json:"error_reason"`
		TokenExpiresAt *string `json:"token_expires_at"`
		LastSyncedAt   *string `json:"last_synced_at"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "account_id_required"})
	}

	status := strings.TrimSpace(strings.ToLower(req.Status))
	if status == "" {
		status = "connected"
	}
	if status != "connected" && status != "disconnected" && status != "token_expired" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}

	var tokenExpiresAt any
	if req.TokenExpiresAt != nil && strings.TrimSpace(*req.TokenExpiresAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.TokenExpiresAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_token_expires_at"})
		}
		tokenExpiresAt = t.UTC()
	}

	var lastSyncedAt any
	if req.LastSyncedAt != nil && strings.TrimSpace(*req.LastSyncedAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.LastSyncedAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_last_synced_at"})
		}
		lastSyncedAt = t.UTC()
	}

	if status == "connected" {
		req.ErrorReason = nil
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
		`INSERT INTO signal_connections (workspace_id, platform, status, account_id, account_name, error_reason, token_expires_at, last_synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW()))
		 ON CONFLICT (workspace_id, platform, account_id)
		 DO UPDATE SET
		   status = EXCLUDED.status,
		   account_name = EXCLUDED.account_name,
		   error_reason = EXCLUDED.error_reason,
		   token_expires_at = EXCLUDED.token_expires_at,
		   last_synced_at = COALESCE(EXCLUDED.last_synced_at, signal_connections.last_synced_at),
		   updated_at = NOW()`,
		workspaceID,
		platform,
		status,
		accountID,
		req.AccountName,
		req.ErrorReason,
		tokenExpiresAt,
		lastSyncedAt,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connection_upsert_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"platform":     platform,
		"account_id":   accountID,
		"status":       status,
		"updated":      true,
	})
}

// PatchSignalConnectionHealth godoc
// PATCH /api/v1/signal/connections/:platform/:accountId/health
func PatchSignalConnectionHealth(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	accountID := strings.TrimSpace(c.Param("accountId"))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "account_id_required"})
	}

	var req struct {
		Status         *string `json:"status"`
		ErrorReason    *string `json:"error_reason"`
		TokenExpiresAt *string `json:"token_expires_at"`
		LastSyncedAt   *string `json:"last_synced_at"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	status := ""
	if req.Status != nil {
		status = strings.TrimSpace(strings.ToLower(*req.Status))
		if status != "connected" && status != "disconnected" && status != "token_expired" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
		}
	}

	var tokenExpiresAt any
	if req.TokenExpiresAt != nil && strings.TrimSpace(*req.TokenExpiresAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.TokenExpiresAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_token_expires_at"})
		}
		tokenExpiresAt = t.UTC()
	}

	var lastSyncedAt any
	if req.LastSyncedAt != nil && strings.TrimSpace(*req.LastSyncedAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.LastSyncedAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_last_synced_at"})
		}
		lastSyncedAt = t.UTC()
	}

	result, err := dbPool.Exec(
		c.Request().Context(),
		`UPDATE signal_connections
		 SET status = COALESCE(NULLIF($4, ''), status),
		     error_reason = CASE
		       WHEN COALESCE(NULLIF($4, ''), status) = 'connected' THEN NULL
		       ELSE COALESCE($5, error_reason)
		     END,
		     token_expires_at = COALESCE($6, token_expires_at),
		     last_synced_at = COALESCE($7, last_synced_at, NOW()),
		     updated_at = NOW()
		 WHERE workspace_id = $1
		   AND platform = $2
		   AND account_id = $3`,
		workspaceID,
		platform,
		accountID,
		status,
		req.ErrorReason,
		tokenExpiresAt,
		lastSyncedAt,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connection_health_patch_failed"})
	}
	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "signal_connection_not_found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"platform":     platform,
		"account_id":   accountID,
		"updated":      true,
	})
}

// UpsertSignalMetrics godoc
// POST /api/v1/signal/metrics
func UpsertSignalMetrics(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	var req struct {
		Metrics []struct {
			VariantID   *string  `json:"variant_id"`
			Platform    string   `json:"platform"`
			Date        string   `json:"date"`
			Impressions int64    `json:"impressions"`
			Clicks      int64    `json:"clicks"`
			SpendUSD    float64  `json:"spend_usd"`
			Conversions int64    `json:"conversions"`
			RevenueUSD  float64  `json:"revenue_usd"`
			Hold25      *float64 `json:"hold_25"`
			Hold50      *float64 `json:"hold_50"`
			Hold75      *float64 `json:"hold_75"`
			Hold100     *float64 `json:"hold_100"`
		} `json:"metrics"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if len(req.Metrics) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "metrics_required"})
	}

	tx, err := dbPool.Begin(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
	}
	defer tx.Rollback(c.Request().Context()) //nolint:errcheck

	rowsUpserted := 0
	for _, metric := range req.Metrics {
		platform := strings.TrimSpace(strings.ToLower(metric.Platform))
		if platform != "meta" && platform != "tiktok" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
		}

		metricDate, parseErr := time.Parse("2006-01-02", strings.TrimSpace(metric.Date))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_date"})
		}

		var variantID any
		if metric.VariantID != nil && strings.TrimSpace(*metric.VariantID) != "" {
			parsed, parseErr := parseUUID(strings.TrimSpace(*metric.VariantID))
			if parseErr != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
			}
			variantID = parsed
		}

		_, execErr := tx.Exec(
			c.Request().Context(),
			`INSERT INTO signal_metrics_daily (
				workspace_id, variant_id, platform, metric_date, impressions, clicks, spend_usd, conversions, revenue_usd, hold_25, hold_50, hold_75, hold_100
			 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			 ON CONFLICT (workspace_id, platform, variant_id, metric_date)
			 DO UPDATE SET
			   impressions = EXCLUDED.impressions,
			   clicks = EXCLUDED.clicks,
			   spend_usd = EXCLUDED.spend_usd,
			   conversions = EXCLUDED.conversions,
			   revenue_usd = EXCLUDED.revenue_usd,
			   hold_25 = EXCLUDED.hold_25,
			   hold_50 = EXCLUDED.hold_50,
			   hold_75 = EXCLUDED.hold_75,
			   hold_100 = EXCLUDED.hold_100,
			   updated_at = NOW()`,
			workspaceID,
			variantID,
			platform,
			metricDate,
			metric.Impressions,
			metric.Clicks,
			metric.SpendUSD,
			metric.Conversions,
			metric.RevenueUSD,
			metric.Hold25,
			metric.Hold50,
			metric.Hold75,
			metric.Hold100,
		)
		if execErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
		}

		rowsUpserted++
	}

	if err := tx.Commit(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"upserted":     rowsUpserted,
	})
}

// GetSignalDashboard godoc
// GET /api/v1/signal/dashboard?days=30
func GetSignalDashboard(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))

	var impressions int64
	var clicks int64
	var spendUSD float64
	var conversions int64
	var revenueUSD float64
	err = dbPool.QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(SUM(impressions), 0),
			COALESCE(SUM(clicks), 0),
			COALESCE(SUM(spend_usd), 0),
			COALESCE(SUM(conversions), 0),
			COALESCE(SUM(revenue_usd), 0)
		 FROM signal_metrics_daily
		 WHERE workspace_id = $1
		   AND metric_date >= CURRENT_DATE - ($2::int - 1)`,
		workspaceID,
		days,
	).Scan(&impressions, &clicks, &spendUSD, &conversions, &revenueUSD)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	ctr := 0.0
	if impressions > 0 {
		ctr = (float64(clicks) / float64(impressions)) * 100
	}
	cpa := 0.0
	if conversions > 0 {
		cpa = spendUSD / float64(conversions)
	}
	roas := 0.0
	if spendUSD > 0 {
		roas = revenueUSD / spendUSD
	}

	var feedbackTotal int64
	var feedbackAccepted int64
	err = dbPool.QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(COUNT(*), 0) AS feedback_total,
			COALESCE(COUNT(*) FILTER (WHERE action = 'accept'), 0) AS feedback_accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1
		   AND created_at >= NOW() - ($2::int || ' days')::interval`,
		workspaceID,
		days,
	).Scan(&feedbackTotal, &feedbackAccepted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	feedbackAcceptanceRate := 0.0
	if feedbackTotal > 0 {
		feedbackAcceptanceRate = (float64(feedbackAccepted) / float64(feedbackTotal)) * 100
	}

	var current7dTotal int64
	var current7dAccepted int64
	var previous7dTotal int64
	var previous7dAccepted int64
	err = dbPool.QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days'), 0) AS current_total,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days' AND action = 'accept'), 0) AS current_accepted,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '14 days' AND created_at < NOW() - INTERVAL '7 days'), 0) AS previous_total,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '14 days' AND created_at < NOW() - INTERVAL '7 days' AND action = 'accept'), 0) AS previous_accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&current7dTotal, &current7dAccepted, &previous7dTotal, &previous7dAccepted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	current7dRate := 0.0
	if current7dTotal > 0 {
		current7dRate = (float64(current7dAccepted) / float64(current7dTotal)) * 100
	}
	previous7dRate := 0.0
	if previous7dTotal > 0 {
		previous7dRate = (float64(previous7dAccepted) / float64(previous7dTotal)) * 100
	}
	acceptanceDeltaPctPoints := current7dRate - previous7dRate

	trendRows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT
			(created_at AT TIME ZONE 'UTC')::date AS day,
			COUNT(*)::bigint AS total,
			COUNT(*) FILTER (WHERE action = 'accept')::bigint AS accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1
		   AND created_at >= NOW() - INTERVAL '14 days'
		 GROUP BY day
		 ORDER BY day ASC`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer trendRows.Close()

	type dailyTrend struct {
		total    int64
		accepted int64
	}
	trendByDate := make(map[string]dailyTrend)
	for trendRows.Next() {
		var day time.Time
		var total int64
		var accepted int64
		if scanErr := trendRows.Scan(&day, &total, &accepted); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}
		dateKey := day.UTC().Format("2006-01-02")
		trendByDate[dateKey] = dailyTrend{total: total, accepted: accepted}
	}
	if rowsErr := trendRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	acceptanceTrend := make([]map[string]any, 0, 14)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := 13; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		dateKey := day.Format("2006-01-02")
		entry := trendByDate[dateKey]
		rate := 0.0
		if entry.total > 0 {
			rate = (float64(entry.accepted) / float64(entry.total)) * 100
		}

		acceptanceTrend = append(acceptanceTrend, map[string]any{
			"date":            dateKey,
			"total":           entry.total,
			"accepted":        entry.accepted,
			"acceptance_rate": rate,
		})
	}

	angleRows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT
			COALESCE(v.angle, 'unknown') AS angle,
			COALESCE(SUM(m.impressions), 0) AS impressions,
			COALESCE(SUM(m.clicks), 0) AS clicks,
			COALESCE(SUM(m.spend_usd), 0) AS spend_usd,
			COALESCE(SUM(m.conversions), 0) AS conversions,
			COALESCE(SUM(m.revenue_usd), 0) AS revenue_usd
		 FROM signal_metrics_daily m
		 LEFT JOIN variants v ON v.id = m.variant_id
		 WHERE m.workspace_id = $1
		   AND m.metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY COALESCE(v.angle, 'unknown')
		 ORDER BY clicks DESC, impressions DESC
		 LIMIT 10`,
		workspaceID,
		days,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer angleRows.Close()

	byAngle := make([]map[string]any, 0)
	for angleRows.Next() {
		var angle string
		var aImpressions int64
		var aClicks int64
		var aSpend float64
		var aConversions int64
		var aRevenue float64
		if scanErr := angleRows.Scan(&angle, &aImpressions, &aClicks, &aSpend, &aConversions, &aRevenue); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}

		aCtr := 0.0
		if aImpressions > 0 {
			aCtr = (float64(aClicks) / float64(aImpressions)) * 100
		}

		byAngle = append(byAngle, map[string]any{
			"angle":            angle,
			"impressions":      aImpressions,
			"clicks":           aClicks,
			"spend_usd":        aSpend,
			"conversions":      aConversions,
			"revenue_usd":      aRevenue,
			"ctr":              aCtr,
			"meets_threshold":  aImpressions >= 1000,
		})
	}
	if rowsErr := angleRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	platformRows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT
			platform,
			COALESCE(SUM(impressions), 0) AS impressions,
			COALESCE(SUM(clicks), 0) AS clicks,
			COALESCE(SUM(spend_usd), 0) AS spend_usd,
			COALESCE(SUM(conversions), 0) AS conversions,
			COALESCE(SUM(revenue_usd), 0) AS revenue_usd
		 FROM signal_metrics_daily
		 WHERE workspace_id = $1
		   AND metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY platform
		 ORDER BY impressions DESC`,
		workspaceID,
		days,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer platformRows.Close()

	byPlatform := make([]map[string]any, 0)
	for platformRows.Next() {
		var platform string
		var pImpressions int64
		var pClicks int64
		var pSpend float64
		var pConversions int64
		var pRevenue float64
		if scanErr := platformRows.Scan(&platform, &pImpressions, &pClicks, &pSpend, &pConversions, &pRevenue); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}

		pCtr := 0.0
		if pImpressions > 0 {
			pCtr = (float64(pClicks) / float64(pImpressions)) * 100
		}

		byPlatform = append(byPlatform, map[string]any{
			"platform":    platform,
			"impressions": pImpressions,
			"clicks":      pClicks,
			"spend_usd":   pSpend,
			"conversions": pConversions,
			"revenue_usd": pRevenue,
			"ctr":         pCtr,
		})
	}
	if rowsErr := platformRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"days":         days,
		"totals": map[string]any{
			"impressions": impressions,
			"clicks":      clicks,
			"spend_usd":   spendUSD,
			"conversions": conversions,
			"revenue_usd": revenueUSD,
			"ctr":         ctr,
			"cpa":         cpa,
			"roas":        roas,
		},
		"recommendation_feedback": map[string]any{
			"total":                       feedbackTotal,
			"accepted":                    feedbackAccepted,
			"acceptance_rate":             feedbackAcceptanceRate,
			"current_7d_rate":             current7dRate,
			"previous_7d_rate":            previous7dRate,
			"acceptance_delta_pct_points": acceptanceDeltaPctPoints,
			"trend":                       acceptanceTrend,
		},
		"by_angle":    byAngle,
		"by_platform": byPlatform,
	})
}

// DetectSignalFatigue godoc
// GET /api/v1/signal/fatigue?days=30&drop_pct=30&sustained_days=3&min_peak_ctr=0.01&min_impressions=1000&persist=true
func DetectSignalFatigue(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	dropPct := parsePositiveFloat(c.QueryParam("drop_pct"), 30)
	if dropPct > 95 {
		dropPct = 95
	}
	dropRatio := dropPct / 100
	sustainedDays := parsePositiveInt(c.QueryParam("sustained_days"), 3)
	if sustainedDays < 3 {
		sustainedDays = 3
	}
	if sustainedDays > 7 {
		sustainedDays = 7
	}
	minPeakCtr := parsePositiveFloat(c.QueryParam("min_peak_ctr"), 0.01)
	minImpressions := int64(parsePositiveInt(c.QueryParam("min_impressions"), 1000))
	persist := parseBool(c.QueryParam("persist"), true)

	windowURL := url.QueryEscape("" + strconv.Itoa(days) + "d")

	rows, err := dbPool.Query(
		c.Request().Context(),
		`WITH daily AS (
			SELECT
				variant_id,
				metric_date,
				COALESCE(SUM(impressions), 0) AS impressions,
				CASE WHEN COALESCE(SUM(impressions), 0) > 0
					THEN COALESCE(SUM(clicks), 0)::float8 / COALESCE(SUM(impressions), 0)::float8
					ELSE 0
				END AS ctr
			FROM signal_metrics_daily
			WHERE workspace_id = $1
			  AND variant_id IS NOT NULL
			  AND metric_date >= CURRENT_DATE - ($2::int - 1)
			GROUP BY variant_id, metric_date
		),
		scored AS (
			SELECT
				variant_id,
				metric_date,
				impressions,
				ctr,
				MAX(ctr) OVER (
					PARTITION BY variant_id
					ORDER BY metric_date
					ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) AS peak_7d
			FROM daily
		),
		flagged AS (
			SELECT
				variant_id,
				metric_date,
				impressions,
				ctr,
				peak_7d,
				CASE
					WHEN peak_7d >= $3::float8
					 AND impressions >= $4::bigint
					 AND ctr <= peak_7d * (1 - $5::float8)
					THEN 1 ELSE 0
				END AS is_drop
			FROM scored
		),
		streak AS (
			SELECT
				variant_id,
				metric_date,
				ctr,
				peak_7d,
				is_drop,
				CASE
					WHEN is_drop = 1
					 AND LAG(is_drop, 1, 0) OVER (PARTITION BY variant_id ORDER BY metric_date) = 1
					 AND LAG(is_drop, 2, 0) OVER (PARTITION BY variant_id ORDER BY metric_date) = 1
					THEN 3 ELSE 0
				END AS sustained_days
			FROM flagged
		),
		candidates AS (
			SELECT DISTINCT ON (variant_id)
				variant_id,
				metric_date AS detected_on,
				ctr AS current_ctr,
				peak_7d,
				((peak_7d - ctr) / NULLIF(peak_7d, 0)) * 100.0 AS drop_pct,
				sustained_days
			FROM streak
			WHERE sustained_days >= $6::int
			ORDER BY variant_id, metric_date DESC
		)
		SELECT
			c.variant_id,
			c.detected_on,
			c.current_ctr,
			c.peak_7d,
			COALESCE(c.drop_pct, 0),
			c.sustained_days,
			COALESCE(v.angle, 'unknown') AS angle
		FROM candidates c
		LEFT JOIN variants v ON v.id = c.variant_id
		ORDER BY c.drop_pct DESC, c.detected_on DESC`,
		workspaceID,
		days,
		minPeakCtr,
		minImpressions,
		dropRatio,
		sustainedDays,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
	}
	defer rows.Close()

	alerts := make([]map[string]any, 0)
	for rows.Next() {
		var variantID pgtype.UUID
		var detectedOn time.Time
		var currentCtr float64
		var peakCtr float64
		var dropComputed float64
		var sustained int
		var angle string
		if scanErr := rows.Scan(&variantID, &detectedOn, &currentCtr, &peakCtr, &dropComputed, &sustained, &angle); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
		}

		suggestedAction := resolveSuggestedAction(dropComputed)
		alert := map[string]any{
			"variant_id":       uuidString(variantID),
			"detected_on":      detectedOn.Format("2006-01-02"),
			"angle":            angle,
			"current_ctr":      currentCtr,
			"peak_ctr":         peakCtr,
			"drop_pct":         dropComputed,
			"sustained_days":   sustained,
			"suggested_action": suggestedAction,
			"refresh_link":     "/dashboard?refresh_variant=" + url.QueryEscape(uuidString(variantID)) + "&window=" + windowURL,
		}
		alerts = append(alerts, alert)

		if persist {
			_, execErr := dbPool.Exec(
				c.Request().Context(),
				`INSERT INTO signal_fatigue_events (
					workspace_id, variant_id, detected_on, current_ctr, peak_ctr, drop_pct, sustained_days, suggested_action, status, first_detected_at, last_detected_at
				 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', NOW(), NOW())
				 ON CONFLICT (workspace_id, variant_id, detected_on)
				 DO UPDATE SET
				   current_ctr = EXCLUDED.current_ctr,
				   peak_ctr = EXCLUDED.peak_ctr,
				   drop_pct = EXCLUDED.drop_pct,
				   sustained_days = EXCLUDED.sustained_days,
				   suggested_action = EXCLUDED.suggested_action,
				   status = 'active',
				   resolved_at = NULL,
				   last_detected_at = NOW(),
				   updated_at = NOW()`,
				workspaceID,
				variantID,
				detectedOn,
				currentCtr,
				peakCtr,
				dropComputed,
				sustained,
				suggestedAction,
			)
			if execErr != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
	}

	if persist && len(alerts) > 0 {
		_ = sendFatigueAlertEmail(c.Request().Context(), claims.OrgID, uuidString(workspaceID), alerts)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":         claims.OrgID,
		"workspace_id":   uuidString(workspaceID),
		"days":           days,
		"drop_pct":       dropPct,
		"sustained_days": sustainedDays,
		"persisted":      persist,
		"alerts":         alerts,
	})
}

// sendFatigueAlertEmail sends a fatigue alert summary via the Resend API.
// Requires RESEND_API_KEY and RESEND_FROM_EMAIL in env. Non-fatal on failure.
func sendFatigueAlertEmail(ctx context.Context, orgID, workspaceID string, alerts []map[string]any) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return nil // not configured — skip silently
	}

	if dbPool == nil {
		return nil
	}

	// Look up notification email for this workspace
	var notificationEmail string
	_ = dbPool.QueryRow(ctx,
		`SELECT COALESCE(notification_email, '') FROM workspaces WHERE org_id = $1 LIMIT 1`,
		orgID,
	).Scan(&notificationEmail)

	if strings.TrimSpace(notificationEmail) == "" {
		return nil // no recipient configured
	}

	count := len(alerts)
	subject := fmt.Sprintf("Qvora Signal: %d creative fatigue alert", count)
	if count != 1 {
		subject = fmt.Sprintf("Qvora Signal: %d creative fatigue alerts", count)
	}

	// Build simple HTML body
	var htmlBuf bytes.Buffer
	htmlBuf.WriteString(`<h2 style="color:#FF3D3D">Creative Fatigue Detected</h2>`)
	htmlBuf.WriteString(fmt.Sprintf(`<p>%d variant(s) in workspace <code>%s</code> are showing sustained CTR decline.</p>`, count, workspaceID))
	htmlBuf.WriteString(`<table border="1" cellpadding="6" style="border-collapse:collapse;font-family:monospace;font-size:12px"><thead><tr><th>Angle</th><th>Variant</th><th>Drop %%</th><th>Action</th></tr></thead><tbody>`)
	for _, alert := range alerts {
		htmlBuf.WriteString(fmt.Sprintf(
			`<tr><td>%v</td><td>%v</td><td>%.1f%%</td><td>%v</td></tr>`,
			alert["angle"], alert["variant_id"],
			toFloat64(alert["drop_pct"]),
			alert["suggested_action"],
		))
	}
	htmlBuf.WriteString(`</tbody></table>`)
	htmlBuf.WriteString(`<p style="margin-top:16px;font-size:11px;color:#888">Sent by Qvora Signal — <a href="https://app.qvora.com/dashboard">View dashboard</a></p>`)

	fromEmail := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = "signal@qvora.ai"
	}

	body, _ := json.Marshal(map[string]any{
		"from":    "Qvora Signal <" + fromEmail + ">",
		"to":      []string{notificationEmail},
		"subject": subject,
		"html":    htmlBuf.String(),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// toFloat64 extracts a float64 from an any value safely.
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}

// ListSignalFatigueEvents godoc
// GET /api/v1/signal/fatigue/events?limit=20&status=active
func ListSignalFatigueEvents(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	limit := parsePositiveInt(c.QueryParam("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	status := strings.TrimSpace(strings.ToLower(c.QueryParam("status")))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "resolved" && status != "all" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}

	query := `SELECT
		f.variant_id,
		f.detected_on,
		f.current_ctr,
		f.peak_ctr,
		f.drop_pct,
		f.sustained_days,
		f.suggested_action,
		f.status,
		f.first_detected_at,
		f.last_detected_at,
		f.resolved_at,
		COALESCE(v.angle, 'unknown') AS angle
	 FROM signal_fatigue_events f
	 LEFT JOIN variants v ON v.id = f.variant_id
	 WHERE f.workspace_id = $1`

	args := []any{workspaceID}
	if status != "all" {
		query += ` AND f.status = $2`
		args = append(args, status)
		query += ` ORDER BY f.last_detected_at DESC LIMIT $3`
		args = append(args, limit)
	} else {
		query += ` ORDER BY f.last_detected_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := dbPool.Query(c.Request().Context(), query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
	}
	defer rows.Close()

	events := make([]map[string]any, 0)
	for rows.Next() {
		var variantID pgtype.UUID
		var detectedOn time.Time
		var currentCtr float64
		var peakCtr float64
		var dropPct float64
		var sustainedDays int
		var suggestedAction string
		var evtStatus string
		var firstDetectedAt time.Time
		var lastDetectedAt time.Time
		var resolvedAt pgtype.Timestamptz
		var angle string
		if scanErr := rows.Scan(
			&variantID,
			&detectedOn,
			&currentCtr,
			&peakCtr,
			&dropPct,
			&sustainedDays,
			&suggestedAction,
			&evtStatus,
			&firstDetectedAt,
			&lastDetectedAt,
			&resolvedAt,
			&angle,
		); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
		}

		events = append(events, map[string]any{
			"variant_id":        uuidString(variantID),
			"detected_on":       detectedOn.Format("2006-01-02"),
			"angle":             angle,
			"current_ctr":       currentCtr,
			"peak_ctr":          peakCtr,
			"drop_pct":          dropPct,
			"sustained_days":    sustainedDays,
			"suggested_action":  suggestedAction,
			"status":            evtStatus,
			"first_detected_at": firstDetectedAt.UTC(),
			"last_detected_at":  lastDetectedAt.UTC(),
			"resolved_at":       tsTime(resolvedAt),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"status":       status,
		"events":       events,
	})
}

// GetSignalRecommendations godoc
// GET /api/v1/signal/recommendations?days=90&refresh=false
func GetSignalRecommendations(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	refresh := parseBool(c.QueryParam("refresh"), false)

	refreshNeeded, err := shouldRefreshRecommendations(c, workspaceID, refresh)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}
	if refreshNeeded {
		if err := regenerateBriefRecommendations(c, workspaceID, days); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
		}
	}

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT angle, suggested_hook, rationale, confidence_score, impression_volume, window_days, last_generated_at
		 FROM signal_brief_recommendations
		 WHERE workspace_id = $1
		 ORDER BY confidence_score DESC, impression_volume DESC
		 LIMIT 3`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}
	defer rows.Close()

	recommendations := make([]map[string]any, 0)
	for rows.Next() {
		var angle string
		var suggestedHook string
		var rationale string
		var confidence float64
		var impressions int64
		var windowDays int
		var lastGenerated time.Time
		if scanErr := rows.Scan(&angle, &suggestedHook, &rationale, &confidence, &impressions, &windowDays, &lastGenerated); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
		}

		recommendations = append(recommendations, map[string]any{
			"angle":             angle,
			"suggested_hook":    suggestedHook,
			"rationale":         rationale,
			"confidence_score":  confidence,
			"impression_volume": impressions,
			"window_days":       windowDays,
			"last_generated_at": lastGenerated.UTC(),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":          claims.OrgID,
		"workspace_id":    uuidString(workspaceID),
		"refreshed":       refreshNeeded,
		"recommendations": recommendations,
	})
}

// RefreshAllSignalRecommendations godoc
// POST /api/v1/internal/signal/recommendations/refresh-all
func RefreshAllSignalRecommendations(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
	if role != "worker" && role != "admin" && role != "system" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	limit := parsePositiveInt(c.QueryParam("limit"), 1000)
	if limit > 5000 {
		limit = 5000
	}

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT id, org_id
		 FROM workspaces
		 ORDER BY created_at ASC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
	}
	defer rows.Close()

	refreshed := 0
	failures := make([]map[string]any, 0)
	for rows.Next() {
		var workspaceID pgtype.UUID
		var orgID string
		if scanErr := rows.Scan(&workspaceID, &orgID); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
		}

		if regenErr := regenerateBriefRecommendations(c, workspaceID, days); regenErr != nil {
			failures = append(failures, map[string]any{
				"org_id": orgID,
				"error":  regenErr.Error(),
			})
			continue
		}

		refreshed++
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"requested_by": claims.UserID,
		"days":         days,
		"refreshed":    refreshed,
		"failures":     failures,
	})
}

// CreateSignalRecommendationFeedback godoc
// POST /api/v1/signal/recommendations/feedback
func CreateSignalRecommendationFeedback(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	var req struct {
		Angle  string `json:"angle"`
		Action string `json:"action"`
		Source string `json:"source"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	angle := strings.TrimSpace(strings.ToLower(req.Angle))
	action := strings.TrimSpace(strings.ToLower(req.Action))
	source := strings.TrimSpace(strings.ToLower(req.Source))
	if angle == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "angle_required"})
	}
	if action != "accept" && action != "ignore" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_action"})
	}
	if source == "" {
		source = "brief_create_panel"
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
		`INSERT INTO signal_recommendation_feedback (workspace_id, angle, action, source, created_by)
		 VALUES ($1, $2, $3, $4, $5)`,
		workspaceID,
		angle,
		action,
		source,
		claims.UserID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendation_feedback_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": uuidString(workspaceID),
		"angle":        angle,
		"action":       action,
		"source":       source,
		"tracked":      true,
	})
}

// ListSignalRecommendationFeedbackByAngle godoc
// GET /api/v1/signal/recommendations/feedback?days=30&limit=10
func ListSignalRecommendationFeedbackByAngle(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	limit := parsePositiveInt(c.QueryParam("limit"), 10)
	if limit > 50 {
		limit = 50
	}

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT
			angle,
			COUNT(*)::bigint AS total,
			COUNT(*) FILTER (WHERE action = 'accept')::bigint AS accepted,
			COUNT(*) FILTER (WHERE action = 'ignore')::bigint AS ignored
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1
		   AND created_at >= NOW() - ($2::int || ' days')::interval
		 GROUP BY angle
		 ORDER BY accepted DESC, total DESC
		 LIMIT $3`,
		workspaceID,
		days,
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendation_feedback_list_failed"})
	}
	defer rows.Close()

	feedbackByAngle := make([]map[string]any, 0)
	for rows.Next() {
		var angle string
		var total int64
		var accepted int64
		var ignored int64
		if scanErr := rows.Scan(&angle, &total, &accepted, &ignored); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendation_feedback_list_failed"})
		}

		rate := 0.0
		if total > 0 {
			rate = (float64(accepted) / float64(total)) * 100
		}

		feedbackByAngle = append(feedbackByAngle, map[string]any{
			"angle":           angle,
			"total":           total,
			"accepted":        accepted,
			"ignored":         ignored,
			"acceptance_rate": rate,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendation_feedback_list_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":            claims.OrgID,
		"workspace_id":      uuidString(workspaceID),
		"days":              days,
		"feedback_by_angle": feedbackByAngle,
	})
}
