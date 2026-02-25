# Internal Collaboration API - Makefile

.PHONY: help dev build test seed clear-db migrate clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make dev        - Run development server with hot reload (Air)"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make seed       - Seed database with sample data"
	@echo "  make clear-db   - Clear all data from database (with confirmation)"
	@echo "  make reset-db   - Clear database and re-seed"
	@echo "  make migrate    - Run database migrations"
	@echo "  make clean      - Clean build artifacts"

# Run development server
dev:
	@echo "Starting development server with Air..."
	air

# Build application
build:
	@echo "Building application..."
	go build -o tmp/main.exe ./cmd/main.go
	@echo "✅ Build complete: tmp/main.exe"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Seed database
seed:
	@echo "Seeding database..."
	go run scripts/seed/main.go

# Clear database
clear-db:
	@echo "Clearing database..."
	go run scripts/clear/main.go

# Reset database (clear + seed)
reset-db: clear-db seed
	@echo "✅ Database reset complete"

# Run migrations
migrate:
	@echo "Running migrations..."
	go run cmd/main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf tmp/
	@echo "✅ Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"
