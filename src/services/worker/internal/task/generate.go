package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// GeneratePayload is the input to the video generation task.
type GeneratePayload struct {
	JobID       string `json:"job_id"`
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`
	PlanTier    string `json:"plan_tier"` // starter | growth | agency
	Angle       string `json:"angle"`
	Script      string `json:"script"`
	Model       string `json:"model"` // veo3 | kling3 | runway4 | sora2
	BrandKitID  string `json:"brand_kit_id,omitempty"`
	// UseAvatar: if true, enqueue avatar task after FAL completes (V2V lip-sync)
	UseAvatar         bool   `json:"use_avatar,omitempty"`
	AudioURL          string `json:"audio_url,omitempty"`
	AvatarProvider    string `json:"avatar_provider,omitempty"`
}

// modelDisplayNames maps internal model keys → human-readable cost-tracking names.
var modelDisplayNames = map[string]string{
	"veo3":    "veo-3.1",
	"kling3":  "kling-3.0",
	"runway4": "runway-gen4",
	"sora2":   "sora-2",
}

func NewGenerateTask(payload GeneratePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal generate payload: %w", err)
	}
	return asynq.NewTask(TypeGenerate, data,
		asynq.Queue("default"),
		asynq.MaxRetry(MaxTaskRetryAttempts),
		asynq.Timeout(30*time.Minute),
	), nil
}

// HandleGenerate submits an async video generation job via the given VideoProvider.
//
// Guards:
//   - Cost circuit breaker: blocks if workspace hourly spend limit hit
//   - FAL concurrency semaphore: max 2 concurrent per workspace (FAL hard limit)
//   - Performance event: records dispatch latency to video_performance_events
//   - Cost event: records estimated USD cost for billing attribution
//
// Always uses async queue submission — NEVER blocking subscribe.
func HandleGenerate(rdb *redis.Client, prov VideoProvider) asynq.HandlerFunc {
	sem := NewFalSemaphore(rdb)
	cb := NewCostCircuitBreaker(rdb)

	return func(ctx context.Context, t *asynq.Task) error {
		var p GeneratePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshal generate payload: %w", err)
		}

		log := slog.With(
			"job_id", p.JobID,
			"variant_id", p.VariantID,
			"workspace_id", p.WorkspaceID,
			"model", p.Model,
		)

		// ----------------------------------------------------------------
		// 1. Cost circuit breaker — check before acquiring semaphore
		// ----------------------------------------------------------------
		modelName := modelDisplayNames[p.Model]
		if modelName == "" {
			modelName = p.Model // fallback for unknown display name
		}
		if err := cb.CheckAndIncrement(ctx, p.WorkspaceID, p.PlanTier, modelName); err != nil {
			log.Warn("cost circuit breaker open", "err", err)
			_ = patchJobStatus(p.JobID, p.WorkspaceID, "failed")
			return fmt.Errorf("cost limit: %w", err)
		}

		// ----------------------------------------------------------------
		// 2. FAL.AI concurrency semaphore (max 2 per workspace)
		// ----------------------------------------------------------------
		slotKey, err := sem.Acquire(ctx, p.WorkspaceID, p.VariantID)
		if errors.Is(err, ErrSemaphoreFull) {
			// Workspace is at FAL concurrency limit — snooze and retry
			log.Info("fal semaphore full, will retry")
			return asynq.SkipRetry // let asynq re-enqueue after backoff
		}
		if err != nil {
			return fmt.Errorf("semaphore acquire: %w", err)
		}
		// Release semaphore on task exit (normal + error paths)
		defer func() {
			if releaseErr := sem.Release(ctx, slotKey); releaseErr != nil {
				log.Warn("semaphore release failed", "err", releaseErr)
			}
		}()

		log.Info("fal semaphore acquired", "slot", slotKey)

		// ----------------------------------------------------------------
		// 3. Submit to video provider async queue
		// ----------------------------------------------------------------
		start := time.Now()
		result, err := prov.Submit(ctx, VideoRequest{
			JobID:          p.JobID,
			VariantID:      p.VariantID,
			WorkspaceID:    p.WorkspaceID,
			PlanTier:       p.PlanTier,
			Angle:          p.Angle,
			Script:         p.Script,
			Model:          p.Model,
			UseAvatar:      p.UseAvatar,
			AudioURL:       p.AudioURL,
			AvatarProvider: p.AvatarProvider,
		})
		dispatchMS := int(time.Since(start).Milliseconds())

		if err != nil {
			_ = patchJobStatus(p.JobID, p.WorkspaceID, "failed")
			recordPerfEvent(ctx, rdb, PerfEvent{
				WorkspaceID: p.WorkspaceID,
				VariantID:   p.VariantID,
				JobID:       p.JobID,
				Stage:       "fal_queue",
				DurationMS:  dispatchMS,
				Model:       modelName,
				ErrorType:   "provider_submit",
				ErrorMsg:    err.Error(),
			})
			return fmt.Errorf("video provider submit: %w", err)
		}
		falRequestID := result.ProviderJobID

		log.Info("video job submitted", "provider", result.Provider, "request_id", falRequestID, "dispatch_ms", dispatchMS)

		// ----------------------------------------------------------------
		// 4. Store FAL request ID on variant (webhook routing)
		// ----------------------------------------------------------------
		if err := patchVariantFalRequestID(ctx, p.VariantID, p.WorkspaceID, falRequestID); err != nil {
			log.Warn("patch fal_request_id failed (non-fatal)", "err", err)
		}

		// ----------------------------------------------------------------
		// 5. Record dispatch perf + cost events
		// ----------------------------------------------------------------
		recordPerfEvent(ctx, rdb, PerfEvent{
			WorkspaceID:  p.WorkspaceID,
			VariantID:    p.VariantID,
			JobID:        p.JobID,
			Stage:        "fal_queue",
			DurationMS:   dispatchMS,
			Model:        modelName,
			FalRequestID: falRequestID,
		})

		if cost, ok := ModelCost[modelName]; ok {
			recordCostEvent(ctx, p.WorkspaceID, p.VariantID, p.JobID,
				"fal_generate", modelName, cost)
		}

		// Semaphore is released via defer when FAL webhook arrives and
		// the postprocess task is enqueued. If the webhook never arrives,
		// the TTL on the semaphore key auto-releases it after falSemaphoreTTL.

		return nil
	}
}

// getEnv is a helper used across this package.
func getEnv(key string) string {
	return os.Getenv(key)
}
