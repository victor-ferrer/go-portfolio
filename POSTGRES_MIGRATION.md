# PostgreSQL Migration Summary

This document summarizes the migration from SQLite (in-memory) to PostgreSQL with Docker.

## Changes Made

### 1. Docker Setup
- **docker-compose.yml**: PostgreSQL 16 Alpine container with volume persistence
- **scripts/init-db.sh**: Database initialization script (runs on first container start)
- **.env.example**: Template for environment variables

### 2. Database Changes
- **Migration files**: Updated to use PostgreSQL-compatible types
  - `TEXT` → `VARCHAR(255)` for string columns
  - `TEXT` → `JSONB` for JSON payload storage
  - `TEXT` → `VARCHAR(64)` for SHA256 uniqueness key
- **New PostgreSQL event store**: `internal/store/postgres_eventstore.go`
  - Connection via environment variable `DATABASE_DSN`
  - Automatic migration on startup
  - PostgreSQL-specific error handling for unique constraint violations

### 3. Test Infrastructure
- **New test file**: `internal/store/postgres_eventstore_test.go`
  - Tests require `DATABASE_DSN` environment variable
  - Tests skip if DATABASE_DSN is not set
  - Automatic cleanup before each test (DELETE FROM events)
- **Old SQLite files removed**:
  - `internal/store/eventstore_sqlite_old.go` (deleted)
  - `internal/store/eventstore_test_sqlite_old.go` (deleted)

### 4. Dependencies
- **Added**:
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/golang-migrate/migrate/v4/database/postgres` - Migration support
- **Removed**:
  - `github.com/mattn/go-sqlite3` - SQLite driver (no longer needed)

### 5. Documentation
- **README.md**: Updated with Docker setup instructions
- **Makefile**: New targets for Docker management
  - `make docker-up` - Start PostgreSQL
  - `make docker-down` - Stop PostgreSQL
  - `make docker-clean` - Remove PostgreSQL and volumes
  - `make db-setup` - Setup database for development/testing
- **CI/CD**: GitHub Actions workflow with PostgreSQL service

## How to Use

### Development Setup
```bash
# 1. Start PostgreSQL database
make docker-up

# 2. Set environment variable
export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"

# 3. Run tests
make test

# 4. Build application
make build

# 5. Run application
make run
```

### Running Tests
```bash
# Start database first
make db-setup

# Set environment variable
export DATABASE_DSN="postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"

# Run tests
go test ./...
```

### CI/CD
GitHub Actions workflow automatically:
1. Starts PostgreSQL service container
2. Runs migrations
3. Executes all tests
4. Builds application

## Database Schema

```sql
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    broker VARCHAR(255) NOT NULL,
    imported_at TIMESTAMP NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    uniqueness_key VARCHAR(64) NOT NULL UNIQUE
);

CREATE INDEX idx_aggregate_id ON events(aggregate_id);
CREATE INDEX idx_broker ON events(broker);
CREATE INDEX idx_created_at ON events(created_at);
```

## Benefits of PostgreSQL

1. **Production-Ready**: PostgreSQL is a robust, production-grade database
2. **JSONB Support**: Native JSON binary format for better performance
3. **Better Tooling**: Rich ecosystem of tools and extensions
4. **Scalability**: Can handle large datasets efficiently
5. **Docker Integration**: Easy to deploy and manage in containers
6. **Testing**: Consistent test environment across development and CI

## Migration Notes

- **No data migration needed**: Repository was using in-memory SQLite (no persistent data)
- **Breaking Change**: Application now requires PostgreSQL and DATABASE_DSN environment variable
- **Backward Compatibility**: None - SQLite support completely removed
- **Testing**: All existing tests pass with PostgreSQL
