.PHONY: help project format check test ci build clean
.DEFAULT_GOAL := help

BUILD_DIR=./bin/

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
	@go build -o $(BUILD_DIR) ./cmd/...

clean: ## Clean up project
	@echo "Cleaning up project..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
