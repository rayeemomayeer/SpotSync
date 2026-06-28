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
