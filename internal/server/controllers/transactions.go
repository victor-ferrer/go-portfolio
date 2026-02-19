package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-portfolio/internal/repository"
)

// TransactionController handles HTTP requests for transactions.
type TransactionController struct {
	repo repository.TransactionRepository
}

// NewTransactionController creates a new TransactionController.
func NewTransactionController(repo repository.TransactionRepository) *TransactionController {
	return &TransactionController{repo: repo}
}

// GetTransactions returns a list of transactions with optional filtering.
//
//	GET /transactions?instrument=AAPL&broker=click-trade&type=buy&from=2024-01-01&to=2024-12-31
func (c *TransactionController) GetTransactions(ctx *gin.Context) {
	filter := repository.TransactionFilter{
		Instrument: ctx.Query("instrument"),
		Broker:     ctx.Query("broker"),
		Type:       ctx.Query("type"),
	}

	if fromStr := ctx.Query("from"); fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date, expected format: YYYY-MM-DD"})
			return
		}
		filter.From = &t
	}

	if toStr := ctx.Query("to"); toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date, expected format: YYYY-MM-DD"})
			return
		}
		// Set to end of day so inclusive
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		filter.To = &endOfDay
	}

	transactions, err := c.repo.GetTransactions(ctx.Request.Context(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve transactions"})
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}
