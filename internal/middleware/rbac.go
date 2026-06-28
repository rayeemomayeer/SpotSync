package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !IsAdmin(c) {
				return domain.ErrForbidden
			}
			return next(c)
		}
	}
}
