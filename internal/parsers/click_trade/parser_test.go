package click_trade

import (
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	file, err := os.Open("testdata.csv")
	if err != nil {
		t.Fatalf("failed to open testdata.csv: %v", err)
	}
	defer file.Close()

	parser := NewParser()
	transactions, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]

	// Test ID
	if tx.ID != "" {
		t.Errorf("expected empty ID (Trade ID was empty), got %q", tx.ID)
	}

	// Test Amount
	if tx.Amount != 17.28 {
		t.Errorf("expected amount 17.28, got %f", tx.Amount)
	}

	// Test Currency
	if tx.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %q", tx.Currency)
	}

	// Test Instrument
	if tx.Instrument != "*Delisted 20260127 (Iberdrola SA - Rights)" {
		t.Errorf("expected instrument '*Delisted 20260127 (Iberdrola SA - Rights)', got %q", tx.Instrument)
	}

	// Test ISIN
	if tx.ISIN != "ES06445809V1" {
		t.Errorf("expected ISIN ES06445809V1, got %q", tx.ISIN)
	}

	// Test Type
	if tx.Type != "Rights" {
		t.Errorf("expected type Rights, got %q", tx.Type)
	}

	// Test Category
	if tx.Category != "Exchange" {
		t.Errorf("expected category Exchange, got %q", tx.Category)
	}

	// Test Description
	if tx.Description != "IBE_D:xmce" {
		t.Errorf("expected description IBE_D:xmce, got %q", tx.Description)
	}

	// Test CreatedAt
	expectedTime, _ := time.Parse("02-Jan-2006", "06-feb-2026")
	if !tx.CreatedAt.Equal(expectedTime) {
		t.Errorf("expected CreatedAt %v, got %v", expectedTime, tx.CreatedAt)
	}
}
