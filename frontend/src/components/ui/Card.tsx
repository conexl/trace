import * as React from 'react';
import { cn } from '@/lib/utils';

export interface CardProps {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
  dashed?: boolean;
  onClick?: React.MouseEventHandler<HTMLDivElement>;
  style?: React.CSSProperties;
}

export const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, children, hover = true, dashed = false, onClick, style }, ref) => (
      <div
        ref={ref}
        className={cn(
          'relative overflow-hidden rounded-xl border bg-surface transition-colors duration-300',
          dashed ? 'border-dashed border-border' : 'border-[0.5px] border-border',
          hover && 'cursor-pointer hover:border-border-glow hover:bg-white/[0.015]',
          className
        )}
        style={style}
        onClick={onClick}
      >
        <div className="h-full">{children}</div>
      </div>
    )
);
Card.displayName = 'Card';
