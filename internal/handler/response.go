package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
)

// JSONSuccess writes a graded success envelope.
func JSONSuccess(c echo.Context, status int, message string, data any) error {
	return c.JSON(status, dto.Success(message, data))
}

// JSONError writes a graded error envelope.
func JSONError(c echo.Context, status int, message string, fieldErrors map[string]string) error {
	return c.JSON(status, dto.Error(message, fieldErrors))
}

// NoContentSuccess writes 200 with an empty data envelope (e.g. cancel reservation).
func NoContentSuccess(c echo.Context, message string) error {
	return JSONSuccess(c, http.StatusOK, message, nil)
}
