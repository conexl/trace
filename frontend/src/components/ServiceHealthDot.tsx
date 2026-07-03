import { cn } from '@/lib/utils';

type HealthStatus = 'healthy' | 'restarting' | 'down';

interface ServiceHealthDotProps {
  status: HealthStatus;
}

export function ServiceHealthDot({ status }: ServiceHealthDotProps) {
  return (
    <div className="relative flex h-2.5 w-2.5 items-center justify-center">
      {status === 'restarting' && (
        <span className="absolute inline-flex h-full w-full animate-pulse-slow rounded-full bg-amber-500/50" />
      )}
      <span
        className={cn(
          'relative inline-flex h-1.5 w-1.5 rounded-full',
          status === 'healthy' && 'bg-accent',
          status === 'restarting' && 'bg-amber-500',
          status === 'down' && 'bg-red-700'
        )}
      />
    </div>
  );
}
