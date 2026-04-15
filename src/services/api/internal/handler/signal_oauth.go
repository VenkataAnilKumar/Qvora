package handler

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
)

// ---------------------------------------------------------------------------
// AES-GCM token encryption / decryption
// Env: SIGNAL_TOKEN_ENCRYPTION_KEY — 32-byte hex (64 hex chars)
// ---------------------------------------------------------------------------

func tokenEncryptionKey() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("SIGNAL_TOKEN_ENCRYPTION_KEY"))
	if len(raw) != 64 {
		return nil, fmt.Errorf("SIGNAL_TOKEN_ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("SIGNAL_TOKEN_ENCRYPTION_KEY is not valid hex: %w", err)
	}
	return key, nil
}

func encryptToken(plaintext string) (string, error) {
	key, err := tokenEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptToken(cipherB64 string) (string, error) {
	key, err := tokenEncryptionKey()
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}
	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("gcm open: %w", err)
	}
	return string(plaintext), nil
}

// ---------------------------------------------------------------------------
// OAuth state — stateless HMAC-signed payload
// Format: base64url(json{org_id,platform,ts}) + "~" + hex(hmac-sha256)
// Env: SIGNAL_OAUTH_STATE_SECRET (falls back to INTERNAL_API_KEY)
// ---------------------------------------------------------------------------

type oauthStatePayload struct {
	OrgID     string `json:"org_id"`
	Platform  string `json:"platform"`
	Timestamp int64  `json:"ts"`
}

func oauthStateSecret() string {
	if s := strings.TrimSpace(os.Getenv("SIGNAL_OAUTH_STATE_SECRET")); s != "" {
		return s
	}
	return strings.TrimSpace(os.Getenv("INTERNAL_API_KEY"))
}

func buildOAuthState(orgID, platform string) (string, error) {
	payload := oauthStatePayload{
		OrgID:     orgID,
		Platform:  platform,
		Timestamp: time.Now().Unix(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("state marshal: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(raw)
	secret := oauthStateSecret()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	sig := hex.EncodeToString(mac.Sum(nil))
	return encoded + "~" + sig, nil
}

func verifyOAuthState(state string) (orgID, platform string, err error) {
	parts := strings.SplitN(state, "~", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid state format")
	}
	encoded, sig := parts[0], parts[1]
	secret := oauthStateSecret()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", "", fmt.Errorf("state signature mismatch")
	}
	raw, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("state decode: %w", err)
	}
	var p oauthStatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", "", fmt.Errorf("state unmarshal: %w", err)
	}
	if time.Now().Unix()-p.Timestamp > 900 {
		return "", "", fmt.Errorf("state expired")
	}
	return p.OrgID, p.Platform, nil
}

// ---------------------------------------------------------------------------
// Schema migration: add encrypted token columns to signal_connections
// ---------------------------------------------------------------------------

