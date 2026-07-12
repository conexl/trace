import * as React from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { Header } from '@/components/Header';
import { DashboardHeader } from '@/components/DashboardHeader';
import { AuthModal } from '@/components/AuthModal';
import { AddServerModal } from '@/components/AddServerModal';
import { PageTransition } from '@/components/PageTransition';
import { cn } from '@/lib/utils';
import type { PairingCode } from '@/lib/types';

export interface LayoutContext {
  onAuthRequired: () => void;
  onAddServer: () => void;
}

interface AddServerSeed {
  pairing?: PairingCode;
  agentName?: string;
}

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [authOpen, setAuthOpen] = React.useState(false);
  const [addServerOpen, setAddServerOpen] = React.useState(false);
  const [addServerSeed, setAddServerSeed] = React.useState<AddServerSeed | null>(null);
  const isDashboard = ['/servers', '/incidents', '/tasks', '/alerts', '/profile'].some((path) =>
    location.pathname === path || location.pathname.startsWith(`${path}/`)
  );
  const hasProfileLanguageRow = location.pathname === '/profile';

  const openAddServer = React.useCallback((seed?: AddServerSeed) => {
    setAddServerSeed(seed ?? null);
    setAddServerOpen(true);
  }, []);

  React.useEffect(() => {
    const params = new URLSearchParams(location.search);
    const pairingCode = params.get('pairing_code');
    if (!pairingCode) return;

    const agentName = params.get('agent_name') || 'home-server';
    const expiresAt = params.get('expires_at') || new Date(Date.now() + 15 * 60 * 1000).toISOString();

    openAddServer({
      agentName,
      pairing: {
        code: pairingCode,
        expires_at: expiresAt,
      },
    });

    params.delete('pairing_code');
    params.delete('agent_name');
    params.delete('expires_at');
    const search = params.toString();
    navigate({ pathname: location.pathname, search: search ? `?${search}` : '' }, { replace: true });
  }, [location.pathname, location.search, navigate, openAddServer]);

  const contextValue = React.useMemo<LayoutContext>(
    () => ({
      onAuthRequired: () => setAuthOpen(true),
      onAddServer: () => openAddServer(),
    }),
    [openAddServer]
  );

  return (
    <div className="relative flex min-h-screen flex-col">
      {isDashboard ? (
        <DashboardHeader onAddServerClick={() => openAddServer()} />
      ) : (
        <Header onLoginClick={() => setAuthOpen(true)} />
      )}
      <PageTransition
        key={location.pathname}
        className={cn('flex min-h-screen flex-col', isDashboard ? (hasProfileLanguageRow ? 'pt-28 sm:pt-14' : 'pt-14') : 'pt-20')}
      >
        <Outlet context={contextValue} />
      </PageTransition>
      <AuthModal open={authOpen} onOpenChange={setAuthOpen} />
      <AddServerModal
        open={addServerOpen}
        onOpenChange={setAddServerOpen}
        initialPairing={addServerSeed?.pairing}
        initialAgentName={addServerSeed?.agentName}
      />
    </div>
  );
}
