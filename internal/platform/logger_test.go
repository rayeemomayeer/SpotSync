package platform_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

func TestNewLogger_levels(t *testing.T) {
	tests := []struct {
		level string
	}{
		{level: "debug"},
		{level: "info"},
		{level: "warn"},
		{level: "error"},
		{level: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger := platform.NewLogger(tt.level)
			if logger == nil {
				t.Fatal("NewLogger returned nil")
			}
			logger.Info("test log", "level", tt.level)
		})
	}
}

func TestRunMigrations_noMigrations(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set")
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := platform.RunMigrations(os.Getenv("DATABASE_URL"), "migrations", log); err != nil {
		t.Fatalf("RunMigrations() error = %v", err)
	}
}
