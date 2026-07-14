package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/realtime"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

const demoReservationHeader = "X-Demo-Reservation"
const idempotencyHeader = "Idempotency-Key"

type ReservationService interface {
	Create(ctx context.Context, userID uint, req dto.CreateReservationRequest, opts service.CreateReservationOptions) (dto.ReservationResponse, error)
	Cancel(ctx context.Context, userID, reservationID uint) (dto.ReservationResponse, error)
	ListMine(ctx context.Context, userID uint) ([]dto.ReservationResponse, error)
	ListAll(ctx context.Context, q dto.PaginationQuery) (service.ListAllResult, error)
}

type ReservationHandler struct {
	reservations ReservationService
	hub          *realtime.Hub
	zones        ZoneAvailabilityInvalidator
}

type ZoneAvailabilityInvalidator interface {
	InvalidateAvailability(ctx context.Context, zoneID uint)
}

func NewReservationHandler(
	reservations ReservationService,
	hub *realtime.Hub,
	zones ZoneAvailabilityInvalidator,
) *ReservationHandler {
	return &ReservationHandler{reservations: reservations, hub: hub, zones: zones}
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
	if key := strings.TrimSpace(c.Request().Header.Get(idempotencyHeader)); key != "" {
		opts.IdempotencyKey = &key
	}
	if middleware.IsDemoMode(c) {
		if sid := strings.TrimSpace(middleware.DemoSessionID(c)); sid != "" {
			opts.DemoSessionID = &sid
		}
	}

	res, err := h.reservations.Create(c.Request().Context(), userID, req, opts)
	if err != nil {
		return err
	}

	h.publishReserved(res)
	h.invalidateZoneCache(c.Request().Context(), res.ZoneID)

	return JSONSuccess(c, http.StatusCreated, "Reservation created successfully", res)
}

func (h *ReservationHandler) publishReserved(res dto.ReservationResponse) {
	if h.hub == nil || res.SpotID == nil {
		return
	}
	h.hub.Publish(realtime.ZoneEvent{
		Type:          realtime.EventSpotReserved,
		ZoneID:        res.ZoneID,
		SpotID:        *res.SpotID,
		UserID:        res.UserID,
		ReservationID: res.ID,
		LicensePlate:  res.LicensePlate,
	})
}

func (h *ReservationHandler) publishReleased(res dto.ReservationResponse) {
	if h.hub == nil || res.SpotID == nil {
		return
	}
	h.hub.Publish(realtime.ZoneEvent{
		Type:          realtime.EventSpotReleased,
		ZoneID:        res.ZoneID,
		SpotID:        *res.SpotID,
		UserID:        res.UserID,
		ReservationID: res.ID,
	})
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

	cancelled, err := h.reservations.Cancel(c.Request().Context(), userID, id)
	if err != nil {
		return err
	}

	h.publishReleased(cancelled)
	h.invalidateZoneCache(c.Request().Context(), cancelled.ZoneID)

	return NoContentSuccess(c, "Reservation cancelled successfully")
}

func (h *ReservationHandler) invalidateZoneCache(ctx context.Context, zoneID uint) {
	if h.zones != nil {
		h.zones.InvalidateAvailability(ctx, zoneID)
	}
}

func (h *ReservationHandler) ListAll(c echo.Context) error {
	var q dto.PaginationQuery
	if err := BindAndValidate(c, &q); err != nil {
		return err
	}

	result, err := h.reservations.ListAll(c.Request().Context(), q)
	if err != nil {
		return err
	}
	list := result.Items
	if list == nil {
		list = []dto.ReservationResponse{}
	}
	if result.Paginated {
		setPaginationHeaders(c, result.Total, result.Page, result.Limit)
	}
	return JSONSuccess(c, http.StatusOK, "Reservations retrieved successfully", list)
}