func ensureSignalOAuthColumns(ctx context.Context) error {
	if dbPool == nil {
		return fmt.Errorf("database_not_initialized")
	}
	for _, stmt := range []string{
		`ALTER TABLE signal_connections ADD COLUMN IF NOT EXISTS encrypted_access_token TEXT`,
		`ALTER TABLE signal_connections ADD COLUMN IF NOT EXISTS encrypted_refresh_token TEXT`,
		`ALTER TABLE signal_connections ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMPTZ`,
		`ALTER TABLE variants ADD COLUMN IF NOT EXISTS platform_asset_id TEXT`,
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS signal_purge_at TIMESTAMPTZ`,
		`ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS notification_email TEXT`,
	} {
		if _, err := dbPool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("schema migration %q: %w", stmt[:60], err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// InitiateSignalOAuth
// GET /api/v1/signal/oauth/:platform/initiate
// Returns { platform, url, state, redirect_uri }
// ---------------------------------------------------------------------------

// InitiateSignalOAuth godoc
func InitiateSignalOAuth(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.OrgID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if claims.OrgRole != "org:admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "requires_admin_role"})
	}

	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}

	state, err := buildOAuthState(claims.OrgID, platform)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "state_generation_failed"})
	}

	appURL := strings.TrimRight(strings.TrimSpace(os.Getenv("NEXT_PUBLIC_APP_URL")), "/")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	redirectURI := appURL + "/settings/integrations/callback/" + platform

	var authURL string
	switch platform {
	case "meta":
		appID := strings.TrimSpace(os.Getenv("META_APP_ID"))
		if appID == "" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "meta_not_configured"})
		}
		params := url.Values{}
		params.Set("client_id", appID)
		params.Set("redirect_uri", redirectURI)
		params.Set("state", state)
		params.Set("scope", "ads_read,read_insights")
		params.Set("response_type", "code")
		authURL = "https://www.facebook.com/v21.0/dialog/oauth?" + params.Encode()

	case "tiktok":
		appID := strings.TrimSpace(os.Getenv("TIKTOK_APP_ID"))
		if appID == "" {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "tiktok_not_configured"})
		}
		params := url.Values{}
		params.Set("app_id", appID)
		params.Set("state", state)
		params.Set("redirect_uri", redirectURI)
		authURL = "https://business-api.tiktok.com/portal/auth?" + params.Encode()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"platform":     platform,
		"url":          authURL,
		"state":        state,
		"redirect_uri": redirectURI,
	})
}

// ---------------------------------------------------------------------------
// HandleSignalOAuthCallback
// GET /api/v1/signal/oauth/:platform/callback?code=...&state=...
// Public — no RequireWorkspace; auth is derived from the signed state token.
// ---------------------------------------------------------------------------

// HandleSignalOAuthCallback godoc
func HandleSignalOAuthCallback(c echo.Context) error {
	platform := strings.TrimSpace(strings.ToLower(c.Param("platform")))
	if platform != "meta" && platform != "tiktok" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
	}

	state := strings.TrimSpace(c.QueryParam("state"))
	code := strings.TrimSpace(c.QueryParam("code"))
	errorParam := strings.TrimSpace(c.QueryParam("error"))

	appURL := strings.TrimRight(strings.TrimSpace(os.Getenv("NEXT_PUBLIC_APP_URL")), "/")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	redirectBase := appURL + "/settings/integrations?platform=" + platform

	if errorParam != "" {
		return c.Redirect(http.StatusFound, redirectBase+"&status=denied&reason="+url.QueryEscape(errorParam))
	}
	if code == "" || state == "" {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=missing_params")
	}

	orgID, _, err := verifyOAuthState(state)
	if err != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=invalid_state")
	}

	if dbPool == nil {
		if _, initErr := queries(c.Request().Context()); initErr != nil {
			return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=database_unavailable")
		}
	}

	q, qErr := queries(c.Request().Context())
	if qErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=database_unavailable")
	}

	workspaceID, wsErr := workspaceIDForOrg(c.Request().Context(), q, orgID)
	if wsErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=workspace_not_found")
	}

	if migrErr := ensureSignalOAuthColumns(c.Request().Context()); migrErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=schema_migration_failed")
	}

	redirectURI := appURL + "/settings/integrations/callback/" + platform

	var (
		accessToken    string
		refreshToken   string
		accountID      string
		accountName    string
		tokenExpiresAt time.Time
		exchangeErr    error
	)

	switch platform {
	case "meta":
		accessToken, tokenExpiresAt, exchangeErr = exchangeMetaOAuthCode(c.Request().Context(), code, redirectURI)
		if exchangeErr == nil {
			accountID, accountName, exchangeErr = fetchMetaAdAccountInfo(c.Request().Context(), accessToken)
		}
	case "tiktok":
		accessToken, refreshToken, tokenExpiresAt, exchangeErr = exchangeTikTokOAuthCode(c.Request().Context(), code, redirectURI)
		if exchangeErr == nil {
			accountID, accountName, exchangeErr = fetchTikTokAdAccountInfo(c.Request().Context(), accessToken)
		}
	}

	if exchangeErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=token_exchange_failed")
	}

	encryptedAccess, encErr := encryptToken(accessToken)
	if encErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=encryption_failed")
	}

	encryptedRefresh := ""
	if refreshToken != "" {
		encryptedRefresh, encErr = encryptToken(refreshToken)
		if encErr != nil {
			return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=encryption_failed")
		}
	}

	_, dbErr := dbPool.Exec(
		c.Request().Context(),
		`INSERT INTO signal_connections (
			workspace_id, platform, status, account_id, account_name,
			encrypted_access_token, encrypted_refresh_token, token_expires_at, last_synced_at
		 ) VALUES ($1, $2, 'connected', $3, $4, $5, NULLIF($6, ''), $7, NOW())
		 ON CONFLICT (workspace_id, platform, account_id)
		 DO UPDATE SET
		   status = 'connected',
		   account_name = EXCLUDED.account_name,
		   encrypted_access_token = EXCLUDED.encrypted_access_token,
		   encrypted_refresh_token = COALESCE(NULLIF(EXCLUDED.encrypted_refresh_token, ''), signal_connections.encrypted_refresh_token),
		   token_expires_at = EXCLUDED.token_expires_at,
		   error_reason = NULL,
		   updated_at = NOW()`,
		workspaceID,
		platform,
		accountID,
		accountName,
		encryptedAccess,
		encryptedRefresh,
		tokenExpiresAt.UTC(),
	)
	if dbErr != nil {
		return c.Redirect(http.StatusFound, redirectBase+"&status=error&reason=connection_save_failed")
	}

	return c.Redirect(http.StatusFound, redirectBase+"&status=connected&account="+url.QueryEscape(accountName))
}

// ---------------------------------------------------------------------------
// Meta OAuth helpers
// ---------------------------------------------------------------------------

type metaTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func exchangeMetaOAuthCode(ctx context.Context, code, redirectURI string) (string, time.Time, error) {
	appID := strings.TrimSpace(os.Getenv("META_APP_ID"))
	appSecret := strings.TrimSpace(os.Getenv("META_APP_SECRET"))
	if appID == "" || appSecret == "" {
		return "", time.Time{}, fmt.Errorf("META_APP_ID or META_APP_SECRET not configured")
	}

	params := url.Values{}
	params.Set("client_id", appID)
	params.Set("client_secret", appSecret)
	params.Set("redirect_uri", redirectURI)
	params.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.facebook.com/v21.0/oauth/access_token?"+params.Encode(), nil)
	if err != nil {
		return "", time.Time{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	var tokenResp metaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, err
	}
	if tokenResp.Error != nil {
		return "", time.Time{}, fmt.Errorf("meta token error: %s", tokenResp.Error.Message)
	}

	// Exchange for long-lived token (60 days)
	longLivedToken, expiresAt, llErr := exchangeMetaLongLivedToken(ctx, tokenResp.AccessToken, appID, appSecret)
	if llErr != nil {
		// Fall back to short-lived token
		longLivedToken = tokenResp.AccessToken
		expiresAt = time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}
	return longLivedToken, expiresAt, nil
}

func exchangeMetaLongLivedToken(ctx context.Context, shortToken, appID, appSecret string) (string, time.Time, error) {
	params := url.Values{}
	params.Set("grant_type", "fb_exchange_token")
	params.Set("client_id", appID)
	params.Set("client_secret", appSecret)
	params.Set("fb_exchange_token", shortToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.facebook.com/v21.0/oauth/access_token?"+params.Encode(), nil)
	if err != nil {
		return "", time.Time{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	var tokenResp metaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, err
	}
	if tokenResp.Error != nil {
		return "", time.Time{}, fmt.Errorf("meta long-lived token error: %s", tokenResp.Error.Message)
	}
	expires := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expires, nil
}

func fetchMetaAdAccountInfo(ctx context.Context, accessToken string) (string, string, error) {
	params := url.Values{}
	params.Set("fields", "id,name")
	params.Set("access_token", accessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.facebook.com/v21.0/me/adaccounts?"+params.Encode(), nil)
	if err != nil {
		return "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	if result.Error != nil {
		return "", "", fmt.Errorf("meta accounts error: %s", result.Error.Message)
	}
	if len(result.Data) == 0 {
		return "", "", fmt.Errorf("no ad accounts found on this Meta user")
	}
	accountID := strings.TrimPrefix(result.Data[0].ID, "act_")
	return accountID, result.Data[0].Name, nil
}

// ---------------------------------------------------------------------------
// TikTok OAuth helpers
// ---------------------------------------------------------------------------

type tikTokTokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken          string  `json:"access_token"`
		RefreshToken         string  `json:"refresh_token"`
		AccessTokenExpiresIn int     `json:"access_token_expires_in"`
		AdvertiserIDs        []int64 `json:"advertiser_ids"`
	} `json:"data"`
}

func exchangeTikTokOAuthCode(ctx context.Context, code, _ string) (string, string, time.Time, error) {
	appID := strings.TrimSpace(os.Getenv("TIKTOK_APP_ID"))
	appSecret := strings.TrimSpace(os.Getenv("TIKTOK_APP_SECRET"))
	if appID == "" || appSecret == "" {
		return "", "", time.Time{}, fmt.Errorf("TIKTOK_APP_ID or TIKTOK_APP_SECRET not configured")
	}

	body, _ := json.Marshal(map[string]string{
		"app_id":    appID,
		"secret":    appSecret,
		"auth_code": code,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://business-api.tiktok.com/open_api/v1.3/oauth2/access_token/",
		bytes.NewReader(body))
	if err != nil {
		return "", "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, err
	}
	defer resp.Body.Close()

	var tokenResp tikTokTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", "", time.Time{}, err
	}
	if tokenResp.Code != 0 {
		return "", "", time.Time{}, fmt.Errorf("tiktok token error %d: %s", tokenResp.Code, tokenResp.Message)
	}

	expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.Data.AccessTokenExpiresIn) * time.Second)
	return tokenResp.Data.AccessToken, tokenResp.Data.RefreshToken, expiresAt, nil
}

func fetchTikTokAdAccountInfo(ctx context.Context, accessToken string) (string, string, error) {
	appID := strings.TrimSpace(os.Getenv("TIKTOK_APP_ID"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://business-api.tiktok.com/open_api/v1.3/advertiser/info/", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Access-Token", accessToken)
	q := req.URL.Query()
	q.Set("app_id", appID)
	q.Set("fields", `["advertiser_id","advertiser_name","status"]`)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				AdvertiserID   int64  `json:"advertiser_id"`
				AdvertiserName string `json:"advertiser_name"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	if result.Code != 0 {
		return "", "", fmt.Errorf("tiktok accounts error %d: %s", result.Code, result.Message)
	}
	if len(result.Data.List) == 0 {
		return "", "", fmt.Errorf("no TikTok ad accounts found")
	}

	accountID := fmt.Sprintf("%d", result.Data.List[0].AdvertiserID)
	return accountID, result.Data.List[0].AdvertiserName, nil
}
