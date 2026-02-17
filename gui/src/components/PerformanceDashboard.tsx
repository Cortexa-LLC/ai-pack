import React, { useState, useEffect } from 'react';
import { usePerformance } from '../hooks/usePerformance';

interface PerformanceDashboardProps {
  apiUrl: string;
}

const PerformanceDashboard: React.FC<PerformanceDashboardProps> = ({ apiUrl }) => {
  const { summary, grades, loading, error, refresh } = usePerformance(apiUrl);
  const [selectedTab, setSelectedTab] = useState<'overview' | 'roles' | 'models' | 'grades'>('overview');

  useEffect(() => {
    // Refresh every 30 seconds
    const interval = setInterval(refresh, 30000);
    return () => clearInterval(interval);
  }, [refresh]);

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

  if (!summary?.enabled) {
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
    <div className="space-y-6">
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
        {(['overview', 'roles', 'models', 'grades'] as const).map((tab) => (
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
        {selectedTab === 'overview' && <OverviewTab summary={summary} />}
        {selectedTab === 'roles' && <RolesTab summary={summary} />}
        {selectedTab === 'models' && <ModelsTab summary={summary} />}
        {selectedTab === 'grades' && <GradesTab grades={grades} />}
      </div>
    </div>
  );
};

// Overview Tab
const OverviewTab: React.FC<{ summary: any }> = ({ summary }) => {
  const costSavings = summary.cost_savings || {};
  const gradeDistribution = summary.summary?.grade_distribution || {};

  return (
    <div className="space-y-6">
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
              total={summary.summary?.total_grades || 0}
            />
          ))}
        </div>
      </div>

      {/* Model Tiers Info */}
      <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">🎯 Model Tiers</h3>
        <div className="space-y-3">
          {summary.model_tiers?.tiers?.map((tier: any) => (
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
  const roleData = summary.summary?.by_role || {};

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
  const modelData = summary.summary?.by_model || {};

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold text-white mb-4">Performance by Model</h3>
      <div className="grid gap-4">
        {Object.entries(modelData).map(([model, data]: [string, any]) => (
          <div key={model} className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-lg font-semibold text-white">{model}</h4>
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
      return b.confidence_score - a.confidence_score;
    } else {
      return b.total_attempts - a.total_attempts;
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
            key={`${grade.model_id}-${grade.role_id}-${grade.project_id}-${idx}`}
            className="bg-gray-800/50 border border-gray-700 rounded-lg p-4"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3 mb-2">
                  <GradeBadge grade={grade.grade} />
                  <span className="font-semibold text-white">{grade.model_id}</span>
                  <span className="text-sm text-gray-400">•</span>
                  <span className="text-sm text-gray-300 capitalize">{grade.role_id}</span>
                </div>

                <div className="grid grid-cols-4 gap-4 mt-3 text-sm">
                  <div>
                    <div className="text-gray-400">Success Rate</div>
                    <div className="font-semibold text-green-400">
                      {(grade.success_rate * 100).toFixed(1)}%
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400">Attempts</div>
                    <div className="font-semibold text-blue-400">{grade.total_attempts}</div>
                  </div>
                  <div>
                    <div className="text-gray-400">Confidence</div>
                    <div className="font-semibold text-purple-400">
                      {(grade.confidence_score * 100).toFixed(0)}%
                    </div>
                  </div>
                  <div>
                    <div className="text-gray-400">Avg Tokens</div>
                    <div className="font-semibold text-gray-300">{grade.average_tokens}</div>
                  </div>
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
