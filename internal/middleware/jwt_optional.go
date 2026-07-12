package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

// JWTAuthFromHeaderOrQuery accepts Authorization Bearer or ?access_token=
// so browser EventSource clients can authenticate SSE streams.
func JWTAuthFromHeaderOrQuery(tokens TokenVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := ""
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if strings.HasPrefix(header, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			}
			if token == "" {
				token = strings.TrimSpace(c.QueryParam("access_token"))
			}
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
