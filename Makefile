.PHONY: help project format check test ci build clean
.DEFAULT_GOAL := help

VERSION=$(shell git describe --tags --always)
BUILD_DIR=./bin/
BINARY_NAME=piphos

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

project: ## Setup project
	@echo "Setting up project..."
	@go mod tidy
	@go mod download
	@go mod verify

format: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

check: ## Lint code
	@echo "Linting code..."
	@go vet ./...

test: ## Test code
	@echo "Running tests..."
	@go test -race -cover ./...

ci: project format check test ## Run CI checks locally
	@echo "CI checks completed."

build: ## Build binaries
	@echo "Building binaries..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-linux-amd64 ./cmd/...
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-linux-arm64 ./cmd/...
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-darwin-amd64 ./cmd/...
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-darwin-arm64 ./cmd/...
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-windows-amd64.exe ./cmd/...
	@GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)$(BINARY_NAME)-$(VERSION)-windows-arm64.exe ./cmd/...

clean: ## Clean up project
	@echo "Cleaning up project..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
