package streaming

import "testing"

func TestIsCodexModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"codex", true},
		{"Codex", true},
		{"codex-mini", true},
		{"CODEX-MINI", true},
		{"gpt-4o-mini", false},
		{"gpt-5.2", false},
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
		{"gpt-5.2", true},
		{"GPT-5.2-mini", true},
		{"gpt-4o-mini", false},
		{"codex", false},
		{"codex-mini", false},
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
