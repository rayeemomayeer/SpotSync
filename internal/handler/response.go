package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
)

func JSONSuccess(c echo.Context, status int, message string, data any) error {
	return c.JSON(status, dto.Success(message, data))
}

func JSONError(c echo.Context, status int, message string, fieldErrors map[string]string) error {
	return c.JSON(status, dto.Error(message, fieldErrors))
}

func NoContentSuccess(c echo.Context, message string) error {
	return JSONSuccess(c, http.StatusOK, message, nil)
}
