import { useGraphQLQuery } from './useGraphQLQuery';

const PERFORMANCE_QUERY = `
  query PerformanceData {
    performanceSummary {
      totalGrades
      gradeDistribution
      byRole
      byModel
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
    }
  }
`;

interface PerformanceData {
  summary: any;
  grades: any[];
  loading: boolean;
  error: string | null;
  refresh: () => void;
}

export const usePerformance = (_apiUrl: string): PerformanceData => {
  const { data, isLoading, error, refetch } = useGraphQLQuery<{
    performanceSummary: any;
    performanceGrades: any[];
  }>('performance', PERFORMANCE_QUERY, undefined, {
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });

  return {
    summary: data?.performanceSummary || null,
    grades: data?.performanceGrades || [],
    loading: isLoading,
    error: error?.message || null,
    refresh: refetch,
  };
};
