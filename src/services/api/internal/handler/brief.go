package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

const scrapeTaskType = "job:scrape"

type inMemoryBrief struct {
	BriefID     string    `json:"brief_id"`
	ScrapeJobID string    `json:"scrape_job_id"`
	OrgID       string    `json:"org_id"`
	ProductURL  string    `json:"product_url"`
	Template    string    `json:"template,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

var briefsStore = struct {
	sync.RWMutex
	byID map[string]inMemoryBrief
}{
	byID: make(map[string]inMemoryBrief),
}

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

	brief := inMemoryBrief{
		BriefID:     uuid.NewString(),
		ScrapeJobID: uuid.NewString(),
		OrgID:       claims.OrgID,
		ProductURL:  req.ProductURL,
		Template:    strings.TrimSpace(req.Template),
		Status:      "queued",
		CreatedAt:   time.Now().UTC(),
	}

	briefsStore.Lock()
	briefsStore.byID[brief.BriefID] = brief
	briefsStore.Unlock()

	if err := enqueueScrapeTask(brief.ScrapeJobID, brief.OrgID, brief.ProductURL); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_enqueue_failed"})
	}

	brief.Status = "scraping"
	briefsStore.Lock()
	briefsStore.byID[brief.BriefID] = brief
	briefsStore.Unlock()

	jobsStore.Lock()
	jobsStore.byID[brief.ScrapeJobID] = inMemoryJob{
		JobID:      brief.ScrapeJobID,
		OrgID:      brief.OrgID,
		ProductURL: brief.ProductURL,
		Model:      "veo3",
		Status:     "scraping",
		CreatedAt:  brief.CreatedAt,
		UpdatedAt:  time.Now().UTC(),
	}
	jobsStore.Unlock()

	return c.JSON(http.StatusAccepted, brief)
}

// ListBriefs godoc
// GET /api/v1/briefs
func ListBriefs(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	briefsStore.RLock()
	briefs := make([]inMemoryBrief, 0, len(briefsStore.byID))
	for _, brief := range briefsStore.byID {
		if brief.OrgID == claims.OrgID {
			briefs = append(briefs, brief)
		}
	}
	briefsStore.RUnlock()

	sort.Slice(briefs, func(i, j int) bool {
		return briefs[i].CreatedAt.After(briefs[j].CreatedAt)
	})

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
