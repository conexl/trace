import * as React from 'react';
import { Check, Copy, Link2, RefreshCw, Terminal } from 'lucide-react';
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
  initialPairing?: PairingCode | null;
  initialAgentName?: string | null;
}

const defaultOrigin = typeof window !== 'undefined' ? window.location.origin : 'https://trace.solen.one';
const apiOrigin = import.meta.env.VITE_API_BASE_URL || defaultOrigin;

export function AddServerModal({ open, onOpenChange, initialPairing, initialAgentName }: AddServerModalProps) {
  const [agentName, setAgentName] = React.useState('home-server');
  const [pairing, setPairing] = React.useState<PairingCode | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [copied, setCopied] = React.useState<'command' | 'link' | null>(null);

  const installCommand = pairing
    ? `curl -fsSL ${defaultOrigin}/install.sh | sudo env TRACE_ENDPOINT=${apiOrigin} TRACE_PAIRING_CODE=${pairing.code} TRACE_AGENT_NAME=${shellEscape(agentName || 'home-server')} sh`
    : `curl -fsSL ${defaultOrigin}/install.sh | sudo env TRACE_ENDPOINT=${apiOrigin} TRACE_PAIRING_CODE=<code> sh`;
  const pairingLink = pairing
    ? `${defaultOrigin}/servers?${new URLSearchParams({
        pairing_code: pairing.code,
        agent_name: agentName || 'home-server',
        expires_at: pairing.expires_at,
      }).toString()}`
    : `${defaultOrigin}/servers?pairing_code=<code>`;

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
      setCopied(null);
      return;
    }

    if (initialAgentName) {
      setAgentName(initialAgentName);
    }

    if (initialPairing?.code) {
      setPairing(initialPairing);
      setLoading(false);
      setError(null);
      return;
    }

    void generateCode();
  }, [generateCode, initialAgentName, initialPairing, open]);

  const copyValue = async (value: string, kind: 'command' | 'link') => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(kind);
      setTimeout(() => setCopied(null), 1500);
    } catch {
      // Clipboard can be unavailable in embedded previews.
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-none overflow-visible p-0 sm:max-w-4xl">
        <DialogHeader>
          <div className="border-b border-white/10 px-5 pb-4 pt-5 pr-12 sm:px-6 sm:pr-14">
            <DialogTitle className="flex items-center gap-2">
              <Terminal className="h-4 w-4 text-accent" />
              Add server
            </DialogTitle>
            <DialogDescription className="mt-2 max-w-xl">
              Generate a one-time pairing code, then run the install command on the target machine.
            </DialogDescription>
          </div>
        </DialogHeader>

        <div className="max-h-[calc(100vh-14rem)] space-y-4 overflow-y-auto px-5 py-5 sm:max-h-[min(680px,calc(100vh-14rem))] sm:px-6">
          <div className="grid min-w-0 gap-3 md:grid-cols-[1fr_auto] md:items-end">
            <div className="min-w-0 space-y-1.5">
              <label className="text-xs font-mono text-muted">Agent name</label>
              <input
                type="text"
                value={agentName}
                onChange={(e) => setAgentName(e.target.value)}
                placeholder="home-server"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
              />
            </div>
            <Button variant="default" size="md" onClick={generateCode} disabled={loading} className="gap-2 md:w-auto">
              <RefreshCw className={loading ? 'h-4 w-4 animate-spin' : 'h-4 w-4'} />
              New code
            </Button>
          </div>

          <div className="rounded-xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),rgba(255,255,255,0.025)_42%,rgba(255,255,255,0.02))] p-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="min-w-0">
                <p className="text-xs uppercase tracking-[0.2em] text-muted">Pairing code</p>
                <p className="mt-2 break-all font-mono text-2xl font-semibold tracking-[0.08em] text-active">
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

          <div className="group relative min-w-0">
            <div className="pointer-events-none absolute left-3 top-4 text-muted">
              <Terminal className="h-4 w-4" />
            </div>
            <pre className="max-h-40 min-w-0 overflow-x-auto whitespace-pre-wrap break-words rounded-lg border border-border bg-canvas py-3 pl-10 pr-12 font-mono text-xs leading-6 text-active">
              {installCommand}
            </pre>
            <button
              onClick={() => copyValue(installCommand, 'command')}
              disabled={!pairing}
              className="absolute right-2 top-2 rounded-md p-1.5 text-muted transition-colors hover:bg-surface-elevated hover:text-active disabled:opacity-40"
              title="Copy"
            >
              {copied === 'command' ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
            </button>
          </div>

          <div className="min-w-0 rounded-xl border border-white/10 bg-white/[0.025] p-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <Link2 className="h-4 w-4 text-muted-soft" />
                  <p className="text-xs uppercase tracking-[0.2em] text-muted">Pairing link</p>
                </div>
                <a
                  href={pairing ? pairingLink : undefined}
                  className="mt-2 block max-w-full truncate font-mono text-xs text-active underline decoration-white/20 underline-offset-4 transition-colors hover:text-white"
                >
                  {pairingLink}
                </a>
                <p className="mt-2 text-xs leading-5 text-muted-soft">
                  Contains a one-time secret. Share only with the operator installing this node.
                </p>
              </div>
              <Button
                variant="default"
                size="sm"
                onClick={() => copyValue(pairingLink, 'link')}
                disabled={!pairing}
                className="shrink-0 gap-2"
              >
                {copied === 'link' ? <Check className="h-3.5 w-3.5 text-accent" /> : <Copy className="h-3.5 w-3.5" />}
                {copied === 'link' ? 'Copied' : 'Copy link'}
              </Button>
            </div>
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
        </div>

        <div className="flex flex-col-reverse gap-2 border-t border-white/10 bg-surface/95 px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          <Button variant="neon" size="sm" onClick={() => copyValue(installCommand, 'command')} disabled={!pairing} className="w-full sm:w-auto">
            {copied === 'command' ? 'Copied' : 'Copy install command'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function shellEscape(value: string) {
  if (/^[A-Za-z0-9._:-]+$/.test(value)) return value;
  return `'${value.replace(/'/g, `'"'"'`)}'`;
}
