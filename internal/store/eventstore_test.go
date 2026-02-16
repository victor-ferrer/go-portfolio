package store

import (
	"context"
	"strconv"
	"testing"
	"time"

	"go-portfolio/internal/domain"
)

func TestUniquenessKeyComputation(t *testing.T) {
	date := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	key1 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.5, 5000.0)
	key2 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.5, 5000.0)
	key3 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.6, 5000.0)

	if key1 != key2 {
		t.Errorf("same inputs should produce same key, got %s and %s", key1, key2)
	}

	if key1 == key3 {
		t.Errorf("different quantity should produce different key")
	}
}

func TestEventStoreAppendAndRetrieve(t *testing.T) {
	// Skip if CGO not enabled
	if isSkipCGOTests() {
		t.Skip("Skipping SQLite tests: CGO_ENABLED=0")
	}

	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	tx := domain.Transaction{
		ID:         "tx-1",
		Amount:     5000.0,
		Quantity:   100.0,
		Type:       "buy",
		Category:   "Trade",
		Instrument: "AAPL",
		Currency:   "USD",
		CreatedAt:  time.Now(),
	}

	event := domain.Event{
		AggregateID:   "AAPL",
		Type:          "TransactionImported",
		Broker:        "ClickTrade",
		ImportedAt:    time.Now(),
		Payload:       tx,
		UniquenessKey: domain.ComputeUniquenessKey(tx.CreatedAt, "AAPL", "Trade", tx.Quantity, tx.Amount),
	}

	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	events, err := store.GetEvents(ctx, "AAPL")
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if events[0].Payload.ID != "tx-1" {
		t.Errorf("expected transaction ID tx-1, got %s", events[0].Payload.ID)
	}

	if events[0].Payload.Quantity != 100.0 {
		t.Errorf("expected quantity 100.0, got %f", events[0].Payload.Quantity)
	}
}

func TestEventStoreDeduplication(t *testing.T) {
	// Skip if CGO not enabled
	if isSkipCGOTests() {
		t.Skip("Skipping SQLite tests: CGO_ENABLED=0")
	}

	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	now := time.Now()
	uniquenessKey := domain.ComputeUniquenessKey(now, "AAPL", "Trade", 100.0, 5000.0)

	tx := domain.Transaction{
		ID:         "tx-1",
		Amount:     5000.0,
		Quantity:   100.0,
		Type:       "buy",
		Category:   "Trade",
		Instrument: "AAPL",
		Currency:   "USD",
		CreatedAt:  now,
	}

	event := domain.Event{
		AggregateID:   "AAPL",
		Type:          "TransactionImported",
		Broker:        "ClickTrade",
		ImportedAt:    time.Now(),
		Payload:       tx,
		UniquenessKey: uniquenessKey,
	}

	// First insert should succeed
	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("failed to append event: %v", err)
	}

	// Duplicate insert should not error (idempotent, returns nil)
	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("duplicate insert should not error: %v", err)
	}

	// Verify only one event exists
	events, err := store.GetEvents(ctx, "AAPL")
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event after duplicate insert, got %d", len(events))
	}
}

func TestEventStoreByBroker(t *testing.T) {
	// Skip if CGO not enabled
	if isSkipCGOTests() {
		t.Skip("Skipping SQLite tests: CGO_ENABLED=0")
	}

	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	// Insert events from different brokers
	for i, broker := range []string{"ClickTrade", "Degiro", "ClickTrade"} {
		tx := domain.Transaction{
			ID:         "tx-" + strconv.Itoa(i),
			Amount:     5000.0,
			Quantity:   100.0,
			Type:       "buy",
			Category:   "Trade",
			Instrument: "AAPL",
			Currency:   "USD",
			CreatedAt:  time.Now(),
		}

		event := domain.Event{
			AggregateID:   "AAPL",
			Type:          "TransactionImported",
			Broker:        broker,
			ImportedAt:    time.Now(),
			Payload:       tx,
			UniquenessKey: domain.ComputeUniquenessKey(time.Now(), "AAPL", "Trade", float64(i), 5000.0),
		}

		if err := store.AppendEvent(ctx, event); err != nil {
			t.Fatalf("failed to append event: %v", err)
		}
	}

	events, err := store.GetEventsByBroker(ctx, "ClickTrade")
	if err != nil {
		t.Fatalf("failed to get events by broker: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 ClickTrade events, got %d", len(events))
	}

	for _, event := range events {
		if event.Broker != "ClickTrade" {
			t.Errorf("expected ClickTrade, got %s", event.Broker)
		}
	}
}

func TestEventStoreGetAllEvents(t *testing.T) {
	// Skip if CGO not enabled
	if isSkipCGOTests() {
		t.Skip("Skipping SQLite tests: CGO_ENABLED=0")
	}

	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:", "")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	// Insert 3 events
	for i := 0; i < 3; i++ {
		tx := domain.Transaction{
			ID:         "tx-" + strconv.Itoa(i),
			Amount:     5000.0,
			Quantity:   100.0,
			Type:       "buy",
			Category:   "Trade",
			Instrument: "AAPL",
			Currency:   "USD",
			CreatedAt:  time.Now(),
		}

		event := domain.Event{
			AggregateID:   "AAPL",
			Type:          "TransactionImported",
			Broker:        "ClickTrade",
			ImportedAt:    time.Now(),
			Payload:       tx,
			UniquenessKey: domain.ComputeUniquenessKey(time.Now(), "AAPL", "Trade", float64(i), 5000.0),
		}

		if err := store.AppendEvent(ctx, event); err != nil {
			t.Fatalf("failed to append event: %v", err)
		}
	}

	events, err := store.GetAllEvents(ctx)
	if err != nil {
		t.Fatalf("failed to get all events: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

// isSkipCGOTests checks if CGO is disabled and tests should be skipped.
// This is a best-effort check since we can't reliably detect CGO_ENABLED at runtime
// for the go-sqlite3 driver. If this returns true, tests are skipped.
func isSkipCGOTests() bool {
	// Try to ping a test database - if it fails with CGO message, skip
	// For now, we'll attempt to create a store and check the error
	_, err := NewSQLiteEventStore(":memory:", "")
	if err != nil && err.Error() == "failed to ping database: Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub" {
		return true
	}
	return false
}
