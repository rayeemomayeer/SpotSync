package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status, message, fieldErrors := mapError(err)

	if status >= http.StatusInternalServerError {
		slog.Error("unhandled server error", "error", err, "path", c.Request().URL.Path)
	}

	if writeErr := JSONError(c, status, message, fieldErrors); writeErr != nil {
		slog.Error("failed to write error response", "error", writeErr)
	}
}

func mapError(err error) (status int, message string, fieldErrors map[string]string) {
	fieldErrors = map[string]string{}

	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		if validationErr.Fields != nil {
			fieldErrors = validationErr.Fields
		}
		msg := validationErr.Message
		if msg == "" {
			msg = "Validation failed"
		}
		return http.StatusBadRequest, msg, fieldErrors
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		msg := "Request error"
		if m, ok := echoErr.Message.(string); ok {
			msg = m
		}
		return echoErr.Code, msg, fieldErrors
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "Resource not found", fieldErrors
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Unauthorized", fieldErrors
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrNotOwner):
		return http.StatusForbidden, "Forbidden", fieldErrors
	case errors.Is(err, domain.ErrZoneFull):
		return http.StatusConflict, "Zone is full", fieldErrors
	case errors.Is(err, domain.ErrSpotTaken):
		return http.StatusConflict, "Spot is already taken", fieldErrors
	case errors.Is(err, domain.ErrSpotUnavailable):
		return http.StatusConflict, "Spot is unavailable", fieldErrors
	case errors.Is(err, domain.ErrDuplicateEmail):
		return http.StatusConflict, "Email already registered", map[string]string{"email": "Email is already in use"}
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "Conflict", fieldErrors
	case errors.Is(err, domain.ErrCapacityBelowActive):
		return http.StatusConflict, "Total capacity cannot be less than active reservations", map[string]string{
			"total_capacity": "Cannot be less than the number of active reservations",
		}
	case errors.Is(err, domain.ErrZoneHasActiveReservations):
		return http.StatusConflict, "Zone has active reservations", map[string]string{
			"zone": "Cannot delete a zone with active reservations",
		}
	default:
		return http.StatusInternalServerError, "Internal server error", fieldErrors
	}
}
