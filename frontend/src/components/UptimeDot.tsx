import { cn } from '@/lib/utils';

interface UptimeDotProps {
  status: 'online' | 'offline' | 'unknown';
}

export function UptimeDot({ status }: UptimeDotProps) {
  const isOnline = status === 'online';

  return (
    <div className="relative flex h-2.5 w-2.5 items-center justify-center">
      {isOnline && (
        <span className="absolute inline-flex h-full w-full animate-pulse-slow rounded-full bg-accent opacity-50" />
      )}
      <span
        className={cn(
          'relative inline-flex h-1.5 w-1.5 rounded-full',
          isOnline ? 'bg-accent' : 'bg-muted'
        )}
      />
    </div>
  );
}
