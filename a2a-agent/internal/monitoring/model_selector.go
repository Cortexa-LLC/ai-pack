package monitoring

import (
	"fmt"
)

// ModelTier represents model cost/capability tiers
type ModelTier int

const (
	TierMinimal ModelTier = 1 // gpt-4o-mini, claude-haiku-4-5
	TierLow     ModelTier = 2 // gpt-5.2-mini
	TierMedium  ModelTier = 3 // claude-sonnet-4-5
	TierHigh    ModelTier = 4 // claude-opus-4-6
)

// ModelInfo holds model metadata
type ModelInfo struct {
	ID          string
	Tier        ModelTier
	Provider    string
	CostPerMIn  float64 // Input cost per 1M tokens
	CostPerMOut float64 // Output cost per 1M tokens
}

// ModelsByTier maps tiers to available models (ordered by preference)
var ModelsByTier = map[ModelTier][]ModelInfo{
	TierMinimal: {
		{ID: "gpt-4o-mini", Tier: TierMinimal, Provider: "openai", CostPerMIn: 0.15, CostPerMOut: 0.60},
		{ID: "claude-haiku-4-5", Tier: TierMinimal, Provider: "anthropic", CostPerMIn: 1.00, CostPerMOut: 5.00},
	},
	TierLow: {
		{ID: "gpt-5.2-mini", Tier: TierLow, Provider: "openai", CostPerMIn: 0.60, CostPerMOut: 2.40},
	},
	TierMedium: {
		{ID: "claude-sonnet-4-6", Tier: TierMedium, Provider: "anthropic", CostPerMIn: 3.00, CostPerMOut: 15.00},
		{ID: "claude-sonnet-4-5", Tier: TierMedium, Provider: "anthropic", CostPerMIn: 3.00, CostPerMOut: 15.00},
		{ID: "claude-sonnet-4-5-20250929", Tier: TierMedium, Provider: "anthropic", CostPerMIn: 3.00, CostPerMOut: 15.00},
	},
	TierHigh: {
		{ID: "claude-opus-4-6", Tier: TierHigh, Provider: "anthropic", CostPerMIn: 5.00, CostPerMOut: 25.00},
	},
}

// RoleDefaultTier maps roles to their default starting tiers
var RoleDefaultTier = map[string]ModelTier{
	"engineer":     TierMinimal, // Start cheap
	"orchestrator": TierLow,     // Needs more capability
	"inspector":    TierMinimal, // Pattern recognition, not raw intelligence
	"architect":    TierMedium,  // Design needs good reasoning
	"reviewer":     TierLow,     // Code review is moderate complexity
	"tester":       TierMinimal, // Test writing is straightforward
	"spelunker":    TierMedium,  // Deep investigation requires strong reasoning
}

// ModelSelector selects the best model based on performance and complexity
type ModelSelector struct {
	gradeManager       *PerformanceGradeManager
	complexityAnalyzer *ComplexityAnalyzer
	minTier            ModelTier
	maxTier            ModelTier
	enabled            bool
}

// ModelSelectionResult holds the selection decision and reasoning
type ModelSelectionResult struct {
	SelectedModel ModelInfo
	Tier          ModelTier
	Reasoning     string
	Complexity    ComplexityLevel
	Grade         *PerformanceGrade
	WasAdjusted   bool // True if tier was adjusted from default
}

// NewModelSelector creates a new adaptive model selector
func NewModelSelector(gradeManager *PerformanceGradeManager, complexityAnalyzer *ComplexityAnalyzer) *ModelSelector {
	return &ModelSelector{
		gradeManager:       gradeManager,
		complexityAnalyzer: complexityAnalyzer,
		minTier:            TierMinimal,
		maxTier:            TierMedium, // Cap at Sonnet 4.6 - Premium tier requires explicit user approval
		enabled:            true,
	}
}

// SetTierLimits sets the min/max tier bounds
func (ms *ModelSelector) SetTierLimits(minTier, maxTier ModelTier) {
	ms.minTier = minTier
	ms.maxTier = maxTier
}

// SetEnabled enables or disables adaptive selection
func (ms *ModelSelector) SetEnabled(enabled bool) {
	ms.enabled = enabled
}

