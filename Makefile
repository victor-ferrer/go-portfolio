.PHONY: build run test clean help

help:
	@echo "Available targets:"
	@echo "  make build   - Build the application"
	@echo "  make run     - Run the application"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Remove build artifacts"

build:
	go build -o bin/portfolio ./cmd/main

run:
	go run ./cmd/main

test:
	go test ./...

clean:
	rm -rf bin/
	go clean
