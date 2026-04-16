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

func ensureWorkspaceMembersTable(c echo.Context) error {
	if store.Pool() == nil {
		return errors.New("database_not_initialized")
	}
	_, err := store.Pool().Exec(
		c.Request().Context(),
		`CREATE TABLE IF NOT EXISTS workspace_members (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			clerk_membership_id TEXT NOT NULL,
			clerk_user_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, clerk_membership_id)
		)`,
	)
	return err
}

func ensureWorkspaceUsageTable(c echo.Context) error {
	if store.Pool() == nil {
		return errors.New("database_not_initialized")
	}
	_, err := store.Pool().Exec(
		c.Request().Context(),
		`CREATE TABLE IF NOT EXISTS workspace_usage (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
			used_variants INTEGER NOT NULL DEFAULT 0,
			last_reset_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_invoice_id TEXT,
			stripe_sub_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	)
	if err != nil {
		return err
	}
	_, err = store.Pool().Exec(
		c.Request().Context(),
		`CREATE TABLE IF NOT EXISTS workspace_usage_events (
			id BIGSERIAL PRIMARY KEY,
			workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			variant_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(workspace_id, variant_id)
		)`,
	)
	return err
}

func incrementWorkspaceUsageForVariant(c echo.Context, workspaceID pgtype.UUID, variantID pgtype.UUID) (usedVariants int64, incremented bool, err error) {
	return IncrementWorkspaceUsageForVariant(c, workspaceID, variantID)
}

// IncrementWorkspaceUsageForVariant increments per-workspace variant usage,
// deduplicating by variantID. Exported for use by other domain packages
// (e.g. media/webhook.go) that cannot call the private handler bridge.
func IncrementWorkspaceUsageForVariant(c echo.Context, workspaceID pgtype.UUID, variantID pgtype.UUID) (usedVariants int64, incremented bool, err error) {
	if store.Pool() == nil {
		return 0, false, errors.New("database_not_initialized")
	}
	if err := ensureWorkspaceUsageTable(c); err != nil {
		return 0, false, err
	}
	tx, err := store.Pool().Begin(c.Request().Context())
	if err != nil {
		return 0, false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(c.Request().Context())
		}
	}()
	insertResult, err := tx.Exec(
		c.Request().Context(),
		`INSERT INTO workspace_usage_events (workspace_id, variant_id)
		 VALUES ($1, $2)
		 ON CONFLICT (workspace_id, variant_id) DO NOTHING`,
		workspaceID, variantID,
	)
	if err != nil {
		return 0, false, err
	}
	incremented = insertResult.RowsAffected() > 0
	if incremented {
		_, err = tx.Exec(
			c.Request().Context(),
			`INSERT INTO workspace_usage (workspace_id, used_variants, last_reset_at)
			 VALUES ($1, 1, NOW())
			 ON CONFLICT (workspace_id)
			 DO UPDATE SET
			   used_variants = workspace_usage.used_variants + 1,
			   updated_at = NOW()`,
			workspaceID,
		)
		if err != nil {
			return 0, false, err
		}
	}
	if scanErr := tx.QueryRow(
		c.Request().Context(),
		`SELECT used_variants FROM workspace_usage WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&usedVariants); scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			usedVariants = 0
		} else {
			err = scanErr
			return 0, false, err
		}
	}
	err = tx.Commit(c.Request().Context())
	if err != nil {
		return 0, false, err
	}
	return usedVariants, incremented, nil
}

// GetWorkspaceUsage handles GET /api/v1/workspaces/:orgId/usage
func GetWorkspaceUsage(c echo.Context) error {
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
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureWorkspaceUsageTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_usage_table_init_failed"})
	}
	var usedVariants int64
	var lastInvoiceID *string
	var stripeSubID *string
	var lastResetAt pgtype.Timestamptz
	err = store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT used_variants, last_invoice_id, stripe_sub_id, last_reset_at
		 FROM workspace_usage
		 WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&usedVariants, &lastInvoiceID, &stripeSubID, &lastResetAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			usedVariants = 0
			lastResetAt = pgtype.Timestamptz{}
		} else {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "usage_lookup_failed"})
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"org_id":                 orgID,
		"workspace_id":           util.UUIDString(workspaceID),
		"used_variants":          usedVariants,
		"last_invoice_id":        lastInvoiceID,
		"stripe_subscription_id": stripeSubID,
		"last_reset_at":          util.TsTime(lastResetAt),
	})
}

