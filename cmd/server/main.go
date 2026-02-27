package main

import (
	"flag"
	"fmt"
	"log"

	"go-portfolio/internal/config"
	"go-portfolio/internal/server"
	"go-portfolio/internal/store"
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

	eventStore, err := store.NewEventStore(cfg.DSN(), cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize event store: %w", err)
	}
	defer eventStore.Close()

	router := server.New(eventStore)

	fmt.Printf("Starting server on %s\n", addr)
	return router.Run(addr)
}
