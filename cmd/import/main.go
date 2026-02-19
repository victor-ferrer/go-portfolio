package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go-portfolio/internal/parsers"
	"go-portfolio/internal/parsers/click_trade"
	"go-portfolio/internal/store"
)

func main() {
	flag.Usage = func() {
		fmt.Println("portfolio-import - Import transactions into the portfolio")
		fmt.Println("\nUsage:")
		fmt.Println("  portfolio-import --file-name <path> --broker <broker-name>")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	fileName := flag.String("file-name", "", "Path to the transaction file to import (required)")
	broker := flag.String("broker", "", "Broker name (currently only 'click-trade' is supported) (required)")
	flag.Parse()

	if *fileName == "" || *broker == "" {
		fmt.Println("Error: Both --file-name and --broker flags are required")
		flag.Usage()
		os.Exit(1)
	}

	if err := runImport(*fileName, *broker); err != nil {
		log.Fatalf("Import failed: %v", err)
	}
}

func runImport(fileName, brokerName string) error {
	ctx := context.Background()

	if brokerName != "click-trade" {
		return fmt.Errorf("unsupported broker: %s (only 'click-trade' is currently supported)", brokerName)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer file.Close()

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

	var parser parsers.Parser
	switch brokerName {
	case "click-trade":
		parser = click_trade.NewParser()
	default:
		return fmt.Errorf("no parser available for broker: %s", brokerName)
	}

	fmt.Printf("Importing transactions from %s (broker: %s)...\n", fileName, brokerName)

	if err := parsers.ParseAndStore(ctx, brokerName, parser, file, eventStore); err != nil {
		return fmt.Errorf("failed to parse and store transactions: %w", err)
	}

	fmt.Println("Import completed successfully!")
	return nil
}
