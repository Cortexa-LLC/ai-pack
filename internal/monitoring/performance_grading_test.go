package monitoring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectIDFromPath(t *testing.T) {
	t.Run("same_path_same_id", func(t *testing.T) {
		id1 := ProjectIDFromPath("/home/user/project")
		id2 := ProjectIDFromPath("/home/user/project")
		if id1 != id2 {
			t.Errorf("expected stable ID: got %q then %q", id1, id2)
		}
	})

	t.Run("different_paths_different_ids", func(t *testing.T) {
		id1 := ProjectIDFromPath("/home/user/project-a")
		id2 := ProjectIDFromPath("/home/user/project-b")
		if id1 == id2 {
			t.Errorf("expected distinct IDs for different paths, both got %q", id1)
		}
	})

	t.Run("id_is_8_hex_chars", func(t *testing.T) {
		id := ProjectIDFromPath("/some/path")
		if len(id) != 8 {
			t.Errorf("expected 8-char hex ID, got len %d: %q", len(id), id)
		}
		for _, ch := range id {
			if !strings.ContainsRune("0123456789abcdef", ch) {
				t.Errorf("expected hex char, got %q in ID %q", ch, id)
			}
		}
	})

	t.Run("empty_string_does_not_panic", func(t *testing.T) {
		id := ProjectIDFromPath("")
		if len(id) != 8 {
			t.Errorf("expected 8-char hex ID for empty path, got %q", id)
		}
	})
}

func TestPerformanceGrading(t *testing.T) {
	// Create temporary directory for test
	tmpDir := filepath.Join(os.TempDir(), "perf_grade_test")
	defer os.RemoveAll(tmpDir)

	// Create grade manager
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// Test recording successful tasks
	for i := 0; i < 10; i++ {
		err := mgr.RecordTaskCompletion(
			"task-"+string(rune(i)),
			"gpt-4o-mini",
			"engineer",
			"/test/project",
			true,  // success
			0,     // retries
			25000, // tokens
			5000,  // 5 seconds
			false, // not escalated
			false, // not downgraded
		)
		if err != nil {
			t.Errorf("Failed to record completion: %v", err)
		}
	}

	// Check grade
	grade := mgr.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade == nil {
		t.Fatal("Expected grade to exist")
	}

	if grade.TotalAttempts != 10 {
		t.Errorf("Expected 10 attempts, got %d", grade.TotalAttempts)
	}

	if grade.Successes != 10 {
		t.Errorf("Expected 10 successes, got %d", grade.Successes)
	}

	if grade.SuccessRate != 1.0 {
		t.Errorf("Expected 100%% success rate, got %.2f", grade.SuccessRate)
	}

	if grade.Grade != "A" {
		t.Errorf("Expected grade A, got %s", grade.Grade)
	}

	// Test that grade persists
	mgr2, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	grade2 := mgr2.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade2 == nil {
		t.Fatal("Expected grade to persist")
	}

	if grade2.TotalAttempts != 10 {
		t.Errorf("Expected persisted grade to have 10 attempts, got %d", grade2.TotalAttempts)
	}
}

func TestGradeCalculation(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "grade_calc_test")
	defer os.RemoveAll(tmpDir)

	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	tests := []struct {
		name          string
		successes     int
		failures      int
		retries       int
		expectedGrade string
	}{
		{"Perfect A", 20, 0, 0, "A"},
		{"Good B", 17, 3, 1, "B"},
		{"OK C", 15, 5, 3, "C"},
		{"Poor D", 12, 8, 5, "D"},
		{"Fail F", 5, 15, 10, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelID := "test-model-" + tt.name

			for i := 0; i < tt.successes; i++ {
				mgr.RecordTaskCompletion("task", modelID, "engineer", "/test", true, 0, 10000, 1000, false, false)
			}

			for i := 0; i < tt.failures; i++ {
				mgr.RecordTaskCompletion("task", modelID, "engineer", "/test", false, 0, 10000, 1000, false, false)
			}

			// Add retries to the last recorded task (approximation)
			if tt.retries > 0 {
				mgr.RecordTaskCompletion("task", modelID, "engineer", "/test", true, tt.retries, 10000, 1000, false, false)
			}

			grade := mgr.GetGrade(modelID, "engineer", "/test")
			if grade == nil {
				t.Fatal("Expected grade to exist")
			}

			if grade.Grade != tt.expectedGrade {
				t.Errorf("Expected grade %s, got %s (success rate: %.2f, retry rate: %.2f)",
					tt.expectedGrade, grade.Grade, grade.SuccessRate, grade.RetryRate)
			}
		})
	}
}

