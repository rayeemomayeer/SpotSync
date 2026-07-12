package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	appmw "github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
)

type NotificationHandler struct {
	repo *repository.NotificationRepository
}

func NewNotificationHandler(repo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

type notificationResponse struct {
	ID        uint       `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ReadAt    *string    `json:"read_at"`
	CreatedAt string     `json:"created_at"`
}

func (h *NotificationHandler) ListMine(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	list, err := h.repo.ListByUser(c.Request().Context(), userID, 30)
	if err != nil {
		return err
	}
	out := make([]notificationResponse, 0, len(list))
	for _, n := range list {
		out = append(out, toNotificationResponse(n))
	}
	return JSONSuccess(c, http.StatusOK, "Notifications retrieved", out)
}

func (h *NotificationHandler) MarkRead(c echo.Context) error {
	userID, ok := appmw.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.repo.MarkRead(c.Request().Context(), userID, id); err != nil {
		return err
	}
	return JSONSuccess(c, http.StatusOK, "Notification marked read", nil)
}

func toNotificationResponse(n models.Notification) notificationResponse {
	var readAt *string
	if n.ReadAt != nil {
		s := n.ReadAt.UTC().Format(timeRFC3339)
		readAt = &s
	}
	return notificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		ReadAt:    readAt,
		CreatedAt: n.CreatedAt.UTC().Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
