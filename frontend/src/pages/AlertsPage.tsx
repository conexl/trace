import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertTriangle, Bell, Info, Server, ShieldAlert, XCircle } from 'lucide-react';
import { useAlerts } from '@/hooks/useAlerts';
import { useAuth } from '@/lib/auth';
import { Card } from '@/components/ui/Card';
import { FaultToleranceOverlay } from '@/components/FaultToleranceOverlay';
import { cn } from '@/lib/utils';
import { EmptyState } from '@/components/EmptyState';
import { PageHeader } from '@/components/PageHeader';
import { StatusBadge } from '@/components/StatusBadge';

function severityIcon(severity: string) {
  switch (severity.toLowerCase()) {
    case 'critical':
    case 'error':
      return <XCircle className="h-4 w-4 text-red-400" />;
    case 'warning':
      return <AlertTriangle className="h-4 w-4 text-amber-500" />;
    case 'info':
      return <Info className="h-4 w-4 text-accent" />;
    default:
      return <ShieldAlert className="h-4 w-4 text-muted" />;
  }
}

function severityClass(severity: string) {
  switch (severity.toLowerCase()) {
    case 'critical':
    case 'error':
      return 'border-red-400/20 bg-red-400/5';
    case 'warning':
      return 'border-amber-500/20 bg-amber-500/5';
    case 'info':
      return 'border-accent/20 bg-accent/5';
    default:
      return 'border-border bg-canvas/40';
  }
}

type SeverityFilter = 'all' | 'critical' | 'warning' | 'info';

export function AlertsPage() {
  const { isAuthenticated } = useAuth();
  const { data: alerts, loading, error, connected, reconnectIn } = useAlerts();
  const [filter, setFilter] = React.useState<SeverityFilter>('all');

  const filtered = React.useMemo(() => {
    if (!alerts) return [];
    if (filter === 'all') return alerts;
    return alerts.filter((a) => a.severity.toLowerCase() === filter);
  }, [alerts, filter]);

  if (!isAuthenticated) {
    return (
      <main className="flex flex-1 flex-col items-center justify-center px-6 text-muted">
        <Bell className="mb-4 h-10 w-10 text-muted/30" />
        <p className="text-sm">Log in to view alerts.</p>
      </main>
    );
  }

  if (loading && !alerts) {
    return (
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
      </main>
    );
  }

  return (
    <FaultToleranceOverlay connected={connected} reconnectIn={reconnectIn} error={error}>
      <main className="page-shell flex flex-1 flex-col px-4 py-6 sm:px-6">
        <PageHeader
          title="Alerts"
          description={`${alerts?.length ?? 0} event${alerts?.length === 1 ? '' : 's'} captured by connected agents.`}
          eyebrow={connected ? 'Live stream connected' : 'Reconnecting stream'}
          actions={<div className="flex items-center gap-1 rounded-lg border border-border bg-white/[0.025] p-1">
            {(['all', 'critical', 'warning', 'info'] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={cn(
                  'rounded-md px-3 py-1 text-[10px] font-medium uppercase transition-colors',
                  filter === f ? 'bg-accent/10 text-accent' : 'text-muted hover:text-active'
                )}
              >
                {f}
              </button>
            ))}
          </div>}
        />

        <div className="space-y-3 pt-5">
          <AnimatePresence mode="popLayout">
            {filtered.length === 0 && (
              <motion.div
                key="empty"
                initial={{ opacity: 0, scale: 0.96 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.96 }}
              >
                <EmptyState icon={Bell} title="No matching alerts" description="Try a different severity filter, or wait for new agent events." />
              </motion.div>
            )}

            {filtered.map((alert) => (
              <motion.div
                key={alert.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.96 }}
                transition={{ duration: 0.25 }}
              >
                <Card className={cn('border p-4 hover:bg-white/[0.025]', severityClass(alert.severity))}>
                  <div className="flex items-start gap-3">
                    <div className="mt-0.5">{severityIcon(alert.severity)}</div>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-sm font-medium text-active">{alert.type}</span>
                        <StatusBadge status={alert.severity} />
                        {alert.subject && (
                          <span className="truncate text-[10px] text-muted">· {alert.subject}</span>
                        )}
                      </div>
                      <p className="mt-1 text-sm text-active">{alert.message}</p>
                      <div className="mt-2 flex flex-wrap items-center gap-3 text-[10px] text-muted">
                        <span className="flex items-center gap-1">
                          <Server className="h-3 w-3" />
                          {alert.server_id.slice(0, 8)}…
                        </span>
                        <span>{new Date(alert.created_at).toLocaleString()}</span>
                        {alert.action && <span>action: {alert.action}</span>}
                        {alert.exit_code !== undefined && <span>exit: {alert.exit_code}</span>}
                      </div>
                    </div>
                  </div>
                </Card>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </main>
    </FaultToleranceOverlay>
  );
}
