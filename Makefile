.PHONY: test test-short test-coverage build clean sonarqube help

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests
	@echo "Running Go tests in a2a-agent..."
	@cd a2a-agent && go test ./... -v

test-short: ## Run tests in short mode (skip slow tests)
	@echo "Running Go tests (short mode) in a2a-agent..."
	@cd a2a-agent && go test ./... -short

test-coverage: ## Run tests with coverage report
	@echo "Running Go tests with coverage..."
	@cd a2a-agent && go test ./... -coverprofile=coverage.out
	@cd a2a-agent && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: a2a-agent/coverage.html"

build: ## Build a2a-agent binaries
	@echo "Building a2a-agent..."
	@cd a2a-agent && go build -o bin/agent-server ./cmd/agent-server
	@cd a2a-agent && go build -o bin/agent ./cmd/agent
	@echo "Binaries built in a2a-agent/bin/"

clean: ## Clean build artifacts and test caches
	@echo "Cleaning..."
	@cd a2a-agent && go clean
	@cd a2a-agent && rm -rf bin/ coverage.out coverage.html
	@rm -rf .scannerwork/
	@echo "Clean complete"

sonarqube: ## Run SonarQube analysis on a2a-agent
	@echo "Running SonarQube analysis..."
	@python3 scripts/validate-with-sonarqube.py a2a-agent --format text

sonarqube-json: ## Run SonarQube analysis with JSON output
	@echo "Running SonarQube analysis (JSON)..."
	@python3 scripts/validate-with-sonarqube.py a2a-agent --format json

lint: ## Run Go linters
	@echo "Running Go linters..."
	@cd a2a-agent && go vet ./...
	@cd a2a-agent && go fmt ./...

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@cd a2a-agent && go fmt ./...
