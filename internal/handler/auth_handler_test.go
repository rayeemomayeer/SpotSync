package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
)

type stubAuthService struct {
	registerFn func(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	loginFn    func(ctx context.Context, req dto.LoginRequest) (dto.LoginData, error)
}

func (s *stubAuthService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	return s.registerFn(ctx, req)
}

func (s *stubAuthService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginData, error) {
	return s.loginFn(ctx, req)
}

func TestAuthHandler_Register(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()

	auth := &stubAuthService{
		registerFn: func(_ context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
			return dto.UserResponse{ID: 1, Name: req.Name, Email: req.Email, Role: req.Role}, nil
		},
	}
	h := handler.NewAuthHandler(auth)

	body := `{"name":"John","email":"john@test.com","password":"password123","role":"driver"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}

	var resp dto.SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Error("success should be true")
	}
}

func TestAuthHandler_LoginUnauthorized(t *testing.T) {
	e := echo.New()
	e.Validator = handler.NewValidator()
	e.HTTPErrorHandler = handler.HTTPErrorHandler

	auth := &stubAuthService{
		loginFn: func(_ context.Context, _ dto.LoginRequest) (dto.LoginData, error) {
			return dto.LoginData{}, domain.ErrInvalidCredentials
		},
	}
	h := handler.NewAuthHandler(auth)

	body := `{"email":"john@test.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		handler.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
