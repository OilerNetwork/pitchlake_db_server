# Pitchlake WebSocket Server - Development Commands

.PHONY: help test test-unit test-integration test-coverage test-verbose test-race build run clean

# Default target
help:
	@echo "Pitchlake WebSocket Server - Available Commands:"
	@echo ""
	@echo "Testing:"
	@echo "  test            Run all tests"
	@echo "  test-unit       Run unit tests only (fast)"
	@echo "  test-integration Run integration tests only"
	@echo "  test-coverage   Run tests with coverage report"
	@echo "  test-verbose    Run tests with verbose output"
	@echo "  test-race       Run tests with race detection"
	@echo ""
	@echo "Development:"
	@echo "  build           Build the application"
	@echo "  run             Run the application"
	@echo "  clean           Clean build artifacts"
	@echo ""

# Testing commands
test:
	@echo "Running all tests..."
	go test ./...

test-unit:
	@echo "Running unit tests (fast)..."
	go test ./server/api/... ./server/validations/...

test-integration:
	@echo "Running integration tests..."
	go test ./server/...

test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...

test-verbose:
	@echo "Running tests with verbose output..."
	go test ./... -v

test-race:
	@echo "Running tests with race detection..."
	go test ./... -race

# Development commands
build:
	@echo "Building application..."
	go build -o pitchlake-backend .

run: build
	@echo "Starting server..."
	@export DB_URL="postgres://pitchlake:pitchlake@localhost:5433/pitchlake?sslmode=disable" && ./pitchlake-backend

clean:
	@echo "Cleaning build artifacts..."
	rm -f pitchlake-backend
