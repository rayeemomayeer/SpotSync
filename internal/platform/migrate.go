package platform

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	appmigrations "github.com/rayeemomayeer/SpotSync/migrations"
)

// RunMigrations applies all pending up migrations embedded in the migrations package.
// migrationsPath is retained for Makefile CLI parity; runtime migrate uses the embedded FS.
func RunMigrations(databaseURL, migrationsPath string, log *slog.Logger) error {
	_ = migrationsPath

	source, err := iofs.New(appmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("open migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
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
		log.Info("database migrations applied")
	}

	return nil
}

// EmbeddedMigrationFiles returns the names of bundled migration SQL files (for tests).
func EmbeddedMigrationFiles() ([]string, error) {
	return fs.Glob(appmigrations.Files, "*.sql")
}
