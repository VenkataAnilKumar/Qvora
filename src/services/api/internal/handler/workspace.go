package handler

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

type inMemoryWorkspace struct {
	OrgID               string `json:"org_id"`
	PlanTier            string `json:"plan_tier"`
	SubscriptionStatus  string `json:"subscription_status"`
}

var workspaceStore = struct {
	sync.RWMutex
	byOrgID map[string]inMemoryWorkspace
}{
	byOrgID: make(map[string]inMemoryWorkspace),
}

func getOrCreateWorkspace(orgID string) inMemoryWorkspace {
	workspaceStore.RLock()
	workspace, ok := workspaceStore.byOrgID[orgID]
	workspaceStore.RUnlock()
	if ok {
		return workspace
	}

	workspace = inMemoryWorkspace{
		OrgID:              orgID,
		PlanTier:           "starter",
		SubscriptionStatus: "trialing",
	}

	workspaceStore.Lock()
	workspaceStore.byOrgID[orgID] = workspace
	workspaceStore.Unlock()

	return workspace
}

func getWorkspacePlanTier(orgID string) string {
	return getOrCreateWorkspace(orgID).PlanTier
}

// GetWorkspace godoc
// GET /api/v1/workspaces/:orgId
func GetWorkspace(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	orgID := c.Param("orgId")

	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	workspace := getOrCreateWorkspace(orgID)
	return c.JSON(http.StatusOK, map[string]any{
		"org_id":              workspace.OrgID,
		"plan_tier":           workspace.PlanTier,
		"subscription_status": workspace.SubscriptionStatus,
	})
}

// UpsertBrandKit godoc
// PUT /api/v1/workspaces/:orgId/brand-kit
func UpsertBrandKit(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
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

	// TODO: upsert via sqlc
	return c.JSON(http.StatusOK, map[string]any{
		"org_id": orgID,
		"name":   req.Name,
	})
}
