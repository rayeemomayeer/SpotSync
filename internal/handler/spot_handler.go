package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
)

type SpotService interface {
	ListByZone(ctx context.Context, zoneID uint) ([]dto.SpotResponse, error)
	UpdateStatus(ctx context.Context, zoneID, spotID uint, status string) (dto.SpotResponse, error)
}

type SpotHandler struct {
	spots SpotService
}

func NewSpotHandler(spots SpotService) *SpotHandler {
	return &SpotHandler{spots: spots}
}

func (h *SpotHandler) ListByZone(c echo.Context) error {
	zoneID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	list, err := h.spots.ListByZone(c.Request().Context(), zoneID)
	if err != nil {
		return err
	}
	if list == nil {
		list = []dto.SpotResponse{}
	}
	return JSONSuccess(c, http.StatusOK, "Spots retrieved successfully", list)
}

func (h *SpotHandler) UpdateStatus(c echo.Context) error {
	zoneID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	spotID, err := parseUintParam(c, "spotId")
	if err != nil {
		return err
	}

	var req dto.UpdateSpotRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	spot, err := h.spots.UpdateStatus(c.Request().Context(), zoneID, spotID, req.Status)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Spot updated successfully", spot)
}
