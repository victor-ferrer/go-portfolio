package parsers

import (
	"context"
	"fmt"
	"io"
	"time"

	"go-portfolio/internal/domain"
	"go-portfolio/internal/store"
)

// Parser interface for different broker parsers.
type Parser interface {
	Parse(reader io.Reader) ([]domain.Transaction, error)
}

// ParseAndStore reads transactions from a parser and stores them as events in the event store.
// It handles deduplication via uniqueness keys computed from transaction attributes.
func ParseAndStore(ctx context.Context, brokerName string, parser Parser, data io.Reader, eventStore store.EventStore) error {
	transactions, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse transactions: %w", err)
	}

	importedAt := time.Now()

	for _, tx := range transactions {
		// Validate required fields for event creation
		if tx.Instrument == "" {
			return fmt.Errorf("transaction %s missing required Instrument field", tx.ID)
		}

		// Compute uniqueness key from date + instrument + category + quantity + amount
		uniquenessKey := domain.ComputeUniquenessKey(tx.CreatedAt, tx.Instrument, tx.Category, tx.Quantity, tx.Amount)

		event := domain.Event{
			AggregateID:   tx.Instrument,
			Type:          "TransactionImported",
			Broker:        brokerName,
			ImportedAt:    importedAt,
			Payload:       tx,
			UniquenessKey: uniquenessKey,
		}

		if err := eventStore.AppendEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to append event for transaction %s: %w", tx.ID, err)
		}
	}

	return nil
}
