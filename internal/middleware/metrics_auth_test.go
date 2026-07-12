package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/domain"
	"github.com/rayeemomayeer/SpotSync/internal/middleware"
)

func TestMetricsAuth_emptyTokenAllows(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	h := middleware.MetricsAuth("")(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})
	if err := h(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("handler should run when token unset")
	}
}

func TestMetricsAuth_rejectsMissing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := middleware.MetricsAuth("secret-token")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	err := h(c)
	if err != domain.ErrUnauthorized {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestMetricsAuth_acceptsBearer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer secret-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	h := middleware.MetricsAuth("secret-token")(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})
	if err := h(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected handler")
	}
}

func TestMetricsAuth_acceptsQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics?token=secret-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/metrics")
	c.QueryParams().Set("token", "secret-token")

	called := false
	h := middleware.MetricsAuth("secret-token")(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})
	if err := h(c); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected handler")
	}
}
