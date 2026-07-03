import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Compass, ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/Button';

export function NotFoundPage() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center px-6 py-12 text-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
        className="flex flex-col items-center"
      >
        <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl border border-border bg-surface">
          <Compass className="h-8 w-8 text-accent" />
        </div>
        <h1 className="text-4xl font-medium tracking-tight text-active">404</h1>
        <p className="mt-2 text-sm text-muted">This page doesn’t exist.</p>
        <div className="mt-6 flex items-center gap-3">
          <Link to="/">
            <Button variant="neon" size="sm" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back home
            </Button>
          </Link>
        </div>
      </motion.div>
    </main>
  );
}
