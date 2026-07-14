// Demo reset cron — purge inactive demo sessions (Render cron / manual).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/internal/service"
)

const defaultInactiveTTL = 24 * time.Hour

func main() {
	if err := run(); err != nil {
		slog.Error("demo-reset failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := platform.NewLogger(cfg.LogLevel)
	slog.SetDefault(log)

	db, err := platform.OpenPostgres(cfg, log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	ttl := defaultInactiveTTL
	if raw := os.Getenv("DEMO_SESSION_INACTIVE_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			ttl = d
		}
	}

	inactiveBefore := time.Now().Add(-ttl)
	demoSvc := service.NewDemoService(repository.NewDemoRepository(db))
	count, stats, err := demoSvc.PurgeStaleSessions(context.Background(), inactiveBefore)
	if err != nil {
		return err
	}

	log.Info("demo sessions purged",
		"sessions", count,
		"reservations_cancelled", stats.ReservationsCancelled,
		"reservations_deleted", stats.ReservationsDeleted,
		"zones_deleted", stats.ZonesDeleted,
		"audit_deleted", stats.AuditDeleted,
		"inactive_before", inactiveBefore.UTC().Format(time.RFC3339),
	)
	return nil
}
