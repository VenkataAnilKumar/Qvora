package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MuxWebhook handles Mux asset + playback webhooks
// POST /webhooks/mux
func MuxWebhook(c echo.Context) error {
	// TODO: verify Mux webhook signature from X-Mux-Signature header
	// Update asset status in PostgreSQL when Mux finishes processing

	var payload map[string]any
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
	}

	eventType, _ := payload["type"].(string)
	_ = eventType // TODO: handle video.asset.ready, video.asset.errored

	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}
