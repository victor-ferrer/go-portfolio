package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"go-portfolio/infrastructure/database"
	"go-portfolio/internal/domain"
)

// setupTestDB creates a test database connection.
// It uses DATABASE_DSN environment variable to connect to the PostgreSQL database.
// If DATABASE_DSN is not set, tests are skipped.
func setupTestDB(t *testing.T) *PostgreSQLEventStore {
	t.Helper()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN environment variable not set. Run 'docker-compose up -d' and set DATABASE_DSN to run tests.")
	}

	migrationsPath := getMigrationsPath(t)
	db, err := database.Connect(dsn, migrationsPath)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	store, err := NewPostgreSQLEventStore(db)
	if err != nil {
		t.Fatalf("failed to create event store: %v", err)
	}

	t.Cleanup(func() { store.Close() })

	// Clean up events table before each test
	_, err = store.db.Exec("DELETE FROM events")
	if err != nil {
		t.Fatalf("failed to clean up events table: %v", err)
	}

	return store
}

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
	store := setupTestDB(t)

	ctx := context.Background()

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
	store := setupTestDB(t)

	ctx := context.Background()

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
	store := setupTestDB(t)

	ctx := context.Background()

	// Insert events from different brokers
	for i, broker := range []string{"ClickTrade", "Degiro", "ClickTrade"} {
		tx := domain.Transaction{
			ID:         fmt.Sprintf("tx-%d", i),
			Amount:     5000.0,
			Quantity:   100.0,
			Type:       "buy",
			Category:   "Trade",
			Instrument: "AAPL",
			Currency:   "USD",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Millisecond),
		}

		event := domain.Event{
			AggregateID:   "AAPL",
			Type:          "TransactionImported",
			Broker:        broker,
			ImportedAt:    time.Now(),
			Payload:       tx,
			UniquenessKey: domain.ComputeUniquenessKey(tx.CreatedAt, "AAPL", "Trade", float64(i), 5000.0),
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
	store := setupTestDB(t)

	ctx := context.Background()

	// Insert 3 events
	for i := 0; i < 3; i++ {
		tx := domain.Transaction{
			ID:         fmt.Sprintf("tx-%d", i),
			Amount:     5000.0,
			Quantity:   100.0,
			Type:       "buy",
			Category:   "Trade",
			Instrument: "AAPL",
			Currency:   "USD",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Millisecond),
		}

		event := domain.Event{
			AggregateID:   "AAPL",
			Type:          "TransactionImported",
			Broker:        "ClickTrade",
			ImportedAt:    time.Now(),
			Payload:       tx,
			UniquenessKey: domain.ComputeUniquenessKey(tx.CreatedAt, "AAPL", "Trade", float64(i), 5000.0),
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

// getMigrationsPath returns the file:// URL path to the migrations directory.
func getMigrationsPath(t *testing.T) string {
	// Get the path of this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get current file path")
	}

	// Navigate from internal/store/postgres_eventstore_test.go to infrastructure/database/migrations/
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "database", "migrations")
	absPath, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("could not resolve migrations path: %v", err)
	}

	// Verify migrations directory exists
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("migrations directory not found at %s: %v", absPath, err)
	}

	// Convert to file:// URL format (file:///C:/path on Windows, file:///path on Unix)
	slashPath := filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" && len(slashPath) > 2 && slashPath[1] == ':' {
		return "file:///" + slashPath
	}
	return "file://" + slashPath
}
