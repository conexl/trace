import * as React from 'react';
import { cn } from '@/lib/utils';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'neon' | 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  asChild?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'md', children, ...props }, ref) => {
    const base =
      'inline-flex items-center justify-center rounded-lg font-medium tracking-tight transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/50 disabled:pointer-events-none disabled:opacity-50';

    const variants = {
      default:
        'bg-white/[0.04] text-active border border-border hover:border-border-glow hover:bg-white/[0.07]',
      neon: 'relative bg-white text-black border border-white shadow-accent-glow hover:bg-white/90 hover:shadow-accent-glow-strong hover:-translate-y-px active:translate-y-0 overflow-hidden',
      ghost: 'text-muted hover:text-active hover:bg-white/[0.05]',
      outline:
        'border border-dashed border-border text-muted hover:border-border-glow hover:text-active bg-transparent',
    };

    const sizes = {
      sm: 'h-8 px-3 text-xs',
      md: 'h-10 px-4 text-sm',
      lg: 'h-12 px-6 text-base',
    };

    return (
      <button
        ref={ref}
        className={cn(base, variants[variant], sizes[size], className)}
        {...props}
      >
        {children}
      </button>
    );
  }
);
Button.displayName = 'Button';
