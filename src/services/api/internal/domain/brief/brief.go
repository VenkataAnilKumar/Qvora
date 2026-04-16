package brief

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// BriefAngleInput is the JSON body shape for inline brief edits and creation.
type BriefAngleInput struct {
	Angle     string  `json:"angle"`
	Headline  string  `json:"headline"`
	Script    string  `json:"script"`
	Cta       string  `json:"cta"`
	VoiceTone *string `json:"voice_tone"`
}

// CreateBrief handles POST /api/v1/briefs
// Called by the web tRPC BFF after AI generation; persists brief + angles + hooks.
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
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}
	brief, err := q.CreateBrief(c.Request().Context(), db.CreateBriefParams{
		WorkspaceID: workspaceID,
		ScrapeJobID: pgtype.UUID{}, // null — scraping done in BFF
		ProductUrl:  req.ProductURL,
		Model:       req.Model,
		Status:      "generated",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_create_failed"})
	}
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
				"brief_id": util.UUIDString(brief.ID),
			})
		}
	}
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
				"brief_id": util.UUIDString(brief.ID),
			})
		}
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"brief_id":    util.UUIDString(brief.ID),
		"org_id":      claims.OrgID,
		"product_url": brief.ProductUrl,
		"model":       brief.Model,
		"template":    strings.TrimSpace(req.Template),
		"status":      brief.Status,
		"angle_count": len(req.Angles),
		"hook_count":  len(req.Hooks),
		"created_at":  util.TsTime(brief.CreatedAt),
	})
}

// ListBriefs handles GET /api/v1/briefs
func ListBriefs(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
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
	for _, b := range rows {
		briefs = append(briefs, map[string]any{
			"brief_id":    util.UUIDString(b.ID),
			"org_id":      claims.OrgID,
			"product_url": b.ProductUrl,
			"model":       b.Model,
			"status":      b.Status,
			"created_at":  util.TsTime(b.CreatedAt),
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"org_id": claims.OrgID,
		"briefs": briefs,
	})
}

// GetBrief handles GET /api/v1/briefs/:id
func GetBrief(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	parsedBriefID, err := util.ParseUUID(strings.TrimSpace(c.Param("id")))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_brief_id"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}
	b, err := q.GetBriefByID(c.Request().Context(), db.GetBriefByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "brief_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}
	angles, _ := q.ListBriefAngles(c.Request().Context(), b.ID)
	hooks, _ := q.ListBriefHooks(c.Request().Context(), b.ID)
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
		"brief_id":    util.UUIDString(b.ID),
		"org_id":      claims.OrgID,
		"product_url": b.ProductUrl,
		"model":       b.Model,
		"status":      b.Status,
		"angles":      angleList,
		"hooks":       hookList,
		"created_at":  util.TsTime(b.CreatedAt),
		"updated_at":  util.TsTime(b.UpdatedAt),
	})
}

// UpdateBriefContent handles PUT /api/v1/briefs/:id/content
// Replaces all brief angles/hooks with the provided payload (inline edits).
func UpdateBriefContent(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	parsedBriefID, err := util.ParseUUID(strings.TrimSpace(c.Param("id")))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_brief_id"})
	}
	var req struct {
		Angles []BriefAngleInput `json:"angles"`
		Hooks  []string          `json:"hooks"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if len(req.Angles) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "angles_required"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspaceID, err := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "workspace_resolve_failed"})
	}
	b, err := q.GetBriefByID(c.Request().Context(), db.GetBriefByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "brief_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}
	if err := ReplaceBriefContent(c.Request().Context(), b.ID, req.Angles, req.Hooks); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_content_update_failed"})
	}
	updatedBrief, err := q.GetBriefByID(c.Request().Context(), db.GetBriefByIDParams{
		ID:          parsedBriefID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "brief_lookup_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"brief_id":    util.UUIDString(updatedBrief.ID),
		"org_id":      claims.OrgID,
		"status":      updatedBrief.Status,
		"angle_count": len(req.Angles),
		"hook_count":  len(req.Hooks),
		"updated_at":  util.TsTime(updatedBrief.UpdatedAt),
		"message":     "brief content updated",
	})
}

// ReplaceBriefContent transactionally replaces all angles and hooks for a brief.
// Exported so that other packages (e.g. regen endpoints) can reuse the logic.
func ReplaceBriefContent(ctx context.Context, briefID pgtype.UUID, angles []BriefAngleInput, hooks []string) error {
	pool := store.Pool()
	if pool == nil {
		return errors.New("database_not_initialized")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, "DELETE FROM brief_angles WHERE brief_id = $1", briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM brief_hooks WHERE brief_id = $1", briefID); err != nil {
		return err
	}
	for _, a := range angles {
		angle := strings.TrimSpace(a.Angle)
		if angle == "" {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO brief_angles (brief_id, angle, headline, script, cta, voice_tone)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			briefID, angle,
			strings.TrimSpace(a.Headline),
			strings.TrimSpace(a.Script),
			strings.TrimSpace(a.Cta),
			a.VoiceTone,
		); err != nil {
			return err
		}
	}
	for _, h := range hooks {
		hook := strings.TrimSpace(h)
		if hook == "" {
			continue
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO brief_hooks (brief_id, hook) VALUES ($1, $2)`,
			briefID, hook,
		); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx,
		"UPDATE briefs SET updated_at = NOW() WHERE id = $1", briefID,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
