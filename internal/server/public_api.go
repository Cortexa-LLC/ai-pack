package server

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
	"github.com/cortexa-llc/ai-pack/internal/protocol"
	"github.com/cortexa-llc/ai-pack/internal/constants"
)

// Public API methods for the server

// SpawnAgentTask spawns an agent task (public method for legacy endpoint)
func (s *AgentServer) SpawnAgentTask(role, task string) (*protocol.ExecuteTaskResponse, error) {
	// Legacy method - use server's root directory as project root
	return s.spawnAgentTask(role, task, "")
}

// ExecuteTaskSync executes a task synchronously and waits for completion
// Used for protocol handler mode (agent:// URLs)
func (s *AgentServer) ExecuteTaskSync(role, task string) (*protocol.ExecuteTaskResponse, error) {
	// Spawn the task - use server's root directory as project root
	response, err := s.spawnAgentTask(role, task, "")
	if err != nil {
		return nil, err
	}

	taskID := response.TaskID

	// Wait for completion by polling status
	for {
		time.Sleep(1 * time.Second)

		status, err := s.getTaskStatus(taskID)
		if err != nil {
			return nil, err
		}

		if status.Status == constants.StatusCompleted {
			response.Status = constants.StatusCompleted
			response.Message = "Task completed successfully"
			return response, nil
		}

		if status.Status == constants.StatusFailed {
			response.Status = constants.StatusFailed
			response.Message = status.Error
			return response, nil
		}

		// Still in progress, continue polling
	}
}

// HandleAgentCard is the public handler for GET /.well-known/agent.json
func (s *AgentServer) HandleAgentCard(w http.ResponseWriter, r *http.Request) {
	s.handleAgentCard(w, r)
}

// HandleA2ADiscovery is the public handler for A2A discovery
func (s *AgentServer) HandleA2ADiscovery(w http.ResponseWriter, r *http.Request) {
	s.handleA2ADiscovery(w, r)
}

// HandleA2AExecute is the public handler for A2A execute
func (s *AgentServer) HandleA2AExecute(w http.ResponseWriter, r *http.Request) {
	s.handleA2AExecute(w, r)
}

// HandleA2AStatus is the public handler for A2A status
func (s *AgentServer) HandleA2AStatus(w http.ResponseWriter, r *http.Request) {
	s.handleA2AStatus(w, r)
}

// HandleStream is the public handler for SSE streaming
func (s *AgentServer) HandleStream(w http.ResponseWriter, r *http.Request) {
	s.handleStream(w, r)
}

// GetMetricsSnapshot returns a snapshot of current server metrics
func (s *AgentServer) GetMetricsSnapshot() monitoring.MetricsSnapshot {
	return monitoring.GlobalMetrics.GetSnapshot()
}

// GetDailyUsage returns today's token usage (aggregated across all projects)
func (s *AgentServer) GetDailyUsage() (*monitoring.DailyUsage, error) {
	s.mu.RLock()
	projectRoots := make([]string, 0, len(s.projectRoots))
	for pr := range s.projectRoots {
		projectRoots = append(projectRoots, pr)
	}
	s.mu.RUnlock()

	// Aggregate across all projects
	aggregated := &monitoring.DailyUsage{
		Date:              time.Now().Format("2006-01-02"),
		ProviderBreakdown: make(map[string]*monitoring.ProviderDailyUsage),
		LastUpdated:       time.Now(),
	}

	for _, projectRoot := range projectRoots {
		pm, err := s.getOrCreateProjectMetrics(projectRoot)
		if err != nil {
			continue // Skip projects with errors
		}

		daily, err := pm.GetToday()
		if err != nil {
			continue // Skip if no data
		}

		// Aggregate totals
		aggregated.TotalInputTokens += daily.TotalInputTokens
		aggregated.TotalOutputTokens += daily.TotalOutputTokens

		// Merge provider breakdown
		for key, usage := range daily.ProviderBreakdown {
			if existing, ok := aggregated.ProviderBreakdown[key]; ok {
				// Add to existing
				existing.Calls += usage.Calls
				existing.InputTokens += usage.InputTokens
				existing.OutputTokens += usage.OutputTokens
				existing.Cost += usage.Cost
			} else {
				// Create new entry (copy to avoid reference issues)
				aggregated.ProviderBreakdown[key] = &monitoring.ProviderDailyUsage{
					Provider:     usage.Provider,
					Model:        usage.Model,
					Calls:        usage.Calls,
					InputTokens:  usage.InputTokens,
					OutputTokens: usage.OutputTokens,
					Cost:         usage.Cost,
				}
			}
		}
	}

	return aggregated, nil
}

