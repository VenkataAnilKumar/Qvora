package handler

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/qvora/api/internal/db"
)

func workspaceIDForOrg(ctx context.Context, q *db.Queries, orgID string) (pgtype.UUID, error) {
	workspace, err := q.GetWorkspaceByOrgID(ctx, orgID)
	if err == nil {
		return workspace.ID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}

	workspace, err = q.UpsertWorkspace(ctx, db.UpsertWorkspaceParams{
		OrgID:     orgID,
		PlanTier:  "starter",
		SubStatus: "trialing",
		TrialEndsAt: pgtype.Timestamptz{
			Time:  time.Now().UTC().Add(7 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		return pgtype.UUID{}, err
	}

	return workspace.ID, nil
}

func parseUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(raw)
	return id, err
}

func uuidString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	u, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return ""
	}
	return u.String()
}

func tsTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time.UTC()
}
