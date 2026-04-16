package identity

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// UpdateWorkspaceSubscription handles PATCH /api/v1/workspaces/:orgId/subscription
func UpdateWorkspaceSubscription(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	var req struct {
		PlanTier           string `json:"plan_tier"`
		SubscriptionStatus string `json:"subscription_status"`
		StripeSubscription string `json:"stripe_subscription_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	planTier := strings.TrimSpace(strings.ToLower(req.PlanTier))
	if planTier == "" {
		planTier = "starter"
	}
	if planTier != "starter" && planTier != "growth" && planTier != "agency" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_plan_tier"})
	}
	subStatus := strings.TrimSpace(strings.ToLower(req.SubscriptionStatus))
	if subStatus == "" {
		subStatus = "trialing"
	}
	if subStatus != "trialing" && subStatus != "active" && subStatus != "past_due" && subStatus != "canceled" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_subscription_status"})
	}
	workspaceID, _, _, trialEndsAt, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	var stripeSubID *string
	if v := strings.TrimSpace(req.StripeSubscription); v != "" {
		stripeSubID = &v
	}
	_, err = store.Pool().Exec(
		c.Request().Context(),
		`UPDATE workspaces
		 SET plan_tier = $2,
		     sub_status = $3,
		     stripe_sub_id = $4,
		     updated_at = NOW()
		 WHERE id = $1`,
		workspaceID, planTier, subStatus, stripeSubID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "subscription_update_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"workspace_id":        util.UUIDString(workspaceID),
		"org_id":              orgID,
		"plan_tier":           planTier,
		"subscription_status": subStatus,
		"stripe_subscription": stripeSubID,
		"trial_ends_at":       util.TsTime(trialEndsAt),
	})
}

// UpdateWorkspaceLifecycle handles PATCH /api/v1/workspaces/:orgId/lifecycle
func UpdateWorkspaceLifecycle(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	var req struct {
		Event string `json:"event"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	eventType := strings.TrimSpace(strings.ToLower(req.Event))
	if eventType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "event_required"})
	}
	switch eventType {
	case "organization.created":
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
			"lifecycle_event":     eventType,
		})
	case "organization.deleted":
		q, err := store.Queries(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
		}
		if store.Pool() == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
		}
		workspace, err := q.GetWorkspaceByOrgID(c.Request().Context(), orgID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "workspace_not_found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_lookup_failed"})
		}
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`UPDATE workspaces
			 SET sub_status = 'canceled',
			     updated_at = NOW()
			 WHERE id = $1`,
			workspace.ID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_cancel_failed"})
		}
		scheduleSignalGDPRPurge(c, orgID)
		updated, err := q.GetWorkspaceByOrgID(c.Request().Context(), orgID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_read_after_write_failed"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"workspace_id":        util.UUIDString(updated.ID),
			"org_id":              orgID,
			"plan_tier":           updated.PlanTier,
			"subscription_status": updated.SubStatus,
			"trial_ends_at":       util.TsTime(updated.TrialEndsAt),
			"lifecycle_event":     eventType,
		})
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_event"})
	}
}

// scheduleSignalGDPRPurge marks a workspace for signal data deletion 90 days
// after subscription cancellation. Declared here to avoid domain/signal → domain/identity cycle.
func scheduleSignalGDPRPurge(c echo.Context, orgID string) {
	if store.Pool() == nil {
		return
	}
	_, _ = store.Pool().Exec(c.Request().Context(),
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS signal_purge_at TIMESTAMPTZ`)
	purgeAt := time.Now().UTC().Add(90 * 24 * time.Hour)
	_, _ = store.Pool().Exec(c.Request().Context(),
		`UPDATE workspaces
		 SET signal_purge_at = $1, updated_at = NOW()
		 WHERE org_id = $2 AND signal_purge_at IS NULL`,
		purgeAt, orgID,
	)
}
