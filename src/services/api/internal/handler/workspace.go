package handler

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/domain/identity"
	"github.com/qvora/api/internal/store"
)

// getWorkspaceForOrg bridges handler files not yet migrated to domain packages.
// Domain packages should call store.GetWorkspaceForOrg directly.
func getWorkspaceForOrg(c echo.Context, orgID string) (pgtype.UUID, string, string, pgtype.Timestamptz, error) {
	return store.GetWorkspaceForOrg(c.Request().Context(), orgID)
}

func GetWorkspace(c echo.Context) error   { return identity.GetWorkspace(c) }
func GetBrandKit(c echo.Context) error    { return identity.GetBrandKit(c) }
func UpsertBrandKit(c echo.Context) error { return identity.UpsertBrandKit(c) }
func UpdateWorkspaceSubscription(c echo.Context) error {
	return identity.UpdateWorkspaceSubscription(c)
}
func UpdateWorkspaceLifecycle(c echo.Context) error { return identity.UpdateWorkspaceLifecycle(c) }
func GetWorkspaceUsage(c echo.Context) error        { return identity.GetWorkspaceUsage(c) }
func SyncWorkspaceMembership(c echo.Context) error  { return identity.SyncWorkspaceMembership(c) }
func ResetWorkspaceUsage(c echo.Context) error      { return identity.ResetWorkspaceUsage(c) }

// incrementWorkspaceUsageForVariant bridges handler/webhook.go → identity domain.
// Remove once webhook.go is migrated to domain/media.
func incrementWorkspaceUsageForVariant(c echo.Context, workspaceID pgtype.UUID, variantID pgtype.UUID) (int64, bool, error) {
	return identity.IncrementWorkspaceUsageForVariant(c, workspaceID, variantID)
}
