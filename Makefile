.PHONY: test test-short test-coverage build build-gui build-kg clean clean-all sonarqube help
.PHONY: install install-agent uninstall uninstall-agent
.PHONY: start-server start-gui start-all stop-all
.PHONY: setup-launchd uninstall-launchd status-launchd setup-kuzu

# Default target
.DEFAULT_GOAL := help

# Directories
PROJECT_ROOT := $(shell pwd)
LAUNCHD_DIR := $(HOME)/Library/LaunchAgents
GUI_DIR := gui

# Kuzu configuration
KUZU_VERSION := 0.8.2
GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
PLATFORM := $(GOOS)-$(GOARCH)
KUZU_DIR := lib/kuzu/$(PLATFORM)

ifeq ($(GOOS),darwin)
  CXX_LIBS := -lc++
else
  CXX_LIBS := -lstdc++
endif

CGO_CFLAGS_KG  := -I$(abspath $(KUZU_DIR)/include)
CGO_LDFLAGS_KG := -L$(abspath $(KUZU_DIR)) -lkuzu $(CXX_LIBS) -lm -ldl -lpthread

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

build: build-agent build-server ## Build agent and agent-server binaries
	@echo "✅ Binaries built in bin/"

build-agent: ## Build the agent binary
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/agent ./cmd/agent

build-server: ## Build the agent-server binary
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/agent-server ./cmd/server

build-kg: ## Build the kg binary (requires: make setup-kuzu first)
	@mkdir -p bin
	CGO_ENABLED=1 \
	CGO_CFLAGS="$(CGO_CFLAGS_KG)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS_KG)" \
	go build -o bin/kg ./cmd/kg

setup-kuzu: ## Download Kuzu static library for current platform
	@bash scripts/download-kuzu.sh $(KUZU_VERSION) $(PLATFORM)

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
	@install -m 755 bin/agent /usr/local/bin/agent
	@install -m 755 bin/agent-server /usr/local/bin/agent-server
	@echo "✅ Agent binaries installed to /usr/local/bin"

uninstall: uninstall-agent ## Uninstall binaries from /usr/local/bin (run with: sudo make uninstall)

uninstall-agent: ## Uninstall agent binaries from /usr/local/bin
	@echo "Uninstalling agent binaries..."
	@rm -f /usr/local/bin/agent /usr/local/bin/agent-server /usr/local/bin/kg
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
	@echo "Running Go tests..."
	@go test ./... -v

test-short: ## Run tests in short mode (skip slow tests)
	@echo "Running Go tests (short mode)..."
	@go test ./... -short

test-coverage: ## Run tests with coverage report
	@echo "Running Go tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-gui: ## Run GUI tests
	@echo "Running GUI tests..."
	@cd $(GUI_DIR) && npm test

# ============================================================================
# CLEAN TARGETS
# ============================================================================

clean: ## Clean build artifacts and test caches
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -rf bin/ coverage.out coverage.html .scannerwork/
	@echo "✅ Clean complete"

clean-gui: ## Clean GUI build artifacts
	@echo "Cleaning GUI..."
	@cd $(GUI_DIR) && rm -rf dist/ node_modules/
	@echo "✅ GUI cleaned"

clean-all: clean clean-gui ## Clean everything (agent + GUI)

# ============================================================================
# CODE QUALITY TARGETS
# ============================================================================

sonarqube: ## Run SonarQube analysis
	@echo "Running SonarQube analysis..."
	@python3 scripts/validate-with-sonarqube.py . --format text

sonarqube-json: ## Run SonarQube analysis with JSON output
	@echo "Running SonarQube analysis (JSON)..."
	@python3 scripts/validate-with-sonarqube.py . --format json

lint: ## Run Go linters
	@echo "Running Go linters..."
	@go vet ./...
	@go fmt ./...

lint-gui: ## Run GUI linters
	@echo "Running GUI linters..."
	@cd $(GUI_DIR) && npm run lint

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...
