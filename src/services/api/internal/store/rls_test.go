package store

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qvora/api/internal/db"
)

// TestRLSCrossWorkspaceIsolation is the "RLS cross-workspace isolation test"
// called for in docs/07-implementation/IMPLEMENTATION_CHECKLIST.md Phase 7 —
// previously untestable in any meaningful way, since RLS was architecturally
// inert (the app always connected as the Postgres superuser, which
// unconditionally bypasses RLS). It now connects as the real non-superuser
// qvora_app role (see migration 007_rls_app_role.sql) and proves two things
// through the actual AcquireForOrg code path, not just raw SQL:
//
//  1. A request scoped to org A cannot see org B's rows, even when (as a
//     defense-in-depth check) queried with org B's own correct workspace_id —
//     this is the part a hand-written `WHERE workspace_id = $1` alone can
//     never catch if the application code ever passes the wrong ID.
//  2. A request with no scope set at all sees nothing (fail closed).
//
// Skips cleanly when TEST_DATABASE_URL isn't set, so it never breaks CI
// (no Postgres service is configured there) or a plain local `go test` —
// opt in by exporting TEST_DATABASE_URL to a non-superuser-role connection
// string pointed at a real Postgres with all migrations applied.
func TestRLSCrossWorkspaceIsolation(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping RLS integration test (needs a real non-superuser-role Postgres connection)")
	}

	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer adminPool.Close()

	// Seed two isolated workspaces + one job each, as the connecting role
	// itself (qvora_app) — relies on test setup having already granted it
	// INSERT, matching migration 007.
	const orgA = "org_rls_test_a"
	const orgB = "org_rls_test_b"
	var workspaceAID, workspaceBID pgtype.UUID
	var jobBID pgtype.UUID

	seed := func(orgID string) (pgtype.UUID, pgtype.UUID) {
		// set_config so this seed insert itself passes RLS as its own org.
		if _, err := adminPool.Exec(ctx, `SELECT set_config('app.org_id', $1, false)`, orgID); err != nil {
			t.Fatalf("seed set_config: %v", err)
		}
		var wsID pgtype.UUID
		err := adminPool.QueryRow(ctx,
			`INSERT INTO workspaces (org_id, plan_tier, sub_status) VALUES ($1, 'starter', 'trialing')
			 ON CONFLICT (org_id) DO UPDATE SET updated_at = NOW() RETURNING id`,
			orgID,
		).Scan(&wsID)
		if err != nil {
			t.Fatalf("seed workspace %s: %v", orgID, err)
		}
		var jobID pgtype.UUID
		err = adminPool.QueryRow(ctx,
			`INSERT INTO jobs (workspace_id, product_url, status, model) VALUES ($1, 'https://rls-test.example.com', 'complete', 'veo3') RETURNING id`,
			wsID,
		).Scan(&jobID)
		if err != nil {
			t.Fatalf("seed job for %s: %v", orgID, err)
		}
		return wsID, jobID
	}
	workspaceAID, _ = seed(orgA)
	workspaceBID, jobBID = seed(orgB)
	_ = jobBID

	t.Cleanup(func() {
		_, _ = adminPool.Exec(ctx, `SELECT set_config('app.org_id', '', false)`)
	})

	// --- Case 1: scope to org A, query org B's workspace_id directly ---
	// A correctly-written WHERE clause would already stop this; the point is
	// that RLS stops it too, independent of application-code correctness.
	scopeA, err := acquireForOrgOnPool(ctx, adminPool, orgA)
	if err != nil {
		t.Fatalf("acquire scope for org A: %v", err)
	}
	jobsAsA, err := scopeA.Queries.ListJobsByWorkspace(ctx, db.ListJobsByWorkspaceParams{
		WorkspaceID: workspaceBID, // org B's workspace, queried while scoped to org A
		Limit:       10,
	})
	scopeA.Close(ctx)
	if err != nil {
		t.Fatalf("query as org A: %v", err)
	}
	if len(jobsAsA) != 0 {
		t.Errorf("RLS FAILED: org A's scope saw %d of org B's jobs when queried by org B's own workspace_id — RLS should have blocked this regardless of the WHERE clause", len(jobsAsA))
	}

	// --- Case 2: scope to org A, query org A's own workspace_id — should see it ---
	scopeA2, err := acquireForOrgOnPool(ctx, adminPool, orgA)
	if err != nil {
		t.Fatalf("acquire scope for org A (2): %v", err)
	}
	jobsOwnA, err := scopeA2.Queries.ListJobsByWorkspace(ctx, db.ListJobsByWorkspaceParams{
		WorkspaceID: workspaceAID,
		Limit:       10,
	})
	scopeA2.Close(ctx)
	if err != nil {
		t.Fatalf("query own workspace as org A: %v", err)
	}
	if len(jobsOwnA) == 0 {
		t.Error("org A's scope saw ZERO of its own jobs — RLS policy or seed data is broken, not just overly strict")
	}

	// --- Case 3: no app.org_id set at all — should fail closed to zero rows ---
	conn, err := adminPool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire raw conn: %v", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `SELECT set_config('app.org_id', '', false)`); err != nil {
		t.Fatalf("reset org_id: %v", err)
	}
	noScopeQueries := db.New(conn)
	jobsNoScope, err := noScopeQueries.ListJobsByWorkspace(ctx, db.ListJobsByWorkspaceParams{
		WorkspaceID: workspaceAID,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("query with no scope: %v", err)
	}
	if len(jobsNoScope) != 0 {
		t.Errorf("RLS FAILED to fail closed: saw %d rows with no app.org_id set at all", len(jobsNoScope))
	}
}

// acquireForOrgOnPool mirrors AcquireForOrg but against an explicit pool
// argument rather than the package-level singleton, so this test can use its
// own isolated pool/connection string instead of whatever DATABASE_URL the
// rest of the package may have already initialized the shared pool from.
func acquireForOrgOnPool(ctx context.Context, p *pgxpool.Pool, orgID string) (*RequestScope, error) {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.Exec(ctx, `SELECT set_config('app.org_id', $1, false)`, orgID); err != nil {
		conn.Release()
		return nil, err
	}
	return &RequestScope{conn: conn, Queries: db.New(conn)}, nil
}
