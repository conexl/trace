import * as React from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip as ReTooltip,
  XAxis,
  YAxis,
} from 'recharts';
import {
  X,
  Globe,
  Search,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  RotateCw,
  Download,
  Server,
  Plus,
  PanelLeftClose,
  PanelLeftOpen,
  AlertOctagon,
  ChevronLeft,
} from 'lucide-react';
import type { DNSResult } from '@/lib/types';
import { useToast } from '@/components/ToastProvider';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { cn } from '@/lib/utils';

interface DnsManagementDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  dns: DNSResult[];
}

type StatusFilter = 'all' | 'ok' | 'error' | 'slow';

export function statusOf(d: DNSResult): StatusFilter {
  if (d.status) return d.status;
  if (d.error) return 'error';
  if (d.matches_public_ip) return 'ok';
  return 'error';
}

function statusBadge(status: StatusFilter) {
  if (status === 'ok')
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-accent/10 px-2 py-0.5 text-[10px] font-medium text-accent">
        <CheckCircle2 className="h-3 w-3" />
        OK
      </span>
    );
  if (status === 'slow')
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-500">
        <AlertTriangle className="h-3 w-3" />
        Slow
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-red-500/10 px-2 py-0.5 text-[10px] font-medium text-red-400">
      <AlertCircle className="h-3 w-3" />
      Error
    </span>
  );
}

