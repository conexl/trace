import * as React from 'react';
import { useState, useMemo, useEffect, useRef } from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Activity,
  AlertTriangle,
  ChevronDown,
  Filter,
  PanelRightOpen,
  Pin,
  Plus,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  X,
  Zap,
} from 'lucide-react';
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

type WatchdogFilter = 'all' | 'down' | 'critical' | 'control' | 'local';

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
  const [drawerQuery, setDrawerQuery] = useState('');
  const [filter, setFilter] = useState<WatchdogFilter>('all');
  const [highlighted, setHighlighted] = useState(0);
  const [restarting, setRestarting] = useState<Record<string, boolean>>({});
  const [manualName, setManualName] = useState('');
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
    setManualName('');
    setShowAdd(false);
  };

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const name = manualName.trim();
    if (name) {
      onAddService(name);
      setManualName('');
      setShowAdd(false);
    }
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

  const rows = useMemo(
    () =>
      processes
        .map((proc) => {
          const policy = servicePolicies[proc.name] ?? defaultPolicy;
          return {
            proc,
            policy,
            health: getHealth(proc, policy, restarting[proc.name]),
            local: localServices.includes(proc.name),
          };
        })
        .sort((a, b) => scoreRow(b) - scoreRow(a) || a.proc.name.localeCompare(b.proc.name)),
    [localServices, processes, restarting, servicePolicies]
  );

  const compactLimit = 4;
  const visible = rows.slice(0, compactLimit);
  const hiddenCount = Math.max(0, processes.length - compactLimit);
  const runningCount = processes.filter((p) => p.running).length;
  const downCount = rows.filter((row) => row.health === 'down').length;
  const criticalCount = rows.filter((row) => row.policy.pin).length;
  const controlCount = rows.filter((row) => row.proc.remote_control && row.proc.service).length;

  const drawerRows = useMemo(() => {
    const normalized = drawerQuery.trim().toLowerCase();
    return rows.filter(({ proc, policy, health, local }) => {
      if (filter === 'down' && health !== 'down') return false;
      if (filter === 'critical' && !policy.pin) return false;
      if (filter === 'control' && !(proc.remote_control && proc.service)) return false;
      if (filter === 'local' && !local) return false;
      if (!normalized) return true;
      return [proc.name, proc.service, proc.match, proc.status, proc.error]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(normalized));
    });
  }, [drawerQuery, filter, rows]);

  return (
    <div className="flex h-full flex-col">
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
        className="mb-3 flex items-center justify-between text-active"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium tracking-tight">Watchdog</span>
          <span className="font-mono text-xs text-muted">
            {runningCount}/{processes.length}
          </span>
          {downCount > 0 && (
            <span className="rounded-full bg-red-500/10 px-1.5 py-0.5 text-[10px] text-red-400">
              {downCount} down
            </span>
          )}
          {controlCount > 0 && (
            <span className="hidden rounded-full bg-accent/10 px-1.5 py-0.5 text-[10px] text-accent sm:inline-flex">
              {controlCount} control
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {!showAdd && (
            <button
              onClick={() => setShowAdd(true)}
              className="flex items-center gap-1 rounded-md border border-accent/40 bg-accent/10 px-2 py-1 font-mono text-[10px] text-accent transition-colors hover:bg-accent/15"
            >
              <Plus className="h-3 w-3" />
              Add
            </button>
          )}
          <button
            type="button"
            onClick={() => onExpandedChange(true)}
            className="inline-flex items-center gap-1 rounded-md border border-border bg-canvas px-2 py-1 font-mono text-[10px] text-muted transition-colors hover:border-accent hover:text-accent active:border-accent active:text-accent"
          >
            <PanelRightOpen className="h-3 w-3" />
            {hiddenCount > 0 ? `[ see all ${hiddenCount}+ ]` : '[ open ]'}
          </button>
        </div>
      </motion.div>

      <AnimatePresence initial={false}>
        {showAdd && (
          <motion.div
            initial={{ opacity: 0, y: -6, height: 0 }}
            animate={{ opacity: 1, y: 0, height: 'auto' }}
            exit={{ opacity: 0, y: -6, height: 0 }}
            transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
            className="relative mb-3 overflow-visible"
          >
            {availableServices.length > 0 ? (
              <>
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
              </>
            ) : (
              <form onSubmit={handleManualSubmit} className="flex items-center gap-2">
                <input
                  ref={inputRef}
                  type="text"
                  value={manualName}
                  onChange={(e) => setManualName(e.target.value)}
                  placeholder="Process or service name…"
                  className="flex-1 rounded-md border border-border bg-canvas px-2 py-1.5 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={!manualName.trim()}
                  className="rounded-md border border-accent/40 bg-accent/10 px-2 py-1.5 text-[10px] text-accent transition-colors hover:bg-accent/15 disabled:opacity-40"
                >
                  Save
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowAdd(false);
                    setManualName('');
                  }}
                  className="rounded-md border border-border bg-canvas px-2 py-1.5 text-[10px] text-muted transition-colors hover:text-active"
                >
                  Cancel
                </button>
              </form>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      <div className="flex-1 overflow-hidden pr-1">
        <div className="space-y-1">
          {visible.length === 0 && (
            <div className="py-8 text-center text-sm text-muted">No services</div>
          )}
          <AnimatePresence initial={false}>
            {visible.map((row, index) => (
              <ServiceRow
                key={row.proc.name}
                compact
                index={index}
                proc={row.proc}
                policy={row.policy}
                health={row.health}
                local={row.local}
                onViewLogs={() => onViewLogs(row.proc.name)}
                onEditPolicy={() => onPolicyChange(row.proc.name, row.policy)}
                onRestart={() => handleAction(row.proc, 'restart')}
                onStop={() => handleAction(row.proc, 'stop')}
                onRemove={() => onRemoveService(row.proc.name)}
              />
            ))}
          </AnimatePresence>
        </div>
      </div>

      <DialogPrimitive.Root open={expanded} onOpenChange={onExpandedChange}>
        <AnimatePresence>
          {expanded && (
            <DialogPrimitive.Portal forceMount>
              <DialogPrimitive.Overlay asChild>
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.18 }}
                  className="fixed inset-0 z-50 bg-canvas/70 backdrop-blur-sm"
                />
              </DialogPrimitive.Overlay>
              <DialogPrimitive.Content asChild>
                <motion.aside
                  initial={{ opacity: 0, x: 36, scale: 0.98 }}
                  animate={{ opacity: 1, x: 0, scale: 1 }}
                  exit={{ opacity: 0, x: 36, scale: 0.98 }}
                  transition={{ duration: 0.24, ease: [0.22, 1, 0.36, 1] }}
                  className="fixed inset-3 z-50 flex flex-col overflow-hidden rounded-2xl border border-border bg-surface shadow-2xl md:inset-y-6 md:left-auto md:right-6 md:w-[min(920px,calc(100vw-3rem))]"
                >
                  <motion.div
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.28, delay: 0.04, ease: [0.22, 1, 0.36, 1] }}
                    className="relative border-b border-border bg-canvas/35 px-4 py-4"
                  >
                    <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-accent/60 to-transparent" />
                    <div className="flex items-start justify-between gap-4">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-accent/30 bg-accent/10 text-accent">
                            <ShieldCheck className="h-4 w-4" />
                          </div>
                          <div>
                            <DialogPrimitive.Title className="text-sm font-medium tracking-tight text-active">
                              Watchdog command center
                            </DialogPrimitive.Title>
                            <DialogPrimitive.Description className="text-xs text-muted">
                              Search services, inspect policy drift, and run safe agent actions.
                            </DialogPrimitive.Description>
                          </div>
                        </div>
                      </div>
                      <DialogPrimitive.Close className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-border text-muted transition-colors hover:border-accent hover:text-accent">
                        <X className="h-4 w-4" />
                      </DialogPrimitive.Close>
                    </div>

                    <div className="mt-3 flex flex-wrap gap-2">
                      <WatchdogStat label="Running" value={`${runningCount}/${processes.length}`} tone="accent" />
                      <WatchdogStat label="Down" value={downCount} tone={downCount > 0 ? 'danger' : 'muted'} />
                      <WatchdogStat label="Pinned" value={criticalCount} tone="accent" />
                      <WatchdogStat
                        label="Control"
                        value={controlCount}
                        tone={controlCount > 0 ? 'accent' : 'muted'}
                      />
                    </div>
                  </motion.div>

                  <motion.div
                    initial={{ opacity: 0, y: -8 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.24, delay: 0.08, ease: [0.22, 1, 0.36, 1] }}
                    className="flex flex-col gap-3 border-b border-border bg-canvas/20 p-4 lg:flex-row lg:items-center"
                  >
                    <div className="flex min-w-0 flex-1 items-center gap-2 rounded-lg border border-border bg-canvas px-3">
                      <Search className="h-4 w-4 text-muted" />
                      <input
                        value={drawerQuery}
                        onChange={(event) => setDrawerQuery(event.target.value)}
                        placeholder="Search by name, unit, status or error..."
                        className="min-w-0 flex-1 bg-transparent py-2 text-sm text-active placeholder:text-muted/50 focus:outline-none"
                      />
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5">
                      <Filter className="mr-1 h-3.5 w-3.5 text-muted" />
                      {filterOptions.map((option) => (
                        <button
                          key={option.value}
                          type="button"
                          onClick={() => setFilter(option.value)}
                          className={cn(
                            'rounded-full border px-2.5 py-1 text-[10px] font-medium transition-colors',
                            filter === option.value
                              ? 'border-accent/50 bg-accent/10 text-accent'
                              : 'border-border bg-canvas text-muted hover:border-accent/50 hover:text-active'
                          )}
                        >
                          {option.label}
                        </button>
                      ))}
                    </div>
                  </motion.div>

                  <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.22, delay: 0.12 }}
                    className="flex-1 overflow-auto p-4"
                  >
                    <div className="mb-3 flex items-center justify-between gap-3 text-xs text-muted">
                      <span>
                        Showing <span className="text-active">{drawerRows.length}</span> of{' '}
                        <span className="text-active">{rows.length}</span> services
                      </span>
                      <span className="hidden items-center gap-1 font-mono text-[10px] uppercase tracking-wide sm:flex">
                        <SlidersHorizontal className="h-3 w-3" />
                        policies are saved to agent desired config
                      </span>
                    </div>

                    <div className="space-y-2">
                      {drawerRows.length === 0 ? (
                        <motion.div
                          initial={{ opacity: 0, y: 8 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{ opacity: 0, y: 8 }}
                          transition={{ duration: 0.18 }}
                          className="rounded-xl border border-dashed border-border bg-canvas/40 p-10 text-center"
                        >
                          <Activity className="mx-auto mb-3 h-6 w-6 text-muted" />
                          <p className="text-sm text-active">No services match this view</p>
                          <p className="mt-1 text-xs text-muted">Try a different filter or search term.</p>
                        </motion.div>
                      ) : (
                        <AnimatePresence initial={false}>
                          {drawerRows.map((row, index) => (
                            <ServiceRow
                              key={row.proc.name}
                              index={index}
                              proc={row.proc}
                              policy={row.policy}
                              health={row.health}
                              local={row.local}
                              onViewLogs={() => onViewLogs(row.proc.name)}
                              onEditPolicy={() => onPolicyChange(row.proc.name, row.policy)}
                              onRestart={() => handleAction(row.proc, 'restart')}
                              onStop={() => handleAction(row.proc, 'stop')}
                              onRemove={() => onRemoveService(row.proc.name)}
                            />
                          ))}
                        </AnimatePresence>
                      )}
                    </div>
                  </motion.div>
                </motion.aside>
              </DialogPrimitive.Content>
            </DialogPrimitive.Portal>
          )}
        </AnimatePresence>
      </DialogPrimitive.Root>
    </div>
  );
}

