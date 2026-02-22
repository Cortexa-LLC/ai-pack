package monitoring

import (
	"fmt"
	"strings"
)

// ComplexityLevel represents the complexity of a task
type ComplexityLevel string

const (
	ComplexityLow      ComplexityLevel = "low"
	ComplexityMedium   ComplexityLevel = "medium"
	ComplexityHigh     ComplexityLevel = "high"
	ComplexityVeryHigh ComplexityLevel = "very_high"
)

// ComplexityIndicators holds patterns for different complexity levels
type ComplexityIndicators struct {
	VeryHigh []string
	High     []string
	Medium   []string
	Low      []string
}

// DefaultComplexityIndicators returns the standard complexity patterns
func DefaultComplexityIndicators() ComplexityIndicators {
	return ComplexityIndicators{
		VeryHigh: []string{
			"architecture", "architect", "design system", "redesign entire",
			"orchestrate", "coordinate multiple", "multi-service",
			"refactor entire", "rewrite from scratch", "complete overhaul",
			"distributed system", "microservices", "service mesh",
			"strategic", "roadmap", "cross-team", "multi-project",
			"breaking changes", "migration path", "backwards compatibility",
		},
		High: []string{
			"complex logic", "algorithm", "optimization", "performance tuning",
			"deep analysis", "root cause", "investigation", "debugging complex",
			"integrate with", "integration", "api design",
			"security", "authentication", "authorization", "encryption",
			"database design", "schema migration", "data model",
			"concurrency", "race condition", "threading", "async",
			"caching strategy", "load balancing", "scaling",
			"new feature", "implement feature", "add functionality",
		},
		Medium: []string{
			"implement", "create", "build", "develop",
			"refactor", "restructure", "reorganize",
			"enhance", "improve", "extend",
			"fix bug", "resolve issue", "address problem",
			"add test", "write tests", "test coverage",
			"update", "modify", "change",
			"integrate", "connect", "link",
		},
		Low: []string{
			"simple", "straightforward", "basic", "minor", "small",
			"typo", "fix typo", "correct spelling",
			"update documentation", "add comment", "improve docs",
			"rename", "move file", "reorganize folder",
			"add log", "logging", "add debug",
			"format", "formatting", "style fix",
			"following pattern", "similar to existing", "copy pattern",
			"update dependency", "bump version",
		},
	}
}

// ComplexityAnalyzer analyzes task complexity from descriptions
type ComplexityAnalyzer struct {
	indicators ComplexityIndicators
}

// NewComplexityAnalyzer creates a new complexity analyzer
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		indicators: DefaultComplexityIndicators(),
	}
}

// AnalyzeComplexity determines the complexity level of a task
func (ca *ComplexityAnalyzer) AnalyzeComplexity(description string) ComplexityLevel {
	if description == "" {
		return ComplexityMedium // Default to medium if no description
	}

	descLower := strings.ToLower(description)

	// Score each complexity level based on indicator matches
	scores := map[ComplexityLevel]int{
		ComplexityLow:      0,
		ComplexityMedium:   0,
		ComplexityHigh:     0,
		ComplexityVeryHigh: 0,
	}

	// Check for very high complexity indicators
	for _, indicator := range ca.indicators.VeryHigh {
		if strings.Contains(descLower, indicator) {
			scores[ComplexityVeryHigh] += 3 // Heavy weight
		}
	}

	// Check for high complexity indicators
	for _, indicator := range ca.indicators.High {
		if strings.Contains(descLower, indicator) {
			scores[ComplexityHigh] += 2
		}
	}

	// Check for medium complexity indicators
	for _, indicator := range ca.indicators.Medium {
		if strings.Contains(descLower, indicator) {
			scores[ComplexityMedium] += 1
		}
	}

	// Check for low complexity indicators
	for _, indicator := range ca.indicators.Low {
		if strings.Contains(descLower, indicator) {
			scores[ComplexityLow] += 1
		}
	}

	// Additional heuristics

	// Long descriptions tend to be more complex
	wordCount := len(strings.Fields(description))
	if wordCount > 200 {
		scores[ComplexityHigh]++
	} else if wordCount > 100 {
		scores[ComplexityMedium]++
	}

	// Multiple sentences suggest more complexity
	sentenceCount := strings.Count(description, ".") + strings.Count(description, "!") + strings.Count(description, "?")
	if sentenceCount > 10 {
		scores[ComplexityHigh]++
	} else if sentenceCount > 5 {
		scores[ComplexityMedium]++
	}

	// Code snippets or technical terms suggest complexity
	if strings.Contains(description, "```") || strings.Contains(description, "`") {
		scores[ComplexityMedium]++
	}

	// Multiple bullet points or lists suggest detailed work
	bulletCount := strings.Count(description, "- ") + strings.Count(description, "* ") + strings.Count(description, "• ")
	if bulletCount > 10 {
		scores[ComplexityHigh]++
	} else if bulletCount > 5 {
		scores[ComplexityMedium]++
	}

	// Find the highest scoring complexity level
	maxScore := 0
	result := ComplexityMedium // Default

	// Check from highest to lowest
	if scores[ComplexityVeryHigh] > maxScore {
		maxScore = scores[ComplexityVeryHigh]
		result = ComplexityVeryHigh
	}
	if scores[ComplexityHigh] > maxScore {
		maxScore = scores[ComplexityHigh]
		result = ComplexityHigh
	}
	if scores[ComplexityMedium] > maxScore {
		maxScore = scores[ComplexityMedium]
		result = ComplexityMedium
	}
	if scores[ComplexityLow] > maxScore {
		maxScore = scores[ComplexityLow]
		result = ComplexityLow
	}

	// If no clear winner, default to medium
	if maxScore == 0 {
		return ComplexityMedium
	}

	return result
}

