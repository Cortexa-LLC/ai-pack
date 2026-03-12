package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// Service handles streaming with multi-provider support and model selection
type Service struct {
	providers          map[string]ProviderFactory
	modelSelector      ModelSelector
	defaultProvider    string
	modelTranslations  map[string]map[string]string // "from->to": {"model": "target", "*pattern*": "target", "*": "default"}
}

// ModelSelector selects the appropriate model and provider
type ModelSelector interface {
	SelectModel(role string, requestedModel string, minContextTokens int) (model string, provider string, err error)
}

// NewService creates a new streaming service
func NewService(selector ModelSelector, defaultProvider string, translations map[string]map[string]string) *Service {
	return &Service{
		providers:         make(map[string]ProviderFactory),
		modelSelector:     selector,
		defaultProvider:   defaultProvider,
		modelTranslations: translations,
	}
}

// RegisterProvider registers a provider factory
func (s *Service) RegisterProvider(factory ProviderFactory) {
	s.providers[factory.GetProviderName()] = factory
}

// CreateStream creates a stream using model selection
func (s *Service) CreateStream(ctx context.Context, role string, req StreamRequest) (StreamProvider, error) {
	// Use model selector to determine which provider to use
	selectedModel, providerName, err := s.modelSelector.SelectModel(role, req.Model, req.MinContextTokens)
	if err != nil {
		monitoring.Logger.Warn("model_selection_failed",
			"role", role,
			"requested_model", req.Model,
			"error", err.Error())
		// Fall back to default provider, keeping the selector's returned model
		// (SelectModel returns a translated default model on error, not the original).
		providerName = s.defaultProvider
		if selectedModel == "" {
			selectedModel = s.translateModelForProvider(req.Model, s.GetProviderForModel(req.Model), s.defaultProvider)
		}
	}

	// Log the selection
	monitoring.Logger.Info("stream_model_selected",
		"role", role,
		"requested", req.Model,
		"selected", selectedModel,
		"provider", providerName)

	// Get the provider factory
	factory, ok := s.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not registered: %s", providerName)
	}

	// Update model in request
	req.Model = selectedModel

	// Create the stream
	stream, err := factory.CreateStream(ctx, req)
	if err != nil {
		monitoring.Logger.Error("stream_creation_failed",
			"provider", providerName,
			"model", selectedModel,
			"error", err.Error())

		// Try to fall back to default provider if not already using it
		if providerName != s.defaultProvider {
			monitoring.Logger.Info("falling_back_to_default_provider",
				"from", providerName,
				"to", s.defaultProvider)

			defaultFactory := s.providers[s.defaultProvider]
			if defaultFactory != nil {
				// Translate model name to one compatible with fallback provider
				fallbackModel := s.translateModelForProvider(req.Model, providerName, s.defaultProvider)

				monitoring.Logger.Info("model_translated_for_fallback",
					"original_model", req.Model,
					"fallback_model", fallbackModel,
					"from_provider", providerName,
					"to_provider", s.defaultProvider)

				req.Model = fallbackModel
				return defaultFactory.CreateStream(ctx, req)
			}
		}

		return nil, err
	}

	return stream, nil
}

// GetProviderForModel returns the provider name that supports a model
func (s *Service) GetProviderForModel(model string) string {
	for _, factory := range s.providers {
		if factory.SupportsModel(model) {
			return factory.GetProviderName()
		}
	}
	return s.defaultProvider
}

// translateModelForProvider translates a model name from one provider to an equivalent
// in another provider using the configured model_translations table.
//
// Keys in each translation map are matched in order:
//  1. Exact match (e.g. "gpt-4o")
//  2. Glob-style substring match (e.g. "*sonnet*" matches any model containing "sonnet")
//  3. Wildcard default ("*")
func (s *Service) translateModelForProvider(model, fromProvider, toProvider string) string {
	if fromProvider == toProvider {
		return model
	}

	key := fromProvider + "->" + toProvider
	table := s.modelTranslations[key]
	if len(table) == 0 {
		return model
	}

	// 1. Exact match
	if target, ok := table[model]; ok {
		return target
	}

	// 2. Glob-style substring match: "*pattern*"
	modelLower := strings.ToLower(model)
	for pattern, target := range table {
		if pattern == "*" || !strings.Contains(pattern, "*") {
			continue
		}
		inner := strings.Trim(pattern, "*")
		if inner != "" && strings.Contains(modelLower, strings.ToLower(inner)) {
			return target
		}
	}

	// 3. Wildcard default
	if target, ok := table["*"]; ok {
		return target
	}

	return model
}
