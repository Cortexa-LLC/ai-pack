package streaming

import (
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// SimpleModelSelector implements model selection logic
type SimpleModelSelector struct {
	defaultModel     string
	openaiAvailable  bool
	agentConfigGetter func(role string) string
}

// NewSimpleModelSelector creates a basic model selector
func NewSimpleModelSelector(defaultModel string, openaiAvailable bool, agentConfigGetter func(role string) string) *SimpleModelSelector {
	return &SimpleModelSelector{
		defaultModel:     defaultModel,
		openaiAvailable:  openaiAvailable,
		agentConfigGetter: agentConfigGetter,
	}
}

// SelectModel selects the appropriate model and provider for a role
func (s *SimpleModelSelector) SelectModel(role string, requestedModel string) (model string, provider string, err error) {
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
	if strings.HasPrefix(strings.ToLower(requestedModel), "gpt-") {
		// OpenAI model requested
		if !s.openaiAvailable {
			// OpenAI not available, fall back to Anthropic
			return s.defaultModel, "anthropic", fmt.Errorf("openai not available, falling back to anthropic")
		}
		return requestedModel, "openai", nil
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
}

// NewPerformanceGradeModelSelector creates a selector that uses performance grades
func NewPerformanceGradeModelSelector(projectID string, defaultModel string, openaiAvailable bool) *PerformanceGradeModelSelector {
	return &PerformanceGradeModelSelector{
		gradeSelector:   monitoring.GlobalModelSelector,
		projectID:       projectID,
		defaultModel:    defaultModel,
		openaiAvailable: openaiAvailable,
	}
}

// SelectModel uses performance grades to select the best model
func (s *PerformanceGradeModelSelector) SelectModel(role string, requestedModel string) (model string, provider string, err error) {
	// If no performance-grade selector available, fall back to default
	if s.gradeSelector == nil {
		monitoring.Logger.Warn("performance_grade_selector_not_initialized", "falling_back_to", s.defaultModel)
		return s.defaultModel, "anthropic", nil
	}

	// Use empty string for taskDescription - selector will use role defaults
	result := s.gradeSelector.SelectModel(role, s.projectID, "")

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

	// Check if provider is available
	if providerName == "openai" && !s.openaiAvailable {
		monitoring.Logger.Warn("openai_not_available_fallback",
			"requested", modelID,
			"falling_back_to", s.defaultModel)
		return s.defaultModel, "anthropic", nil
	}

	return modelID, providerName, nil
}
