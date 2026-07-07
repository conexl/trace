import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertTriangle, CheckCircle2, Clock, Radio, Search, ShieldAlert, Sparkles, TimerReset } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { IncidentDrawer } from '@/components/IncidentDrawer';
import { useAuth } from '@/lib/auth';
import { getIncident, getIncidentMetrics, listIncidents, subscribeToEvents } from '@/lib/api';
import type { Incident, IncidentMetrics } from '@/lib/types';
import { cn } from '@/lib/utils';

type StatusFilter = 'all' | 'open' | 'resolved' | 'suppressed';
type SeverityFilter = 'all' | 'critical' | 'warning';

const statusStyles: Record<string, string> = {
  open: 'border-red-400/30 bg-red-400/10 text-red-400',
  investigating: 'border-amber-500/30 bg-amber-500/10 text-amber-400',
  resolved: 'border-accent/30 bg-accent/10 text-accent',
  suppressed: 'border-muted/30 bg-muted/10 text-muted',
};

const severityStyles: Record<string, string> = {
  critical: 'text-red-400',
  warning: 'text-amber-400',
};

function upsertIncident(items: Incident[], incident: Incident) {
  const next = [incident, ...items.filter((item) => item.id !== incident.id)];
  return next.sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime());
}

function formatRelative(timestamp: string) {
  const delta = Date.now() - new Date(timestamp).getTime();
  if (delta < 60_000) return 'just now';
  if (delta < 3_600_000) return `${Math.floor(delta / 60_000)}m ago`;
  if (delta < 86_400_000) return `${Math.floor(delta / 3_600_000)}h ago`;
  return new Date(timestamp).toLocaleDateString();
}

function formatDuration(seconds: number) {
  if (!seconds) return 'n/a';
  if (seconds < 60) return `${Math.round(seconds)}s`;
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
  return `${(seconds / 3600).toFixed(seconds < 36_000 ? 1 : 0)}h`;
}

