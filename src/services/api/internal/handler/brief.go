package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

// CreateBrief godoc
// POST /api/v1/briefs
// Called by the web tRPC layer after AI generation is complete (scrape + generateObject ×2).
// Persists the brief record and all angles + hooks to the DB.
// No worker enqueue — scraping and generation are done entirely in the Next.js BFF.
func CreateBrief(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		ProductURL string `json:"product_url"`
		Template   string `json:"template"`
		Model      string `json:"model"`
		Angles     []struct {
			Angle     string  `json:"angle"`
			Headline  string  `json:"headline"`
			Script    string  `json:"script"`
			Cta       string  `json:"cta"`
			VoiceTone *string `json:"voice_tone"`
		} `json:"angles"`
		Hooks []string `json:"hooks"`
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

	req.Model = strings.TrimSpace(strings.ToLower(req.Model))
	if req.Model == "" {
		req.Model = "gpt-4o"
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	workspaceID, err := workspaceIDForOrg(c.Request().Context(), q, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}

	// scrape_job_id is intentionally null — scraping is done in the web tRPC layer,
	// not via the asynq worker pipeline. No worker job exists for this brief.
	brief, err := q.CreateBrief(c.Request().Context(), db.CreateBriefParams{
		WorkspaceID: workspaceID,
		ScrapeJobID: pgtype.UUID{}, // null
		ProductUrl:  req.ProductURL,
		Model:       req.Model,
		Status:      "generated",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_create_failed"})
	}

	// Persist angles — fatal: an incomplete brief is worse than a failed request.
	// brief_id is included in error so the caller can clean up or retry.
	for _, a := range req.Angles {
		angle := strings.TrimSpace(a.Angle)
		if angle == "" {
			continue
		}
		if _, err := q.CreateBriefAngle(c.Request().Context(), db.CreateBriefAngleParams{
			BriefID:   brief.ID,
			Angle:     angle,
			Headline:  strings.TrimSpace(a.Headline),
			Script:    strings.TrimSpace(a.Script),
			Cta:       strings.TrimSpace(a.Cta),
			VoiceTone: a.VoiceTone,
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error":    "brief_angle_persist_failed",
				"brief_id": uuidString(brief.ID),
			})
		}
	}

	// Persist hooks — fatal for the same reason.
	for _, h := range req.Hooks {
		hook := strings.TrimSpace(h)
		if hook == "" {
			continue
		}
		if _, err := q.CreateBriefHook(c.Request().Context(), db.CreateBriefHookParams{
			BriefID: brief.ID,
			Hook:    hook,
		}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"error":    "brief_hook_persist_failed",
				"brief_id": uuidString(brief.ID),
			})
		}
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"brief_id":    uuidString(brief.ID),
		"org_id":      claims.OrgID,
		"product_url": brief.ProductUrl,
		"model":       brief.Model,
		"template":    strings.TrimSpace(req.Template),
		"status":      brief.Status,
		"angle_count": len(req.Angles),
		"hook_count":  len(req.Hooks),
		"created_at":  tsTime(brief.CreatedAt),
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
			"brief_id":    uuidString(brief.ID),
			"org_id":      claims.OrgID,
			"product_url": brief.ProductUrl,
			"model":       brief.Model,
			"status":      brief.Status,
			"created_at":  tsTime(brief.CreatedAt),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id": claims.OrgID,
		"briefs": briefs,
	})
}

// GetBrief godoc
// GET /api/v1/briefs/:id
func GetBrief(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	briefID := strings.TrimSpace(c.Param("id"))
	parsedBriefID, err := parseUUID(briefID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_brief_id"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	workspaceID, err := workspaceIDForOrg(c.Request().Context(), q, claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}

	brief, err := q.GetBriefByID(c.Request().Context(), db.GetBriefByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "brief_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}

	angles, _ := q.ListBriefAngles(c.Request().Context(), brief.ID)
	hooks, _ := q.ListBriefHooks(c.Request().Context(), brief.ID)

	angleList := make([]map[string]any, 0, len(angles))
	for _, a := range angles {
		angleList = append(angleList, map[string]any{
			"angle":      a.Angle,
			"headline":   a.Headline,
			"script":     a.Script,
			"cta":        a.Cta,
			"voice_tone": a.VoiceTone,
		})
	}

	hookList := make([]string, 0, len(hooks))
	for _, h := range hooks {
		hookList = append(hookList, h.Hook)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"brief_id":    uuidString(brief.ID),
		"org_id":      claims.OrgID,
		"product_url": brief.ProductUrl,
		"model":       brief.Model,
		"status":      brief.Status,
		"angles":      angleList,
		"hooks":       hookList,
		"created_at":  tsTime(brief.CreatedAt),
		"updated_at":  tsTime(brief.UpdatedAt),
	})
}
