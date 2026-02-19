package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"go-portfolio/internal/domain"
	"go-portfolio/internal/repository"
	"go-portfolio/internal/server/controllers"
)

// stubTransactionRepository is an in-memory TransactionRepository for testing.
type stubTransactionRepository struct {
	transactions []domain.Transaction
}

func (s *stubTransactionRepository) GetTransactions(_ context.Context, filter repository.TransactionFilter) ([]domain.Transaction, error) {
	result := make([]domain.Transaction, 0)
	for _, tx := range s.transactions {
		if filter.Instrument != "" && tx.Instrument != filter.Instrument {
			continue
		}
		if filter.Broker != "" {
			// broker filter is on event metadata, not transaction; skip for stub
		}
		if filter.Type != "" && tx.Type != filter.Type {
			continue
		}
		if filter.From != nil && tx.CreatedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && tx.CreatedAt.After(*filter.To) {
			continue
		}
		result = append(result, tx)
	}
	return result, nil
}

func newTestRouter(repo repository.TransactionRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewTransactionController(repo)
	router.GET("/transactions", ctrl.GetTransactions)
	return router
}

func TestGetTransactions_NoFilter(t *testing.T) {
	repo := &stubTransactionRepository{
		transactions: []domain.Transaction{
			{ID: "1", Instrument: "AAPL", Type: "buy", CreatedAt: time.Now()},
			{ID: "2", Instrument: "GOOG", Type: "sell", CreatedAt: time.Now()},
		},
	}

	router := newTestRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var txs []domain.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}
}

func TestGetTransactions_FilterByInstrument(t *testing.T) {
	repo := &stubTransactionRepository{
		transactions: []domain.Transaction{
			{ID: "1", Instrument: "AAPL", Type: "buy", CreatedAt: time.Now()},
			{ID: "2", Instrument: "GOOG", Type: "sell", CreatedAt: time.Now()},
		},
	}

	router := newTestRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions?instrument=AAPL", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var txs []domain.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
	if txs[0].Instrument != "AAPL" {
		t.Errorf("expected AAPL, got %s", txs[0].Instrument)
	}
}

func TestGetTransactions_FilterByType(t *testing.T) {
	repo := &stubTransactionRepository{
		transactions: []domain.Transaction{
			{ID: "1", Instrument: "AAPL", Type: "buy", CreatedAt: time.Now()},
			{ID: "2", Instrument: "GOOG", Type: "sell", CreatedAt: time.Now()},
		},
	}

	router := newTestRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions?type=buy", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var txs []domain.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 buy transaction, got %d", len(txs))
	}
	if txs[0].Type != "buy" {
		t.Errorf("expected type buy, got %s", txs[0].Type)
	}
}

func TestGetTransactions_FilterByDateRange(t *testing.T) {
	repo := &stubTransactionRepository{
		transactions: []domain.Transaction{
			{ID: "1", Instrument: "AAPL", Type: "buy", CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
			{ID: "2", Instrument: "GOOG", Type: "buy", CreatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
			{ID: "3", Instrument: "MSFT", Type: "buy", CreatedAt: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
		},
	}

	router := newTestRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions?from=2024-01-15&to=2024-11-30", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var txs []domain.Transaction
	if err := json.Unmarshal(w.Body.Bytes(), &txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction in range, got %d", len(txs))
	}
	if txs[0].ID != "2" {
		t.Errorf("expected GOOG transaction, got %s", txs[0].ID)
	}
}

func TestGetTransactions_InvalidFromDate(t *testing.T) {
	repo := &stubTransactionRepository{}
	router := newTestRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions?from=not-a-date", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetTransactions_InvalidToDate(t *testing.T) {
	repo := &stubTransactionRepository{}
	router := newTestRouter(repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/transactions?to=not-a-date", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
