package monitoring

import (
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

// GlobalComplexityAnalyzer is the global complexity analyzer instance
var GlobalComplexityAnalyzer *ComplexityAnalyzer

// InitComplexityAnalyzer initializes the global complexity analyzer
func InitComplexityAnalyzer() {
	GlobalComplexityAnalyzer = NewComplexityAnalyzer()
}
