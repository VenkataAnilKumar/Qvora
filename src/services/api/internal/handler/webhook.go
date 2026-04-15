package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	"go.uber.org/zap"
)

// MuxWebhook handles Mux asset + playback webhooks
// POST /webhooks/mux
// Verifies signature and updates variant with mux_asset_id, mux_playable_id on success
func MuxWebhook(c echo.Context) error {
	// Read body for signature verification (we'll need to re-read it for JSON parsing)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed_to_read_body"})
	}

	// Verify Mux webhook signature
	signature := c.Request().Header.Get("X-Mux-Signature-V2")
	if !verifyMuxSignature(string(body), signature) {
		zap.L().Warn("mux webhook signature verification failed")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_signature"})
	}

	// Parse payload envelope
	var payload MuxWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		zap.L().Error("failed to parse mux payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_json"})
	}

	// Only handle video.asset.ready event (when video is ingested and playback is ready)
	if payload.Type != "video.asset.ready" {
		// Silently accept other event types (e.g., video.asset.errored)
		return c.JSON(http.StatusOK, map[string]bool{"received": true})
	}

	// Parse typed asset data.
	var assetData muxAssetData
	if err := json.Unmarshal(payload.Data, &assetData); err != nil {
		zap.L().Error("invalid mux asset data format", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_data_format"})
	}

	if assetData.ID == "" {
		zap.L().Error("no asset_id in mux payload")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_asset_id"})
	}
	if assetData.Passthrough == "" {
		zap.L().Error("no passthrough variant_id in mux payload", zap.String("asset_id", assetData.ID))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_passthrough"})
	}

	playbackID := ""
	if len(assetData.PlaybackIDs) > 0 {
		playbackID = assetData.PlaybackIDs[0].ID
	}
	if playbackID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_playback_id"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		zap.L().Error("database unavailable for mux webhook", zap.Error(err))
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	var variantID pgtype.UUID
	if err := variantID.Scan(assetData.Passthrough); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}

	updated, err := q.UpdateVariantMuxByID(c.Request().Context(), db.UpdateVariantMuxByIDParams{
		ID:            variantID,
		MuxAssetID:    &assetData.ID,
		MuxPlaybackID: &playbackID,
	})
	if err != nil {
		zap.L().Error("failed to update variant from mux webhook", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_update_failed"})
	}

	zap.L().Info("mux webhook received",
		zap.String("variant_id", assetData.Passthrough),
		zap.String("asset_id", assetData.ID),
		zap.String("playback_id", playbackID),
		zap.String("variant_status", updated.Status),
		zap.String("event_type", payload.Type))

	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}

// MuxWebhookPayload represents the webhook event from Mux
type MuxWebhookPayload struct {
	Type        string      `json:"type"`        // "video.asset.ready", "video.asset.errored", etc.
	Data        json.RawMessage `json:"data"`    // Asset data (id, playback_ids, passthrough)
	CreatedAt   string      `json:"created_at"`  // ISO timestamp
	EventID     string      `json:"id"`          // Webhook event ID (for dedup)
	Attemptnum  int         `json:"attemptnum"`  // Retry attempt
}

type muxAssetData struct {
	ID          string `json:"id"`
	Passthrough string `json:"passthrough"`
	PlaybackIDs []struct {
		ID string `json:"id"`
	} `json:"playback_ids"`
}

// verifyMuxSignature verifies the HMAC-SHA256 signature
// Signature format: mux_v2 <timestamp>.base64_signature
func verifyMuxSignature(body, signature string) bool {
	secret := os.Getenv("MUX_WEBHOOK_SECRET")
	if secret == "" {
		return false
	}

	// Mux signature verification: HMAC-SHA256 of the request body with the webhook secret key
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(body))
	expectedSig := "mux_v2 " + hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// ClerkWebhook handles Clerk organization deletion webhooks
// POST /webhooks/clerk
func ClerkWebhook(c echo.Context) error {
	var payload map[string]any
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
	}

	eventType, _ := payload["type"].(string)
	_ = eventType // TODO: handle organization.deleted event

	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}
