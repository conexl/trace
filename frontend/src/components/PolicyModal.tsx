import * as React from 'react';
import { motion } from 'framer-motion';
import { Settings, Zap, Pin } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog';
import { Button } from '@/components/ui/Button';

export interface ServicePolicy {
  autoRestart: boolean;
  pin: boolean;
  cpuThreshold: number;
  memoryThreshold: number;
}

interface PolicyModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serviceName: string;
  policy: ServicePolicy;
  onSave: (policy: ServicePolicy) => void;
}

export function PolicyModal({ open, onOpenChange, serviceName, policy, onSave }: PolicyModalProps) {
  const [draft, setDraft] = React.useState<ServicePolicy>(policy);

  React.useEffect(() => {
    setDraft(policy);
  }, [policy, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(draft);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md overflow-hidden">
        <motion.div
          initial={{ opacity: 0, y: 14, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
        >
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Settings className="h-4 w-4 text-accent" />
              Policy: {serviceName}
            </DialogTitle>
            <DialogDescription>
              Configure auto-recovery and alert thresholds for this service.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="space-y-5 pt-2">
          <div className="flex items-center justify-between rounded-lg border border-border bg-canvas p-3">
            <div className="flex items-center gap-2 text-sm text-active">
              <Zap className="h-4 w-4 text-accent" />
              Auto-restart
            </div>
            <button
              type="button"
              onClick={() => setDraft((d) => ({ ...d, autoRestart: !d.autoRestart }))}
              className={`relative h-5 w-9 rounded-full transition-colors ${
                draft.autoRestart ? 'bg-accent' : 'bg-surface-elevated border border-border'
              }`}
            >
              <span
                className={`absolute top-0.5 h-4 w-4 rounded-full bg-active transition-transform ${
                  draft.autoRestart ? 'left-[18px]' : 'left-0.5'
                }`}
              />
            </button>
          </div>

          <div className="flex items-center justify-between rounded-lg border border-border bg-canvas p-3">
            <div className="flex items-center gap-2 text-sm text-active">
              <Pin className="h-4 w-4 text-accent" />
              Pin (strict monitoring)
            </div>
            <button
              type="button"
              onClick={() => setDraft((d) => ({ ...d, pin: !d.pin }))}
              className={`relative h-5 w-9 rounded-full transition-colors ${
                draft.pin ? 'bg-accent' : 'bg-surface-elevated border border-border'
              }`}
            >
              <span
                className={`absolute top-0.5 h-4 w-4 rounded-full bg-active transition-transform ${
                  draft.pin ? 'left-[18px]' : 'left-0.5'
                }`}
              />
            </button>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">CPU threshold %</label>
              <p className="text-[10px] text-muted/70">Alert when CPU usage exceeds this value.</p>
              <input
                type="number"
                min={0}
                max={100}
                value={draft.cpuThreshold}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, cpuThreshold: Number(e.target.value) }))
                }
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-mono uppercase text-muted">Memory threshold %</label>
              <p className="text-[10px] text-muted/70">Alert when RAM usage exceeds this value.</p>
              <input
                type="number"
                min={0}
                max={100}
                value={draft.memoryThreshold}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, memoryThreshold: Number(e.target.value) }))
                }
                className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active focus:border-border-focus focus:outline-none font-mono"
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="sm" type="button" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button variant="neon" size="sm" type="submit">
              Save policy
            </Button>
          </div>
        </form>
        </motion.div>
      </DialogContent>
    </Dialog>
  );
}
