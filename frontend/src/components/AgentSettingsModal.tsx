import * as React from 'react';
import { motion } from 'framer-motion';
import {
  Settings,
  Server,
  Zap,
  Moon,
  RefreshCw,
  ShieldAlert,
  Globe,
  Terminal,
  HardDrive,
  Package,
  Trash2,
  Plus,
  X,
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
import { getServerConfig, updateServerConfig } from '@/lib/api';
import type { AgentDesiredConfig } from '@/lib/types';
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
  serverId: string;
  isDemo?: boolean;
  config: AgentConfig;
  onSave: (config: AgentConfig) => void;
}

const DEFAULT_AGENT_CONFIG: AgentConfig = {
  name: '',
  intervalSeconds: 60,
  logPaths: '/var/log/syslog\n/var/log/nginx/access.log',
  loggingLevel: 'INFO',
  updatePolicy: 'check',
  watchdogPollingSeconds: 10,
  watchdogTimeoutSeconds: 30,
  performanceMode: 'balanced',
  fanCurve: 'auto',
  sleepSchedule: { enabled: false, sleepAt: '23:00', wakeAt: '07:00' },
};

const SECONDS_TO_NS = 1_000_000_000;

function defaultDesiredConfig(): AgentDesiredConfig {
  return {
    agent: { name: '', interval: 60 * SECONDS_TO_NS },
    logging: { level: 'INFO' },
    watchdog: { polling_seconds: 10, timeout_seconds: 30 },
    performance: { mode: 'balanced', fan_curve: 'auto' },
    network: { public_ip_url: 'https://api.ipify.org', dns_checks: [], port_checks: [], speed_tests: [] },
    processes: [],
    log_streams: [],
    remote: { tasks_enabled: true, shell_enabled: false, audit_path: '', poll_every: 15 * SECONDS_TO_NS },
    update: { policy: 'check', url: '', sha256: '', signature_url: '', ed25519_public_key: '' },
    hardware: { smart_devices: [] },
    power: { prevent_sleep: false },
    buffer: { path: '', max_events: 1000, mirror_to_stdout: false },
    revision: 0,
  };
}

function agentConfigFromDraft(draft: AgentDesiredConfig): AgentConfig {
  const lines = (draft.log_streams ?? [])
    .map((s) => s.path)
    .filter(Boolean)
    .join('\n');

  return {
    name: draft.agent?.name ?? '',
    intervalSeconds: Math.max(5, Math.round((draft.agent?.interval ?? 60 * SECONDS_TO_NS) / SECONDS_TO_NS)),
    logPaths: lines || DEFAULT_AGENT_CONFIG.logPaths,
    loggingLevel: draft.logging?.level ?? DEFAULT_AGENT_CONFIG.loggingLevel,
    updatePolicy: draft.update?.policy ?? DEFAULT_AGENT_CONFIG.updatePolicy,
    watchdogPollingSeconds: draft.watchdog?.polling_seconds ?? DEFAULT_AGENT_CONFIG.watchdogPollingSeconds,
    watchdogTimeoutSeconds: draft.watchdog?.timeout_seconds ?? DEFAULT_AGENT_CONFIG.watchdogTimeoutSeconds,
    performanceMode: draft.performance?.mode ?? DEFAULT_AGENT_CONFIG.performanceMode,
    fanCurve: draft.performance?.fan_curve ?? DEFAULT_AGENT_CONFIG.fanCurve,
    sleepSchedule: {
      enabled: draft.power?.prevent_sleep ?? false,
      sleepAt: draft.power?.sleep_at ?? DEFAULT_AGENT_CONFIG.sleepSchedule.sleepAt,
      wakeAt: draft.power?.wake_at ?? DEFAULT_AGENT_CONFIG.sleepSchedule.wakeAt,
    },
  };
}

