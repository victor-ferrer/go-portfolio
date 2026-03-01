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

func TestParseBuySellEvents(t *testing.T) {
	csvData := `Client ID,Trade Date,Value Date,Account ID,Trade ID,Position ID,Corporate Action Id,Bk_Record_ Id,Booking Id,Transaction Type,Event,Booked Amount,Currency,Conversion Rate,Conversion cost,Total cost,Realized P/L,IBAN,IBAN owner name,Comment,Correction reason,Instrument,Instrument Symbol,Instrument ISIN,Instrument currency,Type,Exchange Description,From Derivative,Underlying asset type
1,01-Jan-2025,01-Jan-2025,ACC1,,,,,,Trade,Buy 70 @ 44.71 USD,"-3129,70",USD,1,0,0,,,,,,AAPL,AAPL:xnas,US0378331005,USD,Stock,NASDAQ,No,
2,02-Jan-2025,02-Jan-2025,ACC1,,,,,,Trade,Sell -1335 @ 3.94 EUR,"5260,90",EUR,1,0,0,,,,,,PHYS,PHYS:arcx,CA71272H4060,EUR,ETF,NYSE Arca,No,
`
	parser := NewParser()
	transactions, err := parser.Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(transactions))
	}

	buy := transactions[0]
	if buy.Category != "Buy" {
		t.Errorf("buy: expected category %q, got %q", "Buy", buy.Category)
	}
	if buy.Quantity != 70 {
		t.Errorf("buy: expected quantity 70, got %f", buy.Quantity)
	}
	if buy.Price != 44.71 {
		t.Errorf("buy: expected price 44.71, got %f", buy.Price)
	}

	sell := transactions[1]
	if sell.Category != "Sell" {
		t.Errorf("sell: expected category %q, got %q", "Sell", sell.Category)
	}
	if sell.Quantity != 1335 {
		t.Errorf("sell: expected quantity 1335 (absolute), got %f", sell.Quantity)
	}
	if sell.Price != 3.94 {
		t.Errorf("sell: expected price 3.94, got %f", sell.Price)
	}
}

func TestParseEventField(t *testing.T) {
	tests := []struct {
		event    string
		category string
		quantity float64
		price    float64
		ok       bool
	}{
		{"Buy 70 @ 44.71 USD", "Buy", 70, 44.71, true},
		{"Sell -1335 @ 3.94 EUR", "Sell", 1335, 3.94, true},
		{"Exchange", "Exchange", 0, 0, false},
		{"Corporate action", "Corporate action", 0, 0, false},
		{"Buy notanumber @ 44.71 USD", "Buy notanumber @ 44.71 USD", 0, 0, false},
		{"Buy 70 @ notanumber USD", "Buy 70 @ notanumber USD", 0, 0, false},
	}

	for _, tt := range tests {
		cat, qty, prc, ok := parseEventField(tt.event)
		if cat != tt.category || qty != tt.quantity || prc != tt.price || ok != tt.ok {
			t.Errorf("parseEventField(%q) = (%q, %f, %f, %v); want (%q, %f, %f, %v)",
				tt.event, cat, qty, prc, ok, tt.category, tt.quantity, tt.price, tt.ok)
		}
	}
}
