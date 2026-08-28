.PHONY: help install-plugin update-plugin verify-hooks

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
	@echo "  make verify-hooks         # Check hooks are installed and runnable"

install-plugin: ## Register marketplace and install the ai-pack plugin into Claude Code (run once)
	@claude plugin marketplace add $(PROJECT_ROOT) 2>/dev/null || true
	@claude plugin install ai-pack@ai-pack
	@echo "✅ ai-pack plugin installed"

update-plugin: ## Force-resync the installed plugin from local source (update is version-gated)
	@claude plugin uninstall ai-pack@ai-pack || true
	@claude plugin install ai-pack@ai-pack
	@echo "✅ ai-pack plugin resynced from local source"

# Plugin SessionStart/PostToolUse hooks CANNOT be exercised headlessly — plugin
# hooks do not load in `claude -p` at all (verified: the known-good PostToolUse
# hook does not fire there either, while an identical settings.json hook does).
# So this target checks everything that IS mechanically checkable and then
# prints the one step a human must do, rather than pretending to verify firing.
verify-hooks: ## Check installed hooks are present, valid, and runnable (+ print the manual firing check)
	@set -e; \
	VERSION_NOW=$$(python3 -c "import json;print(json.load(open('plugin/.claude-plugin/plugin.json'))['version'])"); \
	CACHE="$$HOME/.claude/plugins/cache/ai-pack/ai-pack/$$VERSION_NOW"; \
	echo "Installed plugin: ai-pack $$VERSION_NOW"; \
	[ -d "$$CACHE" ] || { echo "❌ not installed at $$CACHE — run 'make update-plugin'"; exit 1; }; \
	python3 -c "import json,sys;d=json.load(open('$$CACHE/hooks/hooks.json'));print('  hooks.json valid, events:', ', '.join(d['hooks']))"; \
	for s in $$CACHE/hooks/*.py; do \
	  out=$$(echo '{}' | python3 "$$s" 2>/tmp/vh.err); rc=$$?; \
	  [ $$rc -eq 0 ] || { echo "  ❌ $$(basename $$s) exited $$rc"; cat /tmp/vh.err; exit 1; }; \
	  [ -s /tmp/vh.err ] && { echo "  ❌ $$(basename $$s) wrote to stderr:"; cat /tmp/vh.err; exit 1; }; \
	  echo "  ✅ $$(basename $$s) — exit 0, stderr clean, $${#out} chars on stdout"; \
	done; \
	echo ""; \
	echo "⚠️  Firing itself cannot be checked here: plugin hooks do not load in 'claude -p'."; \
	echo "   Confirm manually in a NEW interactive session (or after /clear):"; \
	echo "     ask \"what does the ai-pack session directive tell you to do first?\""; \
	echo "   If the session does not know, the hook is inert — revert it rather than"; \
	echo "   shipping a directive that never applies."
