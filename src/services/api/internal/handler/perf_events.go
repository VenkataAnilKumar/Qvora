package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
)

// HandleCreatePerfEvent receives performance events from Go workers and
// inserts them into video_performance_events. Workers fire-and-forget —
// this endpoint always returns 202 to avoid slowing down task handlers.
func HandleCreatePerfEvent(c echo.Context) error {
	var req struct {
		WorkspaceID  string `json:"workspace_id"`
		VariantID    string `json:"variant_id"`
		JobID        string `json:"job_id"`
		Stage        string `json:"stage"`
		DurationMS   int    `json:"duration_ms"`
		Model        string `json:"model"`
		FalRequestID string `json:"fal_request_id"`
		ErrorType    string `json:"error_type"`
		ErrorMsg     string `json:"error_msg"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	if req.WorkspaceID == "" || req.Stage == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "workspace_id and stage required"})
	}

	workspaceID, err := parseUUID(req.WorkspaceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_workspace_id"})
	}
	variantID, err := parseUUID(req.VariantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}
	jobID, err := parseUUID(req.JobID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_job_id"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := q.InsertPerfEvent(c.Request().Context(), db.InsertPerfEventParams{
		WorkspaceID:  workspaceID,
		VariantID:    variantID,
		JobID:        jobID,
		Stage:        req.Stage,
		DurationMs:   int32(req.DurationMS),
		Model:        stringPtr(req.Model),
		FalRequestID: stringPtr(req.FalRequestID),
		ErrorType:    stringPtr(req.ErrorType),
		ErrorMsg:     stringPtr(req.ErrorMsg),
		RecordedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert_failed"})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"status": "accepted"})
}

// HandleCreateCostEvent receives cost events from Go workers and
// inserts them into cost_events for billing attribution.
func HandleCreateCostEvent(c echo.Context) error {
	var req struct {
		WorkspaceID  string  `json:"workspace_id"`
		VariantID    string  `json:"variant_id"`
		JobID        string  `json:"job_id"`
		Source       string  `json:"source"`
		Model        string  `json:"model"`
		EstimatedUSD float64 `json:"estimated_usd"`
		Credits      int     `json:"credits"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}
	if req.WorkspaceID == "" || req.Source == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "workspace_id and source required"})
	}
	if req.Credits == 0 {
		req.Credits = 1
	}

	workspaceID, err := parseUUID(req.WorkspaceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_workspace_id"})
	}
	variantID, err := parseUUID(req.VariantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}
	jobID, err := parseUUID(req.JobID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_job_id"})
	}

	var estimated pgtype.Numeric
	if err := estimated.Scan(req.EstimatedUSD); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_estimated_usd"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := q.InsertCostEvent(c.Request().Context(), db.InsertCostEventParams{
		WorkspaceID:  workspaceID,
		VariantID:    variantID,
		JobID:        jobID,
		Source:       req.Source,
		Model:        stringPtr(req.Model),
		EstimatedUsd: estimated,
		Credits:      int32(req.Credits),
		RecordedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert_failed"})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"status": "accepted"})
}

// HandlePatchAvatarJob stores the avatar provider job ID on the variant.
// Called by the avatar task handler after CreateLipSync succeeds.
func HandlePatchAvatarJob(c echo.Context) error {
	if c.Param("id") == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "variant_id_required"})
	}

	var req struct {
		AvatarJobID    string `json:"avatar_job_id"`
		AvatarProvider string `json:"avatar_provider"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	variantID, err := parseUUID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}

	q, err := queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	if err := q.UpdateVariantAvatarJob(c.Request().Context(), db.UpdateVariantAvatarJobParams{
		ID:             variantID,
		AvatarJobID:    stringPtr(req.AvatarJobID),
		AvatarProvider: stringPtr(req.AvatarProvider),
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update_failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
