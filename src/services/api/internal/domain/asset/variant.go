package asset

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/qvora/api/internal/db"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
)

func isVariantPlaybackReady(status string, muxAssetID *string, muxPlaybackID *string) bool {
	if !strings.EqualFold(strings.TrimSpace(status), "complete") {
		return false
	}
	if muxAssetID == nil || strings.TrimSpace(*muxAssetID) == "" {
		return false
	}
	if muxPlaybackID == nil || strings.TrimSpace(*muxPlaybackID) == "" {
		return false
	}
	return true
}

// GetVariantPlaybackURL handles GET /api/v1/variants/:id/playback-url
func GetVariantPlaybackURL(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	workspace, err := q.GetWorkspaceByOrgID(c.Request().Context(), claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "workspace_not_found"})
	}
	variantID := strings.TrimSpace(c.Param("id"))
	if variantID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "variant_id_required"})
	}
	var variantUUID pgtype.UUID
	if err := variantUUID.Scan(variantID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}
	variant, err := q.GetVariantForPlayback(c.Request().Context(), db.GetVariantForPlaybackParams{
		ID:          variantUUID,
		WorkspaceID: workspace.ID,
	})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "variant_not_found"})
	}
	if !isVariantPlaybackReady(variant.Status, variant.MuxAssetID, variant.MuxPlaybackID) {
		return c.JSON(http.StatusConflict, map[string]string{"error": "playback_not_ready"})
	}
	token, expiresAt, err := GenerateMuxPlaybackToken(claims.OrgID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token_generation_failed"})
	}
	playbackID := strings.TrimSpace(*variant.MuxPlaybackID)
	playbackURL := fmt.Sprintf("https://stream.mux.com/%s.m3u8?token=%s", playbackID, token)
	return c.JSON(http.StatusOK, map[string]any{
		"variant_id":    variantID,
		"playback_id":   playbackID,
		"playback_url":  playbackURL,
		"token":         token,
		"token_expires": expiresAt.Format(time.RFC3339),
	})
}

// UpdateVariantFalRequest handles PATCH /api/v1/variants/:id/fal-request
// Internal: called by the worker after fal.queue.submit() to store the request ID.
func UpdateVariantFalRequest(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	variantID := strings.TrimSpace(c.Param("id"))
	if variantID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "variant_id_required"})
	}
	var body struct {
		FalRequestID string `json:"fal_request_id"`
	}
	if err := c.Bind(&body); err != nil || strings.TrimSpace(body.FalRequestID) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "fal_request_id_required"})
	}
	q, err := store.Queries(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	var variantUUID pgtype.UUID
	if err := variantUUID.Scan(variantID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
	}
	falReqID := strings.TrimSpace(body.FalRequestID)
	updated, err := q.UpdateVariantFalRequestID(c.Request().Context(), db.UpdateVariantFalRequestIDParams{
		ID:           variantUUID,
		FalRequestID: &falReqID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"variant_id":     variantID,
		"fal_request_id": updated.FalRequestID,
	})
}

// GenerateMuxPlaybackToken signs a JWT for signed Mux playback.
// Exported for use by domain/media webhook.go.
func GenerateMuxPlaybackToken(workspaceID string) (string, time.Time, error) {
	secret := strings.TrimSpace(os.Getenv("MUX_SECRET_TOKEN"))
	if secret == "" {
		return "", time.Time{}, fmt.Errorf("MUX_SECRET_TOKEN is not set")
	}
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	payload := map[string]any{
		"sub": workspaceID,
		"aud": "v",
		"exp": expiresAt.Unix(),
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}
	enc := base64.RawURLEncoding
	unsigned := enc.EncodeToString(headerJSON) + "." + enc.EncodeToString(payloadJSON)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc.EncodeToString(mac.Sum(nil)), expiresAt, nil
}