export function IncidentsPage() {
  const { isAuthenticated } = useAuth();
  const [incidents, setIncidents] = React.useState<Incident[]>([]);
  const [metrics, setMetrics] = React.useState<IncidentMetrics | null>(null);
  const [selectedIncident, setSelectedIncident] = React.useState<Incident | null>(null);
  const [status, setStatus] = React.useState<StatusFilter>('all');
  const [severity, setSeverity] = React.useState<SeverityFilter>('all');
  const [query, setQuery] = React.useState('');
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [liveEvents, setLiveEvents] = React.useState(0);

  const refresh = React.useCallback(async () => {
    try {
      const [incidentRes, metricsRes] = await Promise.all([
        listIncidents(undefined, 100),
        getIncidentMetrics('7d'),
      ]);
      setIncidents(incidentRes.incidents);
      setMetrics(metricsRes);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load incidents');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    if (!isAuthenticated) return;
    refresh();
  }, [isAuthenticated, refresh]);

  React.useEffect(() => {
    if (!isAuthenticated) return;
    return subscribeToEvents((event) => {
      if (typeof event?.type !== 'string' || !event.type.startsWith('incident.')) return;
      const incident = event.data as Incident | undefined;
      setLiveEvents((count) => count + 1);
      if (incident?.id) {
        setIncidents((items) => upsertIncident(items, incident));
        setSelectedIncident((current) => (current?.id === incident.id ? incident : current));
        getIncidentMetrics('7d').then(setMetrics).catch(() => {});
        return;
      }
      refresh();
    });
  }, [isAuthenticated, refresh]);

  const filtered = React.useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    return incidents.filter((incident) => {
      if (status !== 'all' && incident.status !== status) return false;
      if (severity !== 'all' && incident.severity !== severity) return false;
      if (!normalizedQuery) return true;
      return [incident.title, incident.summary, incident.service_name, incident.server_id]
        .some((value) => value.toLowerCase().includes(normalizedQuery));
    });
  }, [incidents, query, severity, status]);

  const counts = React.useMemo(() => ({
    open: incidents.filter((incident) => incident.status === 'open').length,
    critical: incidents.filter((incident) => incident.severity === 'critical').length,
    resolved: incidents.filter((incident) => incident.status === 'resolved').length,
  }), [incidents]);

  const noisyServices = React.useMemo(() => {
    if (!metrics) return [];
    return Object.entries(metrics.by_service)
      .sort(([, a], [, b]) => b.frequency_per_day - a.frequency_per_day)
      .slice(0, 3);
  }, [metrics]);

  const openIncident = async (incident: Incident) => {
    setSelectedIncident(incident);
    try {
      const fresh = await getIncident(incident.id);
      setSelectedIncident(fresh);
      setIncidents((items) => upsertIncident(items, fresh));
    } catch {
      // Keep the already loaded incident; drawer still has useful context.
    }
  };

  if (!isAuthenticated) {
    return (
      <main className="flex flex-1 flex-col items-center justify-center px-6 text-muted">
        <ShieldAlert className="mb-4 h-10 w-10 text-muted/30" />
        <p className="text-sm">Log in to view incidents.</p>
      </main>
    );
  }

  return (
    <main className="flex flex-1 flex-col px-6 py-10">
      <div className="mb-6 flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <div className="mb-2 flex items-center gap-2 text-[10px] font-mono uppercase tracking-[0.25em] text-accent">
            <Radio className="h-3.5 w-3.5 animate-pulse" />
            Live incident stream · {liveEvents} event{liveEvents === 1 ? '' : 's'}
          </div>
          <h1 className="text-2xl font-medium tracking-tight text-active">Incidents</h1>
          <p className="mt-1 text-sm text-muted">
            Service failures, watchdog actions, AI analysis, and recovery timeline in one place.
          </p>
        </div>
        <div className="grid gap-2 sm:grid-cols-4">
          <Metric label="MTTR" value={formatDuration(metrics?.mttr_seconds ?? 0)} tone="green" />
          <Metric label="Frequency" value={`${(metrics?.frequency_per_day ?? 0).toFixed(1)}/day`} tone="amber" />
          <Metric label="Open" value={metrics?.open ?? counts.open} tone="red" />
          <Metric label="Critical" value={metrics?.critical ?? counts.critical} tone="amber" />
        </div>
      </div>

      <Card className="mb-4 border border-border/80 bg-surface/80 p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-xl border border-accent/20 bg-accent/10 p-2 text-accent">
              <TimerReset className="h-4 w-4" />
            </div>
            <div>
              <div className="text-sm font-medium text-active">Incident metrics · {metrics?.window ?? '7d'}</div>
              <p className="text-xs text-muted">
                MTTR is calculated from resolved incidents; frequency uses all incidents in the rolling window.
              </p>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {noisyServices.length === 0 ? (
              <span className="rounded-full border border-border px-3 py-1 text-[10px] uppercase tracking-wider text-muted">
                no service hotspots
              </span>
            ) : noisyServices.map(([service, serviceMetrics]) => (
              <span key={service} className="rounded-full border border-border bg-canvas/60 px-3 py-1 text-[10px] uppercase tracking-wider text-muted">
                {service}: {serviceMetrics.frequency_per_day.toFixed(1)}/day
              </span>
            ))}
          </div>
        </div>
      </Card>

      <Card className="mb-4 border border-border/80 bg-surface/80 p-3">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div className="relative min-w-0 flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search service, server, title..."
              className="h-10 w-full rounded-lg border border-border bg-canvas pl-9 pr-3 text-sm text-active outline-none transition-colors placeholder:text-muted focus:border-accent"
            />
          </div>
          <div className="flex flex-wrap gap-2">
            {(['all', 'open', 'resolved', 'suppressed'] as const).map((item) => (
              <FilterButton key={item} active={status === item} onClick={() => setStatus(item)}>
                {item}
              </FilterButton>
            ))}
            <div className="mx-1 hidden h-8 w-px bg-border sm:block" />
            {(['all', 'critical', 'warning'] as const).map((item) => (
              <FilterButton key={item} active={severity === item} onClick={() => setSeverity(item)}>
                {item}
              </FilterButton>
            ))}
          </div>
        </div>
      </Card>

      {error && <div className="mb-4 rounded-lg border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-300">{error}</div>}

      {loading ? (
        <div className="flex flex-1 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
        </div>
      ) : (
        <div className="space-y-3">
          <AnimatePresence mode="popLayout">
            {filtered.length === 0 && (
              <motion.div key="empty" initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
                <Card className="flex flex-col items-center justify-center py-16 text-center">
                  <CheckCircle2 className="mb-3 h-8 w-8 text-muted/30" />
                  <p className="text-sm text-muted">No incidents match the current filters.</p>
                </Card>
              </motion.div>
            )}

            {filtered.map((incident, idx) => (
              <motion.button
                key={incident.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.98 }}
                transition={{ delay: Math.min(idx * 0.025, 0.2) }}
                onClick={() => openIncident(incident)}
                className="block w-full text-left"
              >
                <Card className="group border border-border/80 bg-surface/80 p-4 transition-colors hover:border-accent/50 hover:bg-surface-elevated/70">
                  <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                    <div className="flex min-w-0 gap-3">
                      <div className={cn('mt-0.5 rounded-lg border p-2', statusStyles[incident.status] ?? 'border-border text-muted')}>
                        <AlertTriangle className="h-4 w-4" />
                      </div>
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <h2 className="truncate text-sm font-medium text-active">{incident.title}</h2>
                          <span className={cn('rounded-full border px-2 py-0.5 text-[10px] uppercase', statusStyles[incident.status] ?? 'border-border text-muted')}>
                            {incident.status}
                          </span>
                          <span className={cn('font-mono text-[10px] uppercase', severityStyles[incident.severity] ?? 'text-muted')}>
                            {incident.severity}
                          </span>
                        </div>
                        <p className="mt-1 line-clamp-2 text-sm text-muted">{incident.summary}</p>
                        <div className="mt-3 flex flex-wrap items-center gap-3 text-[10px] font-mono uppercase tracking-wider text-muted">
                          <span>service {incident.service_name}</span>
                          <span>server {incident.server_id}</span>
                          <span>{incident.timeline.length} events</span>
                        </div>
                      </div>
                    </div>
                    <div className="flex shrink-0 items-center gap-3 text-[10px] text-muted lg:text-right">
                      <Sparkles className="h-3.5 w-3.5 text-purple-400" />
                      <span>AI ready</span>
                      <Clock className="h-3.5 w-3.5" />
                      <span>{formatRelative(incident.updated_at)}</span>
                    </div>
                  </div>
                </Card>
              </motion.button>
            ))}
          </AnimatePresence>
        </div>
      )}

      <IncidentDrawer
        open={selectedIncident !== null}
        onOpenChange={(open) => !open && setSelectedIncident(null)}
        incident={selectedIncident}
        onActionExecuted={refresh}
      />
    </main>
  );
}

function FilterButton({ active, children, onClick }: { active: boolean; children: React.ReactNode; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'rounded-md px-3 py-1.5 text-[10px] font-medium uppercase transition-colors',
        active ? 'bg-accent/10 text-accent' : 'text-muted hover:text-active'
      )}
    >
      {children}
    </button>
  );
}

function Metric({ label, value, tone }: { label: string; value: React.ReactNode; tone: 'red' | 'amber' | 'green' }) {
  const tones = {
    red: 'border-red-400/20 bg-red-400/10 text-red-300',
    amber: 'border-amber-500/20 bg-amber-500/10 text-amber-300',
    green: 'border-accent/20 bg-accent/10 text-accent',
  };
  return (
    <div className={cn('min-w-28 rounded-xl border px-4 py-3', tones[tone])}>
      <div className="text-[10px] font-mono uppercase tracking-wider opacity-70">{label}</div>
      <div className="mt-1 text-2xl font-semibold">{value}</div>
    </div>
  );
}
