package streaming

import "testing"

func TestIsCodexModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"gpt-5.1-codex", true},
		{"gpt-5.1-codex-mini", true},
		{"gpt-5.2-codex", true},
		{"GPT-5.1-CODEX", true},
		{"gpt-4o-mini", false},
		{"gpt-4.1-mini", false},
		{"o1-preview", false},
		{"claude-3-opus", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := isCodexModel(tt.model)
			if result != tt.expected {
				t.Errorf("isCodexModel(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}

func TestUsesMaxCompletionTokens(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"o1-preview", true},
		{"O1-mini", true},
		{"o3-mini", true},
		{"o4-mini", true},
		{"gpt-5.2", true},
		{"gpt-5", true},
		{"gpt-5-mini", true},
		{"gpt-5.1-codex", true},
		{"gpt-5.2-codex", true},
		{"gpt-4o-mini", false},
		{"gpt-4.1-mini", false},
		{"gpt-4.1-nano", false},
		{"claude-sonnet-4-6", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := usesMaxCompletionTokens(tt.model)
			if result != tt.expected {
				t.Errorf("usesMaxCompletionTokens(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}
