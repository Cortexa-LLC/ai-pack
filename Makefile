.PHONY: help install-plugin update-plugin

# Default target
.DEFAULT_GOAL := help

PROJECT_ROOT := $(shell pwd)

help: ## Show this help message
	@echo "AI-Pack — Claude Code plugin"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start:"
	@echo "  make install-plugin       # Register marketplace and install the plugin (run once)"
	@echo "  make update-plugin        # Refresh the installed plugin after local changes"

install-plugin: ## Register marketplace and install the ai-pack plugin into Claude Code (run once)
	@claude plugin marketplace add $(PROJECT_ROOT) 2>/dev/null || true
	@claude plugin install ai-pack@ai-pack
	@echo "✅ ai-pack plugin installed"

update-plugin: ## Force-resync the installed plugin from local source (update is version-gated)
	@claude plugin uninstall ai-pack@ai-pack || true
	@claude plugin install ai-pack@ai-pack
	@echo "✅ ai-pack plugin resynced from local source"
