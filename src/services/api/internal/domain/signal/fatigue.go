package signal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// resolveSuggestedAction maps a CTR drop percentage to a recommended action.
func resolveSuggestedAction(dropPct float64) string {
	if dropPct >= 50 {
		return "refresh_with_new_angle"
	}
	if dropPct >= 40 {
		return "refresh_hook_and_opening"
	}
	return "refresh_hook"
}

// toFloat64 safely extracts a float64 from an any value.
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	}
	return 0
}

// sendFatigueAlertEmail sends a fatigue alert summary via Resend. Non-fatal on failure.
func sendFatigueAlertEmail(ctx context.Context, orgID, workspaceID string, alerts []map[string]any) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return nil
	}
	if store.Pool() == nil {
		return nil
	}

	var notificationEmail string
	_ = store.Pool().QueryRow(ctx,
		`SELECT COALESCE(notification_email, '') FROM workspaces WHERE org_id = $1 LIMIT 1`,
		orgID,
	).Scan(&notificationEmail)
	if strings.TrimSpace(notificationEmail) == "" {
		return nil
	}

	count := len(alerts)
	subject := fmt.Sprintf("Qvora Signal: %d creative fatigue alert", count)
	if count != 1 {
		subject = fmt.Sprintf("Qvora Signal: %d creative fatigue alerts", count)
	}

	var htmlBuf bytes.Buffer
	htmlBuf.WriteString(`<h2 style="color:#FF3D3D">Creative Fatigue Detected</h2>`)
	htmlBuf.WriteString(fmt.Sprintf(`<p>%d variant(s) in workspace <code>%s</code> are showing sustained CTR decline.</p>`, count, workspaceID))
	htmlBuf.WriteString(`<table border="1" cellpadding="6" style="border-collapse:collapse;font-family:monospace;font-size:12px"><thead><tr><th>Angle</th><th>Variant</th><th>Drop %%</th><th>Action</th></tr></thead><tbody>`)
	for _, alert := range alerts {
		htmlBuf.WriteString(fmt.Sprintf(
			`<tr><td>%v</td><td>%v</td><td>%.1f%%</td><td>%v</td></tr>`,
			alert["angle"], alert["variant_id"],
			toFloat64(alert["drop_pct"]),
			alert["suggested_action"],
		))
	}
	htmlBuf.WriteString(`</tbody></table>`)
	htmlBuf.WriteString(`<p style="margin-top:16px;font-size:11px;color:#888">Sent by Qvora Signal — <a href="https://app.qvora.com/dashboard">View dashboard</a></p>`)

	fromEmail := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = "signal@qvora.ai"
	}

	body, _ := json.Marshal(map[string]any{
		"from":    "Qvora Signal <" + fromEmail + ">",
		"to":      []string{notificationEmail},
		"subject": subject,
		"html":    htmlBuf.String(),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// DetectSignalFatigue handles GET /api/v1/signal/fatigue
func DetectSignalFatigue(c echo.Context) error {
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
	dropPct := parsePositiveFloat(c.QueryParam("drop_pct"), 30)
	if dropPct > 95 {
		dropPct = 95
	}
	dropRatio := dropPct / 100
	sustainedDays := util.ParsePositiveInt(c.QueryParam("sustained_days"), 3)
	if sustainedDays < 3 {
		sustainedDays = 3
	}
	if sustainedDays > 7 {
		sustainedDays = 7
	}
	minPeakCtr := parsePositiveFloat(c.QueryParam("min_peak_ctr"), 0.01)
	minImpressions := int64(util.ParsePositiveInt(c.QueryParam("min_impressions"), 1000))
	persist := parseBool(c.QueryParam("persist"), true)

	windowURL := url.QueryEscape("" + fmt.Sprintf("%d", days) + "d")

	rows, err := store.Pool().Query(
		c.Request().Context(),
		`WITH daily AS (
			SELECT
				variant_id,
				metric_date,
				COALESCE(SUM(impressions), 0) AS impressions,
				CASE WHEN COALESCE(SUM(impressions), 0) > 0
					THEN COALESCE(SUM(clicks), 0)::float8 / COALESCE(SUM(impressions), 0)::float8
					ELSE 0
				END AS ctr
			FROM signal_metrics_daily
			WHERE workspace_id = $1
			  AND variant_id IS NOT NULL
			  AND metric_date >= CURRENT_DATE - ($2::int - 1)
			GROUP BY variant_id, metric_date
		),
		scored AS (
			SELECT
				variant_id,
				metric_date,
				impressions,
				ctr,
				MAX(ctr) OVER (
					PARTITION BY variant_id
					ORDER BY metric_date
					ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
				) AS peak_7d
			FROM daily
		),
		flagged AS (
			SELECT
				variant_id,
				metric_date,
				impressions,
				ctr,
				peak_7d,
				CASE
					WHEN peak_7d >= $3::float8
					 AND impressions >= $4::bigint
					 AND ctr <= peak_7d * (1 - $5::float8)
					THEN 1 ELSE 0
				END AS is_drop
			FROM scored
		),
		streak AS (
			SELECT
				variant_id,
				metric_date,
				ctr,
				peak_7d,
				is_drop,
				CASE
					WHEN is_drop = 1
					 AND LAG(is_drop, 1, 0) OVER (PARTITION BY variant_id ORDER BY metric_date) = 1
					 AND LAG(is_drop, 2, 0) OVER (PARTITION BY variant_id ORDER BY metric_date) = 1
					THEN 3 ELSE 0
				END AS sustained_days
			FROM flagged
		),
		candidates AS (
			SELECT DISTINCT ON (variant_id)
				variant_id,
				metric_date AS detected_on,
				ctr AS current_ctr,
				peak_7d,
				((peak_7d - ctr) / NULLIF(peak_7d, 0)) * 100.0 AS drop_pct,
				sustained_days
			FROM streak
			WHERE sustained_days >= $6::int
			ORDER BY variant_id, metric_date DESC
		)
		SELECT
			c.variant_id,
			c.detected_on,
			c.current_ctr,
			c.peak_7d,
			COALESCE(c.drop_pct, 0),
			c.sustained_days,
			COALESCE(v.angle, 'unknown') AS angle
		FROM candidates c
		LEFT JOIN variants v ON v.id = c.variant_id
		ORDER BY c.drop_pct DESC, c.detected_on DESC`,
		workspaceID,
		days,
		minPeakCtr,
		minImpressions,
		dropRatio,
		sustainedDays,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
	}
	defer rows.Close()

	alerts := make([]map[string]any, 0)
	for rows.Next() {
		var variantID pgtype.UUID
		var detectedOn time.Time
		var currentCtr float64
		var peakCtr float64
		var dropComputed float64
		var sustained int
		var angle string
		if scanErr := rows.Scan(&variantID, &detectedOn, &currentCtr, &peakCtr, &dropComputed, &sustained, &angle); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
		}

		suggestedAction := resolveSuggestedAction(dropComputed)
		alert := map[string]any{
			"variant_id":       util.UUIDString(variantID),
			"detected_on":      detectedOn.Format("2006-01-02"),
			"angle":            angle,
			"current_ctr":      currentCtr,
			"peak_ctr":         peakCtr,
			"drop_pct":         dropComputed,
			"sustained_days":   sustained,
			"suggested_action": suggestedAction,
			"refresh_link":     "/dashboard?refresh_variant=" + url.QueryEscape(util.UUIDString(variantID)) + "&window=" + windowURL,
		}
		alerts = append(alerts, alert)

		if persist {
			_, execErr := store.Pool().Exec(
				c.Request().Context(),
				`INSERT INTO signal_fatigue_events (
					workspace_id, variant_id, detected_on, current_ctr, peak_ctr, drop_pct, sustained_days, suggested_action, status, first_detected_at, last_detected_at
				 ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', NOW(), NOW())
				 ON CONFLICT (workspace_id, variant_id, detected_on)
				 DO UPDATE SET
				   current_ctr = EXCLUDED.current_ctr,
				   peak_ctr = EXCLUDED.peak_ctr,
				   drop_pct = EXCLUDED.drop_pct,
				   sustained_days = EXCLUDED.sustained_days,
				   suggested_action = EXCLUDED.suggested_action,
				   status = 'active',
				   resolved_at = NULL,
				   last_detected_at = NOW(),
				   updated_at = NOW()`,
				workspaceID,
				variantID,
				detectedOn,
				currentCtr,
				peakCtr,
				dropComputed,
				sustained,
				suggestedAction,
			)
			if execErr != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_detect_failed"})
	}

	if persist && len(alerts) > 0 {
		_ = sendFatigueAlertEmail(c.Request().Context(), claims.OrgID, util.UUIDString(workspaceID), alerts)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":         claims.OrgID,
		"workspace_id":   util.UUIDString(workspaceID),
		"days":           days,
		"drop_pct":       dropPct,
		"sustained_days": sustainedDays,
		"persisted":      persist,
		"alerts":         alerts,
	})
}

// ListSignalFatigueEvents handles GET /api/v1/signal/fatigue/events
func ListSignalFatigueEvents(c echo.Context) error {
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

	limit := util.ParsePositiveInt(c.QueryParam("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	status := strings.TrimSpace(strings.ToLower(c.QueryParam("status")))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "resolved" && status != "all" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_status"})
	}

	query := `SELECT
		f.variant_id,
		f.detected_on,
		f.current_ctr,
		f.peak_ctr,
		f.drop_pct,
		f.sustained_days,
		f.suggested_action,
		f.status,
		f.first_detected_at,
		f.last_detected_at,
		f.resolved_at,
		COALESCE(v.angle, 'unknown') AS angle
	 FROM signal_fatigue_events f
	 LEFT JOIN variants v ON v.id = f.variant_id
	 WHERE f.workspace_id = $1`

	args := []any{workspaceID}
	if status != "all" {
		query += ` AND f.status = $2`
		args = append(args, status)
		query += ` ORDER BY f.last_detected_at DESC LIMIT $3`
		args = append(args, limit)
	} else {
		query += ` ORDER BY f.last_detected_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := store.Pool().Query(c.Request().Context(), query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
	}
	defer rows.Close()

	events := make([]map[string]any, 0)
	for rows.Next() {
		var variantID pgtype.UUID
		var detectedOn time.Time
		var currentCtr float64
		var peakCtr float64
		var dropPct float64
		var sustainedDays int
		var suggestedAction string
		var evtStatus string
		var firstDetectedAt time.Time
		var lastDetectedAt time.Time
		var resolvedAt pgtype.Timestamptz
		var angle string
		if scanErr := rows.Scan(
			&variantID,
			&detectedOn,
			&currentCtr,
			&peakCtr,
			&dropPct,
			&sustainedDays,
			&suggestedAction,
			&evtStatus,
			&firstDetectedAt,
			&lastDetectedAt,
			&resolvedAt,
			&angle,
		); scanErr != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
		}

		events = append(events, map[string]any{
			"variant_id":        util.UUIDString(variantID),
			"detected_on":       detectedOn.Format("2006-01-02"),
			"angle":             angle,
			"current_ctr":       currentCtr,
			"peak_ctr":          peakCtr,
			"drop_pct":          dropPct,
			"sustained_days":    sustainedDays,
			"suggested_action":  suggestedAction,
			"status":            evtStatus,
			"first_detected_at": firstDetectedAt.UTC(),
			"last_detected_at":  lastDetectedAt.UTC(),
			"resolved_at":       util.TsTime(resolvedAt),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "signal_fatigue_list_failed"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org_id":       claims.OrgID,
		"workspace_id": util.UUIDString(workspaceID),
		"status":       status,
		"events":       events,
	})
}
