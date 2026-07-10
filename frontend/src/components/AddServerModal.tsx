import * as React from 'react';
import { Check, Copy, RefreshCw, Terminal } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog';
import { Button } from '@/components/ui/Button';
import { createPairingCode } from '@/lib/api';
import type { PairingCode } from '@/lib/types';

interface AddServerModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const defaultOrigin = typeof window !== 'undefined' ? window.location.origin : 'https://trace.solen.one';
const apiOrigin = import.meta.env.VITE_API_BASE_URL || defaultOrigin;

export function AddServerModal({ open, onOpenChange }: AddServerModalProps) {
  const [agentName, setAgentName] = React.useState('home-server');
  const [pairing, setPairing] = React.useState<PairingCode | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [copied, setCopied] = React.useState(false);

  const installCommand = pairing
    ? `curl -fsSL ${defaultOrigin}/install.sh | sudo env TRACE_ENDPOINT=${apiOrigin} TRACE_PAIRING_CODE=${pairing.code} TRACE_AGENT_NAME=${shellEscape(agentName || 'home-server')} sh`
    : `curl -fsSL ${defaultOrigin}/install.sh | sudo env TRACE_ENDPOINT=${apiOrigin} TRACE_PAIRING_CODE=<code> sh`;

  const generateCode = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const code = await createPairingCode();
      setPairing(code);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to create pairing code');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    if (!open) {
      setPairing(null);
      setError(null);
      setCopied(false);
      return;
    }
    void generateCode();
  }, [generateCode, open]);

  const copyCommand = async () => {
    try {
      await navigator.clipboard.writeText(installCommand);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard can be unavailable in embedded previews.
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-accent" />
            Add server
          </DialogTitle>
          <DialogDescription>
            Run one command on the server. Trace will install the agent, pair it, and start the service.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          <div className="grid gap-3 sm:grid-cols-[1fr_auto] sm:items-end">
            <div className="space-y-1.5">
              <label className="text-xs font-mono text-muted">Agent name</label>
              <input
                type="text"
                value={agentName}
                onChange={(e) => setAgentName(e.target.value)}
                placeholder="home-server"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
              />
            </div>
            <Button variant="default" size="md" onClick={generateCode} disabled={loading} className="gap-2">
              <RefreshCw className={loading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
              New code
            </Button>
          </div>

          <div className="rounded-xl border border-white/10 bg-white/[0.035] p-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-muted">Pairing code</p>
                <p className="mt-2 font-mono text-2xl font-semibold tracking-[0.08em] text-active">
                  {pairing?.code ?? 'Generating...'}
                </p>
              </div>
              {pairing && (
                <p className="font-mono text-xs text-muted-soft">
                  Expires {new Date(pairing.expires_at).toLocaleTimeString()}
                </p>
              )}
            </div>
          </div>

          <div className="relative group">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 text-muted">
              <Terminal className="h-4 w-4" />
            </div>
            <pre className="max-h-36 overflow-x-auto whitespace-pre-wrap rounded-lg border border-border bg-canvas py-3 pl-10 pr-12 font-mono text-xs leading-6 text-active">
              {installCommand}
            </pre>
            <button
              onClick={copyCommand}
              disabled={!pairing}
              className="absolute right-2 top-2 rounded-md p-1.5 text-muted transition-colors hover:bg-surface-elevated hover:text-active disabled:opacity-40"
              title="Copy"
            >
              {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
            </button>
          </div>

          <div className="rounded-lg border border-border bg-canvas p-3 text-xs leading-6 text-muted-soft">
            The installer downloads the release binary, writes <span className="font-mono text-active">/etc/homelytics/agent.yaml</span>,
            claims mTLS credentials, and registers a <span className="font-mono text-active">systemd</span> or
            <span className="font-mono text-active"> launchd</span> service. No repository clone is required.
          </div>

          {error && (
            <p className="rounded-md border border-red-900/30 bg-red-950/20 px-3 py-2 text-xs text-red-400">
              {error}
            </p>
          )}

          <div className="flex items-center justify-between pt-1">
            <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
              Close
            </Button>
            <Button variant="neon" size="sm" onClick={copyCommand} disabled={!pairing}>
              {copied ? 'Copied' : 'Copy install command'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function shellEscape(value: string) {
  if (/^[A-Za-z0-9._:-]+$/.test(value)) return value;
  return `'${value.replace(/'/g, `'"'"'`)}'`;
}
