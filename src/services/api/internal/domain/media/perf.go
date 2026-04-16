package media

import (
"fmt"
"net/http"
"strings"
"time"

"github.com/jackc/pgx/v5/pgtype"
"github.com/labstack/echo/v4"
"github.com/qvora/api/internal/db"
appmiddleware "github.com/qvora/api/internal/middleware"
"github.com/qvora/api/internal/store"
"github.com/qvora/api/internal/util"
)

// HandleCreatePerfEvent handles POST /api/v1/internal/perf-events
func HandleCreatePerfEvent(c echo.Context) error {
claims := appmiddleware.GetClaims(c)
if claims == nil || strings.TrimSpace(claims.UserID) == "" {
return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}
role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
if role != "worker" && role != "system" {
return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
}
var req struct {
WorkspaceID  string `json:"workspace_id"`
VariantID    string `json:"variant_id"`
JobID        string `json:"job_id"`
Stage        string `json:"stage"`
DurationMS   int32  `json:"duration_ms"`
Model        string `json:"model"`
FalRequestID string `json:"fal_request_id"`
ErrorType    string `json:"error_type"`
ErrorMsg     string `json:"error_msg"`
}
if err := c.Bind(&req); err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}
req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
req.VariantID = strings.TrimSpace(req.VariantID)
req.JobID = strings.TrimSpace(req.JobID)
req.Stage = strings.TrimSpace(req.Stage)
if req.WorkspaceID == "" || req.VariantID == "" || req.Stage == "" {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "workspace_id_variant_id_stage_required"})
}
q, err := store.Queries(c.Request().Context())
if err != nil {
return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
}
workspaceID, err := util.ParseUUID(req.WorkspaceID)
if err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_workspace_id"})
}
variantID, err := util.ParseUUID(req.VariantID)
if err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
}
var jobID pgtype.UUID
if req.JobID != "" {
if parsed, parseErr := util.ParseUUID(req.JobID); parseErr == nil {
jobID = parsed
}
}
insertErr := q.InsertPerfEvent(c.Request().Context(), db.InsertPerfEventParams{
WorkspaceID:  workspaceID,
VariantID:    variantID,
JobID:        jobID,
Stage:        req.Stage,
DurationMs:   req.DurationMS,
Model:        util.StringPtr(req.Model),
FalRequestID: util.StringPtr(req.FalRequestID),
ErrorType:    util.StringPtr(req.ErrorType),
ErrorMsg:     util.StringPtr(req.ErrorMsg),
RecordedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
})
if insertErr != nil {
return c.JSON(http.StatusInternalServerError, map[string]string{"error": "perf_event_insert_failed"})
}
return c.JSON(http.StatusCreated, map[string]any{
"variant_id": req.VariantID,
"stage":      req.Stage,
"recorded":   true,
})
}

// HandleCreateCostEvent handles POST /api/v1/internal/cost-events
func HandleCreateCostEvent(c echo.Context) error {
claims := appmiddleware.GetClaims(c)
if claims == nil || strings.TrimSpace(claims.UserID) == "" {
return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}
role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
if role != "worker" && role != "system" {
return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
}
var req struct {
WorkspaceID  string  `json:"workspace_id"`
VariantID    string  `json:"variant_id"`
JobID        string  `json:"job_id"`
Source       string  `json:"source"`
Model        string  `json:"model"`
EstimatedUSD float64 `json:"estimated_usd"`
Credits      int32   `json:"credits"`
}
if err := c.Bind(&req); err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}
req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
req.VariantID = strings.TrimSpace(req.VariantID)
req.JobID = strings.TrimSpace(req.JobID)
req.Source = strings.TrimSpace(req.Source)
if req.WorkspaceID == "" || req.VariantID == "" || req.Source == "" {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "workspace_id_variant_id_source_required"})
}
q, err := store.Queries(c.Request().Context())
if err != nil {
return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
}
workspaceID, err := util.ParseUUID(req.WorkspaceID)
if err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_workspace_id"})
}
variantID, err := util.ParseUUID(req.VariantID)
if err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
}
var jobID pgtype.UUID
if req.JobID != "" {
if parsed, parseErr := util.ParseUUID(req.JobID); parseErr == nil {
jobID = parsed
}
}
var estimatedUSD pgtype.Numeric
if scanErr := estimatedUSD.Scan(fmt.Sprintf("%.10f", req.EstimatedUSD)); scanErr != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_estimated_usd"})
}
insertErr := q.InsertCostEvent(c.Request().Context(), db.InsertCostEventParams{
WorkspaceID:  workspaceID,
VariantID:    variantID,
JobID:        jobID,
Source:       req.Source,
Model:        util.StringPtr(req.Model),
EstimatedUsd: estimatedUSD,
Credits:      req.Credits,
RecordedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
})
if insertErr != nil {
return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cost_event_insert_failed"})
}
return c.JSON(http.StatusCreated, map[string]any{
"variant_id":    req.VariantID,
"source":        req.Source,
"estimated_usd": req.EstimatedUSD,
"recorded":      true,
})
}

// HandlePatchAvatarJob handles PATCH /api/v1/internal/variants/:id/avatar-job
func HandlePatchAvatarJob(c echo.Context) error {
claims := appmiddleware.GetClaims(c)
if claims == nil || strings.TrimSpace(claims.UserID) == "" {
return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}
role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
if role != "worker" && role != "system" {
return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
}
variantID := strings.TrimSpace(c.Param("id"))
var req struct {
AvatarJobID    string `json:"avatar_job_id"`
AvatarProvider string `json:"avatar_provider"`
}
if err := c.Bind(&req); err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
}
req.AvatarJobID = strings.TrimSpace(req.AvatarJobID)
req.AvatarProvider = strings.TrimSpace(req.AvatarProvider)
if req.AvatarJobID == "" {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "avatar_job_id_required"})
}
q, err := store.Queries(c.Request().Context())
if err != nil {
return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
}
parsedVariantID, err := util.ParseUUID(variantID)
if err != nil {
return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
}
updateErr := q.UpdateVariantAvatarJob(c.Request().Context(), db.UpdateVariantAvatarJobParams{
ID:             parsedVariantID,
AvatarJobID:    util.StringPtr(req.AvatarJobID),
AvatarProvider: util.StringPtr(req.AvatarProvider),
})
if updateErr != nil {
return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_update_failed"})
}
return c.JSON(http.StatusOK, map[string]any{
"variant_id":      variantID,
"avatar_job_id":   req.AvatarJobID,
"avatar_provider": req.AvatarProvider,
"updated":         true,
})
}
