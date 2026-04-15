package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hibiken/asynq"
)

// GeneratePayload is the input to the video generation task
type GeneratePayload struct {
	JobID       string `json:"job_id"`
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`
	Angle       string `json:"angle"`
	Script      string `json:"script"`
	Model       string `json:"model"` // veo3 | kling3 | runway4 | sora2
	BrandKitID  string `json:"brand_kit_id,omitempty"`
}

func NewGenerateTask(payload GeneratePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal generate payload: %w", err)
	}
	return asynq.NewTask(TypeGenerate, data, asynq.Queue("default")), nil
}

// HandleGenerate submits async video generation job to FAL.AI
// Always uses fal.queue.submit() — NEVER fal.subscribe() (blocks under load)
func HandleGenerate(ctx context.Context, t *asynq.Task) error {
	var payload GeneratePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal generate payload: %w", err)
	}

	falKey := os.Getenv("FAL_KEY")
	if falKey == "" {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("FAL_KEY not set")
	}

	// FAL model endpoint mapping
	modelEndpoints := map[string]string{
		"veo3":    "fal-ai/veo3",
		"kling3":  "fal-ai/kling-video/v3/standard/text-to-video",
		"runway4": "fal-ai/runway-gen4/turbo/text-to-video",
		"sora2":   "fal-ai/sora",
	}

	endpoint, ok := modelEndpoints[payload.Model]
	if !ok {
		return fmt.Errorf("unknown model: %s", payload.Model)
	}

	falRequestID, err := submitFalQueue(ctx, falKey, endpoint, payload)
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("fal queue submit failed: %w", err)
	}
	_ = falRequestID

	// Transition: generating → postprocessing
	if err := patchJobStatus(payload.JobID, payload.WorkspaceID, "postprocessing"); err != nil {
		_ = err // non-fatal
	}

	return nil
}

type falQueueSubmitResponse struct {
	RequestID string `json:"request_id"`
	StatusURL string `json:"status_url"`
	Error     string `json:"error"`
}

func submitFalQueue(ctx context.Context, falKey, endpoint string, payload GeneratePayload) (string, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": payload.Script,
		"input": map[string]any{
			"prompt":           payload.Script,
			"aspect_ratio":     "9:16",
			"duration_seconds": 15,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal fal request: %w", err)
	}

	queueURL := "https://queue.fal.run/" + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, queueURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build fal request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+falKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("submit fal request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read fal response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("fal HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out falQueueSubmitResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("parse fal response: %w", err)
	}
	if out.Error != "" {
		return "", fmt.Errorf("fal API error: %s", out.Error)
	}
	if out.RequestID == "" {
		return "", fmt.Errorf("fal response missing request_id")
	}

	return out.RequestID, nil
}
