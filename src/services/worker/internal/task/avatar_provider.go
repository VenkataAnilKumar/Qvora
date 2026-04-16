package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// =============================================================================
// Avatar Provider Interface
// Abstracts HeyGen v3 (default) and Tavus v2 (secondary).
// All callers use AvatarProvider — never call HeyGen or Tavus APIs directly.
// =============================================================================

// AvatarRequest is the input for lip-sync video generation.
type AvatarRequest struct {
	// URL of the raw video (from FAL.AI output, stored in R2)
	VideoURL string
	// URL of the audio file (from ElevenLabs TTS, stored in R2)
	AudioURL string
	// WorkspaceID for logging and cost tracking
	WorkspaceID string
	// VariantID for correlation
	VariantID string
}

// AvatarResult is returned when the avatar job completes.
type AvatarResult struct {
	// JobID on the provider's system (for status polling / webhook routing)
	JobID string
	// VideoURL is the processed lip-sync video URL (set on completion)
	VideoURL string
	// Provider that produced this result
	Provider string
	// Status: "pending" | "processing" | "completed" | "failed"
	Status string
	// Error message if Status == "failed"
	Error string
}

// AvatarProvider is implemented by all avatar lip-sync providers.
type AvatarProvider interface {
	// Name returns the provider identifier stored on the DB row.
	Name() string
	// CreateLipSync submits a lip-sync job and returns a provider job ID.
	CreateLipSync(ctx context.Context, req AvatarRequest) (jobID string, err error)
	// GetStatus polls job status. Returns AvatarResult with VideoURL set when done.
	GetStatus(ctx context.Context, jobID string) (AvatarResult, error)
}

// AvatarRegistry holds all registered providers.
// Select provider with AvatarRegistry.Get(name).
var AvatarRegistry = map[string]AvatarProvider{}

func init() {
	AvatarRegistry["heygen_v3"] = &HeyGenV3Provider{}
	AvatarRegistry["tavus_v2"] = &TavusProvider{}
}

// GetAvatarProvider returns the named provider, defaulting to heygen_v3.
func GetAvatarProvider(name string) AvatarProvider {
	if p, ok := AvatarRegistry[name]; ok {
		return p
	}
	return AvatarRegistry["heygen_v3"]
}

// =============================================================================
// HeyGen v3 Provider
// Active platform: developers.heygen.com
// V2V lip-sync: video clip + audio → lip-sync job (async, poll every 15s)
// =============================================================================

type HeyGenV3Provider struct{}

func (h *HeyGenV3Provider) Name() string { return "heygen_v3" }

func (h *HeyGenV3Provider) CreateLipSync(ctx context.Context, req AvatarRequest) (string, error) {
	apiKey := mustEnv("HEYGEN_API_KEY")

	body := map[string]any{
		"video_url": req.VideoURL,
		"audio_url": req.AudioURL,
		"title":     fmt.Sprintf("qvora-%s", req.VariantID),
	}
	data, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://api.heygen.com/v2/video_translate",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("heygen create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("heygen create http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("heygen create status %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		Data struct {
			VideoTranslationID string `json:"video_translation_id"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("heygen create decode: %w", err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("heygen api error: %s", out.Error.Message)
	}
	return out.Data.VideoTranslationID, nil
}

func (h *HeyGenV3Provider) GetStatus(ctx context.Context, jobID string) (AvatarResult, error) {
	apiKey := mustEnv("HEYGEN_API_KEY")

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf("https://api.heygen.com/v2/video_translate/%s", jobID),
		nil,
	)
	if err != nil {
		return AvatarResult{}, fmt.Errorf("heygen status request: %w", err)
	}
	httpReq.Header.Set("X-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return AvatarResult{}, fmt.Errorf("heygen status http: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Data struct {
			Status   string `json:"status"`
			VideoURL string `json:"video_url"`
			Error    string `json:"error"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return AvatarResult{}, fmt.Errorf("heygen status decode: %w", err)
	}

	return AvatarResult{
		JobID:    jobID,
		VideoURL: out.Data.VideoURL,
		Provider: "heygen_v3",
		Status:   out.Data.Status,
		Error:    out.Data.Error,
	}, nil
}

