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
- uniquenessKey (hash of date + instrument + category + quantity)
- UNIQUE(uniquenessKey)
```

## Implementation Tasks

### 1. Event Store Layer
- **File**: `internal/store/eventstore.go`
- Create `EventStore` interface with methods:
  - `AppendEvent(ctx context.Context, event Event) error` - append with uniqueness check
  - `GetEvents(ctx context.Context, aggregateID string) ([]Event, error)` - retrieve by instrument
  - `GetEventsByBroker(ctx context.Context, broker string) ([]Event, error)`
- Implement SQLite backend
- Uniqueness key enforced at DB level (UNIQUE constraint on uniquenessKey)
- On duplicate: log warning + return nil (idempotent no-op)

### 2. Event Model & Uniqueness Key
- **File**: `internal/domain/event.go`
- Define `Event` struct:
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
- Helper function: `ComputeUniquenessKey(date, instrument, category, quantity, amount) string`
  - Uses high-precision float formatting (`%.15g`) to avoid hash collisions
  - Includes both quantity and amount to distinguish similar transactions

### 3. Transaction Model Enhancement
- **File**: `internal/domain/transaction.go`
- Add `Quantity float64` field (number of instruments traded)
- Update field semantics:
  - `Amount`: Total transaction value (price × quantity)
  - `Quantity`: Number of instruments/shares traded
  - `Type`: "buy" or "sell" (indicates direction of trade)
  - `Category`: "Trade" or "Corporate Action" (from Event field)

### 4. Parser Integration
- **File**: `internal/parsers/parser.go` (new generic parser wrapper)
- Create `ParseAndStore(ctx context.Context, brokerName string, parser Parser, data io.Reader, store EventStore) error`
  - Parse transactions using broker-specific parser
  - Validate required fields (Instrument must be populated)
  - Compute uniqueness keys using date + instrument + category + quantity + amount
  - Append events to store
  - Log duplicates with warning, continue processing

### 5. Click Trade Parser Update
- **File**: `internal/parsers/click_trade/parser.go`
- Extract `Quantity` from CSV "Quantity" column
- Map CSV "Event" field → `domain.Transaction.Category`
- Ensure `Type` field reflects buy/sell transaction direction

### 6. Projection: Open Positions
- **File**: `internal/projections/positions.go`
- Function: `ProjectOpenPositions(ctx context.Context, events []Event) (map[string]Position, error)`
  - Aggregate events by instrument
  - For buy transactions: increase quantity and cost basis
  - For sell transactions: decrease quantity, preserve cost basis for gain calculations
  - Calculate average cost (total_cost / quantity), current value (quantity × average_cost)
  - Return map[instrument]Position
- Function: `ProjectAnnualizedReturn(ctx context.Context, events []Event, position Position) float64`
  - Calculate return since first investment in instrument
  - Formula: (current_value - total_invested) / total_invested, annualized
  - Annualize using compound growth formula
- Function: `ProjectPortfolioMetrics(ctx context.Context, events []Event, positions map[string]Position) PortfolioMetrics`
  - Aggregate metrics across all positions
  - Calculate total value, total cost, unrealized gains
  - Weight-average annualized returns by cost basis proportion

### 5. Database Setup
- **File**: `internal/store/migrations.go`
- SQLite schema creation for events table
- Ensure UNIQUE(uniquenessKey) constraint

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
    transaction.go (existing, no changes)
    event.go (new)
  store/
    eventstore.go (new interface + SQLite impl)
    migrations.go (new)
  projections/
    positions.go (new)
  parsers/
    parser.go (extend with ParseAndStore)
```

## Testing Strategy
- Unit tests for uniqueness key computation
- Unit tests for deduplication (insert duplicate, verify no-op)
- Integration tests for parser → store flow
- Projection tests for position calculation
- Idempotency tests (import same file twice, verify same state)
