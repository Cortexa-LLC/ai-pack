.PHONY: test test-short test-coverage build build-gui codegen-gui clean clean-all sonarqube help
.PHONY: install install-agent install-agent-mcp uninstall uninstall-agent setup-mcp install-mcp
.PHONY: start-server start-gui start-all stop-server stop-gui stop-all restart-server restart-gui restart-all
.PHONY: setup-services uninstall-services status-services
.PHONY: setup-launchd uninstall-launchd status-launchd
.PHONY: bootstrap
UNAME_S := $(shell uname -s)

# Default target
.DEFAULT_GOAL := help

# Directories
PROJECT_ROOT := $(shell pwd)
LAUNCHD_DIR := $(HOME)/Library/LaunchAgents
GUI_DIR := gui

# CGO is required by go-kuzu (bundled dynamic lib) and go-tree-sitter.
# No external kuzu download needed — go-kuzu v0.11.3+ bundles the library.
CGO := CGO_ENABLED=1

# Version info injected at build time into all binaries via ldflags.
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

MODULE := github.com/cortexa-llc/ai-pack

LDFLAGS_AGENT  := -ldflags "-X main.Version=$(VERSION) \
                             -X main.Commit=$(COMMIT) \
                             -X main.BuildTime=$(BUILD_TIME)"

help: ## Show this help message
	@echo "AI-Pack Build System"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start (fresh clone):"
	@echo "  make bootstrap            # Install all tool dependencies (run once)"
	@echo "  make build install        # Build and install binaries to /usr/local/bin"
	@echo "  make setup-mcp            # Register MCP servers in Claude Code"
	@echo "  make setup-services       # Setup auto-start services (OPTIONAL but recommended)"
	@echo "  kg index                  # Index codebase into knowledge graph"
	@echo ""
	@echo "📦 See INSTALL.md for complete installation and troubleshooting guide"

# ============================================================================
# BOOTSTRAP (first-time environment setup)
# ============================================================================

bootstrap: ## Install all tool dependencies (Go modules, npm, seeds) — run once after clone
	@echo "Bootstrapping AI-Pack development environment..."
	@echo ""
	@echo "→ Downloading Go modules..."
	@go mod download
	@echo "✅ Go modules ready"
	@echo ""
	@echo "→ Installing GUI npm dependencies..."
	@cd $(GUI_DIR) && npm install
	@echo "✅ GUI dependencies installed"
	@echo ""
	@echo "→ Seeding model performance grades (LiveBench)..."
	@python3 scripts/seed-grades.py
	@echo "✅ Performance grades seeded"
	@echo ""
	@echo "✅ Bootstrap complete. Next steps:"
	@echo "   make build install    # Build and install binaries"
	@echo "   make setup-mcp        # Register MCP servers in Claude Code"
	@echo "   make setup-services   # Setup auto-start (macOS/Linux)"

# ============================================================================
# BUILD TARGETS
# ============================================================================

build: build-agent build-server ## Build all binaries
	@echo "✅ Binaries built in bin/"

build-agent: ## Build the agent CLI (requires CGO for SQLite)
	@mkdir -p bin
	$(CGO) go build $(LDFLAGS_AGENT) -o bin/agent ./cmd/agent

build-server: ## Build the agent-server (CGO, go-kuzu bundled)
	@mkdir -p bin
	$(CGO) go build -o bin/agent-server ./cmd/server


codegen-gui: ## Regenerate GraphQL TypeScript types from schema
	@echo "Generating GraphQL types..."
	@if [ ! -d "$(GUI_DIR)/node_modules" ]; then \
		echo "Installing GUI dependencies..."; \
		cd $(GUI_DIR) && npm install; \
	fi
	@cd $(GUI_DIR) && npm run codegen
	@echo "✅ GraphQL types generated in $(GUI_DIR)/src/types/graphql-types.ts"

build-gui: codegen-gui ## Build GUI for production
	@echo "Building GUI..."
	@if [ ! -d "$(GUI_DIR)/node_modules" ]; then \
		echo "Installing GUI dependencies..."; \
		cd $(GUI_DIR) && npm install; \
	fi
	@cd $(GUI_DIR) && npm run build
	@echo "✅ GUI built in $(GUI_DIR)/dist/"

build-all: build build-gui build-agent-mcp ## Build everything (agent + GUI + agent-mcp)

# ============================================================================
# INSTALL TARGETS
# ============================================================================

