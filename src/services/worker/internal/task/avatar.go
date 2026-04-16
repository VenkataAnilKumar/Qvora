package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// TypeAvatar is the asynq task type for HeyGen / Tavus avatar lip-sync.
const TypeAvatar = "job:avatar"

// AvatarPayload is the task payload enqueued after FAL video generation.
// The Go worker calls this task to overlay the ElevenLabs audio on the raw video.
type AvatarPayload struct {
	JobID       string `json:"job_id"`
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`

	// RawVideoURL is the FAL-generated video (no audio yet)
	RawVideoURL string `json:"raw_video_url"`
	// AudioURL is the ElevenLabs TTS output (MP3, stored in R2)
	AudioURL string `json:"audio_url"`

	// PreferredProvider: "heygen_v3" | "tavus_v2" (defaults to "heygen_v3")
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// NewAvatarTask creates an asynq task for avatar lip-sync.
func NewAvatarTask(p AvatarPayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal avatar payload: %w", err)
	}
	return asynq.NewTask(TypeAvatar, data,
		asynq.MaxRetry(3),
		asynq.Timeout(15*time.Minute),
		asynq.Queue("default"),
	), nil
}

// HandleAvatar processes an avatar lip-sync task.
// It selects the preferred provider (HeyGen v3 by default),
// polls for completion, and on success enqueues the postprocess task.
func HandleAvatar(rdb *redis.Client) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p AvatarPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshal avatar payload: %w", err)
		}

		log := slog.With(
			"job_id", p.JobID,
			"variant_id", p.VariantID,
			"workspace_id", p.WorkspaceID,
		)
		log.Info("avatar task started")

		// Select provider
		providerName := p.PreferredProvider
		if providerName == "" {
			providerName = "heygen_v3"
		}
		provider := GetAvatarProvider(providerName)

		req := AvatarRequest{
			VideoURL:    p.RawVideoURL,
			AudioURL:    p.AudioURL,
			WorkspaceID: p.WorkspaceID,
			VariantID:   p.VariantID,
		}

		// Submit lip-sync job
		start := time.Now()
		jobID, err := provider.CreateLipSync(ctx, req)
		if err != nil {
			if provider.Name() == "heygen_v3" && strings.Contains(err.Error(), "status 429") {
				fallback := GetAvatarProvider("tavus_v2")
				log.Warn("heygen rate limited, falling back", "from", provider.Name(), "to", fallback.Name())
				jobID, err = fallback.CreateLipSync(ctx, req)
				if err != nil {
					log.Error("avatar fallback create failed", "provider", fallback.Name(), "err", err)
					return fmt.Errorf("avatar create fallback (%s): %w", fallback.Name(), err)
				}
				provider = fallback
			} else {
				log.Error("avatar create failed", "provider", provider.Name(), "err", err)
				return fmt.Errorf("avatar create (%s): %w", provider.Name(), err)
			}
		}
		log.Info("avatar job submitted", "provider", provider.Name(), "avatar_job_id", jobID)

		// Persist avatar_job_id to DB so webhook can route back if needed
		if err := patchAvatarJobID(ctx, p.VariantID, p.WorkspaceID, jobID, provider.Name()); err != nil {
			log.Warn("patch avatar_job_id failed (non-fatal)", "err", err)
		}

		// Poll for completion (with automatic Tavus fallback on HeyGen 429)
		result, err := PollAvatarWithFallback(ctx, provider, jobID, req)
		if err != nil {
			log.Error("avatar poll failed", "err", err)
			return fmt.Errorf("avatar poll: %w", err)
		}

		elapsed := time.Since(start)
		log.Info("avatar completed",
			"provider", result.Provider,
			"duration_ms", elapsed.Milliseconds(),
			"video_url", result.VideoURL,
		)

		// Record performance event
		recordPerfEvent(ctx, rdb, PerfEvent{
			WorkspaceID: p.WorkspaceID,
			VariantID:   p.VariantID,
			JobID:       p.JobID,
			Stage:       "avatar_lipsync",
			DurationMS:  int(elapsed.Milliseconds()),
		})

		// Hand off to postprocess task (watermark, captions, transcode, R2 upload)
		postPayload := PostprocessPayload{
			JobID:                p.JobID,
			VariantID:            p.VariantID,
			WorkspaceID:          p.WorkspaceID,
			PostprocessRequestID: "",
			InputR2Key:           result.VideoURL,
			OutputR2Key:          fmt.Sprintf("jobs/%s/variants/%s/processed.mp4", p.JobID, p.VariantID),
			Watermark:            true,
			AddCaptions:          false,
			Script:               "",
		}
		postTask, err := NewPostprocessTask(postPayload)
		if err != nil {
			return fmt.Errorf("new postprocess task: %w", err)
		}

		client := asynq.NewClientFromRedisClient(rdb)
		defer client.Close()
		if _, err := client.EnqueueContext(ctx, postTask, asynq.Queue("critical")); err != nil {
			return fmt.Errorf("enqueue postprocess: %w", err)
		}

		log.Info("avatar → postprocess enqueued")
		return nil
	}
}

// patchAvatarJobID stores the provider job ID on the variant row.
func patchAvatarJobID(ctx context.Context, variantID, workspaceID, avatarJobID, providerName string) error {
	apiBase := getEnv("API_BASE_URL")
	internalKey := getEnv("INTERNAL_API_KEY")

	body, _ := json.Marshal(map[string]string{
		"avatar_job_id":   avatarJobID,
		"avatar_provider": providerName,
	})

	return patchJSON(ctx,
		fmt.Sprintf("%s/api/v1/variants/%s/avatar-job", apiBase, variantID),
		workspaceID, internalKey, body,
	)
}
