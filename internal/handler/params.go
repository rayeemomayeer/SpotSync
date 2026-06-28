package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

func parseUintParam(c echo.Context, name string) (uint, error) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, domain.NewValidationError("Validation failed", map[string]string{
			name: "Invalid id",
		})
	}
	return uint(id), nil
}