// SyncWorkspaceMembership handles PATCH /api/v1/workspaces/:orgId/memberships/sync
func SyncWorkspaceMembership(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	var req struct {
		Event        string `json:"event"`
		MembershipID string `json:"membership_id"`
		UserID       string `json:"user_id"`
		Role         string `json:"role"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	eventType := strings.TrimSpace(strings.ToLower(req.Event))
	membershipID := strings.TrimSpace(req.MembershipID)
	userID := strings.TrimSpace(req.UserID)
	role := strings.TrimSpace(strings.ToLower(req.Role))
	if membershipID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "membership_id_required"})
	}
	if eventType == "organizationmembership.created" && userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id_required"})
	}
	if role == "" {
		role = "member"
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureWorkspaceMembersTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_members_table_init_failed"})
	}
	switch eventType {
	case "organizationmembership.created":
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`INSERT INTO workspace_members (workspace_id, clerk_membership_id, clerk_user_id, role, status)
			 VALUES ($1, $2, $3, $4, 'active')
			 ON CONFLICT (workspace_id, clerk_membership_id)
			 DO UPDATE SET
			   clerk_user_id = EXCLUDED.clerk_user_id,
			   role = EXCLUDED.role,
			   status = 'active',
			   updated_at = NOW()`,
			workspaceID, membershipID, userID, role,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "membership_sync_failed"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"org_id":        orgID,
			"workspace_id":  util.UUIDString(workspaceID),
			"event":         eventType,
			"membership_id": membershipID,
			"user_id":       userID,
			"role":          role,
			"member_status": "active",
		})
	case "organizationmembership.deleted":
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`UPDATE workspace_members
			 SET status = 'removed',
			     updated_at = NOW()
			 WHERE workspace_id = $1 AND clerk_membership_id = $2`,
			workspaceID, membershipID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "membership_sync_failed"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"org_id":        orgID,
			"workspace_id":  util.UUIDString(workspaceID),
			"event":         eventType,
			"membership_id": membershipID,
			"member_status": "removed",
		})
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_event"})
	}
}

// ResetWorkspaceUsage handles PATCH /api/v1/workspaces/:orgId/usage/reset
func ResetWorkspaceUsage(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	var req struct {
		InvoiceID          string `json:"invoice_id"`
		StripeSubscription string `json:"stripe_subscription_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	invoiceID := strings.TrimSpace(req.InvoiceID)
	if invoiceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invoice_id_required"})
	}
	stripeSubID := strings.TrimSpace(req.StripeSubscription)
	var stripeSubIDPtr *string
	if stripeSubID != "" {
		stripeSubIDPtr = &stripeSubID
	}
	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureWorkspaceUsageTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_usage_table_init_failed"})
	}
	_, err = store.Pool().Exec(
		c.Request().Context(),
		`INSERT INTO workspace_usage (workspace_id, used_variants, last_reset_at, last_invoice_id, stripe_sub_id)
		 VALUES ($1, 0, NOW(), $2, $3)
		 ON CONFLICT (workspace_id)
		 DO UPDATE SET
		   used_variants = 0,
		   last_reset_at = NOW(),
		   last_invoice_id = EXCLUDED.last_invoice_id,
		   stripe_sub_id = COALESCE(EXCLUDED.stripe_sub_id, workspace_usage.stripe_sub_id),
		   updated_at = NOW()`,
		workspaceID, invoiceID, stripeSubIDPtr,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "usage_reset_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"org_id":                 orgID,
		"workspace_id":           util.UUIDString(workspaceID),
		"invoice_id":             invoiceID,
		"stripe_subscription_id": stripeSubIDPtr,
		"used_variants":          0,
		"reset":                  true,
	})
}