func TestComplexityAnalyzer(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		description string
		expected    ComplexityLevel
	}{
		{"Fix typo in README", ComplexityLow},
		{"Update documentation for API endpoints", ComplexityMedium}, // Update is medium complexity
		{"Implement new user authentication feature", ComplexityHigh},
		{"Refactor the entire architecture", ComplexityVeryHigh},
		{"Design a distributed microservices system", ComplexityVeryHigh},
		{"Add a simple test case", ComplexityLow},
		{"Debug a complex race condition", ComplexityHigh},
		{"Write a function to parse JSON", ComplexityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := analyzer.AnalyzeComplexity(tt.description)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for: %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestModelSelector(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "model_selector_test")
	defer os.RemoveAll(tmpDir)

	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	analyzer := NewComplexityAnalyzer()
	selector := NewModelSelector(mgr, analyzer, false, false)

	// Test 1: No history, low complexity - should use TierLow (the generic default)
	result := selector.SelectModel("engineer", "/test", "Fix a simple typo", 0)
	if result.Tier != TierLow {
		t.Errorf("Expected TierLow for simple task with no history, got %d", result.Tier)
	}

	// Test 2: No history, high complexity - should escalate
	result = selector.SelectModel("engineer", "/test", "Redesign the entire architecture with microservices", 0)
	if result.Tier < TierMedium {
		t.Errorf("Expected at least tier 3 for complex task, got %d", result.Tier)
	}

	// Test 3: Poor performance history - should escalate.
	// Record failures for the representative model of TierLow ("gpt-4.1-mini").
	for i := 0; i < 10; i++ {
		mgr.RecordTaskCompletion("task-"+string(rune(i)), "gpt-4.1-mini", "engineer", "/test", false, 0, 10000, 1000, false, false)
	}

	result = selector.SelectModel("engineer", "/test", "Normal task", 0)
	if result.Tier <= TierLow {
		t.Errorf("Expected tier escalation beyond TierLow after failures, got %d", result.Tier)
	}
	if !result.WasAdjusted {
		t.Error("Expected WasAdjusted to be true after escalation")
	}
}

func TestGradeSummary(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "grade_summary_test")
	defer os.RemoveAll(tmpDir)

	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// Create diverse grades
	models := []string{"gpt-4o-mini", "gpt-4o", "claude-sonnet-4-5"}
	roles := []string{"engineer", "orchestrator", "inspector"}

	for _, model := range models {
		for _, role := range roles {
			// Different success rates
			successes := 8
			failures := 2

			for i := 0; i < successes; i++ {
				mgr.RecordTaskCompletion("task", model, role, "/test", true, 0, 10000, 1000, false, false)
			}
			for i := 0; i < failures; i++ {
				mgr.RecordTaskCompletion("task", model, role, "/test", false, 0, 10000, 1000, false, false)
			}
		}
	}

	summary := mgr.GetSummary()

	if summary.TotalGrades != 9 { // 3 models × 3 roles
		t.Errorf("Expected 9 total grades, got %d", summary.TotalGrades)
	}

	if len(summary.ByRole) != 3 {
		t.Errorf("Expected 3 roles in summary, got %d", len(summary.ByRole))
	}

	// GetSummary pre-seeds all known catalog models, so ByModel will have at least
	// the 3 tested models (and likely more from the catalog pre-seeding).
	for _, model := range models {
		if _, exists := summary.ByModel[model]; !exists {
			t.Errorf("Expected model %s to appear in summary", model)
		}
	}
	if len(summary.ByModel) < len(models) {
		t.Errorf("Expected at least %d models in summary, got %d", len(models), len(summary.ByModel))
	}

	// Check grade distribution
	totalGradesInDistribution := 0
	for _, count := range summary.GradeDistribution {
		totalGradesInDistribution += count
	}
	if totalGradesInDistribution != 9 {
		t.Errorf("Expected 9 total grades in distribution, got %d", totalGradesInDistribution)
	}

	// Verify AverageGrade is populated for models that have data (bug-fix regression check)
	for _, model := range models {
		modelSum, exists := summary.ByModel[model]
		if !exists {
			continue
		}
		if modelSum.TotalAttempts > 0 && modelSum.AverageGrade == "" {
			t.Errorf("Model %s has %d attempts but AverageGrade is empty (attribution bug)", model, modelSum.TotalAttempts)
		}
		validGrades := map[string]bool{"A": true, "B": true, "C": true, "D": true, "F": true}
		if modelSum.TotalAttempts > 0 && !validGrades[modelSum.AverageGrade] {
			t.Errorf("Model %s has invalid AverageGrade %q (expected A/B/C/D/F)", model, modelSum.AverageGrade)
		}
	}
}

