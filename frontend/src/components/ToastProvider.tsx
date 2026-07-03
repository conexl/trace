import * as React from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { CheckCircle2, XCircle, Info, X } from 'lucide-react';
import { cn } from '@/lib/utils';

type ToastType = 'success' | 'error' | 'info';

interface Toast {
  id: string;
  type: ToastType;
  message: string;
}

interface ToastContextValue {
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
  remove: (id: string) => void;
}

const ToastContext = React.createContext<ToastContextValue | null>(null);

export function useToast(): ToastContextValue {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<Toast[]>([]);

  const add = React.useCallback((type: ToastType, message: string) => {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    setToasts((prev) => [...prev, { id, type, message }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  }, []);

  const remove = React.useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const value = React.useMemo(
    () => ({
      success: (message: string) => add('success', message),
      error: (message: string) => add('error', message),
      info: (message: string) => add('info', message),
      remove,
    }),
    [add, remove]
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2">
        <AnimatePresence>
          {toasts.map((toast) => (
            <motion.div
              key={toast.id}
              initial={{ opacity: 0, y: 12, scale: 0.96 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, x: 24 }}
              transition={{ duration: 0.2, ease: 'easeOut' }}
              className={cn(
                'flex w-72 items-start gap-2.5 rounded-lg border bg-surface px-3 py-2.5 shadow-xl',
                toast.type === 'success' && 'border-accent/30',
                toast.type === 'error' && 'border-red-500/30',
                toast.type === 'info' && 'border-border'
              )}
            >
              {toast.type === 'success' && <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-accent" />}
              {toast.type === 'error' && <XCircle className="mt-0.5 h-4 w-4 shrink-0 text-red-400" />}
              {toast.type === 'info' && <Info className="mt-0.5 h-4 w-4 shrink-0 text-muted" />}
              <span className="flex-1 text-xs text-active">{toast.message}</span>
              <button
                onClick={() => remove(toast.id)}
                className="text-muted transition-colors hover:text-active"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
}
