package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/realtime"
)

const sseHeartbeatInterval = 15 * time.Second

type SSEHandler struct {
	hub *realtime.Hub
}

func NewSSEHandler(hub *realtime.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// StreamAllZones streams availability changes for all zones (SSE).
// GET /api/v1/zones/stream
func (h *SSEHandler) StreamAllZones(c echo.Context) error {
	if h.hub == nil {
		return fmt.Errorf("streaming not supported")
	}
	return h.stream(c, h.hub.SubscribeGlobal(), func(ch chan realtime.ZoneEvent) {
		h.hub.UnsubscribeGlobal(ch)
	})
}

func (h *SSEHandler) stream(c echo.Context, ch chan realtime.ZoneEvent, cleanup func(chan realtime.ZoneEvent)) error {
	defer cleanup(ch)

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "text/event-stream")
	res.Header().Set(echo.HeaderCacheControl, "no-cache")
	res.Header().Set(echo.HeaderConnection, "keep-alive")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)

	flusher, ok := res.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	ctx := c.Request().Context()
	ticker := time.NewTicker(sseHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := fmt.Fprintf(res, ": heartbeat\n\n"); err != nil {
				return nil
			}
			flusher.Flush()
		case ev, open := <-ch:
			if !open {
				return nil
			}
			payload, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(res, "event: %s\ndata: %s\n\n", ev.Type, payload); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

// StreamZoneEvents streams spot_reserved / spot_released for a zone (SSE).
// GET /api/v1/zones/:id/events
func (h *SSEHandler) StreamZoneEvents(c echo.Context) error {
	zoneID, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	if h.hub == nil {
		return fmt.Errorf("streaming not supported")
	}

	ch := h.hub.Subscribe(zoneID)
	return h.stream(c, ch, func(c chan realtime.ZoneEvent) {
		h.hub.Unsubscribe(zoneID, c)
	})
}
