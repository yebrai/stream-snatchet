.PHONY: build clean test run install deps lint format

# Build variables
BINARY_NAME=stream-snatchet
BUILD_DIR=build
MAIN_PATH=cmd/stream-snatchet/main.go

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)

# Run with GUI
run-gui:
	@echo "Running $(BINARY_NAME) with GUI..."
	go run $(MAIN_PATH) --gui

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		@echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run quick development checks
dev-check: format test lint

# Release build (optimized)
release:
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  run          - Run the application"
	@echo "  run-gui      - Run the application with GUI"
	@echo "  install      - Install the application"
	@echo "  format       - Format code"
	@echo "  lint         - Lint code"
	@echo "  dev-check    - Run format, test, and lint"
	@echo "  release      - Build optimized release version"
	@echo "  help         - Show this help message"