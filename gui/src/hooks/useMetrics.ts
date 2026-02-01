import { useGraphQLQuery } from './useGraphQLQuery';

/**
 * GraphQL query for fetching system metrics
 */
const METRICS_QUERY = `
  query GetMetrics {
    metrics {
      tasksSpawned
      tasksCompleted
      tasksFailed
      tasksActive
      averageDurationMs
      averageTokensPerTask
      tokenUsage {
        totalTokens
        inputTokens
        outputTokens
      }
      apiCalls {
        total
        success
        failed
      }
      performance {
        cpuUsage
        uptime
      }
    }
  }
`;

/**
 * Metrics data type
 */
export interface Metrics {
  metrics: {
    tasksSpawned: number;
    tasksCompleted: number;
    tasksFailed: number;
    tasksActive: number;
    averageDurationMs: number;
    averageTokensPerTask: number;
    tokenUsage: {
      totalTokens: number;
      inputTokens: number;
      outputTokens: number;
    };
    apiCalls: {
      total: number;
      success: number;
      failed: number;
    };
    performance: {
      cpuUsage: number;
      memoryUsageMB: number;
      goroutines: number;
      uptime: string;
    };
  };
}

/**
 * Hook to fetch system metrics with auto-refresh
 * @returns React Query result with metrics data
 */
export function useMetrics() {
  return useGraphQLQuery<Metrics>('metrics', METRICS_QUERY, undefined, {
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}
