// Auto-generated TypeScript interfaces mirroring a2a-agent/internal/graphql/schema.graphql

export interface HealthStatus {
  status: string
  version: string
  server: string
  features: Record<string, unknown>
}

export interface AgentTask {
  taskID: string
  role: string
  task: string
  status: string
  createdAt: string
  updatedAt: string
  completedAt?: string
  result?: string
  error?: string
  metadata?: Record<string, unknown>
  projectRoot?: string
}

export interface BeadsTask {
  id: string
  title: string
  description: string
  status: string
  dependencies?: string[]
  metadata?: Record<string, unknown>
}

export interface TokenUsage {
  totalTokens: number
  inputTokens: number
  outputTokens: number
}

export interface APICalls {
  total: number
  success: number
  failed: number
}

export interface TurnData {
  taskID: string
  turn: number
  inputTokens: number
  outputTokens: number
  durationMs: number
}

export interface TurnMetrics {
  totalTurns: number
  avgInputPerTurn: number
  avgOutputPerTurn: number
  recentTurns: TurnData[]
}

export interface SessionData {
  taskID: string
  inputTokens: number
  outputTokens: number
  turnCount: number
}

export interface SessionMetrics {
  recentSessions: SessionData[]
}

export interface StreamingMetrics {
  opened: number
  closed: number
  active: number
}

export interface HTTPMetrics {
  totalRequests: number
  errors: number
}

export interface RateLimiting {
  violations: number
}

export interface Performance {
  uptime: string
}

export interface ProviderUsage {
  provider: string
  model: string
  calls: number
  inputTokens: number
  outputTokens: number
}

export interface Metrics {
  tasksSpawned: number
  tasksCompleted: number
  tasksFailed: number
  tasksActive: number
  averageDurationMs: number
  averageTokensPerTask: number
  tokenUsage: TokenUsage
  apiCalls: APICalls
  performance: Performance
  turnMetrics: TurnMetrics
  sessionMetrics: SessionMetrics
  streaming: StreamingMetrics
  http: HTTPMetrics
  rateLimiting: RateLimiting
  providerBreakdown: ProviderUsage[]
}

export interface PerformanceGrade {
  modelID: string
  roleID: string
  projectID: string
  totalAttempts: number
  successes: number
  failures: number
  retries: number
  successRate: number
  retryRate: number
  grade: string
  confidenceScore: number
  averageTokens: number
  averageExecutionTime: number
  escalationCount: number
  downgradeCount: number
  lastUsed: string
  firstUsed: string
  source: string
}

export interface CostSavings {
  baselineCost: number
  actualCost: number
  savings: number
  savingsPercent: number
  totalTasks: number
  avgCostPerTask: number
}

export interface GradeSummary {
  totalGrades: number
  gradeDistribution: Record<string, number>
  byRole: Record<string, unknown>
  byModel: Record<string, unknown>
  modelTiers: Record<string, unknown>
  costSavings: CostSavings
}

export interface ProviderCost {
  provider: string
  model: string
  calls: number
  inputTokens: number
  outputTokens: number
  cost: number
}

export interface ProjectCost {
  projectRoot: string
  projectName: string
  totalCost: number
  totalInputTokens: number
  totalOutputTokens: number
  providerBreakdown: ProviderCost[]
}

export interface LogEntry {
  timestamp: string
  level: string
  message: string
  attributes?: Record<string, unknown>
}

export interface ExecutionEvent {
  eventType: string
  taskID: string
  role: string
  task: string
  timestamp: string
  status?: string
  error?: string
  durationMs?: number
  result?: string
  metadata?: Record<string, unknown>
}

export interface TaskSummary {
  taskID: string
  role: string
  task: string
  status: string
  created: string
  updated: string
  durationMs: number
  error?: string
  result?: string
  events: string[]
}

export interface RetryResult {
  success: boolean
  taskID: string
  message?: string
}

export interface CloseResult {
  success: boolean
  taskID: string
  message?: string
}

export interface DeleteResult {
  success: boolean
  taskID: string
  message?: string
}
