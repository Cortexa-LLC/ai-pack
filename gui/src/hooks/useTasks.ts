import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: api.getTasks,
    refetchInterval: 3000, // Poll every 3 seconds
  });
}
