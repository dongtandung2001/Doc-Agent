package db

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs database migrations from the migrations directory
func RunMigrations(postgresURL string) error {
	// Try multiple paths: ./migrations (when running from service root), migrations (relative)
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath)),
		postgresURL,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
