export interface AgentTask {
  taskId: string;
  taskID: string; // Alias for taskId (GraphQL compatibility)
  beadsTaskId?: string; // Short task ID extracted from full taskId
  status: 'in_progress' | 'completed' | 'failed' | 'queued' | 'IN_PROGRESS' | 'COMPLETED' | 'FAILED' | 'QUEUED' | 'OPEN' | 'BLOCKED' | 'open' | 'blocked';
  role: string;
  description: string;
  task: string; // Alias for description (GraphQL compatibility)
  projectRoot: string;
  error?: string;
  result?: string;
  currentTurn?: number;
  tokensUsed?: number;
  metadata?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

export interface Metrics {
  tasks_spawned: number;
  tasks_in_progress: number;
  tasks_completed: number;
  tasks_failed: number;
  avg_duration_ms: number;
  api_calls_total: number;
  api_calls_success: number;
  api_calls_failed: number;
  rate_limit_violations: number;
  total_input_tokens: number;
  total_output_tokens: number;
  avg_input_per_turn: number;
  avg_output_per_turn: number;
  total_turns: number;
  turn_token_data?: TurnData[];
  task_token_usage?: TaskTokenUsage[];
}

export interface TurnData {
  Turn: number;
  DurationMs: number;
  InputTokens: number;
  OutputTokens: number;
  TaskID: string;
}

export interface TaskTokenUsage {
  TaskID: string;
  InputTokens: number;
  OutputTokens: number;
  TurnCount: number;
}

export interface LogEntry {
  timestamp: string;
  type: 'turn' | 'api' | 'text' | 'tool' | 'success' | 'error' | 'warning' | 'info';
  message: string;
  data?: any;
}

export interface ServerHealth {
  status: 'healthy' | 'unhealthy';
  version: string;
  server: string;
  features: {
    a2a_protocol: boolean;
    monitoring: boolean;
    parallel_execution: boolean;
    sse_streaming: boolean;
  };
}
