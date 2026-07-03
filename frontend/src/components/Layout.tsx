import * as React from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { AnimatePresence } from 'framer-motion';
import { Header } from '@/components/Header';
import { AuthModal } from '@/components/AuthModal';
import { PageTransition } from '@/components/PageTransition';

export interface LayoutContext {
  onAuthRequired: () => void;
}

export function Layout() {
  const location = useLocation();
  const [authOpen, setAuthOpen] = React.useState(false);

  const contextValue = React.useMemo<LayoutContext>(
    () => ({ onAuthRequired: () => setAuthOpen(true) }),
    []
  );

  return (
    <div className="relative flex min-h-screen flex-col">
      <Header onLoginClick={() => setAuthOpen(true)} />
      <AnimatePresence mode="wait">
        <PageTransition key={location.pathname} className="flex min-h-screen flex-col">
          <Outlet context={contextValue} />
        </PageTransition>
      </AnimatePresence>
      <AuthModal open={authOpen} onOpenChange={setAuthOpen} />
    </div>
  );
}
