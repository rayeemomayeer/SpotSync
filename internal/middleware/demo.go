package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	DemoModeHeader       = "X-Demo-Mode"
	DemoSessionHeader    = "X-Demo-Session-Id"
	echoDemoModeKey      = "demo_mode"
	echoDemoSessionIDKey = "demo_session_id"
)

func DemoContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			mode := strings.TrimSpace(c.Request().Header.Get(DemoModeHeader))
			sessionID := strings.TrimSpace(c.Request().Header.Get(DemoSessionHeader))
			if strings.EqualFold(mode, "true") || mode == "1" {
				c.Set(echoDemoModeKey, true)
			}
			if sessionID != "" {
				c.Set(echoDemoSessionIDKey, sessionID)
			}
			return next(c)
		}
	}
}

func IsDemoMode(c echo.Context) bool {
	v, ok := c.Get(echoDemoModeKey).(bool)
	return ok && v
}

func DemoSessionID(c echo.Context) string {
	v, _ := c.Get(echoDemoSessionIDKey).(string)
	return v
}
