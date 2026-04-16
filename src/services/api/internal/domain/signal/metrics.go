package signal

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// UpsertSignalMetrics handles POST /api/v1/signal/metrics
func UpsertSignalMetrics(c echo.Context) error {
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

	var req struct {
		Metrics []struct {
			VariantID   *string  `json:"variant_id"`
			Platform    string   `json:"platform"`
			Date        string   `json:"date"`
			Impressions int64    `json:"impressions"`
			Clicks      int64    `json:"clicks"`
			SpendUSD    float64  `json:"spend_usd"`
			Conversions int64    `json:"conversions"`
			RevenueUSD  float64  `json:"revenue_usd"`
			Hold25      *float64 `json:"hold_25"`
			Hold50      *float64 `json:"hold_50"`
			Hold75      *float64 `json:"hold_75"`
			Hold100     *float64 `json:"hold_100"`
		} `json:"metrics"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_request"})
	}
	if len(req.Metrics) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "metrics_required"})
	}

	tx, err := store.Pool().Begin(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
	}
	defer tx.Rollback(c.Request().Context()) //nolint:errcheck

	rowsUpserted := 0
	for _, metric := range req.Metrics {
		platform := strings.TrimSpace(strings.ToLower(metric.Platform))
		if platform != "meta" && platform != "tiktok" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_platform"})
		}

		metricDate, parseErr := time.Parse("2006-01-02", strings.TrimSpace(metric.Date))
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_date"})
		}

		var variantID any
		if metric.VariantID != nil && strings.TrimSpace(*metric.VariantID) != "" {
			parsed, parseErr := util.ParseUUID(strings.TrimSpace(*metric.VariantID))
			if parseErr != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_variant_id"})
			}
			variantID = parsed
		}

		_, execErr := tx.Exec(
			c.Request().Context(),
			`INSERT INTO signal_metrics_daily (
				workspace_id, variant_id, platform, metric_date, impressions, clicks, spend_usd, conversions, revenue_usd, hold_25, hold_50, hold_75, hold_100
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
			workspaceID,
			variantID,
			platform,
			metricDate,
			metric.Impressions,
			metric.Clicks,
			metric.SpendUSD,
			metric.Conversions,
			metric.RevenueUSD,
			metric.Hold25,
			metric.Hold50,
			metric.Hold75,
			metric.Hold100,
		)
		if execErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
		}

		rowsUpserted++
	}

	if err := tx.Commit(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_metrics_upsert_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"upserted":     rowsUpserted,
	})
}

