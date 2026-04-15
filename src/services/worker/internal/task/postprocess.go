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

// PostprocessPayload is the input to the Rust postprocessor
type PostprocessPayload struct {
	JobID       string `json:"job_id"`      // parent job — used for final status transition
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`
	InputR2Key  string `json:"input_r2_key"`  // raw FAL output in R2
	OutputR2Key string `json:"output_r2_key"` // final processed output
	Watermark   bool   `json:"watermark"`
	AddCaptions bool   `json:"add_captions"`
	Script      string `json:"script,omitempty"` // for caption burn-in
}

func NewPostprocessTask(payload PostprocessPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal postprocess payload: %w", err)
	}
	return asynq.NewTask(TypePostprocess, data, asynq.Queue("critical")), nil
}

// HandlePostprocess calls the Rust Axum postprocessor service
// Rust handles: watermark, captions, transcode, 9:16 reframe (CPU-bound)
func HandlePostprocess(ctx context.Context, t *asynq.Task) error {
	var payload PostprocessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal postprocess payload: %w", err)
	}

	rustURL := os.Getenv("RUST_POSTPROCESS_URL")
	if rustURL == "" {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("RUST_POSTPROCESS_URL not set")
	}

	body, err := json.Marshal(map[string]any{
		"variant_id":    payload.VariantID,
		"workspace_id":  payload.WorkspaceID,
		"input_r2_key":  payload.InputR2Key,
		"output_r2_key": payload.OutputR2Key,
		"watermark":     payload.Watermark,
		"add_captions":  payload.AddCaptions,
		"script":        payload.Script,
	})
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("marshal rust request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(rustURL, "/")+"/process",
		bytes.NewReader(body),
	)
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("build rust request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("post to rust process endpoint: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("rust process HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	// Transition: postprocessing → complete
	// Note: in production this only fires after all variants for the job are done
	if err := patchJobStatus(payload.JobID, payload.WorkspaceID, "complete"); err != nil {
		_ = err // non-fatal
	}

	return nil
}
