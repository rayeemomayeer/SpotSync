package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

const (
	keyUserID    = "userID"
	keyRole      = "role"
	keyRequestID = "requestID"
)

func UserID(c echo.Context) (uint, bool) {
	v, ok := c.Get(keyUserID).(uint)
	return v, ok
}

func Role(c echo.Context) string {
	v, _ := c.Get(keyRole).(string)
	return v
}

func GetRequestID(c echo.Context) string {
	v, _ := c.Get(keyRequestID).(string)
	return v
}

func IsAdmin(c echo.Context) bool {
	return Role(c) == models.RoleAdmin
}