// GetSignalDashboard handles GET /api/v1/signal/dashboard?days=30
func GetSignalDashboard(c echo.Context) error {
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

	var impressions int64
	var clicks int64
	var spendUSD float64
	var conversions int64
	var revenueUSD float64
	err = store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(SUM(impressions), 0),
			COALESCE(SUM(clicks), 0),
			COALESCE(SUM(spend_usd), 0),
			COALESCE(SUM(conversions), 0),
			COALESCE(SUM(revenue_usd), 0)
		 FROM signal_metrics_daily
		 WHERE workspace_id = $1
		   AND metric_date >= CURRENT_DATE - ($2::int - 1)`,
		workspaceID,
		days,
	).Scan(&impressions, &clicks, &spendUSD, &conversions, &revenueUSD)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	ctr := 0.0
	if impressions > 0 {
		ctr = (float64(clicks) / float64(impressions)) * 100
	}
	cpa := 0.0
	if conversions > 0 {
		cpa = spendUSD / float64(conversions)
	}
	roas := 0.0
	if spendUSD > 0 {
		roas = revenueUSD / spendUSD
	}

	var feedbackTotal int64
	var feedbackAccepted int64
	err = store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(COUNT(*), 0) AS feedback_total,
			COALESCE(COUNT(*) FILTER (WHERE action = 'accept'), 0) AS feedback_accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1
		   AND created_at >= NOW() - ($2::int || ' days')::interval`,
		workspaceID,
		days,
	).Scan(&feedbackTotal, &feedbackAccepted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	feedbackAcceptanceRate := 0.0
	if feedbackTotal > 0 {
		feedbackAcceptanceRate = (float64(feedbackAccepted) / float64(feedbackTotal)) * 100
	}

	var current7dTotal int64
	var current7dAccepted int64
	var previous7dTotal int64
	var previous7dAccepted int64
	err = store.Pool().QueryRow(
		c.Request().Context(),
		`SELECT
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days'), 0) AS current_total,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days' AND action = 'accept'), 0) AS current_accepted,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '14 days' AND created_at < NOW() - INTERVAL '7 days'), 0) AS previous_total,
			COALESCE(COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '14 days' AND created_at < NOW() - INTERVAL '7 days' AND action = 'accept'), 0) AS previous_accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1`,
		workspaceID,
	).Scan(&current7dTotal, &current7dAccepted, &previous7dTotal, &previous7dAccepted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	current7dRate := 0.0
	if current7dTotal > 0 {
		current7dRate = (float64(current7dAccepted) / float64(current7dTotal)) * 100
	}
	previous7dRate := 0.0
	if previous7dTotal > 0 {
		previous7dRate = (float64(previous7dAccepted) / float64(previous7dTotal)) * 100
	}
	acceptanceDeltaPctPoints := current7dRate - previous7dRate

	trendRows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT
			(created_at AT TIME ZONE 'UTC')::date AS day,
			COUNT(*)::bigint AS total,
			COUNT(*) FILTER (WHERE action = 'accept')::bigint AS accepted
		 FROM signal_recommendation_feedback
		 WHERE workspace_id = $1
		   AND created_at >= NOW() - INTERVAL '14 days'
		 GROUP BY day
		 ORDER BY day ASC`,
		workspaceID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer trendRows.Close()

	type dailyTrend struct {
		total    int64
		accepted int64
	}
	trendByDate := make(map[string]dailyTrend)
	for trendRows.Next() {
		var day time.Time
		var total int64
		var accepted int64
		if scanErr := trendRows.Scan(&day, &total, &accepted); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}
		dateKey := day.UTC().Format("2006-01-02")
		trendByDate[dateKey] = dailyTrend{total: total, accepted: accepted}
	}
	if rowsErr := trendRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	acceptanceTrend := make([]map[string]any, 0, 14)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := 13; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		dateKey := day.Format("2006-01-02")
		entry := trendByDate[dateKey]
		rate := 0.0
		if entry.total > 0 {
			rate = (float64(entry.accepted) / float64(entry.total)) * 100
		}
		acceptanceTrend = append(acceptanceTrend, map[string]any{
			"date":            dateKey,
			"total":           entry.total,
			"accepted":        entry.accepted,
			"acceptance_rate": rate,
		})
	}

	angleRows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT
			COALESCE(v.angle, 'unknown') AS angle,
			COALESCE(SUM(m.impressions), 0) AS impressions,
			COALESCE(SUM(m.clicks), 0) AS clicks,
			COALESCE(SUM(m.spend_usd), 0) AS spend_usd,
			COALESCE(SUM(m.conversions), 0) AS conversions,
			COALESCE(SUM(m.revenue_usd), 0) AS revenue_usd
		 FROM signal_metrics_daily m
		 LEFT JOIN variants v ON v.id = m.variant_id
		 WHERE m.workspace_id = $1
		   AND m.metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY COALESCE(v.angle, 'unknown')
		 ORDER BY clicks DESC, impressions DESC
		 LIMIT 10`,
		workspaceID,
		days,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer angleRows.Close()

	byAngle := make([]map[string]any, 0)
	for angleRows.Next() {
		var angle string
		var aImpressions int64
		var aClicks int64
		var aSpend float64
		var aConversions int64
		var aRevenue float64
		if scanErr := angleRows.Scan(&angle, &aImpressions, &aClicks, &aSpend, &aConversions, &aRevenue); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}

		aCtr := 0.0
		if aImpressions > 0 {
			aCtr = (float64(aClicks) / float64(aImpressions)) * 100
		}

		byAngle = append(byAngle, map[string]any{
			"angle":           angle,
			"impressions":     aImpressions,
			"clicks":          aClicks,
			"spend_usd":       aSpend,
			"conversions":     aConversions,
			"revenue_usd":     aRevenue,
			"ctr":             aCtr,
			"meets_threshold": aImpressions >= 1000,
		})
	}
	if rowsErr := angleRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	platformRows, err := store.Pool().Query(
		c.Request().Context(),
		`SELECT
			platform,
			COALESCE(SUM(impressions), 0) AS impressions,
			COALESCE(SUM(clicks), 0) AS clicks,
			COALESCE(SUM(spend_usd), 0) AS spend_usd,
			COALESCE(SUM(conversions), 0) AS conversions,
			COALESCE(SUM(revenue_usd), 0) AS revenue_usd
		 FROM signal_metrics_daily
		 WHERE workspace_id = $1
		   AND metric_date >= CURRENT_DATE - ($2::int - 1)
		 GROUP BY platform
		 ORDER BY impressions DESC`,
		workspaceID,
		days,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}
	defer platformRows.Close()

	byPlatform := make([]map[string]any, 0)
	for platformRows.Next() {
		var platform string
		var pImpressions int64
		var pClicks int64
		var pSpend float64
		var pConversions int64
		var pRevenue float64
		if scanErr := platformRows.Scan(&platform, &pImpressions, &pClicks, &pSpend, &pConversions, &pRevenue); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
		}

		pCtr := 0.0
		if pImpressions > 0 {
			pCtr = (float64(pClicks) / float64(pImpressions)) * 100
		}

		byPlatform = append(byPlatform, map[string]any{
			"platform":    platform,
			"impressions": pImpressions,
			"clicks":      pClicks,
			"spend_usd":   pSpend,
			"conversions": pConversions,
			"revenue_usd": pRevenue,
			"ctr":         pCtr,
		})
	}
	if rowsErr := platformRows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_dashboard_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"days":         days,
		"totals": map[string]any{
			"impressions": impressions,
			"clicks":      clicks,
			"spend_usd":   spendUSD,
			"conversions": conversions,
			"revenue_usd": revenueUSD,
			"ctr":         ctr,
			"cpa":         cpa,
			"roas":        roas,
		},
		"recommendation_feedback": map[string]any{
			"total":                       feedbackTotal,
			"accepted":                    feedbackAccepted,
			"acceptance_rate":             feedbackAcceptanceRate,
			"current_7d_rate":             current7dRate,
			"previous_7d_rate":            previous7dRate,
			"acceptance_delta_pct_points": acceptanceDeltaPctPoints,
			"trend":                       acceptanceTrend,
		},
		"by_angle":    byAngle,
		"by_platform": byPlatform,
	})
}
