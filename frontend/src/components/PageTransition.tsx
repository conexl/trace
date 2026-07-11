import { cn } from '@/lib/utils';

interface PageTransitionProps {
  children: React.ReactNode;
  className?: string;
}

export function PageTransition({ children, className }: PageTransitionProps) {
  return (
    <div className={cn('animate-page-in flex min-h-0 flex-1 flex-col', className)}>
      {children}
    </div>
  );
}
