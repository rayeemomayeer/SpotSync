package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const readinessTimeout = 2 * time.Second

// ReadinessChecker verifies backing services are reachable.
type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

// MultiReadinessChecker runs all configured dependency checks.
type MultiReadinessChecker struct {
	Checkers []ReadinessChecker
}

func (m *MultiReadinessChecker) Ping(ctx context.Context) error {
	for _, checker := range m.Checkers {
		if checker == nil {
			continue
		}
		if err := checker.Ping(ctx); err != nil {
			return err
		}
	}
	return nil
}

// RedisReadinessChecker pings Redis when configured.
type RedisReadinessChecker struct {
	PingFn func(ctx context.Context) error
}

func (c *RedisReadinessChecker) Ping(ctx context.Context) error {
	if c == nil || c.PingFn == nil {
		return nil
	}
	return c.PingFn(ctx)
}

// DBReadinessChecker adapts a database pinger for readiness probes.
type DBReadinessChecker struct {
	PingFn func(ctx context.Context) error
}

// Ping implements ReadinessChecker.
func (c *DBReadinessChecker) Ping(ctx context.Context) error {
	if c == nil || c.PingFn == nil {
		return nil
	}
	return c.PingFn(ctx)
}

// HealthHandler serves liveness and readiness endpoints.
type HealthHandler struct {
	checker ReadinessChecker
}

// NewHealthHandler returns a handler that probes checker for /readyz.
func NewHealthHandler(checker ReadinessChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

type healthResponse struct {
	Status string `json:"status"`
}

// Healthz reports process liveness (no dependency checks).
func (h *HealthHandler) Healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

// Readyz reports readiness based on configured dependency checks.
func (h *HealthHandler) Readyz(c echo.Context) error {
	if h.checker != nil {
		ctx, cancel := context.WithTimeout(c.Request().Context(), readinessTimeout)
		defer cancel()

		if err := h.checker.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "not ready"})
		}
	}

	return c.JSON(http.StatusOK, healthResponse{Status: "ready"})
}
