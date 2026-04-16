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
)

// =============================================================================
// VideoProvider interface
// Abstracts T2V generation backends (FAL.AI default).
// All callers use VideoProvider — never call FAL.AI directly from task logic.
// Always uses async queue submission — NEVER blocking subscribe.
// =============================================================================

// VideoRequest is the input for T2V generation.
type VideoRequest struct {
	JobID          string
	VariantID      string
	WorkspaceID    string
	PlanTier       string
	Angle          string
	Script         string
	Model          string // veo3 | kling3 | runway4 | sora2
	UseAvatar      bool
	AudioURL       string
	AvatarProvider string
}

// VideoResult is returned when a video generation job is submitted.
type VideoResult struct {
	// ProviderJobID is the job ID on the provider's system (e.g. FAL request_id).
	ProviderJobID string
	// Provider is the identifier of the provider that processed this request.
	Provider string
}

// VideoProvider abstracts T2V generation backends.
type VideoProvider interface {
	// Name returns the provider identifier.
	Name() string
	// Submit enqueues a video generation job and returns a VideoResult.
	Submit(ctx context.Context, req VideoRequest) (VideoResult, error)
}

// VideoRegistry maps model names → provider.
var VideoRegistry = map[string]VideoProvider{}

func init() {
	fal := NewFalProvider()
	for _, model := range []string{"veo3", "kling3", "runway4", "sora2"} {
		VideoRegistry[model] = fal
	}
}

// GetVideoProvider returns the provider for the given model, or nil if unknown.
func GetVideoProvider(model string) VideoProvider {
	return VideoRegistry[model]
}

// =============================================================================
// FAL.AI Provider
// Always uses fal.queue.submit() — NEVER fal.subscribe() (blocks under load).
// =============================================================================

// falModelEndpoints maps Qvora model names → FAL.AI endpoint paths.
var falModelEndpoints = map[string]string{
	"veo3":    "fal-ai/veo3",
	"kling3":  "fal-ai/kling-video/v3/standard/text-to-video",
	"runway4": "fal-ai/runway-gen4/turbo/text-to-video",
	"sora2":   "fal-ai/sora",
}

// FalProvider implements VideoProvider for FAL.AI.
type FalProvider struct{}

// NewFalProvider returns a new FalProvider.
func NewFalProvider() *FalProvider { return &FalProvider{} }

func (f *FalProvider) Name() string { return "fal" }

// Submit enqueues a video generation job on FAL.AI's async queue.
func (f *FalProvider) Submit(ctx context.Context, req VideoRequest) (VideoResult, error) {
	falKey := os.Getenv("FAL_KEY")
	if falKey == "" {
		return VideoResult{}, fmt.Errorf("FAL_KEY not set")
	}

	endpoint, ok := falModelEndpoints[req.Model]
	if !ok {
		return VideoResult{}, fmt.Errorf("unknown model: %s", req.Model)
	}

	// Resolve webhook URL: explicit env var, or derive from API_BASE_URL
	webhookURL := strings.TrimSpace(os.Getenv("FAL_WEBHOOK_URL"))
	if webhookURL == "" {
		if apiBaseURL := strings.TrimSpace(os.Getenv("API_BASE_URL")); apiBaseURL != "" {
			webhookURL = strings.TrimRight(apiBaseURL, "/") + "/webhooks/fal"
		}
	}

	avatarProvider := req.AvatarProvider
	if avatarProvider == "" {
		avatarProvider = "heygen_v3"
	}

	body, err := json.Marshal(map[string]any{
		"prompt": req.Script,
		"input": map[string]any{
			"prompt":           req.Script,
			"aspect_ratio":     "9:16",
			"duration_seconds": 15,
		},
		"webhook_url": webhookURL,
		"metadata": map[string]any{
			"job_id":          req.JobID,
			"variant_id":      req.VariantID,
			"workspace_id":    req.WorkspaceID,
			"use_avatar":      req.UseAvatar,
			"audio_url":       req.AudioURL,
			"avatar_provider": avatarProvider,
		},
	})
	if err != nil {
		return VideoResult{}, fmt.Errorf("marshal fal request: %w", err)
	}

	queueURL := "https://queue.fal.run/" + strings.TrimPrefix(endpoint, "/")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, queueURL, bytes.NewReader(body))
	if err != nil {
		return VideoResult{}, fmt.Errorf("build fal request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Key "+falKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return VideoResult{}, fmt.Errorf("submit fal request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return VideoResult{}, fmt.Errorf("read fal response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return VideoResult{}, fmt.Errorf("fal HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		RequestID string `json:"request_id"`
		StatusURL string `json:"status_url"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return VideoResult{}, fmt.Errorf("parse fal response: %w", err)
	}
	if out.Error != "" {
		return VideoResult{}, fmt.Errorf("fal API error: %s", out.Error)
	}
	if out.RequestID == "" {
		return VideoResult{}, fmt.Errorf("fal response missing request_id")
	}

	return VideoResult{ProviderJobID: out.RequestID, Provider: f.Name()}, nil
}
