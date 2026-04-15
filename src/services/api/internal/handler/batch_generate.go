package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

const generateTaskType = "job:generate"

type batchGenerateTaskPayload struct {
	JobID       string `json:"job_id"`
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`
	Angle       string `json:"angle"`
	Script      string `json:"script"`
	Model       string `json:"model"`
}

// BatchGenerateVariants queues multiple video generation jobs from a single
// brief. Each spec results in one job with one generation variant.
//
// POST /api/v1/briefs/:briefId/batch-generate
//
//	{
//	  "specs": [{"angle":"problem_solution","hook":"Call out the pain","model":"veo3"}],
//	  "variants_per_spec": 1
//	}
func BatchGenerateVariants(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	briefID := strings.TrimSpace(c.Param("briefId"))
	if briefID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "brief_id_required"})
	}

	var req struct {
		Specs []struct {
			Angle  string `json:"angle"`
			Hook   string `json:"hook"`
			Model  string `json:"model"`
		} `json:"specs"`
		VariantsPerSpec int `json:"variants_per_spec"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if len(req.Specs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "specs_required"})
	}
	if len(req.Specs) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "max_50_specs_per_batch"})
	}
	if req.VariantsPerSpec <= 0 {
		req.VariantsPerSpec = 1
	}

	// Validate and default models in each spec
	for i := range req.Specs {
		req.Specs[i].Model = strings.ToLower(strings.TrimSpace(req.Specs[i].Model))
		if req.Specs[i].Model == "" {
			req.Specs[i].Model = "veo3"
		}
		if _, ok := allowedModels[req.Specs[i].Model]; !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid_model for spec %d: %s", i, req.Specs[i].Model),
			})
		}
		req.Specs[i].Angle = strings.TrimSpace(req.Specs[i].Angle)
		req.Specs[i].Hook = strings.TrimSpace(req.Specs[i].Hook)
	}

	workspaceID, planTier, _, _, err := getWorkspaceForOrg(c, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	// Enforce total variant count against plan tier
	totalRequested := len(req.Specs) * req.VariantsPerSpec
	approvedPerSpec, tierErr := appmiddleware.EnforceVariantLimit(req.VariantsPerSpec, planTier)
	if tierErr != nil {
		if httpErr, ok := tierErr.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, httpErr.Message)
		}
		return c.JSON(http.StatusPaymentRequired, map[string]string{"error": "variant_limit_exceeded"})
	}

	// Resolve brief's product URL from the originating job
	parsedBriefID, parseErr := parseUUID(briefID)
	if parseErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_brief_id"})
	}

	q, qErr := queries(c.Request().Context())
	if qErr != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	sourceJob, dbErr := q.GetJobByID(c.Request().Context(), db.GetJobByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if dbErr != nil {
		if dbErr == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "brief_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}

	// Create one generation job per spec, then queue generation tasks.
	createdJobs := make([]map[string]any, 0, len(req.Specs))
	for _, spec := range req.Specs {
		job, createErr := q.CreateJob(c.Request().Context(), db.CreateJobParams{
			WorkspaceID: workspaceID,
			ProductUrl:  sourceJob.ProductUrl,
			Model:       spec.Model,
		})
		if createErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "job_create_failed"})
		}

		variants := make([]map[string]any, 0, approvedPerSpec)
		for i := 0; i < approvedPerSpec; i++ {
			variant, variantErr := q.CreateVariant(c.Request().Context(), db.CreateVariantParams{
				JobID:       job.ID,
				WorkspaceID: workspaceID,
				Angle:       spec.Angle,
			})
			if variantErr != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_create_failed"})
			}

			if enqueueErr := enqueueBatchGenerateTask(batchGenerateTaskPayload{
				JobID:       uuidString(job.ID),
				VariantID:   uuidString(variant.ID),
				WorkspaceID: uuidString(workspaceID),
				Angle:       spec.Angle,
				Script:      spec.Hook,
				Model:       spec.Model,
			}); enqueueErr != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "batch_enqueue_failed"})
			}

			variants = append(variants, map[string]any{
				"variant_id": uuidString(variant.ID),
				"status":     variant.Status,
			})
		}

		createdJobs = append(createdJobs, map[string]any{
			"job_id":             uuidString(job.ID),
			"angle":              spec.Angle,
			"hook":               spec.Hook,
			"model":              spec.Model,
			"variants_per_spec":  approvedPerSpec,
			"status":             job.Status,
			"variants":           variants,
		})
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"org_id":            claims.OrgID,
		"workspace_id":      uuidString(workspaceID),
		"brief_id":          briefID,
		"plan_tier":         planTier,
		"total_requested":   totalRequested,
		"approved_per_spec": approvedPerSpec,
		"specs_count":       len(req.Specs),
		"jobs":              createdJobs,
		"message":           "batch generation queued",
	})
}

func enqueueBatchGenerateTask(payload batchGenerateTaskPayload) error {
	redisURL := strings.TrimSpace(os.Getenv("RAILWAY_REDIS_URL"))
	if redisURL == "" {
		return fmt.Errorf("RAILWAY_REDIS_URL is required")
	}

	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return fmt.Errorf("parse railway redis url: %w", err)
	}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal generate task payload: %w", err)
	}

	t := asynq.NewTask(generateTaskType, data, asynq.Queue("default"))
	if _, err := client.Enqueue(t); err != nil {
		return fmt.Errorf("enqueue generate task: %w", err)
	}

	return nil
}
