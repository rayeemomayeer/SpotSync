package domain

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrZoneFull           = errors.New("zone is full")
	ErrConflict           = errors.New("conflict")
	ErrDuplicateEmail     = errors.New("email already registered")
	ErrNotOwner           = errors.New("not owner")
	ErrSpotTaken                 = errors.New("spot already taken")
	ErrSpotUnavailable           = errors.New("spot unavailable")
	ErrCapacityBelowActive       = errors.New("total capacity below active reservations")
	ErrZoneHasActiveReservations = errors.New("zone has active reservations")
	ErrRateLimited               = errors.New("rate limit exceeded")
)

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "validation failed"
}

func NewValidationError(message string, fields map[string]string) *ValidationError {
	return &ValidationError{Message: message, Fields: fields}
}
