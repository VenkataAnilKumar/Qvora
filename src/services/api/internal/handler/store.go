package handler

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qvora/api/internal/db"
)

var (
	dbInitOnce sync.Once
	dbInitErr  error
	dbPool     *pgxpool.Pool
	dbQueries  *db.Queries
)

func queries(ctx context.Context) (*db.Queries, error) {
	dbInitOnce.Do(func() {
		databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if databaseURL == "" {
			dbInitErr = errors.New("DATABASE_URL is not set")
			return
		}

		pool, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			dbInitErr = err
			return
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			dbInitErr = err
			return
		}

		dbPool = pool
		dbQueries = db.New(pool)
	})

	if dbInitErr != nil {
		return nil, dbInitErr
	}

	return dbQueries, nil
}
