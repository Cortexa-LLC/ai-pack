package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// Service handles streaming with multi-provider support and model selection
type Service struct {
	providers       map[string]ProviderFactory
	modelSelector   ModelSelector
	defaultProvider string
}

// ModelSelector selects the appropriate model and provider
type ModelSelector interface {
	SelectModel(role string, requestedModel string) (model string, provider string, err error)
}

// NewService creates a new streaming service
func NewService(selector ModelSelector, defaultProvider string) *Service {
	return &Service{
		providers:       make(map[string]ProviderFactory),
		modelSelector:   selector,
		defaultProvider: defaultProvider,
	}
}

// RegisterProvider registers a provider factory
func (s *Service) RegisterProvider(factory ProviderFactory) {
	s.providers[factory.GetProviderName()] = factory
}

// CreateStream creates a stream using model selection
func (s *Service) CreateStream(ctx context.Context, role string, req StreamRequest) (StreamProvider, error) {
	// Use model selector to determine which provider to use
	selectedModel, providerName, err := s.modelSelector.SelectModel(role, req.Model)
	if err != nil {
		monitoring.Logger.Warn("model_selection_failed",
			"role", role,
			"requested_model", req.Model,
			"error", err.Error())
		// Fall back to default provider
		providerName = s.defaultProvider
		selectedModel = req.Model
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

// translateModelForProvider translates a model name from one provider to an equivalent in another provider
func (s *Service) translateModelForProvider(model, fromProvider, toProvider string) string {
	// If providers are the same, no translation needed
	if fromProvider == toProvider {
		return model
	}

	// OpenAI -> Anthropic translations
	if fromProvider == ProviderOpenAI && toProvider == ProviderAnthropic {
		switch model {
		case "gpt-4o", "gpt-4o-2024-08-06":
			return "claude-sonnet-4-5-20250929" // Map GPT-4o to Claude Sonnet 4.5
		case "gpt-4o-mini":
			return "claude-haiku-4-5-20251022" // Map GPT-4o-mini to Claude Haiku 4.5
		case "gpt-5.2-mini":
			return "claude-sonnet-4-5-20250929" // Map GPT-5.2-mini to Claude Sonnet 4.5
		default:
			// Default to Sonnet 4.5 for unknown GPT models
			return "claude-sonnet-4-5-20250929"
		}
	}

	// Anthropic -> OpenAI translations
	if fromProvider == ProviderAnthropic && toProvider == ProviderOpenAI {
		if strings.Contains(strings.ToLower(model), "opus") {
			return "gpt-4o" // Map Opus to GPT-4o
		} else if strings.Contains(strings.ToLower(model), "sonnet") {
			return "gpt-4o" // Map Sonnet to GPT-4o
		} else if strings.Contains(strings.ToLower(model), "haiku") {
			return "gpt-4o-mini" // Map Haiku to GPT-4o-mini
		}
		// Default to GPT-4o for unknown Claude models
		return "gpt-4o"
	}

	// No translation available, return original model
	return model
}
