package signal

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// recommendedHookForAngle returns a suggested hook line for a given angle.
func recommendedHookForAngle(angle string) string {
	switch strings.ToLower(strings.TrimSpace(angle)) {
	case "problem_solution":
		return "Call out the pain, then show the fastest path to outcome."
	case "social_proof":
		return "Lead with proof volume and a concrete before/after result."
	case "transformation":
		return "Contrast old state vs new state in the first 3 seconds."
	case "urgency":
		return "Use a time-bound offer and direct CTA in opening line."
	case "education":
		return "Teach one surprising mechanism and tie it to conversion."
	default:
		return "Open with a concrete benefit and immediate CTA."
	}
}

// computeRecommendationConfidence calculates a confidence score [0, 99].
func computeRecommendationConfidence(impressions int64, ctr float64) float64 {
	if impressions <= 0 {
		return 0
	}
	base := math.Min(80, (float64(impressions)/1000.0)*12.0)
	ctrBoost := math.Min(19, ctr*250.0)
	score := base + ctrBoost
	if score > 99 {
		score = 99
	}
	return math.Round(score*10) / 10
}

// shouldRefreshRecommendations returns true if recommendations are stale or refresh is forced.
func shouldRefreshRecommendations(c echo.Context, workspaceID pgtype.UUID, refresh bool) (bool, error) {
	if refresh {
		return true, nil
	}

	var lastGenerated pgtype.Timestamptz
	err := store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT MAX(last_generated_at)
		 FROM signal_brief_recommendations
		 WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&lastGenerated)
	if err != nil {
		return false, err
	}
	if !lastGenerated.Valid {
		return true, nil
	}
	return time.Since(lastGenerated.Time.UTC()) >= 7*24*time.Hour, nil
}

// regenerateBriefRecommendations recomputes angle-level brief recommendations from signal data.
func regenerateBriefRecommendations(c echo.Context, workspaceID pgtype.UUID, days int) error {
	if store.Pool() == nil {
		return errors.New("database_not_initialized")
	}

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT
			COALESCE(v.angle, 'unknown') AS angle,
			COALESCE(SUM(m.impressions), 0) AS impressions,
			COALESCE(SUM(m.clicks), 0) AS clicks,
			CASE WHEN COALESCE(SUM(m.impressions), 0) > 0
				THEN COALESCE(SUM(m.clicks), 0)::float8 / COALESCE(SUM(m.impressions), 0)::float8
				ELSE 0
			END AS ctr
		 FROM signal_metrics_daily m
		 LEFT JOIN variants v ON v.id = m.variant_id
		 WHERE m.workspace_id = $1
		   AND m.metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY COALESCE(v.angle, 'unknown')
		 HAVING COALESCE(SUM(m.impressions), 0) >= 1000
		 ORDER BY ctr DESC, impressions DESC
		 LIMIT 3`,
		workspaceID,
		days,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]struct {
		angle       string
		impressions int64
		ctr         float64
	}, 0)

	for rows.Next() {
		var angle string
		var impressions int64
		var clicks int64
		var ctr float64
		if scanErr := rows.Scan(&angle, &impressions, &clicks, &ctr); scanErr != nil {
			return scanErr
		}
		items = append(items, struct {
			angle       string
			impressions int64
			ctr         float64
		}{
			angle:       angle,
			impressions: impressions,
			ctr:         ctr,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return rowsErr
	}

	for _, item := range items {
		confidence := computeRecommendationConfidence(item.impressions, item.ctr)
		rationale := "Based on your recent campaigns, " + item.angle + " is outperforming with CTR " +
			fmt.Sprintf("%.2f%%", item.ctr*100) + " across " + strconv.FormatInt(item.impressions, 10) +
			" impressions in the last " + strconv.Itoa(days) + " days."

		_, execErr := store.Pool().Exec(
			c.Request().Context(),
			`INSERT INTO signal_brief_recommendations (
				workspace_id, angle, suggested_hook, rationale, confidence_score, impression_volume, window_days, first_generated_at, last_generated_at
			 ) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			 ON CONFLICT (workspace_id, angle)
			 DO UPDATE SET
			   suggested_hook = EXCLUDED.suggested_hook,
			   rationale = EXCLUDED.rationale,
			   confidence_score = EXCLUDED.confidence_score,
			   impression_volume = EXCLUDED.impression_volume,
			   window_days = EXCLUDED.window_days,
			   last_generated_at = NOW(),
			   updated_at = NOW()`,
			workspaceID,
			item.angle,
			recommendedHookForAngle(item.angle),
			rationale,
			confidence,
			item.impressions,
			days,
		)
		if execErr != nil {
			return execErr
		}
	}

	return nil
}

// GetSignalRecommendations handles GET /api/v1/signal/recommendations
func GetSignalRecommendations(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	workspaceID, _, _, _, err := store.GetWorkspaceForOrg(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	refresh := parseBool(c.QueryParam("refresh"), false)

	refreshNeeded, err := shouldRefreshRecommendations(c, workspaceID, refresh)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}
	if refreshNeeded {
		if err := regenerateBriefRecommendations(c, workspaceID, days); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
		}
	}

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT angle, suggested_hook, rationale, confidence_score, impression_volume, window_days, last_generated_at
		 FROM signal_brief_recommendations
		 WHERE workspace_id = $1
		 ORDER BY confidence_score DESC, impression_volume DESC
		 LIMIT 3`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}
	defer rows.Close()

	recommendations := make([]map[string]any, 0)
	for rows.Next() {
		var angle string
		var suggestedHook string
		var rationale string
		var confidence float64
		var impressions int64
		var windowDays int
		var lastGenerated time.Time
		if scanErr := rows.Scan(&angle, &suggestedHook, &rationale, &confidence, &impressions, &windowDays, &lastGenerated); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
		}

		recommendations = append(recommendations, map[string]any{
			"angle":             angle,
			"suggested_hook":    suggestedHook,
			"rationale":         rationale,
			"confidence_score":  confidence,
			"impression_volume": impressions,
			"window_days":       windowDays,
			"last_generated_at": lastGenerated.UTC(),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":          claims.OrgID,
		"workspace_id":    util.UUIDString(workspaceID),
		"refreshed":       refreshNeeded,
		"recommendations": recommendations,
	})
}

