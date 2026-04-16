package asset

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// ListExports handles GET /api/v1/exports
func ListExports(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT id, variant_id, destination, status, created_at, updated_at
		 FROM exports
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC, created_at DESC
		 LIMIT 100`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "export_list_failed"})
	}
	defer rows.Close()

	exports := make([]map[string]any, 0)
	for rows.Next() {
		var id pgtype.UUID
		var variantID pgtype.UUID
		var destination string
		var status string
		var createdAt pgtype.Timestamptz
		var updatedAt pgtype.Timestamptz

		if scanErr := rows.Scan(
			&id, &variantID, &destination, &status, &createdAt, &updatedAt,
		); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "export_scan_failed"})
		}
		exports = append(exports, map[string]any{
			"id":          util.UUIDString(id),
			"variant_id":  util.UUIDString(variantID),
			"destination": destination,
			"status":      status,
			"created_at":  util.TsTime(createdAt),
			"updated_at":  util.TsTime(updatedAt),
		})
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "export_list_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"org_id":  claims.OrgID,
		"exports": exports,
	})
}
