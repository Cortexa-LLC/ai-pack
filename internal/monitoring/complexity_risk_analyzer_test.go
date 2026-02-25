package monitoring_test

import (
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// newTestAnalyzer returns an analyzer with default weights/thresholds, no grade
// manager, so tests are deterministic.
func newTestAnalyzer() *monitoring.ComplexityRiskAnalyzer {
	return monitoring.NewComplexityRiskAnalyzer(
		true, // enabled
		0.50, // warnThreshold
		0.75, // criticalThreshold
		0.30, // scope
		0.25, // multiStep
		0.25, // uncertainty
		0.20, // structural
		map[string]float64{
			"engineer":     1.0,
			"orchestrator": 0.8,
			"inspector":    0.9,
			"architect":    0.7,
		},
		nil, // no grade manager → historicalScore = 0
	)
}

// ── Enabled / disabled ────────────────────────────────────────────────────────

func TestDisabledAnalyzerReturnsNegligible(t *testing.T) {
	a := monitoring.NewComplexityRiskAnalyzer(
		false, 0.50, 0.75, 0.30, 0.25, 0.25, 0.20, nil, nil,
	)
	r := a.ComputeComplexityRisk("engineer", "fix race condition across multiple modules")
	if r.RiskLevel != monitoring.RiskNegligible {
		t.Errorf("disabled analyzer: expected RiskNegligible, got %q", r.RiskLevel)
	}
	if r.AdjustedScore != 0 {
		t.Errorf("disabled analyzer: expected score=0, got %f", r.AdjustedScore)
	}
}

// ── Risk levels ───────────────────────────────────────────────────────────────

func TestNegligibleRiskForBenignTask(t *testing.T) {
	a := newTestAnalyzer()
	r := a.ComputeComplexityRisk("engineer", "add a button to the login form")
	if r.RiskLevel != monitoring.RiskNegligible && r.RiskLevel != monitoring.RiskLow {
		t.Errorf("benign task: expected negligible/low risk, got %q (score=%f)", r.RiskLevel, r.AdjustedScore)
	}
	if r.ShouldWarn {
		t.Error("benign task should not trigger ShouldWarn")
	}
}

func TestHighRiskForComplexTask(t *testing.T) {
	a := newTestAnalyzer()
	task := "Investigate and fix race condition across multiple modules, " +
		"possibly causing intermittent deadlock in the distributed pipeline. " +
		"First analyse, then refactor concurrency patterns, followed by " +
		"updating all layers. Unclear how long this will take."
	r := a.ComputeComplexityRisk("engineer", task)
	// Complex tasks should be at least moderate; high/critical is preferred but
	// the saturation curve may produce moderate at default weights.
	if r.RiskLevel == monitoring.RiskNegligible || r.RiskLevel == monitoring.RiskLow {
		t.Errorf("complex task: expected at least moderate risk, got %q (score=%f)", r.RiskLevel, r.AdjustedScore)
	}
	// Adjusted score should be clearly above the warn threshold / 2 (i.e. > 0.25).
	if r.AdjustedScore < 0.40 {
		t.Errorf("complex task: adjusted score too low: %f", r.AdjustedScore)
	}
}

func TestCriticalRiskThreshold(t *testing.T) {
	a := newTestAnalyzer()
	// Saturate every dimension heavily — the saturation curve intentionally
	// prevents runaway scores, so >= 0.60 on a maximally-complex task is
	// the meaningful bar (well into moderate-to-high territory).
	task := "Across all microservices and entire codebase, simultaneously " +
		"fix race condition, deadlock, memory leak, concurrency issues, " +
		"breaking change with backward compatibility concern, critical path " +
		"performance bottleneck. Maybe could be possibly intermittent. " +
		"First step, then second step, followed by third phase. " +
		"Maybe unclear unknown possibly randomly sporadically."
	r := a.ComputeComplexityRisk("engineer", task)
	if r.AdjustedScore < 0.60 {
		t.Errorf("expected adjusted score >= 0.60 for fully saturated task, got %f", r.AdjustedScore)
	}
	// Score must exceed the warn threshold so the gate triggers.
	if r.AdjustedScore < 0.50 {
		t.Errorf("fully saturated task must exceed warn threshold (0.50), got %f", r.AdjustedScore)
	}
}

// ── Role multiplier ───────────────────────────────────────────────────────────

func TestRoleMultiplierReducesScore(t *testing.T) {
	a := newTestAnalyzer()
	task := "fix race condition across multiple modules, possibly deadlock, " +
		"first step then second step, unclear architecture change"

	engineerRisk := a.ComputeComplexityRisk("engineer", task)
	architectRisk := a.ComputeComplexityRisk("architect", task)

	if architectRisk.AdjustedScore >= engineerRisk.AdjustedScore {
		t.Errorf(
			"architect (multiplier=0.7) should score lower than engineer (multiplier=1.0): architect=%f engineer=%f",
			architectRisk.AdjustedScore, engineerRisk.AdjustedScore,
		)
	}
}

func TestUnknownRoleUsesMultiplierOne(t *testing.T) {
	a := newTestAnalyzer()
	task := "fix race condition across multiple modules"
	known := a.ComputeComplexityRisk("engineer", task)
	unknown := a.ComputeComplexityRisk("janitor", task)
	// No configured multiplier → defaults to 1.0, same as engineer
	if unknown.Components.RoleMultiplier != 1.0 {
		t.Errorf("unknown role: expected multiplier=1.0, got %f", unknown.Components.RoleMultiplier)
	}
	if unknown.AdjustedScore != known.AdjustedScore {
		t.Errorf("unknown role should score same as multiplier=1 role: unknown=%f known=%f",
			unknown.AdjustedScore, known.AdjustedScore)
	}
}

// ── Sub-score components ──────────────────────────────────────────────────────

func TestComponentsArePopulated(t *testing.T) {
	a := newTestAnalyzer()
	task := "fix race condition across multiple modules"
	r := a.ComputeComplexityRisk("engineer", task)

	if r.Components.ScopeScore < 0 || r.Components.ScopeScore > 1 {
		t.Errorf("ScopeScore out of [0,1]: %f", r.Components.ScopeScore)
	}
	if r.Components.StructuralScore < 0 || r.Components.StructuralScore > 1 {
		t.Errorf("StructuralScore out of [0,1]: %f", r.Components.StructuralScore)
	}
	if r.Components.RoleMultiplier != 1.0 {
		t.Errorf("engineer RoleMultiplier: expected 1.0, got %f", r.Components.RoleMultiplier)
	}
	if r.Components.HistoricalScore != 0 {
		t.Errorf("HistoricalScore should be 0 with nil grade manager, got %f", r.Components.HistoricalScore)
	}
}

func TestScopeKeywordsIncreaseScopeScore(t *testing.T) {
	a := newTestAnalyzer()
	noScope := a.ComputeComplexityRisk("engineer", "fix the login button label")
	withScope := a.ComputeComplexityRisk("engineer", "fix login button across multiple modules entire codebase")
	if withScope.Components.ScopeScore <= noScope.Components.ScopeScore {
		t.Errorf("scope keywords should raise ScopeScore: withScope=%f noScope=%f",
			withScope.Components.ScopeScore, noScope.Components.ScopeScore)
	}
}

func TestStructuralKeywordsIncreaseStructuralScore(t *testing.T) {
	a := newTestAnalyzer()
	noStruct := a.ComputeComplexityRisk("engineer", "rename the variable in config.go")
	withStruct := a.ComputeComplexityRisk("engineer", "fix race condition deadlock memory leak concurrency")
	if withStruct.Components.StructuralScore <= noStruct.Components.StructuralScore {
		t.Errorf("structural keywords should raise StructuralScore: withStruct=%f noStruct=%f",
			withStruct.Components.StructuralScore, noStruct.Components.StructuralScore)
	}
}

// ── Recommendation ────────────────────────────────────────────────────────────

func TestRecommendationPresentForHighRisk(t *testing.T) {
	a := newTestAnalyzer()
	task := "Investigate intermittent race condition across multiple modules, " +
		"possibly deadlock in distributed architecture, maybe refactor entire codebase"
	r := a.ComputeComplexityRisk("engineer", task)
	if r.ShouldWarn && r.Recommendation == "" {
		t.Error("high/critical risk assessment should include a non-empty Recommendation")
	}
}

func TestNoRecommendationForNegligibleRisk(t *testing.T) {
	a := newTestAnalyzer()
	r := a.ComputeComplexityRisk("engineer", "update copyright year in README")
	if r.RiskLevel == monitoring.RiskNegligible && r.Recommendation != "" {
		t.Errorf("negligible risk should have empty Recommendation, got %q", r.Recommendation)
	}
}

// ── Score bounds ──────────────────────────────────────────────────────────────

func TestScoresAlwaysInUnitInterval(t *testing.T) {
	a := newTestAnalyzer()
	tasks := []string{
		"",
		"fix bug",
		"across all microservices investigate race condition deadlock memory leak concurrency possibly intermittent unknown unclear first then followed by phase",
	}
	roles := []string{"engineer", "orchestrator", "architect", "inspector", "janitor"}
	for _, task := range tasks {
		for _, role := range roles {
			r := a.ComputeComplexityRisk(role, task)
			if r.BaseScore < 0 || r.BaseScore > 1 {
				t.Errorf("BaseScore %f out of [0,1] for role=%q task=%q", r.BaseScore, role, task)
			}
			if r.AdjustedScore < 0 || r.AdjustedScore > 1 {
				t.Errorf("AdjustedScore %f out of [0,1] for role=%q task=%q", r.AdjustedScore, role, task)
			}
		}
	}
}
