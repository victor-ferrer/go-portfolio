package click_trade

import (
	"os"
	"strings"
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
	result, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result.Transactions))
	}

	if len(result.FailedRows) != 0 {
		t.Errorf("expected 0 failed rows, got %d", len(result.FailedRows))
	}

	tx := result.Transactions[0]

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

func TestParseInvalidTransactions(t *testing.T) {
	// CSV with one valid row and one row missing the Instrument field
	csvData := `Client ID,Trade Date,Value Date,Account ID,Trade ID,Position ID,Corporate Action Id,Bk_Record_ Id,Booking Id,Transaction Type,Event,Booked Amount,Currency,Conversion Rate,Conversion cost,Total cost,Realized P/L,IBAN,IBAN owner name,Comment,Correction reason,Instrument,Instrument Symbol,Instrument ISIN,Instrument currency,Type,Exchange Description,From Derivative,Underlying asset type
6594837,06-feb-2026,06-feb-2026,15500/SLT5953,,,9498707,2744033532,,Corporate action,Exchange,"17,28",EUR,1,0,0,,,,,,*Delisted 20260127 (Iberdrola SA - Rights),IBE_D:xmce,ES06445809V1,EUR,Rights,BME Spanish Exchanges,No,
6594837,07-feb-2026,07-feb-2026,15500/SLT5953,,,9498708,2744033533,,Corporate action,Exchange,"10,00",EUR,1,0,0,,,,,,,EMPTY_INSTRUMENT:xmce,ES00000000V1,EUR,Rights,BME Spanish Exchanges,No,`

	parser := NewParser()
	result, err := parser.Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Transactions) != 1 {
		t.Errorf("expected 1 valid transaction, got %d", len(result.Transactions))
	}

	if len(result.FailedRows) != 1 {
		t.Errorf("expected 1 failed row, got %d", len(result.FailedRows))
	}

	if len(result.Header) == 0 {
		t.Error("expected non-empty header")
	}
}
