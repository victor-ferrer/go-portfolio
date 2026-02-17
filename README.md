# go-portfolio

An event-sourced portfolio management system built with Go. This application tracks financial transactions from multiple brokers, stores them as immutable events, and projects portfolio positions and performance metrics.

## Features

- **Event Sourcing**: Single source of truth using immutable events
- **Multi-Broker Support**: Parse and import transactions from multiple brokers
- **Automatic Deduplication**: Uniqueness key prevents duplicate transaction imports
- **Position Tracking**: Calculate open positions and cost basis per instrument
- **Performance Metrics**: Compute annualized returns and portfolio-wide metrics
- **SQLite Storage**: Lightweight database with migration support

## Architecture

The system follows event sourcing principles:
- Transactions are parsed and stored as immutable events
- Each event has a uniqueness key (hash of date, instrument, category, quantity, amount)
- Positions and metrics are calculated on-the-fly from event stream
- Database enforces idempotency via UNIQUE constraint on uniqueness keys

## Project Structure

```
.
├── cmd/                           # Command-line applications
│   └── main/                      # Main application entry point
├── internal/                      # Private application code
│   ├── domain/                    # Domain models
│   │   ├── event.go               # Event model and uniqueness key computation
│   │   └── transaction.go         # Transaction model
│   ├── store/                     # Event store implementation
│   │   └── eventstore.go          # SQLite event store with deduplication
│   ├── parsers/                   # Transaction parsers
│   │   ├── parser.go              # Generic parser interface and ParseAndStore
│   │   └── click_trade/           # Click Trade broker parser
│   │       └── parser.go          
│   └── projections/               # Event projections
│       └── positions.go           # Position and metrics calculations
├── migrations/                    # Database migrations
│   └── 000001_create_events_table.up.sql
├── go.mod                         # Module definition
└── go.sum                         # Module checksums
```

## Implementation Status

✅ **Completed:**
- Event store layer with SQLite backend
- Event model with uniqueness key computation (SHA256 hash)
- Transaction model with quantity and category fields
- Parser integration with ParseAndStore function
- Click Trade CSV parser
- Position projections (open positions, annualized returns, portfolio metrics)
- Database migrations
- Comprehensive test coverage

🚧 **In Progress:**
- CLI commands for importing transactions
- CLI commands for viewing positions and metrics

## Getting Started

### Prerequisites

- Go 1.25 or higher
- CGO enabled (required for SQLite)

### Build

```bash
go build -o bin/portfolio ./cmd/main
```

### Run

```bash
go run ./cmd/main
```

### Run Tests

```bash
go test ./...
```

## Database Schema

The event store uses a single `events` table:

```sql
CREATE TABLE events (
    id TEXT PRIMARY KEY,
    aggregate_id TEXT NOT NULL,           -- instrument (e.g., "AAPL")
    type TEXT NOT NULL,                   -- e.g., "TransactionImported"
    broker TEXT NOT NULL,                 -- broker name
    imported_at TIMESTAMP NOT NULL,       -- import timestamp
    payload TEXT NOT NULL,                -- JSON transaction data
    created_at TIMESTAMP NOT NULL,        -- event creation time
    uniqueness_key TEXT NOT NULL UNIQUE   -- deduplication key
);
```

## Key Concepts

### Uniqueness Key
Transactions are deduplicated using a SHA256 hash of:
- Transaction date
- Instrument
- Category
- Quantity (high-precision float)
- Amount (high-precision float)

### Idempotency
Importing the same transaction file multiple times is safe:
- Duplicate events are silently skipped with a warning
- No errors are raised on re-import
- Same file imported twice results in identical state

### Event Projections
Position calculations are performed on-the-fly from events:
- **Open Positions**: Aggregate buys/sells to calculate current holdings
- **Cost Basis**: Track average cost per instrument
- **Annualized Returns**: Compute returns using compound growth formula
- **Portfolio Metrics**: Aggregate across all positions with weighted returns
