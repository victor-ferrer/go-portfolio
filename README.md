# go-portfolio

An event-sourced portfolio management system built with Go. This application tracks financial transactions from multiple brokers, stores them as immutable events, and projects portfolio positions and performance metrics.

## Features

- **Event Sourcing**: Single source of truth using immutable events
- **Multi-Broker Support**: Parse and import transactions from multiple brokers
- **Automatic Deduplication**: Uniqueness key prevents duplicate transaction imports
- **Position Tracking**: Calculate open positions and cost basis per instrument
- **Performance Metrics**: Compute annualized returns and portfolio-wide metrics
- **PostgreSQL Storage**: Production-ready database with Docker support

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
│   │   └── eventstore.go          # PostgreSQL event store with deduplication
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
- Event store layer with PostgreSQL backend
- Event model with uniqueness key computation (SHA256 hash)
- Transaction model with quantity and category fields
- Parser integration with ParseAndStore function
- Click Trade CSV parser
- Position projections (open positions, annualized returns, portfolio metrics)
- Database migrations
- Comprehensive test coverage
- CLI command for importing transactions

🚧 **In Progress:**
- CLI commands for viewing positions and metrics

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Docker 20.10+ and Docker Compose V2 (2.0+)
- PostgreSQL client (optional, for manual database access)

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/victor-ferrer/go-portfolio.git
   cd go-portfolio
   ```

2. **Start PostgreSQL database**
   ```bash
   make docker-up
   ```

3. **Set environment variables**
   ```bash
   export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"
   ```
   
   Or copy the example environment file:
   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

4. **Run migrations**
   Migrations are automatically run when the application starts and connects to the database.

### Build

```bash
make build
```

### Run

```bash
make run
```

## Usage

### Import Transactions

The `import-file` command allows you to import transaction data from broker CSV files.

**Basic Usage:**
```bash
./bin/portfolio import-file --file-name <path-to-csv> --broker <broker-name>
```

**Example:**
```bash
# Set the database connection string
export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"

# Import transactions from Click Trade CSV file
./bin/portfolio import-file --file-name ./data/click-trade-transactions.csv --broker click-trade
```

**Supported Brokers:**
- `click-trade`: Click Trade broker CSV format

**Features:**
- **Automatic Deduplication**: Re-importing the same file will skip duplicate transactions
- **Error Handling**: Clear error messages for missing files, invalid brokers, or connection issues
- **Transaction Tracking**: Each imported transaction is stored as an immutable event

**CSV Format Example (Click Trade):**
The Click Trade CSV should include columns such as:
- `Trade Date`: Date of the transaction
- `Booked Amount`: Total transaction amount
- `Currency`: Transaction currency
- `Instrument`: Stock/security name
- `Instrument ISIN`: ISIN code
- `Type`: Transaction type
- `Event`: Category (Trade, Corporate Action, etc.)
- `Quantity`: Number of shares/units
- `Comment`: Additional description

### Run Tests

```bash
# Start the database first
make docker-up

# Set the DATABASE_DSN environment variable
export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"

# Run tests
make test
```

### Docker Commands

- `make docker-up`: Start PostgreSQL database
- `make docker-down`: Stop PostgreSQL database
- `make docker-clean`: Stop and remove database with volumes
- `make db-setup`: Start database and display connection instructions

## Database Schema

The event store uses a single `events` table in PostgreSQL:

```sql
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,      -- instrument (e.g., "AAPL")
    type VARCHAR(255) NOT NULL,              -- e.g., "TransactionImported"
    broker VARCHAR(255) NOT NULL,            -- broker name
    imported_at TIMESTAMP NOT NULL,          -- import timestamp
    payload JSONB NOT NULL,                  -- JSON transaction data
    created_at TIMESTAMP NOT NULL,           -- event creation time
    uniqueness_key VARCHAR(64) NOT NULL UNIQUE  -- deduplication key (SHA256)
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

## Environment Variables

The application uses the following environment variables:

- `DATABASE_DSN`: PostgreSQL connection string (required)
  - Format: `postgres://user:password@host:port/dbname?sslmode=disable`
  - Example: `postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable`

You can set these variables in a `.env` file (see `.env.example` for template) or export them in your shell.

## Development

### Running with Docker

The easiest way to develop is using Docker for the PostgreSQL database:

```bash
# Start the database
make docker-up

# Set the environment variable
export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"

# Run the application
make run

# Run tests
make test

# Stop the database when done
make docker-down
```

### Migrations

Database migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate). Migrations are automatically applied when the application connects to the database.

Migration files are located in the `migrations/` directory.

## Troubleshooting

### Database Connection Issues

If you encounter connection issues:
1. Ensure Docker is running: `docker ps`
2. Check if PostgreSQL container is healthy: `docker compose ps`
3. Verify DATABASE_DSN is set correctly: `echo $DATABASE_DSN`
4. Check PostgreSQL logs: `docker compose logs postgres`

### Test Failures

If tests fail:
1. Ensure the database is running: `make docker-up`
2. Set DATABASE_DSN environment variable
3. Clear the database: `make docker-clean && make docker-up`
