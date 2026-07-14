package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/outbox"
)

type OutboxHandler struct {
	repo *outbox.Repository
}

func NewOutboxHandler(repo *outbox.Repository) *OutboxHandler {
	return &OutboxHandler{repo: repo}
}

func (h *OutboxHandler) ListDeadLetter(c echo.Context) error {
	limit := 50
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return domain.NewValidationError("Validation failed", map[string]string{"limit": "Must be a positive integer"})
		}
		limit = n
	}
	list, err := h.repo.FetchDeadLettered(c.Request().Context(), limit)
	if err != nil {
		return err
	}
	out := make([]dto.OutboxEventResponse, 0, len(list))
	for _, ev := range list {
		out = append(out, dto.OutboxEventResponse{
			ID:             ev.ID,
			AggregateType:  ev.AggregateType,
			AggregateID:    ev.AggregateID,
			EventType:      ev.EventType,
			Attempts:       ev.Attempts,
			LastError:      ev.LastError,
			DeadLetteredAt: ev.DeadLetteredAt,
			CreatedAt:      ev.CreatedAt,
		})
	}
	return JSONSuccess(c, http.StatusOK, "Dead-letter outbox events retrieved", out)
}

func (h *OutboxHandler) Replay(c echo.Context) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.repo.ReplayDeadLetter(c.Request().Context(), id); err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Outbox event queued for replay", map[string]uint{"id": id})
}
