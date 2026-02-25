package monitoring

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/config"
)

const (
	// GradeSourceBenchmark identifies grades produced by the benchmark script.
	GradeSourceBenchmark = "benchmark"
	// GradeSourceProduction identifies grades recorded from live task executions.
	GradeSourceProduction = "production"
	// GradeSourceLiveBench is the prefix used by seed-grades.py for LiveBench-derived grades.
	GradeSourceLiveBench = "livebench"

	// minSamplesForRuntimeGrade is the number of real task executions required before
	// runtime success/failure data is trusted enough to override a LiveBench-seeded grade.
	// Below this threshold, the seeded grade is preserved as the authoritative signal.
	minSamplesForRuntimeGrade = 5
)

// PerformanceGrade tracks model performance per role and project
type PerformanceGrade struct {
	ModelID   string `json:"model_id"`   // e.g., "gpt-4o-mini"
	RoleID    string `json:"role_id"`    // e.g., "engineer"
	ProjectID string `json:"project_id"` // Project path hash or identifier

	// Success metrics
	TotalAttempts int `json:"total_attempts"`
	Successes     int `json:"successes"`
	Failures      int `json:"failures"`
	Retries       int `json:"retries"`

	// Quality indicators
	TotalTokensUsed      int64   `json:"total_tokens_used"`       // Sum of all tokens
	TotalExecutionTimeMs float64 `json:"total_execution_time_ms"` // Sum of all execution times (ms)
	AverageTokens        int     `json:"average_tokens"`          // Calculated
	AverageExecutionTime float64 `json:"average_execution_time"`  // Calculated in seconds
	ErrorRate            float64 `json:"error_rate"`              // Failures / TotalAttempts
	RetryRate            float64 `json:"retry_rate"`              // Retries / TotalAttempts
	SuccessRate          float64 `json:"success_rate"`            // Successes / TotalAttempts

	// Escalation tracking
	EscalationCount int `json:"escalation_count"` // Times we had to escalate
	DowngradeCount  int `json:"downgrade_count"`  // Times we successfully downgraded

	// Calculated grade
	Grade           string  `json:"grade"`            // A, B, C, D, F
	ConfidenceScore float64 `json:"confidence_score"` // 0.0-1.0 (higher = more data)

	// Time tracking
	LastUsed  time.Time `json:"last_used"`
	FirstUsed time.Time `json:"first_used"`

	// Metadata
	LastTaskID string `json:"last_task_id,omitempty"` // For debugging
	Source     string `json:"source,omitempty"`       // GradeSourceBenchmark or GradeSourceProduction
}

// PerformanceGradeManager manages performance grades with persistent storage
type PerformanceGradeManager struct {
	mu         sync.RWMutex
	grades     map[string]*PerformanceGrade // key: "model:role:project"
	storageDir string
	criteria   *config.GradingCriteriaConfig
}

// NewPerformanceGradeManager creates a new performance grade manager
func NewPerformanceGradeManager(storageDir string, criteria *config.GradingCriteriaConfig) (*PerformanceGradeManager, error) {
	// Use default criteria if none provided
	if criteria == nil {
		defaultCfg := config.DefaultConfig()
		criteria = &defaultCfg.GradingCriteria
	}

	mgr := &PerformanceGradeManager{
		grades:     make(map[string]*PerformanceGrade),
		storageDir: storageDir,
		criteria:   criteria,
	}

	// Create storage directory if needed
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create grades storage directory: %w", err)
	}

	// Load existing grades
	if err := mgr.loadGrades(); err != nil {
		Logger.Warn("failed_to_load_existing_grades", "error", err.Error())
		// Continue anyway - not fatal
	}

	return mgr, nil
}

// gradeKey creates a unique key for a grade
func gradeKey(modelID, roleID, projectID string) string {
	return fmt.Sprintf("%s:%s:%s", modelID, roleID, projectID)
}

