package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hibiken/asynq"
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

// FalWebhook handles FAL completion webhooks and enqueues postprocess tasks.
// POST /webhooks/fal
// Expected payload fields:
// - status: "completed"
// - metadata.variant_id / metadata.job_id / metadata.workspace_id
// - input_r2_key, output_r2_key
func FalWebhook(c echo.Context) error {
	var payload map[string]any
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
	}

	status := strings.ToLower(strings.TrimSpace(findString(payload, "status")))
	if status != "completed" && status != "complete" {
		return c.JSON(http.StatusOK, map[string]bool{"received": true})
	}

	metadata := findMap(payload, "metadata")
	jobID := strings.TrimSpace(findString(metadata, "job_id"))
	variantID := strings.TrimSpace(findString(metadata, "variant_id"))
	workspaceID := strings.TrimSpace(findString(metadata, "workspace_id"))
	inputR2Key := strings.TrimSpace(findString(payload, "input_r2_key"))
	outputR2Key := strings.TrimSpace(findString(payload, "output_r2_key"))

	if inputR2Key == "" {
		inputR2Key = strings.TrimSpace(findString(metadata, "input_r2_key"))
	}
	if outputR2Key == "" {
		outputR2Key = strings.TrimSpace(findString(metadata, "output_r2_key"))
	}

	if jobID == "" || variantID == "" || workspaceID == "" || inputR2Key == "" || outputR2Key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
	}

	redisURL := strings.TrimSpace(os.Getenv("RAILWAY_REDIS_URL"))
	if redisURL == "" {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "redis_not_configured"})
	}

	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "redis_uri_invalid"})
	}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	postprocessPayload := map[string]any{
		"job_id":        jobID,
		"variant_id":    variantID,
		"workspace_id":  workspaceID,
		"input_r2_key":  inputR2Key,
		"output_r2_key": outputR2Key,
		"watermark":     true,
		"add_captions":  false,
	}

	data, err := json.Marshal(postprocessPayload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "marshal_failed"})
	}

	task := asynq.NewTask("job:postprocess", data, asynq.Queue("critical"))
	if _, err := client.Enqueue(task); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "enqueue_failed"})
	}

	q, err := queries(c.Request().Context())
	if err == nil {
		if parsedWorkspaceID, wsErr := parseUUID(workspaceID); wsErr == nil {
			if parsedJobID, parseErr := parseUUID(jobID); parseErr == nil {
				if jobRow, getErr := q.GetJobByID(c.Request().Context(), db.GetJobByIDParams{ID: parsedJobID, WorkspaceID: parsedWorkspaceID}); getErr == nil {
					_, _ = q.UpdateJobStatus(c.Request().Context(), db.UpdateJobStatusParams{ID: jobRow.ID, Status: "postprocessing"})
				}
			}
		}
	}

	zap.L().Info("fal completion webhook received",
		zap.String("job_id", jobID),
		zap.String("variant_id", variantID),
		zap.String("workspace_id", workspaceID),
		zap.String("input_r2_key", inputR2Key),
		zap.String("output_r2_key", outputR2Key),
	)

	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}

func findString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func findMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return map[string]any{}
	}
	out, ok := v.(map[string]any)
	if ok {
		return out
	}
	return map[string]any{}
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
