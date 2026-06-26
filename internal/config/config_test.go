package config_test

import (
	"strings"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/config"
)

func TestLoad_success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRY", "1h")
	t.Setenv("BCRYPT_COST", "12")
	t.Setenv("DB_MAX_OPEN_CONNS", "10")
	t.Setenv("DB_MAX_IDLE_CONNS", "2")
	t.Setenv("DB_CONN_MAX_LIFETIME", "1m")
	t.Setenv("MIGRATE_ON_STARTUP", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should be set")
	}
	if cfg.JWTExpiry.String() != "1h0m0s" {
		t.Errorf("JWTExpiry = %v, want 1h", cfg.JWTExpiry)
	}
	if cfg.DBMaxOpenConns != 10 {
		t.Errorf("DBMaxOpenConns = %d, want 10", cfg.DBMaxOpenConns)
	}
	if cfg.MigrateOnStartup {
		t.Error("MigrateOnStartup should be false")
	}
}

func TestLoad_missingRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error for missing required vars")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("error = %q, want missing required vars message", err)
	}
}

func TestLoad_invalidBcryptCost(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("BCRYPT_COST", "99")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "BCRYPT_COST") {
		t.Fatalf("Load() error = %v, want BCRYPT_COST validation error", err)
	}
}

func TestLoad_idleExceedsOpenConns(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/db")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DB_MAX_OPEN_CONNS", "2")
	t.Setenv("DB_MAX_IDLE_CONNS", "5")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "DB_MAX_IDLE_CONNS") {
		t.Fatalf("Load() error = %v, want idle/open validation error", err)
	}
}
