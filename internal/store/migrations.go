package store

import (
	"context"
	"fmt"
)

// Migration represents a database schema migration.
type Migration struct {
	Name string
	SQL  string
}

// migrations is the ordered list of all schema migrations.
var migrations = []Migration{
	{
		Name: "001_create_events_table",
		SQL: `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			aggregate_id TEXT NOT NULL,
			type TEXT NOT NULL,
			broker TEXT NOT NULL,
			imported_at TIMESTAMP NOT NULL,
			payload TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			uniqueness_key TEXT NOT NULL UNIQUE
		);
		CREATE INDEX IF NOT EXISTS idx_aggregate_id ON events(aggregate_id);
		CREATE INDEX IF NOT EXISTS idx_broker ON events(broker);
		CREATE INDEX IF NOT EXISTS idx_created_at ON events(created_at);
		`,
	},
}

// RunMigrations applies all pending database migrations.
func RunMigrations(ctx context.Context, store *SQLiteEventStore) error {
	for _, migration := range migrations {
		if err := runMigration(ctx, store, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
	}
	return nil
}

// runMigration executes a single migration.
func runMigration(ctx context.Context, store *SQLiteEventStore, migration Migration) error {
	_, err := store.db.ExecContext(ctx, migration.SQL)
	return err
}
