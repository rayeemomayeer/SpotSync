// Package main is the SpotSync API entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/app"
	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
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

	e := app.NewEcho(cfg, db, log, app.Options{
		AuthRateLimitPerMinute: 20,
		EnableRequestLogger:    true,
	})

	addr := ":" + cfg.Port
	log.Info("starting api server", "addr", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Info("server stopped")
	return nil
}
