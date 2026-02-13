# go-portfolio

A portfolio application built with Go.

## Project Structure

```
.
├── cmd/                      # Command-line applications
│   └── main/                 # Main application entry point
├── internal/                 # Private application code
│   └── portfolio/            # Portfolio service logic
├── pkg/                      # Public packages (reusable code)
│   └── models/               # Data models
├── api/                      # API definitions (OpenAPI, gRPC, etc.)
├── web/                      # Web assets (frontend)
├── configs/                  # Configuration files
├── migrations/               # Database migrations
├── scripts/                  # Utility scripts
├── docs/                     # Documentation
├── tests/                    # Integration tests
├── go.mod                    # Module definition
└── go.sum                    # Module checksums
```

## Getting Started

```bash
go build -o bin/portfolio ./cmd/main
go run ./cmd/main
```
