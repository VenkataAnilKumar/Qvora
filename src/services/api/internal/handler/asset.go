package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

// ListAssets godoc
// GET /api/v1/assets
func ListAssets(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)

	// TODO: paginate from PostgreSQL, generate Mux signed playback URLs
	return c.JSON(http.StatusOK, map[string]any{
		"assets": []any{},
		"org_id": claims.OrgID,
	})
}

// DeleteAsset godoc
// DELETE /api/v1/assets/:id
func DeleteAsset(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	assetID := c.Param("id")

	// TODO: verify ownership, delete from R2 + Mux + PostgreSQL
	return c.JSON(http.StatusOK, map[string]any{
		"deleted": assetID,
		"org_id":  claims.OrgID,
	})
}