// RecordTaskCompletion records the outcome of a task
func (m *PerformanceGradeManager) RecordTaskCompletion(
	taskID string,
	modelID string,
	roleID string,
	projectID string,
	success bool,
	retries int,
	tokensUsed int64,
	executionTimeMs int64,
	wasEscalated bool,
	wasDowngraded bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := gradeKey(modelID, roleID, projectID)
	grade, exists := m.grades[key]

	if !exists {
		// Create new grade
		grade = &PerformanceGrade{
			ModelID:   modelID,
			RoleID:    roleID,
			ProjectID: projectID,
			FirstUsed: time.Now(),
			Source:    GradeSourceProduction,
		}
		m.grades[key] = grade
	}

	// Update metrics
	grade.TotalAttempts++
	grade.LastUsed = time.Now()
	grade.LastTaskID = taskID

	if success {
		grade.Successes++
	} else {
		grade.Failures++
	}

	grade.Retries += retries
	grade.TotalTokensUsed += tokensUsed
	grade.TotalExecutionTimeMs += float64(executionTimeMs)

	if wasEscalated {
		grade.EscalationCount++
	}
	if wasDowngraded {
		grade.DowngradeCount++
	}

	// Recalculate derived metrics
	m.recalculateGrade(grade)

	// Persist to disk
	if err := m.saveGrade(grade); err != nil {
		return fmt.Errorf("failed to save grade: %w", err)
	}

	return nil
}

// recalculateGrade updates all calculated fields
func (m *PerformanceGradeManager) recalculateGrade(grade *PerformanceGrade) {
	if grade.TotalAttempts == 0 {
		return
	}

	// Calculate rates
	grade.SuccessRate = float64(grade.Successes) / float64(grade.TotalAttempts)
	grade.ErrorRate = float64(grade.Failures) / float64(grade.TotalAttempts)
	grade.RetryRate = float64(grade.Retries) / float64(grade.TotalAttempts)

	// Calculate averages
	if grade.TotalAttempts > 0 {
		grade.AverageTokens = int(grade.TotalTokensUsed / int64(grade.TotalAttempts))
		grade.AverageExecutionTime = grade.TotalExecutionTimeMs / float64(grade.TotalAttempts) / 1000.0
	}

	// Calculate confidence score (0.0 to 1.0)
	// Full confidence at 20+ samples
	grade.ConfidenceScore = math.Min(1.0, float64(grade.TotalAttempts)/20.0)

	// Preserve LiveBench-seeded grades until we have enough real samples to trust
	// runtime data. Task completion (success=true) only means the agent finished
	// without crashing — it says nothing about output quality. The LiveBench coding
	// score is a more reliable capability signal for low-sample situations.
	if strings.HasPrefix(grade.Source, GradeSourceLiveBench) && grade.TotalAttempts < minSamplesForRuntimeGrade {
		// Rates/averages above are updated for visibility, but the letter grade
		// stays anchored to the seeded value until we have minSamplesForRuntimeGrade runs.
		return
	}

	// Calculate letter grade based on runtime success rate and retry rate
	grade.Grade = m.calculateLetterGrade(grade.SuccessRate, grade.RetryRate)
}

// calculateLetterGrade determines letter grade from metrics using configurable criteria
func (m *PerformanceGradeManager) calculateLetterGrade(successRate, retryRate float64) string {
	// Grade A
	if successRate >= m.criteria.GradeA.MinSuccessRate && retryRate < m.criteria.GradeA.MaxRetryRate {
		return "A"
	}

	// Grade B
	if successRate >= m.criteria.GradeB.MinSuccessRate && retryRate < m.criteria.GradeB.MaxRetryRate {
		return "B"
	}

	// Grade C
	if successRate >= m.criteria.GradeC.MinSuccessRate && retryRate < m.criteria.GradeC.MaxRetryRate {
		return "C"
	}

	// Grade D
	if successRate >= m.criteria.GradeD.MinSuccessRate && retryRate < m.criteria.GradeD.MaxRetryRate {
		return "D"
	}

	// Grade F: Everything else
	return "F"
}