// RefreshAllSignalRecommendations handles POST /api/v1/internal/signal/recommendations/refresh-all
func RefreshAllSignalRecommendations(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
	if role != "worker" && role != "admin" && role != "system" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	if err := ensureSignalTables(c); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_tables_init_failed"})
	}

	days := parseDays(c.QueryParam("days"))
	limit := util.ParsePositiveInt(c.QueryParam("limit"), 1000)
	if limit > 5000 {
		limit = 5000
	}

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT id, org_id
		 FROM workspaces
		 ORDER BY created_at ASC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
	}
	defer rows.Close()

	refreshed := 0
	failures := make([]map[string]any, 0)
	for rows.Next() {
		var workspaceID pgtype.UUID
		var orgID string
		if scanErr := rows.Scan(&workspaceID, &orgID); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
		}

		if regenErr := regenerateBriefRecommendations(c, workspaceID, days); regenErr != nil {
			failures = append(failures, map[string]any{
				"org_id": orgID,
				"error":  regenErr.Error(),
			})
			continue
		}

		refreshed++
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_recommendations_refresh_all_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"refreshed": refreshed,
		"failures":  failures,
		"days":      days,
	})
}

// ---------------------------------------------------------------------------
// Recommendation feedback
// ---------------------------------------------------------------------------

// CreateSignalRecommendationFeedback records whether a user accepted or ignored a recommendation.
// POST /api/v1/signal/recommendations/feedback
func CreateSignalRecommendationFeedback(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	workspaceID, wsErr := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if wsErr != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "workspace_not_found"})
	}

	var req struct {
		Angle  string `json:"angle"`
		Action string `json:"action"`
		Source string `json:"source"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	req.Angle = strings.TrimSpace(req.Angle)
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.Source = strings.TrimSpace(req.Source)
	if req.Angle == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "angle_required"})
	}
	if req.Action != "accept" && req.Action != "ignore" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "action_must_be_accept_or_ignore"})
	}
	if req.Source == "" {
		req.Source = "brief_create_panel"
	}

	_, err := store.Pool().Exec(c.Request().Context(),
		`INSERT INTO signal_recommendation_feedback (workspace_id, angle, action, source, created_by)
		 VALUES ($1, $2, $3, $4, $5)`,
		workspaceID, req.Angle, req.Action, req.Source, claims.UserID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "feedback_insert_failed"})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"angle":  req.Angle,
		"action": req.Action,
		"source": req.Source,
	})
}

// ListSignalRecommendationFeedbackByAngle returns feedback entries for the workspace, optionally filtered by angle.
// GET /api/v1/signal/recommendations/feedback?angle=&limit=
func ListSignalRecommendationFeedbackByAngle(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	workspaceID, wsErr := store.WorkspaceIDForOrg(c.Request().Context(), claims.OrgID)
	if wsErr != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "workspace_not_found"})
	}

	angle := strings.TrimSpace(c.QueryParam("angle"))
	limit := 50
	if l := strings.TrimSpace(c.QueryParam("limit")); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	var rows interface {
		Close()
		Next() bool
		Scan(...any) error
		Err() error
	}
	var queryErr error

	if angle != "" {
		rows, queryErr = store.Pool().Query(c.Request().Context(),
			`SELECT id, angle, action, source, created_by, created_at
			 FROM signal_recommendation_feedback
			 WHERE workspace_id = $1 AND angle = $2
			 ORDER BY created_at DESC LIMIT $3`,
			workspaceID, angle, limit,
		)
	} else {
		rows, queryErr = store.Pool().Query(c.Request().Context(),
			`SELECT id, angle, action, source, created_by, created_at
			 FROM signal_recommendation_feedback
			 WHERE workspace_id = $1
			 ORDER BY created_at DESC LIMIT $2`,
			workspaceID, limit,
		)
	}
	if queryErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "feedback_query_failed"})
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id        int64
			ang, act  string
			src       string
			createdBy *string
			createdAt time.Time
		)
		if scanErr := rows.Scan(&id, &ang, &act, &src, &createdBy, &createdAt); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "feedback_scan_failed"})
		}
		item := map[string]any{
			"id":         id,
			"angle":      ang,
			"action":     act,
			"source":     src,
			"created_by": createdBy,
			"created_at": createdAt,
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "feedback_query_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
		"angle": angle,
		"limit": limit,
	})
}
