package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cortexa-llc/ai-pack/internal/config"
	"github.com/cortexa-llc/ai-pack/internal/constants"
	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

// HandlePerformanceGrades returns all performance grades
func (s *AgentServer) HandlePerformanceGrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if monitoring.GlobalGradeManager == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"grades":  []interface{}{},
			"enabled": false,
		})
		return
	}

	grades := monitoring.GlobalGradeManager.GetAllGrades()

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grades":  grades,
		"enabled": true,
	})
}

// HandlePerformanceSummary returns aggregated performance summary
func (s *AgentServer) HandlePerformanceSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if monitoring.GlobalGradeManager == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
			"message": "Performance grading not enabled",
		})
		return
	}

	summary := monitoring.GlobalGradeManager.GetSummary()

	// Calculate cost savings estimate
	costSavings := s.calculateCostSavings(summary)

	response := map[string]interface{}{
		"enabled":       true,
		"summary":       summary,
		"cost_savings":  costSavings,
		"model_tiers":   getModelTierInfo(),
		"grade_weights": getGradeWeights(),
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandlePerformanceByRole returns performance grades filtered by role
func (s *AgentServer) HandlePerformanceByRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roleID := r.URL.Query().Get("role")
	if roleID == "" {
		http.Error(w, "Missing role parameter", http.StatusBadRequest)
		return
	}

	if monitoring.GlobalGradeManager == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	grades := monitoring.GlobalGradeManager.GetGradesByRole(roleID)

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(grades)
}

// HandlePerformanceByProject returns performance grades filtered by project
func (s *AgentServer) HandlePerformanceByProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.URL.Query().Get("project")
	if projectID == "" {
		http.Error(w, "Missing project parameter", http.StatusBadRequest)
		return
	}
	// The query parameter may be a raw filesystem path; hash it for consistency
	// with the grade storage key which is always a ProjectIDFromPath hash.
	projectID = monitoring.ProjectIDFromPath(projectID)

	if monitoring.GlobalGradeManager == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	grades := monitoring.GlobalGradeManager.GetGradesByProject(projectID)

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(grades)
}

// calculateCostSavings estimates cost savings from adaptive model selection
func (s *AgentServer) calculateCostSavings(summary monitoring.GradeSummary) map[string]interface{} {
	// Calculate what it would cost if everything used Sonnet (baseline)
	baselineCostPerTask := 0.00045 // Rough estimate: 50k tokens * $9/1M
	totalTasks := 0

	for _, roleSum := range summary.ByRole {
		totalTasks += roleSum.TotalAttempts
	}

	baselineCost := float64(totalTasks) * baselineCostPerTask

	// Calculate actual cost based on model distribution
	actualCost := 0.0
	for modelID, modelSum := range summary.ByModel {
		modelInfo, found := monitoring.GetModelInfo(modelID)
		if !found {
			// Unknown model, use medium tier estimate
			actualCost += float64(modelSum.TotalAttempts) * 0.00030
			continue
		}

		// Rough estimate: avg 25k input + 25k output tokens
		avgCost := (25000.0/1000000.0)*modelInfo.CostPerMIn + (25000.0/1000000.0)*modelInfo.CostPerMOut
		actualCost += float64(modelSum.TotalAttempts) * avgCost
	}

	savings := baselineCost - actualCost
	savingsPercent := 0.0
	if baselineCost > 0 {
		savingsPercent = (savings / baselineCost) * 100
	}

	return map[string]interface{}{
		"baseline_cost":   baselineCost,
		"actual_cost":     actualCost,
		"savings":         savings,
		"savings_percent": savingsPercent,
		"total_tasks":     totalTasks,
		"avg_cost_per_task": func() float64 {
			if totalTasks > 0 {
				return actualCost / float64(totalTasks)
			}
			return 0
		}(),
	}
}

// getModelTierInfo returns information about model tiers
func getModelTierInfo() map[string]interface{} {
	return map[string]interface{}{
		"tiers": []map[string]interface{}{
			{
				"tier":        1,
				"name":        "Minimal",
				"models":      []string{"gpt-4o-mini", "claude-haiku-4-5"},
				"cost_range":  "$0.15-1.25/1M",
				"description": "Cost-effective for simple tasks",
			},
			{
				"tier":        2,
				"name":        "Low",
				"models":      []string{"gpt-4.1-mini"},
				"cost_range":  "$0.40-1.60/1M",
				"description": "Good balance for standard work",
			},
			{
				"tier":        3,
				"name":        "Medium",
				"models":      []string{"claude-sonnet-4-5"},
				"cost_range":  "$3.00-15.00/1M",
				"description": "High capability for complex tasks",
			},
			{
				"tier":        4,
				"name":        "High",
				"models":      []string{"claude-opus-4-6"},
				"cost_range":  "$5.00-25.00/1M",
				"description": "Maximum capability for critical work",
			},
		},
	}
}

