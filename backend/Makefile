.PHONY: build run dev test clean migrate migrate-down deps tidy

# Binary name
BINARY=domainsnapshot

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
build:
	$(GOBUILD) -o $(BINARY) ./cmd/server

# Run the application
run: build
	./$(BINARY)

# Run in development mode (no build)
dev:
	$(GORUN) ./cmd/server

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOGET) -v ./...

# Tidy go.mod
tidy:
	$(GOMOD) tidy

# Create database (requires psql)
db-create:
	psql -U postgres -c "CREATE DATABASE domainsnapshot;"

# Drop database (requires psql)
db-drop:
	psql -U postgres -c "DROP DATABASE IF EXISTS domainsnapshot;"

# Run database migrations
migrate:
	psql -d domainsnapshot -f internal/database/migrations/001_initial_schema.up.sql

# Rollback database migrations
migrate-down:
	psql -d domainsnapshot -f internal/database/migrations/001_initial_schema.down.sql

# Reset database (drop, create, migrate)
db-reset: db-drop db-create migrate

# Show help
help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Build and run the application"
	@echo "  make dev           - Run without building (development)"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make tidy          - Tidy go.mod"
	@echo "  make db-create     - Create PostgreSQL database"
	@echo "  make db-drop       - Drop PostgreSQL database"
	@echo "  make migrate       - Run database migrations"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make db-reset      - Reset database (drop, create, migrate)"
