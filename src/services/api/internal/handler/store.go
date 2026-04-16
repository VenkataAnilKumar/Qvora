package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qvora/api/internal/db"
	"github.com/qvora/api/internal/store"
)

// dbPool mirrors the shared pool exposed by internal/store.
// Updated on every successful queries() call so that handler functions that
// access dbPool directly (after calling queries/getWorkspaceForOrg) still work.
var dbPool *pgxpool.Pool

// dbQueries kept for backward-compat during incremental migration.
var dbQueries *db.Queries

// queries initialises the store singleton and returns *db.Queries.
// As a side-effect it updates dbPool so that handler functions which use it
// directly (always after calling queries first) still compile and work.
func queries(ctx context.Context) (*db.Queries, error) {
	q, err := store.Queries(ctx)
	if err == nil {
		dbPool = store.Pool()
		dbQueries = q
	}
	return q, err
}
