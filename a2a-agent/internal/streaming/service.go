package streaming

import (
	"context"
	"fmt"

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
				req.Model = req.Model // Keep the model or use default?
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