// GetGrade retrieves a grade for model/role/project
func (m *PerformanceGradeManager) GetGrade(modelID, roleID, projectID string) *PerformanceGrade {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := gradeKey(modelID, roleID, projectID)
	if grade, exists := m.grades[key]; exists {
		// Return a copy to prevent external modification
		gradeCopy := *grade
		return &gradeCopy
	}

	return nil
}

// GetGradesByRole retrieves all grades for a specific role
func (m *PerformanceGradeManager) GetGradesByRole(roleID string) []*PerformanceGrade {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var grades []*PerformanceGrade
	for _, grade := range m.grades {
		if grade.RoleID == roleID {
			gradeCopy := *grade
			grades = append(grades, &gradeCopy)
		}
	}

	return grades
}

// GetGradesByProject retrieves all grades for a specific project
func (m *PerformanceGradeManager) GetGradesByProject(projectID string) []*PerformanceGrade {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var grades []*PerformanceGrade
	for _, grade := range m.grades {
		if grade.ProjectID == projectID {
			gradeCopy := *grade
			grades = append(grades, &gradeCopy)
		}
	}

	return grades
}

// GetAllGrades retrieves all grades
func (m *PerformanceGradeManager) GetAllGrades() []*PerformanceGrade {
	m.mu.RLock()
	defer m.mu.RUnlock()

	grades := make([]*PerformanceGrade, 0, len(m.grades))
	for _, grade := range m.grades {
		gradeCopy := *grade
		grades = append(grades, &gradeCopy)
	}

	return grades
}

// saveGrade persists a grade to disk
func (m *PerformanceGradeManager) saveGrade(grade *PerformanceGrade) error {
	filename := fmt.Sprintf("%s_%s_%s.json",
		sanitizeFilename(grade.ModelID),
		sanitizeFilename(grade.RoleID),
		sanitizeFilename(grade.ProjectID))

	path := filepath.Join(m.storageDir, filename)

	data, err := json.MarshalIndent(grade, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal grade: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write grade file: %w", err)
	}

	return nil
}

// loadGrades loads all grades from disk
func (m *PerformanceGradeManager) loadGrades() error {
	files, err := os.ReadDir(m.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No grades yet
		}
		return fmt.Errorf("failed to read grades directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(m.storageDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			if Logger != nil {
				Logger.Warn("failed_to_read_grade_file", "file", file.Name(), "error", err.Error())
			}
			continue
		}

		var grade PerformanceGrade
		if err := json.Unmarshal(data, &grade); err != nil {
			if Logger != nil {
				Logger.Warn("failed_to_unmarshal_grade", "file", file.Name(), "error", err.Error())
			}
			continue
		}

		key := gradeKey(grade.ModelID, grade.RoleID, grade.ProjectID)
		m.grades[key] = &grade
	}

	if Logger != nil {
		Logger.Info("loaded_performance_grades", "count", len(m.grades))
	}
	return nil
}

