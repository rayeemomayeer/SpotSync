//go:build integration

package testutil

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/app"
	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupPostgres(t *testing.T) (*gorm.DB, string) {
	t.Helper()

	if os.Getenv("SKIP_TESTCONTAINERS") == "true" {
		t.Skip("SKIP_TESTCONTAINERS=true")
	}

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("spotsync_test"),
		postgres.WithUsername("spotsync"),
		postgres.WithPassword("spotsync"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate postgres container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	if err := platform.RunMigrations(connStr, "migrations", nil); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, connStr
}

func TestConfig(databaseURL string) *config.Config {
	return &config.Config{
		Port:                       "8080",
		DatabaseURL:                databaseURL,
		JWTSecret:                  "integration-test-secret",
		JWTExpiry:                  time.Hour,
		BcryptCost:                 10,
		AllowSelfAdminRegistration: true,
		LogLevel:                   "error",
		DBMaxOpenConns:             10,
		DBMaxIdleConns:             5,
		DBConnMaxLifetime:          5 * time.Minute,
		MigrateOnStartup:           false,
		MigrationsPath:             "migrations",
	}
}

func NewTestServer(t *testing.T, db *gorm.DB, databaseURL string) *httptest.Server {
	t.Helper()

	cfg := TestConfig(databaseURL)
	log := platform.NewLogger(cfg.LogLevel)
	e := app.NewEcho(cfg, db, log, app.Options{
		AuthRateLimitPerMinute: 1000,
		EnableRequestLogger:    false,
	})

	return httptest.NewServer(e)
}

func UniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@test.com", prefix, time.Now().UnixNano())
}
