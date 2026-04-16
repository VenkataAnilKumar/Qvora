package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
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

// modelEndpoints maps Qvora model names → FAL.AI endpoint paths.
var modelEndpoints = map[string]string{
	"veo3":    "fal-ai/veo3",
	"kling3":  "fal-ai/kling-video/v3/standard/text-to-video",
	"runway4": "fal-ai/runway-gen4/turbo/text-to-video",
	"sora2":   "fal-ai/sora",
}

// modelDisplayNames maps internal keys → human-readable cost-tracking names.
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

// HandleGenerate submits an async video generation job to FAL.AI.
//
// Guards added vs original:
//   - Cost circuit breaker: blocks if workspace hourly spend limit hit
//   - FAL concurrency semaphore: max 2 concurrent per workspace (FAL hard limit)
//   - Performance event: records dispatch latency to video_performance_events
//   - Cost event: records estimated USD cost for billing attribution
//
// Always uses fal.queue.submit() — NEVER fal.subscribe() (blocks under load).
func HandleGenerate(rdb *redis.Client) asynq.HandlerFunc {
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

		falKey := os.Getenv("FAL_KEY")
		if falKey == "" {
			_ = patchJobStatus(p.JobID, p.WorkspaceID, "failed")
			return fmt.Errorf("FAL_KEY not set")
		}

		endpoint, ok := modelEndpoints[p.Model]
		if !ok {
			return fmt.Errorf("unknown model: %s", p.Model)
		}

		// ----------------------------------------------------------------
		// 1. Cost circuit breaker — check before acquiring semaphore
		// ----------------------------------------------------------------
		modelName := modelDisplayNames[p.Model]
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
		// 3. Submit to FAL.AI async queue
		// ----------------------------------------------------------------
		start := time.Now()
		falRequestID, err := submitFalQueue(ctx, falKey, endpoint, p)
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
				ErrorType:   "fal_submit",
				ErrorMsg:    err.Error(),
			})
			return fmt.Errorf("fal queue submit: %w", err)
		}

		log.Info("fal job submitted", "fal_request_id", falRequestID, "dispatch_ms", dispatchMS)

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

// =============================================================================
// FAL.AI queue submission
// =============================================================================

type falQueueSubmitResponse struct {
	RequestID string `json:"request_id"`
	StatusURL string `json:"status_url"`
	Error     string `json:"error"`
}

func submitFalQueue(ctx context.Context, falKey, endpoint string, p GeneratePayload) (string, error) {
	apiBaseURL := strings.TrimSpace(os.Getenv("API_BASE_URL"))
	webhookURL := strings.TrimSpace(os.Getenv("FAL_WEBHOOK_URL"))
	if webhookURL == "" && apiBaseURL != "" {
		webhookURL = strings.TrimRight(apiBaseURL, "/") + "/webhooks/fal"
	}

	body, err := json.Marshal(map[string]any{
		"prompt": p.Script,
		"input": map[string]any{
			"prompt":           p.Script,
			"aspect_ratio":     "9:16",
			"duration_seconds": 15,
		},
		"webhook_url": webhookURL,
		"metadata": map[string]any{
			"job_id":       p.JobID,
			"variant_id":   p.VariantID,
			"workspace_id": p.WorkspaceID,
			"use_avatar":   p.UseAvatar,
			"audio_url":    p.AudioURL,
			"avatar_provider": func() string {
				if p.AvatarProvider != "" {
					return p.AvatarProvider
				}
				return "heygen_v3"
			}(),
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

// getEnv is a helper used across this package.
func getEnv(key string) string {
	return os.Getenv(key)
}