// LoadGradesFromDirectory loads grades from a specific directory and merges them into the manager
func (m *PerformanceGradeManager) LoadGradesFromDirectory(directory string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No grades in this directory
		}
		return fmt.Errorf("failed to read grades directory: %w", err)
	}

	loadedCount := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(directory, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			if Logger != nil {
				Logger.Warn("failed_to_read_grade_file", "file", file.Name(), "error", err.Error())
			}
			continue
		}

		var grade PerformanceGrade
		if err := json.Unmarshal(data, &grade); err != nil {
			if Logger != nil {
				Logger.Warn("failed_to_unmarshal_grade", "file", file.Name(), "error", err.Error())
			}
			continue
		}

		key := gradeKey(grade.ModelID, grade.RoleID, grade.ProjectID)

		// If grade already exists, merge the data (sum up attempts)
		if existing, exists := m.grades[key]; exists {
			existing.TotalAttempts += grade.TotalAttempts
			existing.Successes += grade.Successes
			existing.Failures += grade.Failures
			existing.Retries += grade.Retries
			existing.TotalTokensUsed += grade.TotalTokensUsed
			existing.TotalExecutionTimeMs += grade.TotalExecutionTimeMs
			existing.EscalationCount += grade.EscalationCount
			existing.DowngradeCount += grade.DowngradeCount

			// Update timestamps
			if grade.FirstUsed.Before(existing.FirstUsed) {
				existing.FirstUsed = grade.FirstUsed
			}
			if grade.LastUsed.After(existing.LastUsed) {
				existing.LastUsed = grade.LastUsed
				existing.LastTaskID = grade.LastTaskID
			}

			// Recalculate derived metrics
			m.recalculateGrade(existing)
		} else {
			// Add new grade
			m.grades[key] = &grade
		}

		loadedCount++
	}

	if Logger != nil && loadedCount > 0 {
		Logger.Info("loaded_grades_from_directory", "directory", directory, "count", loadedCount)
	}

	return nil
}

// sanitizeFilename removes characters that aren't safe for filenames
func sanitizeFilename(s string) string {
	// Replace problematic characters
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	return s
}

// GetGradeSummary returns a summary of all grades for reporting
type GradeSummary struct {
	TotalGrades       int                     `json:"total_grades"`
	GradeDistribution map[string]int          `json:"grade_distribution"` // A: 10, B: 5, etc.
	ByRole            map[string]RoleSummary  `json:"by_role"`
	ByModel           map[string]ModelSummary `json:"by_model"`
	TopPerformers     []PerformanceGrade      `json:"top_performers"`    // Best grades
	NeedsImprovement  []PerformanceGrade      `json:"needs_improvement"` // Worst grades
}

type RoleSummary struct {
	TotalAttempts int            `json:"total_attempts"`
	Successes     int            `json:"successes"`
	Failures      int            `json:"failures"`
	SuccessRate   float64        `json:"success_rate"`
	Models        map[string]int `json:"models"` // Model usage count
}

type ModelSummary struct {
	TotalAttempts int     `json:"total_attempts"`
	Successes     int     `json:"successes"`
	Failures      int     `json:"failures"`
	TotalRetries  int     `json:"total_retries"`
	SuccessRate   float64 `json:"success_rate"`
	AverageGrade  string  `json:"average_grade"`
}