// getGradeWeights returns grade threshold information
func getGradeWeights() map[string]interface{} {
	return map[string]interface{}{
		"grades": []map[string]interface{}{
			{
				"grade":            "A",
				"min_success_rate": 0.90,
				"max_retry_rate":   0.05,
				"description":      "Excellent performance",
				"action":           "Consider downgrading to save cost",
				"color":            "#10b981",
			},
			{
				"grade":            "B",
				"min_success_rate": 0.80,
				"max_retry_rate":   0.10,
				"description":      "Good performance",
				"action":           "Maintain current tier",
				"color":            "#3b82f6",
			},
			{
				"grade":            "C",
				"min_success_rate": 0.70,
				"max_retry_rate":   0.20,
				"description":      "Acceptable performance",
				"action":           "Monitor closely",
				"color":            "#f59e0b",
			},
			{
				"grade":            "D",
				"min_success_rate": 0.60,
				"max_retry_rate":   0.30,
				"description":      "Poor performance",
				"action":           "Escalate to next tier",
				"color":            "#ef4444",
			},
			{
				"grade":            "F",
				"min_success_rate": 0.0,
				"max_retry_rate":   1.0,
				"description":      "Failing",
				"action":           "Escalate immediately",
				"color":            "#991b1b",
			},
		},
	}
}

// HandlePerformanceReload forces the grade manager to re-read all grade JSON
// files from disk and replace the in-memory grades map.
//
// POST /api/performance/reload
func (s *AgentServer) HandlePerformanceReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if monitoring.GlobalGradeManager == nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Performance grade manager not initialized",
		})
		return
	}

	if err := monitoring.GlobalGradeManager.ReloadGrades(); err != nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "reloaded",
	})
}

// HandleBenchmarkRun launches scripts/benchmark-models.py in the background.
// POST /api/performance/benchmark/run
// Optional JSON body: {"project": "/path/to/project"} (defaults to first known project root)
func (s *AgentServer) HandleBenchmarkRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Allow optional override from request body
	var body struct {
		Project string `json:"project"`
	}
	projectRoot := ""
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		projectRoot = body.Project
	}

	if projectRoot == "" {
		roots := s.GetProjectRoots()
		if len(roots) > 0 {
			projectRoot = roots[0]
		}
	}

	if projectRoot == "" {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "no project root available"})
		return
	}

	// Locate benchmark script relative to the project root
	scriptPath := filepath.Join(projectRoot, "scripts", "benchmark-models.py")
	if _, err := os.Stat(scriptPath); err != nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("benchmark script not found: %s", scriptPath),
		})
		return
	}

	// Pass the canonical data dir so benchmark grades land in the same store
	// the server reads from, regardless of where either binary is invoked.
	dataDir, err := config.DataDir()
	if err != nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	gradesDir := filepath.Join(dataDir, "performance_grades")
	cmd := exec.Command("python3", scriptPath, "--project", projectRoot, "--grades-dir", gradesDir)
	cmd.Dir = projectRoot
	if err := cmd.Start(); err != nil {
		w.Header().Set("Content-Type", constants.ContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Don't wait – let it run asynchronously
	go func() { _ = cmd.Wait() }()

	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
		"pid":    cmd.Process.Pid,
	})
}

// Register performance API routes
func (s *AgentServer) registerPerformanceRoutes() {
	// Performance grade endpoints
	http.HandleFunc("/api/performance/grades", s.HandlePerformanceGrades)
	http.HandleFunc("/api/performance/summary", s.HandlePerformanceSummary)
	http.HandleFunc("/api/performance/by-role", s.HandlePerformanceByRole)
	http.HandleFunc("/api/performance/by-project", s.HandlePerformanceByProject)
	http.HandleFunc("/api/performance/reload", s.HandlePerformanceReload)
	http.HandleFunc("/api/benchmark/run", s.HandleBenchmarkRun)
}
