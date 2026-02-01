import { useState, useEffect } from 'react';
import type { LogEntry } from '../types';

export function useLogStream(taskId: string | null) {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!taskId) {
      setLogs([]);
      setIsConnected(false);
      return;
    }

    setIsConnected(true);

    // Try to fetch initial logs from execution.log
    // For now, use SSE stream endpoint
    const eventSource = new EventSource(`/stream/${taskId}`);

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.onmessage = (event) => {
      if (isPaused) return;

      try {
        const data = JSON.parse(event.data);
        const logEntry: LogEntry = {
          timestamp: data.timestamp || new Date().toISOString(),
          type: inferLogType(data),
          message: formatLogMessage(data),
          data: data.data,
        };

        setLogs(prev => [...prev, logEntry]);
      } catch (error) {
        console.error('Failed to parse log event:', error);
      }
    };

    eventSource.onerror = () => {
      setIsConnected(false);
      eventSource.close();
    };

    return () => {
      eventSource.close();
      setIsConnected(false);
    };
  }, [taskId, isPaused]);

  const clearLogs = () => setLogs([]);

  return { logs, isPaused, setIsPaused, isConnected, clearLogs };
}

function inferLogType(data: any): LogEntry['type'] {
  const msg = data.message || '';
  if (msg.includes('Turn')) return 'turn';
  if (msg.includes('API:')) return 'api';
  if (msg.includes('Text:')) return 'text';
  if (msg.includes('Tool:')) return 'tool';
  if (msg.includes('✓') || msg.includes('success')) return 'success';
  if (msg.includes('❌') || msg.includes('error')) return 'error';
  if (msg.includes('⚠️') || msg.includes('warning')) return 'warning';
  return 'info';
}

function formatLogMessage(data: any): string {
  if (typeof data === 'string') return data;
  if (data.message) return data.message;
  if (data.data?.status) return `Status: ${data.data.status}`;
  return JSON.stringify(data);
}
