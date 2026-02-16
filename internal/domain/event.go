package domain

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Event represents an immutable event in the event store.
type Event struct {
	ID            string        `json:"id"`
	AggregateID   string        `json:"aggregate_id"`   // instrument (e.g., "AAPL")
	Type          string        `json:"type"`           // e.g., "TransactionImported"
	Broker        string        `json:"broker"`         // metadata
	ImportedAt    time.Time     `json:"imported_at"`    // metadata
	Payload       Transaction   `json:"payload"`        // the actual transaction data
	UniquenessKey string        `json:"uniqueness_key"` // hash(date + instrument + category + quantity)
	CreatedAt     time.Time     `json:"created_at"`
}

// ComputeUniquenessKey generates a hash-based uniqueness key from transaction fields.
// Components: date + instrument + category + quantity + amount
// Uses high-precision float formatting to avoid collisions on similar values.
func ComputeUniquenessKey(date time.Time, instrument string, category string, quantity float64, amount float64) string {
	// Use %.15g for high precision float representation to avoid hash collisions
	data := fmt.Sprintf("%s|%s|%s|%.15g|%.15g", date.Format("2006-01-02"), instrument, category, quantity, amount)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
