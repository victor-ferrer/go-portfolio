package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Connect opens a PostgreSQL connection using the provided DSN, verifies connectivity,
// and runs any pending migrations from migrationsPath.
// migrationsPath should be a file:// URL pointing to the migrations directory
// (e.g., "file://./migrations"). If migrationsPath is empty, migrations are skipped.
func Connect(dsn, migrationsPath string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if migrationsPath != "" {
		if err := runMigrations(db, migrationsPath); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return db, nil
}

// runMigrations applies all pending migrations from migrationsPath to the database.
func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
