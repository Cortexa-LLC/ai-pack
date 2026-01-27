package monitoring

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects server performance metrics
type Metrics struct {
	mu sync.RWMutex

	// Task metrics
	TasksSpawned    int64
	TasksCompleted  int64
	TasksFailed     int64
	TasksInProgress int64

	// Performance metrics
	TotalDurationMs int64
	AvgDurationMs   int64

	// API metrics
	APICallsTotal   int64
	APICallsSuccess int64
	APICallsFailed  int64

	// HTTP metrics
	HTTPRequestsTotal int64
	HTTPErrors        int64

	// Streaming metrics
	StreamsOpened int64
	StreamsClosed int64
	StreamsActive int64

	// Rate limiting
	RateLimitViolations int64

	// Token usage tracking
	TotalInputTokens  int64
	TotalOutputTokens int64

	// Detailed task duration tracking
	taskDurations []int64
	maxDurations  int

	// Per-task token tracking (session metrics)
	taskTokenUsage []TaskTokenUsage
	maxTokenUsage  int
}

// TaskTokenUsage tracks token usage for a single task/session
type TaskTokenUsage struct {
	TaskID       string
	InputTokens  int64
	OutputTokens int64
	TurnCount    int64
}

// Global metrics instance
var GlobalMetrics *Metrics

// InitMetrics initializes the global metrics collector
func InitMetrics() {
	GlobalMetrics = &Metrics{
		maxDurations:   1000, // Keep last 1000 task durations for stats
		taskDurations:  make([]int64, 0, 1000),
		maxTokenUsage:  100, // Keep last 100 task token usage records
		taskTokenUsage: make([]TaskTokenUsage, 0, 100),
	}
}

// IncrementTasksSpawned increments spawned tasks counter
func (m *Metrics) IncrementTasksSpawned() {
	atomic.AddInt64(&m.TasksSpawned, 1)
	atomic.AddInt64(&m.TasksInProgress, 1)
}

// IncrementTasksCompleted increments completed tasks counter
func (m *Metrics) IncrementTasksCompleted(durationMs int64) {
	atomic.AddInt64(&m.TasksCompleted, 1)
	atomic.AddInt64(&m.TasksInProgress, -1)
	atomic.AddInt64(&m.TotalDurationMs, durationMs)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to duration tracking
	if len(m.taskDurations) < m.maxDurations {
		m.taskDurations = append(m.taskDurations, durationMs)
	} else {
		// Shift and add (FIFO)
		m.taskDurations = append(m.taskDurations[1:], durationMs)
	}

	// Update average
	if m.TasksCompleted > 0 {
		m.AvgDurationMs = m.TotalDurationMs / m.TasksCompleted
	}
}

// IncrementTasksFailed increments failed tasks counter
func (m *Metrics) IncrementTasksFailed(durationMs int64) {
	atomic.AddInt64(&m.TasksFailed, 1)
	atomic.AddInt64(&m.TasksInProgress, -1)
	atomic.AddInt64(&m.TotalDurationMs, durationMs)
}

// IncrementAPICallsSuccess increments successful API calls
func (m *Metrics) IncrementAPICallsSuccess() {
	atomic.AddInt64(&m.APICallsTotal, 1)
	atomic.AddInt64(&m.APICallsSuccess, 1)
}

// IncrementAPICallsFailed increments failed API calls
func (m *Metrics) IncrementAPICallsFailed() {
	atomic.AddInt64(&m.APICallsTotal, 1)
	atomic.AddInt64(&m.APICallsFailed, 1)
}

// IncrementHTTPRequests increments HTTP request counter
func (m *Metrics) IncrementHTTPRequests() {
	atomic.AddInt64(&m.HTTPRequestsTotal, 1)
}

// IncrementHTTPErrors increments HTTP error counter
func (m *Metrics) IncrementHTTPErrors() {
	atomic.AddInt64(&m.HTTPErrors, 1)
}

// IncrementStreamsOpened increments opened streams counter
func (m *Metrics) IncrementStreamsOpened() {
	atomic.AddInt64(&m.StreamsOpened, 1)
	atomic.AddInt64(&m.StreamsActive, 1)
}

// IncrementStreamsClosed increments closed streams counter
func (m *Metrics) IncrementStreamsClosed() {
	atomic.AddInt64(&m.StreamsClosed, 1)
	atomic.AddInt64(&m.StreamsActive, -1)
}

// IncrementRateLimitViolations increments rate limit violations
func (m *Metrics) IncrementRateLimitViolations() {
	atomic.AddInt64(&m.RateLimitViolations, 1)
}

