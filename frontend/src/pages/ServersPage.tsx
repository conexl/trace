import { AnimatePresence, motion } from 'framer-motion';
import { Crown, Server } from 'lucide-react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import { useAuth } from '@/lib/auth';
import { useServers } from '@/hooks/useServers';
import type { LayoutContext } from '@/components/Layout';
import { ServerCard } from '@/components/ServerCard';
import { Button } from '@/components/ui/Button';
import { EmptyState } from '@/components/EmptyState';
import { PageHeader } from '@/components/PageHeader';

export function ServersPage() {
  const { data: servers, loading } = useServers();
  const { isAuthenticated, user } = useAuth();
  const { onAuthRequired, onAddServer } = useOutletContext<LayoutContext>();
  const navigate = useNavigate();
  const hasServers = (servers?.length ?? 0) > 0;
  const serverLimit = user?.subscription.limits.max_servers ?? 1;
  const atServerLimit = isAuthenticated && (servers?.length ?? 0) >= serverLimit;

  const handleAddClick = () => {
    if (!isAuthenticated) {
      onAuthRequired();
      return;
    }
    if (atServerLimit) {
      navigate('/billing');
      return;
    }
    onAddServer();
  };

  const emptyState = (
    <EmptyState
      icon={Server}
      title={atServerLimit ? 'Node limit reached' : 'Connect your first node'}
      description={atServerLimit ? 'Upgrade the workspace plan to connect more servers.' : isAuthenticated ? 'Generate a pairing code, install the agent, and this workspace will start receiving telemetry.' : 'Sign in to create a pairing code and connect a node.'}
      action={<Button variant="neon" size="md" onClick={handleAddClick}>{atServerLimit ? 'Upgrade plan' : 'Add node'}</Button>}
    />
  );

  if (!isAuthenticated) {
    return (
      <main className="page-shell flex flex-1 flex-col px-4 py-6 sm:px-6">
        <PageHeader title="Dashboard" description="Sign in to view connected nodes, incidents and operations." />
        <div className="pt-6">
          <EmptyState
            icon={Server}
            title="Sign in to open your dashboard"
            description="Your workspace keeps server state and operational actions private to your account."
            action={<Button variant="neon" size="md" onClick={() => navigate('/login')}>Sign in</Button>}
          />
        </div>
      </main>
    );
  }

  if (loading && !servers) {
    return (
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
      </main>
    );
  }

  return (
      <main className="page-shell flex flex-1 flex-col px-4 py-6 sm:px-6">
        <AnimatePresence mode="wait">
          {!hasServers ? (
            <div className="pt-2">{emptyState}</div>
          ) : (
            <motion.div
              key="grid"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.15, duration: 0.4 }}
              className="w-full"
            >
              <PageHeader
                title="Nodes"
                description="Connected servers, agent state and configuration delivery."
                eyebrow={`${servers?.length ?? 0} of ${serverLimit} nodes · ${user?.subscription.plan ?? 'free'} plan`}
                actions={
                  atServerLimit ? (
                    <button
                      type="button"
                      onClick={() => navigate('/billing')}
                      className="flex items-center gap-2 rounded-full border border-amber-soft/30 bg-amber-soft/10 px-3 py-1.5 text-xs font-medium text-amber-soft"
                    >
                      <Crown className="h-3.5 w-3.5" />
                      Free limit reached
                    </button>
                  ) : null
                }
              />

              <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {servers?.map((server, idx) => (
                  <ServerCard key={server.id} server={server} index={idx} />
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>
  );
}
