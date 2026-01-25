# A2A Agent Build Configuration
.PHONY: all build server agent install clean test help

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.Commit=$(COMMIT) \
           -X main.BuildTime=$(BUILD_TIME)

# Directories
BIN_DIR := bin
CMD_DIR := cmd

# Default target
all: build

## build: Build all binaries
build: server agent

## server: Build the agent server
server:
	@echo "Building agent-server..."
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/agent-server ./$(CMD_DIR)/agent-server
	@echo "✅ Built $(BIN_DIR)/agent-server"

## agent: Build the agent CLI
agent:
	@echo "Building agent CLI..."
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/agent ./$(CMD_DIR)/agent
	@echo "✅ Built $(BIN_DIR)/agent"

## install: Install binaries to /usr/local/bin
install: build
	@echo "Installing binaries..."
	@install -m 755 $(BIN_DIR)/agent-server /usr/local/bin/agent-server
	@install -m 755 $(BIN_DIR)/agent /usr/local/bin/agent
	@echo "✅ Installed to /usr/local/bin"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BIN_DIR)/agent-server $(BIN_DIR)/agent
	@echo "✅ Cleaned"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Formatted"

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run 2>/dev/null || go vet ./...

## dev-server: Build and run server in development mode
dev-server: server
	@echo "Starting server in development mode..."
	@./$(BIN_DIR)/agent-server --server

## version: Show version info
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

## help: Show this help message
help:
	@echo "A2A Agent Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
