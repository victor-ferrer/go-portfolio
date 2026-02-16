package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go-portfolio/internal/domain"
)

// EventStore defines the interface for storing and retrieving events.
type EventStore interface {
	// AppendEvent appends an event to the store. On duplicate uniqueness key, logs warning and returns nil (idempotent).
	AppendEvent(ctx context.Context, event domain.Event) error

	// GetEvents retrieves all events for a given aggregate (instrument).
	GetEvents(ctx context.Context, aggregateID string) ([]domain.Event, error)

	// GetEventsByBroker retrieves all events imported from a specific broker.
	GetEventsByBroker(ctx context.Context, broker string) ([]domain.Event, error)

	// GetAllEvents retrieves all events from the store.
	GetAllEvents(ctx context.Context) ([]domain.Event, error)

	// Close closes the database connection.
	Close() error
}

// SQLiteEventStore implements EventStore using SQLite.
type SQLiteEventStore struct {
	db  *sql.DB
	dsn string
}

// NewSQLiteEventStore creates a new SQLite event store.
// If dsn is ":memory:", it creates an in-memory database.
// migrationsPath should be the directory containing migration files (e.g., "file://./migrations").
// If migrationsPath is empty, uses built-in migrations.
func NewSQLiteEventStore(dsn, migrationsPath string) (*SQLiteEventStore, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &SQLiteEventStore{db: db, dsn: dsn}

	// Run migrations
	if migrationsPath != "" {
		// Use golang-migrate with external migration files
		if err := store.MigrateWithPath(migrationsPath); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		// Use built-in migrations
		if err := RunMigrations(context.Background(), store); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return store, nil
}

// MigrateWithPath runs all pending migrations from the specified path using golang-migrate.
// migrationsPath should be in the format "file://path/to/migrations".
func (s *SQLiteEventStore) MigrateWithPath(migrationsPath string) error {
	m, err := migrate.New(migrationsPath, "sqlite3://"+s.dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// AppendEvent appends an event to the store.
// On duplicate uniqueness key, logs a warning and returns nil (idempotent).
func (s *SQLiteEventStore) AppendEvent(ctx context.Context, event domain.Event) error {
	// Generate ID if not provided
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Set creation time if not provided
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
	INSERT INTO events (id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		event.ID,
		event.AggregateID,
		event.Type,
		event.Broker,
		event.ImportedAt,
		string(payloadJSON),
		event.CreatedAt,
		event.UniquenessKey,
	)

	if err != nil {
		// Check if it's a UNIQUE constraint violation (duplicate)
		if isUniqueConstraintError(err) {
			log.Printf("warning: duplicate event (uniqueness_key=%s), skipping", event.UniquenessKey)
			return nil
		}
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for a given aggregate (instrument).
func (s *SQLiteEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.Event, error) {
	query := `
	SELECT id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key
	FROM events
	WHERE aggregate_id = ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetEventsByBroker retrieves all events imported from a specific broker.
func (s *SQLiteEventStore) GetEventsByBroker(ctx context.Context, broker string) ([]domain.Event, error) {
	query := `
	SELECT id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key
	FROM events
	WHERE broker = ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, broker)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by broker: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// GetAllEvents retrieves all events from the store.
func (s *SQLiteEventStore) GetAllEvents(ctx context.Context) ([]domain.Event, error) {
	query := `
	SELECT id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key
	FROM events
	ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

// scanEvents converts SQL rows to Event structs.
func (s *SQLiteEventStore) scanEvents(rows *sql.Rows) ([]domain.Event, error) {
	var events []domain.Event

	for rows.Next() {
		var event domain.Event
		var payloadJSON string

		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.Type,
			&event.Broker,
			&event.ImportedAt,
			&payloadJSON,
			&event.CreatedAt,
			&event.UniquenessKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Unmarshal payload
		if err := json.Unmarshal([]byte(payloadJSON), &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// Close closes the database connection.
func (s *SQLiteEventStore) Close() error {
	return s.db.Close()
}

// isUniqueConstraintError checks if an error is a UNIQUE constraint violation.
func isUniqueConstraintError(err error) bool {
	return err != nil && (err.Error() == "UNIQUE constraint failed: events.uniqueness_key" ||
		err.Error() == "UNIQUE constraint failed")
}
