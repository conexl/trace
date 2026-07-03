import * as React from 'react';
import { useState, useMemo, useEffect, useRef } from 'react';
import { ChevronDown, Search, Plus, Zap, Pin } from 'lucide-react';
import { enqueueServiceAction } from '@/lib/api';
import type { ProcessSnapshot } from '@/lib/types';
import { ServiceHealthDot } from '@/components/ServiceHealthDot';
import { ServiceActionMenu } from '@/components/ServiceActionMenu';
import type { ServicePolicy } from '@/components/PolicyModal';
import { cn, formatBytes } from '@/lib/utils';

interface ServicesPanelProps {
  serverId: string;
  processes: ProcessSnapshot[];
  localServices: string[];
  availableServices: string[];
  expanded: boolean;
  servicePolicies: Record<string, ServicePolicy>;
  onExpandedChange: (expanded: boolean) => void;
  onAddService: (name: string) => void;
  onRemoveService: (name: string) => void;
  onViewLogs: (name: string) => void;
  onPolicyChange: (name: string, policy: ServicePolicy) => void;
}

export function ServicesPanel({
  serverId,
  processes,
  localServices,
  availableServices,
  expanded,
  servicePolicies,
  onExpandedChange,
  onAddService,
  onRemoveService,
  onViewLogs,
  onPolicyChange,
}: ServicesPanelProps) {
  const [showAdd, setShowAdd] = useState(false);
  const [query, setQuery] = useState('');
  const [highlighted, setHighlighted] = useState(0);
  const [restarting, setRestarting] = useState<Record<string, boolean>>({});
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    const existing = new Set(processes.map((p) => p.name));
    return availableServices
      .filter((s) => !existing.has(s))
      .filter((s) => s.toLowerCase().includes(query.toLowerCase()))
      .slice(0, 8);
  }, [availableServices, processes, query]);

  useEffect(() => {
    setHighlighted(0);
  }, [filtered.length, query]);

  useEffect(() => {
    if (showAdd) inputRef.current?.focus();
  }, [showAdd]);

  const handleAction = async (proc: ProcessSnapshot, action: 'start' | 'stop' | 'restart') => {
    const processName = proc.name;
    if (action === 'restart') {
      setRestarting((prev) => ({ ...prev, [processName]: true }));
    }
    try {
      await enqueueServiceAction(serverId, proc.service, action);
    } catch {
      // Service actions are allowed only for agent-configured remote_control services.
    } finally {
      if (action === 'restart') {
        setTimeout(() => {
          setRestarting((prev) => ({ ...prev, [processName]: false }));
        }, 2600);
      }
    }
  };

  const handleSelect = (name: string) => {
    onAddService(name);
    setQuery('');
    setShowAdd(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlighted((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlighted((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const name = filtered[highlighted];
      if (name) handleSelect(name);
    } else if (e.key === 'Escape') {
      setShowAdd(false);
    }
  };

  const visible = expanded ? processes : processes.slice(0, 5);
  const hiddenCount = Math.max(0, processes.length - 5);

  return (
    <div className="flex h-full flex-col">
      <div className="mb-3 flex items-center justify-between text-active">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium tracking-tight">Watchdog</span>
          <span className="font-mono text-xs text-muted">
            {processes.filter((p) => p.running).length}/{processes.length}
          </span>
        </div>
        <div className="flex items-center gap-2">
          {!showAdd && (
            <button
              onClick={() => setShowAdd(true)}
              disabled={availableServices.length === 0}
              className="flex items-center gap-1 rounded-md border border-accent/40 bg-accent/10 px-2 py-1 font-mono text-[10px] text-accent transition-colors hover:bg-accent/15"
            >
              <Plus className="h-3 w-3" />
              Add
            </button>
          )}
          <button
            onClick={() => onExpandedChange(!expanded)}
            className="rounded-md border border-border bg-canvas px-2 py-1 font-mono text-[10px] text-muted transition-colors hover:border-accent hover:text-accent"
          >
            {expanded ? '[ see less ]' : hiddenCount > 0 ? `[ see all ${hiddenCount}+ ]` : '[ see all ]'}
          </button>
        </div>
      </div>

      {showAdd && (
        <div className="relative mb-3">
          <div className="flex items-center gap-2 rounded-md border border-border bg-canvas px-2">
            <Search className="h-3.5 w-3.5 text-muted" />
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Search running services…"
              className="flex-1 bg-transparent py-1.5 text-xs text-active placeholder:text-muted/50 focus:outline-none"
            />
            <ChevronDown className="h-3.5 w-3.5 text-muted" />
          </div>
          {filtered.length > 0 && (
            <div
              ref={listRef}
              className="absolute z-30 mt-1 max-h-40 w-full overflow-auto rounded-md border border-border bg-surface py-1 shadow-xl"
            >
              {filtered.map((name, idx) => (
                <button
                  key={name}
                  onClick={() => handleSelect(name)}
                  className={cn(
                    'block w-full px-3 py-1.5 text-left text-xs transition-colors',
                    idx === highlighted
                      ? 'bg-accent/10 text-accent'
                      : 'text-active hover:bg-surface-elevated'
                  )}
                >
                  {name}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="flex-1 overflow-auto pr-1">
        <div className="space-y-1">
          {visible.length === 0 && (
            <div className="py-8 text-center text-sm text-muted">No services</div>
          )}
          {visible.map((proc) => {
            const policy = servicePolicies[proc.name] ?? {
              autoRestart: false,
              pin: false,
              cpuThreshold: 80,
              memoryThreshold: 80,
            };
            const health: Parameters<typeof ServiceHealthDot>[0]['status'] = proc.running
              ? restarting[proc.name]
                ? 'restarting'
                : 'healthy'
              : policy.autoRestart
                ? 'restarting'
                : 'down';

            return (
              <div
                key={proc.name}
                className="flex items-center justify-between rounded-lg border border-transparent py-2 hover:border-border hover:bg-surface/50 px-2 -mx-2 transition-colors"
              >
                <div className="flex min-w-0 flex-1 items-center gap-3">
                  <ServiceHealthDot status={health} />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => onViewLogs(proc.name)}
                        className="truncate text-left text-sm font-medium tracking-tight text-active transition-colors hover:text-accent"
                      >
                        {proc.name}
                      </button>
                      {localServices.includes(proc.name) && (
                        <span className="rounded-full bg-surface-elevated px-1.5 py-0.5 text-[10px] text-muted">
                          local
                        </span>
                      )}
                      {proc.remote_control && proc.service && (
                        <span className="rounded-full bg-accent/10 px-1.5 py-0.5 text-[10px] text-accent">
                          remote
                        </span>
                      )}
                      <div className="flex items-center gap-1">
                        {policy.autoRestart && (
                          <span title="Auto-restart">
                            <Zap className="h-3 w-3 text-amber-500" />
                          </span>
                        )}
                        {policy.pin && (
                          <span title="Pinned">
                            <Pin className="h-3 w-3 text-accent" />
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="mt-0.5 flex gap-3 font-mono text-xs text-muted">
                      {proc.pid ? <span>pid {proc.pid}</span> : null}
                      {proc.cpu_percent !== undefined ? (
                        <span>cpu {proc.cpu_percent.toFixed(1)}%</span>
                      ) : null}
                      {proc.memory_rss ? <span>mem {formatBytes(proc.memory_rss)}</span> : null}
                      {proc.error ? (
                        <span className="text-amber-muted truncate max-w-[120px]">{proc.error}</span>
                      ) : null}
                    </div>
                  </div>
                </div>

                <div className="ml-3">
                  <ServiceActionMenu
                    serviceName={proc.name}
                    onViewLogs={() => onViewLogs(proc.name)}
                    onEditPolicy={() => onPolicyChange(proc.name, policy)}
                    onRestart={() => handleAction(proc, 'restart')}
                    onStop={() => handleAction(proc, 'stop')}
                    onRemove={() => onRemoveService(proc.name)}
                    canControl={Boolean(proc.remote_control && proc.service)}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
