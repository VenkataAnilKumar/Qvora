package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

// scheduleSignalGDPRPurge marks a workspace for signal data deletion 90 days
// after subscription cancellation. Called from UpdateWorkspaceLifecycle.
func scheduleSignalGDPRPurge(c echo.Context, orgID string) {
	if dbPool == nil {
		return
	}
	// Ensure column exists (idempotent)
	_, _ = dbPool.Exec(c.Request().Context(),
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS signal_purge_at TIMESTAMPTZ`)

	purgeAt := time.Now().UTC().Add(90 * 24 * time.Hour)
	_, _ = dbPool.Exec(c.Request().Context(),
		`UPDATE workspaces
		 SET signal_purge_at = $1, updated_at = NOW()
		 WHERE org_id = $2 AND signal_purge_at IS NULL`,
		purgeAt, orgID,
	)
}

// HandleSignalGDPRCleanup deletes signal data for workspaces whose purge date
// has passed. Runs daily via the worker scheduler.
// POST /api/v1/internal/signal/gdpr/cleanup
func HandleSignalGDPRCleanup(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	role := strings.ToLower(strings.TrimSpace(claims.OrgRole))
	if role != "worker" && role != "system" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	// Ensure column exists before querying
	_, _ = dbPool.Exec(c.Request().Context(),
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS signal_purge_at TIMESTAMPTZ`)

	rows, err := dbPool.Query(
		c.Request().Context(),
		`SELECT id FROM workspaces WHERE signal_purge_at IS NOT NULL AND signal_purge_at <= NOW()`,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "gdpr_query_failed"})
	}
	defer rows.Close()

	var workspaceIDs []pgtype.UUID
	for rows.Next() {
		var id pgtype.UUID
		if scanErr := rows.Scan(&id); scanErr != nil {
			continue
		}
		workspaceIDs = append(workspaceIDs, id)
	}
	rows.Close()

	signalTables := []string{
		"signal_recommendation_feedback",
		"signal_brief_recommendations",
		"signal_fatigue_events",
		"signal_metrics_daily",
		"signal_connections",
	}

	deletedCount := 0
	for _, wsID := range workspaceIDs {
		for _, table := range signalTables {
			_, _ = dbPool.Exec(c.Request().Context(),
				`DELETE FROM `+table+` WHERE workspace_id = $1`, wsID)
		}
		// Mark purge as executed so it does not run again
		_, _ = dbPool.Exec(c.Request().Context(),
			`UPDATE workspaces SET signal_purge_at = NULL, updated_at = NOW() WHERE id = $1`,
			wsID,
		)
		deletedCount++
	}

	return c.JSON(http.StatusOK, map[string]any{
		"deleted_workspaces": deletedCount,
		"message":            "gdpr signal cleanup complete",
	})
}
