package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	var req dto.CreatePaymentRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	payment, err := h.svc.Record(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusCreated, "Payment recorded", service.ToPaymentResponse(payment))
}

func (h *PaymentHandler) Get(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	payment, err := h.svc.Get(c.Request().Context(), userID, id, appmw.Role(c))
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Payment retrieved", service.ToPaymentResponse(payment))
}

func (h *PaymentHandler) ListMine(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	list, err := h.svc.ListMine(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	out := make([]dto.PaymentResponse, 0, len(list))
	for i := range list {
		out = append(out, service.ToPaymentResponse(&list[i]))
	}
	return JSONSuccess(c, http.StatusOK, "Payments retrieved", out)
}

func (h *PaymentHandler) Refund(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req dto.CreateRefundRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}
	ref, err := h.svc.RecordRefund(c.Request().Context(), userID, id, appmw.Role(c), req)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusCreated, "Refund recorded", service.ToRefundResponse(ref))
}
