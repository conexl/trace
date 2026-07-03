import * as React from 'react';
import { motion } from 'framer-motion';
import {
  Settings,
  Server,
  Zap,
  Moon,
  RefreshCw,
  ShieldAlert,
} from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog';
import { Button } from '@/components/ui/Button';
import { useToast } from '@/components/ToastProvider';
import { cn } from '@/lib/utils';

export interface AgentConfig {
  name: string;
  intervalSeconds: number;
  logPaths: string;

  loggingLevel: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  updatePolicy: 'manual' | 'check' | 'auto';

  watchdogPollingSeconds: number;
  watchdogTimeoutSeconds: number;

  performanceMode: 'high' | 'balanced' | 'power-save';
  fanCurve: 'auto' | 'quiet' | 'max';

  sleepSchedule: {
    enabled: boolean;
    sleepAt: string;
    wakeAt: string;
  };
}

interface AgentSettingsModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: AgentConfig;
  onSave: (config: AgentConfig) => void;
}

function Toggle({
  checked,
  onChange,
}: {
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <button
      type="button"
      onClick={() => onChange(!checked)}
      className={cn(
        'relative h-5 w-9 rounded-full transition-colors',
        checked ? 'bg-accent' : 'bg-surface-elevated border border-border'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-4 w-4 rounded-full bg-active transition-transform',
          checked ? 'left-[18px]' : 'left-0.5'
        )}
      />
    </button>
  );
}

function Section({
  icon: Icon,
  title,
  children,
}: {
  icon: React.ElementType;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-3 rounded-xl border border-border bg-canvas/40 p-4">
      <div className="flex items-center gap-2 text-sm font-medium uppercase tracking-wide text-muted">
        <Icon className="h-4 w-4" />
        {title}
      </div>
      {children}
    </div>
  );
}

export function AgentSettingsModal({ open, onOpenChange, config, onSave }: AgentSettingsModalProps) {
  const [draft, setDraft] = React.useState<AgentConfig>(config);
  const { success } = useToast();

  React.useEffect(() => {
    setDraft(config);
  }, [config, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(draft);
    success('Agent settings saved');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl overflow-hidden">
        <motion.div
          initial={{ opacity: 0, y: 14, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
        >
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Settings className="h-4 w-4 text-accent" />
              Agent settings
            </DialogTitle>
            <DialogDescription>
              Configure identity, power, hardware and watchdog behaviour.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="grid max-h-[70vh] gap-5 overflow-auto py-2 pr-1">
            {/* Identity */}
            <Section icon={Server} title="Identity">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">Agent name</label>
                  <input
                    type="text"
                    value={draft.name}
                    onChange={(e) => setDraft((d) => ({ ...d, name: e.target.value }))}
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted" title="Seconds between agent uploads">
                    Snapshot interval
                  </label>
                  <input
                    type="number"
                    min={5}
                    max={3600}
                    value={draft.intervalSeconds}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, intervalSeconds: Number(e.target.value) }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                  />
                </div>
              </div>
            </Section>

            {/* Agent behaviour */}
            <Section icon={RefreshCw} title="Agent behaviour">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">Update policy</label>
                  <select
                    value={draft.updatePolicy}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, updatePolicy: e.target.value as AgentConfig['updatePolicy'] }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  >
                    <option value="manual">Manual updates</option>
                    <option value="check">Check for updates</option>
                    <option value="auto">Auto-apply updates</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">Logging level</label>
                  <select
                    value={draft.loggingLevel}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, loggingLevel: e.target.value as AgentConfig['loggingLevel'] }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  >
                    <option value="DEBUG">DEBUG</option>
                    <option value="INFO">INFO</option>
                    <option value="WARN">WARN</option>
                    <option value="ERROR">ERROR</option>
                  </select>
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-mono uppercase text-muted" title="One path per line">Log paths</label>
                <textarea
                  value={draft.logPaths}
                  onChange={(e) => setDraft((d) => ({ ...d, logPaths: e.target.value }))}
                  rows={3}
                  className="w-full resize-none rounded-md border border-border bg-canvas px-3 py-2 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none font-mono"
                />
              </div>
            </Section>

            {/* Watchdog */}
            <Section icon={ShieldAlert} title="Watchdog strategy">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted" title="How often services are checked">
                    Polling interval
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={300}
                    value={draft.watchdogPollingSeconds}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, watchdogPollingSeconds: Number(e.target.value) }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted" title="Seconds before a service is marked down">
                    Failure timeout
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={600}
                    value={draft.watchdogTimeoutSeconds}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, watchdogTimeoutSeconds: Number(e.target.value) }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                  />
                </div>
              </div>
            </Section>

            {/* Power & Hardware */}
            <Section icon={Zap} title="Power & hardware">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">
                    Performance mode
                  </label>
                  <select
                    value={draft.performanceMode}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, performanceMode: e.target.value as AgentConfig['performanceMode'] }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  >
                    <option value="high">High Performance</option>
                    <option value="balanced">Balanced</option>
                    <option value="power-save">Power Save</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">Fan curve</label>
                  <select
                    value={draft.fanCurve}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, fanCurve: e.target.value as AgentConfig['fanCurve'] }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  >
                    <option value="auto">Auto</option>
                    <option value="quiet">Quiet</option>
                    <option value="max">Max</option>
                  </select>
                </div>
              </div>

              <div className="space-y-3 rounded-lg border border-border bg-canvas/50 p-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Moon className="h-3.5 w-3.5 text-muted" />
                    <span className="text-sm text-active">Sleep / wake schedule</span>
                  </div>
                  <Toggle
                    checked={draft.sleepSchedule.enabled}
                    onChange={(enabled) =>
                      setDraft((d) => ({ ...d, sleepSchedule: { ...d.sleepSchedule, enabled } }))
                    }
                  />
                </div>
                {draft.sleepSchedule.enabled && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    className="grid gap-3 overflow-hidden sm:grid-cols-2"
                  >
                    <div className="space-y-1">
                      <label className="text-xs font-mono uppercase text-muted">Sleep at</label>
                      <input
                        type="time"
                        value={draft.sleepSchedule.sleepAt}
                        onChange={(e) =>
                          setDraft((d) => ({
                            ...d,
                            sleepSchedule: { ...d.sleepSchedule, sleepAt: e.target.value },
                          }))
                        }
                        className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                      />
                    </div>
                    <div className="space-y-1">
                      <label className="text-xs font-mono uppercase text-muted">Wake at</label>
                      <input
                        type="time"
                        value={draft.sleepSchedule.wakeAt}
                        onChange={(e) =>
                          setDraft((d) => ({
                            ...d,
                            sleepSchedule: { ...d.sleepSchedule, wakeAt: e.target.value },
                          }))
                        }
                        className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                      />
                    </div>
                  </motion.div>
                )}
              </div>
            </Section>

            <div className="flex justify-end gap-2 pt-1">
              <Button variant="ghost" size="sm" type="button" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button variant="neon" size="sm" type="submit">
                Save settings
              </Button>
            </div>
          </form>
        </motion.div>
      </DialogContent>
    </Dialog>
  );
}
