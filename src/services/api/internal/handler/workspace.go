package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
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

func getWorkspaceForOrg(c echo.Context, orgID string) (workspaceID pgtype.UUID, planTier string, subStatus string, trialEndsAt pgtype.Timestamptz, err error) {
	q, err := queries(c.Request().Context())
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}

	workspaceID, err = workspaceIDForOrg(c.Request().Context(), q, orgID)
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}

	workspace, err := q.GetWorkspaceByOrgID(c.Request().Context(), orgID)
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}

	return workspaceID, workspace.PlanTier, workspace.SubStatus, workspace.TrialEndsAt, nil
}

func getLatestBrandKit(c echo.Context, workspaceID pgtype.UUID) (*brandKitRecord, error) {
	if dbPool == nil {
		return nil, errors.New("database_not_initialized")
	}

	row := dbPool.QueryRow(
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

// GetWorkspace godoc
// GET /api/v1/workspaces/:orgId
func GetWorkspace(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	orgID := c.Param("orgId")

	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	workspaceID, planTier, subStatus, trialEndsAt, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workspace_id":        uuidString(workspaceID),
		"org_id":              orgID,
		"plan_tier":           planTier,
		"subscription_status": subStatus,
		"trial_ends_at":       tsTime(trialEndsAt),
	})
}

// GetBrandKit godoc
// GET /api/v1/workspaces/:orgId/brand-kit
func GetBrandKit(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, orgID)
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
		"id":                uuidString(brand.ID),
		"workspace_id":      uuidString(workspaceID),
		"name":              brand.Name,
		"logo_r2_key":       brand.LogoR2Key,
		"primary_color":     brand.PrimaryColor,
		"secondary_color":   brand.SecondaryColor,
		"font_family":       brand.FontFamily,
		"watermark_enabled": brand.WatermarkEnabled,
		"created_at":        tsTime(brand.CreatedAt),
		"updated_at":        tsTime(brand.UpdatedAt),
	})
}

