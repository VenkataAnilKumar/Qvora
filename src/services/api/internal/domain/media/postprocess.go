package media

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

type postprocessCallbackRequest struct {
	RequestID     string  `json:"request_id"`
	JobID         string  `json:"job_id"`
	VariantID     string  `json:"variant_id"`
	WorkspaceID   string  `json:"workspace_id"`
	Status        string  `json:"status"` // success | failed
	OutputR2Key   string  `json:"output_r2_key"`
	MuxAssetID    *string `json:"mux_asset_id"`
	MuxPlayableID *string `json:"mux_playable_id"`
	DurationMS    *int64  `json:"duration_ms"`
	ErrorMessage  *string `json:"error_message"`
}

// HandlePostprocessCallback handles POST /api/v1/internal/postprocess/callback
// Accepts success/failure callbacks from the Rust postprocessor.
func HandlePostprocessCallback(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
	if role != "worker" && role != "system" && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	var req postprocessCallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	req.RequestID = strings.TrimSpace(req.RequestID)
	req.JobID = strings.TrimSpace(req.JobID)
	req.VariantID = strings.TrimSpace(req.VariantID)
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	req.Status = strings.TrimSpace(strings.ToLower(req.Status))
	req.OutputR2Key = strings.TrimSpace(req.OutputR2Key)
	if req.RequestID == "" || req.VariantID == "" || req.WorkspaceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "request_id_variant_id_workspace_id_required"})
	}
	if req.Status != "success" && req.Status != "failed" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}
	parsedWorkspaceID, err := util.ParseUUID(req.WorkspaceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_workspace_id"})
	}
	parsedVariantID, err := util.ParseUUID(req.VariantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}
	if claims.OrgID != req.WorkspaceID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "workspace_context_mismatch"})
	}
	payload, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "payload_marshal_failed"})
	}
	insertResult, err := store.Pool().Exec(
		c.Request().Context(),
		`INSERT INTO postprocess_callbacks (workspace_id, variant_id, request_id, job_id, status, payload)
		 VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6)
		 ON CONFLICT (workspace_id, variant_id, request_id) DO NOTHING`,
		parsedWorkspaceID, parsedVariantID, req.RequestID, req.JobID, req.Status, payload,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "postprocess_callback_insert_failed"})
	}
	// Idempotent: already processed
	if insertResult.RowsAffected() == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"accepted":   true,
			"idempotent": true,
			"request_id": req.RequestID,
			"variant_id": req.VariantID,
		})
	}
	if req.Status == "success" {
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`UPDATE variants
			 SET status = CASE WHEN status = 'complete' THEN status ELSE 'postprocessing' END,
			     r2_key = CASE WHEN NULLIF($3, '') IS NULL THEN r2_key ELSE $3 END,
			     mux_asset_id = COALESCE($4, mux_asset_id),
			     mux_playback_id = COALESCE($5, mux_playback_id),
			     duration_secs = COALESCE($6::numeric / 1000.0, duration_secs),
			     updated_at = NOW()
			 WHERE id = $1 AND workspace_id = $2`,
			parsedVariantID, parsedWorkspaceID,
			req.OutputR2Key, req.MuxAssetID, req.MuxPlayableID, req.DurationMS,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_status_update_failed"})
		}
		if strings.TrimSpace(req.JobID) != "" {
			if parsedJobID, parseErr := util.ParseUUID(req.JobID); parseErr == nil {
				_, _ = store.Pool().Exec(
					c.Request().Context(),
					`UPDATE jobs
					 SET status = CASE WHEN status = 'complete' THEN status ELSE 'postprocessing' END,
					     error_msg = NULL,
					     updated_at = NOW()
					 WHERE id = $1`,
					parsedJobID,
				)
			}
		}
	} else {
		_, err = store.Pool().Exec(
			c.Request().Context(),
			`UPDATE variants
			 SET status = CASE WHEN status = 'complete' THEN status ELSE 'failed' END,
			     updated_at = NOW()
			 WHERE id = $1 AND workspace_id = $2`,
			parsedVariantID, parsedWorkspaceID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_status_update_failed"})
		}
		if strings.TrimSpace(req.JobID) != "" {
			if parsedJobID, parseErr := util.ParseUUID(req.JobID); parseErr == nil {
				_, _ = store.Pool().Exec(
					c.Request().Context(),
					`UPDATE jobs
					 SET status = CASE WHEN status = 'complete' THEN status ELSE 'failed' END,
					     error_msg = $2,
					     updated_at = NOW()
					 WHERE id = $1`,
					parsedJobID, req.ErrorMessage,
				)
			}
		}
	}
	q, qErr := store.Queries(c.Request().Context())
	if qErr != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	variant, getErr := q.GetVariantForPlayback(c.Request().Context(), db.GetVariantForPlaybackParams{
		ID:          parsedVariantID,
		WorkspaceID: parsedWorkspaceID,
	})
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "variant_not_found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_lookup_failed"})
	}
	response := map[string]any{
		"accepted":       true,
		"idempotent":     false,
		"request_id":     req.RequestID,
		"variant_id":     req.VariantID,
		"workspace_id":   req.WorkspaceID,
		"variant_status": variant.Status,
	}
	if req.JobID != "" {
		response["job_id"] = req.JobID
	}
	if req.MuxAssetID != nil {
		response["mux_asset_id"] = strings.TrimSpace(*req.MuxAssetID)
	}
	if req.MuxPlayableID != nil {
		response["mux_playable_id"] = strings.TrimSpace(*req.MuxPlayableID)
	}
	if req.ErrorMessage != nil {
		response["error_message"] = strings.TrimSpace(*req.ErrorMessage)
	}
	return c.JSON(http.StatusOK, response)
}
