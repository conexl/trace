import * as React from 'react';
import { motion } from 'framer-motion';
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
  ({ className, children, hover = true, dashed = false, onClick, style }, ref) => {
    const localRef = React.useRef<HTMLDivElement | null>(null);
    const [position, setPosition] = React.useState({ x: 0, y: 0 });
    const [isHovered, setIsHovered] = React.useState(false);

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
      const rect = (ref as React.RefObject<HTMLDivElement>)?.current?.getBoundingClientRect() ??
        localRef.current?.getBoundingClientRect();
      if (!rect) return;
      setPosition({ x: e.clientX - rect.left, y: e.clientY - rect.top });
    };

    const setRefs = (node: HTMLDivElement | null) => {
      localRef.current = node;
      if (typeof ref === 'function') ref(node);
      else if (ref) (ref as React.MutableRefObject<HTMLDivElement | null>).current = node;
    };

    return (
      <motion.div
        ref={setRefs}
        className={cn(
          'relative overflow-hidden rounded-xl border bg-surface transition-colors duration-300',
          dashed ? 'border-dashed border-border' : 'border-[0.5px] border-border',
          hover && 'hover:border-border-glow cursor-pointer',
          className
        )}
        style={style}
        onClick={onClick}
        onMouseMove={handleMouseMove}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
        whileHover={hover ? { scale: 1.01 } : undefined}
        transition={{ type: 'spring', stiffness: 400, damping: 30 }}
      >
        {hover && (
          <motion.div
            className="pointer-events-none absolute inset-0 opacity-0 transition-opacity duration-500"
            animate={{ opacity: isHovered ? 1 : 0 }}
            style={{
              background: `radial-gradient(400px circle at ${position.x}px ${position.y}px, rgba(255, 255, 255, 0.06), transparent 60%)`,
            }}
          />
        )}
        <div className="relative z-10 h-full">{children}</div>
      </motion.div>
    );
  }
);
Card.displayName = 'Card';