// UpsertBrandKit godoc
// PUT /api/v1/workspaces/:orgId/brand-kit
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

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
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
		_, err = dbPool.Exec(
			c.Request().Context(),
			`INSERT INTO brands (workspace_id, name, primary_color, secondary_color, font_family, watermark_enabled)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			workspaceID,
			trimmedName,
			trimmedPrimaryColor,
			secondaryColor,
			fontFamily,
			req.WatermarkEnabled,
		)
	} else {
		_, err = dbPool.Exec(
			c.Request().Context(),
			`UPDATE brands
			 SET name = $2,
			     primary_color = $3,
			     secondary_color = $4,
			     font_family = $5,
			     watermark_enabled = $6,
			     updated_at = NOW()
			 WHERE id = $1`,
			brand.ID,
			trimmedName,
			trimmedPrimaryColor,
			secondaryColor,
			fontFamily,
			req.WatermarkEnabled,
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
		"id":                uuidString(updated.ID),
		"org_id":            orgID,
		"workspace_id":      uuidString(workspaceID),
		"name":              updated.Name,
		"logo_r2_key":       updated.LogoR2Key,
		"primary_color":     updated.PrimaryColor,
		"secondary_color":   updated.SecondaryColor,
		"font_family":       updated.FontFamily,
		"watermark_enabled": updated.WatermarkEnabled,
		"created_at":        tsTime(updated.CreatedAt),
		"updated_at":        tsTime(updated.UpdatedAt),
	})
}

// UpdateWorkspaceSubscription godoc
// PATCH /api/v1/workspaces/:orgId/subscription
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

	workspaceID, _, _, trialEndsAt, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	var stripeSubID *string
	if v := strings.TrimSpace(req.StripeSubscription); v != "" {
		stripeSubID = &v
	}

	_, err = dbPool.Exec(
		c.Request().Context(),
		`UPDATE workspaces
		 SET plan_tier = $2,
		     sub_status = $3,
		     stripe_sub_id = $4,
		     updated_at = NOW()
		 WHERE id = $1`,
		workspaceID,
		planTier,
		subStatus,
		stripeSubID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "subscription_update_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"workspace_id":        uuidString(workspaceID),
		"org_id":              orgID,
		"plan_tier":           planTier,
		"subscription_status": subStatus,
		"stripe_subscription": stripeSubID,
		"trial_ends_at":       tsTime(trialEndsAt),
	})
}

// UpdateWorkspaceLifecycle godoc
// PATCH /api/v1/workspaces/:orgId/lifecycle
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
		workspaceID, planTier, subStatus, trialEndsAt, err := getWorkspaceForOrg(c, orgID)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"workspace_id":        uuidString(workspaceID),
			"org_id":              orgID,
			"plan_tier":           planTier,
			"subscription_status": subStatus,
			"trial_ends_at":       tsTime(trialEndsAt),
			"lifecycle_event":     eventType,
		})

	case "organization.deleted":
		q, err := queries(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
		}
		if dbPool == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
		}

		workspace, err := q.GetWorkspaceByOrgID(c.Request().Context(), orgID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "workspace_not_found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_lookup_failed"})
		}

		_, err = dbPool.Exec(
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

		// Schedule signal data purge 90 days after cancellation (SIGNAL-09)
		scheduleSignalGDPRPurge(c, orgID)

		updated, err := q.GetWorkspaceByOrgID(c.Request().Context(), orgID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_read_after_write_failed"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"workspace_id":        uuidString(updated.ID),
			"org_id":              orgID,
			"plan_tier":           updated.PlanTier,
			"subscription_status": updated.SubStatus,
			"trial_ends_at":       tsTime(updated.TrialEndsAt),
			"lifecycle_event":     eventType,
		})
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_event"})
	}
}

func ensureWorkspaceMembersTable(c echo.Context) error {
	if dbPool == nil {
		return errors.New("database_not_initialized")
	}

	_, err := dbPool.Exec(
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
	if dbPool == nil {
		return errors.New("database_not_initialized")
	}

	_, err := dbPool.Exec(
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

	_, err = dbPool.Exec(
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
	if dbPool == nil {
		return 0, false, errors.New("database_not_initialized")
	}
	if err := ensureWorkspaceUsageTable(c); err != nil {
		return 0, false, err
	}

	tx, err := dbPool.Begin(c.Request().Context())
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
		workspaceID,
		variantID,
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

// GetWorkspaceUsage godoc
// GET /api/v1/workspaces/:orgId/usage
func GetWorkspaceUsage(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	orgID := c.Param("orgId")
	if orgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := ensureWorkspaceUsageTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_usage_table_init_failed"})
	}

	var usedVariants int64
	var lastInvoiceID *string
	var stripeSubID *string
	var lastResetAt pgtype.Timestamptz

	err = dbPool.QueryRow(
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
		"org_id":                orgID,
		"workspace_id":          uuidString(workspaceID),
		"used_variants":         usedVariants,
		"last_invoice_id":       lastInvoiceID,
		"stripe_subscription_id": stripeSubID,
		"last_reset_at":         tsTime(lastResetAt),
	})
}

// SyncWorkspaceMembership godoc
// PATCH /api/v1/workspaces/:orgId/memberships/sync
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

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := ensureWorkspaceMembersTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_members_table_init_failed"})
	}

	switch eventType {
	case "organizationmembership.created":
		_, err = dbPool.Exec(
			c.Request().Context(),
			`INSERT INTO workspace_members (workspace_id, clerk_membership_id, clerk_user_id, role, status)
			 VALUES ($1, $2, $3, $4, 'active')
			 ON CONFLICT (workspace_id, clerk_membership_id)
			 DO UPDATE SET
			   clerk_user_id = EXCLUDED.clerk_user_id,
			   role = EXCLUDED.role,
			   status = 'active',
			   updated_at = NOW()`,
			workspaceID,
			membershipID,
			userID,
			role,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "membership_sync_failed"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"org_id":         orgID,
			"workspace_id":   uuidString(workspaceID),
			"event":          eventType,
			"membership_id":  membershipID,
			"user_id":        userID,
			"role":           role,
			"member_status":  "active",
		})

	case "organizationmembership.deleted":
		_, err = dbPool.Exec(
			c.Request().Context(),
			`UPDATE workspace_members
			 SET status = 'removed',
			     updated_at = NOW()
			 WHERE workspace_id = $1 AND clerk_membership_id = $2`,
			workspaceID,
			membershipID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "membership_sync_failed"})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"org_id":         orgID,
			"workspace_id":   uuidString(workspaceID),
			"event":          eventType,
			"membership_id":  membershipID,
			"member_status":  "removed",
		})
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_event"})
	}
}

// ResetWorkspaceUsage godoc
// PATCH /api/v1/workspaces/:orgId/usage/reset
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

	workspaceID, _, _, _, err := getWorkspaceForOrg(c, orgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if dbPool == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := ensureWorkspaceUsageTable(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_usage_table_init_failed"})
	}

	_, err = dbPool.Exec(
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
		workspaceID,
		invoiceID,
		stripeSubIDPtr,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "usage_reset_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":                orgID,
		"workspace_id":          uuidString(workspaceID),
		"invoice_id":            invoiceID,
		"stripe_subscription_id": stripeSubIDPtr,
		"used_variants":         0,
		"reset":                 true,
	})
}
