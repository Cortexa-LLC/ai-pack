
import { AgentTask } from '../hooks/useTasks';

/**
 * Props for TaskList component
 */
interface TaskListProps {
  /** Array of agent tasks to display */
  tasks: AgentTask[];
}

/**
 * Get status badge color based on task status
 */
function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'completed':
      return 'bg-green-900 text-green-300';
    case 'running':
      return 'bg-blue-900 text-blue-300';
    case 'failed':
      return 'bg-red-900 text-red-300';
    case 'pending':
      return 'bg-gray-700 text-gray-300';
    default:
      return 'bg-gray-700 text-gray-300';
  }
}

/**
 * Component for displaying a list of agent tasks
 */
function TaskList({ tasks }: TaskListProps) {
  if (tasks.length === 0) {
    return (
      <div className="text-center py-12 text-gray-400">
        No tasks currently running
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {tasks.map((task) => (
        <div
          key={task.taskID}
          className="bg-gray-800 rounded-lg p-6 border border-gray-700"
        >
          <div className="flex items-start justify-between mb-3">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <span className="inline-block px-3 py-1 text-sm font-medium rounded-full bg-purple-900 text-purple-300">
                  {task.role}
                </span>
                <span className={`inline-block px-3 py-1 text-sm font-medium rounded-full ${getStatusColor(task.status)}`}>
                  {task.status}
                </span>
                {task.taskID && (
                  <span className="text-xs text-gray-500">
                    {task.taskID}
                  </span>
                )}
              </div>
              <p className="text-white font-medium">{task.task}</p>
            </div>
          </div>

          {task.error && (
            <div className="mt-3 p-3 bg-red-900/20 border border-red-800 rounded text-red-400 text-sm">
              Error: {task.error}
            </div>
          )}

          {task.result && (
            <div className="mt-3 p-3 bg-green-900/20 border border-green-800 rounded text-green-400 text-sm">
              Result: {task.result}
            </div>
          )}

          <div className="mt-3 text-xs text-gray-500">
            Created: {new Date(task.createdAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
            {task.completedAt && ` • Completed: ${new Date(task.completedAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}`}
          </div>
        </div>
      ))}
    </div>
  );
}

export default TaskList;
