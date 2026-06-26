package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rayeemomayeer/SpotSync/internal/handler"
)

func TestHealthHandler_Healthz(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewHealthHandler(nil)
	if err := h.Healthz(c); err != nil {
		t.Fatalf("Healthz() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

func TestHealthHandler_Readyz_ready(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	checker := &handler.DBReadinessChecker{
		PingFn: func(ctx context.Context) error { return nil },
	}
	h := handler.NewHealthHandler(checker)

	if err := h.Readyz(c); err != nil {
		t.Fatalf("Readyz() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHealthHandler_Readyz_notReady(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	checker := &handler.DBReadinessChecker{
		PingFn: func(ctx context.Context) error { return errors.New("db down") },
	}
	h := handler.NewHealthHandler(checker)

	if err := h.Readyz(c); err != nil {
		t.Fatalf("Readyz() error = %v", err)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
