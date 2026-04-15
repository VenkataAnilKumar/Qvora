package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

type inMemoryJob struct {
	JobID      string    `json:"job_id"`
	OrgID      string    `json:"org_id"`
	ProductURL string    `json:"product_url"`
	Model      string    `json:"model"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var jobsStore = struct {
	sync.RWMutex
	byID map[string]inMemoryJob
}{
	byID: make(map[string]inMemoryJob),
}

var allowedModels = map[string]struct{}{
	"veo3":    {},
	"kling3":  {},
	"runway4": {},
	"sora2":   {},
}

var allowedJobStatuses = map[string]struct{}{
	"queued":         {},
	"scraping":       {},
	"briefing":       {},
	"generating":     {},
	"postprocessing": {},
	"complete":       {},
	"failed":         {},
}

// SubmitJob godoc
// POST /api/v1/jobs
func SubmitJob(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		ProductURL       string `json:"product_url"`
		Model            string `json:"model"`
		VariantsPerAngle int    `json:"variants_per_angle"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	req.ProductURL = strings.TrimSpace(req.ProductURL)
	if req.ProductURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "product_url_required"})
	}
	if _, err := url.ParseRequestURI(req.ProductURL); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_product_url"})
	}

	req.Model = strings.ToLower(strings.TrimSpace(req.Model))
	if req.Model == "" {
		req.Model = "veo3"
	}
	if _, ok := allowedModels[req.Model]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_model"})
	}

	if req.VariantsPerAngle <= 0 {
		req.VariantsPerAngle = 3
	}

	planTier := getWorkspacePlanTier(claims.OrgID)

	approvedVariants, err := appmiddleware.EnforceVariantLimit(req.VariantsPerAngle, planTier)
	if err != nil {
		if httpErr, ok := err.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, httpErr.Message)
		}
		return c.JSON(http.StatusPaymentRequired, map[string]string{"error": "variant_limit_exceeded"})
	}

	job := inMemoryJob{
		JobID:      uuid.NewString(),
		OrgID:      claims.OrgID,
		ProductURL: req.ProductURL,
		Model:      req.Model,
		Status:     "queued",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	jobsStore.Lock()
	jobsStore.byID[job.JobID] = job
	jobsStore.Unlock()

	// TODO: enqueue asynq task to Railway Redis (TCP)
	// scrape task → generation task → postprocess task

	return c.JSON(http.StatusAccepted, map[string]any{
		"job_id":             job.JobID,
		"org_id":             job.OrgID,
		"status":             job.Status,
		"product_url":        job.ProductURL,
		"model":              job.Model,
		"variants_per_angle": approvedVariants,
		"plan_tier":          planTier,
		"created_at":         job.CreatedAt,
		"updated_at":         job.UpdatedAt,
	})
}

// GetJob godoc
// GET /api/v1/jobs/:id
func GetJob(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	jobID := c.Param("id")

	jobsStore.RLock()
	job, ok := jobsStore.byID[jobID]
	jobsStore.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "job_not_found"})
	}
	if job.OrgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"job_id":      jobID,
		"org_id":      job.OrgID,
		"status":      job.Status,
		"product_url": job.ProductURL,
		"model":       job.Model,
		"created_at":  job.CreatedAt,
		"updated_at":  job.UpdatedAt,
	})
}

// UpdateJobStatus godoc
// PATCH /api/v1/jobs/:id/status
func UpdateJobStatus(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	jobID := strings.TrimSpace(c.Param("id"))
	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	req.Status = strings.TrimSpace(strings.ToLower(req.Status))
	if _, ok := allowedJobStatuses[req.Status]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}

	jobsStore.Lock()
	job, ok := jobsStore.byID[jobID]
	if !ok {
		jobsStore.Unlock()
		return c.JSON(http.StatusNotFound, map[string]string{"error": "job_not_found"})
	}
	if job.OrgID != claims.OrgID {
		jobsStore.Unlock()
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	job.Status = req.Status
	job.UpdatedAt = time.Now().UTC()
	jobsStore.byID[jobID] = job
	jobsStore.Unlock()

	return c.JSON(http.StatusOK, map[string]any{
		"job_id":      job.JobID,
		"org_id":      job.OrgID,
		"status":      job.Status,
		"product_url": job.ProductURL,
		"model":       job.Model,
		"created_at":  job.CreatedAt,
		"updated_at":  job.UpdatedAt,
	})
}

// ListJobs godoc
// GET /api/v1/jobs
func ListJobs(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	limit := 20
	if rawLimit := strings.TrimSpace(c.QueryParam("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_limit"})
		}
		if parsed > 50 {
			parsed = 50
		}
		limit = parsed
	}

	jobsStore.RLock()
	jobs := make([]inMemoryJob, 0, len(jobsStore.byID))
	for _, job := range jobsStore.byID {
		if job.OrgID == claims.OrgID {
			jobs = append(jobs, job)
		}
	}
	jobsStore.RUnlock()

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	if len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return c.JSON(http.StatusOK, map[string]any{
		"jobs":        jobs,
		"org_id":      claims.OrgID,
		"next_cursor": nil,
	})
}

// StreamJob godoc
// GET /api/v1/jobs/:id/stream — SSE endpoint
func StreamJob(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	jobID := c.Param("id")

	jobsStore.RLock()
	job, ok := jobsStore.byID[jobID]
	jobsStore.RUnlock()
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "job_not_found"})
	}
	if job.OrgID != claims.OrgID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	type ssePayload struct {
		Type      string `json:"type"`
		JobID     string `json:"job_id"`
		Status    string `json:"status"`
		Message   string `json:"message"`
		Progress  int    `json:"progress"`
		Timestamp string `json:"timestamp"`
	}

	ctx := c.Request().Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			jobsStore.RLock()
			current, exists := jobsStore.byID[jobID]
			jobsStore.RUnlock()
			if !exists {
				return nil
			}

			eventType, progress, message := statusEvent(current.Status)
			data, err := json.Marshal(ssePayload{
				Type:      eventType,
				JobID:     jobID,
				Status:    current.Status,
				Message:   message,
				Progress:  progress,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			})
			if err != nil {
				return nil
			}

			_, _ = fmt.Fprintf(c.Response(), "data: %s\n\n", data)
			c.Response().Flush()

			if current.Status == "complete" || current.Status == "failed" {
				return nil
			}
		}
	}

}

func statusEvent(status string) (eventType string, progress int, message string) {
	switch status {
	case "queued":
		return "job_queued", 5, "job accepted"
	case "scraping":
		return "scraping_started", 20, "scraping product page"
	case "briefing":
		return "brief_started", 55, "generating brief"
	case "generating":
		return "generation_started", 75, "generating variants"
	case "postprocessing":
		return "postprocessing_started", 90, "postprocessing video"
	case "complete":
		return "job_complete", 100, "job complete"
	case "failed":
		return "job_failed", 100, "job failed"
	default:
		return "job_queued", 5, "job accepted"
	}
}
