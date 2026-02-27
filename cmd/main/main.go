package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go-portfolio/internal/config"
	"go-portfolio/internal/parsers"
	"go-portfolio/internal/parsers/click_trade"
	"go-portfolio/internal/store"
)

func main() {
	// Define subcommands
	importCmd := flag.NewFlagSet("import-file", flag.ExitOnError)
	fileName := importCmd.String("file-name", "", "Path to the transaction file to import (required)")
	broker := importCmd.String("broker", "", "Broker name (currently only 'click-trade' is supported) (required)")

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
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("go-portfolio - Portfolio management tool")
	fmt.Println("\nUsage:")
	fmt.Println("  portfolio import-file --file-name <path> --broker <broker-name>")
	fmt.Println("\nCommands:")
	fmt.Println("  import-file    Import transactions from a file")
	fmt.Println("\nOptions:")
	fmt.Println("  --file-name    Path to the transaction file")
	fmt.Println("  --broker       Broker name (currently supported: click-trade)")
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

	// Get database configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize event store
	eventStore, err := store.NewEventStore(cfg.DSN(), cfg.MigrationsPath)
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

