package server

import (
	"github.com/gin-gonic/gin"

	"go-portfolio/internal/repository"
	"go-portfolio/internal/server/controllers"
	"go-portfolio/internal/store"
)

// New creates and configures a Gin engine with all routes registered.
func New(eventStore store.EventStore) *gin.Engine {
	router := gin.Default()

	txRepo := repository.NewTransactionRepository(eventStore)
	txController := controllers.NewTransactionController(txRepo)

	router.GET("/transactions", txController.GetTransactions)

	return router
}
