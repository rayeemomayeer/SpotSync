package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
)

func TestHTTPErrorHandler_domainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"not found", domain.ErrNotFound, http.StatusNotFound, "Resource not found"},
		{"unauthorized", domain.ErrUnauthorized, http.StatusUnauthorized, "Unauthorized"},
		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, "Unauthorized"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "Forbidden"},
		{"zone full", domain.ErrZoneFull, http.StatusConflict, "Zone is full"},
		{"duplicate email", domain.ErrDuplicateEmail, http.StatusConflict, "Email already registered"},
		{"internal", errors.New("pq: connection refused"), http.StatusInternalServerError, "Internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler.HTTPErrorHandler(tt.err, c)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body dto.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Success {
				t.Error("success should be false")
			}
			if body.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", body.Message, tt.wantMsg)
			}
		})
	}
}

func TestHTTPErrorHandler_validationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := domain.NewValidationError("Validation failed", map[string]string{
		"email": "Must be a valid email address",
	})
	handler.HTTPErrorHandler(err, c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	var body dto.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Errors["email"] == "" {
		t.Error("expected email field error")
	}
}

func TestBindAndValidate_registerRequest(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()

	body := `{"name":"John","email":"not-an-email","password":"short","role":"driver"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var payload dto.RegisterRequest
	err := handler.BindAndValidate(c, &payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error type = %T, want *domain.ValidationError", err)
	}
	if validationErr.Fields["email"] == "" {
		t.Error("expected email validation error")
	}
	if validationErr.Fields["password"] == "" {
		t.Error("expected password validation error")
	}
}

func TestJSONSuccess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.JSONSuccess(c, http.StatusCreated, "Created", map[string]string{"id": "1"}); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}

	var body dto.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Error("success should be true")
	}
}
