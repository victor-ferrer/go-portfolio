package store

import (
	"context"
	"database/sql"

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

// NewEventStore creates a new event store using the provided PostgreSQL database connection.
func NewEventStore(db *sql.DB) (EventStore, error) {
	return NewPostgreSQLEventStore(db)
}
