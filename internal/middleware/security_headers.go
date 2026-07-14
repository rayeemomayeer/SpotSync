package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func SecurityHeaders() echo.MiddlewareFunc {
	enableHSTS := strings.EqualFold(os.Getenv("ENABLE_HSTS"), "true")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			h.Set("X-XSS-Protection", "0")
			if enableHSTS {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			return next(c)
		}
	}
}
