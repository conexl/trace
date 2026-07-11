import * as React from 'react';
import { motion } from 'framer-motion';
import { Terminal, Clock, CheckCircle2, XCircle, RotateCcw } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { useToast } from '@/components/ToastProvider';
import { getAuditLogs, listTasks } from '@/lib/api';
import type { AuditLog, Task } from '@/lib/types';
import { cn } from '@/lib/utils';
import { EmptyState } from '@/components/EmptyState';
import { PageHeader } from '@/components/PageHeader';
import { StatusBadge } from '@/components/StatusBadge';

export function TasksPage() {
  const [logs, setLogs] = React.useState<AuditLog[]>([]);
  const [tasks, setTasks] = React.useState<Task[]>([]);
  const [tab, setTab] = React.useState<'audit' | 'tasks'>('tasks');
  const [loading, setLoading] = React.useState(true);
  const { error: showError } = useToast();

  const fetchData = React.useCallback(async () => {
    setLoading(true);
    try {
      if (tab === 'audit') {
        const res = await getAuditLogs();
        setLogs(Array.isArray(res.audit_logs) ? res.audit_logs : []);
      } else {
        const res = await listTasks();
        setTasks(Array.isArray(res.tasks) ? res.tasks : []);
      }
    } catch (err) {
      showError(err instanceof Error ? err.message : 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  }, [tab, showError]);

  React.useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <main className="page-shell flex flex-1 flex-col px-4 py-6 sm:px-6">
      <PageHeader
        title={tab === 'tasks' ? 'Operations' : 'Audit log'}
        description={tab === 'tasks' ? 'Queued and completed actions requested for connected nodes.' : 'A chronological record of workspace actions.'}
        eyebrow="Control plane"
        actions={<div className="flex items-center gap-1 rounded-lg border border-border bg-white/[0.025] p-1">
          <button
            onClick={() => setTab('tasks')}
            className={cn(
              'rounded-md px-3 py-1 text-xs font-medium transition-colors',
              tab === 'tasks' ? 'bg-accent/10 text-accent' : 'text-muted hover:text-active'
            )}
          >
            Tasks
          </button>
          <button
            onClick={() => setTab('audit')}
            className={cn(
              'rounded-md px-3 py-1 text-xs font-medium transition-colors',
              tab === 'audit' ? 'bg-accent/10 text-accent' : 'text-muted hover:text-active'
            )}
          >
            Audit Log
          </button>
        </div>}
      />

      {loading ? (
        <div className="flex flex-1 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
        </div>
      ) : (
        <div className="space-y-3 pt-5">
          {tab === 'audit' ? (
            logs.map((log, idx) => (
              <motion.div
                key={log.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.03 }}
              >
                <Card className="p-4 hover:bg-white/[0.025]">
                  <div className="flex items-start gap-3">
                    <div className="mt-1 rounded-md bg-surface-elevated p-2">
                      <Clock className="h-4 w-4 text-muted" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-sm font-medium text-active">{log.action}</span>
                        <span className="text-xs text-muted">on</span>
                        <span className="font-mono text-xs text-accent">{log.target}</span>
                      </div>
                      <p className="mt-1 text-sm text-muted-soft">{log.details}</p>
                      <div className="mt-2 flex items-center gap-3 text-[10px] font-mono uppercase tracking-wider text-muted">
                        <span>{log.user_email}</span>
                        <span>•</span>
                        <span>{new Date(log.timestamp).toLocaleString()}</span>
                      </div>
                    </div>
                  </div>
                </Card>
              </motion.div>
            ))
          ) : (
            tasks.map((task, idx) => (
              <motion.div
                key={task.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.03 }}
              >
                <Card className="p-4 hover:bg-white/[0.025]">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start gap-3">
                      <div className="mt-1 rounded-md bg-surface-elevated p-2">
                        {task.status === 'completed' ? (
                          <CheckCircle2 className="h-4 w-4 text-accent" />
                        ) : task.status === 'failed' ? (
                          <XCircle className="h-4 w-4 text-red-400" />
                        ) : (
                          <RotateCcw className="h-4 w-4 animate-spin text-amber-500" />
                        )}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm font-medium text-active">{task.name}</span>
                          <StatusBadge status={task.status} />
                        </div>
                        <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-[10px] font-mono text-muted">
                          <span>ID: {task.id}</span>
                          <span>Server: {task.server_id}</span>
                          {task.created_by && <span>By: {task.created_by}</span>}
                        </div>
                        {task.result && (
                          <div className="mt-3 rounded-md bg-black p-3 font-mono text-[11px] leading-relaxed">
                            {task.result.stdout && (
                              <div className="text-accent/90">
                                <span className="text-muted">stdout:</span> {task.result.stdout}
                              </div>
                            )}
                            {task.result.stderr && (
                              <div className="mt-1 text-red-400/90">
                                <span className="text-muted">stderr:</span> {task.result.stderr}
                              </div>
                            )}
                            {task.result.error && (
                              <div className="mt-1 text-red-500 font-bold">
                                <span className="text-muted">error:</span> {task.result.error}
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-[10px] font-mono text-muted">
                        {new Date(task.created_at).toLocaleString()}
                      </div>
                      {task.result?.duration_ms && (
                        <div className="mt-1 text-[10px] font-mono text-muted-soft">
                          took {task.result.duration_ms}ms
                        </div>
                      )}
                    </div>
                  </div>
                </Card>
              </motion.div>
            ))
          )}

          {((tab === 'audit' && logs.length === 0) || (tab === 'tasks' && tasks.length === 0)) && (
            <EmptyState
              icon={Terminal}
              title={tab === 'audit' ? 'No audit events yet' : 'No tasks yet'}
              description={tab === 'audit' ? 'Workspace actions will be recorded here as the team uses Trace.' : 'Tasks appear when an operator or an incident workflow requests work from an agent.'}
            />
          )}
        </div>
      )}
    </main>
  );
}
