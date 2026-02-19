package server

import (
	"os"
	"testing"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/config"
)

func TestGetFallbackChain(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("ANTHROPIC_API_TOKEN")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	srv, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-opus-20240229", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	ms := &ModelSelector{server: srv}

	tests := []struct {
		failedModel   string
		expectedChain []string
	}{
		{
			failedModel:   "gpt-5.2",
			expectedChain: []string{"gpt-5.2-mini", "claude-3-opus-20240229"},
		},
		{
			failedModel:   "gpt-5.2-mini",
			expectedChain: []string{"gpt-4o-mini", "claude-3-opus-20240229"},
		},
		{
			failedModel:   "gpt-4o-mini",
			expectedChain: []string{"codex", "claude-3-opus-20240229"},
		},
		{
			failedModel:   "codex",
			expectedChain: []string{"codex-mini", "claude-3-opus-20240229"},
		},
		{
			failedModel:   "codex-mini",
			expectedChain: []string{"claude-3-opus-20240229"},
		},
		{
			failedModel:   "unknown-model",
			expectedChain: []string{"claude-3-opus-20240229"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.failedModel, func(t *testing.T) {
			result := ms.GetFallbackChain(tt.failedModel)
			if len(result) != len(tt.expectedChain) {
				t.Errorf("GetFallbackChain(%q) returned %d models, want %d",
					tt.failedModel, len(result), len(tt.expectedChain))
				return
			}
			for i := range result {
				if result[i] != tt.expectedChain[i] {
					t.Errorf("GetFallbackChain(%q)[%d] = %q, want %q",
						tt.failedModel, i, result[i], tt.expectedChain[i])
				}
			}
		})
	}
}

func TestGetFallbackChainNoOpenAI(t *testing.T) {
	// Setup test environment without OpenAI
	tmpDir := t.TempDir()
	os.Setenv("ANTHROPIC_API_TOKEN", "test-token")
	defer os.Unsetenv("ANTHROPIC_API_TOKEN")

	cfg := config.DefaultConfig()
	cfg.API.Mode = "direct"
	srv, err := NewAgentServer(tmpDir, 3, 4000, "claude-3-opus-20240229", cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	ms := &ModelSelector{server: srv}

	tests := []string{"gpt-5.2", "gpt-4o-mini", "codex", "codex-mini", "unknown-model"}

	for _, failedModel := range tests {
		t.Run(failedModel, func(t *testing.T) {
			result := ms.GetFallbackChain(failedModel)
			expectedChain := []string{"claude-3-opus-20240229"}
			if len(result) != len(expectedChain) {
				t.Errorf("GetFallbackChain(%q) without OpenAI returned %d models, want %d",
					failedModel, len(result), len(expectedChain))
				return
			}
			if result[0] != expectedChain[0] {
				t.Errorf("GetFallbackChain(%q) without OpenAI = %q, want %q",
					failedModel, result[0], expectedChain[0])
			}
		})
	}
}
