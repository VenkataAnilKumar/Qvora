package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

	// TODO: POST to Rust Axum /process with payload
	//   Rust service fetches from R2, processes (watermark, captions, transcode, 9:16 reframe), uploads back to R2
	//   Then upload to Mux and store mux_asset_id + mux_playback_id in video_variants
	_ = rustURL
	_ = ctx

	// Transition: postprocessing → complete
	// Note: in production this only fires after all variants for the job are done
	if err := patchJobStatus(payload.JobID, payload.WorkspaceID, "complete"); err != nil {
		_ = err // non-fatal
	}

	return nil
}
