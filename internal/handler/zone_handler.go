package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
)

type ZoneService interface {
	Create(ctx context.Context, req dto.CreateZoneRequest) (dto.ZoneResponse, error)
	List(ctx context.Context) ([]dto.ZoneResponse, error)
	GetByID(ctx context.Context, id uint) (dto.ZoneResponse, error)
}

type ZoneHandler struct {
	zones ZoneService
}

func NewZoneHandler(zones ZoneService) *ZoneHandler {
	return &ZoneHandler{zones: zones}
}

func (h *ZoneHandler) Create(c echo.Context) error {
	var req dto.CreateZoneRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	zone, err := h.zones.Create(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusCreated, "Zone created successfully", zone)
}

func (h *ZoneHandler) List(c echo.Context) error {
	zones, err := h.zones.List(c.Request().Context())
	if err != nil {
		return err
	}
	if zones == nil {
		zones = []dto.ZoneResponse{}
	}
	return JSONSuccess(c, http.StatusOK, "Zones retrieved successfully", zones)
}

func (h *ZoneHandler) GetByID(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	zone, err := h.zones.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Zone retrieved successfully", zone)
}
