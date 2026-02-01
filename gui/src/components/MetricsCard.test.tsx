import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import MetricsCard from './MetricsCard';

describe('MetricsCard', () => {
  it('should render title and value', () => {
    render(<MetricsCard title="Tasks Completed" value={42} />);
    
    expect(screen.getByText('Tasks Completed')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });

  it('should render with subtitle', () => {
    render(<MetricsCard title="CPU Usage" value="25.5%" subtitle="Current" />);
    
    expect(screen.getByText('CPU Usage')).toBeInTheDocument();
    expect(screen.getByText('25.5%')).toBeInTheDocument();
    expect(screen.getByText('Current')).toBeInTheDocument();
  });

  it('should apply custom color classes', () => {
    render(
      <MetricsCard title="Test" value="100" colorClass="text-green-500" />
    );
    
    const valueElement = screen.getByText('100');
    expect(valueElement.className).toContain('text-green-500');
  });

  it('should handle numeric and string values', () => {
    const { rerender } = render(<MetricsCard title="Test" value={123} />);
    expect(screen.getByText('123')).toBeInTheDocument();
    
    rerender(<MetricsCard title="Test" value="abc" />);
    expect(screen.getByText('abc')).toBeInTheDocument();
  });
});
