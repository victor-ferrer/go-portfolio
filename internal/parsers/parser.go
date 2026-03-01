package parsers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go-portfolio/internal/domain"
	"go-portfolio/internal/store"
)

// ParseResult holds the results of parsing: successfully parsed transactions,
// the CSV header row, and any raw rows that failed validation.
type ParseResult struct {
	Header       []string
	Transactions []domain.Transaction
	FailedRows   [][]string
}

// Parser interface for different broker parsers.
type Parser interface {
	Parse(reader io.Reader) (ParseResult, error)
}

// ParseAndStore reads transactions from a parser and stores them as events in the event store.
// It handles deduplication via uniqueness keys computed from transaction attributes.
// If failedFilePath is non-empty, any invalid transactions are written to that file in CSV format.
func ParseAndStore(ctx context.Context, brokerName string, parser Parser, data io.Reader, eventStore store.EventStore, failedFilePath string) error {
	result, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse transactions: %w", err)
	}

	if len(result.FailedRows) > 0 {
		log.Printf("warning: %d invalid transaction(s) skipped", len(result.FailedRows))
		if failedFilePath != "" {
			if writeErr := writeFailedCSV(failedFilePath, result.Header, result.FailedRows); writeErr != nil {
				log.Printf("warning: failed to write failed transactions to %s: %v", failedFilePath, writeErr)
			} else {
				log.Printf("invalid transactions written to %s", failedFilePath)
			}
		}
	}

	importedAt := time.Now()

	for _, tx := range result.Transactions {
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

// writeFailedCSV writes the header and failed rows to a CSV file at the given path.
func writeFailedCSV(filePath string, header []string, rows [][]string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create failed file %s: %w", filePath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if err := w.WriteAll(rows); err != nil {
		return fmt.Errorf("failed to write rows: %w", err)
	}
	w.Flush()
	return w.Error()
}
