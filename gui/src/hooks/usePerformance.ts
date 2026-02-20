import { useCallback } from 'react';
import { useGraphQLQuery } from './useGraphQLQuery';

const PERFORMANCE_QUERY = `
  query PerformanceData {
    performanceSummary {
      totalGrades
      gradeDistribution
      byRole
      byModel
      modelTiers
      costSavings {
        baselineCost
        actualCost
        savings
        savingsPercent
        totalTasks
        avgCostPerTask
      }
    }
    performanceGrades {
      modelID
      roleID
      projectID
      totalAttempts
      successes
      failures
      retries
      successRate
      retryRate
      grade
      confidenceScore
      averageTokens
      averageExecutionTime
      escalationCount
      downgradeCount
      lastUsed
      firstUsed
      source
    }
  }
`;

interface PerformanceData {
  summary: any;
  grades: any[];
  loading: boolean;
  error: string | null;
  /** Re-fetches already-loaded data from the GraphQL cache / server. */
  refresh: () => void;
  /**
   * Asks the agent-server to reload grade JSON files from disk, then
   * immediately re-fetches the GraphQL data so the UI reflects the new state.
   */
  reload: () => Promise<void>;
}

export const usePerformance = (_apiUrl: string): PerformanceData => {
  const { data, isLoading, error, refetch } = useGraphQLQuery<{
    performanceSummary: any;
    performanceGrades: any[];
  }>('performance', PERFORMANCE_QUERY, undefined, {
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });

  const reload = useCallback(async () => {
    await fetch('/api/performance/reload', { method: 'POST' });
    refetch();
  }, [refetch]);

  return {
    summary: data?.performanceSummary || null,
    grades: data?.performanceGrades || [],
    loading: isLoading,
    error: error?.message || null,
    refresh: refetch,
    reload,
  };
};