// SelectModel chooses the best model for a task
func (ms *ModelSelector) SelectModel(
	role string,
	projectID string,
	taskDescription string,
) ModelSelectionResult {
	// If adaptive selection is disabled, use default
	if !ms.enabled {
		defaultTier := ms.getDefaultTier(role)
		model := ms.getBestModelFromTier(defaultTier)
		return ModelSelectionResult{
			SelectedModel: model,
			Tier:          defaultTier,
			Reasoning:     "Adaptive selection disabled, using default tier",
			WasAdjusted:   false,
		}
	}

	// 1. Analyze task complexity
	complexity := ms.complexityAnalyzer.AnalyzeComplexity(taskDescription)

	// 2. Get default tier for role
	defaultTier := ms.getDefaultTier(role)

	// 3. Start with default tier
	selectedTier := defaultTier

	// 4. Check if we have performance history
	var grade *PerformanceGrade
	var reasoning string
	wasAdjusted := false

	// Try to find grade for current default model
	defaultModel := ms.getBestModelFromTier(defaultTier)
	grade = ms.gradeManager.GetGrade(defaultModel.ID, role, projectID)

	// 5. Adjust tier based on performance history (if confident)
	if grade != nil && grade.ConfidenceScore > 0.5 {
		switch grade.Grade {
		case "A":
			// Performing excellently - can we downgrade to save cost?
			if selectedTier > ms.minTier && selectedTier > TierMinimal {
				// Don't downgrade below complexity minimum
				complexityMinTier := ModelTier(GetMinimumTier(complexity))
				if selectedTier-1 >= complexityMinTier {
					selectedTier = selectedTier - 1
					reasoning = fmt.Sprintf("Downgraded from tier %d to %d (Grade A, %.0f%% success rate)",
						defaultTier, selectedTier, grade.SuccessRate*100)
					wasAdjusted = true
				}
			}

		case "B":
			// Good performance - keep current tier
			reasoning = fmt.Sprintf("Maintaining tier %d (Grade B, %.0f%% success rate)",
				selectedTier, grade.SuccessRate*100)

		case "C":
			// Marginal performance - consider escalating
			if grade.TotalAttempts > 5 {
				reasoning = fmt.Sprintf("Warning: Grade C (%.0f%% success rate) - consider escalation if problems persist",
					grade.SuccessRate*100)
			}

		case "D", "F":
			// Poor performance - escalate
			if selectedTier < ms.maxTier {
				selectedTier = selectedTier + 1
				reasoning = fmt.Sprintf("Escalated from tier %d to %d (Grade %s, %.0f%% success rate)",
					defaultTier, selectedTier, grade.Grade, grade.SuccessRate*100)
				wasAdjusted = true
			}
		}
	}

	// 6. Override tier based on complexity if needed
	complexityMinTier := ModelTier(GetMinimumTier(complexity))
	if complexityMinTier > selectedTier {
		oldTier := selectedTier
		selectedTier = complexityMinTier
		if reasoning == "" {
			reasoning = fmt.Sprintf("Escalated from tier %d to %d due to %s complexity",
				oldTier, selectedTier, complexity)
		} else {
			reasoning += fmt.Sprintf("; adjusted to tier %d for %s complexity",
				selectedTier, complexity)
		}
		wasAdjusted = true
	}

	// 7. Enforce tier limits
	if selectedTier < ms.minTier {
		selectedTier = ms.minTier
	}
	if selectedTier > ms.maxTier {
		selectedTier = ms.maxTier
	}

	// 8. Select best model from tier
	selectedModel := ms.getBestModelFromTier(selectedTier)

	// 9. Build default reasoning if none set
	if reasoning == "" {
		if grade != nil {
			reasoning = fmt.Sprintf("Using tier %d based on %s complexity (Grade %s, %.0f%% success, %d samples)",
				selectedTier, complexity, grade.Grade, grade.SuccessRate*100, grade.TotalAttempts)
		} else {
			reasoning = fmt.Sprintf("Using tier %d based on %s complexity (no performance history yet)",
				selectedTier, complexity)
		}
	}

	return ModelSelectionResult{
		SelectedModel: selectedModel,
		Tier:          selectedTier,
		Reasoning:     reasoning,
		Complexity:    complexity,
		Grade:         grade,
		WasAdjusted:   wasAdjusted,
	}
}

// getDefaultTier returns the default tier for a role
func (ms *ModelSelector) getDefaultTier(role string) ModelTier {
	if tier, exists := RoleDefaultTier[role]; exists {
		return tier
	}
	return TierLow // Default fallback
}

// getBestModelFromTier selects the best (usually cheapest/preferred) model from a tier
func (ms *ModelSelector) getBestModelFromTier(tier ModelTier) ModelInfo {
	models, exists := ModelsByTier[tier]
	if !exists || len(models) == 0 {
		// Fallback to tier 2 if tier not found
		models = ModelsByTier[TierLow]
	}

	// Return first model (preferred)
	return models[0]
}

// GetModelInfo looks up model information by ID
func GetModelInfo(modelID string) (ModelInfo, bool) {
	for _, models := range ModelsByTier {
		for _, model := range models {
			if model.ID == modelID {
				return model, true
			}
		}
	}
	return ModelInfo{}, false
}

// GetModelTier returns the tier for a given model ID
func GetModelTier(modelID string) (ModelTier, bool) {
	info, found := GetModelInfo(modelID)
	if !found {
		return TierLow, false
	}
	return info.Tier, true
}

// EstimateCost estimates the cost for a task based on model and token usage
func EstimateCost(modelID string, inputTokens, outputTokens int64) float64 {
	info, found := GetModelInfo(modelID)
	if !found {
		return 0.0
	}

	inputCost := float64(inputTokens) / 1_000_000.0 * info.CostPerMIn
	outputCost := float64(outputTokens) / 1_000_000.0 * info.CostPerMOut

	return inputCost + outputCost
}

// Global model selector instance
var GlobalModelSelector *ModelSelector

// InitModelSelector initializes the global model selector
func InitModelSelector(gradeManager *PerformanceGradeManager, complexityAnalyzer *ComplexityAnalyzer) {
	GlobalModelSelector = NewModelSelector(gradeManager, complexityAnalyzer)
}