const defaultPolicy: ServicePolicy = {
  autoRestart: false,
  pin: false,
  cpuThreshold: 80,
  memoryThreshold: 80,
};

const filterOptions: Array<{ value: WatchdogFilter; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'down', label: 'Down' },
  { value: 'critical', label: 'Pinned' },
  { value: 'control', label: 'Control' },
  { value: 'local', label: 'Local' },
];

function getHealth(
  proc: ProcessSnapshot,
  policy: ServicePolicy,
  restarting?: boolean
): Parameters<typeof ServiceHealthDot>[0]['status'] {
  if (proc.running) return restarting ? 'restarting' : 'healthy';
  return policy.autoRestart ? 'restarting' : 'down';
}

function scoreRow(row: {
  proc: ProcessSnapshot;
  policy: ServicePolicy;
  health: Parameters<typeof ServiceHealthDot>[0]['status'];
  local: boolean;
}) {
  let score = 0;
  if (row.health === 'down') score += 100;
  if (row.health === 'restarting') score += 70;
  if (row.policy.pin) score += 20;
  if (row.proc.remote_control && row.proc.service) score += 5;
  if (row.local) score -= 3;
  return score;
}

function WatchdogStat({
  label,
  value,
  tone,
}: {
  label: string;
  value: number | string;
  tone: 'accent' | 'danger' | 'muted';
}) {
  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.98 }}
      transition={{ duration: 0.18, ease: [0.22, 1, 0.36, 1] }}
      className={cn(
        'inline-flex items-center gap-2 rounded-full border bg-canvas/55 px-3 py-1.5 transition-colors duration-200',
        tone === 'accent' && 'border-accent/25',
        tone === 'danger' && 'border-red-500/25 bg-red-950/10',
        tone === 'muted' && 'border-border'
      )}
    >
      <div
        className={cn(
          'font-mono text-sm leading-none',
          tone === 'accent' && 'text-accent',
          tone === 'danger' && 'text-red-400',
          tone === 'muted' && 'text-muted'
        )}
      >
        {value}
      </div>
      <div className="text-[10px] uppercase tracking-wide text-muted">{label}</div>
    </motion.div>
  );
}

