package click_trade

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go-portfolio/internal/domain"
)

// Parser handles parsing of Click Trade data.
type Parser struct {
	// Add fields as needed
}

// NewParser creates a new Click Trade parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads a CSV file and returns a slice of transactions.
func (p *Parser) Parse(reader io.Reader) ([]domain.Transaction, error) {
	csvReader := csv.NewReader(reader)
	
	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Create a map of column names to indices
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[strings.TrimSpace(col)] = i
	}

	var transactions []domain.Transaction

	// Read data rows
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		tx, err := p.parseRow(record, columnIndex)
		if err != nil {
			// Log error but continue processing other rows
			continue
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// parseRow converts a CSV row into a Transaction.
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

	// ISIN
	if idx, ok := columnIndex["Instrument ISIN"]; ok && idx < len(record) {
		tx.ISIN = strings.TrimSpace(record[idx])
	}

	// Type
	if idx, ok := columnIndex["Type"]; ok && idx < len(record) {
		tx.Type = strings.TrimSpace(record[idx])
	}

	// Category from Event or Transaction Type
	if idx, ok := columnIndex["Event"]; ok && idx < len(record) {
		tx.Category = strings.TrimSpace(record[idx])
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
