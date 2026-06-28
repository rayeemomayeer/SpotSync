package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

type TokenVerifier interface {
	Verify(tokenString string) (uint, string, error)
}

func JWTAuth(tokens TokenVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if !strings.HasPrefix(header, "Bearer ") {
				return domain.ErrUnauthorized
			}

			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if token == "" {
				return domain.ErrUnauthorized
			}

			userID, role, err := tokens.Verify(token)
			if err != nil {
				return err
			}

			c.Set(keyUserID, userID)
			c.Set(keyRole, role)
			return next(c)
		}
	}
}

var _ TokenVerifier = (*platform.TokenManager)(nil)
