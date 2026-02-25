package monitoring

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cortexa-llc/ai-pack/internal/constants"
)

// ModelTier represents model cost/capability tiers
type ModelTier int

const (
	TierMinimal ModelTier = 1 // gpt-4o-mini, gpt-4.1-nano, claude-haiku-4-5
	TierLow     ModelTier = 2 // gpt-4.1-mini
	TierMedium  ModelTier = 3 // gpt-4.1, gpt-5.1-codex-mini, claude-sonnet-4-6
	TierHigh    ModelTier = 4 // gpt-5.1-codex, gpt-5.2-codex, claude-opus-4-6
)

// ModelClass describes what kind of work a model is suited for.
// Used to filter out code-completion models from agentic roles.
type ModelClass string

const (
	// ClassAgentic: instruction-following models capable of autonomous multi-step
	// tool use. Required for roles that execute plans (engineer, spelunker, etc.).
	ClassAgentic ModelClass = "agentic"
	// ClassCompletion: code-generation / fill-in-the-middle models (codex family).
	// Good for code quality tasks; NOT suitable for agentic execution loops.
	ClassCompletion ModelClass = "completion"
	// ClassReasoning: extended-thinking models (o-series). Suitable for planning
	// and analysis; tool use support varies by model.
	ClassReasoning ModelClass = "reasoning"
)

// ParseClassString converts a class name string (from role config) to a ModelClass.
// Returns "" (no filter) if the string is empty or unrecognized.
func ParseClassString(s string) ModelClass {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "agentic":
		return ClassAgentic
	case "completion":
		return ClassCompletion
	case "reasoning":
		return ClassReasoning
	default:
		return ""
	}
}

// ModelInfo holds model metadata
type ModelInfo struct {
	ID            string
	Tier          ModelTier
	Class         ModelClass // What kind of work this model is suited for
	Provider      string
	CostPerMIn    float64 // Input cost per 1M tokens
	CostPerMOut   float64 // Output cost per 1M tokens
	ContextWindow int     // Max context in tokens (0 = unknown, no filtering applied)
}

// ModelsByTier maps tiers to available models (ordered by preference)
var ModelsByTier = map[ModelTier][]ModelInfo{
	TierMinimal: {
		{ID: "gpt-4o-mini", Tier: TierMinimal, Class: ClassAgentic, Provider: constants.ProviderOpenAI, CostPerMIn: 0.15, CostPerMOut: 0.60, ContextWindow: 128_000},
		{ID: "gpt-4.1-nano", Tier: TierMinimal, Class: ClassAgentic, Provider: constants.ProviderOpenAI, CostPerMIn: 0.10, CostPerMOut: 0.40, ContextWindow: 1_047_576},
		{ID: "claude-haiku-4-5", Tier: TierMinimal, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 1.00, CostPerMOut: 5.00, ContextWindow: 200_000},
		{ID: "gemini-2.5-flash-lite", Tier: TierMinimal, Class: ClassAgentic, Provider: constants.ProviderGemini, CostPerMIn: 0.10, CostPerMOut: 0.40, ContextWindow: 1_048_576},
	},
	TierLow: {
		{ID: "gpt-4.1-mini", Tier: TierLow, Class: ClassAgentic, Provider: constants.ProviderOpenAI, CostPerMIn: 0.40, CostPerMOut: 1.60, ContextWindow: 1_047_576},
		{ID: "o4-mini", Tier: TierLow, Class: ClassReasoning, Provider: constants.ProviderOpenAI, CostPerMIn: 1.10, CostPerMOut: 4.40, ContextWindow: 200_000},
		{ID: "gemini-2.5-flash", Tier: TierLow, Class: ClassAgentic, Provider: constants.ProviderGemini, CostPerMIn: 0.30, CostPerMOut: 2.50, ContextWindow: 1_048_576},
	},
	TierMedium: {
		{ID: "gpt-5.1-codex-mini", Tier: TierMedium, Class: ClassCompletion, Provider: constants.ProviderOpenAI, CostPerMIn: 1.50, CostPerMOut: 6.00, ContextWindow: 0},
		{ID: "gpt-4.1", Tier: TierMedium, Class: ClassAgentic, Provider: constants.ProviderOpenAI, CostPerMIn: 2.00, CostPerMOut: 8.00, ContextWindow: 1_047_576},
		{ID: "claude-sonnet-4-6", Tier: TierMedium, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "claude-sonnet-4-5", Tier: TierMedium, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "claude-sonnet-4-5-20250929", Tier: TierMedium, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 3.00, CostPerMOut: 15.00, ContextWindow: 200_000},
		{ID: "gemini-2.5-pro", Tier: TierMedium, Class: ClassAgentic, Provider: constants.ProviderGemini, CostPerMIn: 1.25, CostPerMOut: 10.00, ContextWindow: 1_048_576},
	},
	TierHigh: {
		{ID: "gpt-5.1-codex", Tier: TierHigh, Class: ClassCompletion, Provider: constants.ProviderOpenAI, CostPerMIn: 3.00, CostPerMOut: 12.00, ContextWindow: 0},
		{ID: "gpt-5.2-codex", Tier: TierHigh, Class: ClassCompletion, Provider: constants.ProviderOpenAI, CostPerMIn: 5.00, CostPerMOut: 20.00, ContextWindow: 0},
		{ID: "claude-opus-4-5", Tier: TierHigh, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 15.00, CostPerMOut: 75.00, ContextWindow: 200_000},
		{ID: "claude-opus-4-6", Tier: TierHigh, Class: ClassAgentic, Provider: constants.ProviderAnthropic, CostPerMIn: 15.00, CostPerMOut: 75.00, ContextWindow: 200_000},
		{ID: "gemini-3-pro-preview", Tier: TierHigh, Class: ClassAgentic, Provider: constants.ProviderGemini, CostPerMIn: 5.00, CostPerMOut: 20.00, ContextWindow: 0},
	},
}

