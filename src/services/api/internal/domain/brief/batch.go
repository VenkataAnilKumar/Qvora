package brief

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
	"github.com/qvora/api/internal/domain/media"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// generateTaskType mirrors the worker's TypeGenerate constant.
// TODO: consolidate into a shared constants package once domain/media is complete.
const generateTaskType = "job:generate"

// allowedModels delegates to domain/media.AllowedModels — single source of truth.

type batchGenerateTaskPayload struct {
	JobID       string `json:"job_id"`
	VariantID   string `json:"variant_id"`
	WorkspaceID string `json:"workspace_id"`
	Angle       string `json:"angle"`
	Script      string `json:"script"`
	Model       string `json:"model"`
}

// BatchGenerateVariants handles POST /api/v1/briefs/:briefId/batch-generate
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
			Angle string `json:"angle"`
			Hook  string `json:"hook"`
			Model string `json:"model"`
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
	for i := range req.Specs {
		req.Specs[i].Model = strings.ToLower(strings.TrimSpace(req.Specs[i].Model))
		if req.Specs[i].Model == "" {
			req.Specs[i].Model = "veo3"
		}
		if _, ok := media.AllowedModels[req.Specs[i].Model]; !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("invalid_model for spec %d: %s", i, req.Specs[i].Model),
			})
		}
		req.Specs[i].Angle = strings.TrimSpace(req.Specs[i].Angle)
		req.Specs[i].Hook = strings.TrimSpace(req.Specs[i].Hook)
	}
	workspaceID, planTier, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	totalRequested := len(req.Specs) * req.VariantsPerSpec
	approvedPerSpec, tierErr := appmiddleware.EnforceVariantLimit(req.VariantsPerSpec, planTier)
	if tierErr != nil {
		if httpErr, ok := tierErr.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, httpErr.Message)
		}
		return c.JSON(http.StatusPaymentRequired, map[string]string{"error": "variant_limit_exceeded"})
	}
	parsedBriefID, err := util.ParseUUID(briefID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_brief_id"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	b, err := q.GetBriefByID(c.Request().Context(), db.GetBriefByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "brief_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}
	createdJobs := make([]map[string]any, 0, len(req.Specs))
	for _, spec := range req.Specs {
		job, err := q.CreateJob(c.Request().Context(), db.CreateJobParams{
			WorkspaceID: workspaceID,
			ProductUrl:  b.ProductUrl,
			Model:       spec.Model,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "job_create_failed"})
		}
		variants := make([]map[string]any, 0, approvedPerSpec)
		for i := 0; i < approvedPerSpec; i++ {
			variant, err := q.CreateVariant(c.Request().Context(), db.CreateVariantParams{
				JobID:       job.ID,
				WorkspaceID: workspaceID,
				Angle:       spec.Angle,
			})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_create_failed"})
			}
			if err := enqueueBatchGenerateTask(batchGenerateTaskPayload{
				JobID:       util.UUIDString(job.ID),
				VariantID:   util.UUIDString(variant.ID),
				WorkspaceID: util.UUIDString(workspaceID),
				Angle:       spec.Angle,
				Script:      spec.Hook,
				Model:       spec.Model,
			}); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "batch_enqueue_failed"})
			}
			variants = append(variants, map[string]any{
				"variant_id": util.UUIDString(variant.ID),
				"status":     variant.Status,
			})
		}
		createdJobs = append(createdJobs, map[string]any{
			"job_id":            util.UUIDString(job.ID),
			"angle":             spec.Angle,
			"hook":              spec.Hook,
			"model":             spec.Model,
			"variants_per_spec": approvedPerSpec,
			"status":            job.Status,
			"variants":          variants,
		})
	}
	return c.JSON(http.StatusAccepted, map[string]any{
		"org_id":            claims.OrgID,
		"workspace_id":      util.UUIDString(workspaceID),
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
	t := asynq.NewTask(generateTaskType, data, asynq.Queue("default"), asynq.MaxRetry(3))
	if _, err := client.Enqueue(t); err != nil {
		return fmt.Errorf("enqueue generate task: %w", err)
	}
	return nil
}