// GetSummary returns a comprehensive summary of all grades
func (m *PerformanceGradeManager) GetSummary() GradeSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := GradeSummary{
		TotalGrades:       len(m.grades),
		GradeDistribution: make(map[string]int),
		ByRole:            make(map[string]RoleSummary),
		ByModel:           make(map[string]ModelSummary),
	}

	// Pre-seed ByModel with every known model so the UI always shows all models
	// even before they have accumulated any performance history.
	for _, models := range ModelsByTier {
		for _, info := range models {
			if _, exists := summary.ByModel[info.ID]; !exists {
				summary.ByModel[info.ID] = ModelSummary{}
			}
		}
	}

	// Collect all grades for analysis
	allGrades := make([]PerformanceGrade, 0, len(m.grades))
	for _, grade := range m.grades {
		allGrades = append(allGrades, *grade)

		// Count grade distribution
		summary.GradeDistribution[grade.Grade]++

		// Aggregate by role
		roleSum := summary.ByRole[grade.RoleID]
		roleSum.TotalAttempts += grade.TotalAttempts
		roleSum.Successes += grade.Successes
		roleSum.Failures += grade.Failures
		if roleSum.Models == nil {
			roleSum.Models = make(map[string]int)
		}
		roleSum.Models[grade.ModelID] += grade.TotalAttempts
		if roleSum.TotalAttempts > 0 {
			roleSum.SuccessRate = float64(roleSum.Successes) / float64(roleSum.TotalAttempts)
		}
		summary.ByRole[grade.RoleID] = roleSum

		// Aggregate by model
		modelSum := summary.ByModel[grade.ModelID]
		modelSum.TotalAttempts += grade.TotalAttempts
		modelSum.Successes += grade.Successes
		modelSum.Failures += grade.Failures
		modelSum.TotalRetries += grade.Retries
		if modelSum.TotalAttempts > 0 {
			modelSum.SuccessRate = float64(modelSum.Successes) / float64(modelSum.TotalAttempts)
		}
		summary.ByModel[grade.ModelID] = modelSum
	}

	// Find top performers (Grade A with high confidence)
	for _, grade := range allGrades {
		if grade.Grade == "A" && grade.ConfidenceScore >= 0.5 {
			summary.TopPerformers = append(summary.TopPerformers, grade)
			if len(summary.TopPerformers) >= 10 {
				break
			}
		}
	}

	// Find needs improvement (Grade D/F with high confidence)
	for _, grade := range allGrades {
		if (grade.Grade == "D" || grade.Grade == "F") && grade.ConfidenceScore >= 0.5 {
			summary.NeedsImprovement = append(summary.NeedsImprovement, grade)
			if len(summary.NeedsImprovement) >= 10 {
				break
			}
		}
	}

	// Calculate AverageGrade for each model using aggregated success/retry rates
	for modelID, modelSum := range summary.ByModel {
		if modelSum.TotalAttempts > 0 {
			retryRate := float64(modelSum.TotalRetries) / float64(modelSum.TotalAttempts)
			modelSum.AverageGrade = m.calculateLetterGrade(modelSum.SuccessRate, retryRate)
			summary.ByModel[modelID] = modelSum
		}
	}

	return summary
}

// ShouldEscalate determines if we should escalate to a higher tier model
func (m *PerformanceGradeManager) ShouldEscalate(modelID, roleID, projectID string) bool {
	grade := m.GetGrade(modelID, roleID, projectID)
	if grade == nil {
		return false // No history yet
	}

	// Only escalate if we have enough confidence in the data
	if grade.ConfidenceScore < 0.5 {
		return false
	}

	// Escalate on Grade D or F
	return grade.Grade == "D" || grade.Grade == "F"
}

// ShouldDowngrade determines if we can downgrade to a cheaper model
func (m *PerformanceGradeManager) ShouldDowngrade(modelID, roleID, projectID string, consecutiveSuccesses int) bool {
	grade := m.GetGrade(modelID, roleID, projectID)
	if grade == nil {
		return false
	}

	// Only downgrade if we have high confidence
	if grade.ConfidenceScore < 0.75 {
		return false
	}

	// Downgrade on Grade A with many consecutive successes
	return grade.Grade == "A" && consecutiveSuccesses >= 10
}

// ReloadGrades re-reads all grade JSON files from the storage directory and
// replaces the in-memory grades map.  Safe for concurrent use.
func (m *PerformanceGradeManager) ReloadGrades() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset the grades map before reloading to remove stale entries.
	m.grades = make(map[string]*PerformanceGrade)

	files, err := os.ReadDir(m.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No grades yet; not an error
		}
		return fmt.Errorf("failed to read grades directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(m.storageDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable files
		}

		var grade PerformanceGrade
		if err := json.Unmarshal(data, &grade); err != nil {
			continue // Skip malformed files
		}

		key := fmt.Sprintf("%s:%s:%s", grade.ModelID, grade.RoleID, grade.ProjectID)
		m.grades[key] = &grade
	}

	return nil
}

// Global performance grade manager instance
var GlobalGradeManager *PerformanceGradeManager

// InitGradeManager initializes the global performance grade manager
func InitGradeManager(storageDir string, criteria *config.GradingCriteriaConfig) error {
	mgr, err := NewPerformanceGradeManager(storageDir, criteria)
	if err != nil {
		return err
	}
	GlobalGradeManager = mgr
	return nil
}
