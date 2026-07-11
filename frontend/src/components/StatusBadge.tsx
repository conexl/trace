import { cn } from '@/lib/utils';

const toneByStatus: Record<string, string> = {
  active: 'border-emerald-400/25 bg-emerald-400/10 text-emerald-200',
  online: 'border-emerald-400/25 bg-emerald-400/10 text-emerald-200',
  completed: 'border-emerald-400/25 bg-emerald-400/10 text-emerald-200',
  resolved: 'border-emerald-400/25 bg-emerald-400/10 text-emerald-200',
  running: 'border-sky-400/25 bg-sky-400/10 text-sky-200',
  pending: 'border-amber-400/25 bg-amber-400/10 text-amber-200',
  warning: 'border-amber-400/25 bg-amber-400/10 text-amber-200',
  investigating: 'border-amber-400/25 bg-amber-400/10 text-amber-200',
  open: 'border-red-400/25 bg-red-400/10 text-red-200',
  failed: 'border-red-400/25 bg-red-400/10 text-red-200',
  error: 'border-red-400/25 bg-red-400/10 text-red-200',
  offline: 'border-red-400/25 bg-red-400/10 text-red-200',
};

export function StatusBadge({ status, className }: { status: string; className?: string }) {
  const normalized = status.toLowerCase();
  return (
    <span className={cn('inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide', toneByStatus[normalized] ?? 'border-border bg-white/[0.03] text-muted-soft', className)}>
      {status.replace(/_/g, ' ')}
    </span>
  );
}
