package streaming

import (
	"log/slog"
	"os"
	"testing"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

func TestMain(m *testing.M) {
	monitoring.InitLogger(slog.LevelError) // silence logs during tests
	os.Exit(m.Run())
}

func TestSimpleModelSelector_SelectModel(t *testing.T) {
	tests := []struct {
		name             string
		defaultModel     string
		openaiAvailable  bool
		requestedModel   string
		role             string
		expectedModel    string
		expectedProvider string
		expectError      bool
	}{
		{
			name:             "codex with OpenAI available",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "codex",
			role:             "engineer",
			expectedModel:    "codex",
			expectedProvider: "openai",
			expectError:      false,
		},
		{
			name:             "codex-mini with OpenAI available",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "codex-mini",
			role:             "engineer",
			expectedModel:    "codex-mini",
			expectedProvider: "openai",
			expectError:      false,
		},
		{
			name:             "CODEX case insensitive",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "CODEX",
			role:             "engineer",
			expectedModel:    "CODEX",
			expectedProvider: "openai",
			expectError:      false,
		},
		{
			name:             "codex without OpenAI available",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  false,
			requestedModel:   "codex",
			role:             "engineer",
			expectedModel:    "claude-3-opus-20240229",
			expectedProvider: "anthropic",
			expectError:      true,
		},
		{
			name:             "gpt-4o-mini with OpenAI available",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "gpt-4o-mini",
			role:             "engineer",
			expectedModel:    "gpt-4o-mini",
			expectedProvider: "openai",
			expectError:      false,
		},
		{
			name:             "claude model with OpenAI available",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "claude-3-opus-20240229",
			role:             "architect",
			expectedModel:    "claude-3-opus-20240229",
			expectedProvider: "anthropic",
			expectError:      false,
		},
		{
			name:             "empty model falls back to default",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "",
			role:             "engineer",
			expectedModel:    "claude-3-opus-20240229",
			expectedProvider: "anthropic",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewSimpleModelSelector(tt.defaultModel, tt.openaiAvailable, nil)
			model, provider, err := selector.SelectModel(tt.role, tt.requestedModel)

			if (err != nil) != tt.expectError {
				t.Errorf("SelectModel() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if model != tt.expectedModel {
				t.Errorf("SelectModel() model = %v, want %v", model, tt.expectedModel)
			}

			if provider != tt.expectedProvider {
				t.Errorf("SelectModel() provider = %v, want %v", provider, tt.expectedProvider)
			}
		})
	}
}

func TestPerformanceGradeModelSelector_SelectModel(t *testing.T) {
	tests := []struct {
		name             string
		projectID        string
		defaultModel     string
		openaiAvailable  bool
		requestedModel   string
		role             string
		expectedModel    string
		expectedProvider string
	}{
		{
			name:             "explicit codex request with OpenAI",
			projectID:        "test-project",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "codex",
			role:             "engineer",
			expectedModel:    "codex",
			expectedProvider: "openai",
		},
		{
			name:             "explicit codex-mini request with OpenAI",
			projectID:        "test-project",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "codex-mini",
			role:             "engineer",
			expectedModel:    "codex-mini",
			expectedProvider: "openai",
		},
		{
			name:             "explicit codex request without OpenAI",
			projectID:        "test-project",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  false,
			requestedModel:   "codex",
			role:             "engineer",
			expectedModel:    "claude-3-opus-20240229",
			expectedProvider: "anthropic",
		},
		{
			name:             "explicit gpt request with OpenAI",
			projectID:        "test-project",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "gpt-4o-mini",
			role:             "engineer",
			expectedModel:    "gpt-4o-mini",
			expectedProvider: "openai",
		},
		{
			name:             "no explicit request defaults to base",
			projectID:        "test-project",
			defaultModel:     "claude-3-opus-20240229",
			openaiAvailable:  true,
			requestedModel:   "",
			role:             "engineer",
			expectedModel:    "claude-3-opus-20240229",
			expectedProvider: "anthropic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewPerformanceGradeModelSelector(tt.projectID, tt.defaultModel, tt.openaiAvailable)
			model, provider, err := selector.SelectModel(tt.role, tt.requestedModel)

			if err != nil {
				t.Errorf("SelectModel() unexpected error = %v", err)
				return
			}

			if model != tt.expectedModel {
				t.Errorf("SelectModel() model = %v, want %v", model, tt.expectedModel)
			}

			if provider != tt.expectedProvider {
				t.Errorf("SelectModel() provider = %v, want %v", provider, tt.expectedProvider)
			}
		})
	}
}
