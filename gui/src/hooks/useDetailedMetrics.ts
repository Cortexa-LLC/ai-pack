import { useQuery } from '@tanstack/react-query';

/**
 * Detailed metrics data from REST endpoint
 */
export interface DetailedMetrics {
  tasks_spawned: number;
  tasks_completed: number;
  tasks_failed: number;
  tasks_in_progress: number;
  total_duration_ms: number;
  avg_duration_ms: number;
  api_calls_total: number;
  api_calls_success: number;
  api_calls_failed: number;
  http_requests_total: number;
  http_errors: number;
  streams_opened: number;
  streams_closed: number;
  streams_active: number;
  rate_limit_violations: number;
  total_input_tokens: number;
  total_output_tokens: number;
  task_token_usage?: TaskTokenUsage[];
  total_turns: number;
  avg_input_per_turn: number;
  avg_output_per_turn: number;
  average_tokens_per_task: number;
  cpu_usage: number;
  uptime: number;
  turn_token_data?: TurnTokenData[];
  timestamp: string;
}

export interface TurnTokenData {
  TaskID: string;
  Turn: number;
  InputTokens: number;
  OutputTokens: number;
  DurationMs: number;
}

export interface TaskTokenUsage {
  TaskID: string;
  InputTokens: number;
  OutputTokens: number;
  TurnCount: number;
}

/**
 * Hook to fetch detailed metrics from REST endpoint
 */
export function useDetailedMetrics() {
  return useQuery<DetailedMetrics>({
    queryKey: ['detailedMetrics'],
    queryFn: async () => {
      const response = await fetch('/metrics');
      if (!response.ok) {
        throw new Error('Failed to fetch detailed metrics');
      }
      return response.json();
    },
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}
