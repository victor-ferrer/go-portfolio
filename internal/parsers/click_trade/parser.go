package click_trade

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go-portfolio/internal/domain"
	"go-portfolio/internal/parsers"
)

// Parser handles parsing of Click Trade data.
type Parser struct {
	// Add fields as needed
}

// NewParser creates a new Click Trade parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a CSV file and returns a ParseResult containing valid transactions
// and any raw rows that failed validation (missing required fields).
func (p *Parser) Parse(reader io.Reader) (parsers.ParseResult, error) {
	csvReader := csv.NewReader(reader)

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return parsers.ParseResult{}, fmt.Errorf("failed to read header: %w", err)
	}

	// Create a map of column names to indices
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[strings.TrimSpace(col)] = i
	}

	result := parsers.ParseResult{
		Header: header,
	}

	// Read data rows
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return parsers.ParseResult{}, fmt.Errorf("failed to read row: %w", err)
		}

		tx, err := p.parseRow(record, columnIndex)
		if err != nil {
			// Row failed validation: collect it for the failed output file
			result.FailedRows = append(result.FailedRows, record)
			continue
		}

		result.Transactions = append(result.Transactions, tx)
	}

	return result, nil
}

// parseEventField parses an event string like "Buy 70 @ 44.71 USD" or
// "Sell -1335 @ 3.94 EUR" into its components.
// Returns (category, quantity, price, true) on success, or (eventStr, 0, 0, false)
// when the string doesn't match the structured format.
func parseEventField(event string) (category string, quantity float64, price float64, ok bool) {
	parts := strings.Fields(event)
	// Expected format: <Category> <Quantity> @ <Price> [<Currency>]
	if len(parts) >= 4 && parts[2] == "@" {
		qtyStr := strings.ReplaceAll(parts[1], ",", ".")
		qty, err1 := strconv.ParseFloat(qtyStr, 64)
		priceStr := strings.ReplaceAll(parts[3], ",", ".")
		prc, err2 := strconv.ParseFloat(priceStr, 64)
		if err1 == nil && err2 == nil {
			if qty < 0 {
				qty = -qty
			}
			return parts[0], qty, prc, true
		}
	}
	return event, 0, 0, false
}

// parseRow converts a CSV row into a Transaction.
// Returns an error if the row is missing required fields (e.g., Instrument).
func (p *Parser) parseRow(record []string, columnIndex map[string]int) (domain.Transaction, error) {
	tx := domain.Transaction{}

	// Trade ID as transaction ID
	if idx, ok := columnIndex["Trade ID"]; ok && idx < len(record) {
		tx.ID = strings.TrimSpace(record[idx])
	}

	// Booked Amount
	if idx, ok := columnIndex["Booked Amount"]; ok && idx < len(record) {
		amountStr := strings.TrimSpace(record[idx])
		amountStr = strings.ReplaceAll(amountStr, ",", ".")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err == nil {
			tx.Amount = amount
		}
	}

	// Currency
	if idx, ok := columnIndex["Currency"]; ok && idx < len(record) {
		tx.Currency = strings.TrimSpace(record[idx])
	}

	// Instrument
	if idx, ok := columnIndex["Instrument"]; ok && idx < len(record) {
		tx.Instrument = strings.TrimSpace(record[idx])
	}

	// Validate required fields
	if tx.Instrument == "" {
		return domain.Transaction{}, fmt.Errorf("missing required Instrument field")
	}

	// ISIN
	if idx, ok := columnIndex["Instrument ISIN"]; ok && idx < len(record) {
		tx.ISIN = strings.TrimSpace(record[idx])
	}

	// Type
	if idx, ok := columnIndex["Type"]; ok && idx < len(record) {
		tx.Type = strings.TrimSpace(record[idx])
	}

	// Category, Quantity, and Price from Event field
	// Event field can be a plain string (e.g. "Exchange") or a structured string
	// like "Buy 70 @ 44.71 USD" or "Sell -1335 @ 3.94 EUR".
	if idx, ok := columnIndex["Event"]; ok && idx < len(record) {
		eventStr := strings.TrimSpace(record[idx])
		category, quantity, price, ok := parseEventField(eventStr)
		tx.Category = category
		if ok {
			tx.Quantity = quantity
			tx.Price = price
		}
	}

	// Quantity column (overrides event-derived quantity if present)
	if idx, ok := columnIndex["Quantity"]; ok && idx < len(record) {
		quantityStr := strings.TrimSpace(record[idx])
		quantityStr = strings.ReplaceAll(quantityStr, ",", ".")
		quantity, err := strconv.ParseFloat(quantityStr, 64)
		if err == nil {
			tx.Quantity = quantity
		}
	}

	// Description from Comment or Instrument Symbol
	if idx, ok := columnIndex["Comment"]; ok && idx < len(record) {
		tx.Description = strings.TrimSpace(record[idx])
	}
	if tx.Description == "" {
		if idx, ok := columnIndex["Instrument Symbol"]; ok && idx < len(record) {
			tx.Description = strings.TrimSpace(record[idx])
		}
	}

	// Trade Date as CreatedAt
	if idx, ok := columnIndex["Trade Date"]; ok && idx < len(record) {
		dateStr := strings.TrimSpace(record[idx])
		parsedTime, err := time.Parse("02-Jan-2006", dateStr)
		if err == nil {
			tx.CreatedAt = parsedTime
		}
	}

	return tx, nil
}
