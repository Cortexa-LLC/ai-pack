import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import TaskList from './TaskList';
import { AgentTask } from '../hooks/useTasks';

describe('TaskList', () => {
  const mockTasks: AgentTask[] = [
    {
      taskID: 'task-1',
      role: 'engineer',
      task: 'Fix bug in authentication',
      status: 'running',
      createdAt: '2026-01-31T12:00:00Z',
      updatedAt: '2026-01-31T12:30:00Z',
      completedAt: null,
      result: null,
      error: null,
      metadata: null,
      beadsTaskID: 'bd-123',
    },
    {
      taskID: 'task-2',
      role: 'tester',
      task: 'Write integration tests',
      status: 'completed',
      createdAt: '2026-01-31T11:00:00Z',
      updatedAt: '2026-01-31T11:45:00Z',
      completedAt: '2026-01-31T11:45:00Z',
      result: 'Tests passed',
      error: null,
      metadata: null,
      beadsTaskID: 'bd-456',
    },
  ];

  it('should render task list', () => {
    render(<TaskList tasks={mockTasks} />);
    
    expect(screen.getByText('Fix bug in authentication')).toBeInTheDocument();
    expect(screen.getByText('Write integration tests')).toBeInTheDocument();
  });

  it('should display task roles', () => {
    render(<TaskList tasks={mockTasks} />);
    
    expect(screen.getByText('engineer')).toBeInTheDocument();
    expect(screen.getByText('tester')).toBeInTheDocument();
  });

  it('should show empty state when no tasks', () => {
    render(<TaskList tasks={[]} />);
    
    expect(screen.getByText(/no tasks/i)).toBeInTheDocument();
  });

  it('should display task status', () => {
    render(<TaskList tasks={mockTasks} />);
    
    expect(screen.getByText('running')).toBeInTheDocument();
    expect(screen.getByText('completed')).toBeInTheDocument();
  });
});
