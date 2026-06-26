package platform

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rayeemomayeer/SpotSync/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenPostgres connects to PostgreSQL via GORM and applies connection-pool tuning.
func OpenPostgres(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres sql handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if log != nil {
		log.Info("postgres connected",
			"max_open_conns", cfg.DBMaxOpenConns,
			"max_idle_conns", cfg.DBMaxIdleConns,
			"conn_max_lifetime", cfg.DBConnMaxLifetime.String(),
		)
	}

	return db, nil
}
