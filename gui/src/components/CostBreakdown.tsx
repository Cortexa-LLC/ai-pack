interface ProviderCost {
  provider: string;
  model: string;
  calls: number;
  inputTokens: number;
  outputTokens: number;
  cost: number;
  percentage: number;
}

interface CostBreakdownProps {
  totalInputTokens: number;
  totalOutputTokens: number;
  providers?: ProviderCost[];
}

/**
 * Cost breakdown by provider and model
 * Shows API costs per service (OpenAI vs Anthropic)
 */
function CostBreakdown({ totalInputTokens, totalOutputTokens, providers }: CostBreakdownProps) {
  // Calculate estimated costs
  // TODO: Wire this up to real-time data from backend
  const calculateDefaultCost = () => {
    // Claude Sonnet pricing: $3 input / $15 output per 1M tokens
    const inputCost = (totalInputTokens / 1_000_000) * 3.00;
    const outputCost = (totalOutputTokens / 1_000_000) * 15.00;
    return inputCost + outputCost;
  };

  const totalCost = providers?.reduce((sum, p) => sum + p.cost, 0) || calculateDefaultCost();

  // Calculate percentages if we have provider data
  const displayProviders = providers
    ? providers.map(p => ({
        ...p,
        percentage: totalCost > 0 ? (p.cost / totalCost) * 100 : 0,
      }))
    : [
        {
          provider: 'anthropic',
          model: 'claude-sonnet-4-5',
          calls: 0,
          inputTokens: totalInputTokens,
          outputTokens: totalOutputTokens,
          cost: calculateDefaultCost(),
          percentage: 100,
        },
      ];

  const formatCost = (cost: number) => `$${cost.toFixed(4)}`;
  const formatTokens = (tokens: number) => `${(tokens / 1_000_000).toFixed(3)}M`;

  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <h3 className="text-sm font-semibold text-gray-300 mb-3 flex items-center justify-between">
        <span>💰 Cost Breakdown</span>
        <span className="text-lg font-bold text-green-400">{formatCost(totalCost)}</span>
      </h3>

      <div className="space-y-3">
        {displayProviders.map((provider, idx) => (
          <div key={idx} className="bg-gray-750 rounded p-3 border border-gray-600">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className={`text-xs px-2 py-0.5 rounded font-semibold ${
                  provider.provider === 'openai' ? 'bg-green-900 text-green-300' : 'bg-purple-900 text-purple-300'
                }`}>
                  {provider.provider === 'openai' ? '🟢 OpenAI' : '🔵 Anthropic'}
                </span>
                <span className="text-xs text-gray-400">{provider.model}</span>
              </div>
              <span className="text-sm font-bold text-gray-300">{formatCost(provider.cost)}</span>
            </div>

            <div className="grid grid-cols-3 gap-2 text-xs">
              <div>
                <div className="text-gray-500">Input</div>
                <div className="text-gray-300 font-mono">{formatTokens(provider.inputTokens)}</div>
              </div>
              <div>
                <div className="text-gray-500">Output</div>
                <div className="text-gray-300 font-mono">{formatTokens(provider.outputTokens)}</div>
              </div>
              <div>
                <div className="text-gray-500">Share</div>
                <div className="text-gray-300 font-mono">{provider.percentage.toFixed(0)}%</div>
              </div>
            </div>

            {provider.calls > 0 && (
              <div className="mt-2 text-xs text-gray-500">
                {provider.calls} API call{provider.calls !== 1 ? 's' : ''}
              </div>
            )}
          </div>
        ))}
      </div>

      {!providers && totalInputTokens === 0 && totalOutputTokens === 0 && (
        <div className="mt-3 p-2 bg-blue-900/20 border border-blue-700/50 rounded text-xs text-blue-300">
          💡 No API calls yet. Cost breakdown will appear after first request.
        </div>
      )}

      {!providers && (totalInputTokens > 0 || totalOutputTokens > 0) && (
        <div className="mt-3 p-2 bg-yellow-900/20 border border-yellow-700/50 rounded text-xs text-yellow-300">
          ⚠️ Provider breakdown not available. Showing estimated cost based on default model.
        </div>
      )}

      <div className="mt-3 pt-3 border-t border-gray-700">
        <div className="text-xs text-gray-500 space-y-1">
          <div className="flex justify-between">
            <span>Total tokens:</span>
            <span className="font-mono">{formatTokens(totalInputTokens + totalOutputTokens)}</span>
          </div>
          <div className="flex justify-between">
            <span>Average per task:</span>
            <span className="font-mono">{formatTokens((totalInputTokens + totalOutputTokens) / Math.max(1, displayProviders[0].calls || 1))}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default CostBreakdown;
