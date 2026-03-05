package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go-portfolio/internal/domain"
)

// StockQuotationRepository defines the interface for stock quotation persistence.
type StockQuotationRepository interface {
	// Insert stores a new stock quotation and returns it with the generated ID.
	Insert(ctx context.Context, q domain.StockQuotation) (domain.StockQuotation, error)

	// GetByTicker retrieves all quotations for the given ticker symbol.
	GetByTicker(ctx context.Context, ticker string) ([]domain.StockQuotation, error)
}

// PostgreSQLStockQuotationRepository implements StockQuotationRepository using PostgreSQL.
type PostgreSQLStockQuotationRepository struct {
	db *sql.DB
}

// NewStockQuotationRepository creates a new StockQuotationRepository backed by PostgreSQL.
func NewStockQuotationRepository(db *sql.DB) (StockQuotationRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db must not be nil")
	}
	return &PostgreSQLStockQuotationRepository{db: db}, nil
}

// Insert stores a new stock quotation in the database. A UUID is generated for the ID.
func (r *PostgreSQLStockQuotationRepository) Insert(ctx context.Context, q domain.StockQuotation) (domain.StockQuotation, error) {
	q.ID = uuid.New().String()

	query := `INSERT INTO stock_quotations (id, ticker, price, volume) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, q.ID, q.Ticker, q.Price, q.Volume)
	if err != nil {
		return domain.StockQuotation{}, fmt.Errorf("failed to insert stock quotation: %w", err)
	}

	return q, nil
}

// GetByTicker retrieves all stock quotations for the given ticker symbol.
func (r *PostgreSQLStockQuotationRepository) GetByTicker(ctx context.Context, ticker string) ([]domain.StockQuotation, error) {
	query := `SELECT id, ticker, price, volume FROM stock_quotations WHERE ticker = $1`

	rows, err := r.db.QueryContext(ctx, query, ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to query stock quotations: %w", err)
	}
	defer rows.Close()

	var quotations []domain.StockQuotation
	for rows.Next() {
		var q domain.StockQuotation
		if err := rows.Scan(&q.ID, &q.Ticker, &q.Price, &q.Volume); err != nil {
			return nil, fmt.Errorf("failed to scan stock quotation: %w", err)
		}
		quotations = append(quotations, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock quotations: %w", err)
	}

	return quotations, nil
}
