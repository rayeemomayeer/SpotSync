package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

const demoReservationHeader = "X-Demo-Reservation"

type ReservationService interface {
	Create(ctx context.Context, userID uint, req dto.CreateReservationRequest, opts service.CreateReservationOptions) (dto.ReservationResponse, error)
	Cancel(ctx context.Context, userID, reservationID uint) error
	ListMine(ctx context.Context, userID uint) ([]dto.ReservationResponse, error)
	ListAll(ctx context.Context, q dto.PaginationQuery) ([]dto.ReservationResponse, error)
}

type ReservationHandler struct {
	reservations ReservationService
}

func NewReservationHandler(reservations ReservationService) *ReservationHandler {
	return &ReservationHandler{reservations: reservations}
}

func (h *ReservationHandler) Create(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}

	var req dto.CreateReservationRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	opts := service.CreateReservationOptions{
		DemoReservation: isDemoReservation(c),
	}

	res, err := h.reservations.Create(c.Request().Context(), userID, req, opts)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusCreated, "Reservation created successfully", res)
}

func isDemoReservation(c echo.Context) bool {
	v := strings.TrimSpace(c.Request().Header.Get(demoReservationHeader))
	return strings.EqualFold(v, "true") || v == "1"
}

func (h *ReservationHandler) ListMine(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}

	list, err := h.reservations.ListMine(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	if list == nil {
		list = []dto.ReservationResponse{}
	}
	return JSONSuccess(c, http.StatusOK, "Reservations retrieved successfully", list)
}

func (h *ReservationHandler) Cancel(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.reservations.Cancel(c.Request().Context(), userID, id); err != nil {
		return err
	}

	return NoContentSuccess(c, "Reservation cancelled successfully")
}

func (h *ReservationHandler) ListAll(c echo.Context) error {
	var q dto.PaginationQuery
	if err := BindAndValidate(c, &q); err != nil {
		return err
	}

	list, err := h.reservations.ListAll(c.Request().Context(), q)
	if err != nil {
		return err
	}
	if list == nil {
		list = []dto.ReservationResponse{}
	}
	return JSONSuccess(c, http.StatusOK, "Reservations retrieved successfully", list)
}
