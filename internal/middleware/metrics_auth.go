package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

// MetricsAuth gates /metrics behind a shared bearer token when configured.
// When token is empty, the endpoint stays open (local/dev convenience).
func MetricsAuth(token string) echo.MiddlewareFunc {
	expected := strings.TrimSpace(token)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if expected == "" {
				return next(c)
			}

			got := bearerOrQueryToken(c)
			if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
				return domain.ErrUnauthorized
			}
			return next(c)
		}
	}
}

func bearerOrQueryToken(c echo.Context) string {
	header := c.Request().Header.Get(echo.HeaderAuthorization)
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	if q := strings.TrimSpace(c.QueryParam("token")); q != "" {
		return q
	}
	return ""
}
