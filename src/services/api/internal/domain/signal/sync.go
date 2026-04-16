package signal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
)

// metricRow is one day's performance data for one ad creative on one platform.
type metricRow struct {
	AdID        string
	AdName      string
	MetricDate  time.Time
	Impressions int64
	Clicks      int64
	SpendUSD    float64
	Conversions int64
	RevenueUSD  float64
	Hold25      *float64
	Hold50      *float64
	Hold75      *float64
	Hold100     *float64
}

// SyncSignalMetricsAll handles POST /api/v1/internal/signal/metrics/sync-all
// Worker-triggered; iterates every active connection and pulls ad metrics.
func SyncSignalMetricsAll(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	role := strings.ToLower(strings.TrimSpace(claims.OrgRole))
	if role != "worker" && role != "system" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}

	var req struct {
		Days int `json:"days"`
	}
	if bindErr := c.Bind(&req); bindErr != nil || req.Days <= 0 {
		req.Days = 30
	}
	if req.Days > 90 {
		req.Days = 90
	}

	if migrErr := ensureSignalOAuthColumns(c.Request().Context()); migrErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "schema_migration_failed"})
	}

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT id, workspace_id, platform, account_id, encrypted_access_token
		 FROM signal_connections
		 WHERE status = 'connected'
		   AND encrypted_access_token IS NOT NULL
		   AND (token_expires_at IS NULL OR token_expires_at > NOW() + INTERVAL '10 minutes')
		 ORDER BY last_synced_at ASC NULLS FIRST
		 LIMIT 100`,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "connections_query_failed"})
	}
	defer rows.Close()

	type connInfo struct {
		id          int64
		workspaceID pgtype.UUID
		platform    string
		accountID   string
		encToken    string
	}
	var connections []connInfo
	for rows.Next() {
		var conn connInfo
		if scanErr := rows.Scan(&conn.id, &conn.workspaceID, &conn.platform, &conn.accountID, &conn.encToken); scanErr != nil {
			continue
		}
		connections = append(connections, conn)
	}
	rows.Close()

	since := time.Now().UTC().AddDate(0, 0, -req.Days)
	until := time.Now().UTC()

	synced := 0
	failed := 0
	for _, conn := range connections {
		accessToken, decErr := decryptToken(conn.encToken)
		if decErr != nil {
			failed++
			continue
		}

		var metrics []metricRow
		var fetchErr error
		switch conn.platform {
		case "meta":
			metrics, fetchErr = fetchMetaAdMetrics(c.Request().Context(), accessToken, conn.accountID, since, until)
		case "tiktok":
			metrics, fetchErr = fetchTikTokAdMetrics(c.Request().Context(), accessToken, conn.accountID, since, until)
		}

		if fetchErr != nil {
			_, _ = store.Pool().Exec(c.Request().Context(),
				`UPDATE signal_connections SET error_reason = $1, updated_at = NOW() WHERE id = $2`,
				"sync_fetch_failed: "+fetchErr.Error(), conn.id,
			)
			failed++
			continue
		}

		if upsertErr := upsertPlatformMetrics(c.Request().Context(), conn.workspaceID, conn.platform, metrics); upsertErr != nil {
			failed++
			continue
		}

		_, _ = store.Pool().Exec(c.Request().Context(),
			`UPDATE signal_connections SET last_synced_at = NOW(), error_reason = NULL, updated_at = NOW() WHERE id = $1`,
			conn.id,
		)
		synced++
	}

	return c.JSON(http.StatusOK, map[string]any{
		"total":   len(connections),
		"synced":  synced,
		"failed":  failed,
		"days":    req.Days,
		"message": "metrics sync complete",
	})
}

// upsertPlatformMetrics writes metric rows into signal_metrics_daily.
// Resolves variant_id by platform_asset_id when available.
func upsertPlatformMetrics(ctx context.Context, workspaceID pgtype.UUID, platform string, metrics []metricRow) error {
	if store.Pool() == nil {
		return fmt.Errorf("database_not_initialized")
	}
	if len(metrics) == 0 {
		return nil
	}

	tx, err := store.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, m := range metrics {
		var variantID any
		if m.AdID != "" {
			var vid pgtype.UUID
			lookErr := tx.QueryRow(ctx,
				`SELECT id FROM variants WHERE workspace_id = $1 AND platform_asset_id = $2 LIMIT 1`,
				workspaceID, m.AdID,
			).Scan(&vid)
			if lookErr == nil && vid.Valid {
				variantID = vid
			}
		}

		_, execErr := tx.Exec(ctx,
			`INSERT INTO signal_metrics_daily (
				workspace_id, variant_id, platform, metric_date,
				impressions, clicks, spend_usd, conversions, revenue_usd,
				hold_25, hold_50, hold_75, hold_100
			 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			 ON CONFLICT (workspace_id, platform, variant_id, metric_date)
			 DO UPDATE SET
			   impressions = EXCLUDED.impressions,
			   clicks = EXCLUDED.clicks,
			   spend_usd = EXCLUDED.spend_usd,
			   conversions = EXCLUDED.conversions,
			   revenue_usd = EXCLUDED.revenue_usd,
			   hold_25 = EXCLUDED.hold_25,
			   hold_50 = EXCLUDED.hold_50,
			   hold_75 = EXCLUDED.hold_75,
			   hold_100 = EXCLUDED.hold_100,
			   updated_at = NOW()`,
			workspaceID, variantID, platform, m.MetricDate,
			m.Impressions, m.Clicks, m.SpendUSD, m.Conversions, m.RevenueUSD,
			m.Hold25, m.Hold50, m.Hold75, m.Hold100,
		)
		if execErr != nil {
			return execErr
		}
	}

	return tx.Commit(ctx)
}

// ---------------------------------------------------------------------------
// Meta Insights API client
// ---------------------------------------------------------------------------

func fetchMetaAdMetrics(ctx context.Context, accessToken, accountID string, since, until time.Time) ([]metricRow, error) {
	adAccountID := accountID
	if !strings.HasPrefix(adAccountID, "act_") {
		adAccountID = "act_" + accountID
	}

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("fields", "ad_id,ad_name,impressions,clicks,spend,actions,action_values,video_continuous_2_sec_watched_actions,video_thruplay_watched_actions")
	params.Set("level", "ad")
	params.Set("time_increment", "1")
	params.Set("time_range", fmt.Sprintf(`{"since":"%s","until":"%s"}`,
		since.Format("2006-01-02"), until.Format("2006-01-02")))
	params.Set("limit", "1000")

	apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/insights?%s", adAccountID, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			AdID        string `json:"ad_id"`
			AdName      string `json:"ad_name"`
			DateStop    string `json:"date_stop"`
			Impressions string `json:"impressions"`
			Clicks      string `json:"clicks"`
			Spend       string `json:"spend"`
			Actions     []struct {
				ActionType string `json:"action_type"`
				Value      string `json:"value"`
			} `json:"actions"`
			ActionValues []struct {
				ActionType string `json:"action_type"`
				Value      string `json:"value"`
			} `json:"action_values"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("meta insights error %d: %s", result.Error.Code, result.Error.Message)
	}

	var rows []metricRow
	for _, d := range result.Data {
		metricDate, parseErr := time.Parse("2006-01-02", d.DateStop)
		if parseErr != nil {
			continue
		}
		m := metricRow{
			AdID:        d.AdID,
			AdName:      d.AdName,
			MetricDate:  metricDate,
			Impressions: signalParseInt64(d.Impressions),
			Clicks:      signalParseInt64(d.Clicks),
			SpendUSD:    signalParseFloat64(d.Spend),
		}
		for _, action := range d.Actions {
			if action.ActionType == "purchase" || action.ActionType == "offsite_conversion.fb_pixel_purchase" {
				m.Conversions += signalParseInt64(action.Value)
			}
		}
		for _, av := range d.ActionValues {
			if av.ActionType == "purchase" || av.ActionType == "offsite_conversion.fb_pixel_purchase" {
				m.RevenueUSD += signalParseFloat64(av.Value)
			}
		}
		rows = append(rows, m)
	}
	return rows, nil
}

