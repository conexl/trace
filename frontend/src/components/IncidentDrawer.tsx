import * as React from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X, AlertTriangle, RefreshCw, Shield, Wrench, RotateCcw, CheckCircle, Clock, Sparkles, Zap } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import type { Incident, IncidentAnalysis } from '@/lib/types';
import { analyzeIncident, executeIncidentAction, fallbackIncidentActions, getIncidentActions } from '@/lib/api';
import { cn } from '@/lib/utils';

interface IncidentDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  incident: Incident | null;
  onActionExecuted?: () => void;
}

const severityColors = {
  critical: 'text-red-400 bg-red-400/10 border-red-400/30',
  warning: 'text-yellow-400 bg-yellow-400/10 border-yellow-400/30',
};

const statusColors = {
  open: 'text-red-400',
  investigating: 'text-yellow-400',
  resolved: 'text-green-400',
  suppressed: 'text-gray-400',
};

const eventIcons: Record<string, React.ReactNode> = {
  crash: <AlertTriangle className="h-4 w-4 text-red-400" />,
  restart: <RefreshCw className="h-4 w-4 text-yellow-400" />,
  action: <Wrench className="h-4 w-4 text-blue-400" />,
  log: <Clock className="h-4 w-4 text-gray-400" />,
  resolved: <CheckCircle className="h-4 w-4 text-green-400" />,
};

