package service

import (
	"context"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
}

type AuthService struct {
	users                  UserStore
	tokens                 *platform.TokenManager
	bcryptCost             int
	allowSelfAdminRegister bool
}

func NewAuthService(users UserStore, tokens *platform.TokenManager, bcryptCost int, allowSelfAdminRegister bool) *AuthService {
	return &AuthService{
		users:                  users,
		tokens:                 tokens,
		bcryptCost:             bcryptCost,
		allowSelfAdminRegister: allowSelfAdminRegister,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return dto.UserResponse{}, err
	}

	user := &models.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    normalizeEmail(req.Email),
		Password: string(hash),
		Role:     resolveRegistrationRole(req.Role, s.allowSelfAdminRegister),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserFromModel(*user), nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (dto.LoginData, error) {
	user, err := s.users.FindByEmail(ctx, normalizeEmail(req.Email))
	if err != nil {
		return dto.LoginData{}, err
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return dto.LoginData{}, domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user.ID, user.Role)
	if err != nil {
		return dto.LoginData{}, err
	}

	return dto.LoginData{
		Token: token,
		User:  dto.UserFromModel(*user),
	}, nil
}

func (s *AuthService) Me(ctx context.Context, userID uint) (dto.UserResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return dto.UserResponse{}, err
	}
	if user == nil {
		return dto.UserResponse{}, domain.ErrUnauthorized
	}
	return dto.UserFromModel(*user), nil
}

func resolveRegistrationRole(requested string, allowSelfAdmin bool) string {
	if allowSelfAdmin {
		return requested
	}
	return models.RoleDriver
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
