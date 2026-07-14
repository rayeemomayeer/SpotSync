package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/openapi"
)

func OpenAPIYAML(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml", openapi.Spec)
}
