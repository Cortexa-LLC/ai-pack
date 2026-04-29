import type { AgentTask, Metrics, ServerHealth } from '../types';

const API_BASE = '';  // Proxied through Vite

export const api = {
  // Health check
  async health(): Promise<ServerHealth> {
    const res = await fetch(`${API_BASE}/health`);
    return res.json();
  },

  // Agent tasks
  async getTasks(): Promise<{ tasks: AgentTask[]; count: number }> {
    const res = await fetch(`${API_BASE}/a2a/tasks`);
    return res.json();
  },

  async getTaskStatus(taskId: string): Promise<AgentTask> {
    const res = await fetch(`${API_BASE}/a2a/status/${taskId}`);
    return res.json();
  },

  // Metrics
  async getMetrics(): Promise<Metrics> {
    const res = await fetch(`${API_BASE}/metrics`);
    return res.json();
  },


  // Spawn agent
  async spawnAgent(role: string, taskId: string): Promise<{ task_id: string }> {
    const res = await fetch(`${API_BASE}/a2a/execute`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        method: 'execute',
        params: { role, task: taskId },
        id: 1,
      }),
    });
    const data = await res.json();
    return data.result;
  },
};

export function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return `${(num / 1_000_000).toFixed(1)}M`;
  }
  if (num >= 1_000) {
    return `${(num / 1_000).toFixed(1)}K`;
  }
  return num.toString();
}

export function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60_000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  const minutes = Math.floor(ms / 60_000);
  const seconds = Math.floor((ms % 60_000) / 1000);
  return `${minutes}m ${seconds}s`;
}