// ModelSelector selects the best model based on performance and complexity
type ModelSelector struct {
	gradeManager        *PerformanceGradeManager
	complexityAnalyzer  *ComplexityAnalyzer
	minTier             ModelTier
	maxTier             ModelTier
	enabled             bool
	openaiAvailable     bool
	geminiAvailable     bool
	roleDefaultTiers    map[string]ModelTier  // per-role starting tier from **Tier:** role config
	roleRequiredClasses map[string]ModelClass // per-role class filter from **Class:** role config
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

// SetRoleDefaultTier registers a per-role starting tier from the **Tier:** role config field.
// This allows role .md files to influence where grade-based selection begins.
func (ms *ModelSelector) SetRoleDefaultTier(role string, tier ModelTier) {
	if ms.roleDefaultTiers == nil {
		ms.roleDefaultTiers = make(map[string]ModelTier)
	}
	ms.roleDefaultTiers[role] = tier
}

// SetRoleRequiredClass registers a class filter for a role from the **Class:** role config field.
// When set, only models of that class are eligible for selection in that role.
func (ms *ModelSelector) SetRoleRequiredClass(role string, class ModelClass) {
	if ms.roleRequiredClasses == nil {
		ms.roleRequiredClasses = make(map[string]ModelClass)
	}
	ms.roleRequiredClasses[role] = class
}

// getRequiredClass returns the class filter registered for a role, or "" if none.
func (ms *ModelSelector) getRequiredClass(role string) ModelClass {
	if ms.roleRequiredClasses != nil {
		if class, ok := ms.roleRequiredClasses[role]; ok {
			return class
		}
	}
	return ""
}

// ParseTierString converts a tier name string (from role config) to a ModelTier.
// Returns 0 (unset) if the string is empty or unrecognized.
func ParseTierString(s string) ModelTier {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "minimal":
		return TierMinimal
	case "low":
		return TierLow
	case "medium":
		return TierMedium
	case "high":
		return TierHigh
	default:
		return 0 // unset
	}
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
	// Reload grades from disk on every selection to reflect real-time failures.
	if ms.gradeManager != nil {
		if err := ms.gradeManager.ReloadGrades(); err != nil {
			Logger.Warn("grade_reload_failed", "error", err.Error())
		}
	}

	// If adaptive selection is disabled, use default
	if !ms.enabled {
		defaultTier := ms.getDefaultTier(role)
		requiredClass := ms.getRequiredClass(role)
		model := ms.getBestAvailableModelFromTier(defaultTier, role, projectID, minContextTokens, requiredClass)
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
	// Apply class filter from role config (**Class:** field).
	requiredClass := ms.getRequiredClass(role)
	selectedModel := ms.getBestAvailableModelFromTier(selectedTier, role, projectID, minContextTokens, requiredClass)

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

// getDefaultTier returns the default starting tier for a role.
// Uses the per-role override registered via SetRoleDefaultTier (from **Tier:** in .md),
// falling back to TierLow when no override is set.
func (ms *ModelSelector) getDefaultTier(role string) ModelTier {
	if ms.roleDefaultTiers != nil {
		if tier, ok := ms.roleDefaultTiers[role]; ok && tier > 0 {
			return tier
		}
	}
	return TierLow
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
// If requiredClass is non-empty, only models of that class are considered.
// Falls back to cheapest available (ignoring class) when filtering would leave no candidates.
func (ms *ModelSelector) getBestAvailableModelFromTier(tier ModelTier, role, projectID string, minContextTokens int, requiredClass ModelClass) ModelInfo {
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

	// Apply class filter when the role specifies a required class.
	// If no models of the required class exist in this tier, skip the filter so
	// the agent can still run rather than failing silently.
	if requiredClass != "" {
		classFiltered := make([]ModelInfo, 0, len(available))
		for _, m := range available {
			if m.Class == requiredClass {
				classFiltered = append(classFiltered, m)
			}
		}
		if len(classFiltered) > 0 {
			available = classFiltered
		}
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

	// No confirmed passing model — return cheapest that isn't actively failing (D/F).
	for _, m := range available {
		grade := ms.gradeManager.GetGrade(m.ID, role, projectID)
		if grade != nil && grade.ConfidenceScore > 0.1 && (grade.Grade == "D" || grade.Grade == "F") {
			continue // skip known bad models
		}
		return m
	}
	// All models have failing grades — pick cheapest as last resort.
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