export function IncidentDrawer({ open, onOpenChange, incident, onActionExecuted }: IncidentDrawerProps) {
  const [executing, setExecuting] = React.useState<string | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [analysis, setAnalysis] = React.useState<IncidentAnalysis | null>(null);
  const [analyzing, setAnalyzing] = React.useState(false);
  const [actions, setActions] = React.useState(fallbackIncidentActions);

  React.useEffect(() => {
    if (!open) return;
    let canceled = false;
    getIncidentActions()
      .then((res) => {
        if (!canceled) setActions(res.actions);
      })
      .catch(() => {
        if (!canceled) setActions(fallbackIncidentActions);
      });
    return () => {
      canceled = true;
    };
  }, [open]);

  React.useEffect(() => {
    setAnalysis(null);
    setError(null);
  }, [incident?.id]);

  if (!incident) return null;

  const handleAnalyze = async () => {
    setAnalyzing(true);
    setError(null);
    try {
      const result = await analyzeIncident(incident.id);
      setAnalysis(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Analysis failed');
    } finally {
      setAnalyzing(false);
    }
  };

  const handleAction = async (actionName: string) => {
    setExecuting(actionName);
    setError(null);
    try {
      await executeIncidentAction(incident.id, actionName);
      onActionExecuted?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setExecuting(null);
    }
  };

  const formatTime = (ts: string) => {
    const date = new Date(ts);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const formatDate = (ts: string) => {
    const date = new Date(ts);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

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
                className="fixed inset-0 z-50 bg-canvas/60 backdrop-blur-sm"
                onClick={() => onOpenChange(false)}
              />
            </DialogPrimitive.Overlay>
            <DialogPrimitive.Content asChild>
              <motion.div
                initial={{ x: '100%' }}
                animate={{ x: 0 }}
                exit={{ x: '100%' }}
                transition={{ type: 'spring', damping: 30, stiffness: 300 }}
                className="fixed right-0 top-0 z-50 flex h-full w-full max-w-2xl flex-col border-l border-border bg-surface shadow-2xl"
              >
                {/* Header */}
                <div className="flex items-center justify-between border-b border-border px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className={cn('rounded-md border p-2', severityColors[incident.severity])}>
                      <AlertTriangle className="h-4 w-4" />
                    </div>
                    <div>
                      <DialogPrimitive.Title className="font-mono text-sm text-active">
                        {incident.title}
                      </DialogPrimitive.Title>
                      <div className="mt-0.5 flex items-center gap-2 text-xs text-muted">
                        <span className={statusColors[incident.status]}>{incident.status}</span>
                        <span>•</span>
                        <span>{incident.service_name}</span>
                        <span>•</span>
                        <span>{formatDate(incident.created_at)}</span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <motion.button
                      onClick={handleAnalyze}
                      disabled={analyzing}
                      whileHover={{ scale: 1.05 }}
                      whileTap={{ scale: 0.96 }}
                      className={cn(
                        'flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors',
                        analyzing
                          ? 'border-purple-400/30 bg-purple-400/10 text-purple-300'
                          : 'border-purple-400/30 bg-purple-400/10 text-purple-400 hover:bg-purple-400/20'
                      )}
                    >
                      <Sparkles className={cn('h-3.5 w-3.5', analyzing && 'animate-pulse')} />
                      <span>{analyzing ? 'Analyzing...' : 'AI Analyze'}</span>
                    </motion.button>
                    <DialogPrimitive.Close className="flex h-7 w-7 items-center justify-center rounded-md border border-border text-muted transition-colors hover:border-accent hover:text-accent">
                      <X className="h-4 w-4" />
                    </DialogPrimitive.Close>
                  </div>
                </div>

                {/* AI Analysis */}
                <AnimatePresence>
                  {analysis && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="border-b border-purple-400/20 bg-purple-400/5 px-6 py-4"
                    >
                      <div className="flex items-center gap-2 mb-3">
                        <Sparkles className="h-4 w-4 text-purple-400" />
                        <h3 className="text-xs font-medium uppercase tracking-wider text-purple-400">AI Analysis</h3>
                        <span className="ml-auto text-xs text-muted">
                          Confidence: {(analysis.confidence * 100).toFixed(0)}%
                        </span>
                      </div>
                      <div className="space-y-3">
                        <div>
                          <div className="text-xs font-medium text-active mb-1">Summary</div>
                          <p className="text-xs text-muted">{analysis.summary}</p>
                        </div>
                        <div>
                          <div className="text-xs font-medium text-active mb-1">Root Cause</div>
                          <p className="text-xs text-muted">{analysis.root_cause}</p>
                        </div>
                        {analysis.suggestions.length > 0 && (
                          <div>
                            <div className="text-xs font-medium text-active mb-2">Suggestions</div>
                            <ul className="space-y-1">
                              {analysis.suggestions.map((suggestion, idx) => (
                                <li key={idx} className="flex items-start gap-2 text-xs text-muted">
                                  <Zap className="h-3 w-3 text-purple-400 mt-0.5 shrink-0" />
                                  <span>{suggestion}</span>
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>

                {/* Timeline */}
                <div className="flex-1 overflow-auto px-6 py-4">
                  <div className="mb-4">
                    <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-muted">Timeline</h3>
                    <div className="relative space-y-4">
                      {incident.timeline.map((event, idx) => (
                        <div key={event.id} className="flex gap-3">
                          <div className="flex flex-col items-center">
                            <div className="flex h-8 w-8 items-center justify-center rounded-full border border-border bg-surface">
                              {eventIcons[event.type] || <Clock className="h-4 w-4 text-gray-400" />}
                            </div>
                            {idx < incident.timeline.length - 1 && (
                              <div className="flex-1 w-px bg-border" />
                            )}
                          </div>
                          <div className="flex-1 pb-4">
                            <div className="flex items-center justify-between">
                              <span className="text-sm font-medium text-active">{event.title}</span>
                              <span className="font-mono text-xs text-muted">{formatTime(event.timestamp)}</span>
                            </div>
                            {event.message && (
                              <p className="mt-1 text-xs text-muted">{event.message}</p>
                            )}
                            {event.exit_code !== undefined && event.exit_code !== 0 && (
                              <div className="mt-2 inline-flex items-center gap-1.5 rounded-md bg-red-400/10 px-2 py-1 font-mono text-xs text-red-400">
                                exit code: {event.exit_code}
                              </div>
                            )}
                            {event.actor && (
                              <div className="mt-2 text-xs text-muted">
                                by <span className="text-accent">{event.actor}</span>
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Actions */}
                <div className="border-t border-border px-6 py-4">
                  <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-muted">Actions</h3>
                  {error && (
                    <div className="mb-3 rounded-md bg-red-400/10 px-3 py-2 text-xs text-red-400">
                      {error}
                    </div>
                  )}
                  <div className="grid grid-cols-2 gap-2">
                    {actions.map((action) => (
                      <button
                        key={action.name}
                        disabled={!action.enabled || executing !== null}
                        onClick={() => handleAction(action.name)}
                        className={cn(
                          'flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors',
                          action.enabled
                            ? 'border-border text-muted hover:border-accent hover:text-accent'
                            : 'cursor-not-allowed border-border/50 text-muted/50'
                        )}
                      >
                        {action.name === 'restart' && <RefreshCw className={cn('h-4 w-4', executing === action.name && 'animate-spin')} />}
                        {action.name === 'disable-watchdog' && <Shield className="h-4 w-4" />}
                        {action.name === 'diagnostics' && <Wrench className="h-4 w-4" />}
                        {action.name === 'rollback-config' && <RotateCcw className="h-4 w-4" />}
                        <span>{action.label}</span>
                        {action.coming_soon && (
                          <span className="ml-auto text-xs text-muted/50">Soon</span>
                        )}
                      </button>
                    ))}
                  </div>
                </div>
              </motion.div>
            </DialogPrimitive.Content>
          </DialogPrimitive.Portal>
        )}
      </AnimatePresence>
    </DialogPrimitive.Root>
  );
}
