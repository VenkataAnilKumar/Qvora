package signal

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// HandleSignalGDPRCleanup handles POST /api/v1/internal/signal/gdpr-cleanup
// Worker/system only: deletes signal data for workspaces past their purge date.
func HandleSignalGDPRCleanup(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
	if role != "worker" && role != "admin" && role != "system" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	// Ensure signal_purge_at column exists (added in Phase 2 oauth flow).
	_, _ = store.Pool().Exec(
		c.Request().Context(),
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS signal_purge_at TIMESTAMPTZ`,
	)

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT id FROM workspaces WHERE signal_purge_at IS NOT NULL AND signal_purge_at <= NOW()`,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_gdpr_cleanup_failed"})
	}
	defer rows.Close()

	type purgedEntry struct {
		WorkspaceID string `json:"workspace_id"`
	}
	purged := make([]purgedEntry, 0)

	for rows.Next() {
		var workspaceID pgtype.UUID
		if scanErr := rows.Scan(&workspaceID); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_gdpr_cleanup_failed"})
		}

		for _, table := range []string{
			"signal_recommendation_feedback",
			"signal_brief_recommendations",
			"signal_fatigue_events",
			"signal_metrics_daily",
			"signal_connections",
		} {
			_, execErr := store.Pool().Exec(
				c.Request().Context(),
				"DELETE FROM "+table+" WHERE workspace_id = $1",
				workspaceID,
			)
			if execErr != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_gdpr_cleanup_failed"})
			}
		}

		_, execErr := store.Pool().Exec(
			c.Request().Context(),
			`UPDATE workspaces SET signal_purge_at = NULL, updated_at = NOW() WHERE id = $1`,
			workspaceID,
		)
		if execErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_gdpr_cleanup_failed"})
		}

		purged = append(purged, purgedEntry{WorkspaceID: util.UUIDString(workspaceID)})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_gdpr_cleanup_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"purged": purged,
		"count":  len(purged),
	})
}
