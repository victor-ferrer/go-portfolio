package store

import (
	"context"
	"testing"
	"time"

	"go-portfolio/internal/domain"
)

func TestUniquenessKeyComputation(t *testing.T) {
	date := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	key1 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.5)
	key2 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.5)
	key3 := domain.ComputeUniquenessKey(date, "AAPL", "Trade", 100.6)

	if key1 != key2 {
		t.Errorf("same inputs should produce same key, got %s and %s", key1, key2)
	}

	if key1 == key3 {
		t.Errorf("different quantity should produce different key")
	}
}

func TestEventStoreAppendAndRetrieve(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	tx := domain.Transaction{
		ID:         "tx-1",
		Amount:     100.0,
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
		UniquenessKey: domain.ComputeUniquenessKey(tx.CreatedAt, "AAPL", "Trade", 100.0),
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
}

func TestEventStoreDeduplication(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	uniquenessKey := domain.ComputeUniquenessKey(time.Now(), "AAPL", "Trade", 100.0)

	tx := domain.Transaction{
		ID:         "tx-1",
		Amount:     100.0,
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
	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	// Insert events from different brokers
	for i, broker := range []string{"ClickTrade", "Degiro", "ClickTrade"} {
		tx := domain.Transaction{
			ID:         string(rune(i)),
			Amount:     100.0,
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
			UniquenessKey: domain.ComputeUniquenessKey(time.Now(), "AAPL", "Trade", float64(i)),
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
	ctx := context.Background()
	store, err := NewSQLiteEventStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}
	defer store.Close()

	// Insert 3 events
	for i := 0; i < 3; i++ {
		tx := domain.Transaction{
			ID:         string(rune(i)),
			Amount:     100.0,
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
			UniquenessKey: domain.ComputeUniquenessKey(time.Now(), "AAPL", "Trade", float64(i)),
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
