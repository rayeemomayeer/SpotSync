package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
	"github.com/rayeemomayeer/SpotSync/internal/middleware"
	"github.com/rayeemomayeer/SpotSync/internal/models"
)

type stubVerifier struct {
	userID uint
	role   string
	err    error
}

func (s stubVerifier) Verify(string) (uint, string, error) {
	return s.userID, s.role, s.err
}

func TestJWTAuth(t *testing.T) {
	e := echo.New()
	e.Use(middleware.JWTAuth(stubVerifier{userID: 5, role: models.RoleDriver}))
	e.GET("/", func(c echo.Context) error {
		id, ok := middleware.UserID(c)
		if !ok || id != 5 {
			t.Fatalf("user id = %d, ok = %v", id, ok)
		}
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestJWTAuth_missingToken(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handler.HTTPErrorHandler
	e.Use(middleware.JWTAuth(stubVerifier{}))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAdmin_forbidden(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handler.HTTPErrorHandler
	e.Use(middleware.JWTAuth(stubVerifier{userID: 1, role: models.RoleDriver}))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, middleware.RequireAdmin())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireAdmin_allowed(t *testing.T) {
	e := echo.New()
	e.Use(middleware.JWTAuth(stubVerifier{userID: 1, role: models.RoleAdmin}))
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, middleware.RequireAdmin())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRequireAdmin_noRole(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = handler.HTTPErrorHandler
	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }, middleware.RequireAdmin())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
