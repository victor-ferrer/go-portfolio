package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go-portfolio/internal/domain"
)

// PostgreSQLEventStore implements EventStore using PostgreSQL.
type PostgreSQLEventStore struct {
	db *sql.DB
}

// NewPostgreSQLEventStore creates a new PostgreSQL event store using an existing database connection.
func NewPostgreSQLEventStore(db *sql.DB) (*PostgreSQLEventStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db must not be nil")
	}
	return &PostgreSQLEventStore{db: db}, nil
}

// AppendEvent appends an event to the store.
// On duplicate uniqueness key, logs a warning and returns nil (idempotent).
func (s *PostgreSQLEventStore) AppendEvent(ctx context.Context, event domain.Event) error {
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
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
		if isPostgresUniqueConstraintError(err) {
			log.Printf("warning: duplicate event skipped (instrument=%s, date=%s, category=%s, broker=%s)",
				event.AggregateID,
				event.Payload.CreatedAt.Format("2006-01-02"),
				event.Payload.Category,
				event.Broker,
			)
			return nil
		}
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for a given aggregate (instrument).
func (s *PostgreSQLEventStore) GetEvents(ctx context.Context, aggregateID string) ([]domain.Event, error) {
	query := `
	SELECT id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key
	FROM events
	WHERE aggregate_id = $1
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
func (s *PostgreSQLEventStore) GetEventsByBroker(ctx context.Context, broker string) ([]domain.Event, error) {
	query := `
	SELECT id, aggregate_id, type, broker, imported_at, payload, created_at, uniqueness_key
	FROM events
	WHERE broker = $1
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
func (s *PostgreSQLEventStore) GetAllEvents(ctx context.Context) ([]domain.Event, error) {
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
func (s *PostgreSQLEventStore) scanEvents(rows *sql.Rows) ([]domain.Event, error) {
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
func (s *PostgreSQLEventStore) Close() error {
	return s.db.Close()
}

// isPostgresUniqueConstraintError checks if an error is a UNIQUE constraint violation.
func isPostgresUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// Check if error is a PostgreSQL error with unique constraint violation code
	if pqErr, ok := err.(*pq.Error); ok {
		// Code 23505 is unique_violation in PostgreSQL
		return pqErr.Code == "23505" && pqErr.Constraint == "events_uniqueness_key_key"
	}
	return false
}
