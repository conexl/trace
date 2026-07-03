import * as React from 'react';
import { Copy, Check, Terminal } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog';
import { Button } from '@/components/ui/Button';
import { claimPairing } from '@/lib/api';
import type { PairingResponse } from '@/lib/types';

interface AddServerModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AddServerModal({ open, onOpenChange }: AddServerModalProps) {
  const [token, setToken] = React.useState('');
  const [agentName, setAgentName] = React.useState('home-server');
  const [hostname, setHostname] = React.useState('');
  const [copied, setCopied] = React.useState(false);
  const [claiming, setClaiming] = React.useState(false);
  const [claimError, setClaimError] = React.useState<string | null>(null);
  const [claimed, setClaimed] = React.useState<PairingResponse | null>(null);
  const [copiedCredential, setCopiedCredential] = React.useState<string | null>(null);

  const bashCommand = 'homelytics-agent -config ./agent.yaml -pair -pair-dir ./certs';

  const copyCommand = async () => {
    try {
      await navigator.clipboard.writeText(bashCommand);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // ignore
    }
  };

  const copyCredential = async (label: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedCredential(label);
      setTimeout(() => setCopiedCredential(null), 1500);
    } catch {
      // Browser clipboard permission can be unavailable in embedded previews.
    }
  };

  const handleClaim = async () => {
    if (!token || !agentName) return;
    setClaiming(true);
    setClaimError(null);
    try {
      const response = await claimPairing(token, agentName, hostname || agentName);
      setClaimed(response);
    } catch (err) {
      setClaimError(err instanceof Error ? err.message : 'claim failed');
    } finally {
      setClaiming(false);
    }
  };

  React.useEffect(() => {
    if (!open) {
      setToken('');
      setClaimed(null);
      setClaimError(null);
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-accent" />
            Добавить сервер
          </DialogTitle>
          <DialogDescription>
            Скопируйте команду и запустите в терминале на сервере.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-mono text-muted">Pairing token</label>
              <input
                type="text"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="pair-once"
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none font-mono"
              />
            </div>
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
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-mono text-muted">Hostname (optional)</label>
            <input
              type="text"
              value={hostname}
              onChange={(e) => setHostname(e.target.value)}
              placeholder="mac-mini"
              className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none"
            />
          </div>

          <div className="relative group">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 text-muted">
              <Terminal className="h-4 w-4" />
            </div>
            <pre className="overflow-x-auto rounded-lg border border-border bg-canvas py-3 pl-10 pr-12 font-mono text-xs text-active">
              {bashCommand}
            </pre>
            <button
              onClick={copyCommand}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted transition-colors hover:bg-surface-elevated hover:text-active"
              title="Copy"
            >
              {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
            </button>
          </div>
          <p className="text-[11px] leading-relaxed text-muted">
            CLI pairing reads backend URL, token and agent name from <span className="font-mono">agent.yaml</span>.
            The token here can also claim credentials directly from the dashboard for a quick demo.
          </p>

          <div className="flex items-center justify-between pt-1">
            <Button variant="ghost" size="sm" onClick={() => onOpenChange(false)}>
              Закрыть
            </Button>
            <Button
              variant="neon"
              size="sm"
              onClick={handleClaim}
              disabled={claiming || !token || !agentName}
            >
              {claiming ? 'Claiming…' : 'Claim credentials'}
            </Button>
          </div>

          {claimError && (
            <p className="rounded-md border border-red-900/30 bg-red-950/20 px-3 py-2 text-xs text-red-400">
              {claimError}
            </p>
          )}

          {claimed && (
            <div className="space-y-2 rounded-lg border border-border bg-canvas p-3">
              <div className="flex items-center justify-between text-xs">
                <span className="text-muted">agent_id</span>
                <span className="font-mono text-active">{claimed.agent_id}</span>
              </div>
              <div className="text-xs text-muted">Certificates issued. Save them before closing.</div>
              <div className="grid grid-cols-3 gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="text-[10px]"
                  onClick={() => copyCredential('ca.pem', claimed.ca_cert_pem)}
                >
                  {copiedCredential === 'ca.pem' ? 'copied' : 'ca.pem'}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="text-[10px]"
                  onClick={() => copyCredential('agent.pem', claimed.certificate_pem)}
                >
                  {copiedCredential === 'agent.pem' ? 'copied' : 'agent.pem'}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="text-[10px]"
                  onClick={() => copyCredential('agent-key.pem', claimed.private_key_pem)}
                >
                  {copiedCredential === 'agent-key.pem' ? 'copied' : 'agent-key.pem'}
                </Button>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
