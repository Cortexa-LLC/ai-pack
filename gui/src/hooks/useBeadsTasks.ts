import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function useBeadsTasks() {
  return useQuery({
    queryKey: ['beads-tasks'],
    queryFn: api.getBeadsTasks,
    refetchInterval: 3000, // Poll every 3 seconds
  });
}
