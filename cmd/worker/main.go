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

	db, err := platform.OpenPostgres(cfg, log)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if cfg.MigrateOnStartup {
		if err := platform.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath, log); err != nil {
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
	var publisher worker.EventPublisher = worker.NoopPublisher{}
	if redisClient != nil {
		publisher = worker.NewRedisPublisher(redisClient)
	}

	relay := worker.NewRelay(outboxRepo, publisher, log)
	expiry := worker.NewExpiryEngine(db, outboxRepo, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.RunRelayLoop(ctx, relay, 2*time.Second, log)
	go worker.RunExpiryLoop(ctx, expiry, 30*time.Second, log)

	log.Info("worker started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("worker shutting down")
	cancel()
	time.Sleep(500 * time.Millisecond)
	return nil
}
