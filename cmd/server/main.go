package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return fmt.Errorf("DATABASE_DSN environment variable is not set")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://./migrations"
	}

	eventStore, err := store.NewEventStore(dsn, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize event store: %w", err)
	}
	defer eventStore.Close()

	router := server.New(eventStore)

	fmt.Printf("Starting server on %s\n", addr)
	return router.Run(addr)
}
