package models_test

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"gorm.io/gorm"
)

func TestMigrationCreatesGradedTables(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := platform.RunMigrations(databaseURL, "", "migrations", log); err != nil {
		if isPostgresUnavailable(err) {
			t.Skipf("postgres not available: %v", err)
		}
		t.Fatalf("RunMigrations() error = %v", err)
	}

	cfg := &config.Config{
		DatabaseURL:       databaseURL,
		DBMaxOpenConns:    5,
		DBMaxIdleConns:    2,
		DBConnMaxLifetime: 5 * time.Minute,
	}

	db, err := platform.OpenPostgres(cfg, nil)
	if err != nil {
		t.Fatalf("OpenPostgres() error = %v", err)
	}

	migrator := db.Migrator()
	for _, table := range []string{"users", "parking_zones", "reservations"} {
		if !migrator.HasTable(table) {
			t.Errorf("expected table %q to exist after migration", table)
		}
	}

	assertColumnExists(t, db, "users", "email")
	assertColumnExists(t, db, "parking_zones", "total_capacity")
	assertColumnExists(t, db, "reservations", "license_plate")
}

func assertColumnExists(t *testing.T, db *gorm.DB, table, column string) {
	t.Helper()
	if !db.Migrator().HasColumn(table, column) {
		t.Errorf("expected column %q on table %q", column, table)
	}
}

func isPostgresUnavailable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "password authentication failed") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connect: connection")
}
