.PHONY: build run test clean help docker-up docker-down docker-clean db-setup

help:
	@echo "Available targets:"
	@echo "  make build        - Build all binaries"
	@echo "  make run          - Run the import tool"
	@echo "  make run-server   - Run the HTTP server"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make docker-up    - Start PostgreSQL database with Docker"
	@echo "  make docker-down  - Stop PostgreSQL database"
	@echo "  make docker-clean - Stop and remove PostgreSQL database and volumes"
	@echo "  make db-setup     - Start database and set up environment for tests"

build:
	go build -o bin/portfolio-import ./cmd/import
	go build -o bin/portfolio-server ./cmd/server

run:
	go run ./cmd/import

run-server:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin/
	go clean

docker-up:
	docker compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@docker compose exec -T postgres pg_isready -U portfolio -d go-portfolio || sleep 2
	@echo "PostgreSQL is ready!"

docker-down:
	docker compose down

docker-clean:
	docker compose down -v
	docker volume rm go-portfolio_postgres_data 2>/dev/null || true

db-setup: docker-up
	@echo "Database is ready for testing!"
	@echo "Set DATABASE_DSN environment variable:"
	@echo "export DATABASE_DSN=postgres://portfolio:portfolio@localhost:5432/go-portfolio?sslmode=disable"
