package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type stubZoneService struct {
	listFn   func(ctx context.Context, q dto.ZoneListQuery) ([]dto.ZoneResponse, error)
	updateFn func(ctx context.Context, id uint, req dto.UpdateZoneRequest) (dto.ZoneResponse, error)
	deleteFn func(ctx context.Context, id uint) error
}

func (s *stubZoneService) Create(context.Context, dto.CreateZoneRequest, *uint) (dto.ZoneResponse, error) {
	return dto.ZoneResponse{}, nil
}

func (s *stubZoneService) List(ctx context.Context, q dto.ZoneListQuery) ([]dto.ZoneResponse, error) {
	if s.listFn != nil {
		return s.listFn(ctx, q)
	}
	return nil, nil
}

func (s *stubZoneService) GetByID(context.Context, uint) (dto.ZoneResponse, error) {
	return dto.ZoneResponse{}, nil
}

func (s *stubZoneService) Update(ctx context.Context, id uint, req dto.UpdateZoneRequest) (dto.ZoneResponse, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, id, req)
	}
	return dto.ZoneResponse{}, nil
}

func (s *stubZoneService) Delete(ctx context.Context, id uint) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

type stubReservationService struct {
	listAllFn func(ctx context.Context, q dto.PaginationQuery) (service.ListAllResult, error)
}

func (s *stubReservationService) Create(context.Context, uint, dto.CreateReservationRequest, service.CreateReservationOptions) (dto.ReservationResponse, error) {
	return dto.ReservationResponse{}, nil
}

func (s *stubReservationService) Cancel(context.Context, uint, uint) (dto.ReservationResponse, error) {
	return dto.ReservationResponse{}, nil
}

func (s *stubReservationService) ListMine(context.Context, uint) ([]dto.ReservationResponse, error) {
	return nil, nil
}

func (s *stubReservationService) ListAll(ctx context.Context, q dto.PaginationQuery) (service.ListAllResult, error) {
	if s.listAllFn != nil {
		return s.listAllFn(ctx, q)
	}
	return service.ListAllResult{}, nil
}

func TestZoneHandler_ListInvalidSort(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()
	e.HTTPErrorHandler = handler.HTTPErrorHandler

	h := handler.NewZoneHandler(&stubZoneService{}, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/zones?sort=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.List(c); err != nil {
		handler.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestZoneHandler_DeleteActiveReservationsConflict(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handler.HTTPErrorHandler

	h := handler.NewZoneHandler(&stubZoneService{
		deleteFn: func(context.Context, uint) error {
			return domain.ErrZoneHasActiveReservations
		},
	}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/zones/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	if err := h.Delete(c); err != nil {
		handler.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestReservationHandler_ListAllPaginationHeaders(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()

	h := handler.NewReservationHandler(&stubReservationService{
		listAllFn: func(_ context.Context, _ dto.PaginationQuery) (service.ListAllResult, error) {
			return service.ListAllResult{
				Items:     []dto.ReservationResponse{{ID: 1}},
				Total:     42,
				Page:      2,
				Limit:     10,
				Paginated: true,
			}, nil
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reservations?page=2&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListAll(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("X-Total-Count"); got != "42" {
		t.Fatalf("X-Total-Count=%q", got)
	}
	if got := rec.Header().Get("X-Page"); got != "2" {
		t.Fatalf("X-Page=%q", got)
	}
	if got := rec.Header().Get("X-Limit"); got != "10" {
		t.Fatalf("X-Limit=%q", got)
	}

	var resp dto.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatal("expected success envelope")
	}
}

func TestAuthHandler_Me(t *testing.T) {
	e := echo.New()
	h := handler.NewAuthHandler(&stubAuthService{
		meFn: func(_ context.Context, userID uint) (dto.UserResponse, error) {
			return dto.UserResponse{ID: userID, Email: "a@b.com"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(7))

	if err := h.Me(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestZoneHandler_Update(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()

	h := handler.NewZoneHandler(&stubZoneService{
		updateFn: func(_ context.Context, id uint, req dto.UpdateZoneRequest) (dto.ZoneResponse, error) {
			return dto.ZoneResponse{ID: id, Name: req.Name}, nil
		},
	}, nil)

	body := `{"name":"Updated Lot","type":"general","total_capacity":10,"price_per_hour":4.5}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/zones/3", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("3")

	if err := h.Update(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
