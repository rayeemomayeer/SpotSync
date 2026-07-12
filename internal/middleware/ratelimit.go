package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"golang.org/x/time/rate"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

func newIPRateLimiter(requestsPerMinute int) *ipRateLimiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 10
	}
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:    requestsPerMinute,
	}
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.limit, l.burst)
		l.limiters[ip] = lim
	}
	return lim
}

func IPRateLimit(requestsPerMinute int) echo.MiddlewareFunc {
	limiter := newIPRateLimiter(requestsPerMinute)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.get(c.RealIP()).Allow() {
				c.Response().Header().Set("Retry-After", "60")
				return domain.ErrRateLimited
			}
			return next(c)
		}
	}
}
