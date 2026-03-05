package domain

// StockQuotation represents a price quotation for a stock instrument.
type StockQuotation struct {
	ID     string   `json:"id"`
	Ticker string   `json:"ticker"`
	Price  float64  `json:"price"`
	Volume *float64 `json:"volume,omitempty"` // optional
}