func TestConfidenceScore(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "confidence_test")
	defer os.RemoveAll(tmpDir)

	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// Test confidence progression
	tests := []struct {
		attempts        int
		expectedMinConf float64
		expectedMaxConf float64
	}{
		{1, 0.0, 0.1},
		{5, 0.2, 0.3},
		{10, 0.4, 0.6},
		{20, 0.9, 1.0},
		{30, 1.0, 1.0},
	}

	for _, tt := range tests {
		modelID := "confidence-test"
		for i := 0; i < tt.attempts; i++ {
			mgr.RecordTaskCompletion("task", modelID, "engineer", "/test", true, 0, 10000, 1000, false, false)
		}

		grade := mgr.GetGrade(modelID, "engineer", "/test")
		if grade == nil {
			t.Fatal("Expected grade to exist")
		}

		if grade.ConfidenceScore < tt.expectedMinConf || grade.ConfidenceScore > tt.expectedMaxConf {
			t.Errorf("After %d attempts, expected confidence between %.2f and %.2f, got %.2f",
				tt.attempts, tt.expectedMinConf, tt.expectedMaxConf, grade.ConfidenceScore)
		}

		// Reset for next test using a fresh directory so disk state doesn't carry over
		tmpDir = t.TempDir()
		mgr, err = NewPerformanceGradeManager(tmpDir, nil)
		if err != nil {
			t.Fatalf("Failed to reset grade manager: %v", err)
		}
	}
}

func TestModelClassFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	analyzer := NewComplexityAnalyzer()
	selector := NewModelSelector(mgr, analyzer, true, false)
	selector.SetRoleDefaultTier("engineer", TierMedium)
	selector.SetRoleRequiredClass("engineer", ClassAgentic)

	// With ClassAgentic filter, gpt-5.1-codex-mini (ClassCompletion) must not be selected.
	result := selector.SelectModel("engineer", "/test/project", "implement a feature", 0)
	if result.SelectedModel.Class == ClassCompletion {
		t.Errorf("ClassAgentic filter allowed a ClassCompletion model: %s", result.SelectedModel.ID)
	}
	if result.SelectedModel.Class != ClassAgentic {
		t.Errorf("Expected ClassAgentic model, got class=%q model=%s", result.SelectedModel.Class, result.SelectedModel.ID)
	}
}

func TestModelClassFallbackWhenNoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	analyzer := NewComplexityAnalyzer()
	// TierMinimal has no ClassCompletion models — selector must not panic or return empty.
	selector := NewModelSelector(mgr, analyzer, true, false)
	selector.SetRoleDefaultTier("codegen", TierMinimal)
	selector.SetRoleRequiredClass("codegen", ClassCompletion)

	result := selector.SelectModel("codegen", "/test/project", "generate code", 0)
	if result.SelectedModel.ID == "" {
		t.Error("Expected a model even when class filter has no matches in tier")
	}
}

// ---------------------------------------------------------------------------
// Catastrophic failure tests
// ---------------------------------------------------------------------------

// TestCatastrophicFailureImmediate verifies that RecordCatastrophicFailure
// forces Grade F on the very first call, bypassing the 5-run anchor.
func TestCatastrophicFailureImmediate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// A single catastrophic failure must immediately produce Grade F.
	err = mgr.RecordCatastrophicFailure(
		"task-cat-1",
		"gpt-4o-mini",
		"engineer",
		"/test/project",
		"agent corrupted entity.go (179 lines → 52-line stub)",
	)
	if err != nil {
		t.Fatalf("RecordCatastrophicFailure returned error: %v", err)
	}

	grade := mgr.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade == nil {
		t.Fatal("Expected grade to exist after catastrophic failure")
	}

	if grade.Grade != "F" {
		t.Errorf("Expected Grade F after catastrophic failure, got %q", grade.Grade)
	}
	if grade.CatastrophicFailures != 1 {
		t.Errorf("Expected CatastrophicFailures=1, got %d", grade.CatastrophicFailures)
	}
	if grade.Failures != 1 {
		t.Errorf("Expected Failures=1, got %d", grade.Failures)
	}
}

