package parsers

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"testing"

	"go-portfolio/internal/domain"
)

// mockEventStore is a simple in-memory event store for testing.
type mockEventStore struct {
	events []domain.Event
}

func (m *mockEventStore) AppendEvent(_ context.Context, event domain.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventStore) GetEvents(_ context.Context, _ string) ([]domain.Event, error) {
	return m.events, nil
}

func (m *mockEventStore) GetEventsByBroker(_ context.Context, _ string) ([]domain.Event, error) {
	return m.events, nil
}

func (m *mockEventStore) GetAllEvents(_ context.Context) ([]domain.Event, error) {
	return m.events, nil
}

func (m *mockEventStore) Close() error { return nil }

// mockParser returns a ParseResult with both valid transactions and failed rows.
type mockParser struct {
	result ParseResult
}

func (p *mockParser) Parse(_ io.Reader) (ParseResult, error) {
	return p.result, nil
}

func TestParseAndStore_WritesFailedCSV(t *testing.T) {
	validCSV := `Client ID,Trade Date,Instrument
6594837,06-feb-2026,AAPL`

	failedRow := []string{"6594837", "07-feb-2026", ""} // missing instrument

	mp := &mockParser{
		result: ParseResult{
			Header:       []string{"Client ID", "Trade Date", "Instrument"},
			Transactions: []domain.Transaction{{Instrument: "AAPL"}},
			FailedRows:   [][]string{failedRow},
		},
	}

	ms := &mockEventStore{}

	failedFile, err := os.CreateTemp(t.TempDir(), "failed_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	failedFilePath := failedFile.Name()
	failedFile.Close()

	ctx := context.Background()
	if err := ParseAndStore(ctx, "test-broker", mp, strings.NewReader(validCSV), ms, failedFilePath); err != nil {
		t.Fatalf("ParseAndStore returned error: %v", err)
	}

	// Verify that the event store received the valid transaction
	if len(ms.events) != 1 {
		t.Errorf("expected 1 event stored, got %d", len(ms.events))
	}

	// Verify the failed CSV file was written with header + failed row
	f, err := os.Open(failedFilePath)
	if err != nil {
		t.Fatalf("failed to open failed file: %v", err)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("failed to read failed CSV: %v", err)
	}

	if len(records) != 2 { // header + 1 failed row
		t.Errorf("expected 2 records in failed CSV (header + 1 row), got %d", len(records))
	}

	if strings.Join(records[0], ",") != "Client ID,Trade Date,Instrument" {
		t.Errorf("unexpected header in failed CSV: %v", records[0])
	}

	if strings.Join(records[1], ",") != strings.Join(failedRow, ",") {
		t.Errorf("unexpected failed row in CSV: %v", records[1])
	}
}

func TestParseAndStore_NoFailedFile_WhenNoFailedRows(t *testing.T) {
	mp := &mockParser{
		result: ParseResult{
			Header:       []string{"Client ID", "Trade Date", "Instrument"},
			Transactions: []domain.Transaction{{Instrument: "AAPL"}},
			FailedRows:   nil,
		},
	}

	ms := &mockEventStore{}

	failedFilePath := t.TempDir() + "/should_not_exist_failed.csv"

	ctx := context.Background()
	if err := ParseAndStore(ctx, "test-broker", mp, strings.NewReader(""), ms, failedFilePath); err != nil {
		t.Fatalf("ParseAndStore returned error: %v", err)
	}

	// The failed file should NOT be created when there are no failed rows
	if _, err := os.Stat(failedFilePath); !os.IsNotExist(err) {
		t.Errorf("expected failed file to not exist when there are no failed rows, but it does")
	}
}
