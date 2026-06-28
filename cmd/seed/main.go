package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"github.com/rayeemomayeer/SpotSync/internal/models"
	"github.com/rayeemomayeer/SpotSync/internal/platform"
	"github.com/rayeemomayeer/SpotSync/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SEED_ADMIN_NAME"))
	if name == "" {
		name = "Admin"
	}

	if email == "" || password == "" {
		return fmt.Errorf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required")
	}

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

	if err := platform.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath, log); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	existing, err := userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		log.Info("admin already exists", "email", email)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     models.RoleAdmin,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		return err
	}

	log.Info("admin user created", "email", email, "id", user.ID)
	return nil
}
