package store

import (
	"context"

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

// NewEventStore creates a new event store using PostgreSQL.
// dsn should be a PostgreSQL connection string (e.g., "postgres://user:pass@host:port/dbname?sslmode=disable").
// If dsn is empty, it reads from DATABASE_DSN environment variable.
// migrationsPath should be the directory containing migration files (e.g., "file://./migrations").
func NewEventStore(dsn, migrationsPath string) (EventStore, error) {
	return NewPostgreSQLEventStore(dsn, migrationsPath)
}
