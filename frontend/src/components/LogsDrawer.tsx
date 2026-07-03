import * as React from 'react';
import * as DialogPrimitive from '@radix-ui/react-dialog';
import { X, ScrollText } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import type { LogChunk } from '@/lib/types';

interface LogsDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serviceName: string;
  logs: LogChunk[];
}

export function LogsDrawer({ open, onOpenChange, serviceName, logs }: LogsDrawerProps) {
  const scrollRef = React.useRef<HTMLDivElement>(null);

  const lines = logs
    .filter((chunk) => chunk.name.toLowerCase() === serviceName.toLowerCase())
    .flatMap((chunk) =>
      chunk.data
        .split('\n')
        .filter((line) => line.trim().length > 0)
        .map((line) => ({ line, name: chunk.name }))
    )
    .slice(-300);

  React.useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [lines.length, open]);

  return (
    <DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <DialogPrimitive.Portal forceMount>
            <DialogPrimitive.Overlay asChild>
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="fixed inset-0 z-50 bg-canvas/60 backdrop-blur-sm"
                onClick={() => onOpenChange(false)}
              />
            </DialogPrimitive.Overlay>
            <DialogPrimitive.Content asChild>
              <motion.div
                initial={{ x: '100%' }}
                animate={{ x: 0 }}
                exit={{ x: '100%' }}
                transition={{ type: 'spring', damping: 30, stiffness: 300 }}
                className="fixed right-0 top-0 z-50 flex h-full w-full max-w-xl flex-col border-l border-border bg-surface shadow-2xl"
              >
                <div className="flex items-center justify-between border-b border-border px-5 py-4">
                  <div className="flex items-center gap-2">
                    <ScrollText className="h-4 w-4 text-accent" />
                    <DialogPrimitive.Title className="font-mono text-sm text-active">
                      logs / {serviceName}
                    </DialogPrimitive.Title>
                  </div>
                  <DialogPrimitive.Close className="flex h-7 w-7 items-center justify-center rounded-md border border-border text-muted transition-colors hover:border-accent hover:text-accent">
                    <X className="h-4 w-4" />
                  </DialogPrimitive.Close>
                </div>

                <div
                  ref={scrollRef}
                  className="flex-1 overflow-auto bg-black px-5 py-4 font-mono text-xs leading-relaxed"
                >
                  {lines.length === 0 ? (
                    <div className="flex h-full flex-col items-center justify-center gap-2 text-center text-muted/50">
                      <ScrollText className="h-6 w-6 text-muted/30" />
                      <span>No log stream for {serviceName}</span>
                    </div>
                  ) : (
                    lines.map(({ line }, idx) => (
                      <div key={idx} className="flex gap-3">
                        <span className="shrink-0 text-accent/60">❯</span>
                        <span className="text-accent/90">{line}</span>
                      </div>
                    ))
                  )}
                </div>
              </motion.div>
            </DialogPrimitive.Content>
          </DialogPrimitive.Portal>
        )}
      </AnimatePresence>
    </DialogPrimitive.Root>
  );
}
