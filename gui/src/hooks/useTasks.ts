import { useGraphQLQuery } from './useGraphQLQuery';

/**
 * GraphQL query for fetching agent tasks
 */
const TASKS_QUERY = `
  query GetTasks {
    tasks {
      taskID
      role
      task
      status
      createdAt
      updatedAt
      completedAt
      result
      error
      metadata
      beadsTaskID
    }
  }
`;

/**
 * Agent task data type
 */
export interface AgentTask {
  taskID: string;
  role: string;
  task: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  result: string | null;
  error: string | null;
  metadata: Record<string, any> | null;
  beadsTaskID: string | null;
}

/**
 * Tasks response type
 */
export interface TasksData {
  tasks: AgentTask[];
}

/**
 * Hook to fetch active agent tasks with auto-refresh
 * @returns React Query result with tasks data
 */
export function useTasks() {
  return useGraphQLQuery<TasksData>('tasks', TASKS_QUERY, undefined, {
    refetchInterval: 3000, // Refresh every 3 seconds
  });
}