function applyAgentConfig(draft: AgentDesiredConfig, cfg: AgentConfig): AgentDesiredConfig {
  return {
    ...draft,
    agent: { name: cfg.name, interval: cfg.intervalSeconds * SECONDS_TO_NS },
    logging: { level: cfg.loggingLevel },
    watchdog: { polling_seconds: cfg.watchdogPollingSeconds, timeout_seconds: cfg.watchdogTimeoutSeconds },
    performance: { mode: cfg.performanceMode, fan_curve: cfg.fanCurve },
    log_streams: cfg.logPaths
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((path, idx) => ({
        name: `log-${idx + 1}`,
        path,
        max_bytes: 128 * 1024,
      })),
    power: {
      ...draft.power,
      prevent_sleep: cfg.sleepSchedule.enabled,
      sleep_at: cfg.sleepSchedule.enabled ? cfg.sleepSchedule.sleepAt : undefined,
      wake_at: cfg.sleepSchedule.enabled ? cfg.sleepSchedule.wakeAt : undefined,
    },
    update: {
      ...draft.update,
      policy: cfg.updatePolicy,
    },
  };
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

function NumberField({
  label,
  value,
  min,
  max,
  onChange,
}: {
  label: string;
  value: number;
  min?: number;
  max?: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="space-y-1.5">
      <label className="text-xs font-mono uppercase text-muted">{label}</label>
      <input
        type="number"
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
      />
    </div>
  );
}

function SelectField<T extends string>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: { value: T; label: string }[];
  onChange: (value: T) => void;
}) {
  return (
    <div className="space-y-1.5">
      <label className="text-xs font-mono uppercase text-muted">{label}</label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as T)}
        className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );
}

