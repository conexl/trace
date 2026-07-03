import * as React from 'react';
import { motion } from 'framer-motion';
import { Plus } from 'lucide-react';
import { cn } from '@/lib/utils';

interface NeonButtonProps {
  layoutId?: string;
  children?: React.ReactNode;
  className?: string;
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
}

export function NeonButton({
  className,
  layoutId,
  children,
  onClick,
  disabled,
  type = 'button',
}: NeonButtonProps) {
  return (
    <motion.button
      type={type}
      layoutId={layoutId}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'group relative inline-flex items-center justify-center gap-3 rounded-2xl bg-canvas px-10 py-6 text-xl font-medium tracking-tight text-accent border-2 border-accent shadow-accent-glow transition-shadow duration-300 hover:shadow-accent-glow-strong focus:outline-none',
        className
      )}
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
    >
      <Plus className="h-6 w-6 transition-transform duration-300 group-hover:rotate-90" />
      <span className="transition-transform duration-300 group-hover:-translate-y-px">
        {children ?? 'Добавить первый сервер'}
      </span>
    </motion.button>
  );
}
