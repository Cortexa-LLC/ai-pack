import React, { useState, useEffect } from 'react';
import { usePerformance } from '../hooks/usePerformance';

interface PerformanceDashboardProps {
  apiUrl: string;
}

const PerformanceDashboard: React.FC<PerformanceDashboardProps> = ({ apiUrl }) => {
  const { summary, grades, loading, error, refresh } = usePerformance(apiUrl);
  const [selectedTab, setSelectedTab] = useState<'overview' | 'models' | 'grades' | 'benchmark'>('overview');
  const [benchmarkRunning, setBenchmarkRunning] = useState(false);
  const [benchmarkError, setBenchmarkError] = useState<string | null>(null);
  const [totalCost, setTotalCost] = useState<number | null>(null);

  useEffect(() => {
    // Refresh every 30 seconds
    const interval = setInterval(refresh, 30000);
    return () => clearInterval(interval);
  }, [refresh]);

  useEffect(() => {
    // Fetch daily metrics to calculate total cost
    fetch('/metrics/daily/last30')
      .then(res => res.json())
      .then(data => {
        const cost = data.reduce((sum: number, day: any) => {
          const dayCost = Object.values(day.provider_breakdown || {}).reduce(
            (daySum: number, provider: any) => daySum + (provider.cost || 0),
            0
          );
          return sum + dayCost;
        }, 0);
        setTotalCost(cost);
      })
      .catch(err => console.error('Failed to fetch daily metrics:', err));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-3 text-gray-400">Loading performance data...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-900/20 border border-red-700/50 rounded-lg">
        <p className="text-red-400">Failed to load performance data: {error}</p>
        <button
          onClick={refresh}
          className="mt-2 px-4 py-2 bg-red-700/30 hover:bg-red-700/50 rounded text-sm"
        >
          Retry
        </button>
      </div>
    );
  }

  if (!summary || summary.totalGrades === 0) {
    return (
      <div className="p-6 bg-blue-900/20 border border-blue-700/50 rounded-lg">
        <h3 className="text-lg font-semibold text-blue-300 mb-2">Performance Grading Disabled</h3>
        <p className="text-blue-200 text-sm">
          Performance grading system is not yet enabled. It will activate automatically after the first task completion.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6 px-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-white">Performance Dashboard</h2>
        <button
          onClick={refresh}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm flex items-center gap-2"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-2 border-b border-gray-700">
        {(['overview', 'models', 'grades', 'benchmark'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setSelectedTab(tab)}
            className={`px-4 py-2 font-medium capitalize transition-colors ${
              selectedTab === tab
                ? 'text-blue-400 border-b-2 border-blue-400'
                : 'text-gray-400 hover:text-gray-300'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div>
        {selectedTab === 'overview' && <OverviewTab summary={summary} totalCost={totalCost} />}
        {selectedTab === 'models' && <ModelsTab summary={summary} />}
        {selectedTab === 'grades' && <GradesTab grades={grades.filter((g: any) => g.source !== 'benchmark')} />}
        {selectedTab === 'benchmark' && (
          <BenchmarkTab
            grades={grades.filter((g: any) => g.source === 'benchmark')}
            apiUrl={apiUrl}
            running={benchmarkRunning}
            error={benchmarkError}
            onRun={async () => {
              setBenchmarkRunning(true);
              setBenchmarkError(null);
              try {
                const res = await fetch(`${apiUrl}/api/benchmark/run`, { method: 'POST' });
                if (!res.ok) {
                  const body = await res.text();
                  setBenchmarkError(`Server error ${res.status}: ${body}`);
                } else {
                  // Poll for new grades
                  setTimeout(refresh, 3000);
                }
              } catch (err: any) {
                setBenchmarkError(err.message || 'Network error');
              } finally {
                setBenchmarkRunning(false);
              }
            }}
          />
        )}
      </div>
    </div>
  );
};

// Overview Tab
const OverviewTab: React.FC<{ summary: any; totalCost: number | null }> = ({ summary, totalCost }) => {
  const costSavings = summary.costSavings || {};
  const gradeDistribution = summary.gradeDistribution || {};

  return (
    <div className="space-y-6">
      {/* Total Cost Card */}
      {totalCost !== null && (
        <div className="bg-gradient-to-br from-blue-900/30 to-blue-800/20 border border-blue-700/50 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-blue-300 mb-4">💵 Total Cost (Last 30 Days)</h3>
          <div className="flex items-baseline gap-2">
            <span className="text-5xl font-bold text-blue-400">${totalCost.toFixed(2)}</span>
            <span className="text-lg text-gray-400">USD</span>
          </div>
          <p className="text-sm text-gray-400 mt-2">Aggregated across all registered projects</p>
        </div>
      )}

      {/* Cost Savings Card */}
      <div className="bg-gradient-to-br from-green-900/30 to-green-800/20 border border-green-700/50 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-green-300 mb-4">💰 Cost Optimization</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            label="Cost Savings"
            value={`$${costSavings.savings?.toFixed(2) || '0.00'}`}
            subtext={`${costSavings.savings_percent?.toFixed(1) || '0'}% saved`}
            color="green"
          />
          <StatCard
            label="Actual Cost"
            value={`$${costSavings.actual_cost?.toFixed(2) || '0.00'}`}
            subtext={`${costSavings.total_tasks || 0} tasks`}
            color="blue"
          />
          <StatCard
            label="Baseline Cost"
            value={`$${costSavings.baseline_cost?.toFixed(2) || '0.00'}`}
            subtext="If all used Sonnet"
            color="gray"
          />
          <StatCard
            label="Avg Cost/Task"
            value={`$${costSavings.avg_cost_per_task?.toFixed(4) || '0.0000'}`}
            subtext="Per task average"
            color="purple"
          />
        </div>
      </div>

      {/* Grade Distribution */}
      <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">📊 Grade Distribution</h3>
        <div className="grid grid-cols-5 gap-4">
          {['A', 'B', 'C', 'D', 'F'].map((grade) => (
            <GradeCard
              key={grade}
              grade={grade}
              count={gradeDistribution[grade] || 0}
              total={summary.totalGrades || 0}
            />
          ))}
        </div>
      </div>

      {/* Model Tiers Info */}
      <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">🎯 Model Tiers</h3>
        <div className="space-y-3">
          {summary.modelTiers?.tiers?.map((tier: any) => (
            <div key={tier.tier} className="flex items-center justify-between p-3 bg-gray-700/30 rounded">
              <div className="flex-1">
                <div className="flex items-center gap-3">
                  <span className="font-mono font-semibold text-blue-400">Tier {tier.tier}</span>
                  <span className="font-semibold text-white">{tier.name}</span>
                  <span className="text-sm text-gray-400">{tier.cost_range}</span>
                </div>
                <p className="text-sm text-gray-400 mt-1">{tier.description}</p>
              </div>
              <div className="text-sm text-gray-500">
                {tier.models.join(', ')}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// Roles Tab
const RolesTab: React.FC<{ summary: any }> = ({ summary }) => {
  const roleData = summary.byRole || {};

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-white mb-4">Performance by Role</h3>
      <div className="grid gap-4">
        {Object.entries(roleData).map(([role, data]: [string, any]) => (
          <div key={role} className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-lg font-semibold text-white capitalize">{role}</h4>
              <div className="text-sm text-gray-400">
                {data.total_attempts} attempts
              </div>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <div className="text-sm text-gray-400">Success Rate</div>
                <div className="text-2xl font-bold text-green-400">
                  {(data.success_rate * 100).toFixed(1)}%
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-400">Successes</div>
                <div className="text-2xl font-bold text-blue-400">{data.successes}</div>
              </div>
              <div>
                <div className="text-sm text-gray-400">Failures</div>
                <div className="text-2xl font-bold text-red-400">{data.failures}</div>
              </div>
            </div>

            {/* Model usage */}
            {data.models && Object.keys(data.models).length > 0 && (
              <div className="mt-4 pt-4 border-t border-gray-700">
                <div className="text-sm text-gray-400 mb-2">Models Used</div>
                <div className="flex flex-wrap gap-2">
                  {Object.entries(data.models).map(([model, count]: [string, any]) => (
                    <span
                      key={model}
                      className="px-2 py-1 bg-gray-700/50 rounded text-xs text-gray-300"
                    >
                      {model}: {count}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

// Models Tab
const ModelsTab: React.FC<{ summary: any }> = ({ summary }) => {
  const modelData = summary.byModel || {};

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-white mb-4">Performance by Model</h3>
      <div className="grid gap-4">
        {Object.entries(modelData).map(([model, data]: [string, any]) => (
          <div key={model} className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-lg font-semibold text-white">{model}</h4>
              <div className="text-sm text-gray-400">
                {data.totalAttempts} attempts
              </div>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <div className="text-sm text-gray-400">Success Rate</div>
                <div className="text-2xl font-bold text-green-400">
                  {((data.successRate || 0) * 100).toFixed(1)}%
                </div>
              </div>
              <div>
                <div className="text-sm text-gray-400">Successes</div>
                <div className="text-2xl font-bold text-blue-400">{data.successes}</div>
              </div>
              <div>
                <div className="text-sm text-gray-400">Failures</div>
                <div className="text-2xl font-bold text-red-400">{data.failures}</div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Grades Tab
const GradesTab: React.FC<{ grades: any[] }> = ({ grades }) => {
  const [sortBy, setSortBy] = useState<'grade' | 'confidence' | 'attempts'>('grade');

  const sortedGrades = [...(grades || [])].sort((a, b) => {
    if (sortBy === 'grade') {
      const gradeOrder = { A: 0, B: 1, C: 2, D: 3, F: 4 };
      return (gradeOrder[a.grade as keyof typeof gradeOrder] || 5) -
             (gradeOrder[b.grade as keyof typeof gradeOrder] || 5);
    } else if (sortBy === 'confidence') {
      return b.confidenceScore - a.confidenceScore;
    } else {
      return b.totalAttempts - a.totalAttempts;
    }
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-white">All Performance Grades</h3>
        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as any)}
          className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm"
        >
          <option value="grade">Sort by Grade</option>
          <option value="confidence">Sort by Confidence</option>
          <option value="attempts">Sort by Attempts</option>
        </select>
      </div>

      <div className="grid gap-3">
        {sortedGrades.map((grade, idx) => (
          <div
            key={`${grade.modelID}-${grade.roleID}-${grade.projectID}-${idx}`}
            className="bg-gray-800/50 border border-gray-700 rounded-lg p-4"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  <GradeBadge grade={grade.grade} />
                  <span className="font-semibold text-white">{grade.modelID}</span>
                  <span className="text-sm text-gray-400">•</span>
                  <span className="text-sm text-gray-300 capitalize">{grade.roleID}</span>
                </div>

                <div className={`grid gap-4 mt-3 text-sm ${grade.averageTokens ? 'grid-cols-4' : 'grid-cols-3'}`}>
                  <div>
                    <div className="text-gray-400">Success Rate</div>
                    <div className="font-semibold text-green-400">
                      {(grade.successRate * 100).toFixed(1)}%
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400">Attempts</div>
                    <div className="font-semibold text-blue-400">{grade.totalAttempts}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Confidence</div>
                    <div className="font-semibold text-purple-400">
                      {(grade.confidenceScore * 100).toFixed(0)}%
                    </div>
                  </div>
                  {grade.averageTokens && (
                    <div>
                      <div className="text-gray-400">Avg Tokens</div>
                      <div className="font-semibold text-gray-300">{grade.averageTokens.toLocaleString()}</div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {sortedGrades.length === 0 && (
        <div className="text-center py-12 text-gray-400">
          No performance grades recorded yet. Grades will appear after task completions.
        </div>
      )}
    </div>
  );
};

// ─── Benchmark Tab ────────────────────────────────────────────────────────────
interface BenchmarkTabProps {
  grades: any[];
  apiUrl: string;
  running: boolean;
  error: string | null;
  onRun: () => void;
}

const BenchmarkTab: React.FC<BenchmarkTabProps> = ({ grades, running, error, onRun }) => {
  const [sortBy, setSortBy] = useState<'grade' | 'model' | 'attempts'>('grade');

  const sorted = [...grades].sort((a, b) => {
    if (sortBy === 'grade') {
      const ord: Record<string, number> = { A: 0, B: 1, C: 2, D: 3, F: 4 };
      return (ord[a.grade] ?? 5) - (ord[b.grade] ?? 5);
    }
    if (sortBy === 'model') return (a.modelID || '').localeCompare(b.modelID || '');
    return (b.totalAttempts || 0) - (a.totalAttempts || 0);
  });

  const gradeColor = (g: string) => {
    const map: Record<string, string> = { A: 'text-green-400', B: 'text-blue-400', C: 'text-yellow-400', D: 'text-orange-400', F: 'text-red-400' };
    return map[g] || 'text-gray-400';
  };

  return (
    <div className="space-y-4">
      {/* Header row */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-white">Benchmark Results</h3>
          <p className="text-sm text-gray-400">
            Grades produced by <code>scripts/benchmark-models.py</code> — kept separate from
            production grades.
          </p>
        </div>
        <button
          onClick={onRun}
          disabled={running}
          className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
            running
              ? 'bg-gray-700 text-gray-400 cursor-not-allowed'
              : 'bg-indigo-600 hover:bg-indigo-500 text-white'
          }`}
        >
          {running ? (
            <>
              <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
              </svg>
              Running…
            </>
          ) : (
            <>
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Run Benchmark
            </>
          )}
        </button>
      </div>

      {error && (
        <div className="bg-red-900/30 border border-red-700 rounded-lg p-3 text-red-300 text-sm">
          {error}
        </div>
      )}

      {running && (
        <div className="bg-indigo-900/20 border border-indigo-700 rounded-lg p-3 text-indigo-300 text-sm">
          Benchmark is running in the background. Results will appear here once complete — this may
          take a few minutes.
        </div>
      )}

      {/* Sort Controls */}
      <div className="flex gap-2">
        {(['grade', 'model', 'attempts'] as const).map((s) => (
          <button
            key={s}
            onClick={() => setSortBy(s)}
            className={`px-3 py-1 rounded text-sm transition-colors ${
              sortBy === s ? 'bg-gray-600 text-white' : 'bg-gray-800 text-gray-400 hover:text-white'
            }`}
          >
            Sort by {s}
          </button>
        ))}
      </div>

      {/* Results Table */}
      {sorted.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-400 border-b border-gray-700">
                <th className="pb-2 pr-4">Model</th>
                <th className="pb-2 pr-4">Role</th>
                <th className="pb-2 pr-4">Grade</th>
                <th className="pb-2 pr-4">Success Rate</th>
                <th className="pb-2 pr-4">Attempts</th>
                <th className="pb-2 pr-4">Avg Latency</th>
                <th className="pb-2">Last Run</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800">
              {sorted.map((g, i) => {
                const avgLatency = g.totalAttempts > 0
                  ? Math.round(g.totalExecutionTimeMs / g.totalAttempts)
                  : 0;
                return (
                  <tr key={i} className="hover:bg-gray-800/50 transition-colors">
                    <td className="py-2 pr-4 font-mono text-xs text-white">{g.modelID}</td>
                    <td className="py-2 pr-4 text-gray-300">{g.roleID || '—'}</td>
                    <td className={`py-2 pr-4 font-bold text-lg ${gradeColor(g.grade)}`}>{g.grade}</td>
                    <td className="py-2 pr-4 text-gray-300">
                      {g.successRate !== undefined ? `${Math.round(g.successRate * 100)}%` : '—'}
                    </td>
                    <td className="py-2 pr-4 text-gray-300">{g.totalAttempts}</td>
                    <td className="py-2 pr-4 text-gray-300">{avgLatency > 0 ? `${avgLatency} ms` : '—'}</td>
                    <td className="py-2 text-gray-500 text-xs">
                      {g.lastUsed ? new Date(g.lastUsed).toLocaleString() : '—'}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-center py-12 text-gray-400">
          No benchmark results yet. Click <strong>Run Benchmark</strong> to generate grades.
        </div>
      )}
    </div>
  );
};

// Helper Components
const StatCard: React.FC<{ label: string; value: string; subtext: string; color: string }> = ({
  label,
  value,
  subtext,
  color,
}) => {
  const colors = {
    green: 'text-green-400',
    blue: 'text-blue-400',
    gray: 'text-gray-400',
    purple: 'text-purple-400',
  };

  return (
    <div className="bg-gray-800/30 rounded-lg p-4">
      <div className="text-sm text-gray-400 mb-1">{label}</div>
      <div className={`text-2xl font-bold ${colors[color as keyof typeof colors]}`}>{value}</div>
      <div className="text-xs text-gray-500 mt-1">{subtext}</div>
    </div>
  );
};

const GradeCard: React.FC<{ grade: string; count: number; total: number }> = ({
  grade,
  count,
  total,
}) => {
  const percentage = total > 0 ? (count / total) * 100 : 0;

  const colors = {
    A: { bg: 'bg-green-900/30', border: 'border-green-700/50', text: 'text-green-400' },
    B: { bg: 'bg-blue-900/30', border: 'border-blue-700/50', text: 'text-blue-400' },
    C: { bg: 'bg-yellow-900/30', border: 'border-yellow-700/50', text: 'text-yellow-400' },
    D: { bg: 'bg-orange-900/30', border: 'border-orange-700/50', text: 'text-orange-400' },
    F: { bg: 'bg-red-900/30', border: 'border-red-700/50', text: 'text-red-400' },
  };

  const color = colors[grade as keyof typeof colors];

  return (
    <div className={`${color.bg} border ${color.border} rounded-lg p-4 text-center`}>
      <div className={`text-3xl font-bold ${color.text} mb-2`}>{grade}</div>
      <div className="text-2xl font-semibold text-white mb-1">{count}</div>
      <div className="text-sm text-gray-400">{percentage.toFixed(0)}%</div>
    </div>
  );
};

const GradeBadge: React.FC<{ grade: string }> = ({ grade }) => {
  const colors = {
    A: 'bg-green-900/30 text-green-400 border-green-700/50',
    B: 'bg-blue-900/30 text-blue-400 border-blue-700/50',
    C: 'bg-yellow-900/30 text-yellow-400 border-yellow-700/50',
    D: 'bg-orange-900/30 text-orange-400 border-orange-700/50',
    F: 'bg-red-900/30 text-red-400 border-red-700/50',
  };

  const color = colors[grade as keyof typeof colors] || colors.F;

  return (
    <span className={`px-3 py-1 rounded-full text-sm font-bold border ${color}`}>
      Grade {grade}
    </span>
  );
};

export default PerformanceDashboard;
