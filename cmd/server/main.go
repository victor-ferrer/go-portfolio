package main

import (
	"flag"
	"fmt"
	"log"

	"go-portfolio/infrastructure/database"
	"go-portfolio/internal/config"
	"go-portfolio/internal/server"
	"go-portfolio/internal/store"
	webui "go-portfolio/web"
)

func main() {
	flag.Usage = func() {
		fmt.Println("portfolio-server - Start the portfolio HTTP server")
		fmt.Println("\nUsage:")
		fmt.Println("  portfolio-server [--addr <address>]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	addr := flag.String("addr", ":8080", "Address to listen on")
	flag.Parse()

	if err := runServer(*addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runServer(addr string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := database.Connect(cfg.DSN(), cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	eventStore, err := store.NewEventStore(db)
	if err != nil {
		return fmt.Errorf("failed to initialize event store: %w", err)
	}
	defer eventStore.Close()

	router := server.New(eventStore, webui.StaticFiles)

	fmt.Printf("Starting server on %s\n", addr)
	return router.Run(addr)
}
