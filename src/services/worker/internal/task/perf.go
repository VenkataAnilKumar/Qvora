package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// PerfEvent is a single performance measurement for one pipeline stage.
// Written to the video_performance_events table via the Go API.
type PerfEvent struct {
	WorkspaceID  string
	VariantID    string
	JobID        string
	Stage        string
	DurationMS   int
	Model        string
	FalRequestID string
	ErrorType    string
	ErrorMsg     string
}

// recordPerfEvent ships a PerfEvent to the API asynchronously.
// Non-blocking — failures are logged but never returned as errors.
// The worker must not fail a task because observability writes fail.
func recordPerfEvent(ctx context.Context, rdb *redis.Client, e PerfEvent) {
	// Fire-and-forget in a goroutine so it never blocks the task handler
	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		apiBase := getEnv("API_BASE_URL")
		internalKey := getEnv("INTERNAL_API_KEY")
		if apiBase == "" {
			return
		}

		body, err := json.Marshal(map[string]any{
			"workspace_id":   e.WorkspaceID,
			"variant_id":     e.VariantID,
			"job_id":         e.JobID,
			"stage":          e.Stage,
			"duration_ms":    e.DurationMS,
			"model":          e.Model,
			"fal_request_id": e.FalRequestID,
			"error_type":     e.ErrorType,
			"error_msg":      e.ErrorMsg,
		})
		if err != nil {
			return
		}

		if err := postJSON(sendCtx,
			fmt.Sprintf("%s/api/v1/internal/perf-events", apiBase),
			e.WorkspaceID, internalKey, body,
		); err != nil {
			slog.Warn("perf event write failed", "stage", e.Stage, "err", err)
		}
	}()
}

// recordCostEvent ships a cost event to the API for billing attribution.
func recordCostEvent(ctx context.Context, workspaceID, variantID, jobID, source, model string, estimatedUSD float64) {
	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		apiBase := getEnv("API_BASE_URL")
		internalKey := getEnv("INTERNAL_API_KEY")
		if apiBase == "" {
			return
		}

		body, _ := json.Marshal(map[string]any{
			"workspace_id":  workspaceID,
			"variant_id":    variantID,
			"job_id":        jobID,
			"source":        source,
			"model":         model,
			"estimated_usd": estimatedUSD,
			"credits":       1,
		})

		if err := postJSON(sendCtx,
			fmt.Sprintf("%s/api/v1/internal/cost-events", apiBase),
			workspaceID, internalKey, body,
		); err != nil {
			slog.Warn("cost event write failed", "source", source, "err", err)
		}
	}()
}
