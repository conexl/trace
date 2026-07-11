import type * as React from 'react';
import { Card } from '@/components/ui/Card';

interface EmptyStateProps {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

export function EmptyState({ icon: Icon, title, description, action, className }: EmptyStateProps) {
  return (
    <Card hover={false} className={className}>
      <div className="flex min-h-64 flex-col items-center justify-center px-6 py-12 text-center">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-white/[0.035]">
          <Icon className="h-5 w-5 text-muted-soft" />
        </div>
        <h2 className="mt-4 text-sm font-semibold text-active">{title}</h2>
        {description && <p className="mt-1.5 max-w-sm text-sm leading-6 text-muted">{description}</p>}
        {action && <div className="mt-5">{action}</div>}
      </div>
    </Card>
  );
}
