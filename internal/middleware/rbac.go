package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

func IsPlatformAdmin(c echo.Context) bool {
	role := Role(c)
	return role == models.RoleAdmin || role == models.RoleSaaSAdmin
}

func IsOrgAdmin(c echo.Context) bool {
	return Role(c) == models.RoleOrgAdmin
}

func IsAdmin(c echo.Context) bool {
	return IsPlatformAdmin(c) || IsOrgAdmin(c)
}

// RequireAdmin allows graded admin, saas_admin, or org_admin.
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

func RequirePlatformAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !IsPlatformAdmin(c) {
				return domain.ErrForbidden
			}
			return next(c)
		}
	}
}
