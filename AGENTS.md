# Implementation Plan: Event-Sourced Transaction Persistence

## Overview
Build an event-sourced transaction system that:
- Parses transactions from multiple brokers into `domain.Transaction` entries
- Stores events immutably in an event store
- Deduplicates transactions at the event store level
- Projects open positions per instrument
- Calculates annualized returns on-the-fly

## Architecture

### Core Principles
- **Event Sourcing**: Single source of truth is the event store, not derived state
- **Event Store Only**: No separate transactions table; projections calculated from events
- **Immutable Events**: Events are append-only with broker name and import timestamp metadata
- **Idempotent Imports**: Duplicate detection via uniqueness key prevents re-processing

### Event Store Schema
```
events table:
- id (UUID, primary key)
- aggregateID (instrument)
- type (string, e.g., "TransactionImported")
- broker (string, metadata)
- importedAt (timestamp, metadata)
- payload (JSON, domain.Transaction data)
- createdAt (timestamp)
- uniquenessKey (hash of date + instrument + category + quantity + amount)
- UNIQUE(uniquenessKey)
```

## Implementation Tasks

### ✅ 1. Event Store Layer - COMPLETED
- **File**: `internal/store/eventstore.go`
- ✅ Created `EventStore` interface with methods:
  - `AppendEvent(ctx context.Context, event Event) error` - append with uniqueness check
  - `GetEvents(ctx context.Context, aggregateID string) ([]Event, error)` - retrieve by instrument
  - `GetEventsByBroker(ctx context.Context, broker string) ([]Event, error)`
  - `GetAllEvents(ctx context.Context) ([]Event, error)` - retrieve all events
- ✅ Implemented SQLite backend
- ✅ Uniqueness key enforced at DB level (UNIQUE constraint on uniquenessKey)
- ✅ On duplicate: log warning + return nil (idempotent no-op)
- ✅ Migration support via golang-migrate
- ✅ Comprehensive tests with proper migrations path handling

### ✅ 2. Event Model & Uniqueness Key - COMPLETED
- **File**: `internal/domain/event.go`
- ✅ Defined `Event` struct:
  ```go
  type Event struct {
    ID            string
    AggregateID   string        // instrument (e.g., "AAPL")
    Type          string        // e.g., "TransactionImported"
    Broker        string        // metadata
    ImportedAt    time.Time     // metadata
    Payload       Transaction   // the actual transaction data
    UniquenessKey string        // hash(date + instrument + category + quantity + amount)
    CreatedAt     time.Time
  }
  ```
- ✅ Helper function: `ComputeUniquenessKey(date, instrument, category, quantity, amount) string`
  - Uses high-precision float formatting (`%.15g`) to avoid hash collisions
  - Includes both quantity and amount to distinguish similar transactions
  - SHA256 hashing for robust deduplication

### ✅ 3. Transaction Model Enhancement - COMPLETED
- **File**: `internal/domain/transaction.go`
- ✅ Added `Quantity float64` field (number of instruments traded)
- ✅ Updated field semantics:
  - `Amount`: Total transaction value (price × quantity)
  - `Quantity`: Number of instruments/shares traded
  - `Type`: "buy" or "sell" (indicates direction of trade)
  - `Category`: "Trade" or "Corporate Action" (from Event field)

### ✅ 4. Parser Integration - COMPLETED
- **File**: `internal/parsers/parser.go`
- ✅ Created `ParseAndStore(ctx context.Context, brokerName string, parser Parser, data io.Reader, store EventStore) error`
  - Parses transactions using broker-specific parser
  - Validates required fields (Instrument must be populated)
  - Computes uniqueness keys using date + instrument + category + quantity + amount
  - Appends events to store
  - Logs duplicates with warning, continues processing

### ✅ 5. Click Trade Parser Update - COMPLETED
- **File**: `internal/parsers/click_trade/parser.go`
- ✅ Extracts `Quantity` from CSV "Quantity" column
- ✅ Maps CSV "Event" field → `domain.Transaction.Category`
- ✅ Ensures `Type` field reflects buy/sell transaction direction
- ✅ Parses trade dates and amounts with proper formatting
- ✅ Includes comprehensive test coverage

### ✅ 6. Projection: Open Positions - COMPLETED
- **File**: `internal/projections/positions.go`
- ✅ Function: `ProjectOpenPositions(ctx context.Context, events []Event) (map[string]Position, error)`
  - Aggregates events by instrument
  - For buy transactions: increases quantity and cost basis
  - For sell transactions: decreases quantity, preserves cost basis for gain calculations
  - Calculates average cost (total_cost / quantity), current value (quantity × average_cost)
  - Returns map[instrument]Position
