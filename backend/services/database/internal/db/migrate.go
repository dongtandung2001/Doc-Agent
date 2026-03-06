package db

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs database migrations from the migrations directory
func RunMigrations(postgresURL string) error {
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	var m *migrate.Migrate
	for attempt := 1; attempt <= 10; attempt++ {
		m, err = migrate.New(
			fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath)),
			postgresURL,
		)
		if err == nil {
			break
		}
		log.Printf("Migration attempt %d/10 failed: %v — retrying in 3s", attempt, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
