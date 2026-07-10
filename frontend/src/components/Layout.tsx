import * as React from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { AnimatePresence } from 'framer-motion';
import { Header } from '@/components/Header';
import { DashboardHeader } from '@/components/DashboardHeader';
import { AuthModal } from '@/components/AuthModal';
import { AddServerModal } from '@/components/AddServerModal';
import { PageTransition } from '@/components/PageTransition';
import { cn } from '@/lib/utils';

export interface LayoutContext {
  onAuthRequired: () => void;
}

export function Layout() {
  const location = useLocation();
  const [authOpen, setAuthOpen] = React.useState(false);
  const [addServerOpen, setAddServerOpen] = React.useState(false);
  const isDashboard = ['/servers', '/incidents', '/tasks', '/alerts'].some((path) =>
    location.pathname === path || location.pathname.startsWith(`${path}/`)
  );

  const contextValue = React.useMemo<LayoutContext>(
    () => ({ onAuthRequired: () => setAuthOpen(true) }),
    []
  );

  return (
    <div className="relative flex min-h-screen flex-col">
      {isDashboard ? (
        <DashboardHeader onAddServerClick={() => setAddServerOpen(true)} />
      ) : (
        <Header onLoginClick={() => setAuthOpen(true)} />
      )}
      <AnimatePresence mode="wait">
        <PageTransition
          key={location.pathname}
          className={cn('flex min-h-screen flex-col', isDashboard ? 'pt-14' : 'pt-32 lg:pt-20')}
        >
          <Outlet context={contextValue} />
        </PageTransition>
      </AnimatePresence>
      <AuthModal open={authOpen} onOpenChange={setAuthOpen} />
      <AddServerModal open={addServerOpen} onOpenChange={setAddServerOpen} />
    </div>
  );
}