// ---------------------------------------------------------------------------
// TikTok Report API client
// ---------------------------------------------------------------------------

func fetchTikTokAdMetrics(ctx context.Context, accessToken, accountID string, since, until time.Time) ([]metricRow, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"advertiser_id": accountID,
		"report_type":   "BASIC",
		"data_level":    "AUCTION_AD",
		"dimensions":    []string{"ad_id", "stat_time_day"},
		"metrics": []string{
			"ad_name",
			"impressions",
			"clicks",
			"spend",
			"conversions",
			"video_watched_2s",
			"video_watched_6s",
		},
		"start_date": since.Format("2006-01-02"),
		"end_date":   until.Format("2006-01-02"),
		"page":       1,
		"page_size":  1000,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://business-api.tiktok.com/open_api/v1.3/report/integrated/get/",
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				Dimensions struct {
					AdID        string `json:"ad_id"`
					StatTimeDay string `json:"stat_time_day"`
				} `json:"dimensions"`
				Metrics struct {
					AdName         string `json:"ad_name"`
					Impressions    string `json:"impressions"`
					Clicks         string `json:"clicks"`
					Spend          string `json:"spend"`
					Conversions    string `json:"conversions"`
					VideoWatched2s string `json:"video_watched_2s"`
					VideoWatched6s string `json:"video_watched_6s"`
				} `json:"metrics"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("tiktok report error %d: %s", result.Code, result.Message)
	}

	var rows []metricRow
	for _, item := range result.Data.List {
		metricDate, parseErr := time.Parse("2006-01-02", item.Dimensions.StatTimeDay)
		if parseErr != nil {
			continue
		}

		impressions := signalParseInt64(item.Metrics.Impressions)
		watched2s := signalParseFloat64(item.Metrics.VideoWatched2s)
		watched6s := signalParseFloat64(item.Metrics.VideoWatched6s)

		var hold25, hold50 *float64
		if impressions > 0 {
			h25 := (watched2s / float64(impressions)) * 100.0
			hold25 = &h25
			h50 := (watched6s / float64(impressions)) * 100.0
			hold50 = &h50
		}

		m := metricRow{
			AdID:        item.Dimensions.AdID,
			AdName:      item.Metrics.AdName,
			MetricDate:  metricDate,
			Impressions: impressions,
			Clicks:      signalParseInt64(item.Metrics.Clicks),
			SpendUSD:    signalParseFloat64(item.Metrics.Spend),
			Conversions: signalParseInt64(item.Metrics.Conversions),
			Hold25:      hold25,
			Hold50:      hold50,
		}
		rows = append(rows, m)
	}
	return rows, nil
}

// signalParseInt64 safely parses a numeric string to int64.
func signalParseInt64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

// signalParseFloat64 safely parses a numeric string to float64.
func signalParseFloat64(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
