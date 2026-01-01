.PHONY: all build test clean run docker docker-run lint fmt help

# Variables
BINARY_NAME=mcp-calculator
DOCKER_IMAGE=mcp-calculator
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/mcp-calculator

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run the server locally
run: build
	@echo "Starting server..."
	./bin/$(BINARY_NAME)

# Run in development mode with auto-reload
dev:
	@echo "Starting in development mode..."
	MCP_LOG_LEVEL=debug go run ./cmd/mcp-calculator

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofumpt -l -w .

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

# Run Docker container
docker-run: docker
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):latest

# Run with docker-compose
compose-up:
	docker-compose up --build

compose-down:
	docker-compose down

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Install development tools
tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate go.sum
mod:
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Run tests and build"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Build and run the server"
	@echo "  dev           - Run in development mode"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  docker        - Build Docker image"
	@echo "  docker-run    - Build and run Docker container"
	@echo "  compose-up    - Start with docker-compose"
	@echo "  compose-down  - Stop docker-compose"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  tools         - Install development tools"
	@echo "  help          - Show this help"
