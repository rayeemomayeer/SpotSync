package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/dto"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type mockUserStore struct {
	byEmail map[string]*models.User
	nextID  uint
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{byEmail: make(map[string]*models.User), nextID: 1}
}

func (m *mockUserStore) Create(_ context.Context, user *models.User) error {
	if _, exists := m.byEmail[user.Email]; exists {
		return domain.ErrDuplicateEmail
	}
	user.ID = m.nextID
	m.nextID++
	copy := *user
	m.byEmail[user.Email] = &copy
	return nil
}

func (m *mockUserStore) FindByEmail(_ context.Context, email string) (*models.User, error) {
	user, ok := m.byEmail[email]
	if !ok {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func newTestAuthService(store *mockUserStore, allowAdmin bool) *service.AuthService {
	tokens := platform.NewTokenManager("test-secret-key", time.Hour)
	return service.NewAuthService(store, tokens, bcrypt.MinCost, allowAdmin)
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	store := newMockUserStore()
	auth := newTestAuthService(store, true)

	user, err := auth.Register(context.Background(), dto.RegisterRequest{
		Name:     "Jane Doe",
		Email:    "Jane@Example.com",
		Password: "password123",
		Role:     models.RoleDriver,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Email != "jane@example.com" {
		t.Errorf("email = %q, want normalized lowercase", user.Email)
	}
	if user.Role != models.RoleDriver {
		t.Errorf("role = %q", user.Role)
	}

	login, err := auth.Login(context.Background(), dto.LoginRequest{
		Email:    "jane@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if login.Token == "" {
		t.Error("expected token")
	}
	if login.User.ID != user.ID {
		t.Errorf("login user id = %d, want %d", login.User.ID, user.ID)
	}
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	store := newMockUserStore()
	auth := newTestAuthService(store, true)

	req := dto.RegisterRequest{
		Name:     "A",
		Email:    "dup@test.com",
		Password: "password123",
		Role:     models.RoleDriver,
	}
	if _, err := auth.Register(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	_, err := auth.Register(context.Background(), req)
	if !errors.Is(err, domain.ErrDuplicateEmail) {
		t.Fatalf("error = %v, want ErrDuplicateEmail", err)
	}
}

func TestAuthService_LoginInvalidCredentials(t *testing.T) {
	store := newMockUserStore()
	auth := newTestAuthService(store, true)

	_, err := auth.Login(context.Background(), dto.LoginRequest{
		Email:    "missing@test.com",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_AdminRegistrationGate(t *testing.T) {
	tests := []struct {
		name       string
		allowAdmin bool
		wantRole   string
	}{
		{"honors admin when allowed", true, models.RoleAdmin},
		{"forces driver when disallowed", false, models.RoleDriver},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockUserStore()
			auth := newTestAuthService(store, tt.allowAdmin)

			user, err := auth.Register(context.Background(), dto.RegisterRequest{
				Name:     "Admin User",
				Email:    tt.name + "@test.com",
				Password: "password123",
				Role:     models.RoleAdmin,
			})
			if err != nil {
				t.Fatal(err)
			}
			if user.Role != tt.wantRole {
				t.Errorf("role = %q, want %q", user.Role, tt.wantRole)
			}
		})
	}
}
