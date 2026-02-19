package repository

import (
	"context"
	"time"

	"go-portfolio/internal/domain"
	"go-portfolio/internal/store"
)

// TransactionFilter holds optional filters for querying transactions.
type TransactionFilter struct {
	Instrument string
	Broker     string
	Type       string
	From       *time.Time
	To         *time.Time
}

// TransactionRepository defines the interface for retrieving transactions.
type TransactionRepository interface {
	GetTransactions(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error)
}

// EventStoreTransactionRepository implements TransactionRepository using an EventStore.
type EventStoreTransactionRepository struct {
	store store.EventStore
}

// NewTransactionRepository creates a new TransactionRepository backed by the given EventStore.
func NewTransactionRepository(s store.EventStore) TransactionRepository {
	return &EventStoreTransactionRepository{store: s}
}

// GetTransactions retrieves transactions from the event store, applying the provided filters.
func (r *EventStoreTransactionRepository) GetTransactions(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error) {
	var events []domain.Event
	var err error

	if filter.Instrument != "" {
		events, err = r.store.GetEvents(ctx, filter.Instrument)
	} else if filter.Broker != "" {
		events, err = r.store.GetEventsByBroker(ctx, filter.Broker)
	} else {
		events, err = r.store.GetAllEvents(ctx)
	}

	if err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(events))
	for _, e := range events {
		if e.Type != "TransactionImported" {
			continue
		}

		tx := e.Payload

		// Apply broker filter when fetching all events
		if filter.Broker != "" && e.Broker != filter.Broker {
			continue
		}

		// Apply type filter
		if filter.Type != "" && tx.Type != filter.Type {
			continue
		}

		// Apply date range filters
		if filter.From != nil && tx.CreatedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && tx.CreatedAt.After(*filter.To) {
			continue
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}
