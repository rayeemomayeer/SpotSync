package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

type rateLimitStore interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type redisRateStore struct {
	eval func(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error)
}

const redisFixedWindowScript = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
if current > tonumber(ARGV[2]) then
  return 0
end
return 1
`

func (s *redisRateStore) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	n, err := s.eval(ctx, redisFixedWindowScript, []string{key}, window.Milliseconds(), limit)
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// IPRateLimitRedis uses Redis fixed-window counters when available; falls back to in-memory.
func IPRateLimitRedis(requestsPerMinute int, store rateLimitStore) echo.MiddlewareFunc {
	if store == nil {
		return IPRateLimit(requestsPerMinute)
	}
	if requestsPerMinute < 1 {
		requestsPerMinute = 10
	}
	window := time.Minute
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := fmt.Sprintf("spotsync:ratelimit:auth:%s", c.RealIP())
			ok, err := store.Allow(c.Request().Context(), key, requestsPerMinute, window)
			if err != nil {
				// Fail open to in-process limiter on Redis errors.
				return IPRateLimit(requestsPerMinute)(next)(c)
			}
			if !ok {
				c.Response().Header().Set("Retry-After", "60")
				return domain.ErrRateLimited
			}
			return next(c)
		}
	}
}

func NewRedisRateStore(eval func(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error)) rateLimitStore {
	if eval == nil {
		return nil
	}
	return &redisRateStore{eval: eval}
}