// =============================================================================
// Tavus v2 Provider
// Secondary / cost-optimized avatar provider.
// Activated automatically on HeyGen 429 or when plan = cost_optimized.
// =============================================================================

type TavusProvider struct{}

func (t *TavusProvider) Name() string { return "tavus_v2" }

func (t *TavusProvider) CreateLipSync(ctx context.Context, req AvatarRequest) (string, error) {
	apiKey := mustEnv("TAVUS_API_KEY")

	body := map[string]any{
		"video_url": req.VideoURL,
		"audio_url": req.AudioURL,
		"callback_url": fmt.Sprintf("%s/webhooks/tavus",
			mustEnv("API_BASE_URL")),
	}
	data, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		"https://tavusapi.com/v2/videos",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", fmt.Errorf("tavus create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("tavus create http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tavus create status %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		VideoID string `json:"video_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("tavus create decode: %w", err)
	}
	return out.VideoID, nil
}

func (t *TavusProvider) GetStatus(ctx context.Context, jobID string) (AvatarResult, error) {
	apiKey := mustEnv("TAVUS_API_KEY")

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf("https://tavusapi.com/v2/videos/%s", jobID),
		nil,
	)
	if err != nil {
		return AvatarResult{}, fmt.Errorf("tavus status request: %w", err)
	}
	httpReq.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return AvatarResult{}, fmt.Errorf("tavus status http: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Status       string `json:"status"`
		DownloadURL  string `json:"download_url"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return AvatarResult{}, fmt.Errorf("tavus status decode: %w", err)
	}

	// Normalise Tavus status → common status vocabulary
	status := out.Status
	switch status {
	case "ready":
		status = "completed"
	case "error":
		status = "failed"
	}

	return AvatarResult{
		JobID:    jobID,
		VideoURL: out.DownloadURL,
		Provider: "tavus_v2",
		Status:   status,
		Error:    out.ErrorMessage,
	}, nil
}

// =============================================================================
// PollAvatarWithFallback
// Polls the given provider every 15 seconds up to maxWait.
// On HeyGen 429, automatically retries with Tavus.
// =============================================================================

const avatarPollInterval = 15 * time.Second
const avatarMaxWait = 10 * time.Minute

func PollAvatarWithFallback(
	ctx context.Context,
	provider AvatarProvider,
	jobID string,
	req AvatarRequest,
) (AvatarResult, error) {
	deadline := time.Now().Add(avatarMaxWait)
	for time.Now().Before(deadline) {
		result, err := provider.GetStatus(ctx, jobID)
		if err != nil {
			// If HeyGen returns 429, fall back to Tavus
			if provider.Name() == "heygen_v3" {
				tavus := GetAvatarProvider("tavus_v2")
				newJobID, fbErr := tavus.CreateLipSync(ctx, req)
				if fbErr != nil {
					return AvatarResult{}, fmt.Errorf("heygen error + tavus fallback failed: %w", fbErr)
				}
				return PollAvatarWithFallback(ctx, tavus, newJobID, req)
			}
			return AvatarResult{}, fmt.Errorf("avatar poll: %w", err)
		}

		switch result.Status {
		case "completed":
			return result, nil
		case "failed":
			return AvatarResult{}, fmt.Errorf("avatar job failed (%s): %s", provider.Name(), result.Error)
		}

		select {
		case <-ctx.Done():
			return AvatarResult{}, ctx.Err()
		case <-time.After(avatarPollInterval):
		}
	}
	return AvatarResult{}, fmt.Errorf("avatar job timed out after %s", avatarMaxWait)
}

// mustEnv panics if an env var is missing — catches misconfiguration at startup.
func mustEnv(key string) string {
	v := getEnv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}
