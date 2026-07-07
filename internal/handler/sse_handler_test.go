package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
	"github.com/rayeemomayeer/SpotSync/internal/realtime"
)

func TestSSEHandler_StreamZoneEvents(t *testing.T) {
	hub := realtime.NewHub()
	sse := handler.NewSSEHandler(hub)
	e := echo.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/zones/7/events", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("7")

	go func() {
		_ = sse.StreamZoneEvents(c)
	}()

	time.Sleep(30 * time.Millisecond)
	hub.Publish(realtime.ZoneEvent{
		Type:          realtime.EventSpotReserved,
		ZoneID:        7,
		SpotID:        3,
		UserID:        9,
		ReservationID: 42,
		LicensePlate:  "XYZ-9999",
	})

	deadline := time.After(2 * time.Second)
	for {
		body := rec.Body.String()
		if strings.Contains(body, "spot_reserved") && strings.Contains(body, "XYZ-9999") {
			cancel()
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for SSE payload, body=%q", rec.Body.String())
		case <-time.After(15 * time.Millisecond):
		}
	}
}
