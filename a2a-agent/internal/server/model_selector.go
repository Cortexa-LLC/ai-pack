package server

import (
	"fmt"
	"strings"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants"
	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

// ModelSelector handles intelligent model selection with fallback logic
type ModelSelector struct {
	server *AgentServer
}

// NewModelSelector creates a new model selector
func NewModelSelector(server *AgentServer) *ModelSelector {
	return &ModelSelector{server: server}
}

// SelectProviderWithFallback selects the best available provider for a model
// Priority: gpt-5.1-codex → gpt-4.1 → gpt-4.1-mini → gpt-4o-mini → Claude Sonnet
// Note: gpt-4o is excluded — demonstrated unreliable tool use (false positive completions)
func (ms *ModelSelector) SelectProviderWithFallback(requestedModel string) (LLMProvider, string, error) {
	// Check if model is OpenAI (gpt-* or o-series)
	if strings.HasPrefix(requestedModel, "gpt-") {
		// gpt-4o is excluded — demonstrated unreliable tool use (false positive completions)
		if requestedModel == "gpt-4o" {
			monitoring.Logger.Warn("model_excluded",
				"requested", requestedModel,
				"reason", "unreliable_tool_use",
				"redirect", "gpt-4.1-mini")
			requestedModel = "gpt-4.1-mini"
		}
		// Check if OpenAI is available
		if ms.server.openaiProvider != nil {
			monitoring.Logger.Info("model_selected", "requested", requestedModel, "provider", constants.ProviderOpenAI)
			return NewOpenAIProvider(&ms.server.openaiClient, requestedModel, ms.server.maxTokens), requestedModel, nil
		}

		// OpenAI not available - fall back to Claude
		fallbackModel := ms.server.model // Default Claude model
		monitoring.Logger.Warn("model_fallback_to_claude",
			"requested", requestedModel,
			"fallback", fallbackModel,
			"reason", "openai_api_key_not_set")

		return ms.server.anthropicProvider, fallbackModel, nil
	}

	// Claude model requested or default
	monitoring.Logger.Info("model_selected", "requested", requestedModel, "provider", constants.ProviderAnthropic)
	return NewAnthropicProvider(&ms.server.client, requestedModel, ms.server.maxTokens), requestedModel, nil
}

// RecommendModel suggests the best available model based on task complexity
// Favors models in availability order: gpt-5.1-codex → gpt-4.1 → gpt-4.1-mini → gpt-4o-mini → Claude
func (ms *ModelSelector) RecommendModel(taskDescription string) string {
	desc := strings.ToLower(taskDescription)

	// Analyze task complexity
	isSimple := strings.Contains(desc, "simple") ||
		strings.Contains(desc, "format") ||
		strings.Contains(desc, "rename") ||
		strings.Contains(desc, "update docs") ||
		strings.Contains(desc, "typo")

	isBackground := strings.Contains(desc, "batch") ||
		strings.Contains(desc, "bulk") ||
		strings.Contains(desc, "test") && strings.Contains(desc, "run")

	isComplex := strings.Contains(desc, "architect") ||
		strings.Contains(desc, "design system") ||
		strings.Contains(desc, "optimize") ||
		strings.Contains(desc, "performance")

	isCritical := strings.Contains(desc, "security") ||
		strings.Contains(desc, "production") ||
		strings.Contains(desc, "critical")

	// Priority-based recommendation (favor availability)
	if ms.server.openaiProvider == nil {
		// No OpenAI available - use Claude
		return ms.server.model
	}

	// Recommend based on task type (OpenAI available)
	// Note: gpt-4o intentionally excluded — unreliable tool use
	switch {
	case isCritical:
		// Critical tasks - use Claude Sonnet for maximum reliability
		return ms.server.model

	case isComplex:
		// Complex tasks - use gpt-5.1-codex (best reasoning for code)
		return "gpt-5.1-codex"

	case isBackground:
		// Background/bulk - use gpt-4.1-mini (good quality, low cost)
		return "gpt-4.1-mini"

	case isSimple:
		// Simple tasks - use gpt-4o-mini (cheapest)
		return "gpt-4o-mini"

	default:
		// Default: gpt-4.1-mini (reliable, economical)
		return "gpt-4.1-mini"
	}
}

// GetProviderCost returns estimated cost per 1M tokens for a model
func (ms *ModelSelector) GetProviderCost(model string) string {
	switch {
	case strings.Contains(model, "gpt-5.2-codex"):
		return "$5.00 input / $20.00 output per 1M tokens"
	case strings.Contains(model, "gpt-5.1-codex-mini"):
		return "$1.50 input / $6.00 output per 1M tokens"
	case strings.Contains(model, "gpt-5.1-codex"):
		return "$3.00 input / $12.00 output per 1M tokens"
	case strings.HasPrefix(model, "gpt-4.1-nano"):
		return "$0.10 input / $0.40 output per 1M tokens"
	case strings.HasPrefix(model, "gpt-4.1-mini"):
		return "$0.40 input / $1.60 output per 1M tokens"
	case strings.HasPrefix(model, "gpt-4.1"):
		return "$2.00 input / $8.00 output per 1M tokens"
	case strings.HasPrefix(model, "gpt-4o-mini"):
		return "$0.15 input / $0.60 output per 1M tokens"
	case strings.HasPrefix(model, "gpt-4o"):
		return "$2.50 input / $10.00 output per 1M tokens"
	case strings.Contains(model, "haiku"):
		return "$0.25 input / $1.25 output per 1M tokens"
	case strings.Contains(model, "sonnet"):
		return "$3.00 input / $15.00 output per 1M tokens"
	case strings.Contains(model, "opus"):
		return "$15.00 input / $75.00 output per 1M tokens"
	default:
		return "Unknown"
	}
}

// ValidateModelAvailability checks if a model is available and returns warnings
func (ms *ModelSelector) ValidateModelAvailability(model string) []string {
	var warnings []string

	if strings.HasPrefix(model, "gpt-") && ms.server.openaiProvider == nil {
		warnings = append(warnings, fmt.Sprintf(
			"⚠️  Model '%s' requires OPENAI_API_KEY (not set). Will fall back to Claude Sonnet.",
			model))
		warnings = append(warnings, "💡 Run setup script: ./scripts/setup-api-keys.sh")
	}

	return warnings
}

// CalculateCost estimates cost based on token usage (corrected pricing)
func CalculateCost(model string, inputTokens, outputTokens int) float64 {
	// Cost per 1M tokens
	var inputCost, outputCost float64

	switch {
	case strings.Contains(model, "gpt-5.2-codex"):
		inputCost = 5.00
		outputCost = 20.00
	case strings.Contains(model, "gpt-5.1-codex-mini"):
		inputCost = 1.50
		outputCost = 6.00
	case strings.Contains(model, "gpt-5.1-codex"):
		inputCost = 3.00
		outputCost = 12.00
	case strings.HasPrefix(model, "gpt-4.1-nano"):
		inputCost = 0.10
		outputCost = 0.40
	case strings.HasPrefix(model, "gpt-4.1-mini"):
		inputCost = 0.40
		outputCost = 1.60
	case strings.HasPrefix(model, "gpt-4.1"):
		inputCost = 2.00
		outputCost = 8.00
	case strings.HasPrefix(model, "gpt-4o-mini"):
		inputCost = 0.15
		outputCost = 0.60
	case strings.HasPrefix(model, "gpt-4o"):
		inputCost = 2.50
		outputCost = 10.00
	case strings.Contains(model, "haiku"):
		inputCost = 0.25
		outputCost = 1.25
	case strings.Contains(model, "sonnet"):
		inputCost = 3.00
		outputCost = 15.00
	case strings.Contains(model, "opus"):
		inputCost = 15.00
		outputCost = 75.00
	default:
		return 0
	}

	// Convert to actual cost (tokens / 1M * cost per 1M)
	totalCost := (float64(inputTokens)/1000000.0)*inputCost +
		(float64(outputTokens)/1000000.0)*outputCost

	return totalCost
}

// GetFallbackChain returns the fallback chain for a failed model
func (ms *ModelSelector) GetFallbackChain(failedModel string) []string {
	// If OpenAI available, use OpenAI fallback chain
	if ms.server.openaiKey != "" {
		switch failedModel {
		case "gpt-5.2-codex":
			return []string{"gpt-5.1-codex", ms.server.model}
		case "gpt-5.1-codex":
			return []string{"gpt-5.1-codex-mini", ms.server.model}
		case "gpt-5.1-codex-mini":
			return []string{"gpt-4.1-mini", ms.server.model}
		case "gpt-4.1":
			return []string{"gpt-4.1-mini", ms.server.model}
		case "gpt-4.1-mini":
			return []string{"gpt-4o-mini", ms.server.model}
		case "gpt-4o-mini":
			return []string{ms.server.model}
		default:
			return []string{ms.server.model}
		}
	}

	// No OpenAI - Claude only
	return []string{ms.server.model}
}

// ModelStats tracks usage and costs per model
type ModelStats struct {
	Model         string
	Provider      string
	TaskCount     int
	InputTokens   int
	OutputTokens  int
	EstimatedCost float64
}

// FormatCost formats cost as a currency string
func FormatCost(cost float64) string {
	return fmt.Sprintf("$%.4f", cost)
}
