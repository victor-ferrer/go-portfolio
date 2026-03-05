package repository

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"go-portfolio/infrastructure/database"
	"go-portfolio/internal/domain"
)

// setupStockQuotationTestDB creates a test database connection for stock quotation tests.
// It uses DATABASE_DSN environment variable. If not set, tests are skipped.
func setupStockQuotationTestDB(t *testing.T) StockQuotationRepository {
	t.Helper()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN environment variable not set. Run 'docker-compose up -d' and set DATABASE_DSN to run tests.")
	}

	migrationsPath := getStockQuotationMigrationsPath(t)
	db, err := database.Connect(dsn, migrationsPath)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	repo, err := NewStockQuotationRepository(db)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	// Clean up table before each test
	_, err = db.Exec("DELETE FROM stock_quotations")
	if err != nil {
		t.Fatalf("failed to clean up stock_quotations table: %v", err)
	}

	return repo
}

func TestStockQuotationInsert(t *testing.T) {
	repo := setupStockQuotationTestDB(t)
	ctx := context.Background()

	volume := 1000.0
	q := domain.StockQuotation{
		Ticker: "AAPL",
		Price:  150.5,
		Volume: &volume,
	}

	inserted, err := repo.Insert(ctx, q)
	if err != nil {
		t.Fatalf("failed to insert stock quotation: %v", err)
	}

	if inserted.ID == "" {
		t.Error("expected generated ID, got empty string")
	}
	if inserted.Ticker != q.Ticker {
		t.Errorf("expected ticker %s, got %s", q.Ticker, inserted.Ticker)
	}
	if inserted.Price != q.Price {
		t.Errorf("expected price %f, got %f", q.Price, inserted.Price)
	}
	if inserted.Volume == nil || *inserted.Volume != volume {
		t.Errorf("expected volume %f, got %v", volume, inserted.Volume)
	}
}

func TestStockQuotationInsertWithoutVolume(t *testing.T) {
	repo := setupStockQuotationTestDB(t)
	ctx := context.Background()

	q := domain.StockQuotation{
		Ticker: "GOOGL",
		Price:  2800.0,
		Volume: nil,
	}

	inserted, err := repo.Insert(ctx, q)
	if err != nil {
		t.Fatalf("failed to insert stock quotation without volume: %v", err)
	}

	if inserted.ID == "" {
		t.Error("expected generated ID, got empty string")
	}
	if inserted.Volume != nil {
		t.Errorf("expected nil volume, got %v", inserted.Volume)
	}
}

func TestStockQuotationGetByTicker(t *testing.T) {
	repo := setupStockQuotationTestDB(t)
	ctx := context.Background()

	// Insert two quotations for the same ticker and one for a different ticker
	for _, ticker := range []string{"AAPL", "AAPL", "MSFT"} {
		_, err := repo.Insert(ctx, domain.StockQuotation{Ticker: ticker, Price: 100.0})
		if err != nil {
			t.Fatalf("failed to insert stock quotation: %v", err)
		}
	}

	quotations, err := repo.GetByTicker(ctx, "AAPL")
	if err != nil {
		t.Fatalf("failed to get stock quotations: %v", err)
	}

	if len(quotations) != 2 {
		t.Errorf("expected 2 quotations for AAPL, got %d", len(quotations))
	}

	for _, q := range quotations {
		if q.Ticker != "AAPL" {
			t.Errorf("expected ticker AAPL, got %s", q.Ticker)
		}
		if q.ID == "" {
			t.Error("expected non-empty ID")
		}
	}
}

func TestStockQuotationGetByTickerEmpty(t *testing.T) {
	repo := setupStockQuotationTestDB(t)
	ctx := context.Background()

	quotations, err := repo.GetByTicker(ctx, "UNKNOWN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(quotations) != 0 {
		t.Errorf("expected 0 quotations, got %d", len(quotations))
	}
}

// getStockQuotationMigrationsPath returns the file:// URL path to the migrations directory.
func getStockQuotationMigrationsPath(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get current file path")
	}

	// Navigate from internal/repository/ to infrastructure/database/migrations/
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "database", "migrations")
	absPath, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("could not resolve migrations path: %v", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("migrations directory not found at %s: %v", absPath, err)
	}

	slashPath := filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" && len(slashPath) > 2 && slashPath[1] == ':' {
		return "file:///" + slashPath
	}
	return "file://" + slashPath
}
