package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/middleware"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginData, error)
	Me(ctx context.Context, userID uint) (dto.UserResponse, error)
}

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.auth.Register(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusCreated, "User registered successfully", user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	data, err := h.auth.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "Login successful", data)
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}

	user, err := h.auth.Me(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return JSONSuccess(c, http.StatusOK, "User retrieved successfully", user)
}