export function DnsManagementDrawer({ open, onOpenChange, dns }: DnsManagementDrawerProps) {
  const { success } = useToast();
  const [query, setQuery] = React.useState('');
  const [statusFilter, setStatusFilter] = React.useState<StatusFilter>('all');
  const [selectedGroup, setSelectedGroup] = React.useState<string>('all');
  const [activeDomain, setActiveDomain] = React.useState<DNSResult | null>(null);
  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [records, setRecords] = React.useState<DNSResult[]>([]);
  const [customGroups, setCustomGroups] = React.useState<Set<string>>(new Set());
  const [sidebarOpen, setSidebarOpen] = React.useState(true);
  const [showAddForm, setShowAddForm] = React.useState(false);
  const [mobileDetailsOpen, setMobileDetailsOpen] = React.useState(false);
  const [groupToDelete, setGroupToDelete] = React.useState<string | null>(null);

  const [newDomain, setNewDomain] = React.useState('');
  const [newName, setNewName] = React.useState('');
  const [newGroup, setNewGroup] = React.useState('Default');
  const [newGroupInput, setNewGroupInput] = React.useState('');
  const [isNewGroup, setIsNewGroup] = React.useState(false);
  const [groupDraft, setGroupDraft] = React.useState('');

  React.useEffect(() => {
    if (open) {
      setQuery('');
      setStatusFilter('all');
      setSelectedGroup('all');
      setSelected(new Set());
      setRecords(dns);
      setCustomGroups(new Set());
      setSidebarOpen(true);
      setShowAddForm(false);
      setNewDomain('');
      setNewName('');
      setNewGroup('Default');
      setNewGroupInput('');
      setIsNewGroup(false);
      setGroupDraft('');
      setMobileDetailsOpen(false);
      setActiveDomain(dns[0] ?? null);
    }
  }, [open, dns]);

  const allGroups = React.useMemo(() => {
    const set = new Set<string>();
    records.forEach((d) => set.add(d.group || 'Default'));
    customGroups.forEach((g) => set.add(g));
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [records, customGroups]);

  const groupCounts = React.useMemo(() => {
    const map = new Map<string, number>();
    records.forEach((d) => {
      const g = d.group || 'Default';
      map.set(g, (map.get(g) || 0) + 1);
    });
    return map;
  }, [records]);

  const filtered = React.useMemo(() => {
    const q = query.toLowerCase();
    return records.filter((d) => {
      const group = d.group || 'Default';
      const matchesGroup = selectedGroup === 'all' || group === selectedGroup;
      const matchesSearch =
        d.domain.toLowerCase().includes(q) || d.name.toLowerCase().includes(q);
      const s = statusOf(d);
      const matchesStatus = statusFilter === 'all' || s === statusFilter;
      return matchesGroup && matchesSearch && matchesStatus;
    });
  }, [records, selectedGroup, query, statusFilter]);

  const toggleSelect = (domain: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(domain)) next.delete(domain);
      else next.add(domain);
      return next;
    });
  };

  const toggleAll = () => {
    if (selected.size === filtered.length && filtered.length > 0) {
      setSelected(new Set());
    } else {
      setSelected(new Set(filtered.map((d) => d.domain)));
    }
  };

  const handleAddDomain = (e: React.FormEvent) => {
    e.preventDefault();
    const domain = newDomain.trim();
    if (!domain || records.some((d) => d.domain === domain)) return;
    const group = isNewGroup
      ? newGroupInput.trim() || 'Default'
      : newGroup;
    const record: DNSResult = {
      name: newName.trim() || domain,
      domain,
      records: [],
      matches_public_ip: false,
      group,
      latency_ms: 0,
      status: 'ok',
      history: [
        { ts: Date.now(), latency_ms: 0, ok: true },
      ],
    };
    setRecords((prev) => [...prev, record]);
    setActiveDomain(record);
    setNewDomain('');
    setNewName('');
    setNewGroup('Default');
    setNewGroupInput('');
    setIsNewGroup(false);
    setShowAddForm(false);
    success(`Domain ${record.domain} added`);
    if (group !== 'Default') {
      setCustomGroups((prev) => {
        const next = new Set(prev);
        next.add(group);
        return next;
      });
    }
  };

  const handleAddGroup = (e: React.FormEvent) => {
    e.preventDefault();
    const name = groupDraft.trim();
    if (!name) return;
    setCustomGroups((prev) => {
      const next = new Set(prev);
      next.add(name);
      return next;
    });
    setGroupDraft('');
    success(`Group "${name}" created`);
  };

  const handleDeleteGroup = (group: string) => {
    setRecords((prev) =>
      prev.map((r) => (r.group === group ? { ...r, group: undefined } : r))
    );
    setCustomGroups((prev) => {
      const next = new Set(prev);
      next.delete(group);
      return next;
    });
    if (selectedGroup === group) setSelectedGroup('all');
    success(`Group "${group}" deleted`);
  };

  const okCount = records.filter((d) => statusOf(d) === 'ok').length;

  function DomainDetails({ onBack }: { onBack?: () => void }) {
    if (!activeDomain) {
      return (
        <div className="flex h-full flex-col items-center justify-center p-6 text-center text-sm text-muted">
          <Server className="mb-2 h-6 w-6 text-muted/40" />
          Select a domain to view details.
        </div>
      );
    }

    return (
      <div className="flex h-full flex-col">
        <div className="border-b border-border p-4">
          {onBack && (
            <button
              onClick={onBack}
              className="mb-2 flex items-center gap-1 text-[10px] text-muted transition-colors hover:text-active"
            >
              <ChevronLeft className="h-3 w-3" />
              Back to domains
            </button>
          )}
          <div className="mb-2 text-[10px] font-mono uppercase text-muted">Domain details</div>
          <h3 className="break-all text-sm font-medium text-active">{activeDomain.domain}</h3>
          <div className="mt-2 flex items-center gap-2">
            {statusBadge(statusOf(activeDomain))}
            <span className="text-[10px] text-muted">{activeDomain.group || 'Default'}</span>
          </div>
        </div>

        <div className="flex-1 overflow-auto p-4">
          <div className="mb-4 space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted">Current latency</span>
              <span className="font-mono text-active">
                {activeDomain.latency_ms != null ? `${activeDomain.latency_ms} ms` : '—'}
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted">Public IP match</span>
              <span
                className={cn(
                  'font-mono',
                  activeDomain.matches_public_ip ? 'text-accent' : 'text-red-400'
                )}
              >
                {activeDomain.matches_public_ip ? 'yes' : 'no'}
              </span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted">Records</span>
              <span className="font-mono text-active">{activeDomain.records.length}</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="flex items-center gap-1.5 text-muted">
                <AlertOctagon className="h-3 w-3 text-amber-500" />
                Critical alerts
              </span>
              <button
                type="button"
                onClick={() => {
                  setRecords((prev) =>
                    prev.map((r) =>
                      r.domain === activeDomain.domain ? { ...r, critical: !r.critical } : r
                    )
                  );
                  setActiveDomain((prev) => (prev ? { ...prev, critical: !prev.critical } : prev));
                }}
                className={cn(
                  'relative h-5 w-9 rounded-full transition-colors',
                  activeDomain.critical
                    ? 'bg-accent'
                    : 'bg-surface-elevated border border-border'
                )}
              >
                <span
                  className={cn(
                    'absolute top-0.5 h-4 w-4 rounded-full bg-active transition-transform',
                    activeDomain.critical ? 'left-[18px]' : 'left-0.5'
                  )}
                />
              </button>
            </div>
          </div>

          {activeDomain.history && activeDomain.history.length > 0 && (
            <div className="mb-4">
              <div className="mb-2 flex items-center justify-between text-[10px] font-mono uppercase text-muted">
                <span>Latency (24h)</span>
                <span className="normal-case text-muted/60">ms</span>
              </div>
              <div className="h-40 rounded-lg border border-border bg-canvas p-2">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={activeDomain.history}>
                    <defs>
                      <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor="#00F576" stopOpacity={0.25} />
                        <stop offset="100%" stopColor="#00F576" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#1A1F2C" />
                    <XAxis
                      dataKey="ts"
                      tickFormatter={(v) =>
                        new Date(v).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                      }
                      stroke="#64748B"
                      tick={{ fill: '#64748B', fontSize: 9 }}
                      axisLine={false}
                      tickLine={false}
                      minTickGap={24}
                    />
                    <YAxis
                      stroke="#64748B"
                      tick={{ fill: '#64748B', fontSize: 9 }}
                      axisLine={false}
                      tickLine={false}
                      width={30}
                    />
                    <ReTooltip
                      contentStyle={{
                        background: '#0B0D13',
                        border: '1px solid #1A1F2C',
                        borderRadius: 8,
                        fontSize: 11,
                      }}
                      itemStyle={{ color: '#F8FAFC' }}
                      labelFormatter={(v) => new Date(v).toLocaleString()}
                      formatter={(value) => [`${value} ms`, 'latency']}
                    />
                    <Area
                      type="monotone"
                      dataKey="latency_ms"
                      stroke="#00F576"
                      strokeWidth={2}
                      fill="url(#latencyGradient)"
                      isAnimationActive={false}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          <div>
            <div className="mb-2 text-[10px] font-mono uppercase text-muted">Resolution history</div>
            <div className="space-y-1.5">
              {(activeDomain.history || [])
                .slice(-8)
                .reverse()
                .map((point, idx) => (
                  <div
                    key={idx}
                    className="flex items-center justify-between rounded-md border border-border bg-canvas px-2 py-1.5"
                  >
                    <span className="text-[10px] text-muted">
                      {new Date(point.ts).toLocaleTimeString()}
                    </span>
                    <span
                      className={cn(
                        'font-mono text-[10px]',
                        point.ok ? 'text-accent' : 'text-red-400'
                      )}
                    >
                      {point.latency_ms} ms
                    </span>
                  </div>
                ))}
              {(!activeDomain.history || activeDomain.history.length === 0) && (
                <div className="flex flex-col items-center justify-center gap-2 py-6 text-center text-xs text-muted">
                  <Server className="h-5 w-5 text-muted/30" />
                  No history available for this domain.
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <DialogPrimitive.Portal forceMount>
            <DialogPrimitive.Overlay asChild>
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="fixed inset-0 z-50 bg-canvas/70 backdrop-blur-sm"
                onClick={() => onOpenChange(false)}
              />
            </DialogPrimitive.Overlay>
            <DialogPrimitive.Content asChild>
              <motion.div
                initial={{ opacity: 0, y: 24, scale: 0.98 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: 24, scale: 0.98 }}
                transition={{ type: 'spring', damping: 28, stiffness: 260 }}
                className="fixed inset-4 z-50 flex flex-col overflow-hidden rounded-2xl border border-border bg-surface shadow-2xl md:inset-6"
              >
                <header className="flex items-center justify-between border-b border-border px-5 py-4">
                  <div className="flex items-center gap-3">
                    <motion.button
                      onClick={() => setSidebarOpen((s) => !s)}
                      whileTap={{ scale: 0.95 }}
                      className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-canvas text-muted transition-colors hover:border-accent hover:text-accent"
                      title={sidebarOpen ? 'Hide groups' : 'Show groups'}
                    >
                      {sidebarOpen ? (
                        <PanelLeftClose className="h-4 w-4" />
                      ) : (
                        <PanelLeftOpen className="h-4 w-4" />
                      )}
                    </motion.button>
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-canvas">
                      <Globe className="h-4 w-4 text-accent" />
                    </div>
                    <div>
                      <DialogPrimitive.Title className="text-sm font-medium tracking-tight text-active">
                        DNS Management Hub
                      </DialogPrimitive.Title>
                      <p className="text-[11px] text-muted">
                        {okCount}/{records.length} domains OK
                      </p>
                    </div>
                  </div>
                  <DialogPrimitive.Close className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted transition-colors hover:border-accent hover:text-accent">
                    <X className="h-4 w-4" />
                  </DialogPrimitive.Close>
                </header>

                <div className="flex flex-1 overflow-hidden">
                  {/* Side navigation */}
                  <motion.aside
                    initial={false}
                    animate={{
                      width: sidebarOpen ? 176 : 0,
                      opacity: sidebarOpen ? 1 : 0,
                    }}
                    transition={{ duration: 0.25, ease: 'easeInOut' }}
                    className="flex flex-col overflow-hidden border-r border-border bg-canvas/30"
                  >
                    <div className="flex w-44 flex-col">
                      <div className="flex items-center justify-between p-3 pb-2">
                        <span className="px-2 text-[10px] font-mono uppercase text-muted">Groups</span>
                      </div>
                      <div className="flex-1 overflow-auto px-3 pb-2">
                        <button
                          onClick={() => setSelectedGroup('all')}
                          className={cn(
                            'flex w-full items-center justify-between rounded-md px-2.5 py-1.5 text-left text-xs transition-colors',
                            selectedGroup === 'all'
                              ? 'bg-accent/10 text-accent'
                              : 'text-active hover:bg-surface-elevated'
                          )}
                        >
                          <span>All groups</span>
                          <span className="font-mono text-[10px] text-muted">{records.length}</span>
                        </button>
                        {allGroups.map((group) => (
                          <div
                            key={group}
                            className={cn(
                              'group flex items-center justify-between rounded-md px-2.5 py-1.5 text-left text-xs transition-colors',
                              selectedGroup === group
                                ? 'bg-accent/10 text-accent'
                                : 'text-active hover:bg-surface-elevated'
                            )}
                          >
                            <button
                              onClick={() => setSelectedGroup(group)}
                              className="flex flex-1 items-center justify-between truncate"
                            >
                              <span className="truncate">{group}</span>
                              <span className="font-mono text-[10px] text-muted">
                                {groupCounts.get(group) || 0}
                              </span>
                            </button>
                            {group !== 'Default' && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setGroupToDelete(group);
                                }}
                                className="ml-1.5 text-muted opacity-0 transition-opacity hover:text-red-400 group-hover:opacity-100"
                                title="Delete group"
                              >
                                <X className="h-3 w-3" />
                              </button>
                            )}
                          </div>
                        ))}
                      </div>

                      <div className="border-t border-border p-3">
                        <form onSubmit={handleAddGroup} className="flex items-center gap-1.5">
                          <input
                            type="text"
                            value={groupDraft}
                            onChange={(e) => setGroupDraft(e.target.value)}
                            placeholder="New group"
                            className="min-w-0 flex-1 rounded-md border border-border bg-canvas px-2 py-1 text-[10px] text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                          />
                          <button
                            type="submit"
                            disabled={!groupDraft.trim()}
                            className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-canvas text-muted transition-colors hover:border-accent hover:text-accent disabled:opacity-40"
                          >
                            <Plus className="h-3 w-3" />
                          </button>
                        </form>
                      </div>
                    </div>
                  </motion.aside>

                  {/* Center table */}
                  <section className="flex min-w-0 flex-1 flex-col">
                    <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-start sm:justify-between">
                      <div className="relative flex-1">
                        <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted" />
                        <input
                          type="text"
                          value={query}
                          onChange={(e) => setQuery(e.target.value)}
                          placeholder="Search domains…"
                          className="w-full rounded-md border border-border bg-canvas py-1.5 pl-8 pr-3 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                        />
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        <div className="flex items-center gap-1 rounded-lg border border-border bg-canvas p-0.5">
                          {(['all', 'ok', 'slow', 'error'] as StatusFilter[]).map((s) => (
                            <button
                              key={s}
                              onClick={() => setStatusFilter(s)}
                              className={cn(
                                'rounded-md px-2.5 py-1 text-[10px] font-medium uppercase transition-colors',
                                statusFilter === s
                                  ? 'bg-accent/10 text-accent'
                                  : 'text-muted hover:text-active'
                              )}
                            >
                              {s}
                            </button>
                          ))}
                        </div>
                        <button
                          onClick={() => setShowAddForm((s) => !s)}
                          className={cn(
                            'inline-flex items-center gap-1 rounded-md border px-2 py-1 text-[10px] font-medium transition-colors',
                            showAddForm
                              ? 'border-accent bg-accent/10 text-accent'
                              : 'border-border bg-canvas text-active hover:border-accent hover:text-accent'
                          )}
                        >
                          <Plus className="h-3 w-3" />
                          Add domain
                        </button>
                      </div>
                    </div>

                    <AnimatePresence initial={false}>
                      {showAddForm && (
                        <motion.div
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: 'auto', opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.2, ease: 'easeInOut' }}
                          className="overflow-hidden border-b border-border bg-canvas/20"
                        >
                          <form onSubmit={handleAddDomain} className="flex flex-col gap-2 p-4 sm:flex-row sm:items-end">
                            <div className="flex-1 space-y-1">
                              <label className="text-[10px] font-mono uppercase text-muted">Domain</label>
                              <input
                                type="text"
                                value={newDomain}
                                onChange={(e) => setNewDomain(e.target.value)}
                                placeholder="example.com"
                                required
                                className="w-full rounded-md border border-border bg-canvas px-3 py-1.5 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                              />
                            </div>
                            <div className="flex-1 space-y-1">
                              <label className="text-[10px] font-mono uppercase text-muted">Name</label>
                              <input
                                type="text"
                                value={newName}
                                onChange={(e) => setNewName(e.target.value)}
                                placeholder="Optional label"
                                className="w-full rounded-md border border-border bg-canvas px-3 py-1.5 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                              />
                            </div>
                            <div className="flex-1 space-y-1">
                              <label className="text-[10px] font-mono uppercase text-muted">Group</label>
                              {!isNewGroup ? (
                                <select
                                  value={newGroup}
                                  onChange={(e) => {
                                    if (e.target.value === '__new__') {
                                      setIsNewGroup(true);
                                      setNewGroupInput('');
                                    } else {
                                      setNewGroup(e.target.value);
                                    }
                                  }}
                                  className="w-full rounded-md border border-border bg-canvas px-3 py-1.5 text-xs text-active focus:border-border-focus focus:outline-none"
                                >
                                  <option value="Default">Default</option>
                                  {allGroups
                                    .filter((g) => g !== 'Default')
                                    .map((g) => (
                                      <option key={g} value={g}>
                                        {g}
                                      </option>
                                    ))}
                                  <option value="__new__">+ New group</option>
                                </select>
                              ) : (
                                <div className="flex items-center gap-1.5">
                                  <input
                                    type="text"
                                    value={newGroupInput}
                                    onChange={(e) => setNewGroupInput(e.target.value)}
                                    placeholder="Group name"
                                    autoFocus
                                    className="min-w-0 flex-1 rounded-md border border-border bg-canvas px-3 py-1.5 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                                  />
                                  <button
                                    type="button"
                                    onClick={() => {
                                      setIsNewGroup(false);
                                      setNewGroupInput('');
                                    }}
                                    className="text-[10px] text-muted hover:text-active"
                                  >
                                    Cancel
                                  </button>
                                </div>
                              )}
                            </div>
                            <div className="flex items-center gap-2">
                              <button
                                type="button"
                                onClick={() => setShowAddForm(false)}
                                className="rounded-md border border-border bg-canvas px-3 py-1.5 text-[10px] text-muted transition-colors hover:text-active"
                              >
                                Cancel
                              </button>
                              <button
                                type="submit"
                                disabled={!newDomain.trim()}
                                className="rounded-md border border-accent/40 bg-accent/10 px-3 py-1.5 text-[10px] text-accent transition-colors hover:bg-accent/15 disabled:opacity-40"
                              >
                                Save
                              </button>
                            </div>
                          </form>
                        </motion.div>
                      )}
                    </AnimatePresence>

                    <div className="flex items-center justify-between border-b border-border bg-canvas/20 px-4 py-2">
                      <div className="flex items-center gap-2 text-xs text-active">
                        <input
                          type="checkbox"
                          checked={filtered.length > 0 && selected.size === filtered.length}
                          onChange={toggleAll}
                          className="h-3.5 w-3.5 rounded border-border bg-canvas accent-accent"
                        />
                        <span className="text-muted">
                          {selected.size > 0 ? `${selected.size} selected` : `${filtered.length} domains`}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          disabled={selected.size === 0}
                          onClick={() => console.log('recheck', Array.from(selected))}
                          className="inline-flex items-center gap-1 rounded-md border border-border bg-canvas px-2 py-1 text-[10px] text-active transition-colors hover:border-accent hover:text-accent disabled:opacity-40"
                        >
                          <RotateCw className="h-3 w-3" />
                          Recheck
                        </button>
                        <button
                          disabled={selected.size === 0}
                          onClick={() => console.log('export', Array.from(selected))}
                          className="inline-flex items-center gap-1 rounded-md border border-border bg-canvas px-2 py-1 text-[10px] text-active transition-colors hover:border-accent hover:text-accent disabled:opacity-40"
                        >
                          <Download className="h-3 w-3" />
                          Export
                        </button>
                      </div>
                    </div>

                    <div className="flex-1 overflow-auto">
                      <table className="w-full text-left text-xs">
                        <thead className="sticky top-0 z-10 bg-surface">
                          <tr className="border-b border-border text-[10px] uppercase text-muted">
                            <th className="w-8 px-4 py-2"></th>
                            <th className="px-4 py-2">Domain</th>
                            <th className="hidden px-4 py-2 sm:table-cell">Group</th>
                            <th className="px-4 py-2">Status</th>
                            <th className="px-4 py-2 text-right">Latency</th>
                          </tr>
                        </thead>
                        <tbody>
                          {filtered.map((d) => {
                            const status = statusOf(d);
                            const isActive = activeDomain?.domain === d.domain;
                            return (
                              <tr
                                key={d.domain}
                                onClick={() => {
                                  setActiveDomain(d);
                                  setMobileDetailsOpen(true);
                                }}
                                className={cn(
                                  'cursor-pointer border-b border-border transition-colors',
                                  isActive ? 'bg-accent/5' : 'hover:bg-surface-elevated/50'
                                )}
                              >
                                <td className="px-4 py-2.5" onClick={(e) => e.stopPropagation()}>
                                  <input
                                    type="checkbox"
                                    checked={selected.has(d.domain)}
                                    onChange={() => toggleSelect(d.domain)}
                                    className="h-3.5 w-3.5 rounded border-border bg-canvas accent-accent"
                                  />
                                </td>
                                <td className="px-4 py-2.5">
                                  <div className="flex items-center gap-1.5">
                                    <span className="font-medium text-active">{d.domain}</span>
                                    {d.critical && (
                                      <span title="Critical domain">
                                        <AlertOctagon className="h-3 w-3 text-amber-500" />
                                      </span>
                                    )}
                                  </div>
                                  <div className="text-[10px] text-muted">{d.name}</div>
                                </td>
                                <td className="hidden px-4 py-2.5 text-muted sm:table-cell">
                                  {d.group || 'Default'}
                                </td>
                                <td className="px-4 py-2.5">{statusBadge(status)}</td>
                                <td className="px-4 py-2.5 text-right font-mono text-muted">
                                  {d.latency_ms != null ? `${d.latency_ms} ms` : '—'}
                                </td>
                              </tr>
                            );
                          })}
                          {filtered.length === 0 && (
                            <tr>
                              <td colSpan={5} className="py-12 text-center">
                                <div className="flex flex-col items-center justify-center gap-2 text-sm text-muted">
                                  <Search className="h-6 w-6 text-muted/30" />
                                  No domains match the filters.
                                </div>
                              </td>
                            </tr>
                          )}
                        </tbody>
                      </table>
                    </div>
                  </section>

                  {/* Details */}
                  <aside className="hidden w-72 flex-col border-l border-border bg-canvas/30 lg:flex">
                    <DomainDetails />
                  </aside>

                  {/* Mobile details slide-over */}
                  {mobileDetailsOpen && activeDomain && (
                    <motion.div
                      initial={{ x: '100%' }}
                      animate={{ x: 0 }}
                      exit={{ x: '100%' }}
                      transition={{ type: 'spring', damping: 30, stiffness: 300 }}
                      className="absolute inset-y-0 right-0 z-30 flex w-full flex-col border-l border-border bg-surface sm:w-80 lg:hidden"
                    >
                      <DomainDetails onBack={() => setMobileDetailsOpen(false)} />
                    </motion.div>
                  )}
                </div>
              </motion.div>
            </DialogPrimitive.Content>
          </DialogPrimitive.Portal>
        )}
      </AnimatePresence>

      <ConfirmationDialog
        open={groupToDelete !== null}
        onOpenChange={(open) => !open && setGroupToDelete(null)}
        title="Delete group"
        description={
          groupToDelete
            ? `Delete "${groupToDelete}"? Domains in this group will be moved to Default.`
            : ''
        }
        confirmLabel="Delete"
        variant="danger"
        onConfirm={() => groupToDelete && handleDeleteGroup(groupToDelete)}
      />
    </DialogPrimitive.Root>
  );
}
