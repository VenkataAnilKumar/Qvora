package media

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
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	"github.com/qvora/api/internal/domain/identity"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
	"go.uber.org/zap"
)

// MuxWebhookPayload represents the webhook event envelope from Mux.
type MuxWebhookPayload struct {
	Type       string          `json:"type"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  string          `json:"created_at"`
	EventID    string          `json:"id"`
	Attemptnum int             `json:"attemptnum"`
}

type muxAssetData struct {
	ID          string `json:"id"`
	Passthrough string `json:"passthrough"`
	PlaybackIDs []struct {
		ID string `json:"id"`
	} `json:"playback_ids"`
}

// MuxWebhook handles POST /webhooks/mux
func MuxWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed_to_read_body"})
	}
	signature := c.Request().Header.Get("X-Mux-Signature-V2")
	if !verifyMuxSignature(string(body), signature) {
		zap.L().Warn("mux webhook signature verification failed")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid_signature"})
	}
	var payload MuxWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		zap.L().Error("failed to parse mux payload", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_json"})
	}
	if strings.TrimSpace(payload.EventID) != "" {
		duplicate, dedupErr := recordMuxWebhookEvent(c, payload.EventID, payload.Type, body)
		if dedupErr != nil {
			zap.L().Error("failed to persist mux event for dedup", zap.Error(dedupErr), zap.String("event_id", payload.EventID))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mux_event_dedup_failed"})
		}
		if duplicate {
			zap.L().Info("duplicate mux webhook ignored", zap.String("event_id", payload.EventID), zap.String("event_type", payload.Type))
			return c.JSON(http.StatusOK, map[string]any{"received": true, "duplicate": true})
		}
	}
	if payload.Type != "video.asset.ready" {
		return c.JSON(http.StatusOK, map[string]bool{"received": true})
	}
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
	q, err := store.Queries(c.Request().Context())
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
	variants, listErr := q.ListVariantsByJob(c.Request().Context(), updated.JobID)
	if listErr != nil {
		zap.L().Warn("failed to list variants for job completion check",
			zap.Error(listErr),
			zap.String("job_id", util.UUIDString(updated.JobID)),
		)
	} else if areAllVariantsComplete(variants) {
		if _, updateErr := q.UpdateJobStatus(c.Request().Context(), db.UpdateJobStatusParams{
			ID:     updated.JobID,
			Status: "complete",
		}); updateErr != nil {
			zap.L().Warn("failed to mark job complete after mux webhook",
				zap.Error(updateErr),
				zap.String("job_id", util.UUIDString(updated.JobID)),
			)
		} else {
			zap.L().Info("job marked complete after all variants finished",
				zap.String("job_id", util.UUIDString(updated.JobID)),
			)
		}
	}
	usedVariants, incremented, usageErr := identity.IncrementWorkspaceUsageForVariant(c, updated.WorkspaceID, updated.ID)
	if usageErr != nil {
		zap.L().Error("failed to increment workspace usage",
			zap.Error(usageErr),
			zap.String("workspace_id", util.UUIDString(updated.WorkspaceID)),
			zap.String("variant_id", util.UUIDString(updated.ID)),
		)
	} else {
		zap.L().Info("workspace usage updated",
			zap.String("workspace_id", util.UUIDString(updated.WorkspaceID)),
			zap.String("variant_id", util.UUIDString(updated.ID)),
			zap.Bool("incremented", incremented),
			zap.Int64("used_variants", usedVariants),
		)
	}
	zap.L().Info("mux webhook received",
		zap.String("variant_id", assetData.Passthrough),
		zap.String("job_id", util.UUIDString(updated.JobID)),
		zap.String("asset_id", assetData.ID),
		zap.String("playback_id", playbackID),
		zap.String("variant_status", updated.Status),
		zap.String("event_type", payload.Type),
	)
	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}

// FalWebhook handles POST /webhooks/fal
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
	rawVideoURL := strings.TrimSpace(findString(metadata, "raw_video_url"))
	audioURL := strings.TrimSpace(findString(metadata, "audio_url"))
	useAvatar := findString(metadata, "use_avatar") == "true"
	inputR2Key := strings.TrimSpace(findString(payload, "input_r2_key"))
	outputR2Key := strings.TrimSpace(findString(payload, "output_r2_key"))
	videoPayload := findMap(findMap(payload, "payload"), "video")
	if inputR2Key == "" {
		inputR2Key = strings.TrimSpace(findString(metadata, "input_r2_key"))
	}
	if outputR2Key == "" {
		outputR2Key = strings.TrimSpace(findString(metadata, "output_r2_key"))
	}
	if rawVideoURL == "" {
		rawVideoURL = strings.TrimSpace(findString(videoPayload, "url"))
	}
	if rawVideoURL == "" {
		rawVideoURL = strings.TrimSpace(findString(payload, "video_url"))
	}
	if jobID == "" || variantID == "" || workspaceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
	}
	if !useAvatar && (inputR2Key == "" || outputR2Key == "") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
	}
	if useAvatar && (rawVideoURL == "" || audioURL == "") {
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
	var (
		taskData  []byte
		taskType  string
		taskQueue = "critical"
	)
	if useAvatar && rawVideoURL != "" && audioURL != "" {
		avatarPayload := map[string]any{
			"job_id":             jobID,
			"variant_id":         variantID,
			"workspace_id":       workspaceID,
			"raw_video_url":      rawVideoURL,
			"audio_url":          audioURL,
			"preferred_provider": "heygen_v3",
		}
		taskData, err = json.Marshal(avatarPayload)
		taskType = "job:avatar"
	} else {
		postprocessPayload := map[string]any{
			"job_id":                 jobID,
			"variant_id":             variantID,
			"workspace_id":           workspaceID,
			"postprocess_request_id": uuid.NewString(),
			"input_r2_key":           inputR2Key,
			"output_r2_key":          outputR2Key,
			"watermark":              true,
			"add_captions":           false,
		}
		taskData, err = json.Marshal(postprocessPayload)
		taskType = "job:postprocess"
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "marshal_failed"})
	}
	enqueuedTask := asynq.NewTask(taskType, taskData, asynq.Queue(taskQueue), asynq.MaxRetry(10), asynq.Timeout(20*time.Minute))
	if _, err := client.Enqueue(enqueuedTask); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "enqueue_failed"})
	}
	q, err := store.Queries(c.Request().Context())
	if err == nil {
		if parsedWorkspaceID, wsErr := util.ParseUUID(workspaceID); wsErr == nil {
			if parsedJobID, parseErr := util.ParseUUID(jobID); parseErr == nil {
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
		zap.String("raw_video_url", rawVideoURL),
		zap.String("audio_url", audioURL),
		zap.Bool("use_avatar", useAvatar),
		zap.String("task_type", taskType),
	)
	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}

// ClerkWebhook handles POST /webhooks/clerk
func ClerkWebhook(c echo.Context) error {
	var payload map[string]any
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
	}
	eventType, _ := payload["type"].(string)
	_ = eventType // TODO: handle organization.deleted event
	return c.JSON(http.StatusOK, map[string]bool{"received": true})
}

func recordMuxWebhookEvent(c echo.Context, eventID, eventType string, body []byte) (bool, error) {
	if store.Pool() == nil {
		return false, fmt.Errorf("database_not_initialized")
	}
	result, err := store.Pool().Exec(
		c.Request().Context(),
		`INSERT INTO mux_webhook_events (event_id, event_type, payload)
		 VALUES ($1, $2, $3::jsonb)
		 ON CONFLICT (event_id) DO NOTHING`,
		eventID,
		eventType,
		string(body),
	)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 0, nil
}

func areAllVariantsComplete(variants []db.Variant) bool {
	if len(variants) == 0 {
		return false
	}
	for _, variant := range variants {
		if !strings.EqualFold(strings.TrimSpace(variant.Status), "complete") {
			return false
		}
	}
	return true
}

func verifyMuxSignature(body, signature string) bool {
	secret := os.Getenv("MUX_WEBHOOK_SECRET")
	if secret == "" {
		return false
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(body))
	expectedSig := "mux_v2 " + hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSig))
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
