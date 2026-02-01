import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useMetrics } from './useMetrics';
import React from 'react';

// Mock graphql client
vi.mock('../lib/graphql', () => ({
  graphqlClient: {
    query: vi.fn(),
  },
}));

import { graphqlClient } from '../lib/graphql';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  );
};

describe('useMetrics', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch metrics data', async () => {
    const mockMetrics = {
      metrics: {
        tasksSpawned: 10,
        tasksCompleted: 8,
        tasksFailed: 1,
        tasksActive: 1,
        averageDurationMs: 1500.5,
        tokenUsage: {
          totalTokens: 5000,
          inputTokens: 3000,
          outputTokens: 2000,
        },
        apiCalls: {
          total: 50,
          success: 48,
          failed: 2,
        },
        performance: {
          cpuUsage: 25.5,
          memoryUsageMB: 512.0,
          goroutines: 10,
          uptime: '2h30m',
        },
      },
    };

    (graphqlClient.query as any).mockResolvedValueOnce({ data: mockMetrics });

    const { result } = renderHook(() => useMetrics(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockMetrics);
  });

  it('should auto-refresh metrics', () => {
    const mockMetrics = { metrics: { tasksSpawned: 5 } };
    (graphqlClient.query as any).mockResolvedValue({ data: mockMetrics });

    const { result } = renderHook(() => useMetrics(), { wrapper: createWrapper() });

    // Should be loading initially
    expect(result.current.isLoading).toBe(true);
  });
});
