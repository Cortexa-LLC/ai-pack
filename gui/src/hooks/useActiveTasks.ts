import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function useActiveTasks() {
  return useQuery({
    queryKey: ['active-tasks'],
    queryFn: api.getTasks,
    refetchInterval: 2000, // Poll every 2 seconds
  });
}
