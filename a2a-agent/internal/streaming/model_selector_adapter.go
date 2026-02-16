package streaming

import (
	"fmt"
	"strings"
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
