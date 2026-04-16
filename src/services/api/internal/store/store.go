package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qvora/api/internal/db"
)

var (
	initOnce sync.Once
	initErr  error
	pool     *pgxpool.Pool
	queries_ *db.Queries
)

// Pool returns the shared pgxpool.Pool. May be nil if not yet initialised.
func Pool() *pgxpool.Pool { return pool }

// Queries initialises the pool on first call and returns the shared *db.Queries.
func Queries(ctx context.Context) (*db.Queries, error) {
	initOnce.Do(func() {
		databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if databaseURL == "" {
			initErr = errors.New("DATABASE_URL is not set")
			return
		}
		p, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			initErr = err
			return
		}
		if err := p.Ping(ctx); err != nil {
			p.Close()
			initErr = err
			return
		}
		pool = p
		queries_ = db.New(p)
	})
	if initErr != nil {
		return nil, initErr
	}
	return queries_, nil
}

// WorkspaceIDForOrg returns the workspace UUID for the given Clerk org, creating
// a new workspace record (trialing, starter) on first access.
func WorkspaceIDForOrg(ctx context.Context, orgID string) (pgtype.UUID, error) {
	q, err := Queries(ctx)
	if err != nil {
		return pgtype.UUID{}, err
	}
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

// GetWorkspaceForOrg returns the full workspace record fields needed by most
// handlers: workspaceID, planTier, subStatus, trialEndsAt.
func GetWorkspaceForOrg(ctx context.Context, orgID string) (workspaceID pgtype.UUID, planTier string, subStatus string, trialEndsAt pgtype.Timestamptz, err error) {
	workspaceID, err = WorkspaceIDForOrg(ctx, orgID)
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}
	q, err := Queries(ctx)
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}
	workspace, err := q.GetWorkspaceByOrgID(ctx, orgID)
	if err != nil {
		return pgtype.UUID{}, "", "", pgtype.Timestamptz{}, err
	}
	return workspaceID, workspace.PlanTier, workspace.SubStatus, workspace.TrialEndsAt, nil
}
