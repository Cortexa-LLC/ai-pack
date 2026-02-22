package streaming

import (
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// resolveProvider returns the provider name for a model ID based on naming convention.
func resolveProvider(modelID string, openaiAvailable, geminiAvailable bool, defaultModel string) (string, string, error) {
	lower := strings.ToLower(modelID)
	if strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "o4") {
		if !openaiAvailable {
			return defaultModel, ProviderAnthropic, fmt.Errorf("openai not available, falling back to default")
		}
		return modelID, ProviderOpenAI, nil
	}
	if strings.Contains(lower, "gemini") {
		if !geminiAvailable {
			return defaultModel, ProviderAnthropic, fmt.Errorf("gemini not available, falling back to default")
		}
		return modelID, ProviderGemini, nil
	}
	return modelID, ProviderAnthropic, nil
}

// SimpleModelSelector implements model selection logic
type SimpleModelSelector struct {
	defaultModel      string
	openaiAvailable   bool
	geminiAvailable   bool
	agentConfigGetter func(role string) string
}

// NewSimpleModelSelector creates a basic model selector
func NewSimpleModelSelector(defaultModel string, openaiAvailable bool, agentConfigGetter func(role string) string) *SimpleModelSelector {
	return &SimpleModelSelector{
		defaultModel:      defaultModel,
		openaiAvailable:   openaiAvailable,
		agentConfigGetter: agentConfigGetter,
	}
}

// SelectModel selects the appropriate model and provider for a role
func (s *SimpleModelSelector) SelectModel(role string, requestedModel string, _ int) (model string, provider string, err error) {
	// Get model from agent config if role is provided
	if role != "" && s.agentConfigGetter != nil {
		configModel := s.agentConfigGetter(role)
		if configModel != "" {
			requestedModel = configModel
		}
	}

	// Use default if no model specified
	if requestedModel == "" {
		requestedModel = s.defaultModel
	}

	// Determine provider based on model name
	modelLower := strings.ToLower(requestedModel)
	if strings.HasPrefix(modelLower, "gpt-") || strings.HasPrefix(modelLower, "o1") || strings.HasPrefix(modelLower, "o3") || strings.HasPrefix(modelLower, "o4") {
		// OpenAI model requested
		if !s.openaiAvailable {
			// OpenAI not available, fall back to Anthropic
			return s.defaultModel, "anthropic", fmt.Errorf("openai not available, falling back to anthropic")
		}
		return requestedModel, "openai", nil
	}

	if strings.Contains(modelLower, "gemini") {
		if !s.geminiAvailable {
			return s.defaultModel, ProviderAnthropic, fmt.Errorf("gemini not available, falling back to anthropic")
		}
		return requestedModel, ProviderGemini, nil
	}

	// Claude or unknown model - use Anthropic
	return requestedModel, "anthropic", nil
}

// PerformanceGradeModelSelector adapts monitoring.ModelSelector to streaming.ModelSelector interface
type PerformanceGradeModelSelector struct {
	gradeSelector    *monitoring.ModelSelector
	projectID        string
	defaultModel     string
	openaiAvailable  bool
	geminiAvailable  bool
}

// NewPerformanceGradeModelSelector creates a selector that uses performance grades
func NewPerformanceGradeModelSelector(projectID string, defaultModel string, openaiAvailable bool, geminiAvailable bool) *PerformanceGradeModelSelector {
	return &PerformanceGradeModelSelector{
		gradeSelector:   monitoring.GlobalModelSelector,
		projectID:       projectID,
		defaultModel:    defaultModel,
		openaiAvailable: openaiAvailable,
		geminiAvailable: geminiAvailable,
	}
}

// SelectModel uses performance grades to select the best model.
//
//   - requestedModel == "" → grade-based selection runs; picks cheapest effective model
//   - requestedModel != "" → explicit role-config pin; honored as-is (bypasses grades)
//
// The server passes "" when the role config has no explicit model so grades decide.
// It passes the pinned model ID only when the role config explicitly requests one.
func (s *PerformanceGradeModelSelector) SelectModel(role string, requestedModel string, minContextTokens int) (model string, provider string, err error) {
	// Explicit pin in role config — honor it without consulting grades.
	if requestedModel != "" {
		monitoring.Logger.Info("model_pinned_by_role_config",
			"role", role,
			"model", requestedModel,
		)
		return s.resolveProviderForModel(requestedModel)
	}

	// If no performance-grade selector available, fall back to default model
	if s.gradeSelector == nil {
		monitoring.Logger.Warn("performance_grade_selector_not_initialized", "falling_back_to", s.defaultModel)
		return s.resolveProviderForModel(s.defaultModel)
	}

	// Use empty string for taskDescription - selector will use role defaults
	result := s.gradeSelector.SelectModel(role, s.projectID, "", minContextTokens)

	monitoring.Logger.Info("model_selected_by_performance",
		"role", role,
		"selected_model", result.SelectedModel.ID,
		"tier", result.Tier,
		"reasoning", result.Reasoning,
		"complexity", result.Complexity,
	)

	// Determine provider from model ID
	providerName := result.SelectedModel.Provider
	modelID := result.SelectedModel.ID

	// Check if provider is available; if not, fall back to the Anthropic default.
	if providerName == ProviderOpenAI && !s.openaiAvailable {
		monitoring.Logger.Warn("openai_not_available_fallback",
			"requested", modelID,
			"falling_back_to", s.defaultModel)
		return s.defaultModel, ProviderAnthropic, nil
	}
	if providerName == ProviderGemini && !s.geminiAvailable {
		monitoring.Logger.Warn("gemini_not_available_fallback",
			"requested", modelID,
			"falling_back_to", s.defaultModel)
		return s.defaultModel, ProviderAnthropic, nil
	}

	return modelID, providerName, nil
}

// resolveProviderForModel resolves the provider for a given model ID, respecting
// availability of OpenAI/Gemini keys and falling back to the Anthropic default.
func (s *PerformanceGradeModelSelector) resolveProviderForModel(modelID string) (string, string, error) {
	return resolveProvider(modelID, s.openaiAvailable, s.geminiAvailable, s.defaultModel)
}