// TestCatastrophicFailureBypassesAnchor verifies that a seeded Grade C model
// is immediately downgraded to F on a catastrophic failure, even though it has
// fewer than minSamplesForRuntimeGrade (5) real-task executions.
func TestCatastrophicFailureBypassesAnchor(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// Simulate 2 successful runs (well below the 5-run anchor).
	for i := 0; i < 2; i++ {
		if err := mgr.RecordTaskCompletion(
			"task-ok-"+string(rune('0'+i)),
			"gpt-4o-mini", "engineer", "/test/project",
			true, 0, 10000, 3000, false, false,
		); err != nil {
			t.Fatalf("RecordTaskCompletion error: %v", err)
		}
	}

	grade := mgr.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade == nil {
		t.Fatal("Expected grade to exist after 2 successes")
	}
	// With only 2 runs, the anchor should still be holding a non-F grade.
	if grade.Grade == "F" {
		t.Fatal("Grade is already F before the catastrophic failure – test setup is wrong")
	}

	// Now record the catastrophic failure.
	if err := mgr.RecordCatastrophicFailure(
		"task-cat-1",
		"gpt-4o-mini", "engineer", "/test/project",
		"out-of-scope file write detected",
	); err != nil {
		t.Fatalf("RecordCatastrophicFailure error: %v", err)
	}

	grade = mgr.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade.Grade != "F" {
		t.Errorf("Expected Grade F after catastrophic failure, got %q", grade.Grade)
	}
}

// TestCatastrophicFailureLocksGrade verifies that once Grade F is set by a
// catastrophic failure, subsequent successful runs cannot raise the grade.
func TestCatastrophicFailureLocksGrade(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	// First: a catastrophic failure.
	if err := mgr.RecordCatastrophicFailure(
		"task-cat-1",
		"gpt-4o-mini", "engineer", "/test/project",
		"file corruption",
	); err != nil {
		t.Fatalf("RecordCatastrophicFailure error: %v", err)
	}

	// Then: many successful runs.
	for i := 0; i < 20; i++ {
		if err := mgr.RecordTaskCompletion(
			"task-ok-"+string(rune('0'+i%10)),
			"gpt-4o-mini", "engineer", "/test/project",
			true, 0, 10000, 3000, false, false,
		); err != nil {
			t.Fatalf("RecordTaskCompletion error: %v", err)
		}
	}

	grade := mgr.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade.Grade != "F" {
		t.Errorf("Expected Grade F to remain locked even after 20 successes, got %q", grade.Grade)
	}
	if grade.CatastrophicFailures != 1 {
		t.Errorf("Expected CatastrophicFailures still 1, got %d", grade.CatastrophicFailures)
	}
}

// TestCatastrophicFailurePersisted verifies that a catastrophic failure's
// Grade F survives a reload of the grade manager from disk.
func TestCatastrophicFailurePersisted(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to create grade manager: %v", err)
	}

	if err := mgr.RecordCatastrophicFailure(
		"task-cat-1",
		"gpt-4o-mini", "engineer", "/test/project",
		"file corruption",
	); err != nil {
		t.Fatalf("RecordCatastrophicFailure error: %v", err)
	}

	// Reload from the same directory.
	mgr2, err := NewPerformanceGradeManager(tmpDir, nil)
	if err != nil {
		t.Fatalf("Failed to reload grade manager: %v", err)
	}

	grade := mgr2.GetGrade("gpt-4o-mini", "engineer", "/test/project")
	if grade == nil {
		t.Fatal("Expected grade to be reloaded from disk")
	}
	if grade.Grade != "F" {
		t.Errorf("Expected persisted Grade F, got %q", grade.Grade)
	}
	if grade.CatastrophicFailures != 1 {
		t.Errorf("Expected persisted CatastrophicFailures=1, got %d", grade.CatastrophicFailures)
	}
}
