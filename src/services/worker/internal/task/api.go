package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// patchJobStatus calls the Go API PATCH /api/v1/jobs/:id/status.
// Workers call this to report status transitions back to the orchestrator.
// Non-fatal errors are logged by the caller — a failed callback never aborts the task.
func patchJobStatus(jobID, workspaceID, status string) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	body, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("marshal status request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/api/v1/jobs/%s/status", apiBase, jobID),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Internal service-to-service auth (not Clerk JWT)
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", workspaceID)
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("patch job status request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("patch job status %q returned HTTP %d", status, resp.StatusCode)
	}
	return nil
}

// patchVariantFalRequestID stores the FAL request ID on a variant record via
// PATCH /api/v1/variants/:id/fal-request. Non-fatal — caller logs on error.
func patchVariantFalRequestID(ctx context.Context, variantID, workspaceID, falRequestID string) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	body, err := json.Marshal(map[string]string{"fal_request_id": falRequestID})
	if err != nil {
		return fmt.Errorf("marshal fal-request body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		fmt.Sprintf("%s/api/v1/variants/%s/fal-request", apiBase, variantID),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build fal-request patch: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", workspaceID)
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("patch fal-request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("patch fal-request returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func refreshAllSignalRecommendations(ctx context.Context, days int) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	body, err := json.Marshal(map[string]int{"days": days})
	if err != nil {
		return fmt.Errorf("marshal refresh recommendations body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/internal/signal/recommendations/refresh-all?days=%d", apiBase, days),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build refresh recommendations request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", "system")
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh recommendations request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("refresh recommendations returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func syncSignalMetrics(ctx context.Context) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/internal/signal/metrics/sync-all", apiBase),
		bytes.NewReader([]byte(`{}`)),
	)
	if err != nil {
		return fmt.Errorf("build sync metrics request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", "system")
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sync metrics request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sync metrics returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func runSignalGDPRCleanup(ctx context.Context) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/internal/signal/gdpr/cleanup", apiBase),
		bytes.NewReader([]byte(`{}`)),
	)
	if err != nil {
		return fmt.Errorf("build gdpr cleanup request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", "system")
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gdpr cleanup request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gdpr cleanup returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func runJobStuckReconciliation(ctx context.Context, maxAgeMinutes int) error {
	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		return fmt.Errorf("API_BASE_URL not set")
	}
	apiKey := os.Getenv("INTERNAL_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}
	if maxAgeMinutes <= 0 {
		maxAgeMinutes = 45
	}

	body, _ := json.Marshal(map[string]int{"max_age_minutes": maxAgeMinutes})

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/internal/jobs/reconcile-stuck?max_age_minutes=%d", apiBase, maxAgeMinutes),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build reconcile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", "system")
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("job reconcile request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("job reconcile returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func patchJSON(ctx context.Context, url, workspaceID, internalKey string, body []byte) error {
	if internalKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build patch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", workspaceID)
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", internalKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("patch request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("patch request returned HTTP %d", resp.StatusCode)
	}

	return nil
}

func postJSON(ctx context.Context, url, workspaceID, internalKey string, body []byte) error {
	if internalKey == "" {
		return fmt.Errorf("INTERNAL_API_KEY not set")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("build post request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "worker")
	req.Header.Set("X-Org-Id", workspaceID)
	req.Header.Set("X-Org-Role", "worker")
	req.Header.Set("X-Internal-Api-Key", internalKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("post request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return fmt.Errorf("post request returned HTTP %d", resp.StatusCode)
	}

	return nil
}
