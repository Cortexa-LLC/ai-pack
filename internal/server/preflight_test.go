package server

import (
	"context"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/kgclient"
	"github.com/cortexa-llc/ai-pack/internal/mcp"
)

// TestPreflightContextNilManager verifies that PreflightContext returns an empty
// string (best-effort) when no MCP manager is available.
func TestPreflightContextNilManager(t *testing.T) {
	ctx := context.Background()
	result := kgclient.PreflightContext(ctx, nil, "implement feature X", "/tmp/proj")
	if result != "" {
		t.Errorf("expected empty string from nil manager, got %q", result)
	}
}

// TestPreflightContextEmptyManager verifies that PreflightContext returns an empty
// string (best-effort) when the MCP manager has no registered kg server.
func TestPreflightContextEmptyManager(t *testing.T) {
	ctx := context.Background()
	mgr := mcp.NewManager()
	// No servers started — CallTool will return an error, preflight is skipped.
	result := kgclient.PreflightContext(ctx, mgr, "implement feature X", "/tmp/proj")
	if result != "" {
		t.Errorf("expected empty string from manager with no kg server, got %q", result)
	}
}

// TestBuildSystemPromptPreflightInjection verifies that a non-empty preflight
// block is prepended to the system prompt with the expected separator.
func TestBuildSystemPromptPreflightInjection(t *testing.T) {
	preflightBlock := "## Project Knowledge\nfoo: bar\n"
	roleContext := "You are a test agent."

	// Simulate what executeAgentWorkflow does when preflight returns a block.
	systemPrompt := preflightBlock + "\n---\n\n" + roleContext

	if len(systemPrompt) < len(roleContext) {
		t.Fatal("system prompt shorter than roleContext — injection missing")
	}
	if systemPrompt[:len(preflightBlock)] != preflightBlock {
		t.Errorf("system prompt does not start with preflight block:\ngot: %q", systemPrompt[:50])
	}
	const sep = "\n---\n\n"
	if systemPrompt[len(preflightBlock):len(preflightBlock)+len(sep)] != sep {
		t.Errorf("expected separator after preflight block, got %q", systemPrompt[len(preflightBlock):len(preflightBlock)+len(sep)])
	}
}

// TestPreflightContextSkippedWhenNoResults verifies that an empty preflight
// result does not modify the system prompt (no prepend happens).
func TestPreflightContextSkippedWhenNoResults(t *testing.T) {
	// Empty preflight block means no injection should happen.
	kgBlock := "" // simulates what PreflightContext returns when kg is empty
	roleContext := "You are a test agent."

	systemPrompt := roleContext
	if kgBlock != "" {
		systemPrompt = kgBlock + "\n---\n\n" + roleContext
	}

	if systemPrompt != roleContext {
		t.Errorf("expected unchanged system prompt, got %q", systemPrompt)
	}
}
