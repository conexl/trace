import { useMemo } from 'react';
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Area,
} from 'recharts';
import { formatDuration } from '@/lib/utils';

export interface ChartSeries {
  key: string;
  name: string;
  color?: string;
  fill?: boolean;
}

interface MetricChartProps {
  data: Record<string, number | string>[];
  series: ChartSeries[];
  yDomain?: [number, number];
  showGrid?: boolean;
  className?: string;
}

export function MetricChart({
  data,
  series,
  yDomain = [0, 100],
  showGrid = false,
  className,
}: MetricChartProps) {
  const gradientId = useMemo(() => `greenGradient-${Math.random().toString(36).slice(2, 9)}`, []);

  const firstFillSeries = series.find((s) => s.fill);

  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 8, right: 8, bottom: 8, left: 0 }}>
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#00F576" stopOpacity={0.08} />
              <stop offset="100%" stopColor="#00F576" stopOpacity={0} />
            </linearGradient>
          </defs>
          {showGrid && (
            <CartesianGrid stroke="rgba(26, 31, 44, 0.6)" vertical={false} />
          )}
          <XAxis dataKey="ts" hide />
          <YAxis domain={yDomain} hide />
          <Tooltip
            cursor={{ stroke: '#263147', strokeWidth: 1, strokeDasharray: '4 4' }}
            content={({ active, payload, label }) => {
              if (!active || !payload?.length) return null;
              return (
                <div className="rounded-md border border-border bg-surface-elevated px-3 py-2 text-xs shadow-md">
                  <div className="mb-1 text-muted">{typeof label === 'number' ? formatDuration(label) : label}</div>
                  {payload.map((p) => (
                    <div key={String(p.dataKey)} className="flex items-center gap-2 font-mono">
                      <span
                        className="inline-block h-1.5 w-1.5 rounded-full"
                        style={{ backgroundColor: p.color }}
                      />
                      <span className="text-active">{String(p.name ?? p.dataKey)}:</span>
                      <span className="text-accent">{Number(p.value).toFixed(1)}</span>
                    </div>
                  ))}
                </div>
              );
            }}
          />
          {firstFillSeries && (
            <Area
              type="monotone"
              dataKey={firstFillSeries.key}
              stroke="none"
              fill={`url(#${gradientId})`}
              isAnimationActive={false}
            />
          )}
          {series.map((s) => (
            <Line
              key={s.key}
              type="monotone"
              dataKey={s.key}
              name={s.name}
              stroke={s.color ?? '#00F576'}
              strokeWidth={1.5}
              dot={false}
              isAnimationActive={false}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
