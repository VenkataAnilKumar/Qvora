package signal

import (
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// ListSignalConnections handles GET /api/v1/signal/connections
func ListSignalConnections(c echo.Context) error {
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

	_, err = store.Pool().Exec(
		c.Request().Context(),
		`UPDATE signal_connections
		 SET status = 'token_expired',
		     error_reason = COALESCE(error_reason, 'oauth_token_expired'),
		     updated_at = NOW()
		 WHERE workspace_id = $1
		   AND status = 'connected'
		   AND token_expires_at IS NOT NULL
		   AND token_expires_at < NOW()`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT platform, status, account_id, account_name, error_reason, token_expires_at, last_synced_at, created_at, updated_at
		 FROM signal_connections
		 WHERE workspace_id = $1
		 ORDER BY updated_at DESC, created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}
	defer rows.Close()

	connections := make([]map[string]any, 0)
	for rows.Next() {
		var platform string
		var status string
		var accountID string
		var accountName *string
		var errorReason *string
		var tokenExpiresAt pgtype.Timestamptz
		var lastSyncedAt pgtype.Timestamptz
		var createdAt pgtype.Timestamptz
		var updatedAt pgtype.Timestamptz
		if scanErr := rows.Scan(&platform, &status, &accountID, &accountName, &errorReason, &tokenExpiresAt, &lastSyncedAt, &createdAt, &updatedAt); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_scan_failed"})
		}

		connections = append(connections, map[string]any{
			"platform":         platform,
			"status":           status,
			"account_id":       accountID,
			"account_name":     accountName,
			"error_reason":     errorReason,
			"token_expires_at": util.TsTime(tokenExpiresAt),
			"last_synced_at":   util.TsTime(lastSyncedAt),
			"created_at":       util.TsTime(createdAt),
			"updated_at":       util.TsTime(updatedAt),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connections_list_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"connections":  connections,
	})
}

// UpsertSignalConnection handles PUT /api/v1/signal/connections/:platform
func UpsertSignalConnection(c echo.Context) error {
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

	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}

	var req struct {
		AccountID      string  `json:"account_id"`
		AccountName    *string `json:"account_name"`
		Status         string  `json:"status"`
		ErrorReason    *string `json:"error_reason"`
		TokenExpiresAt *string `json:"token_expires_at"`
		LastSyncedAt   *string `json:"last_synced_at"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "account_id_required"})
	}

	status := strings.TrimSpace(strings.ToLower(req.Status))
	if status == "" {
		status = "connected"
	}
	if status != "connected" && status != "disconnected" && status != "token_expired" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}

	var tokenExpiresAt any
	if req.TokenExpiresAt != nil && strings.TrimSpace(*req.TokenExpiresAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.TokenExpiresAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_token_expires_at"})
		}
		tokenExpiresAt = t.UTC()
	}

	var lastSyncedAt any
	if req.LastSyncedAt != nil && strings.TrimSpace(*req.LastSyncedAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.LastSyncedAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_last_synced_at"})
		}
		lastSyncedAt = t.UTC()
	}

	if status == "connected" {
		req.ErrorReason = nil
	}

	_, err = store.Pool().Exec(
		c.Request().Context(),
		`INSERT INTO signal_connections (workspace_id, platform, status, account_id, account_name, error_reason, token_expires_at, last_synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW()))
		 ON CONFLICT (workspace_id, platform, account_id)
		 DO UPDATE SET
		   status = EXCLUDED.status,
		   account_name = EXCLUDED.account_name,
		   error_reason = EXCLUDED.error_reason,
		   token_expires_at = EXCLUDED.token_expires_at,
		   last_synced_at = COALESCE(EXCLUDED.last_synced_at, signal_connections.last_synced_at),
		   updated_at = NOW()`,
		workspaceID,
		platform,
		status,
		accountID,
		req.AccountName,
		req.ErrorReason,
		tokenExpiresAt,
		lastSyncedAt,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connection_upsert_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"platform":     platform,
		"account_id":   accountID,
		"status":       status,
		"updated":      true,
	})
}

// PatchSignalConnectionHealth handles PATCH /api/v1/signal/connections/:platform/:accountId/health
func PatchSignalConnectionHealth(c echo.Context) error {
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

	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	accountID := strings.TrimSpace(c.Param("accountId"))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}
	if accountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "account_id_required"})
	}

	var req struct {
		Status         *string `json:"status"`
		ErrorReason    *string `json:"error_reason"`
		TokenExpiresAt *string `json:"token_expires_at"`
		LastSyncedAt   *string `json:"last_synced_at"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}

	status := ""
	if req.Status != nil {
		status = strings.TrimSpace(strings.ToLower(*req.Status))
		if status != "connected" && status != "disconnected" && status != "token_expired" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
		}
	}

	var tokenExpiresAt any
	if req.TokenExpiresAt != nil && strings.TrimSpace(*req.TokenExpiresAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.TokenExpiresAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_token_expires_at"})
		}
		tokenExpiresAt = t.UTC()
	}

	var lastSyncedAt any
	if req.LastSyncedAt != nil && strings.TrimSpace(*req.LastSyncedAt) != "" {
		t, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(*req.LastSyncedAt))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_last_synced_at"})
		}
		lastSyncedAt = t.UTC()
	}

	result, err := store.Pool().Exec(
		c.Request().Context(),
		`UPDATE signal_connections
		 SET status = COALESCE(NULLIF($4, ''), status),
		     error_reason = CASE
		       WHEN COALESCE(NULLIF($4, ''), status) = 'connected' THEN NULL
		       ELSE COALESCE($5, error_reason)
		     END,
		     token_expires_at = COALESCE($6, token_expires_at),
		     last_synced_at = COALESCE($7, last_synced_at, NOW()),
		     updated_at = NOW()
		 WHERE workspace_id = $1
		   AND platform = $2
		   AND account_id = $3`,
		workspaceID,
		platform,
		accountID,
		status,
		req.ErrorReason,
		tokenExpiresAt,
		lastSyncedAt,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_connection_health_patch_failed"})
	}
	if result.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "signal_connection_not_found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"platform":     platform,
		"account_id":   accountID,
		"updated":      true,
	})
}
