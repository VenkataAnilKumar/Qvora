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

// ListAssets handles GET /api/v1/assets
func ListAssets(c echo.Context) error {
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
		`SELECT id, job_id, angle, status, mux_asset_id, mux_playback_id, r2_key, duration_secs, created_at, updated_at
		 FROM variants
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC, created_at DESC
		 LIMIT 100`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "asset_list_failed"})
	}
	defer rows.Close()

	assets := make([]map[string]any, 0)
	for rows.Next() {
		var id pgtype.UUID
		var jobID pgtype.UUID
		var angle string
		var status string
		var muxAssetID *string
		var muxPlaybackID *string
		var r2Key *string
		var durationSecs pgtype.Numeric
		var createdAt pgtype.Timestamptz
		var updatedAt pgtype.Timestamptz

		if scanErr := rows.Scan(
			&id, &jobID, &angle, &status,
			&muxAssetID, &muxPlaybackID, &r2Key,
			&durationSecs, &createdAt, &updatedAt,
		); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "asset_scan_failed"})
		}
		assets = append(assets, map[string]any{
			"id":              util.UUIDString(id),
			"job_id":          util.UUIDString(jobID),
			"angle":           angle,
			"status":          status,
			"mux_asset_id":    muxAssetID,
			"mux_playback_id": muxPlaybackID,
			"r2_key":          r2Key,
			"duration_secs":   durationSecs,
			"created_at":      util.TsTime(createdAt),
			"updated_at":      util.TsTime(updatedAt),
		})
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "asset_list_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"assets": assets,
		"org_id": claims.OrgID,
	})
}

// DeleteAsset handles DELETE /api/v1/assets/:id
func DeleteAsset(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	assetID := c.Param("id")
	if strings.TrimSpace(assetID) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "asset_id_required"})
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	parsedAssetID, err := util.ParseUUID(assetID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_asset_id"})
	}
	result, err := store.Pool().Exec(
		c.Request().Context(),
		`DELETE FROM variants WHERE id = $1 AND workspace_id = $2`,
		parsedAssetID, workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "asset_delete_failed"})
	}
	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "asset_not_found"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"deleted": assetID,
		"org_id":  claims.OrgID,
	})
}
