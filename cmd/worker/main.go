// Package main is the SpotSync release worker entrypoint.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/outbox"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"github.com/rayeemomayeer/SpotSync/internal/service"
	"github.com/rayeemomayeer/SpotSync/internal/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker fatal", "error", err)
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

	ctxRoot := context.Background()
	shutdownTracer, err := platform.InitTracer(ctxRoot, "spotsync-worker", log)
	if err != nil {
		return fmt.Errorf("init tracer: %w", err)
	}
	defer func() {
		_ = shutdownTracer(context.Background())
	}()

	db, err := platform.OpenPostgres(cfg, log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if cfg.MigrateOnStartup {
		if err := platform.RunMigrations(cfg.DatabaseURL, cfg.DatabaseMigrateURL, cfg.MigrationsPath, log); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
	}

	redisClient, err := platform.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Warn("redis unavailable, relay publishes noop", "error", err)
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	outboxRepo := outbox.NewRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	var publisher worker.EventPublisher = worker.NoopPublisher{}
	if redisClient != nil {
		publisher = worker.NewRedisPublisher(redisClient, notifRepo)
	}

	relay := worker.NewRelay(outboxRepo, publisher, log)
	expiry := worker.NewExpiryEngine(db, outboxRepo, log)
	demoSvc := service.NewDemoService(repository.NewDemoRepository(db))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.RunRelayLoop(ctx, relay, 2*time.Second, log)
	go worker.RunExpiryLoop(ctx, expiry, 30*time.Second, log)
	go runDemoResetLoop(ctx, demoSvc, log)

	log.Info("worker started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("worker shutting down")
	cancel()
	time.Sleep(500 * time.Millisecond)
	return nil
}

func runDemoResetLoop(ctx context.Context, demoSvc *service.DemoService, log *slog.Logger) {
	interval := 6 * time.Hour
	if raw := os.Getenv("DEMO_RESET_INTERVAL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d >= time.Minute {
			interval = d
		}
	}
	ttl := 24 * time.Hour
	if raw := os.Getenv("DEMO_SESSION_INACTIVE_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			ttl = d
		}
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, stats, err := demoSvc.PurgeStaleSessions(ctx, time.Now().Add(-ttl))
			if err != nil && log != nil {
				log.Error("demo reset sweep failed", "error", err)
				continue
			}
			if count > 0 && log != nil {
				log.Info("demo reset sweep", "sessions", count, "deleted_reservations", stats.ReservationsDeleted)
			}
		}
	}
}
