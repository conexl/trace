import { motion, AnimatePresence } from 'framer-motion';

interface FaultToleranceOverlayProps {
  connected: boolean;
  reconnectIn: number;
  error?: Error | null;
  children: React.ReactNode;
}

export function FaultToleranceOverlay({ connected, reconnectIn, error, children }: FaultToleranceOverlayProps) {
  return (
    <motion.div
      className="relative flex flex-1 flex-col min-h-0"
      animate={connected ? { filter: 'blur(0px) grayscale(0%)' } : { filter: 'blur(1px) grayscale(30%)' }}
      transition={{ duration: 0.3 }}
    >
      <AnimatePresence>
        {!connected && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            className="absolute left-0 right-0 top-0 z-50 flex flex-col items-center justify-center gap-1 py-2"
          >
            <div className="flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-red-500 animate-pulse" />
              <span className="rounded-md border border-border bg-surface px-3 py-1 font-mono text-xs text-active shadow-xl">
                [status: {error?.message || 'connection lost'} · reconnecting in {reconnectIn}s…]
              </span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      {children}
    </motion.div>
  );
}
