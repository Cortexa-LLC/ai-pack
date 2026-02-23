import { useGraphQLQuery } from './useGraphQLQuery';
import type { ProviderUsage } from '../types/graphql-types';

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
        uptime
      }
      providerBreakdown {
        provider
        model
        calls
        inputTokens
        outputTokens
      }
    }
  }
`;

// ProviderUsage is imported from ../types/graphql-types
export type { ProviderUsage };

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
      uptime: string;
    };
    providerBreakdown: ProviderUsage[];
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
