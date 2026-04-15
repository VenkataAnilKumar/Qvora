package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/qvora/api/internal/db"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

const scrapeTaskType = "job:scrape"

// CreateBrief godoc
// POST /api/v1/briefs
func CreateBrief(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		ProductURL string `json:"product_url"`
		Template   string `json:"template"`
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

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	workspaceID, err := workspaceIDForOrg(c.Request().Context(), q, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}

	job, err := q.CreateJob(c.Request().Context(), db.CreateJobParams{
		WorkspaceID: workspaceID,
		ProductUrl:  req.ProductURL,
		Model:       "veo3",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "job_create_failed"})
	}

	brief, err := q.CreateBrief(c.Request().Context(), db.CreateBriefParams{
		WorkspaceID: workspaceID,
		ScrapeJobID: job.ID,
		ProductUrl:  req.ProductURL,
		Model:       "veo3",
		Status:      "queued",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_create_failed"})
	}

	if err := enqueueScrapeTask(uuidString(job.ID), claims.OrgID, req.ProductURL); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_enqueue_failed"})
	}

	updatedBrief, err := q.UpdateBriefStatus(c.Request().Context(), db.UpdateBriefStatusParams{
		ID:     brief.ID,
		Status: "scraping",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_update_failed"})
	}

	if _, err := q.UpdateJobStatus(c.Request().Context(), db.UpdateJobStatusParams{
		ID:     job.ID,
		Status: "scraping",
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "job_update_failed"})
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"brief_id":      uuidString(updatedBrief.ID),
		"scrape_job_id": uuidString(updatedBrief.ScrapeJobID),
		"org_id":        claims.OrgID,
		"product_url":   updatedBrief.ProductUrl,
		"template":      strings.TrimSpace(req.Template),
		"status":        updatedBrief.Status,
		"created_at":    tsTime(updatedBrief.CreatedAt),
	})
}

// ListBriefs godoc
// GET /api/v1/briefs
func ListBriefs(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	workspaceID, err := workspaceIDForOrg(c.Request().Context(), q, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}

	rows, err := q.ListBriefsByWorkspace(c.Request().Context(), db.ListBriefsByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       50,
		Offset:      0,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "briefs_list_failed"})
	}

	briefs := make([]map[string]any, 0, len(rows))
	for _, brief := range rows {
		briefs = append(briefs, map[string]any{
			"brief_id":      uuidString(brief.ID),
			"scrape_job_id": uuidString(brief.ScrapeJobID),
			"org_id":        claims.OrgID,
			"product_url":   brief.ProductUrl,
			"status":        brief.Status,
			"created_at":    tsTime(brief.CreatedAt),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id": claims.OrgID,
		"briefs": briefs,
	})
}

func enqueueScrapeTask(jobID, workspaceID, productURL string) error {
	redisURL := strings.TrimSpace(os.Getenv("RAILWAY_REDIS_URL"))
	if redisURL == "" {
		return fmt.Errorf("RAILWAY_REDIS_URL not set")
	}

	redisOpt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return fmt.Errorf("parse redis uri: %w", err)
	}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	payload, err := json.Marshal(map[string]string{
		"job_id":       jobID,
		"workspace_id": workspaceID,
		"product_url":  productURL,
	})
	if err != nil {
		return fmt.Errorf("marshal scrape payload: %w", err)
	}

	task := asynq.NewTask(scrapeTaskType, payload, asynq.Queue("default"))
	if _, err := client.Enqueue(task); err != nil {
		return fmt.Errorf("enqueue scrape task: %w", err)
	}

	return nil
}
