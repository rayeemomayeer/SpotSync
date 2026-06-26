package platform

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending up migrations from migrationsPath against databaseURL.
func RunMigrations(databaseURL, migrationsPath string, log *slog.Logger) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	sourceURL := migrationSourceURL(absPath)
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if log != nil && (srcErr != nil || dbErr != nil) {
			log.Warn("close migrator", "source_err", srcErr, "db_err", dbErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	if log != nil {
		log.Info("database migrations applied", "path", absPath)
	}

	return nil
}

func migrationSourceURL(absPath string) string {
	path := filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" {
		return "file:///" + path
	}
	return "file://" + path
}