export function AgentSettingsModal({
  open,
  onOpenChange,
  serverId,
  isDemo,
  config,
  onSave,
}: AgentSettingsModalProps) {
  const [draft, setDraft] = React.useState<AgentDesiredConfig>(() => applyAgentConfig(defaultDesiredConfig(), config));
  const [loading, setLoading] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const { success, error: showError } = useToast();

  React.useEffect(() => {
    setDraft((prev) => applyAgentConfig(prev, config));
  }, [config, open]);

  React.useEffect(() => {
    if (!open || isDemo || !serverId) return;

    let canceled = false;
    setLoading(true);
    getServerConfig(serverId)
      .then((cfg) => {
        if (canceled) return;
        const merged = { ...defaultDesiredConfig(), ...cfg };
        setDraft(applyAgentConfig(merged, config));
      })
      .catch((err) => {
        if (canceled) return;
        showError(err instanceof Error ? err.message : 'Failed to load agent config');
      })
      .finally(() => setLoading(false));

    return () => {
      canceled = true;
    };
  }, [open, serverId, isDemo, showError, config]);

  const updateAgentConfig = (patch: Partial<AgentConfig> | ((prev: AgentConfig) => AgentConfig)) => {
    setDraft((prev) => {
      const current = agentConfigFromDraft(prev);
      const next = typeof patch === 'function' ? patch(current) : { ...current, ...patch };
      return applyAgentConfig(prev, next);
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!isDemo && serverId) {
      setSaving(true);
      try {
        const toSave = { ...draft, updated_at: new Date().toISOString() };
        const saved = await updateServerConfig(serverId, toSave);
        setDraft(saved);
        success('Agent settings saved');
      } catch (err) {
        showError(err instanceof Error ? err.message : 'Failed to save agent config');
        setSaving(false);
        return;
      } finally {
        setSaving(false);
      }
    }

    onSave(agentConfigFromDraft(draft));
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl overflow-hidden">
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
              Configure identity, network, power, remote access and update policy.
            </DialogDescription>
          </DialogHeader>

          {loading && (
            <div className="flex items-center justify-center py-8">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-accent" />
            </div>
          )}

          {!loading && (
            <form onSubmit={handleSubmit} className="grid max-h-[75vh] gap-5 overflow-auto py-2 pr-1">
              {/* Identity */}
              <Section icon={Server} title="Identity">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Agent name</label>
                    <input
                      type="text"
                      value={draft.agent.name}
                      onChange={(e) => updateAgentConfig({ name: e.target.value })}
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                    />
                  </div>
                  <NumberField
                    label="Snapshot interval"
                    value={Math.round(draft.agent.interval / SECONDS_TO_NS)}
                    min={5}
                    max={3600}
                    onChange={(v) => updateAgentConfig({ intervalSeconds: v })}
                  />
                </div>
              </Section>

              {/* Agent behaviour */}
              <Section icon={RefreshCw} title="Agent behaviour">
                <div className="grid gap-4 sm:grid-cols-2">
                  <SelectField
                    label="Update policy"
                    value={draft.update.policy}
                    options={[
                      { value: 'manual', label: 'Manual updates' },
                      { value: 'check', label: 'Check for updates' },
                      { value: 'auto', label: 'Auto-apply updates (DANGEROUS)' },
                    ]}
                    onChange={(updatePolicy) => updateAgentConfig({ updatePolicy })}
                  />
                  <SelectField
                    label="Logging level"
                    value={draft.logging.level}
                    options={[
                      { value: 'DEBUG', label: 'DEBUG' },
                      { value: 'INFO', label: 'INFO' },
                      { value: 'WARN', label: 'WARN' },
                      { value: 'ERROR', label: 'ERROR' },
                    ]}
                    onChange={(loggingLevel) => updateAgentConfig({ loggingLevel })}
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted" title="One path per line">Log paths</label>
                  <textarea
                    value={draft.log_streams.map((s) => s.path).join('\n')}
                    onChange={(e) => updateAgentConfig({ logPaths: e.target.value })}
                    rows={3}
                    className="w-full resize-none rounded-md border border-border bg-canvas px-3 py-2 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none font-mono"
                  />
                </div>
              </Section>

              {/* Watchdog */}
              <Section icon={ShieldAlert} title="Watchdog strategy">
                <div className="grid gap-4 sm:grid-cols-2">
                  <NumberField
                    label="Polling interval"
                    value={draft.watchdog.polling_seconds}
                    min={1}
                    max={300}
                    onChange={(v) =>
                      setDraft((d) => ({ ...d, watchdog: { ...d.watchdog, polling_seconds: v } }))
                    }
                  />
                  <NumberField
                    label="Failure timeout"
                    value={draft.watchdog.timeout_seconds}
                    min={1}
                    max={600}
                    onChange={(v) =>
                      setDraft((d) => ({ ...d, watchdog: { ...d.watchdog, timeout_seconds: v } }))
                    }
                  />
                </div>
              </Section>

              {/* Power & Hardware */}
              <Section icon={Zap} title="Power & hardware">
                <div className="grid gap-4 sm:grid-cols-2">
                  <SelectField
                    label="Performance mode"
                    value={draft.performance.mode}
                    options={[
                      { value: 'high', label: 'High Performance' },
                      { value: 'balanced', label: 'Balanced' },
                      { value: 'power-save', label: 'Power Save' },
                    ]}
                    onChange={(mode) =>
                      setDraft((d) => ({ ...d, performance: { ...d.performance, mode } }))
                    }
                  />
                  <SelectField
                    label="Fan curve"
                    value={draft.performance.fan_curve}
                    options={[
                      { value: 'auto', label: 'Auto' },
                      { value: 'quiet', label: 'Quiet' },
                      { value: 'max', label: 'Max' },
                    ]}
                    onChange={(fan_curve) =>
                      setDraft((d) => ({ ...d, performance: { ...d.performance, fan_curve } }))
                    }
                  />
                </div>

                <div className="space-y-3 rounded-lg border border-border bg-canvas/50 p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Moon className="h-3.5 w-3.5 text-muted" />
                      <span className="text-sm text-active">Sleep / wake schedule</span>
                    </div>
                    <Toggle
                      checked={draft.power.prevent_sleep}
                      onChange={(enabled) =>
                        updateAgentConfig((prev) => ({
                          ...prev,
                          sleepSchedule: { ...prev.sleepSchedule, enabled },
                        }))
                      }
                    />
                  </div>
                  {draft.power.prevent_sleep && (
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
                          value={draft.power.sleep_at ?? '23:00'}
                          onChange={(e) =>
                            updateAgentConfig((prev) => ({
                              ...prev,
                              sleepSchedule: { ...prev.sleepSchedule, sleepAt: e.target.value },
                            }))
                          }
                          className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-xs font-mono uppercase text-muted">Wake at</label>
                        <input
                          type="time"
                          value={draft.power.wake_at ?? '07:00'}
                          onChange={(e) =>
                            updateAgentConfig((prev) => ({
                              ...prev,
                              sleepSchedule: { ...prev.sleepSchedule, wakeAt: e.target.value },
                            }))
                          }
                          className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                        />
                      </div>
                    </motion.div>
                  )}
                </div>

                <StringListEditor
                  label="SMART devices"
                  placeholder="/dev/sda"
                  values={draft.hardware.smart_devices}
                  onChange={(smart_devices) => setDraft((d) => ({ ...d, hardware: { ...d.hardware, smart_devices } }))}
                />
              </Section>

              {/* Network */}
              <Section icon={Globe} title="Network">
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase text-muted">Public IP URL</label>
                  <input
                    type="text"
                    value={draft.network.public_ip_url}
                    onChange={(e) =>
                      setDraft((d) => ({ ...d, network: { ...d.network, public_ip_url: e.target.value } }))
                    }
                    className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                  />
                </div>

                <ObjectListEditor
                  label="Port checks"
                  fields={[
                    { key: 'name', label: 'Name', type: 'text' },
                    { key: 'address', label: 'Address', type: 'text' },
                    { key: 'timeout', label: 'Timeout (s)', type: 'number' },
                  ]}
                  rows={draft.network.port_checks.map((c) => ({
                    name: c.name,
                    address: c.address,
                    timeout: Math.round(c.timeout / SECONDS_TO_NS),
                  }))}
                  onChange={(rows) =>
                    setDraft((d) => ({
                      ...d,
                      network: {
                        ...d.network,
                        port_checks: rows.map((r) => ({
                          name: String(r.name),
                          address: String(r.address),
                          timeout: Number(r.timeout) * SECONDS_TO_NS,
                        })),
                      },
                    }))
                  }
                />

                <ObjectListEditor
                  label="Speed tests"
                  fields={[
                    { key: 'name', label: 'Name', type: 'text' },
                    { key: 'url', label: 'URL', type: 'text' },
                    { key: 'max_bytes', label: 'Max bytes', type: 'number' },
                    { key: 'timeout', label: 'Timeout (s)', type: 'number' },
                  ]}
                  rows={draft.network.speed_tests.map((t) => ({
                    name: t.name,
                    url: t.url,
                    max_bytes: t.max_bytes,
                    timeout: Math.round(t.timeout / SECONDS_TO_NS),
                  }))}
                  onChange={(rows) =>
                    setDraft((d) => ({
                      ...d,
                      network: {
                        ...d.network,
                        speed_tests: rows.map((r) => ({
                          name: String(r.name),
                          url: String(r.url),
                          max_bytes: Number(r.max_bytes),
                          timeout: Number(r.timeout) * SECONDS_TO_NS,
                        })),
                      },
                    }))
                  }
                />
              </Section>

              {/* Remote */}
              <Section icon={Terminal} title="Remote control">
                <div className="flex items-center justify-between rounded-lg border border-border bg-canvas/50 p-3">
                  <div className="flex items-center gap-2">
                    <RefreshCw className="h-3.5 w-3.5 text-muted" />
                    <div className="flex flex-col">
                      <span className="text-sm text-active">Tasks enabled</span>
                      <span className="text-[10px] text-amber-muted uppercase font-mono">Potentially dangerous</span>
                    </div>
                  </div>
                  <Toggle
                    checked={draft.remote.tasks_enabled}
                    onChange={(tasks_enabled) =>
                      setDraft((d) => ({ ...d, remote: { ...d.remote, tasks_enabled } }))
                    }
                  />
                </div>
                <div
                  className="flex items-center justify-between rounded-lg border border-border bg-canvas/50 p-3 opacity-60"
                  title="Remote shell is disabled until cloud-side authorization is implemented"
                >
                  <div className="flex items-center gap-2">
                    <Terminal className="h-3.5 w-3.5 text-muted" />
                    <span className="text-sm text-active">Shell enabled</span>
                  </div>
                  <Toggle
                    checked={false}
                    onChange={() => {}}
                  />
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Audit path</label>
                    <input
                      type="text"
                      value={draft.remote.audit_path}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, remote: { ...d.remote, audit_path: e.target.value } }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                    />
                  </div>
                  <NumberField
                    label="Poll every (s)"
                    value={Math.round(draft.remote.poll_every / SECONDS_TO_NS)}
                    min={1}
                    max={300}
                    onChange={(v) =>
                      setDraft((d) => ({ ...d, remote: { ...d.remote, poll_every: v * SECONDS_TO_NS } }))
                    }
                  />
                </div>
              </Section>

              {/* Buffer */}
              <Section icon={HardDrive} title="Buffer">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Buffer path</label>
                    <input
                      type="text"
                      value={draft.buffer.path}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, buffer: { ...d.buffer, path: e.target.value } }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                    />
                  </div>
                  <NumberField
                    label="Max events"
                    value={draft.buffer.max_events}
                    min={1}
                    max={100000}
                    onChange={(v) =>
                      setDraft((d) => ({ ...d, buffer: { ...d.buffer, max_events: v } }))
                    }
                  />
                </div>
                <div className="flex items-center justify-between rounded-lg border border-border bg-canvas/50 p-3">
                  <span className="text-sm text-active">Mirror to stdout</span>
                  <Toggle
                    checked={draft.buffer.mirror_to_stdout}
                    onChange={(mirror_to_stdout) =>
                      setDraft((d) => ({ ...d, buffer: { ...d.buffer, mirror_to_stdout } }))
                    }
                  />
                </div>
              </Section>

              {/* Update */}
              <Section icon={Package} title="Agent update">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Update URL</label>
                    <input
                      type="text"
                      value={draft.update.url}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, update: { ...d.update, url: e.target.value } }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">SHA256</label>
                    <input
                      type="text"
                      value={draft.update.sha256}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, update: { ...d.update, sha256: e.target.value } }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                    />
                  </div>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Signature URL</label>
                    <input
                      type="text"
                      value={draft.update.signature_url}
                      onChange={(e) =>
                        setDraft((d) => ({ ...d, update: { ...d.update, signature_url: e.target.value } }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase text-muted">Ed25519 public key</label>
                    <input
                      type="text"
                      value={draft.update.ed25519_public_key}
                      onChange={(e) =>
                        setDraft((d) => ({
                          ...d,
                          update: { ...d.update, ed25519_public_key: e.target.value },
                        }))
                      }
                      className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
                    />
                  </div>
                </div>
              </Section>

              <div className="flex justify-end gap-2 pt-1">
                <Button variant="ghost" size="sm" type="button" onClick={() => onOpenChange(false)}>
                  Cancel
                </Button>
                <Button variant="neon" size="sm" type="submit" disabled={saving}>
                  {saving ? 'Saving…' : 'Save settings'}
                </Button>
              </div>
            </form>
          )}
        </motion.div>
      </DialogContent>
    </Dialog>
  );
}