install: install-agent install-agent-mcp ## Install all binaries to /usr/local/bin, then run: make setup-mcp

install-agent: build-agent build-server ## Install agent binaries to /usr/local/bin
	@echo "Installing agent binaries..."
	@install -m 755 bin/agent /usr/local/bin/agent
	@install -m 755 bin/agent-server /usr/local/bin/agent-server
	@echo "✅ Agent binaries installed to /usr/local/bin"
	@echo ""
	@echo "Initializing configuration..."
	@if [ -n "$$SUDO_USER" ]; then \
		sudo -u "$$SUDO_USER" python3 scripts/init-config.py; \
	else \
		python3 scripts/init-config.py; \
	fi
	@echo "✅ Configuration ready at ~/.ai-pack/config.json"

build-agent-mcp: ## Build agent-mcp MCP stdio server (CGO_ENABLED=0)
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/agent-mcp ./cmd/agent-mcp

install-agent-mcp: build-agent-mcp ## Install agent-mcp binary to /usr/local/bin
	@echo "Installing agent-mcp..."
	@install -m 755 bin/agent-mcp /usr/local/bin/agent-mcp
	@echo "✅ agent-mcp installed to /usr/local/bin/agent-mcp"

install-mcp: install-agent-mcp ## Build, install agent-mcp, register in ~/.claude.json, and copy orchestrate skill
	@echo "Registering agent-mcp in ~/.claude.json..."
	@python3 scripts/register-mcp.py
	@echo ""
	@echo "Installing orchestrate skill..."
	@mkdir -p ~/.claude/skills
	@cp config/skills/orchestrate.md ~/.claude/skills/orchestrate.md
	@echo "✅ ~/.claude/skills/orchestrate.md installed"
	@echo ""
	@echo "✅ install-mcp complete. Restart Claude Code to pick up the changes."

setup-mcp: ## Register AI-Pack MCP servers globally (~/.claude/settings.json)
	@python3 scripts/setup-mcp.py

setup-mcp-local: ## Register AI-Pack MCP servers in project-local .claude/settings.local.json
	@python3 scripts/setup-mcp.py --local

uninstall: uninstall-agent ## Uninstall agent binaries from /usr/local/bin (run with: sudo make uninstall)

uninstall-agent: ## Uninstall agent binaries from /usr/local/bin
	@echo "Uninstalling agent binaries..."
	@rm -f /usr/local/bin/agent /usr/local/bin/agent-server
	@echo "✅ Agent binaries uninstalled"

# ============================================================================
# START/STOP TARGETS (via launchctl on macOS, systemd on Linux)
# Run 'make setup-services' first to install the background services.
# Use 'make run-*' for foreground mode (no service manager needed).
# ============================================================================

start-server: ## Start agent server via service manager
	@python3 scripts/setup-services.py start-server

start-gui: ## Start GUI via service manager
	@python3 scripts/setup-services.py start-gui

start-all: start-server start-gui ## Start both services

stop-server: ## Stop agent server
	@python3 scripts/setup-services.py stop-server

stop-gui: ## Stop GUI
	@python3 scripts/setup-services.py stop-gui

stop-all: stop-server stop-gui ## Stop all services

restart-server: stop-server start-server ## Restart agent server

restart-gui: stop-gui start-gui ## Restart GUI

restart-all: stop-all start-all ## Restart all services

run-server: ## Run agent server in foreground (no service manager)
	@echo "Starting AI-Pack Agent Server (foreground)..."
	@python3 scripts/start-all.py --server-only

run-gui: ## Run GUI in foreground (no service manager)
	@echo "Starting GUI (foreground)..."
	@python3 scripts/start-all.py --gui-only

run-all: ## Run both services in foreground (no service manager)
	@echo "Starting AI-Pack (foreground)..."
	@python3 scripts/start-all.py

# ============================================================================
# SERVICE MANAGEMENT (macOS launchd / Linux systemd)
# ============================================================================

setup-services: ## Install and enable background services (macOS/Linux)
	@python3 scripts/setup-services.py install

uninstall-services: ## Remove background services
	@python3 scripts/setup-services.py uninstall

status-services: ## Show service status
	@python3 scripts/setup-services.py status

# Backwards-compat aliases
setup-launchd: setup-services ## Alias for setup-services (macOS)
uninstall-launchd: uninstall-services ## Alias for uninstall-services
status-launchd: status-services ## Alias for status-services

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
