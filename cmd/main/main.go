package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go-portfolio/internal/parsers"
	"go-portfolio/internal/parsers/click_trade"
	"go-portfolio/internal/server"
	"go-portfolio/internal/store"
)

func main() {
	// Define subcommands
	importCmd := flag.NewFlagSet("import-file", flag.ExitOnError)
	fileName := importCmd.String("file-name", "", "Path to the transaction file to import (required)")
	broker := importCmd.String("broker", "", "Broker name (currently only 'click-trade' is supported) (required)")

	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	addr := serverCmd.String("addr", ":8080", "Address to listen on (default: :8080)")

	// Parse command-line arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "import-file":
		importCmd.Parse(os.Args[2:])
		if *fileName == "" || *broker == "" {
			fmt.Println("Error: Both --file-name and --broker flags are required")
			importCmd.PrintDefaults()
			os.Exit(1)
		}
		if err := runImport(*fileName, *broker); err != nil {
			log.Fatalf("Import failed: %v", err)
		}
	case "server":
		serverCmd.Parse(os.Args[2:])
		if err := runServer(*addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("go-portfolio - Portfolio management tool")
	fmt.Println("\nUsage:")
	fmt.Println("  portfolio import-file --file-name <path> --broker <broker-name>")
	fmt.Println("  portfolio server [--addr <address>]")
	fmt.Println("\nCommands:")
	fmt.Println("  import-file    Import transactions from a file")
	fmt.Println("  server         Start the HTTP server")
	fmt.Println("\nOptions:")
	fmt.Println("  --file-name    Path to the transaction file")
	fmt.Println("  --broker       Broker name (currently supported: click-trade)")
	fmt.Println("  --addr         Address to listen on (default: :8080)")
}

func runImport(fileName, brokerName string) error {
	ctx := context.Background()

	// Validate broker name
	if brokerName != "click-trade" {
		return fmt.Errorf("unsupported broker: %s (only 'click-trade' is currently supported)", brokerName)
	}

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()

	// Get database connection string from environment or use default
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return fmt.Errorf("DATABASE_DSN environment variable is not set")
	}

	// Get migrations path from environment or use default
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://./migrations"
	}

	// Initialize event store
	eventStore, err := store.NewEventStore(dsn, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize event store: %w", err)
	}
	defer eventStore.Close()

	// Create parser based on broker name
	var parser parsers.Parser
	switch brokerName {
	case "click-trade":
		parser = click_trade.NewParser()
	default:
		return fmt.Errorf("no parser available for broker: %s", brokerName)
	}

	// Parse and store transactions
	fmt.Printf("Importing transactions from %s (broker: %s)...\n", fileName, brokerName)
	
	if err := parsers.ParseAndStore(ctx, brokerName, parser, file, eventStore); err != nil {
		return fmt.Errorf("failed to parse and store transactions: %w", err)
	}

	fmt.Println("Import completed successfully!")
	return nil
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