function StringListEditor({
  label,
  placeholder,
  values,
  onChange,
}: {
  label: string;
  placeholder?: string;
  values: string[];
  onChange: (values: string[]) => void;
}) {
  const [draft, setDraft] = React.useState('');

  const add = () => {
    const value = draft.trim();
    if (!value || values.includes(value)) return;
    onChange([...values, value]);
    setDraft('');
  };

  return (
    <div className="space-y-2">
      <label className="text-xs font-mono uppercase text-muted">{label}</label>
      <div className="flex flex-wrap gap-2">
        {values.map((v) => (
          <span
            key={v}
            className="inline-flex items-center gap-1 rounded-md border border-border bg-canvas px-2 py-1 text-xs text-active"
          >
            {v}
            <button
              type="button"
              onClick={() => onChange(values.filter((x) => x !== v))}
              className="text-muted hover:text-red-400"
            >
              <X className="h-3 w-3" />
            </button>
          </span>
        ))}
      </div>
      <div className="flex gap-2">
        <input
          type="text"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), add())}
          placeholder={placeholder}
          className="min-w-0 flex-1 rounded-md border border-border bg-canvas px-3 py-1.5 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
        />
        <button
          type="button"
          onClick={add}
          disabled={!draft.trim()}
          className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-border bg-canvas text-muted transition-colors hover:border-accent hover:text-accent disabled:opacity-40"
        >
          <Plus className="h-3 w-3" />
        </button>
      </div>
    </div>
  );
}

