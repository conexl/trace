import { useState } from 'react';
import { Play, Square, RotateCw } from 'lucide-react';
import { enqueueServiceAction } from '@/lib/api';
import type { ProcessSnapshot } from '@/lib/types';
import { cn, formatBytes } from '@/lib/utils';

interface ProcessListProps {
  serverId: string;
  processes: ProcessSnapshot[];
}

export function ProcessList({ serverId, processes }: ProcessListProps) {
  const [pending, setPending] = useState<Record<string, boolean>>({});

  const handleAction = async (processName: string, action: 'start' | 'stop' | 'restart') => {
    const key = `${processName}-${action}`;
    setPending((prev) => ({ ...prev, [key]: true }));
    try {
      const proc = processes.find((item) => item.name === processName);
      if (!proc?.service || !proc.remote_control) {
        throw new Error('service is not remote-controllable');
      }
      await enqueueServiceAction(serverId, proc.service, action);
    } catch {
      // Tasks are agent-allowlisted; failures are expected if the task is not configured.
    } finally {
      setTimeout(() => setPending((prev) => ({ ...prev, [key]: false })), 600);
    }
  };

  return (
    <div className="space-y-1">
      {processes.length === 0 && (
        <div className="py-8 text-center text-sm text-muted">No watched processes</div>
      )}
      {processes.map((proc) => (
        <div
          key={proc.name}
          className="flex items-center justify-between rounded-lg border border-transparent py-2.5 hover:border-border hover:bg-surface/50 px-2 -mx-2 transition-colors"
        >
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="truncate text-sm font-medium tracking-tight text-active">
                {proc.name}
              </span>
              <span
                className={cn(
                  'rounded-full px-1.5 py-0.5 text-[11px] font-mono uppercase leading-none',
                  proc.running
                    ? 'bg-accent/10 text-accent'
                    : 'bg-amber-500/10 text-amber-muted'
                )}
              >
                {proc.status || (proc.running ? 'running' : 'down')}
              </span>
            </div>
            <div className="mt-1 flex gap-3 font-mono text-xs text-muted">
              {proc.pid ? <span>pid {proc.pid}</span> : null}
              {proc.cpu_percent !== undefined ? (
                <span>cpu {proc.cpu_percent.toFixed(1)}%</span>
              ) : null}
              {proc.memory_rss ? <span>mem {formatBytes(proc.memory_rss)}</span> : null}
              {proc.error ? <span className="text-amber-muted truncate max-w-[140px]">{proc.error}</span> : null}
            </div>
          </div>

          <div className="ml-3 flex items-center gap-1.5">
            <ActionButton
              icon={Play}
              label="start"
              onClick={() => handleAction(proc.name, 'start')}
              loading={pending[`${proc.name}-start`]}
              disabled={!proc.remote_control || !proc.service}
            />
            <ActionButton
              icon={Square}
              label="stop"
              onClick={() => handleAction(proc.name, 'stop')}
              loading={pending[`${proc.name}-stop`]}
              disabled={!proc.remote_control || !proc.service}
            />
            <ActionButton
              icon={RotateCw}
              label="restart"
              onClick={() => handleAction(proc.name, 'restart')}
              loading={pending[`${proc.name}-restart`]}
              disabled={!proc.remote_control || !proc.service}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function ActionButton({
  icon: Icon,
  label,
  onClick,
  loading,
  disabled,
}: {
  icon: React.ElementType;
  label: string;
  onClick: () => void;
  loading?: boolean;
  disabled?: boolean;
}) {
  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
      disabled={loading || disabled}
      title={label}
      className={cn(
        'flex h-7 w-7 items-center justify-center rounded-md border border-border bg-canvas',
        'text-muted transition-all duration-200',
        'hover:border-accent hover:text-accent hover:bg-accent/5 hover:shadow-[0_0_10px_rgba(0,245,118,0.12)]',
        'disabled:opacity-50'
      )}
    >
      <Icon className={cn('h-4 w-4', loading && 'animate-spin')} />
    </button>
  );
}
