package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimiter uses Upstash Redis (HTTP) for per-user rate limiting.
// 60 requests per minute per user. Unauthenticated requests use IP.
func RateLimiter() echo.MiddlewareFunc {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("UPSTASH_REDIS_REST_URL"),
		Password: os.Getenv("UPSTASH_REDIS_REST_TOKEN"),
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := rateLimitKey(c)
			ctx := context.Background()

			pipe := rdb.Pipeline()
			incr := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, 60*1_000_000_000) // 60 seconds
			if _, err := pipe.Exec(ctx); err != nil {
				// Redis unavailable — fail open (don't block requests)
				return next(c)
			}

			count := incr.Val()
			c.Response().Header().Set("X-RateLimit-Limit", "60")
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(max(0, 60-count), 10))

			if count > 60 {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate_limit_exceeded"})
			}

			return next(c)
		}
	}
}

func rateLimitKey(c echo.Context) string {
	claims := GetClaims(c)
	if claims != nil && claims.UserID != "" {
		return "rl:user:" + claims.UserID
	}
	return "rl:ip:" + c.RealIP()
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