// GetDailyUsageRange returns token usage for a date range (aggregated across all projects)
func (s *AgentServer) GetDailyUsageRange(startDate, endDate string) ([]*monitoring.DailyUsage, error) {
	s.mu.RLock()
	projectRoots := make([]string, 0, len(s.projectRoots))
	for pr := range s.projectRoots {
		projectRoots = append(projectRoots, pr)
	}
	s.mu.RUnlock()

	// Collect all daily usage by date
	byDate := make(map[string]*monitoring.DailyUsage)

	for _, projectRoot := range projectRoots {
		pm, err := s.getOrCreateProjectMetrics(projectRoot)
		if err != nil {
			continue
		}

		dailies, err := pm.GetDateRange(startDate, endDate)
		if err != nil {
			continue
		}

		// Merge into byDate map
		for _, daily := range dailies {
			if existing, ok := byDate[daily.Date]; ok {
				// Aggregate
				existing.TotalInputTokens += daily.TotalInputTokens
				existing.TotalOutputTokens += daily.TotalOutputTokens

				// Merge provider breakdown
				for key, usage := range daily.ProviderBreakdown {
					if existingUsage, ok := existing.ProviderBreakdown[key]; ok {
						existingUsage.Calls += usage.Calls
						existingUsage.InputTokens += usage.InputTokens
						existingUsage.OutputTokens += usage.OutputTokens
						existingUsage.Cost += usage.Cost
					} else {
						existing.ProviderBreakdown[key] = &monitoring.ProviderDailyUsage{
							Provider:     usage.Provider,
							Model:        usage.Model,
							Calls:        usage.Calls,
							InputTokens:  usage.InputTokens,
							OutputTokens: usage.OutputTokens,
							Cost:         usage.Cost,
						}
					}
				}

				if daily.LastUpdated.After(existing.LastUpdated) {
					existing.LastUpdated = daily.LastUpdated
				}
			} else {
				// Create new entry
				newDaily := &monitoring.DailyUsage{
					Date:              daily.Date,
					TotalInputTokens:  daily.TotalInputTokens,
					TotalOutputTokens: daily.TotalOutputTokens,
					ProviderBreakdown: make(map[string]*monitoring.ProviderDailyUsage),
					LastUpdated:       daily.LastUpdated,
				}
				for key, usage := range daily.ProviderBreakdown {
					newDaily.ProviderBreakdown[key] = &monitoring.ProviderDailyUsage{
						Provider:     usage.Provider,
						Model:        usage.Model,
						Calls:        usage.Calls,
						InputTokens:  usage.InputTokens,
						OutputTokens: usage.OutputTokens,
						Cost:         usage.Cost,
					}
				}
				byDate[daily.Date] = newDaily
			}
		}
	}

	// Convert map to sorted slice
	result := make([]*monitoring.DailyUsage, 0, len(byDate))
	for _, daily := range byDate {
		result = append(result, daily)
	}

	// Sort by date
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	return result, nil
}

// GetLast30DaysUsage returns token usage for the last 30 days (aggregated across all projects)
func (s *AgentServer) GetLast30DaysUsage() ([]*monitoring.DailyUsage, error) {
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	return s.GetDailyUsageRange(startDate, endDate)
}

// GetMaxConcurrent returns the maximum concurrent agents setting
func (s *AgentServer) GetMaxConcurrent() int {
	return s.maxConcurrent
}

// LoggingMiddleware creates HTTP middleware for request logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := context.Background()

		monitoring.GlobalMetrics.IncrementHTTPRequests()

		// Wrap response writer to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(lrw, r)

		// Log request
		durationMs := time.Since(start).Milliseconds()
		monitoring.LogHTTPRequest(ctx, r.Method, r.URL.Path, lrw.statusCode, durationMs)

		if lrw.statusCode >= 400 {
			monitoring.GlobalMetrics.IncrementHTTPErrors()
		}
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher for SSE streaming support
func (lrw *loggingResponseWriter) Flush() {
	if flusher, ok := lrw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
