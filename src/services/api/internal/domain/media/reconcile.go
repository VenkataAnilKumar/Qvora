package media

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	appmiddleware "github.com/qvora/api/internal/middleware"
	"github.com/qvora/api/internal/store"
	"github.com/qvora/api/internal/util"
)

// ReconcileStuckJobs handles POST /api/v1/internal/jobs/reconcile-stuck
// Marks stale in-flight jobs/variants as failed to prevent infinite limbo states.
func ReconcileStuckJobs(c echo.Context) error {
	claims := appmiddleware.GetClaims(c)
	if claims == nil || strings.TrimSpace(claims.UserID) == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	role := strings.TrimSpace(strings.ToLower(claims.OrgRole))
	if role != "worker" && role != "system" && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}
	if store.Pool() == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database_unavailable"})
	}
	maxAgeMinutes := util.ParsePositiveInt(c.QueryParam("max_age_minutes"), 45)
	if maxAgeMinutes < 15 {
		maxAgeMinutes = 15
	}
	if maxAgeMinutes > 720 {
		maxAgeMinutes = 720
	}
	var failedJobs int64
	if err := store.Pool().QueryRow(
		c.Request().Context(),
		`WITH stale_jobs AS (
			SELECT id
			FROM jobs
			WHERE status IN ('scraping','generating','postprocessing')
			  AND updated_at < NOW() - ($1::int || ' minutes')::interval
		), updated AS (
			UPDATE jobs
			SET status = 'failed',
			    error_msg = COALESCE(error_msg, 'reconciled_stuck_job_timeout'),
			    updated_at = NOW()
			WHERE id IN (SELECT id FROM stale_jobs)
			RETURNING id
		)
		SELECT COUNT(*)::bigint FROM updated`,
		maxAgeMinutes,
	).Scan(&failedJobs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "job_reconcile_failed"})
	}
	var failedVariants int64
	if err := store.Pool().QueryRow(
		c.Request().Context(),
		`WITH stale_variants AS (
			SELECT id
			FROM variants
			WHERE status IN ('generating','postprocessing')
			  AND updated_at < NOW() - ($1::int || ' minutes')::interval
		), updated AS (
			UPDATE variants
			SET status = CASE WHEN status = 'complete' THEN status ELSE 'failed' END,
			    updated_at = NOW()
			WHERE id IN (SELECT id FROM stale_variants)
			RETURNING id
		)
		SELECT COUNT(*)::bigint FROM updated`,
		maxAgeMinutes,
	).Scan(&failedVariants); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "variant_reconcile_failed"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"max_age_minutes": maxAgeMinutes,
		"failed_jobs":     failedJobs,
		"failed_variants": failedVariants,
	})
}
