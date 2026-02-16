package graphql

// TokenUsage represents token usage statistics
type TokenUsage struct {
	TotalTokens  int `json:"totalTokens"`
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// APICalls represents API call statistics
type APICalls struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// Performance represents system performance metrics
type Performance struct {
	CPUUsage float64 `json:"cpuUsage"`
	Uptime   string  `json:"uptime"`
}

// Metrics represents aggregated system metrics
type Metrics struct {
	TasksSpawned         int               `json:"tasksSpawned"`
	TasksCompleted       int               `json:"tasksCompleted"`
	TasksFailed          int               `json:"tasksFailed"`
	TasksActive          int               `json:"tasksActive"`
	AverageDurationMs    float64           `json:"averageDurationMs"`
	AverageTokensPerTask int               `json:"averageTokensPerTask"`
	TokenUsage           *TokenUsage       `json:"tokenUsage"`
	APICalls             *APICalls         `json:"apiCalls"`
	Performance          *Performance      `json:"performance"`
	TurnMetrics          *TurnMetrics      `json:"turnMetrics"`
	SessionMetrics       *SessionMetrics   `json:"sessionMetrics"`
	Streaming            *StreamingMetrics `json:"streaming"`
	HTTP                 *HTTPMetrics      `json:"http"`
	RateLimiting         *RateLimiting     `json:"rateLimiting"`
	ProviderBreakdown    []ProviderUsage   `json:"providerBreakdown"`
}

// TurnData represents token usage for a single turn
type TurnData struct {
	TaskID       string `json:"taskID"`
	Turn         int    `json:"turn"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
	DurationMs   int    `json:"durationMs"`
}

// TurnMetrics represents per-turn token statistics
type TurnMetrics struct {
	TotalTurns      int        `json:"totalTurns"`
	AvgInputPerTurn int        `json:"avgInputPerTurn"`
	AvgOutputPerTurn int        `json:"avgOutputPerTurn"`
	RecentTurns     []TurnData `json:"recentTurns"`
}

// SessionData represents token usage for a single session/task
type SessionData struct {
	TaskID       string `json:"taskID"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
	TurnCount    int    `json:"turnCount"`
}

// SessionMetrics represents session/task token statistics
type SessionMetrics struct {
	RecentSessions []SessionData `json:"recentSessions"`
}

// StreamingMetrics represents streaming connection statistics
type StreamingMetrics struct {
	Opened int `json:"opened"`
	Closed int `json:"closed"`
	Active int `json:"active"`
}

// HTTPMetrics represents HTTP request statistics
type HTTPMetrics struct {
	TotalRequests int `json:"totalRequests"`
	Errors        int `json:"errors"`
}

// RateLimiting represents rate limiting statistics
type RateLimiting struct {
	Violations int `json:"violations"`
}

// ProviderUsage represents token usage for a specific provider/model
type ProviderUsage struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	Calls        int    `json:"calls"`
	InputTokens  int    `json:"inputTokens"`
	OutputTokens int    `json:"outputTokens"`
}

// Package graphql provides GraphQL API types and resolvers
