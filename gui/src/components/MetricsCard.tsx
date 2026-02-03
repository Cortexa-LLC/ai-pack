

/**
 * Props for MetricsCard component
 */
interface MetricsCardProps {
  /** Card title */
  title: string;
  /** Metric value (number or string) */
  value: number | string;
  /** Optional subtitle */
  subtitle?: string;
  /** Optional Tailwind color class for value */
  colorClass?: string;
  /** Optional click handler */
  onClick?: () => void;
  /** Optional detail text shown on hover/click */
  detail?: string;
}

/**
 * Card component for displaying a single metric
 */
function MetricsCard({ title, value, subtitle, colorClass = 'text-blue-400', onClick, detail }: MetricsCardProps) {
  const isClickable = !!onClick;

  return (
    <div
      className={`bg-gray-800 rounded-lg p-2 border border-gray-700 ${isClickable ? 'cursor-pointer hover:border-gray-600 hover:bg-gray-750 transition-colors' : ''}`}
      onClick={onClick}
      title={detail}
    >
      <h3 className="text-xs font-medium text-gray-400 mb-0.5">{title}</h3>
      <div className={`text-lg font-bold ${colorClass}`}>{value}</div>
      {subtitle && <p className="text-xs text-gray-500 mt-0.5">{subtitle}</p>}
    </div>
  );
}

export default MetricsCard;
