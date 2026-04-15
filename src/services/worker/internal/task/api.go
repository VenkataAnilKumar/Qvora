package task

import (
	"bytes"
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

	body, _ := json.Marshal(map[string]string{"status": status})

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
