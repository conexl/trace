import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { NavLink, useNavigate } from 'react-router-dom';
import { Activity, AlertTriangle, CreditCard, LogIn, LogOut, User, Workflow } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { cn } from '@/lib/utils';

const dashboardTabs = [
  { to: '/servers', label: 'Nodes', icon: Activity },
  { to: '/incidents', label: 'Incidents', icon: AlertTriangle },
  { to: '/tasks', label: 'Tasks', icon: Workflow },
];

export function DashboardHeader() {
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const [confirmLogout, setConfirmLogout] = React.useState(false);
  const plan = user?.subscription.plan ?? 'free';

  return (
    <header className="fixed left-0 right-0 top-0 z-40 border-b border-white/10 bg-black/78 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between gap-3 px-4 sm:px-6">
        <button
          type="button"
          onClick={() => navigate('/servers')}
          className="group flex items-center gap-3 rounded-xl px-1.5 py-1 text-left transition-colors hover:bg-white/[0.04]"
        >
          <img
            src="/logo.svg"
            alt="Trace"
            className="h-8 w-8 object-contain drop-shadow-[0_0_12px_rgba(255,255,255,0.18)]"
          />
          <div className="hidden sm:block">
            <p className="text-sm font-semibold tracking-tight text-active">Trace</p>
            <p className="text-[10px] uppercase tracking-[0.18em] text-muted">Dashboard</p>
          </div>
        </button>

        <nav className="flex min-w-0 flex-1 justify-center overflow-x-auto">
          <div className="flex items-center gap-1 rounded-xl border border-white/10 bg-white/[0.03] p-1">
            {dashboardTabs.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  cn(
                    'flex shrink-0 items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-medium transition-all',
                    isActive
                      ? 'bg-white text-black shadow-[0_8px_24px_rgba(255,255,255,0.10)]'
                      : 'text-muted-soft hover:bg-white/[0.05] hover:text-active'
                  )
                }
              >
                <item.icon className="h-3.5 w-3.5" />
                {item.label}
              </NavLink>
            ))}
          </div>
        </nav>

        <div className="flex items-center gap-2">
          {isAuthenticated && (
            <button
              type="button"
              onClick={() => navigate('/billing')}
              className="hidden items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 py-1.5 text-xs font-medium capitalize text-muted-soft transition-colors hover:border-white/20 hover:bg-white/[0.07] hover:text-active md:flex"
            >
              <CreditCard className="h-3.5 w-3.5" />
              {plan}
            </button>
          )}

          <AnimatePresence mode="wait">
            {!isAuthenticated ? (
              <motion.div
                key="login"
                initial={{ opacity: 0, x: 8 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 8 }}
                transition={{ duration: 0.2 }}
              >
                <Button variant="neon" size="sm" onClick={() => navigate('/login')} className="gap-2">
                  <LogIn className="h-4 w-4" />
                  <span className="hidden sm:inline">Login</span>
                </Button>
              </motion.div>
            ) : (
              <motion.div
                key="session"
                initial={{ opacity: 0, scale: 0.96 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.96 }}
                transition={{ duration: 0.2 }}
                className="flex items-center gap-2"
              >
                <div className="hidden h-8 items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 lg:flex">
                  <User className="h-3.5 w-3.5 text-muted-soft" />
                  <span className="max-w-[150px] truncate font-mono text-xs text-active">
                    {user?.email ?? 'Session'}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setConfirmLogout(true)}
                  className="h-8 w-8 p-0 text-muted hover:text-active"
                  title="Logout"
                >
                  <LogOut className="h-4 w-4" />
                </Button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>

      <ConfirmationDialog
        open={confirmLogout}
        onOpenChange={setConfirmLogout}
        title="Log out"
        description="Are you sure you want to end this session?"
        confirmLabel="Log out"
        variant="danger"
        onConfirm={logout}
      />
    </header>
  );
}