- ✅ Function: `ProjectAnnualizedReturn(ctx context.Context, events []Event, position Position) float64`
  - Calculates return since first investment in instrument
  - Formula: (current_value - total_invested) / total_invested, annualized
  - Annualizes using compound growth formula: (1 + totalReturn)^(1/yearsHeld) - 1
  - Handles edge cases (zero division, NaN, Inf)
- ✅ Function: `ProjectPortfolioMetrics(ctx context.Context, events []Event, positions map[string]Position) PortfolioMetrics`
  - Aggregates metrics across all positions
  - Calculates total value, total cost, unrealized gains
  - Weight-averages annualized returns by cost basis proportion

### ✅ 7. Database Setup - COMPLETED
- **File**: `internal/infrastructure/migrations/000001_create_events_table.up.sql`
- ✅ SQLite schema creation for events table
- ✅ UNIQUE(uniquenessKey) constraint
- ✅ Indexes on aggregate_id, broker, and created_at for query performance
- ✅ Migration down script for rollback support

### ✅ 8. HTTP API Server - COMPLETED
- **Files**: `cmd/server/main.go`, `internal/server/server.go`, `internal/server/controllers/transactions.go`, `internal/repository/transaction_repository.go`
- ✅ `portfolio-server` binary with `--addr` flag (default `:8080`)
- ✅ Gin-based HTTP router configured in `internal/server/server.go`
- ✅ `GET /transactions` endpoint with query parameter filters:
  - `instrument`, `broker`, `type`, `from` (YYYY-MM-DD), `to` (YYYY-MM-DD)
- ✅ `TransactionRepository` interface backed by EventStore
- ✅ Controller unit tests in `internal/server/controllers/transactions_test.go`
- ✅ Separate import binary (`cmd/import/main.go`) as `portfolio-import`

## Key Behaviors

### Deduplication
- **Uniqueness Key**: `SHA256(date + "|" + instrument + "|" + category + "|" + quantity + "|" + amount)`
  - High-precision float formatting (`%.15g`) prevents collisions from rounding
  - Includes both quantity and amount to distinguish similar transactions
- **Enforcement**: Database UNIQUE constraint on uniquenessKey column
- **On Duplicate**: Log warning, skip silently (idempotent)
- **Recovery**: Importing same file twice = no-op

### Idempotency
- Each import is idempotent
- Same batch re-imported = no errors, duplicate events logged as warnings
- Import batch ID not required; derived from broker + importedAt + hash

### Error Handling
- Parse errors: fail fast, log error, abort import
- Store errors (duplicates): log warning, continue
- Store errors (other): fail fast, return error

## File Structure
```
internal/
  domain/
    transaction.go ✅ (enhanced with Quantity field)
    event.go ✅ (new)
    event_test.go ✅ (tests for uniqueness key)
  store/
    eventstore.go ✅ (new interface + PostgreSQL impl)
    eventstore_test.go ✅ (comprehensive tests)
  projections/
    positions.go ✅ (new)
  parsers/
    parser.go ✅ (ParseAndStore implementation)
    click_trade/
      parser.go ✅ (updated with Quantity and Category)
      parser_test.go ✅ (tests)
  infrastructure/
    database.go ✅ (new infrastructure package with Connect and runMigrations)
    migrations/
      000001_create_events_table.up.sql ✅
      000001_create_events_table.down.sql ✅
  repository/
    transaction_repository.go ✅ (EventStore-backed TransactionRepository)
  server/
    server.go ✅ (Gin router setup)
    controllers/
      transactions.go ✅ (GET /transactions handler)
      transactions_test.go ✅ (controller unit tests)
cmd/
  import/
    main.go ✅ (portfolio-import binary)
  server/
    main.go ✅ (portfolio-server binary)
  main/
    main.go ✅ (legacy CLI with import-file command)
```

## Testing Strategy - ✅ COMPLETED
- ✅ Unit tests for uniqueness key computation
- ✅ Unit tests for deduplication (insert duplicate, verify no-op)
- ✅ Integration tests for parser → store flow
- ✅ Projection tests for position calculation (covered by implementation)
- ✅ Idempotency tests (import same event twice, verify same state)

## Next Steps - TODO

### ✅ 9. CLI Commands - COMPLETED
- ✅ `portfolio-import` binary: import transaction files (`--file-name`, `--broker`)
- ✅ `portfolio-server` binary: HTTP server for querying transactions (`--addr`)

### 10. Configuration
- ✅ Database connection configuration via DATABASE_DSN environment variable
- ✅ Migrations path configuration via MIGRATIONS_PATH environment variable (default: file://./internal/infrastructure/migrations)
- [ ] Support for multiple broker configurations

### 11. Documentation
- ✅ Updated README.md with current implementation status
- ✅ Updated README.md with CLI usage documentation
- ✅ Updated README.md with HTTP API server documentation
- ✅ Updated AGENTS.md with completion checklist
- ✅ Add examples of CSV import usage in README
- ✅ Add HTTP server usage and API endpoint docs in README
