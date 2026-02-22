package monitoring

import (
	"fmt"
	"sort"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants"
)

// ModelTier represents model cost/capability tiers
type ModelTier int

const (
	TierMinimal ModelTier = 1 // gpt-4o-mini, gpt-4.1-nano, claude-haiku-4-5
	TierLow     ModelTier = 2 // gpt-4.1-mini
	TierMedium  ModelTier = 3 // gpt-4.1, gpt-5.1-codex-mini, claude-sonnet-4-6
	TierHigh    ModelTier = 4 // gpt-5.1-codex, gpt-5.2-codex, claude-opus-4-6
)

// ModelInfo holds model metadata
type ModelInfo struct {
	ID            string
	Tier          ModelTier
	Provider      string
	CostPerMIn    float64 // Input cost per 1M tokens
	CostPerMOut   float64 // Output cost per 1M tokens
	ContextWindow int     // Max context in tokens (0 = unknown, no filtering applied)
}

// ModelsByTier maps tiers to available models (ordered by preference)
var ModelsByTier = map[ModelTier][]ModelInfo{
	TierMinimal: {
		{ID: "gpt-4o-mini", Tier: TierMinimal, Provider: constants.ProviderOpenAI, CostPerMIn: 0.15, CostPerMOut: 0.60, ContextWindow: 128_000},
		{ID: "gpt-4.1-nano", Tier: TierMinimal, Provider: constants.ProviderOpenAI, CostPerMIn: 0.10, CostPerMOut: 0.40, ContextWindow: 1_047_576},
		{ID: "claude-haiku-4-5", Tier: TierMinimal, Provider: constants.ProviderAnthropic, CostPerMIn: 1.00, CostPerMOut: 5.00, ContextWindow: 200_000},
		{ID: "gemini-2.5-flash-lite", Tier: TierMinimal, Provider: constants.ProviderGemini, CostPerMIn: 0.10, CostPerMOut: 0.40, ContextWindow: 1_048_576},
	},
	TierLow: {
		{ID: "gpt-4.1-mini", Tier: TierLow, Provider: constants.ProviderOpenAI, CostPerMIn: 0.40, CostPerMOut: 1.60, ContextWindow: 1_047_576},
		{ID: "o4-mini", Tier: TierLow, Provider: constants.ProviderOpenAI, CostPerMIn: 1.10, CostPerMOut: 4.40, ContextWindow: 200_000},
		{ID: "gemini-2.5-flash", Tier: TierLow, Provider: constants.ProviderGemini, CostPerMIn: 0.30, CostPerMOut: 2.50, ContextWindow: 1_048_576},
	},
	TierMedium: {
		{ID: "gpt-5.1-codex-mini", Tier: TierMedium, Provider: constants.ProviderOpenAI, CostPerMIn: 1.50, CostPerMOut: 6.00, ContextWindow: 0},
		{ID: "gpt-4.1", Tier: TierMedium, Provider: constants.ProviderOpenAI, CostPerMIn: 2.00, CostPerMOut: 8.00, ContextWindow: 1_047_576},
		{ID: "claude-sonnet-4-6", Tier: TierMedium, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "claude-sonnet-4-5", Tier: TierMedium, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "claude-sonnet-4-5-20250929", Tier: TierMedium, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "gemini-2.5-pro", Tier: TierMedium, Provider: constants.ProviderGemini, CostPerMIn: 1.25, CostPerMOut: 10.00, ContextWindow: 1_048_576},
	},
	TierHigh: {
		{ID: "gpt-5.1-codex", Tier: TierHigh, Provider: constants.ProviderOpenAI, CostPerMIn: 3.00, CostPerMOut: 12.00, ContextWindow: 0},
		{ID: "gpt-5.2-codex", Tier: TierHigh, Provider: constants.ProviderOpenAI, CostPerMIn: 5.00, CostPerMOut: 20.00, ContextWindow: 0},
		{ID: "claude-opus-4-5", Tier: TierHigh, Provider: constants.ProviderAnthropic, CostPerMIn: 15.00, CostPerMOut: 75.00, ContextWindow: 200_000},
		{ID: "claude-opus-4-6", Tier: TierHigh, Provider: constants.ProviderAnthropic, CostPerMIn: 15.00, CostPerMOut: 75.00, ContextWindow: 200_000},
		{ID: "gemini-3-pro-preview", Tier: TierHigh, Provider: constants.ProviderGemini, CostPerMIn: 5.00, CostPerMOut: 20.00, ContextWindow: 0},
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
	openaiAvailable    bool
	geminiAvailable    bool
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
func NewModelSelector(gradeManager *PerformanceGradeManager, complexityAnalyzer *ComplexityAnalyzer, openaiAvailable, geminiAvailable bool) *ModelSelector {
	return &ModelSelector{
		gradeManager:       gradeManager,
		complexityAnalyzer: complexityAnalyzer,
		minTier:            TierMinimal,
		maxTier:            TierMedium, // Cap at Sonnet 4.6 - Premium tier requires explicit user approval
		enabled:            true,
		openaiAvailable:    openaiAvailable,
		geminiAvailable:    geminiAvailable,
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

// SelectModel chooses the best model for a task.
// minContextTokens is the minimum context window the selected model must support
// (0 means no constraint).
func (ms *ModelSelector) SelectModel(
	role string,
	projectID string,
	taskDescription string,
	minContextTokens int,
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

	// 5. Adjust tier based on performance history (if confident enough).
	// Threshold 0.1 ≈ 5 samples — low enough for benchmark data to influence selection.
	if grade != nil && grade.ConfidenceScore > 0.1 {
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

	// 8. Select best available model from tier (cheapest with passing grade, or cheapest available)
	selectedModel := ms.getBestAvailableModelFromTier(selectedTier, role, projectID, minContextTokens)

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

// getBestModelFromTier returns the first model in a tier (used for grade lookups,
// where we want a deterministic "representative" model, not cost-optimised).
func (ms *ModelSelector) getBestModelFromTier(tier ModelTier) ModelInfo {
	models, exists := ModelsByTier[tier]
	if !exists || len(models) == 0 {
		models = ModelsByTier[TierLow]
	}
	return models[0]
}

// getBestAvailableModelFromTier returns the cheapest available model from a tier
// that has an acceptable grade (A or B) and meets the minimum context requirement.
// Falls back to cheapest available when no grade data is present or no model
// achieves a passing grade yet.
func (ms *ModelSelector) getBestAvailableModelFromTier(tier ModelTier, role, projectID string, minContextTokens int) ModelInfo {
	models, exists := ModelsByTier[tier]
	if !exists || len(models) == 0 {
		models = ModelsByTier[TierLow]
	}

	// Filter to providers that are available and meet the context window requirement.
	available := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		if m.Provider == constants.ProviderOpenAI && !ms.openaiAvailable {
			continue
		}
		if m.Provider == constants.ProviderGemini && !ms.geminiAvailable {
			continue
		}
		// Skip models with a known context window that is too small for the role.
		// ContextWindow == 0 means unknown — allow it through.
		if minContextTokens > 0 && m.ContextWindow > 0 && m.ContextWindow < minContextTokens {
			continue
		}
		available = append(available, m)
	}
	if len(available) == 0 {
		// All models filtered out — fall back to full list so the agent can still run.
		available = models
	}

	// Sort by combined cost (cheapest first) as the default preference.
	sort.Slice(available, func(i, j int) bool {
		costI := available[i].CostPerMIn + available[i].CostPerMOut
		costJ := available[j].CostPerMIn + available[j].CostPerMOut
		return costI < costJ
	})

	// Find cheapest model with a confirmed passing grade (A or B).
	for _, m := range available {
		grade := ms.gradeManager.GetGrade(m.ID, role, projectID)
		if grade != nil && grade.ConfidenceScore > 0.1 && (grade.Grade == "A" || grade.Grade == "B") {
			return m
		}
	}

	// No confirmed passing model yet — return cheapest available so we start
	// gathering real performance data on the most cost-effective option.
	return available[0]
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
func InitModelSelector(gradeManager *PerformanceGradeManager, complexityAnalyzer *ComplexityAnalyzer, openaiAvailable, geminiAvailable bool) {
	GlobalModelSelector = NewModelSelector(gradeManager, complexityAnalyzer, openaiAvailable, geminiAvailable)
}
