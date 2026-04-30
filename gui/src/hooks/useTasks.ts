import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export type { AgentTask } from '../types';

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: api.getTasks,
    refetchInterval: 3000, // Poll every 3 seconds
  });
}
