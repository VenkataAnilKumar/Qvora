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

// ScrapePayload is the input to the scrape task
type ScrapePayload struct {
	JobID       string `json:"job_id"`
	WorkspaceID string `json:"workspace_id"`
	ProductURL  string `json:"product_url"`
}

// ProductExtraction is the structured response from the Modal Playwright scraper
type ProductExtraction struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Price       string   `json:"price"`
	Features    []string `json:"features"`
	ProofPoints []string `json:"proof_points"`
	ImageURLs   []string `json:"image_urls"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
}

// NewScrapeTask creates an asynq task for URL scraping via Modal Playwright
func NewScrapeTask(payload ScrapePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal scrape payload: %w", err)
	}
	return asynq.NewTask(TypeScrape, data, asynq.Queue("default"), asynq.MaxRetry(MaxTaskRetryAttempts)), nil
}

// HandleScrape processes the scrape task.
// Calls Modal serverless Playwright endpoint, parses the ProductExtraction,
// then enqueues generation tasks directly and advances job status to "generating".
func HandleScrape(ctx context.Context, t *asynq.Task) error {
	var payload ScrapePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal scrape payload: %w", err)
	}

	modalEndpoint := os.Getenv("MODAL_SCRAPER_ENDPOINT")
	if modalEndpoint == "" {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("MODAL_SCRAPER_ENDPOINT not set")
	}

	// POST to Modal Playwright scraper
	reqBody, err := json.Marshal(map[string]string{
		"job_id":       payload.JobID,
		"workspace_id": payload.WorkspaceID,
		"product_url":  payload.ProductURL,
	})
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("marshal scraper request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, modalEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("build scraper request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("scraper request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("scraper returned HTTP %d", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("read scraper response: %w", err)
	}

	// Validate response is parseable.
	var extraction ProductExtraction
	if err := json.Unmarshal(rawBody, &extraction); err != nil {
		_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "failed")
		return fmt.Errorf("parse scraper response: %w", err)
	}

	// Transition: scraping → generating
	_ = patchJobStatus(payload.JobID, payload.WorkspaceID, "generating") // non-fatal

	// Enqueue generation tasks directly (worker-side AI brief stage removed).
	redisURL := os.Getenv("RAILWAY_REDIS_URL")
	if redisURL == "" {
		return fmt.Errorf("RAILWAY_REDIS_URL not set for generate enqueue")
	}
	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return fmt.Errorf("parse redis URL: %w", err)
	}
	client := asynq.NewClient(redisOpt)
	defer client.Close() //nolint:errcheck

	for _, variant := range buildGenerateVariants(payload, extraction) {
		genTask, err := NewGenerateTask(variant)
		if err != nil {
			return fmt.Errorf("create generate task: %w", err)
		}
		if _, err := client.Enqueue(genTask); err != nil {
			return fmt.Errorf("enqueue generate task: %w", err)
		}
	}

	return nil
}

func buildGenerateVariants(payload ScrapePayload, extraction ProductExtraction) []GeneratePayload {
	name := strings.TrimSpace(extraction.Name)
	if name == "" {
		name = "This product"
	}

	description := strings.TrimSpace(extraction.Description)
	feature := "high-impact results"
	if len(extraction.Features) > 0 {
		feature = strings.TrimSpace(extraction.Features[0])
	}
	if feature == "" {
		feature = "high-impact results"
	}

	scriptA := fmt.Sprintf("%s helps you fix the real problem fast. Stop wasting spend on guesswork and start seeing cleaner results with %s. Try it now.", name, feature)
	scriptB := fmt.Sprintf("People switch to %s when they need proof, not promises. Built around %s so every use feels obvious. Tap to get started.", name, feature)
	scriptC := fmt.Sprintf("If speed and performance matter, %s is built for you. %s without the usual friction. Start your first run today.", name, feature)

	if description != "" {
		scriptA = description
	}

	return []GeneratePayload{
		{
			JobID:       payload.JobID,
			VariantID:   fmt.Sprintf("%s-problem", payload.JobID),
			WorkspaceID: payload.WorkspaceID,
			Angle:       "problem_solution",
			Script:      scriptA,
			Model:       "veo3",
		},
		{
			JobID:       payload.JobID,
			VariantID:   fmt.Sprintf("%s-proof", payload.JobID),
			WorkspaceID: payload.WorkspaceID,
			Angle:       "social_proof",
			Script:      scriptB,
			Model:       "veo3",
		},
		{
			JobID:       payload.JobID,
			VariantID:   fmt.Sprintf("%s-urgency", payload.JobID),
			WorkspaceID: payload.WorkspaceID,
			Angle:       "urgency",
			Script:      scriptC,
			Model:       "veo3",
		},
	}
}
