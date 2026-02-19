import { useState, useEffect, useRef } from 'react';
import { useMetrics } from './hooks/useMetrics';
import { useTasks } from './hooks/useTasks';
import { useDetailedMetrics } from './hooks/useDetailedMetrics';
import MetricsCard from './components/MetricsCard';
import ChatPanel from './components/ChatPanel';
import CostBreakdown from './components/CostBreakdown';
import PerformanceDashboard from './components/PerformanceDashboard';

/**
 * Format milliseconds into human-readable duration
 */
function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms.toFixed(0)}ms`;
  const seconds = ms / 1000;
  if (seconds < 60) return `${seconds.toFixed(1)}s`;
  const minutes = seconds / 60;
  if (minutes < 60) return `${minutes.toFixed(1)}m`;
  const hours = minutes / 60;
  if (hours < 24) return `${hours.toFixed(1)}h`;
  const days = hours / 24;
  return `${days.toFixed(1)}d`;
}

/**
 * Format large numbers into human-readable format (K/M/G)
 */
function formatNumber(num: number): string {
  if (num < 1000) return num.toString();
  if (num < 1000000) return `${(num / 1000).toFixed(1)}K`;
  if (num < 1000000000) return `${(num / 1000000).toFixed(1)}M`;
  return `${(num / 1000000000).toFixed(1)}G`;
}

/**
 * Filter and clean log lines - removes JSON structured logs and formats escaped newlines
 */
function filterLogLines(lines: string[]): string[] {
  const filtered: string[] = [];

  for (const line of lines) {
    const trimmed = line.trim();

    // Filter out empty lines
    if (!trimmed) continue;

    // Filter out JSON structured log lines (start with { and contain typical log fields)
    if (trimmed.startsWith('{') && (trimmed.includes('"time"') || trimmed.includes('"level"') || trimmed.includes('"msg"'))) {
      continue;
    }

    // Check if line contains many escaped newlines (likely serialized output)
    // If it has more than 10 \n sequences, it's probably a large text block that should be split
    const escapedNewlineCount = (trimmed.match(/\\n/g) || []).length;
    if (escapedNewlineCount > 10) {
      // Split on \n and add each part as a separate line, filtering out JSON artifacts
      const parts = trimmed.split('\\n')
        .map((p: string) => p.trim())
        .filter((p: string) => {
          // Filter out empty parts
          if (!p) return false;
          // Filter out parts that are only JSON delimiters
          // Be aggressive: remove anything that looks like JSON wrapper
          if (p.match(/^["']*[}{\]]+["']*$/)) {
            return false;
          }
          return true;
        });
      filtered.push(...parts);
    } else {
      // Normal line, just add it (but still filter JSON wrapper artifacts)
      if (trimmed.match(/^["']*[}{\]]+["']*$/)) {
        continue;
      }
      filtered.push(line);
    }
  }

  return filtered;
}

/**
 * Main application component for AI-Pack monitoring dashboard
 */
declare const __APP_VERSION__: string;
declare const __GIT_COMMIT__: string;

function App() {
  const { data: metricsData, isLoading: metricsLoading, isError: metricsError } = useMetrics();
  const { data: tasksData, isLoading: tasksLoading, isError: tasksError, refetch: refetchTasks } = useTasks();
  const { data: detailedMetrics } = useDetailedMetrics();
  const [serverVersion, setServerVersion] = useState<string | null>(null);
  const [selectedTask, setSelectedTask] = useState<string | null>(null);
  const [selectedMetric, setSelectedMetric] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'tasks' | 'server-logs' | 'task-logs' | 'performance'>('tasks');
  const [logs, setLogs] = useState<string[]>([]);
  const [serverLogs, setServerLogs] = useState<any[]>([]);
  const [taskDateFilter, setTaskDateFilter] = useState<'today' | '3d' | '5d' | '7d' | '10d' | '30d' | 'all'>(() => {
    const saved = localStorage.getItem('ai-pack-task-date-filter');
    return (saved as any) || '30d';
  });
  const [followLogs, setFollowLogs] = useState(true);
  const [followServerLogs, setFollowServerLogs] = useState(true);

  // Fetch server version once on mount
  useEffect(() => {
    fetch('/health').then(r => r.json()).then(data => {
      if (data.version) setServerVersion(`${data.version}-${data.commit || 'unknown'}`);
    }).catch(() => {});
  }, []);

  // Save date filter preference to localStorage
  useEffect(() => {
    localStorage.setItem('ai-pack-task-date-filter', taskDateFilter);
  }, [taskDateFilter]);

  // Filter tasks by date
  const filterTasksByDate = (tasks: any[]) => {
    if (taskDateFilter === 'all') return tasks;

    const now = new Date();
    let cutoffDays: number;

    switch (taskDateFilter) {
      case 'today':
        cutoffDays = 1;
        break;
      case '3d':
        cutoffDays = 3;
        break;
      case '5d':
        cutoffDays = 5;
        break;
      case '7d':
        cutoffDays = 7;
        break;
      case '10d':
        cutoffDays = 10;
        break;
      case '30d':
        cutoffDays = 30;
        break;
      default:
        cutoffDays = 30;
    }

    const cutoffDate = new Date(now.getTime() - cutoffDays * 24 * 60 * 60 * 1000);

    return tasks.filter(task => {
      const taskDate = new Date(task.completedAt || task.updatedAt || task.createdAt);
      return taskDate >= cutoffDate;
    });
  };

  // Column widths for server logs table
  const [logColumnWidths, setLogColumnWidths] = useState(() => {
    const saved = localStorage.getItem('ai-pack-log-column-widths');
    return saved ? JSON.parse(saved) : { time: 100, level: 65, message: 130, attributes: 0 }; // 0 = flex/auto
  });
  const [resizingColumn, setResizingColumn] = useState<string | null>(null);
  const [columnResizeStartX, setColumnResizeStartX] = useState(0);
  const [columnResizeStartWidth, setColumnResizeStartWidth] = useState(0);

  // Column widths for Recent Turns table
  const [turnColumnWidths, setTurnColumnWidths] = useState(() => {
    const saved = localStorage.getItem('ai-pack-turn-column-widths');
    return saved ? JSON.parse(saved) : { turn: 60, duration: 80, input: 80, output: 80, task: 0 }; // 0 = flex/auto
  });
  const [resizingTurnColumn, setResizingTurnColumn] = useState<string | null>(null);
  const [turnResizeStartX, setTurnResizeStartX] = useState(0);
  const [turnResizeStartWidth, setTurnResizeStartWidth] = useState(0);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const serverLogsEndRef = useRef<HTMLTableRowElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const serverLogsEventSourceRef = useRef<EventSource | null>(null);
  const followLogsRef = useRef(followLogs);
  const followServerLogsRef = useRef(followServerLogs);

  // Confirmation modal state
  const [confirmModal, setConfirmModal] = useState<{
    show: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({
    show: false,
    title: '',
    message: '',
    onConfirm: () => {},
  });

  const [alertModal, setAlertModal] = useState<{
    show: boolean;
    title: string;
    message: string;
  }>({
    show: false,
    title: '',
    message: '',
  });

  const showAlert = (title: string, message: string) => {
    setAlertModal({ show: true, title, message });
  };

  // Panel widths (resizable) - only chat panel now
  const [chatWidth, setChatWidth] = useState(() => {
    const saved = localStorage.getItem('ai-pack-chat-width');
    return saved ? parseInt(saved) : 384; // default 384px (w-96)
  });
  const [isResizingChat, setIsResizingChat] = useState(false);
  const [resizeStartX, setResizeStartX] = useState(0);
  const [resizeStartWidth, setResizeStartWidth] = useState(0);

  // Save panel widths to localStorage
  useEffect(() => {
    localStorage.setItem('ai-pack-chat-width', chatWidth.toString());
  }, [chatWidth]);

  // Handle resize for chat panel
  useEffect(() => {
    if (!isResizingChat) return;

    const handleMouseMove = (e: MouseEvent) => {
      const delta = resizeStartX - e.clientX; // Positive delta = dragging left = wider panel
      const newWidth = resizeStartWidth + delta;
      setChatWidth(Math.max(300, Math.min(800, newWidth))); // Min 300px, max 800px
    };

    const handleMouseUp = () => {
      setIsResizingChat(false);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isResizingChat, resizeStartX, resizeStartWidth]);

  // Update follow refs when state changes
  useEffect(() => {
    followLogsRef.current = followLogs;
  }, [followLogs]);

  useEffect(() => {
    followServerLogsRef.current = followServerLogs;
  }, [followServerLogs]);

  // Auto-scroll to bottom when following logs
  useEffect(() => {
    if (followLogsRef.current && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'auto' });
    }
  }, [logs]);

  // Auto-scroll server logs
  useEffect(() => {
    if (followServerLogsRef.current && serverLogsEndRef.current) {
      serverLogsEndRef.current.scrollIntoView({ behavior: 'auto' });
    }
  }, [serverLogs]);

  // Save log column widths
  useEffect(() => {
    localStorage.setItem('ai-pack-log-column-widths', JSON.stringify(logColumnWidths));
  }, [logColumnWidths]);

  // Save turn column widths
  useEffect(() => {
    localStorage.setItem('ai-pack-turn-column-widths', JSON.stringify(turnColumnWidths));
  }, [turnColumnWidths]);

  // Handle column resize for server logs
  useEffect(() => {
    if (!resizingColumn) return;

    const handleMouseMove = (e: MouseEvent) => {
      const delta = e.clientX - columnResizeStartX;
      const newWidth = Math.max(50, columnResizeStartWidth + delta);
      setLogColumnWidths((prev: Record<string, number>) => ({
        ...prev,
        [resizingColumn]: newWidth
      }));
    };

    const handleMouseUp = () => {
      setResizingColumn(null);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [resizingColumn, columnResizeStartX, columnResizeStartWidth]);

  // Handle column resize for Recent Turns table
  useEffect(() => {
    if (!resizingTurnColumn) return;

    const handleMouseMove = (e: MouseEvent) => {
      const delta = e.clientX - turnResizeStartX;
      const newWidth = Math.max(50, turnResizeStartWidth + delta);
      setTurnColumnWidths((prev: Record<string, number>) => ({
        ...prev,
        [resizingTurnColumn]: newWidth
      }));
    };

    const handleMouseUp = () => {
      setResizingTurnColumn(null);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [resizingTurnColumn, turnResizeStartX, turnResizeStartWidth]);

  // Fetch server logs when on server logs tab
  useEffect(() => {
    if (activeTab !== 'server-logs') {
      if (serverLogsEventSourceRef.current) {
        serverLogsEventSourceRef.current.close();
        serverLogsEventSourceRef.current = null;
      }
      return;
    }

    // Fetch recent logs first
    fetch('/logs/recent?limit=50')
      .then(res => res.json())
      .then(data => {
        if (data.logs) {
          setServerLogs(data.logs);
        }
      })
      .catch(err => console.error('Failed to load recent logs:', err));

    // Set up SSE for streaming logs
    try {
      const eventSource = new EventSource('/logs/stream');
      serverLogsEventSourceRef.current = eventSource;

      eventSource.addEventListener('log', (event) => {
        try {
          const logEntry = JSON.parse(event.data);
          setServerLogs(prev => [...prev, logEntry].slice(-200)); // Keep last 200 logs
        } catch (err) {
          console.error('Failed to parse log entry:', err);
        }
      });

      eventSource.onerror = (err) => {
        console.error('Server logs stream error:', err);
        eventSource.close();
      };
    } catch (err) {
      console.error('Failed to set up server logs stream:', err);
    }

    return () => {
      if (serverLogsEventSourceRef.current) {
        serverLogsEventSourceRef.current.close();
        serverLogsEventSourceRef.current = null;
      }
    };
  }, [activeTab]);

  // Helper function to select a task and switch to task logs tab
  const selectTask = (taskId: string) => {
    setSelectedTask(taskId);
    setActiveTab('task-logs');
    // Push to browser history so back button works
    window.history.pushState({ taskId, tab: 'task-logs' }, '', `#task/${taskId}`);
  };

  const cancelTask = async (taskID: string, event?: React.MouseEvent, skipConfirm = false) => {
    // Stop propagation if called from within a clickable card
    if (event) {
      event.stopPropagation();
    }

    if (!skipConfirm) {
      // Show confirmation modal
      setConfirmModal({
        show: true,
        title: 'Cancel Task',
        message: `Are you sure you want to cancel this task?\n\nTask ID: ${taskID}`,
        onConfirm: () => {
          setConfirmModal({ show: false, title: '', message: '', onConfirm: () => {} });
          cancelTask(taskID, undefined, true);
        },
      });
      return;
    }

    try {
      // First, cancel the running agent
      const cancelResponse = await fetch('/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: `
            mutation CancelAgent($taskID: String!) {
              cancelAgent(taskID: $taskID)
            }
          `,
          variables: { taskID },
        }),
      });

      const cancelResult = await cancelResponse.json();

      if (cancelResult.errors) {
        showAlert('Error', `Failed to cancel task: ${cancelResult.errors[0].message}`);
        return;
      }

      // Then, close the task in Beads
      const closeResponse = await fetch('/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            mutation CloseTask($taskID: String!) {
              closeTask(taskID: $taskID) {
                success
                message
                taskID
              }
            }
          `,
          variables: { taskID },
        }),
      });

      const closeResult = await closeResponse.json();

      if (closeResult.errors) {
        showAlert('Warning', `Task cancelled but failed to close: ${closeResult.errors[0].message}`);
      } else if (closeResult.data?.closeTask?.success) {
        // Refresh tasks to show updated status
        setTimeout(() => refetchTasks(), 300);
      } else {
        showAlert('Warning', 'Task cancelled but failed to close: ' + (closeResult.data?.closeTask?.message || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to cancel task:', error);
      showAlert('Error', 'Failed to cancel task. Check console for details.');
    }
  };

  const retryTask = async (taskID: string, taskDescription: string, event?: React.MouseEvent, skipConfirm = false) => {
    // Stop propagation if called from within a clickable card
    if (event) {
      event.stopPropagation();
    }

    if (!skipConfirm) {
      // Show confirmation modal
      setConfirmModal({
        show: true,
        title: 'Retry Task',
        message: `Are you sure you want to retry this task?\n\n${taskDescription.split('\n')[0]}`,
        onConfirm: () => {
          setConfirmModal({ show: false, title: '', message: '', onConfirm: () => {} });
          retryTask(taskID, taskDescription, undefined, true);
        },
      });
      return;
    }

    try {
      const response = await fetch('/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: `
            mutation RetryTask($taskID: String!) {
              retryTask(taskID: $taskID) {
                success
                message
                taskID
              }
            }
          `,
          variables: { taskID },
        }),
      });

      const result = await response.json();

      if (result.errors) {
        showAlert('Error', `Failed to retry task: ${result.errors[0].message}`);
      } else if (result.data?.retryTask?.success) {
        const newTaskID = result.data.retryTask.taskID;
        showAlert('Success', `Task retried! New task ID: ${newTaskID}`);
        // Switch to the new task and show its logs
        selectTask(newTaskID);
      } else {
        showAlert('Error', 'Failed to retry task: ' + (result.data?.retryTask?.message || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to retry task:', error);
      showAlert('Error', 'Failed to retry task. Check console for details.');
    }
  };

  const closeTask = async (taskID: string, taskDescription: string, event?: React.MouseEvent, skipConfirm = false) => {
    // Stop propagation if called from within a clickable card
    if (event) {
      event.stopPropagation();
    }

    if (!skipConfirm) {
      // Show confirmation modal
      setConfirmModal({
        show: true,
        title: 'Close Task',
        message: `Are you sure you want to close/dismiss this task?\n\n${taskDescription.split('\n')[0]}`,
        onConfirm: () => {
          setConfirmModal({ show: false, title: '', message: '', onConfirm: () => {} });
          closeTask(taskID, taskDescription, undefined, true);
        },
      });
      return;
    }

    try {
      const response = await fetch('/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            mutation CloseTask($taskID: String!) {
              closeTask(taskID: $taskID) {
                success
                message
                taskID
              }
            }
          `,
          variables: { taskID },
        }),
      });

      const result = await response.json();

      if (result.errors) {
        showAlert('Error', `Failed to close task: ${result.errors[0].message}`);
      } else if (result.data?.closeTask?.success) {
        // Refresh tasks to show updated status (small delay to ensure backend update completes)
        setTimeout(() => refetchTasks(), 300);
      } else {
        showAlert('Error', 'Failed to close task: ' + (result.data?.closeTask?.message || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to close task:', error);
      showAlert('Error', 'Failed to close task. Check console for details.');
    }
  };

  const startAgent = async (taskID: string, taskDescription: string, projectRoot?: string | null, event?: React.MouseEvent, skipConfirm = false, role = 'engineer') => {
    // Stop propagation if called from within a clickable card
    if (event) {
      event.stopPropagation();
    }

    if (!skipConfirm) {
      // Show confirmation modal
      const roleName = role === 'orchestrator' ? 'Orchestrator' : 'Engineer';
      setConfirmModal({
        show: true,
        title: `Start ${roleName}`,
        message: `Start ${roleName.toLowerCase()} for this task?\n\n${taskDescription.split('\n')[0]}`,
        onConfirm: () => {
          setConfirmModal({ show: false, title: '', message: '', onConfirm: () => {} });
          startAgent(taskID, taskDescription, projectRoot, undefined, true, role);
        },
      });
      return;
    }

    try {
      const response = await fetch(`/a2a/start/${taskID}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          role: role,
          project_root: projectRoot || '', // Use task's project root
        }),
      });

      const result = await response.json();

      if (!response.ok) {
        showAlert('Error', `Failed to start agent: ${result.message || response.statusText}`);
        return;
      }

      if (result.success) {
        const newTaskID = result.task_id;
        // Switch to the new task and show its logs (no success dialog needed)
        selectTask(newTaskID);
      } else {
        showAlert('Error', 'Failed to start agent: ' + (result.message || 'Unknown error'));
      }
    } catch (error) {
      console.error('Failed to start agent:', error);
      showAlert('Error', 'Failed to start agent. Check console for details.');
    }
  };

  // Fetch and stream logs when a task is selected or when switching to task logs tab
  useEffect(() => {
    if (!selectedTask || activeTab !== 'task-logs') {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      if (!selectedTask) {
        setLogs([]);
      }
      return;
    }

    // Clear logs immediately when switching tasks to avoid showing stale logs
    setLogs([]);

    // Check if task is active (only check once, don't refetch when tasksData updates)
    const task = tasksData?.tasks.find(t => t.taskID === selectedTask);
    const isActiveTask = task?.status === 'IN_PROGRESS' || task?.status === 'in_progress';

    // Fetch initial logs
    fetch(`http://localhost:8080/a2a/tasks/${selectedTask}/logs`)
      .then(res => {
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        return res.text();
      })
      .then(text => {
        if (!text || text.trim() === '') {
          setLogs(['No logs available for this task']);
          return;
        }
        const allLines = text.split('\n');
        const filteredLines = filterLogLines(allLines);
        setLogs(filteredLines.length > 0 ? filteredLines : ['No logs available']);
      })
      .catch(err => {
        console.error('Error fetching logs:', err);
        setLogs([
          'Unable to load logs for this task.',
          '',
          'This may be because:',
          '- The task has not started yet',
          '- The task logs have been archived',
          '- The log endpoint is not available',
          '',
          `Error: ${err.message}`
        ]);
      });

    // Set up SSE for log streaming (only for active tasks)
    if (isActiveTask) {
      try {
        const eventSource = new EventSource(`http://localhost:8080/a2a/tasks/${selectedTask}/logs?stream=true`);
        eventSourceRef.current = eventSource;

        // Handle connected event
        eventSource.addEventListener('connected', (event) => {
          console.log('Log stream connected:', event.data);
        });

        // Handle log events
        eventSource.addEventListener('log', (event) => {
          try {
            // Sanitize event data to handle control characters that break JSON parsing
            // Remove non-printable control characters (except newline, tab, carriage return which should be escaped)
            const sanitizedData = event.data.replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]/g, '');

            const logData = JSON.parse(sanitizedData);
            if (logData.line && logData.line.trim()) {
              const line = logData.line;
              const trimmed = line.trim();

              // Filter out JSON structured logs
              if (trimmed.startsWith('{') && (trimmed.includes('"time"') || trimmed.includes('"level"') || trimmed.includes('"msg"'))) {
                return; // Skip JSON log lines
              }

              // Check if line contains many escaped newlines (likely serialized output)
              const escapedNewlineCount = (trimmed.match(/\\n/g) || []).length;
              if (escapedNewlineCount > 10) {
                // Split on \n and add each part as a separate line, filtering out JSON artifacts
                const parts = trimmed.split('\\n')
                  .map((p: string) => p.trim())
                  .filter((p: string) => {
                    // Filter out empty parts
                    if (!p) return false;
                    // Filter out parts that are only JSON delimiters - be aggressive
                    if (p.match(/^["']*[}{\]]+["']*$/)) {
                      return false;
                    }
                    return true;
                  });
                setLogs(prev => [...prev, ...parts]);
              } else {
                // Normal line - filter JSON wrapper artifacts aggressively
                if (trimmed.match(/^["']*[}{\]]+["']*$/)) {
                  return; // Skip JSON wrapper characters
                }
                setLogs(prev => [...prev, line]);
              }
            }
          } catch (err) {
            // Silently skip malformed log lines - they're often just streaming artifacts
            // Uncomment for debugging: console.warn('Skipped malformed log data:', err);
          }
        });

        // Handle complete event
        eventSource.addEventListener('complete', (event) => {
          console.log('Task completed:', event.data);
          eventSource.close();
        });

        eventSource.onerror = (err) => {
          console.error('SSE error:', err);
          eventSource.close();
        };

        return () => {
          eventSource.close();
        };
      } catch (err) {
        console.error('Failed to set up SSE:', err);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedTask, activeTab]);

  // Handle browser back button
  useEffect(() => {
    const handlePopState = (event: PopStateEvent) => {
      // If user presses back, return to tasks view
      if (event.state?.taskId) {
        // Going back from a task, but history has task state - do nothing
        return;
      } else if (selectedTask) {
        // Going back from task logs to tasks view
        setSelectedTask(null);
        setActiveTab('tasks');
      }
    };

    window.addEventListener('popstate', handlePopState);

    // Set initial state if needed
    if (!window.history.state) {
      window.history.replaceState({ tab: 'tasks' }, '', '#tasks');
    }

    return () => {
      window.removeEventListener('popstate', handlePopState);
    };
  }, [selectedTask]);

  return (
    <div className="h-screen flex flex-col bg-gray-900 text-white">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700 px-4 py-3 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="AI-Pack" className="h-8 w-8" />
            <h1 className="text-xl font-bold">AI-Pack Console</h1>
          </div>
          <div className="text-xs text-gray-500 font-mono">
            gui {__APP_VERSION__}-{__GIT_COMMIT__}{serverVersion ? ` · server ${serverVersion}` : ''}
          </div>
        </div>
      </header>

      {/* Main Layout: Left (Tasks + Metrics) and Right (Chat) */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left: Tasks and Metrics */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {/* Tabs Section - Takes remaining space */}
          <div className="flex-1 flex flex-col overflow-hidden min-h-0">
            {/* Tab Bar */}
            <div className="flex items-center border-b border-gray-700 bg-gray-800 px-4 flex-shrink-0">
              <button
                onClick={() => setActiveTab('tasks')}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'tasks'
                    ? 'border-blue-500 text-blue-400'
                    : 'border-transparent text-gray-400 hover:text-gray-300'
                }`}
              >
                📋 Tasks
              </button>
              <button
                onClick={() => setActiveTab('server-logs')}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'server-logs'
                    ? 'border-blue-500 text-blue-400'
                    : 'border-transparent text-gray-400 hover:text-gray-300'
                }`}
              >
                🖥️ Server Log
              </button>
              <button
                onClick={() => setActiveTab('task-logs')}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'task-logs'
                    ? 'border-blue-500 text-blue-400'
                    : 'border-transparent text-gray-400 hover:text-gray-300'
                }`}
              >
                📄 Task Logs {selectedTask && '✓'}
              </button>
              <button
                onClick={() => setActiveTab('performance')}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'performance'
                    ? 'border-blue-500 text-blue-400'
                    : 'border-transparent text-gray-400 hover:text-gray-300'
                }`}
              >
                📊 Performance
              </button>
              {selectedTask && activeTab === 'task-logs' && (
                <div className="ml-auto flex items-center gap-2">
                  {/* Only show Following button for in-progress tasks */}
                  {(() => {
                    const task = tasksData?.tasks.find(t => t.taskID === selectedTask);
                    return task?.status === 'IN_PROGRESS' || task?.status === 'in_progress';
                  })() && (
                    <button
                      onClick={() => setFollowLogs(!followLogs)}
                      className={`px-2 py-1 text-xs rounded ${followLogs ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300'}`}
                      title={followLogs ? 'Following logs' : 'Follow disabled'}
                    >
                      {followLogs ? '📍 Following' : '⏸ Paused'}
                    </button>
                  )}
                  <button
                    onClick={() => {
                      // Use browser back to maintain history
                      window.history.back();
                    }}
                    className="text-gray-400 hover:text-white text-sm px-2"
                  >
                    ✕ Close
                  </button>
                </div>
              )}
            </div>

            {/* Tab Content */}
            {activeTab === 'tasks' && (
              <div className="flex-1 overflow-hidden p-4 flex flex-col">

            {/* Date Filter */}
            <div className="mb-3 flex items-center gap-2">
              <label className="text-sm text-gray-400">Show tasks from:</label>
              <select
                value={taskDateFilter}
                onChange={(e) => setTaskDateFilter(e.target.value as 'today' | '3d' | '5d' | '7d' | '10d' | '30d' | 'all')}
                className="bg-gray-800 border border-gray-600 rounded px-2 py-1 text-sm text-gray-300 focus:outline-none focus:border-blue-500"
              >
                <option value="today">Today</option>
                <option value="3d">Last 3 days</option>
                <option value="5d">Last 5 days</option>
                <option value="7d">Last 7 days</option>
                <option value="10d">Last 10 days</option>
                <option value="30d">Last 30 days</option>
                <option value="all">All time</option>
              </select>
            </div>

            {tasksLoading && (
              <div className="text-gray-400">Loading tasks...</div>
            )}

            {tasksError && (
              <div className="bg-red-900/20 border border-red-800 rounded p-3 text-red-400">
                Error loading tasks.
              </div>
            )}

            {tasksData && (
              <div className="flex-1 min-h-0 overflow-hidden">
                <div className="grid grid-cols-5 gap-4 h-full">
                    {/* Queued/Orphaned Lane */}
                    <div className="flex flex-col h-full min-h-0 border-2 border-blue-600 rounded-lg bg-gray-800/50">
                      <div className="bg-blue-900 border-b-2 border-blue-600 p-3 flex-shrink-0">
                        <h3 className="font-semibold text-blue-300 flex items-center justify-between text-base">
                          <span>⏳ Queued / Ready</span>
                          <span className="text-sm bg-blue-800 px-2 py-1 rounded-full">
                            {tasksData.tasks.filter(t => t.status === 'QUEUED' || t.status === 'queued' || t.status === 'OPEN' || t.status === 'open').length}
                          </span>
                        </h3>
                      </div>
                      <div className="flex-1 space-y-2 overflow-auto p-3">
                        {tasksData.tasks.filter(t => t.status === 'QUEUED' || t.status === 'queued' || t.status === 'OPEN' || t.status === 'open').length === 0 ? (
                          <div className="text-center text-gray-500 text-sm py-8">
                            No queued tasks
                          </div>
                        ) : (
                          tasksData.tasks
                            .filter(t => t.status === 'QUEUED' || t.status === 'queued' || t.status === 'OPEN' || t.status === 'open')
                            .sort((a, b) => {
                              const timeCompare = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
                              return timeCompare !== 0 ? timeCompare : a.taskID.localeCompare(b.taskID);
                            })
                            .map(task => {
                              const isOpen = task.status === 'OPEN' || task.status === 'open';
                              return (
                                <div
                                  key={task.taskID}
                                  className="bg-gray-800 border border-gray-700 hover:border-blue-500 rounded-lg p-2 transition-colors relative"
                                >
                                  <div
                                    className="cursor-pointer pb-8"
                                    onClick={() => selectTask(task.taskID)}
                                  >
                                    <div className="text-xs font-medium text-white mb-1 line-clamp-2" title={task.task}>
                                      {task.task.split('\n')[0]}
                                    </div>
                                    <div className="text-xs text-gray-500 mb-1">
                                      {task.taskID}
                                    </div>
                                    {task.createdAt && (
                                      <div className="text-xs text-gray-500">
                                        Created: {new Date(task.createdAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
                                      </div>
                                    )}
                                  </div>
                                  <div className="absolute bottom-2 right-2 flex gap-1">
                                    {isOpen && (
                                      <button
                                        onClick={(e) => startAgent(task.taskID, task.task, task.projectRoot, e)}
                                        className="w-6 h-6 flex items-center justify-center bg-green-600 hover:bg-green-700 text-white rounded transition-colors text-xs"
                                        title="Start agent for this task"
                                      >
                                        ▶️
                                      </button>
                                    )}
                                    {task.metadata?.beads_status !== 'closed' && (
                                      <button
                                        onClick={(e) => closeTask(task.taskID, task.task, e)}
                                        className="w-6 h-6 flex items-center justify-center bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors text-xs"
                                        title="Close/dismiss task"
                                      >
                                        ✕
                                      </button>
                                    )}
                                  </div>
                                </div>
                              );
                            })
                        )}
                      </div>
                    </div>

                    {/* In Progress Lane */}
                    <div className="flex flex-col h-full min-h-0 border-2 border-yellow-600 rounded-lg bg-gray-800/50">
                      <div className="bg-yellow-900 border-b-2 border-yellow-600 p-3 flex-shrink-0">
                        <h3 className="font-semibold text-yellow-300 flex items-center justify-between text-base">
                          <span>🔄 In Progress</span>
                          <span className="text-sm bg-yellow-800 px-2 py-1 rounded-full">
                            {tasksData.tasks.filter(t => t.status === 'IN_PROGRESS' || t.status === 'in_progress').length}
                          </span>
                        </h3>
                      </div>
                      <div className="flex-1 space-y-2 overflow-auto p-3">
                        {tasksData.tasks.filter(t => t.status === 'IN_PROGRESS' || t.status === 'in_progress').length === 0 ? (
                          <div className="text-center text-gray-500 text-sm py-8">
                            No tasks in progress
                          </div>
                        ) : (
                          tasksData.tasks
                            .filter(t => t.status === 'IN_PROGRESS' || t.status === 'in_progress')
                            .sort((a, b) => {
                              const timeCompare = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
                              return timeCompare !== 0 ? timeCompare : a.taskID.localeCompare(b.taskID);
                            })
                            .map(task => (
                              <div
                                key={task.taskID}
                                className="bg-gray-800 border border-gray-700 rounded-lg p-2 hover:border-yellow-500 cursor-pointer transition-colors relative"
                                onClick={() => selectTask(task.taskID)}
                              >
                                <div className="pb-8">
                                  <div className="text-xs font-medium text-white mb-1 line-clamp-2" title={task.task}>
                                    {task.task.split('\n')[0]}
                                  </div>
                                  <div className="text-xs text-gray-500 mb-1">
                                    {task.taskID}
                                  </div>
                                  {task.createdAt && (
                                    <div className="text-xs text-gray-500">
                                      Created: {new Date(task.createdAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
                                    </div>
                                  )}
                                </div>
                                <div className="absolute bottom-2 right-2">
                                  <button
                                    onClick={(e) => cancelTask(task.taskID, e)}
                                    className="w-6 h-6 flex items-center justify-center bg-red-600 hover:bg-red-700 text-white rounded transition-colors text-xs"
                                    title="Cancel task"
                                  >
                                    🛑
                                  </button>
                                </div>
                              </div>
                            ))
                        )}
                      </div>
                    </div>

                    {/* Blocked Lane */}
                    <div className="flex flex-col h-full min-h-0 border-2 border-orange-600 rounded-lg bg-gray-800/50">
                      <div className="bg-orange-900 border-b-2 border-orange-600 p-3 flex-shrink-0">
                        <h3 className="font-semibold text-orange-300 flex items-center justify-between text-base">
                          <span>🚧 Blocked</span>
                          <span className="text-sm bg-orange-800 px-2 py-1 rounded-full">
                            {tasksData.tasks.filter(t => t.status === 'BLOCKED' || t.status === 'blocked').length}
                          </span>
                        </h3>
                      </div>
                      <div className="flex-1 space-y-2 overflow-auto p-3">
                        {tasksData.tasks.filter(t => t.status === 'BLOCKED' || t.status === 'blocked').length === 0 ? (
                          <div className="text-center text-gray-500 text-sm py-8">
                            No blocked tasks
                          </div>
                        ) : (
                          tasksData.tasks
                            .filter(t => t.status === 'BLOCKED' || t.status === 'blocked')
                            .sort((a, b) => {
                              const timeCompare = new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
                              return timeCompare !== 0 ? timeCompare : a.taskID.localeCompare(b.taskID);
                            })
                            .map(task => (
                              <div
                                key={task.taskID}
                                className="bg-gray-800 border border-gray-700 rounded-lg p-2 hover:border-orange-500 cursor-pointer transition-colors relative"
                                onClick={() => selectTask(task.taskID)}
                              >
                                <div className="pb-8">
                                  <div className="text-xs font-medium text-white mb-1 line-clamp-2" title={task.task}>
                                    {task.task.split('\n')[0]}
                                  </div>
                                  <div className="text-xs text-gray-500 mb-1">
                                    {task.taskID}
                                  </div>
                                  {task.error && (
                                    <div className="text-xs text-orange-400 mt-1 line-clamp-2" title={task.error}>
                                      ⚠️ {task.error}
                                    </div>
                                  )}
                                  {task.updatedAt && (
                                    <div className="text-xs text-gray-500 mt-1">
                                      Blocked: {new Date(task.updatedAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
                                    </div>
                                  )}
                                </div>
                                <div className="absolute bottom-2 right-2 flex gap-1">
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      // Extract Beads task ID from agent task ID (remove timestamp suffix)
                                      const beadsTaskId = task.taskID.replace(/-\d{8}-\d{6}$/, '');
                                      startAgent(beadsTaskId, task.task, task.projectRoot, e, false, 'orchestrator');
                                    }}
                                    className="w-6 h-6 flex items-center justify-center bg-purple-600 hover:bg-purple-700 text-white rounded transition-colors text-xs"
                                    title="Send to Orchestrator for review"
                                  >
                                    🤖
                                  </button>
                                  <button
                                    onClick={(e) => closeTask(task.taskID, task.task, e)}
                                    className="w-6 h-6 flex items-center justify-center bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors text-xs"
                                    title="Close/dismiss task"
                                  >
                                    ✕
                                  </button>
                                </div>
                              </div>
                            ))
                        )}
                      </div>
                    </div>

                    {/* Completed Lane */}
                    <div className="flex flex-col h-full min-h-0 border-2 border-green-600 rounded-lg bg-gray-800/50">
                      <div className="bg-green-900 border-b-2 border-green-600 p-3 flex-shrink-0">
                        <h3 className="font-semibold text-green-300 flex items-center justify-between text-base">
                          <span>✅ Completed</span>
                          <span className="text-sm bg-green-800 px-2 py-1 rounded-full">
                            {filterTasksByDate(tasksData.tasks.filter(t => t.status === 'COMPLETED' || t.status === 'completed')).length}
                          </span>
                        </h3>
                      </div>
                      <div className="flex-1 space-y-2 overflow-auto p-3">
                        {tasksData.tasks.filter(t => t.status === 'COMPLETED' || t.status === 'completed').length === 0 ? (
                          <div className="text-center text-gray-500 text-sm py-8">
                            No completed tasks
                          </div>
                        ) : (
                          filterTasksByDate(
                            tasksData.tasks.filter(t => t.status === 'COMPLETED' || t.status === 'completed')
                          )
                            .sort((a, b) => {
                              const timeCompare = new Date(b.completedAt || b.updatedAt).getTime() - new Date(a.completedAt || a.updatedAt).getTime();
                              return timeCompare !== 0 ? timeCompare : a.taskID.localeCompare(b.taskID);
                            })
                            .map(task => (
                              <div
                                key={task.taskID}
                                className="bg-gray-800 border border-gray-700 rounded-lg p-2 hover:border-green-500 transition-colors relative"
                              >
                                <div
                                  className="cursor-pointer pb-8"
                                  onClick={() => selectTask(task.taskID)}
                                >
                                  <div className="text-xs font-medium text-white mb-1 line-clamp-2" title={task.task}>
                                    {task.task.split('\n')[0]}
                                  </div>
                                  <div className="text-xs text-gray-500 mb-1">
                                    {task.taskID}
                                  </div>
                                  {task.completedAt && (
                                    <div className="text-xs text-gray-500">
                                      Completed: {new Date(task.completedAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
                                    </div>
                                  )}
                                </div>
                                {task.metadata?.beads_status !== 'closed' && (
                                  <div className="absolute bottom-2 right-2">
                                    <button
                                      onClick={(e) => closeTask(task.taskID, task.task, e)}
                                      className="w-6 h-6 flex items-center justify-center bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors text-xs"
                                      title="Close/dismiss task"
                                    >
                                      ✕
                                    </button>
                                  </div>
                                )}
                              </div>
                            ))
                        )}
                      </div>
                    </div>

                    {/* Failed Lane */}
                    <div className="flex flex-col h-full min-h-0 border-2 border-red-600 rounded-lg bg-gray-800/50">
                      <div className="bg-red-900 border-b-2 border-red-600 p-3 flex-shrink-0">
                        <h3 className="font-semibold text-red-300 flex items-center justify-between text-base">
                          <span>❌ Failed</span>
                          <span className="text-sm bg-red-800 px-2 py-1 rounded-full">
                            {filterTasksByDate(tasksData.tasks.filter(t => t.status === 'FAILED' || t.status === 'failed')).length}
                          </span>
                        </h3>
                      </div>
                      <div className="flex-1 space-y-2 overflow-auto p-3">
                        {tasksData.tasks.filter(t => t.status === 'FAILED' || t.status === 'failed').length === 0 ? (
                          <div className="text-center text-gray-500 text-sm py-8">
                            No failed tasks
                          </div>
                        ) : (
                          filterTasksByDate(
                            tasksData.tasks.filter(t => t.status === 'FAILED' || t.status === 'failed')
                          )
                            .sort((a, b) => {
                              const timeCompare = new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
                              return timeCompare !== 0 ? timeCompare : a.taskID.localeCompare(b.taskID);
                            })
                            .map(task => (
                              <div
                                key={task.taskID}
                                className="bg-gray-800 border border-gray-700 rounded-lg p-2 hover:border-red-500 transition-colors relative"
                              >
                                <div
                                  className="cursor-pointer pb-8"
                                  onClick={() => selectTask(task.taskID)}
                                >
                                  <div className="text-xs font-medium text-white mb-1 line-clamp-2" title={task.task}>
                                    {task.task.split('\n')[0]}
                                  </div>
                                  <div className="text-xs text-gray-500 mb-1">
                                    {task.taskID}
                                  </div>
                                  {(task.completedAt || task.updatedAt) && (
                                    <div className="text-xs text-gray-500">
                                      Failed: {new Date(task.completedAt || task.updatedAt).toLocaleString(undefined, { month: 'numeric', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })}
                                    </div>
                                  )}
                                </div>
                                {(task.metadata?.beads_status !== 'closed' || task.metadata?.beads_status === undefined) && (
                                  <div className="absolute bottom-2 right-2 flex gap-1">
                                    <button
                                      onClick={(e) => retryTask(task.taskID, task.task, e)}
                                      className="w-6 h-6 flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors text-xs"
                                      title="Retry task"
                                    >
                                      🔄
                                    </button>
                                    <button
                                      onClick={(e) => closeTask(task.taskID, task.task, e)}
                                      className="w-6 h-6 flex items-center justify-center bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors text-xs"
                                      title="Close/dismiss task"
                                    >
                                      ✕
                                    </button>
                                  </div>
                                )}
                              </div>
                            ))
                        )}
                      </div>
                    </div>
                  </div>
              </div>
            )}
            </div>
            )}

            {/* Server Logs Tab Content */}
            {activeTab === 'server-logs' && (
              <div className="flex-1 flex flex-col overflow-hidden">
                <div className="p-3 bg-gray-800 border-b border-gray-700 flex items-center justify-between">
                  <div>
                    <h3 className="text-md font-semibold">Agent Server Logs</h3>
                    <div className="text-xs text-gray-500">Real-time server activity</div>
                  </div>
                  <button
                    onClick={() => setFollowServerLogs(!followServerLogs)}
                    className={`px-2 py-1 text-xs rounded ${followServerLogs ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-300'}`}
                    title={followServerLogs ? 'Following logs' : 'Follow disabled'}
                  >
                    {followServerLogs ? '📍 Following' : '⏸ Paused'}
                  </button>
                </div>
                <div className="flex-1 overflow-auto bg-gray-900">
                  {serverLogs.length === 0 ? (
                    <div className="p-4 text-gray-500 text-sm">Loading server logs...</div>
                  ) : (
                    <table className="w-full text-xs font-mono" style={{ tableLayout: 'fixed' }}>
                      <colgroup>
                        <col style={{ width: `${logColumnWidths.time}px` }} />
                        <col style={{ width: `${logColumnWidths.level}px` }} />
                        <col style={{ width: `${logColumnWidths.message}px` }} />
                        <col />
                      </colgroup>
                      <thead className="sticky top-0 bg-gray-800 border-b border-gray-700">
                        <tr>
                          <th className="text-left p-2 text-gray-400 font-medium relative group">
                            Time
                            <div
                              className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                              onMouseDown={(e) => {
                                setColumnResizeStartX(e.clientX);
                                setColumnResizeStartWidth(logColumnWidths.time);
                                setResizingColumn('time');
                              }}
                            />
                          </th>
                          <th className="text-left p-2 text-gray-400 font-medium relative group">
                            Level
                            <div
                              className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                              onMouseDown={(e) => {
                                setColumnResizeStartX(e.clientX);
                                setColumnResizeStartWidth(logColumnWidths.level);
                                setResizingColumn('level');
                              }}
                            />
                          </th>
                          <th className="text-left p-2 text-gray-400 font-medium relative group">
                            Message
                            <div
                              className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                              onMouseDown={(e) => {
                                setColumnResizeStartX(e.clientX);
                                setColumnResizeStartWidth(logColumnWidths.message);
                                setResizingColumn('message');
                              }}
                            />
                          </th>
                          <th className="text-left p-2 text-gray-400 font-medium">Attributes</th>
                        </tr>
                      </thead>
                      <tbody>
                        {serverLogs.map((log, idx) => {
                          const level = log.level || 'INFO';
                          const levelColor =
                            level === 'ERROR' ? 'text-red-400' :
                            level === 'WARN' ? 'text-yellow-400' :
                            level === 'DEBUG' ? 'text-gray-500' :
                            'text-blue-400';

                          // Parse timestamp to show time with milliseconds and timezone abbreviation
                          let time = '';
                          if (log.timestamp) {
                            const date = new Date(log.timestamp);
                            const timeStr = date.toLocaleTimeString('en-US', {
                              hour12: false,
                              hour: '2-digit',
                              minute: '2-digit',
                              second: '2-digit'
                            });
                            const ms = date.getMilliseconds().toString().padStart(3, '0');
                            // Get timezone abbreviation (e.g., PST, UTC, EST)
                            const tzFormatter = new Intl.DateTimeFormat('en-US', {
                              timeZoneName: 'short'
                            });
                            const tzParts = tzFormatter.formatToParts(date);
                            const tzName = tzParts.find(part => part.type === 'timeZoneName')?.value || '';
                            time = `${timeStr}.${ms} ${tzName}`;
                          }

                          const msg = log.message || '';
                          const attrs = log.attrs || {};

                          // Flatten nested attrs - if attrs has a 'message' field, show it prominently
                          const flattenedAttrs: Record<string, any> = {};
                          Object.entries(attrs).forEach(([key, value]) => {
                            if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
                              // Flatten nested objects
                              Object.entries(value).forEach(([nestedKey, nestedValue]) => {
                                flattenedAttrs[`${key}.${nestedKey}`] = nestedValue;
                              });
                            } else {
                              flattenedAttrs[key] = value;
                            }
                          });

                          // Format attributes for display
                          const renderAttribute = (key: string, value: any) => {
                            const valueStr = Array.isArray(value) ? JSON.stringify(value) : String(value);
                            const colorClass =
                              key === 'error' || key.includes('error') ? 'text-red-400' :
                              key === 'status_code' ? (value >= 400 ? 'text-red-400' : 'text-green-400') :
                              key.includes('token') || key.includes('input') || key.includes('output') ? 'text-purple-400' :
                              key.includes('duration') || key.includes('_ms') ? 'text-amber-400' :
                              key === 'method' ? 'text-cyan-400' :
                              key === 'path' ? 'text-blue-400' :
                              key.includes('task_id') || key.includes('task-id') ? 'text-green-400' :
                              key === 'message' || key.includes('.message') ? 'text-white' :
                              'text-gray-300';

                            return (
                              <span key={key} className="inline-block mr-3">
                                <span className="text-gray-500">{key}=</span>
                                <span className={colorClass}>{valueStr}</span>
                              </span>
                            );
                          };

                          return (
                            <tr key={idx} className="border-b border-gray-800 hover:bg-gray-800/50">
                              <td className="p-2 text-gray-500 font-mono text-xs truncate">{time}</td>
                              <td className={`p-2 font-semibold text-xs truncate ${levelColor}`}>{level}</td>
                              <td className="p-2 text-gray-300 text-xs truncate" title={msg}>{msg}</td>
                              <td className="p-2 text-xs">
                                {Object.entries(flattenedAttrs).map(([key, value]) => renderAttribute(key, value))}
                              </td>
                            </tr>
                          );
                        })}
                        <tr ref={serverLogsEndRef}><td colSpan={4}></td></tr>
                      </tbody>
                    </table>
                  )}
                </div>
              </div>
            )}

            {/* Task Logs Tab Content */}
            {activeTab === 'task-logs' && selectedTask && (
              <div className="flex-1 flex flex-col overflow-hidden">
                <div className="p-4 bg-gray-800 border-b border-gray-700">
                  <div className="flex-1">
                    <h3 className="text-md font-semibold mb-1">Task Logs</h3>
                    <div className="text-xs text-gray-500">Task ID: {selectedTask}</div>
                  </div>
                </div>
                <div className="flex-1 overflow-auto p-4 bg-gray-900 font-mono text-xs">
                  {logs.length === 0 ? (
                    <div className="text-gray-500">Loading logs...</div>
                  ) : (
                    <div className="space-y-1">
                      {logs.map((log, idx) => {
                        // Check if this is a JSON line (StreamEvent)
                        if (log.trim().startsWith('{')) {
                          try {
                            const event = JSON.parse(log);
                            // Format JSON events nicely
                            if (event.type && event.task_id) {
                              // Check if error contains embedded JSON
                              let errorText = event.data?.error || '';
                              let embeddedJson = null;

                              // Try to extract embedded JSON from error message
                              // Find the first { and try to parse from there to the end
                              const jsonStart = errorText.indexOf('{');
                              if (jsonStart !== -1) {
                                try {
                                  const jsonStr = errorText.substring(jsonStart);
                                  embeddedJson = JSON.parse(jsonStr);
                                  // Remove embedded JSON from error text
                                  errorText = errorText.substring(0, jsonStart).trim();
                                } catch (e) {
                                  // Ignore parse errors
                                }
                              }

                              return (
                                <div key={`${selectedTask}-${idx}`} className="text-xs bg-gray-800 border border-gray-700 rounded p-2 my-1">
                                  <div className="flex items-start gap-2 mb-2">
                                    <span className="text-blue-400 font-mono font-semibold">{event.type}</span>
                                    <span className="text-gray-500 text-xs">
                                      {new Date(event.timestamp).toLocaleTimeString()}
                                    </span>
                                  </div>
                                  {errorText && (
                                    <div className="text-red-400 mb-2">{errorText}</div>
                                  )}
                                  {embeddedJson && (
                                    <div className="bg-gray-900 border border-gray-600 rounded p-2 text-xs">
                                      <div className="text-orange-400 font-mono mb-1">API Error:</div>
                                      <div className="text-gray-300">
                                        <span className="text-gray-500">Type:</span> {embeddedJson.error?.type || 'unknown'}
                                      </div>
                                      {embeddedJson.error?.message && (
                                        <div className="text-gray-300 mt-1">
                                          <span className="text-gray-500">Message:</span> {embeddedJson.error.message}
                                        </div>
                                      )}
                                      {embeddedJson.request_id && (
                                        <div className="text-gray-500 mt-1 text-xs">
                                          Request ID: {embeddedJson.request_id}
                                        </div>
                                      )}
                                    </div>
                                  )}
                                </div>
                              );
                            }
                          } catch (e) {
                            // Not valid JSON, display as-is
                          }
                        }
                        // Regular log line
                        return (
                          <div key={`${selectedTask}-${idx}-${log.substring(0, 20)}`} className="text-gray-300 whitespace-pre-wrap break-words">
                            {log}
                          </div>
                        );
                      })}
                      <div ref={logsEndRef} />
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Show message when task logs tab is active but no task selected */}
            {activeTab === 'task-logs' && !selectedTask && (
              <div className="flex-1 flex items-center justify-center text-gray-500">
                <div className="text-center">
                  <p className="text-lg mb-2">📄 No task selected</p>
                  <p className="text-sm">Select a task from the Tasks tab to view its logs</p>
                </div>
              </div>
            )}

            {/* Performance Dashboard Tab */}
            {activeTab === 'performance' && (
              <div className="flex-1 overflow-auto">
                <PerformanceDashboard apiUrl={window.location.origin} />
              </div>
            )}
          </div>

          {/* Bottom Section: Performance Metrics (Grafana-style) */}
          <div className="border-t border-gray-700 bg-gray-850 overflow-y-auto flex-shrink-0 max-h-64">
            <div className="p-2">
              {metricsLoading && (
                <div className="text-gray-400 text-sm">Loading metrics...</div>
              )}

              {metricsError && (
                <div className="bg-red-900/20 border border-red-800 rounded p-2 text-red-400 text-sm">
                  Error loading metrics.
                </div>
              )}

              {metricsData && (
                <div className="space-y-2">
                  {/* Performance Metrics - Single Row */}
                  <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-7 gap-2">
                    <MetricsCard
                      title="Active Tasks"
                      value={metricsData.metrics.tasksActive}
                      colorClass="text-yellow-400"
                      detail={`Currently running tasks`}
                    />
                    <MetricsCard
                      title="Tasks"
                      value={formatNumber(metricsData.metrics.tasksSpawned)}
                      subtitle={`${metricsData.metrics.tasksCompleted}✓ ${metricsData.metrics.tasksFailed}✗`}
                      colorClass="text-blue-400"
                      detail={`Click for task details`}
                      onClick={() => setSelectedMetric('tasks')}
                    />
                    <MetricsCard
                      title="Avg Duration"
                      value={formatDuration(metricsData.metrics.averageDurationMs)}
                      colorClass="text-amber-400"
                      detail={`Click for performance details`}
                      onClick={() => setSelectedMetric('performance')}
                    />
                    <MetricsCard
                      title="Tokens"
                      value={formatNumber(metricsData.metrics.tokenUsage.totalTokens)}
                      subtitle={`${formatNumber(metricsData.metrics.tokenUsage.inputTokens)}↑ ${formatNumber(metricsData.metrics.tokenUsage.outputTokens)}↓`}
                      colorClass="text-purple-400"
                      detail={`Click for token details`}
                      onClick={() => setSelectedMetric('tokens')}
                    />
                    <MetricsCard
                      title="API Calls"
                      value={formatNumber(metricsData.metrics.apiCalls.total)}
                      subtitle={`${metricsData.metrics.apiCalls.success}✓ ${metricsData.metrics.apiCalls.failed}✗`}
                      colorClass="text-cyan-400"
                      detail={`Click for API details`}
                      onClick={() => setSelectedMetric('api')}
                    />
                    <MetricsCard
                      title="Success Rate"
                      value={metricsData.metrics.apiCalls.total > 0
                        ? `${((metricsData.metrics.apiCalls.success / metricsData.metrics.apiCalls.total) * 100).toFixed(1)}%`
                        : '100%'}
                      colorClass="text-emerald-400"
                      detail={`API call success rate`}
                    />
                    <MetricsCard
                      title="Uptime"
                      value={metricsData.metrics.performance.uptime || '0s'}
                      colorClass="text-lime-400"
                      detail={`Click for system status`}
                      onClick={() => setSelectedMetric('system')}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Metric Detail Modal */}
        {selectedMetric && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setSelectedMetric(null)}>
            <div className="bg-gray-800 border border-gray-700 rounded-lg max-w-4xl w-full max-h-[90vh] flex flex-col" onClick={(e) => e.stopPropagation()}>
              <div className="flex items-center justify-between p-6 pb-4 border-b border-gray-700 flex-shrink-0">
                <h3 className="text-lg font-semibold">
                  {selectedMetric === 'tokens' && 'Token Usage Details'}
                  {selectedMetric === 'api' && 'API Call Details'}
                  {selectedMetric === 'performance' && 'Performance Details'}
                  {selectedMetric === 'tasks' && 'Task Overview'}
                  {selectedMetric === 'system' && 'System Status'}
                </h3>
                <button
                  onClick={() => setSelectedMetric(null)}
                  className="text-gray-400 hover:text-white text-xl"
                >
                  ✕
                </button>
              </div>
              <div className="overflow-y-auto p-6 pt-4" style={{
                scrollbarWidth: 'thin',
                scrollbarColor: '#4B5563 #1F2937'
              }}>

              {metricsData && selectedMetric === 'tokens' && (
                <div className="space-y-4">
                  {/* Overview */}
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Total Input</div>
                      <div className="text-2xl font-bold text-indigo-400">
                        {formatNumber(metricsData.metrics.tokenUsage.inputTokens)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tokenUsage.inputTokens.toLocaleString()} tokens
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Total Output</div>
                      <div className="text-2xl font-bold text-violet-400">
                        {formatNumber(metricsData.metrics.tokenUsage.outputTokens)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tokenUsage.outputTokens.toLocaleString()} tokens
                      </div>
                    </div>
                    {detailedMetrics && detailedMetrics.tasks_completed > 0 && (
                      <>
                        <div className="bg-gray-900 rounded p-4">
                          <div className="text-sm text-gray-400 mb-1">Avg Input/Task</div>
                          <div className="text-2xl font-bold text-indigo-400">
                            {formatNumber(Math.floor(detailedMetrics.total_input_tokens / detailedMetrics.tasks_completed))}
                          </div>
                          <div className="text-xs text-gray-500 mt-1">
                            {Math.floor(detailedMetrics.total_input_tokens / detailedMetrics.tasks_completed).toLocaleString()} tokens per task
                          </div>
                        </div>
                        <div className="bg-gray-900 rounded p-4">
                          <div className="text-sm text-gray-400 mb-1">Avg Output/Task</div>
                          <div className="text-2xl font-bold text-violet-400">
                            {formatNumber(Math.floor(detailedMetrics.total_output_tokens / detailedMetrics.tasks_completed))}
                          </div>
                          <div className="text-xs text-gray-500 mt-1">
                            {Math.floor(detailedMetrics.total_output_tokens / detailedMetrics.tasks_completed).toLocaleString()} tokens per task
                          </div>
                        </div>
                      </>
                    )}
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Input/Output Ratio</div>
                      <div className="text-2xl font-bold text-cyan-400">
                        {metricsData.metrics.tokenUsage.outputTokens > 0
                          ? (metricsData.metrics.tokenUsage.inputTokens / metricsData.metrics.tokenUsage.outputTokens).toFixed(1)
                          : '0.0'}:1
                      </div>
                      <div className="text-xs text-gray-500 mt-1">Input per output token</div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Cache Efficiency</div>
                      <div className="text-2xl font-bold text-green-400">
                        {metricsData.metrics.tokenUsage.inputTokens > 0 && (metricsData.metrics.tokenUsage.inputTokens / metricsData.metrics.tokenUsage.outputTokens) > 50
                          ? 'Excellent'
                          : metricsData.metrics.tokenUsage.inputTokens > 0 && (metricsData.metrics.tokenUsage.inputTokens / metricsData.metrics.tokenUsage.outputTokens) > 20
                          ? 'Good'
                          : 'Normal'}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tokenUsage.inputTokens > 0 && (metricsData.metrics.tokenUsage.inputTokens / metricsData.metrics.tokenUsage.outputTokens) > 50
                          ? 'High I/O ratio indicates likely caching'
                          : 'Based on I/O ratio'}
                      </div>
                    </div>
                  </div>

                  {/* Cost Breakdown by Provider/Model */}
                  <CostBreakdown
                    totalInputTokens={metricsData.metrics.tokenUsage.inputTokens}
                    totalOutputTokens={metricsData.metrics.tokenUsage.outputTokens}
                    providers={metricsData.metrics.providerBreakdown?.length > 0 ? metricsData.metrics.providerBreakdown.map(p => {
                      // Calculate cost based on hardcoded pricing (TODO: fetch from config)
                      const pricing = {
                        'anthropic:claude-sonnet-4-5': [3.00, 15.00],
                        'anthropic:claude-sonnet-4-5-20250929': [3.00, 15.00],
                        'anthropic:claude-haiku-4-5': [0.25, 1.25],
                        'openai:gpt-4o': [2.50, 10.00],
                        'openai:gpt-4o-mini': [0.15, 0.60],
                        'openai:gpt-5.2-mini': [0.60, 2.40],
                      };
                      const key = `${p.provider}:${p.model}`;
                      const prices = pricing[key as keyof typeof pricing] || [3.00, 15.00]; // Default to Sonnet pricing
                      const inputCost = (p.inputTokens / 1_000_000) * prices[0];
                      const outputCost = (p.outputTokens / 1_000_000) * prices[1];
                      const cost = inputCost + outputCost;

                      return {
                        ...p,
                        cost,
                        percentage: 0, // Will be calculated in CostBreakdown
                      };
                    }) : undefined}
                  />

                  {/* Per-Turn Averages */}
                  {detailedMetrics && detailedMetrics.total_turns > 0 && (
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-3 font-semibold">🔄 Per-Turn Averages</div>
                      <div className="grid grid-cols-4 gap-4">
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Total Turns</div>
                          <div className="text-xl font-bold text-blue-400">
                            {detailedMetrics.total_turns.toLocaleString()}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Avg Input/Turn</div>
                          <div className="text-xl font-bold text-indigo-400">
                            {formatNumber(detailedMetrics.avg_input_per_turn)}
                          </div>
                          <div className="text-xs text-gray-500 mt-0.5">
                            {detailedMetrics.avg_input_per_turn.toLocaleString()}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Avg Output/Turn</div>
                          <div className="text-xl font-bold text-violet-400">
                            {formatNumber(detailedMetrics.avg_output_per_turn)}
                          </div>
                          <div className="text-xs text-gray-500 mt-0.5">
                            {detailedMetrics.avg_output_per_turn.toLocaleString()}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Avg Turn Ratio</div>
                          <div className="text-xl font-bold text-cyan-400">
                            {detailedMetrics.avg_output_per_turn > 0
                              ? (detailedMetrics.avg_input_per_turn / detailedMetrics.avg_output_per_turn).toFixed(1)
                              : '0.0'}:1
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Recent Turns */}
                  {detailedMetrics && detailedMetrics.turn_token_data && detailedMetrics.turn_token_data.length > 0 && (
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-3 font-semibold">📊 Recent Turns (last {Math.min(10, detailedMetrics.turn_token_data.length)})</div>
                      <div className="overflow-x-auto">
                        <table className="w-full text-xs font-mono" style={{ tableLayout: 'fixed' }}>
                          <colgroup>
                            <col style={{ width: `${turnColumnWidths.turn}px` }} />
                            <col style={{ width: `${turnColumnWidths.duration}px` }} />
                            <col style={{ width: `${turnColumnWidths.input}px` }} />
                            <col style={{ width: `${turnColumnWidths.output}px` }} />
                            <col />
                          </colgroup>
                          <thead className="text-gray-400 border-b border-gray-700">
                            <tr>
                              <th className="text-left py-2 px-2 relative group">
                                Turn
                                <div
                                  className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                  onMouseDown={(e) => {
                                    setResizingTurnColumn('turn');
                                    setTurnResizeStartX(e.clientX);
                                    setTurnResizeStartWidth(turnColumnWidths.turn);
                                  }}
                                />
                              </th>
                              <th className="text-right py-2 px-2 relative group">
                                Duration
                                <div
                                  className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                  onMouseDown={(e) => {
                                    setResizingTurnColumn('duration');
                                    setTurnResizeStartX(e.clientX);
                                    setTurnResizeStartWidth(turnColumnWidths.duration);
                                  }}
                                />
                              </th>
                              <th className="text-right py-2 px-2 relative group">
                                Input
                                <div
                                  className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                  onMouseDown={(e) => {
                                    setResizingTurnColumn('input');
                                    setTurnResizeStartX(e.clientX);
                                    setTurnResizeStartWidth(turnColumnWidths.input);
                                  }}
                                />
                              </th>
                              <th className="text-right py-2 px-2 relative group">
                                Output
                                <div
                                  className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                  onMouseDown={(e) => {
                                    setResizingTurnColumn('output');
                                    setTurnResizeStartX(e.clientX);
                                    setTurnResizeStartWidth(turnColumnWidths.output);
                                  }}
                                />
                              </th>
                              <th className="text-left py-2 px-2">Task</th>
                            </tr>
                          </thead>
                          <tbody className="text-gray-300">
                            {detailedMetrics.turn_token_data.slice(-10).map((turn, idx) => (
                              <tr key={idx} className="border-b border-gray-800">
                                <td className="py-2 px-2 text-blue-400">{turn.Turn}</td>
                                <td className="py-2 px-2 text-right text-amber-400">{formatDuration(turn.DurationMs)}</td>
                                <td className="py-2 px-2 text-right text-indigo-400">{formatNumber(turn.InputTokens)}</td>
                                <td className="py-2 px-2 text-right text-violet-400">{formatNumber(turn.OutputTokens)}</td>
                                <td className="py-2 px-2 text-gray-400 truncate" title={turn.TaskID}>
                                  {turn.TaskID}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Recent Sessions */}
                  {detailedMetrics && detailedMetrics.task_token_usage && detailedMetrics.task_token_usage.length > 0 && (
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-3 font-semibold">📋 Recent Sessions (last {Math.min(5, detailedMetrics.task_token_usage.length)})</div>
                      <div className="space-y-3">
                        {detailedMetrics.task_token_usage.slice(-5).reverse().map((session, idx) => (
                          <div key={idx} className="bg-gray-800 rounded p-3">
                            <div className="text-xs font-mono text-gray-300 mb-2">{session.TaskID}</div>
                            <div className="grid grid-cols-3 gap-2 text-xs">
                              <div>
                                <span className="text-gray-500">Turns:</span>{' '}
                                <span className="text-blue-400 font-semibold">{session.TurnCount}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Input:</span>{' '}
                                <span className="text-indigo-400 font-semibold">{formatNumber(session.InputTokens)}</span>
                                <span className="text-gray-600 ml-1">({session.InputTokens.toLocaleString()})</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Output:</span>{' '}
                                <span className="text-violet-400 font-semibold">{formatNumber(session.OutputTokens)}</span>
                                <span className="text-gray-600 ml-1">({session.OutputTokens.toLocaleString()})</span>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {metricsData && selectedMetric === 'api' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Total API Calls</div>
                      <div className="text-2xl font-bold text-cyan-400">
                        {formatNumber(metricsData.metrics.apiCalls.total)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.apiCalls.total.toLocaleString()} calls
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Success Rate</div>
                      <div className="text-2xl font-bold text-green-400">
                        {metricsData.metrics.apiCalls.total > 0
                          ? `${((metricsData.metrics.apiCalls.success / metricsData.metrics.apiCalls.total) * 100).toFixed(1)}%`
                          : '100%'}
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Successful</div>
                      <div className="text-2xl font-bold text-green-400">
                        {formatNumber(metricsData.metrics.apiCalls.success)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.apiCalls.success.toLocaleString()} calls
                        {metricsData.metrics.apiCalls.total > 0
                          ? ` • ${((metricsData.metrics.apiCalls.success / metricsData.metrics.apiCalls.total) * 100).toFixed(1)}% of total`
                          : ' • 100% of total'}
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Failed</div>
                      <div className="text-2xl font-bold text-red-400">
                        {formatNumber(metricsData.metrics.apiCalls.failed)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.apiCalls.failed.toLocaleString()} calls
                        {metricsData.metrics.apiCalls.total > 0
                          ? ` • ${((metricsData.metrics.apiCalls.failed / metricsData.metrics.apiCalls.total) * 100).toFixed(1)}% of total`
                          : ' • 0% of total'}
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {metricsData && selectedMetric === 'performance' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-3 gap-4">
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Average Duration</div>
                      <div className="text-2xl font-bold text-amber-400">
                        {formatDuration(metricsData.metrics.averageDurationMs)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.averageDurationMs.toFixed(0)}ms per task
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Uptime</div>
                      <div className="text-2xl font-bold text-lime-400">
                        {metricsData.metrics.performance.uptime || '0s'}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">Server running time</div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Tasks Active</div>
                      <div className="text-2xl font-bold text-yellow-400">
                        {metricsData.metrics.tasksActive}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tasksCompleted} completed, {metricsData.metrics.tasksFailed} failed
                      </div>
                    </div>
                  </div>
                  {metricsData.metrics.averageTokensPerTask > 0 && (
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-2">Token Efficiency</div>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Avg Tokens/Task</div>
                          <div className="text-xl font-bold text-purple-400">
                            {formatNumber(metricsData.metrics.averageTokensPerTask)}
                          </div>
                          <div className="text-xs text-gray-500 mt-1">
                            {metricsData.metrics.averageTokensPerTask.toLocaleString()} tokens
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Total Tasks</div>
                          <div className="text-xl font-bold text-blue-400">
                            {metricsData.metrics.tasksSpawned}
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {metricsData && selectedMetric === 'tasks' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Total Spawned</div>
                      <div className="text-2xl font-bold text-blue-400">
                        {formatNumber(metricsData.metrics.tasksSpawned)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tasksSpawned.toLocaleString()} tasks created
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Success Rate</div>
                      <div className="text-2xl font-bold text-green-400">
                        {metricsData.metrics.tasksSpawned > 0
                          ? `${((metricsData.metrics.tasksCompleted / metricsData.metrics.tasksSpawned) * 100).toFixed(1)}%`
                          : '0%'}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.tasksCompleted} completed / {metricsData.metrics.tasksFailed} failed
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Currently Active</div>
                      <div className="text-2xl font-bold text-yellow-400">
                        {formatNumber(metricsData.metrics.tasksActive)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">Tasks in progress</div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Avg Task Duration</div>
                      <div className="text-2xl font-bold text-amber-400">
                        {formatDuration(metricsData.metrics.averageDurationMs)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.averageDurationMs.toFixed(0)}ms average
                      </div>
                    </div>
                  </div>
                  <div className="bg-gray-900 rounded p-4">
                    <div className="text-sm text-gray-400 mb-2">Task Efficiency</div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="text-xs text-gray-500 mb-1">Tokens per Task</div>
                        <div className="text-xl font-bold text-purple-400">
                          {formatNumber(metricsData.metrics.averageTokensPerTask)}
                        </div>
                        <div className="text-xs text-gray-500 mt-1">Average consumption</div>
                      </div>
                      <div>
                        <div className="text-xs text-gray-500 mb-1">Input/Output Ratio</div>
                        <div className="text-xl font-bold text-cyan-400">
                          {metricsData.metrics.tokenUsage.outputTokens > 0
                            ? (metricsData.metrics.tokenUsage.inputTokens / metricsData.metrics.tokenUsage.outputTokens).toFixed(1)
                            : '0.0'}:1
                        </div>
                        <div className="text-xs text-gray-500 mt-1">Input to output ratio</div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {metricsData && selectedMetric === 'system' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">Server Uptime</div>
                      <div className="text-2xl font-bold text-lime-400">
                        {metricsData.metrics.performance.uptime || '0s'}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">Time since server started</div>
                    </div>
                    <div className="bg-gray-900 rounded p-4">
                      <div className="text-sm text-gray-400 mb-1">API Success Rate</div>
                      <div className="text-2xl font-bold text-green-400">
                        {metricsData.metrics.apiCalls.total > 0
                          ? `${((metricsData.metrics.apiCalls.success / metricsData.metrics.apiCalls.total) * 100).toFixed(1)}%`
                          : '100%'}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {metricsData.metrics.apiCalls.success} / {metricsData.metrics.apiCalls.total} calls
                      </div>
                    </div>
                    <div className="bg-gray-900 rounded p-4 col-span-2">
                      <div className="text-sm text-gray-400 mb-2">Resource Usage</div>
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Total Token Throughput</div>
                          <div className="text-xl font-bold text-purple-400">
                            {formatNumber(metricsData.metrics.tokenUsage.totalTokens)}
                          </div>
                          <div className="text-xs text-gray-500 mt-1">
                            {metricsData.metrics.tokenUsage.totalTokens.toLocaleString()} tokens processed
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 mb-1">Average Task Cost</div>
                          <div className="text-xl font-bold text-amber-400">
                            {formatNumber(metricsData.metrics.averageTokensPerTask)}
                          </div>
                          <div className="text-xs text-gray-500 mt-1">Tokens per completed task</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}
              </div>
            </div>
          </div>
        )}

        {/* Right: Chat Panel */}
        {/* Resize handle for chat panel */}
        <div
          className="w-1 bg-gray-700 hover:bg-blue-500 cursor-col-resize flex-shrink-0 transition-colors"
          onMouseDown={(e) => {
            setResizeStartX(e.clientX);
            setResizeStartWidth(chatWidth);
            setIsResizingChat(true);
          }}
          title="Drag to resize"
        />
        <div style={{ width: `${chatWidth}px` }} className="flex-shrink-0">
          <ChatPanel />
        </div>
      </div>

      {/* Confirmation Modal */}
      {confirmModal.show && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4 border border-gray-700">
            <h3 className="text-lg font-semibold text-white mb-3">{confirmModal.title}</h3>
            <p className="text-gray-300 text-sm mb-6 whitespace-pre-line">{confirmModal.message}</p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setConfirmModal({ show: false, title: '', message: '', onConfirm: () => {} })}
                className="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={confirmModal.onConfirm}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}

      {alertModal.show && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4 border border-gray-700">
            <h3 className="text-lg font-semibold text-white mb-3">{alertModal.title}</h3>
            <p className="text-gray-300 text-sm mb-6 whitespace-pre-line">{alertModal.message}</p>
            <div className="flex justify-end">
              <button
                onClick={() => setAlertModal({ show: false, title: '', message: '' })}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors"
              >
                OK
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
