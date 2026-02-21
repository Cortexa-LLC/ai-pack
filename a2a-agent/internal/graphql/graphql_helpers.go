package graphql

import (
	"time"

	"github.com/cortexa-llc/ai-pack/a2a-agent/internal/monitoring"
)

func convertTaskInfoToAgentTask(taskInfo *TaskInfo) *AgentTask {
	// Convert metadata from map[string]string to map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range taskInfo.Metadata {
		metadata[k] = v
	}

	return &AgentTask{
		TaskID:      taskInfo.TaskID,
		Role:        taskInfo.Role,
		Task:        taskInfo.Task,
		Status:      taskInfo.Status,
		CreatedAt:   taskInfo.CreatedAt,
		UpdatedAt:   taskInfo.UpdatedAt,
		CompletedAt: taskInfo.CompletedAt,
		Result:      taskInfo.Result,
		Error:       taskInfo.Error,
		Metadata:    metadata,
		ProjectRoot: taskInfo.ProjectRoot,
	}
}

func convertMonitoringGrade(mg *monitoring.PerformanceGrade) *PerformanceGrade {
	return &PerformanceGrade{
		ModelID:              mg.ModelID,
		RoleID:               mg.RoleID,
		ProjectID:            mg.ProjectID,
		TotalAttempts:        mg.TotalAttempts,
		Successes:            mg.Successes,
		Failures:             mg.Failures,
		Retries:              mg.Retries,
		SuccessRate:          mg.SuccessRate,
		RetryRate:            mg.RetryRate,
		Grade:                mg.Grade,
		ConfidenceScore:      mg.ConfidenceScore,
		AverageTokens:        mg.AverageTokens,
		AverageExecutionTime: mg.AverageExecutionTime,
		EscalationCount:      mg.EscalationCount,
		DowngradeCount:       mg.DowngradeCount,
		LastUsed:             mg.LastUsed.Format(time.RFC3339),
		FirstUsed:            mg.FirstUsed.Format(time.RFC3339),
		Source:               mg.Source,
	}
}

func calculateCostSavings() *CostSavings {
	if monitoring.GlobalMetrics == nil {
		return &CostSavings{
			BaselineCost:   0,
			ActualCost:     0,
			Savings:        0,
			SavingsPercent: 0,
			TotalTasks:     0,
			AvgCostPerTask: 0,
		}
	}

	metrics := monitoring.GlobalMetrics.GetSnapshot()

	// Estimate costs from token usage
	// Pricing estimates (per 1M tokens):
	// Sonnet: $3 input, $15 output
	// Haiku: $0.25 input, $1.25 output
	inputTokens := float64(metrics.TotalInputTokens)
	outputTokens := float64(metrics.TotalOutputTokens)

	avgInputCostPer1M := 1.625  // ($3 + $0.25) / 2
	avgOutputCostPer1M := 8.125 // ($15 + $1.25) / 2
	actualCost := (inputTokens/1000000)*avgInputCostPer1M + (outputTokens/1000000)*avgOutputCostPer1M

	// Baseline assumes all tasks used Sonnet (most expensive)
	sonnetInputCostPer1M := 3.0
	sonnetOutputCostPer1M := 15.0
	baselineCost := (inputTokens/1000000)*sonnetInputCostPer1M + (outputTokens/1000000)*sonnetOutputCostPer1M

	savings := baselineCost - actualCost
	savingsPercent := 0.0
	if baselineCost > 0 {
		savingsPercent = (savings / baselineCost) * 100
	}

	totalTasks := float64(metrics.TasksCompleted)
	avgCostPerTask := 0.0
	if totalTasks > 0 {
		avgCostPerTask = actualCost / totalTasks
	}

	return &CostSavings{
		BaselineCost:   baselineCost,
		ActualCost:     actualCost,
		Savings:        savings,
		SavingsPercent: savingsPercent,
		TotalTasks:     int(totalTasks),
		AvgCostPerTask: avgCostPerTask,
	}
}
