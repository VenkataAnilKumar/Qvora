package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Health godoc
// GET /api/v1/health
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":    "ok",
		"ok":        true,
		"service":   "qvora-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
