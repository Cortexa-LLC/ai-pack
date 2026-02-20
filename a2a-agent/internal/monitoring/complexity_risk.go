package monitoring

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// v2: Composite structural risk scorer
// ---------------------------------------------------------------------------

// RiskLevel is an ordered severity tier for the v2 composite scorer.
type RiskLevel string

const (
	RiskNegligible RiskLevel = "negligible"
	RiskLow        RiskLevel = "low"
	RiskModerate   RiskLevel = "moderate"
	RiskHigh       RiskLevel = "high"
	RiskCritical   RiskLevel = "critical"
)

// RiskComponents holds the raw [0,1] sub-scores that compose the final score.
type RiskComponents struct {
	ScopeScore       float64 `json:"scope_score"`
	MultiStepScore   float64 `json:"multi_step_score"`
	UncertaintyScore float64 `json:"uncertainty_score"`
	StructuralScore  float64 `json:"structural_score"`
	HistoricalScore  float64 `json:"historical_score"`
	RoleMultiplier   float64 `json:"role_multiplier"`
}

// ComplexityRiskAssessment is the v2 assessment result.
type ComplexityRiskAssessment struct {
	BaseScore      float64        `json:"base_score"`
	AdjustedScore  float64        `json:"adjusted_score"`
	RiskLevel      RiskLevel      `json:"risk_level"`
	Components     RiskComponents `json:"components"`
	Recommendation string         `json:"recommendation"`
	ShouldWarn     bool           `json:"should_warn"` // true when >= RiskHigh
}

// ComplexityRiskAnalyzer wraps the v2 composite scorer logic.
type ComplexityRiskAnalyzer struct {
	cfg          complexityGateConfig
	gradeManager *PerformanceGradeManager // may be nil
}

// complexityGateConfig is a local mirror of config.ComplexityGateConfig so
// the monitoring package does not import the config package.
type complexityGateConfig struct {
	Enabled           bool
	WarnThreshold     float64
	CriticalThreshold float64
	Weights           complexityWeights
	RoleMultipliers   map[string]float64
}

type complexityWeights struct {
	Scope       float64
	MultiStep   float64
	Uncertainty float64
	Structural  float64
	Historical  float64
}

// NewComplexityRiskAnalyzer builds a v2 scorer from raw config values.
//
// Parameters mirror config.ComplexityGateConfig to avoid a cross-package import cycle.
// The historical sub-score weight is fixed internally at 0.15 because the
// config struct does not expose it (historical data comes from the grade
// manager at runtime rather than a static weight).
func NewComplexityRiskAnalyzer(
	enabled bool,
	warnThreshold, criticalThreshold float64,
	scopeW, multiStepW, uncertaintyW, structuralW float64,
	roleMultipliers map[string]float64,
	gm *PerformanceGradeManager,
) *ComplexityRiskAnalyzer {
	const defaultHistoricalWeight = 0.15
	return &ComplexityRiskAnalyzer{
		cfg: complexityGateConfig{
			Enabled:           enabled,
			WarnThreshold:     warnThreshold,
			CriticalThreshold: criticalThreshold,
			Weights: complexityWeights{
				Scope:       scopeW,
				MultiStep:   multiStepW,
				Uncertainty: uncertaintyW,
				Structural:  structuralW,
				Historical:  defaultHistoricalWeight,
			},
			RoleMultipliers: roleMultipliers,
		},
		gradeManager: gm,
	}
}

// ---------------------------------------------------------------------------
// Signal libraries
// ---------------------------------------------------------------------------

var scopeSignals = []string{
	"across", "multiple modules", "several services", "entire codebase",
	"all layers", "end-to-end", "distributed", "microservices",
	"cross-cutting", "full stack", "everywhere", "all components",
}

var multiStepSignals = []string{
	"first", "then", "followed by", "next", "finally", "step",
	"phase", "stage", "sequence", "pipeline", "workflow",
}

var uncertaintySignals = []string{
	"maybe", "unclear", "unknown", "possibly", "intermittent",
	"randomly", "sporadically", "seems", "appears", "might", "could",
}

var structuralSignals = []string{
	"race condition", "deadlock", "memory leak", "concurrency",
	"breaking change", "backward compatibility", "critical path",
	"performance bottleneck", "security vulnerability", "data corruption",
}

// ---------------------------------------------------------------------------
// Scoring helpers
// ---------------------------------------------------------------------------

// saturate converts a raw hit ratio into a [0,1] score using a simple
// square-root saturation curve so that a single hit already registers but
// additional hits have diminishing returns.
func saturate(hits, total int) float64 {
	if total == 0 {
		return 0
	}
	ratio := float64(hits) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	// sqrt gives ~0.71 for a 50% hit-rate; saturates smoothly to 1.0
	result := 0.0
	for i := 0; i < 10; i++ {
		result = result*0.5 + ratio*0.5
	}
	// simple approximation: use the ratio itself weighted by sqrt(ratio)
	// sqrt(x) in Go is math.Sqrt but we avoid the import; use ratio^0.5 via iteration
	// Use a quick Newton's method to compute sqrt(ratio)
	x := ratio
	if x == 0 {
		return 0
	}
	// 3 iterations of Newton's: x_(n+1) = (x_n + ratio/x_n) / 2
	for i := 0; i < 6; i++ {
		x = (x + ratio/x) / 2
	}
	return x * x * (3 - 2*x) // smoothstep for a nicer curve
}