function ObjectListEditor({
  label,
  fields,
  rows,
  onChange,
}: {
  label: string;
  fields: { key: string; label: string; type: 'text' | 'number' }[];
  rows: Record<string, string | number>[];
  onChange: (rows: Record<string, string | number>[]) => void;
}) {
  const add = () => {
    const empty: Record<string, string | number> = {};
    fields.forEach((f) => (empty[f.key] = f.type === 'number' ? 0 : ''));
    onChange([...rows, empty]);
  };

  const update = (idx: number, key: string, value: string | number) => {
    onChange(rows.map((r, i) => (i === idx ? { ...r, [key]: value } : r)));
  };

  const remove = (idx: number) => {
    onChange(rows.filter((_, i) => i !== idx));
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-xs font-mono uppercase text-muted">{label}</label>
        <button
          type="button"
          onClick={add}
          className="flex items-center gap-1 rounded-md border border-border bg-canvas px-2 py-1 text-[10px] text-active transition-colors hover:border-accent hover:text-accent"
        >
          <Plus className="h-3 w-3" />
          Add
        </button>
      </div>
      <div className="space-y-2">
        {rows.map((row, idx) => (
          <div key={idx} className="flex items-start gap-2 rounded-lg border border-border bg-canvas/50 p-2">
            {fields.map((f) => (
              <div key={f.key} className="min-w-0 flex-1">
                <input
                  type={f.type === 'number' ? 'number' : 'text'}
                  value={row[f.key]}
                  onChange={(e) => update(idx, f.key, f.type === 'number' ? Number(e.target.value) : e.target.value)}
                  placeholder={f.label}
                  className="w-full rounded-md border border-border bg-canvas px-2 py-1 text-xs text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
                />
              </div>
            ))}
            <button
              type="button"
              onClick={() => remove(idx)}
              className="mt-1 text-muted hover:text-red-400"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
        {rows.length === 0 && (
          <div className="rounded-lg border border-dashed border-border bg-canvas/30 py-4 text-center text-xs text-muted">
            No entries
          </div>
        )}
      </div>
    </div>
  );
}
