.PHONY: test test-short test-coverage build build-gui clean clean-all sonarqube help
.PHONY: install install-agent uninstall uninstall-agent
.PHONY: start-server start-gui start-all stop-all
.PHONY: setup-launchd uninstall-launchd status-launchd

# Default target
.DEFAULT_GOAL := help

# Directories
PROJECT_ROOT := $(shell pwd)
LAUNCHD_DIR := $(HOME)/Library/LaunchAgents
GUI_DIR := gui

help: ## Show this help message
	@echo "AI-Pack Build System"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start:"
	@echo "  make build install        # Build and install everything"
	@echo "  make start-all            # Start server and GUI"
	@echo "  make setup-launchd        # Setup auto-start with launchd"

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: ## Build a2a-agent binaries
	@echo "Building a2a-agent..."
	@cd a2a-agent && $(MAKE) build
	@echo "✅ Binaries built in a2a-agent/bin/"

build-gui: ## Build GUI for production
	@echo "Building GUI..."
	@if [ ! -d "$(GUI_DIR)/node_modules" ]; then \
		echo "Installing GUI dependencies..."; \
		cd $(GUI_DIR) && npm install; \
	fi
	@cd $(GUI_DIR) && npm run build
	@echo "✅ GUI built in $(GUI_DIR)/dist/"

build-all: build build-gui ## Build everything (agent + GUI)

# ============================================================================
# INSTALL TARGETS
# ============================================================================

install: install-agent ## Install binaries to /usr/local/bin (run with: sudo make install)

install-agent: build ## Install agent binaries to /usr/local/bin
	@echo "Installing agent binaries..."
	@cd a2a-agent && $(MAKE) install
	@echo "✅ Agent binaries installed to /usr/local/bin"

uninstall: uninstall-agent ## Uninstall binaries from /usr/local/bin (run with: sudo make uninstall)

uninstall-agent: ## Uninstall agent binaries from /usr/local/bin
	@echo "Uninstalling agent binaries..."
	@cd a2a-agent && $(MAKE) uninstall
	@echo "✅ Agent binaries uninstalled"

# ============================================================================
# START/STOP TARGETS
# ============================================================================

start-server: ## Start the agent server (foreground)
	@echo "Starting AI-Pack Agent Server..."
	@python3 scripts/start-all.py --server-only

start-gui: ## Start the GUI dev server (foreground)
	@echo "Starting GUI..."
	@python3 scripts/start-all.py --gui-only

start-all: ## Start both agent server and GUI (foreground)
	@echo "Starting AI-Pack (Server + GUI)..."
	@python3 scripts/start-all.py

stop-all: ## Stop all running services
	@echo "Stopping services..."
	@pkill -f "agent-server" || true
	@pkill -f "vite.*gui" || true
	@echo "✅ Services stopped"

# ============================================================================
# LAUNCHD TARGETS (macOS)
# ============================================================================

setup-launchd: ## Install launchd plists for auto-start (macOS)
	@echo "Installing launchd configuration..."
	@python3 scripts/setup-launchd.py install
	@echo ""
	@echo "✅ launchd setup complete"
	@echo ""
	@echo "Services will start automatically on login."
	@echo "To start now: make status-launchd"

uninstall-launchd: ## Uninstall launchd plists (macOS)
	@echo "Uninstalling launchd configuration..."
	@python3 scripts/setup-launchd.py uninstall
	@echo "✅ launchd configuration removed"

status-launchd: ## Show launchd service status (macOS)
	@python3 scripts/setup-launchd.py status

# ============================================================================
# TEST TARGETS
# ============================================================================

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

test-gui: ## Run GUI tests
	@echo "Running GUI tests..."
	@cd $(GUI_DIR) && npm test

# ============================================================================
# CLEAN TARGETS
# ============================================================================

clean: ## Clean build artifacts and test caches
	@echo "Cleaning a2a-agent..."
	@cd a2a-agent && go clean
	@cd a2a-agent && rm -rf bin/ coverage.out coverage.html
	@rm -rf .scannerwork/
	@echo "✅ Clean complete"

clean-gui: ## Clean GUI build artifacts
	@echo "Cleaning GUI..."
	@cd $(GUI_DIR) && rm -rf dist/ node_modules/
	@echo "✅ GUI cleaned"

clean-all: clean clean-gui ## Clean everything (agent + GUI)

# ============================================================================
# CODE QUALITY TARGETS
# ============================================================================

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

lint-gui: ## Run GUI linters
	@echo "Running GUI linters..."
	@cd $(GUI_DIR) && npm run lint

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@cd a2a-agent && go fmt ./...
