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

// RequestScope wraps a single acquired connection and a *db.Queries built from
// it, with app.org_id set for that connection's session. This exists because
// RLS policies key on current_setting('app.org_id', TRUE), which is
// connection-session-scoped — the shared Queries() above wraps the pool
// directly, where each query can land on a different pooled connection, so
// there is no stable session to attach a per-request setting to. Acquiring a
// dedicated connection per request is the same technique PostgREST/Supabase
// use under the hood for the same reason.
//
// Must be paired with a deferred Close() call, which resets the session
// setting before releasing the connection — otherwise a later request reusing
// this same pooled connection could silently inherit the wrong org_id.
type RequestScope struct {
	conn    *pgxpool.Conn
	Queries *db.Queries
}

// AcquireForOrg acquires a dedicated connection from the shared pool and sets
// app.org_id for its session via set_config() — not a literal `SET x = $1`,
// which Postgres does not allow as a bind parameter.
func AcquireForOrg(ctx context.Context, orgID string) (*RequestScope, error) {
	if pool == nil {
		if _, err := Queries(ctx); err != nil {
			return nil, err
		}
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, `SELECT set_config('app.org_id', $1, false)`, orgID); err != nil {
		conn.Release()
		return nil, err
	}
	return &RequestScope{conn: conn, Queries: db.New(conn)}, nil
}

// Close resets the session's app.org_id and releases the connection back to
// the pool. Always call via defer immediately after a successful AcquireForOrg.
func (s *RequestScope) Close(ctx context.Context) {
	_, _ = s.conn.Exec(ctx, `SELECT set_config('app.org_id', '', false)`)
	s.conn.Release()
}

// Conn exposes the underlying connection directly (satisfies db.DBTX — the
// same Exec/Query/QueryRow shape as *pgxpool.Pool) for call sites that use raw
// SQL via store.Pool() rather than sqlc's generated Queries wrapper.
func (s *RequestScope) Conn() *pgxpool.Conn {
	return s.conn
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
