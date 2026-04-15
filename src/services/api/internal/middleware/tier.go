package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// PlanTierLimits enforces variant generation limits per Qvora pricing tier.
// Starter: 3 variants/angle | Growth: 10 variants/angle | Agency: unlimited
// Server-side enforcement — NEVER trust client-side limits.
var planVariantLimits = map[string]int{
	"starter": 3,
	"growth":  10,
	"agency":  -1, // unlimited
}

// EnforceVariantLimit validates the workspace plan tier against requested variant count.
// Expects orgPlanTier in echo context (set by workspace loader middleware in Phase 1).
func EnforceVariantLimit(requestedCount int, planTier string) (int, error) {
	limit, ok := planVariantLimits[planTier]
	if !ok {
		limit = planVariantLimits["starter"] // default to most restrictive
	}

	if limit == -1 {
		return requestedCount, nil // agency — unlimited
	}

	if requestedCount > limit {
		return 0, echo.NewHTTPError(http.StatusPaymentRequired, map[string]string{
			"error": "variant_limit_exceeded",
			"limit": strconv.Itoa(limit),
			"tier":  planTier,
		})
	}

	return requestedCount, nil
}