// AnalyzeWithDetails returns complexity with scoring details
func (ca *ComplexityAnalyzer) AnalyzeWithDetails(description string) (ComplexityLevel, map[string]interface{}) {
	complexity := ca.AnalyzeComplexity(description)

	details := map[string]interface{}{
		"complexity":   complexity,
		"word_count":   len(strings.Fields(description)),
		"char_count":   len(description),
		"has_code":     strings.Contains(description, "```") || strings.Contains(description, "`"),
		"bullet_count": strings.Count(description, "- ") + strings.Count(description, "* "),
	}

	// Find matched indicators
	descLower := strings.ToLower(description)
	matched := []string{}

	allIndicators := append(ca.indicators.VeryHigh, ca.indicators.High...)
	allIndicators = append(allIndicators, ca.indicators.Medium...)
	allIndicators = append(allIndicators, ca.indicators.Low...)

	for _, indicator := range allIndicators {
		if strings.Contains(descLower, indicator) {
			matched = append(matched, indicator)
			if len(matched) >= 10 {
				break // Don't overwhelm with matches
			}
		}
	}

	details["matched_indicators"] = matched

	return complexity, details
}

// GetMinimumTier returns the minimum model tier for a complexity level
func GetMinimumTier(complexity ComplexityLevel) int {
	switch complexity {
	case ComplexityLow:
		return 1 // gpt-4o-mini, haiku
	case ComplexityMedium:
		return 2 // gpt-4o
	case ComplexityHigh:
		return 3 // claude-sonnet
	case ComplexityVeryHigh:
		return 3 // claude-sonnet (can escalate to opus if grade is poor)
	default:
		return 2 // Default to tier 2
	}
}

// Global complexity analyzer instance
var GlobalComplexityAnalyzer *ComplexityAnalyzer

// InitComplexityAnalyzer initializes the global complexity analyzer
func InitComplexityAnalyzer() {
	GlobalComplexityAnalyzer = NewComplexityAnalyzer()
}

// ── Debug complexity gate ─────────────────────────────────────────────────────

// ComplexityAssessment captures the result of a debug-task complexity pre-check.
type ComplexityAssessment struct {
	// Level is the overall complexity of the task description.
	Level ComplexityLevel

	// DebugSignals are the bug/debug keywords found in the task description.
	DebugSignals []string

	// MultiModuleSignals are cross-module indicators found in the task description.
	MultiModuleSignals []string

	// Recommendation is a human-readable guidance string for the orchestrator.
	Recommendation string
}

// debugTaskPatterns are keywords that indicate this is a bug-fix / debug task.
var debugTaskPatterns = []string{
	"bug", "fix", "debug", "error", "crash", "fail", "broken", "issue",
	"regression", "flaky", "panic", "exception", "traceback", "stack trace",
	"not working", "incorrect", "unexpected", "wrong output",
}

// multiModulePatterns indicate that the issue spans more than one module.
var multiModulePatterns = []string{
	"multiple modules", "across modules", "several packages",
	"multiple files", "multiple services", "cross-service",
	"multiple components", "across packages", "end-to-end",
	"integration", "distributed", "microservice",
	"cascading", "propagates", "affects multiple",
}

// AssessDebugComplexity is a pre-spawn complexity gate for debug tasks.
//
// When the task description contains bug/debug signals and is assessed as
// high-complexity AND spans multiple modules, the function returns
// (assessment, true), signalling that a deeper investigation may be warranted
// before proceeding with the task.
//
// When the task is simple or contains no debug signals, the function returns
// (assessment, false) and the caller should proceed normally.
func AssessDebugComplexity(role, taskDescription string) (ComplexityAssessment, bool) {
	assessment := ComplexityAssessment{}

	lower := strings.ToLower(taskDescription)

	// Collect matching debug signals.
	for _, p := range debugTaskPatterns {
		if strings.Contains(lower, p) {
			assessment.DebugSignals = append(assessment.DebugSignals, p)
		}
	}

	// No debug signals → not a debug task; proceed normally.
	if len(assessment.DebugSignals) == 0 {
		return assessment, false
	}

	// Collect matching multi-module signals.
	for _, p := range multiModulePatterns {
		if strings.Contains(lower, p) {
			assessment.MultiModuleSignals = append(assessment.MultiModuleSignals, p)
		}
	}

	// Determine overall complexity using the standard analyzer.
	analyzer := NewComplexityAnalyzer()
	assessment.Level = analyzer.AnalyzeComplexity(taskDescription)

	// Trigger the investigation gate only when both conditions hold:
	//   1. the task is rated High or VeryHigh complexity, AND
	//   2. at least one multi-module signal was detected.
	needsInvestigation := (assessment.Level == ComplexityHigh || assessment.Level == ComplexityVeryHigh) &&
		len(assessment.MultiModuleSignals) > 0

	if needsInvestigation {
		assessment.Recommendation = fmt.Sprintf(
			"Complex debug task detected (complexity: %s) with multi-module signals: [%s]. "+
				"Recommend routing through the Inspector role for root-cause analysis before "+
				"assigning to an engineer to avoid thrashing on an unbounded debugging session.",
			assessment.Level,
			strings.Join(assessment.MultiModuleSignals, ", "),
		)
	}

	return assessment, needsInvestigation
}

