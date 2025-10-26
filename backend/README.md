# Backend Go Service

Go backend service for the OCR web application.

## Structure

```
backend/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and queries
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── services/        # Business logic
│   └── client/          # External service clients
└── pkg/
    ├── logger/          # Logging utilities
    ├── validator/       # Input validation
    ├── utils/           # Helper functions
    └── errors/          # Custom error types
```

## Development

### Prerequisites
- Go 1.21+
- PostgreSQL 16+
- Redis 7+

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Run the server:
```bash
go run cmd/server/main.go
```

3. Run tests:
```bash
go test -v ./...
```

### Build

```bash
go build -o bin/server cmd/server/main.go
```

## API Endpoints

See PROJECT_PLAN.md for full API documentation.
