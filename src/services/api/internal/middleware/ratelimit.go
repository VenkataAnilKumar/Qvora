package middleware

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimiter uses Upstash Redis for per-user rate limiting.
// 60 requests per minute per user. Unauthenticated requests use IP.
// Requires UPSTASH_REDIS_URL in rediss://default:TOKEN@host:6379 format.
func RateLimiter() echo.MiddlewareFunc {
	opt, err := redis.ParseURL(os.Getenv("UPSTASH_REDIS_URL"))
	if err != nil {
		panic("ratelimit: invalid UPSTASH_REDIS_URL: " + err.Error())
	}
	rdb := redis.NewClient(opt)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := rateLimitKey(c)
			ctx := c.Request().Context()

			pipe := rdb.Pipeline()
			incr := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, 60*time.Second)
			if _, err := pipe.Exec(ctx); err != nil {
				// Redis unavailable — fail open (don't block requests)
				return next(c)
			}

			count := incr.Val()
			c.Response().Header().Set("X-RateLimit-Limit", "60")
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.FormatInt(max(int64(0), 60-count), 10))

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