function ServiceRow({
  proc,
  policy,
  health,
  local,
  compact,
  index = 0,
  onViewLogs,
  onEditPolicy,
  onRestart,
  onStop,
  onRemove,
}: {
  proc: ProcessSnapshot;
  policy: ServicePolicy;
  health: Parameters<typeof ServiceHealthDot>[0]['status'];
  local: boolean;
  compact?: boolean;
  index?: number;
  onViewLogs: () => void;
  onEditPolicy: () => void;
  onRestart: () => void;
  onStop: () => void;
  onRemove: () => void;
}) {
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: compact ? 4 : 8, scale: compact ? 1 : 0.99 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: compact ? -4 : -8, scale: compact ? 1 : 0.99 }}
      whileHover={compact ? undefined : { y: -1 }}
      transition={{
        duration: 0.22,
        delay: Math.min(index * (compact ? 0.035 : 0.025), 0.18),
        ease: [0.22, 1, 0.36, 1],
      }}
      className={cn(
        'flex items-center justify-between rounded-lg border transition-colors duration-200',
        compact
          ? '-mx-2 border-transparent px-2 py-2 hover:border-border hover:bg-surface/50'
          : 'border-border bg-canvas/35 px-3 py-3 hover:border-accent/30 hover:bg-surface-elevated/35'
      )}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <ServiceHealthDot status={health} />
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <button
              onClick={onViewLogs}
              className="truncate text-left text-sm font-medium tracking-tight text-active transition-colors hover:text-accent"
            >
              {proc.name}
            </button>
            {local && (
              <span className="rounded-full bg-surface-elevated px-1.5 py-0.5 text-[10px] text-muted">
                local
              </span>
            )}
            {proc.remote_control && proc.service && (
              <span className="rounded-full bg-accent/10 px-1.5 py-0.5 text-[10px] text-accent">
                control
              </span>
            )}
            {health === 'down' && (
              <span className="inline-flex items-center gap-1 rounded-full bg-red-500/10 px-1.5 py-0.5 text-[10px] text-red-400">
                <AlertTriangle className="h-3 w-3" />
                down
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
          <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 font-mono text-xs text-muted">
            {proc.service ? <span>unit {proc.service}</span> : null}
            {proc.pid ? <span>pid {proc.pid}</span> : null}
            {proc.last_exit_code !== undefined ? <span>exit {proc.last_exit_code}</span> : null}
            {proc.cpu_percent !== undefined ? <span>cpu {proc.cpu_percent.toFixed(1)}%</span> : null}
            {proc.memory_rss ? <span>mem {formatBytes(proc.memory_rss)}</span> : null}
            {proc.error ? <span className="truncate text-amber-muted sm:max-w-[260px]">{proc.error}</span> : null}
          </div>
        </div>
      </div>

      <div className="ml-3">
        <ServiceActionMenu
          serviceName={proc.name}
          onViewLogs={onViewLogs}
          onEditPolicy={onEditPolicy}
          onRestart={onRestart}
          onStop={onStop}
          onRemove={onRemove}
          canControl={Boolean(proc.remote_control && proc.service)}
        />
      </div>
    </motion.div>
  );
}
