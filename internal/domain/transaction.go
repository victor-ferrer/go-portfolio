package domain

import "time"

// Transaction represents a financial transaction.
type Transaction struct {
	ID          string    `json:"id"`
	Amount      float64   `json:"amount"`       // Total transaction value (price × quantity)
	Quantity    float64   `json:"quantity"`     // Number of instruments traded
	Type        string    `json:"type"`         // e.g., "buy", "sell"
	Category    string    `json:"category"`     // e.g., "Trade", "Corporate Action"
	Description string    `json:"description"`
	Currency    string    `json:"currency"`
	Instrument  string    `json:"instrument"`
	ISIN        string    `json:"isin"`
	CreatedAt   time.Time `json:"created_at"`
}
