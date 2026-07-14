package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

type DemoHandler struct {
	svc *service.DemoService
}

func NewDemoHandler(svc *service.DemoService) *DemoHandler {
	return &DemoHandler{svc: svc}
}

type DemoResetResponse struct {
	ReservationsCancelled int64 `json:"reservations_cancelled"`
	ReservationsDeleted   int64 `json:"reservations_deleted"`
	ZonesDeleted          int64 `json:"zones_deleted"`
	AuditDeleted          int64 `json:"audit_deleted"`
}

func (h *DemoHandler) Reset(c echo.Context) error {
	if !appmw.IsDemoMode(c) {
		return domain.ErrForbidden
	}
	sessionID := strings.TrimSpace(appmw.DemoSessionID(c))
	if sessionID == "" {
		return domain.NewValidationError("Validation failed", map[string]string{
			"demo_session_id": "Required in demo mode",
		})
	}
	stats, err := h.svc.ResetSession(c.Request().Context(), sessionID)
	if err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Demo session reset", DemoResetResponse{
		ReservationsCancelled: stats.ReservationsCancelled,
		ReservationsDeleted:   stats.ReservationsDeleted,
		ZonesDeleted:          stats.ZonesDeleted,
		AuditDeleted:          stats.AuditDeleted,
	})
}
