package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/qvora/api/internal/db"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// workspaceIDForOrg is a handler-level bridge to store.WorkspaceIDForOrg.
// The q parameter is ignored — initialisation is handled by the store singleton.
// Domain packages should call store.WorkspaceIDForOrg directly.
func workspaceIDForOrg(ctx context.Context, _ *db.Queries, orgID string) (pgtype.UUID, error) {
	return store.WorkspaceIDForOrg(ctx, orgID)
}

// parseUUID bridges the legacy handler call to util.ParseUUID.
func parseUUID(raw string) (pgtype.UUID, error) {
	return util.ParseUUID(raw)
}

// uuidString bridges the legacy handler call to util.UUIDString.
func uuidString(id pgtype.UUID) string {
	return util.UUIDString(id)
}

// tsTime bridges the legacy handler call to util.TsTime.
func tsTime(ts pgtype.Timestamptz) time.Time {
	return util.TsTime(ts)
}