// RecordTokenUsage records token usage for a completed task/session
func (m *Metrics) RecordTokenUsage(taskID string, inputTokens, outputTokens int64, turnCount int64) {
	// Update global totals
	atomic.AddInt64(&m.TotalInputTokens, inputTokens)
	atomic.AddInt64(&m.TotalOutputTokens, outputTokens)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to per-task tracking
	usage := TaskTokenUsage{
		TaskID:       taskID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TurnCount:    turnCount,
	}

	if len(m.taskTokenUsage) < m.maxTokenUsage {
		m.taskTokenUsage = append(m.taskTokenUsage, usage)
	} else {
		// Shift and add (FIFO)
		m.taskTokenUsage = append(m.taskTokenUsage[1:], usage)
	}
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy task token usage for snapshot
	tokenUsageCopy := make([]TaskTokenUsage, len(m.taskTokenUsage))
	copy(tokenUsageCopy, m.taskTokenUsage)

	return MetricsSnapshot{
		TasksSpawned:        atomic.LoadInt64(&m.TasksSpawned),
		TasksCompleted:      atomic.LoadInt64(&m.TasksCompleted),
		TasksFailed:         atomic.LoadInt64(&m.TasksFailed),
		TasksInProgress:     atomic.LoadInt64(&m.TasksInProgress),
		TotalDurationMs:     atomic.LoadInt64(&m.TotalDurationMs),
		AvgDurationMs:       atomic.LoadInt64(&m.AvgDurationMs),
		APICallsTotal:       atomic.LoadInt64(&m.APICallsTotal),
		APICallsSuccess:     atomic.LoadInt64(&m.APICallsSuccess),
		APICallsFailed:      atomic.LoadInt64(&m.APICallsFailed),
		HTTPRequestsTotal:   atomic.LoadInt64(&m.HTTPRequestsTotal),
		HTTPErrors:          atomic.LoadInt64(&m.HTTPErrors),
		StreamsOpened:       atomic.LoadInt64(&m.StreamsOpened),
		StreamsClosed:       atomic.LoadInt64(&m.StreamsClosed),
		StreamsActive:       atomic.LoadInt64(&m.StreamsActive),
		RateLimitViolations: atomic.LoadInt64(&m.RateLimitViolations),
		TotalInputTokens:    atomic.LoadInt64(&m.TotalInputTokens),
		TotalOutputTokens:   atomic.LoadInt64(&m.TotalOutputTokens),
		TaskTokenUsage:      tokenUsageCopy,
		Timestamp:           time.Now(),
	}
}

// MetricsSnapshot is a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	TasksSpawned        int64            `json:"tasks_spawned"`
	TasksCompleted      int64            `json:"tasks_completed"`
	TasksFailed         int64            `json:"tasks_failed"`
	TasksInProgress     int64            `json:"tasks_in_progress"`
	TotalDurationMs     int64            `json:"total_duration_ms"`
	AvgDurationMs       int64            `json:"avg_duration_ms"`
	APICallsTotal       int64            `json:"api_calls_total"`
	APICallsSuccess     int64            `json:"api_calls_success"`
	APICallsFailed      int64            `json:"api_calls_failed"`
	HTTPRequestsTotal   int64            `json:"http_requests_total"`
	HTTPErrors          int64            `json:"http_errors"`
	StreamsOpened       int64            `json:"streams_opened"`
	StreamsClosed       int64            `json:"streams_closed"`
	StreamsActive       int64            `json:"streams_active"`
	RateLimitViolations int64            `json:"rate_limit_violations"`
	TotalInputTokens    int64            `json:"total_input_tokens"`
	TotalOutputTokens   int64            `json:"total_output_tokens"`
	TaskTokenUsage      []TaskTokenUsage `json:"task_token_usage,omitempty"`
	Timestamp           time.Time        `json:"timestamp"`
}

// SuccessRate returns the task success rate as a percentage
func (s *MetricsSnapshot) SuccessRate() float64 {
	total := s.TasksCompleted + s.TasksFailed
	if total == 0 {
		return 0.0
	}
	return float64(s.TasksCompleted) / float64(total) * 100.0
}

// APISuccessRate returns the API call success rate as a percentage
func (s *MetricsSnapshot) APISuccessRate() float64 {
	if s.APICallsTotal == 0 {
		return 0.0
	}
	return float64(s.APICallsSuccess) / float64(s.APICallsTotal) * 100.0
}

// AvgInputTokensPerTask returns average input tokens per completed task
func (s *MetricsSnapshot) AvgInputTokensPerTask() int64 {
	if s.TasksCompleted == 0 {
		return 0
	}
	return s.TotalInputTokens / s.TasksCompleted
}

// AvgOutputTokensPerTask returns average output tokens per completed task
func (s *MetricsSnapshot) AvgOutputTokensPerTask() int64 {
	if s.TasksCompleted == 0 {
		return 0
	}
	return s.TotalOutputTokens / s.TasksCompleted
}

// InputOutputRatio returns the ratio of input to output tokens
func (s *MetricsSnapshot) InputOutputRatio() float64 {
	if s.TotalOutputTokens == 0 {
		return 0.0
	}
	return float64(s.TotalInputTokens) / float64(s.TotalOutputTokens)
}
