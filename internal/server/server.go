package server

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-portfolio/internal/repository"
	"go-portfolio/internal/server/controllers"
	"go-portfolio/internal/store"
)

// New creates and configures a Gin engine with all routes registered.
// staticFiles, if non-nil, is served as the root web UI (embedded FS with a "dist" sub-directory).
func New(eventStore store.EventStore, staticFiles fs.FS) *gin.Engine {
	router := gin.Default()

	txRepo := repository.NewTransactionRepository(eventStore)
	txController := controllers.NewTransactionController(txRepo)

	router.GET("/transactions", txController.GetTransactions)

	if staticFiles != nil {
		distFS, err := fs.Sub(staticFiles, "dist")
		if err == nil {
			router.NoRoute(gin.WrapH(http.FileServer(http.FS(distFS))))
		}
	}

	return router
}
