.PHONY: help build run test clean docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  make build        - Build the backend server"
	@echo "  make run          - Run the backend server"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"

# Build the backend
build:
	@echo "Building backend..."
	cd backend && go build -o ../server ./cmd/server

# Run the backend
run:
	@echo "Running backend..."
	cd backend && go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	cd backend && go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd backend && go test -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f server
	cd backend && go clean

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	cd backend && go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	cd backend && go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	cd backend && golangci-lint run

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t go_infra_link:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 go_infra_link:latest

# Development: run with auto-reload (requires air)
dev:
	@echo "Starting development server with auto-reload..."
	cd backend && air
