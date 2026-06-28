package domain

import "errors"

// Sentinel domain errors mapped to HTTP status codes at the handler edge.
var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrZoneFull           = errors.New("zone is full")
	ErrConflict           = errors.New("conflict")
	ErrDuplicateEmail     = errors.New("email already registered")
)

// ValidationError carries field-level validation messages for 400 responses.
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

// NewValidationError builds a field-level validation error.
func NewValidationError(message string, fields map[string]string) *ValidationError {
	return &ValidationError{Message: message, Fields: fields}
}
