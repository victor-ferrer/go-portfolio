# go-portfolio

An event-sourced portfolio management system built with Go. This application tracks financial transactions from multiple brokers, stores them as immutable events, and projects portfolio positions and performance metrics.

## Features

- **Event Sourcing**: Single source of truth using immutable events
- **Multi-Broker Support**: Parse and import transactions from multiple brokers
- **Automatic Deduplication**: Uniqueness key prevents duplicate transaction imports
- **Position Tracking**: Calculate open positions and cost basis per instrument
- **Performance Metrics**: Compute annualized returns and portfolio-wide metrics
- **PostgreSQL Storage**: Production-ready database with Docker support
- **HTTP API**: Query transactions via a RESTful HTTP server

## Architecture

The system follows event sourcing principles:
- Transactions are parsed and stored as immutable events
- Each event has a uniqueness key (hash of date, instrument, category, quantity, amount)
- Positions and metrics are calculated on-the-fly from event stream
- Database enforces idempotency via UNIQUE constraint on uniqueness keys

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
- CLI tool for importing transactions (`portfolio-import`)
- HTTP API server with transaction querying (`portfolio-server`)
- Transaction repository with filtering support

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Node.js 18+ and npm (for building the web UI)
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
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=go-portfolio
   export DB_USER=portfolio
   export DB_PASSWORD=portfolio
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
# Install web UI dependencies (first time only)
make web-install

make build
```

This produces two binaries in `bin/`:
- `bin/portfolio-import` — import transaction CSV files
- `bin/portfolio-server` — start the HTTP API server

### Run

**Import transactions:**
```bash
make run
```

**Start the HTTP server:**
```bash
make run-server
```

## Usage

### Import Transactions

The `portfolio-import` tool allows you to import transaction data from broker CSV files.

**Basic Usage:**
```bash
./bin/portfolio-import --file-name <path-to-csv> --broker <broker-name>
```

**Example:**
```bash
# Set the database environment variables
export DB_HOST=localhost DB_PORT=5432 DB_NAME=go-portfolio DB_USER=portfolio DB_PASSWORD=portfolio

# Import transactions from Click Trade CSV file
./bin/portfolio-import --file-name ./data/click-trade-transactions.csv --broker click-trade
```

**Using Docker:**

Place your CSV files in the `./data` directory, then run the import service via Docker Compose:

```bash
# Create the data directory if it doesn't exist
mkdir -p data

# Copy your CSV file into it
cp /path/to/your/transactions.csv data/

# Start the database (if not already running)
make docker-up

# Run the import container, passing the file path inside /data
docker compose --profile import run --rm import \
  --file-name /data/transactions.csv \
  --broker click-trade
```

The `./data` directory on your host is mounted to `/data` inside the container, so any CSV files placed there are accessible to the import tool.

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

### HTTP API Server

The `portfolio-server` starts an HTTP server for querying stored transactions.

**Start the server:**
```bash
export DB_HOST=localhost DB_PORT=5432 DB_NAME=go-portfolio DB_USER=portfolio DB_PASSWORD=portfolio
./bin/portfolio-server --addr :8080
```

The `--addr` flag is optional and defaults to `:8080`.

#### Endpoints

**`GET /transactions`** — List all transactions with optional filters.

| Query Parameter | Type   | Description                                      |
|-----------------|--------|--------------------------------------------------|
| `instrument`    | string | Filter by instrument name (e.g. `AAPL`)          |
| `broker`        | string | Filter by broker (e.g. `click-trade`)            |
| `type`          | string | Filter by transaction type (`buy` or `sell`)     |
| `from`          | string | Start date inclusive, format `YYYY-MM-DD`        |
| `to`            | string | End date inclusive, format `YYYY-MM-DD`          |

**Examples:**
```bash
# Get all transactions
curl http://localhost:8080/transactions

# Filter by instrument
curl "http://localhost:8080/transactions?instrument=AAPL"

# Filter by broker and date range
curl "http://localhost:8080/transactions?broker=click-trade&from=2024-01-01&to=2024-12-31"

# Filter by type
curl "http://localhost:8080/transactions?type=buy"
```

**Response format:**
```json
[
  {
    "instrument": "AAPL",
    "type": "buy",
    "category": "Trade",
    "amount": 1500.00,
    "quantity": 10,
    "currency": "USD",
    "createdAt": "2024-06-15T00:00:00Z"
  }
]
```

### Run Tests

```bash
# Start the database first
make docker-up

# Set the database environment variables
export DB_HOST=localhost DB_PORT=5432 DB_NAME=go-portfolio DB_USER=portfolio DB_PASSWORD=portfolio

# Run tests
make test
```

### Docker Commands

- `make docker-up`: Start PostgreSQL database
- `make docker-down`: Stop PostgreSQL database
- `make docker-clean`: Stop and remove database with volumes
- `make db-setup`: Start database and display connection instructions
- `docker compose --profile import run --rm import --file-name /data/<file> --broker <broker>`: Run the import tool in Docker (mounts `./data` as `/data`)

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

- `DB_HOST`: PostgreSQL host (required)
- `DB_PORT`: PostgreSQL port (required)
- `DB_NAME`: PostgreSQL database name (required)
- `DB_USER`: PostgreSQL username (required)
- `DB_PASSWORD`: PostgreSQL password (required)
- `DB_SSLMODE`: SSL mode (optional, defaults to `disable`)
- `MIGRATIONS_PATH`: Path to migration files (optional, defaults to `file://./migrations`)

You can set these variables in a `.env` file (see `.env.example` for template) or export them in your shell.

## Development

### Running with Docker

The easiest way to develop is using Docker for the PostgreSQL database:

```bash
# Start the database
make docker-up

# Set the database environment variables
export DB_HOST=localhost DB_PORT=5432 DB_NAME=go-portfolio DB_USER=portfolio DB_PASSWORD=portfolio

# Run the import tool
make run

# Run the HTTP server
make run-server

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
3. Verify `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` are set correctly
4. Check PostgreSQL logs: `docker compose logs postgres`

### Test Failures

If tests fail:
1. Ensure the database is running: `make docker-up`
2. Set the `DB_*` environment variables (see [Environment Variables](#environment-variables))
3. Clear the database: `make docker-clean && make docker-up`