func countHits(lower string, signals []string) int {
	hits := 0
	for _, s := range signals {
		if strings.Contains(lower, s) {
			hits++
		}
	}
	return hits
}

// ---------------------------------------------------------------------------
// ComputeComplexityRisk — main entry point
// ---------------------------------------------------------------------------

// ComputeComplexityRisk runs the v2 composite scorer for any role.
//
// If the analyzer is disabled it returns a negligible assessment immediately.
func (a *ComplexityRiskAnalyzer) ComputeComplexityRisk(role, taskDescription string) ComplexityRiskAssessment {
	if !a.cfg.Enabled {
		return ComplexityRiskAssessment{RiskLevel: RiskNegligible}
	}

	lower := strings.ToLower(taskDescription)

	scopeScore := saturate(countHits(lower, scopeSignals), len(scopeSignals))
	multiStepScore := saturate(countHits(lower, multiStepSignals), len(multiStepSignals))
	uncertaintyScore := saturate(countHits(lower, uncertaintySignals), len(uncertaintySignals))
	structuralScore := saturate(countHits(lower, structuralSignals), len(structuralSignals))
	historicalScore := a.historicalScore(role)

	w := a.cfg.Weights
	totalW := w.Scope + w.MultiStep + w.Uncertainty + w.Structural + w.Historical
	if totalW == 0 {
		totalW = 1
	}

	baseScore := (scopeScore*w.Scope +
		multiStepScore*w.MultiStep +
		uncertaintyScore*w.Uncertainty +
		structuralScore*w.Structural +
		historicalScore*w.Historical) / totalW

	roleMultiplier := a.roleMultiplier(role)
	adjustedScore := baseScore * roleMultiplier
	if adjustedScore > 1.0 {
		adjustedScore = 1.0
	}

	level := a.scoreToLevel(adjustedScore)

	components := RiskComponents{
		ScopeScore:       scopeScore,
		MultiStepScore:   multiStepScore,
		UncertaintyScore: uncertaintyScore,
		StructuralScore:  structuralScore,
		HistoricalScore:  historicalScore,
		RoleMultiplier:   roleMultiplier,
	}

	return ComplexityRiskAssessment{
		BaseScore:      baseScore,
		AdjustedScore:  adjustedScore,
		RiskLevel:      level,
		Components:     components,
		Recommendation: a.recommend(level, role),
		ShouldWarn:     level == RiskHigh || level == RiskCritical,
	}
}

// historicalScore returns a [0,1] penalty when the grade manager records a
// poor recent success rate for this role.
func (a *ComplexityRiskAnalyzer) historicalScore(role string) float64 {
	if a.gradeManager == nil {
		return 0
	}
	grades := a.gradeManager.GetGradesByRole(role)
	for _, g := range grades {
		if g != nil && g.TotalAttempts >= 5 {
			failRate := 1.0 - g.SuccessRate
			if failRate < 0 {
				failRate = 0
			}
			return failRate
		}
	}
	return 0
}

// roleMultiplier returns the configured multiplier for a role, defaulting to 1.0.
func (a *ComplexityRiskAnalyzer) roleMultiplier(role string) float64 {
	if m, ok := a.cfg.RoleMultipliers[strings.ToLower(role)]; ok {
		return m
	}
	return 1.0
}

func (a *ComplexityRiskAnalyzer) scoreToLevel(score float64) RiskLevel {
	switch {
	case score < 0.10:
		return RiskNegligible
	case score < 0.30:
		return RiskLow
	case score < 0.50:
		return RiskModerate
	case score < a.cfg.CriticalThreshold:
		return RiskHigh
	default:
		return RiskCritical
	}
}

func (a *ComplexityRiskAnalyzer) recommend(level RiskLevel, role string) string {
	switch level {
	case RiskNegligible:
		return ""
	case RiskLow:
		return fmt.Sprintf("Task is well-scoped for %s. Proceed directly.", role)
	case RiskModerate:
		return fmt.Sprintf(
			"Task has moderate structural complexity for role %q. "+
				"Consider decomposing before starting.", role)
	case RiskHigh:
		return fmt.Sprintf(
			"High structural risk detected for role %q. "+
				"Recommend investigation or task decomposition before proceeding.", role)
	default:
		return fmt.Sprintf(
			"Critical structural risk for role %q. "+
				"Strong recommendation: request investigation first; "+
				"avoid direct implementation to prevent thrashing.", role)
	}
}
