package identity

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

type brandKitRecord struct {
	ID               pgtype.UUID
	Name             string
	LogoR2Key        *string
	PrimaryColor     string
	SecondaryColor   *string
	FontFamily       *string
	WatermarkEnabled bool
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
}

func getLatestBrandKit(c echo.Context, workspaceID pgtype.UUID) (*brandKitRecord, error) {
	if store.Pool() == nil {
		return nil, errors.New("database_not_initialized")
	}
	row := store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT id, name, logo_r2_key, primary_color, secondary_color, font_family, watermark_enabled, created_at, updated_at
		 FROM brands
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC, created_at DESC
		 LIMIT 1`,
		workspaceID,
	)
	var out brandKitRecord
	if err := row.Scan(
		&out.ID,
		&out.Name,
		&out.LogoR2Key,
		&out.PrimaryColor,
		&out.SecondaryColor,
		&out.FontFamily,
		&out.WatermarkEnabled,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

// GetWorkspace handles GET /api/v1/workspaces/:orgId
func GetWorkspace(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	workspaceID, planTier, subStatus, trialEndsAt, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"workspace_id":        util.UUIDString(workspaceID),
		"org_id":              orgID,
		"plan_tier":           planTier,
		"subscription_status": subStatus,
		"trial_ends_at":       util.TsTime(trialEndsAt),
	})
}

// GetBrandKit handles GET /api/v1/workspaces/:orgId/brand-kit
func GetBrandKit(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	brand, err := getLatestBrandKit(c, workspaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brand_kit_lookup_failed"})
	}
	if brand == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "brand_kit_not_found"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":                util.UUIDString(brand.ID),
		"workspace_id":      util.UUIDString(workspaceID),
		"name":              brand.Name,
		"logo_r2_key":       brand.LogoR2Key,
		"primary_color":     brand.PrimaryColor,
		"secondary_color":   brand.SecondaryColor,
		"font_family":       brand.FontFamily,
		"watermark_enabled": brand.WatermarkEnabled,
		"created_at":        util.TsTime(brand.CreatedAt),
		"updated_at":        util.TsTime(brand.UpdatedAt),
	})
}

// UpsertBrandKit handles PUT /api/v1/workspaces/:orgId/brand-kit
func UpsertBrandKit(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	var req struct {
		Name             string `json:"name"`
		PrimaryColor     string `json:"primary_color"`
		SecondaryColor   string `json:"secondary_color,omitempty"`
		FontFamily       string `json:"font_family,omitempty"`
		WatermarkEnabled bool   `json:"watermark_enabled"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if strings.TrimSpace(req.Name) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name_required"})
	}
	if strings.TrimSpace(req.PrimaryColor) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "primary_color_required"})
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	brand, err := getLatestBrandKit(c, workspaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brand_kit_lookup_failed"})
	}
	trimmedName := strings.TrimSpace(req.Name)
	trimmedPrimaryColor := strings.TrimSpace(req.PrimaryColor)
	var secondaryColor *string
	if v := strings.TrimSpace(req.SecondaryColor); v != "" {
		secondaryColor = &v
	}
	var fontFamily *string
	if v := strings.TrimSpace(req.FontFamily); v != "" {
		fontFamily = &v
	}
	if brand == nil {
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`INSERT INTO brands (workspace_id, name, primary_color, secondary_color, font_family, watermark_enabled)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			workspaceID, trimmedName, trimmedPrimaryColor, secondaryColor, fontFamily, req.WatermarkEnabled,
		)
	} else {
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`UPDATE brands
			 SET name = $2,
			     primary_color = $3,
			     secondary_color = $4,
			     font_family = $5,
			     watermark_enabled = $6,
			     updated_at = NOW()
			 WHERE id = $1`,
			brand.ID, trimmedName, trimmedPrimaryColor, secondaryColor, fontFamily, req.WatermarkEnabled,
		)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brand_kit_upsert_failed"})
	}
	updated, err := getLatestBrandKit(c, workspaceID)
	if err != nil || updated == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brand_kit_read_after_write_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":                util.UUIDString(updated.ID),
		"org_id":            orgID,
		"workspace_id":      util.UUIDString(workspaceID),
		"name":              updated.Name,
		"logo_r2_key":       updated.LogoR2Key,
		"primary_color":     updated.PrimaryColor,
		"secondary_color":   updated.SecondaryColor,
		"font_family":       updated.FontFamily,
		"watermark_enabled": updated.WatermarkEnabled,
		"created_at":        util.TsTime(updated.CreatedAt),
		"updated_at":        util.TsTime(updated.UpdatedAt),
	})
}
