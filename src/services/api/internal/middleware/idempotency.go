package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	// IdempotencyKeyHeader is the header clients send to deduplicate mutations.
	// Clients must generate a UUID per request and retry with the same key.
	IdempotencyKeyHeader = "X-Idempotency-Key"

	// idempotencyKeyCtx is the context key where the extracted key is stored.
	idempotencyKeyCtx = "idempotency_key"
)

// Idempotency extracts the X-Idempotency-Key header and stores it in the
// Echo context. Downstream handlers read it via GetIdempotencyKey(c).
//
// The header is optional on GET/DELETE; required on POST mutations that create
// jobs, briefs, or exports. If required=true and the header is missing,
// the middleware returns 400.
//
// The actual deduplication happens at the DB layer via the UNIQUE INDEX on
// (workspace_id, idempotency_key) — not in this middleware.
func Idempotency(required bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get(IdempotencyKeyHeader)

			if required && key == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "X-Idempotency-Key header is required for this endpoint",
					"hint":  "Generate a UUID v4 per request and retry with the same key to avoid duplicates",
				})
			}

			c.Set(idempotencyKeyCtx, key)
			return next(c)
		}
	}
}

// GetIdempotencyKey retrieves the idempotency key from the Echo context.
// Returns "" if no key was provided (optional use case).
func GetIdempotencyKey(c echo.Context) string {
	v, _ := c.Get(idempotencyKeyCtx).(string)
	return v
}
