import * as React from 'react';
import { LayoutGroup, AnimatePresence, motion } from 'framer-motion';
import { Crown, Plus } from 'lucide-react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import { useAuth } from '@/lib/auth';
import { useServers } from '@/hooks/useServers';
import type { LayoutContext } from '@/components/Layout';
import { NeonButton } from '@/components/NeonButton';
import { ServerCard } from '@/components/ServerCard';
import { AddServerModal } from '@/components/AddServerModal';
import { Card } from '@/components/ui/Card';

export function ServersPage() {
  const { data: servers, loading } = useServers();
  const { isAuthenticated, user } = useAuth();
  const { onAuthRequired } = useOutletContext<LayoutContext>();
  const navigate = useNavigate();
  const [modalOpen, setModalOpen] = React.useState(false);
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
    setModalOpen(true);
  };

  const emptyState = (
    <motion.div
      key="empty"
      initial={{ opacity: 0, scale: 0.96 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.96 }}
      transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
      className="flex flex-1 flex-col items-center justify-center"
    >
      <NeonButton layoutId="add-server-action" onClick={handleAddClick}>
        {atServerLimit ? 'Upgrade for more servers' : 'Добавить первый сервер'}
      </NeonButton>
      {!isAuthenticated && (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3 }}
          className="mt-4 text-sm text-muted"
        >
          Войдите, чтобы добавить узел
        </motion.p>
      )}
    </motion.div>
  );

  if (!isAuthenticated) {
    return (
      <LayoutGroup>
        <main className="flex flex-1 flex-col px-6 py-10">
          <AnimatePresence mode="wait">{emptyState}</AnimatePresence>
        </main>
      </LayoutGroup>
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
    <LayoutGroup>
      <main className="flex flex-1 flex-col px-6 py-10">
        <AnimatePresence mode="wait">
          {!hasServers ? (
            emptyState
          ) : (
            <motion.div
              key="grid"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.15, duration: 0.4 }}
              className="w-full"
            >
              <div className="mb-6 flex items-center justify-between">
                <div>
                  <h1 className="text-2xl font-medium tracking-tight text-active">Узлы</h1>
                  <p className="mt-1 text-sm text-muted">
                    {servers?.length} / {serverLimit} nodes on {user?.subscription.plan ?? 'free'}
                  </p>
                </div>
                {atServerLimit && (
                  <button
                    type="button"
                    onClick={() => navigate('/billing')}
                    className="flex items-center gap-2 rounded-full border border-amber-soft/30 bg-amber-soft/10 px-3 py-1.5 text-xs font-medium text-amber-soft"
                  >
                    <Crown className="h-3.5 w-3.5" />
                    Free limit reached
                  </button>
                )}
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                <motion.div layoutId="add-server-action">
                  <Card
                    dashed
                    hover
                    className="flex h-40 cursor-pointer flex-col items-center justify-center gap-3 border-dashed"
                    onClick={handleAddClick}
                  >
                    <div className="flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface">
                      {atServerLimit ? (
                        <Crown className="h-5 w-5 text-amber-soft" />
                      ) : (
                        <Plus className="h-5 w-5 text-accent" />
                      )}
                    </div>
                    <span className="text-sm font-medium tracking-tight text-muted">
                      {atServerLimit ? 'Upgrade to add nodes' : 'Добавить узел'}
                    </span>
                  </Card>
                </motion.div>

                {servers?.map((server, idx) => (
                  <ServerCard key={server.id} server={server} index={idx} />
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      <AddServerModal open={modalOpen} onOpenChange={setModalOpen} />
    </LayoutGroup>
  );
}
