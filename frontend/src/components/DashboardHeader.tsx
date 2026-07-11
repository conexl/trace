import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { NavLink, useNavigate } from 'react-router-dom';
import { Activity, AlertTriangle, Bell, CreditCard, LogIn, LogOut, Plus, User, Workflow } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { cn } from '@/lib/utils';

const dashboardTabs = [
  { to: '/servers', label: 'Nodes', icon: Activity },
  { to: '/incidents', label: 'Incidents', icon: AlertTriangle },
  { to: '/alerts', label: 'Alerts', icon: Bell },
  { to: '/tasks', label: 'Tasks', icon: Workflow },
];

interface DashboardHeaderProps {
  onAddServerClick: () => void;
}

export function DashboardHeader({ onAddServerClick }: DashboardHeaderProps) {
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const [confirmLogout, setConfirmLogout] = React.useState(false);
  const plan = user?.subscription.plan ?? 'free';

  return (
    <header className="fixed left-0 right-0 top-0 z-40 border-b border-white/10 bg-black/78 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between gap-2 px-3 sm:gap-3 sm:px-6">
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
                title={item.label}
                className={({ isActive }) =>
                  cn(
                    'flex shrink-0 items-center gap-1.5 rounded-lg px-2 py-1.5 text-[10px] font-medium transition-all sm:text-xs sm:gap-2 sm:px-3',
                    isActive
                      ? 'bg-white text-black shadow-[0_8px_24px_rgba(255,255,255,0.10)]'
                      : 'text-muted-soft hover:bg-white/[0.05] hover:text-active'
                  )
                }
              >
                {({ isActive }) => (
                  <>
                    <item.icon className="h-3.5 w-3.5" />
                    <span className={cn('hidden min-[420px]:inline', isActive && 'inline')}>{item.label}</span>
                  </>
                )}
              </NavLink>
            ))}
          </div>
        </nav>

        <div className="flex items-center gap-1.5 sm:gap-2">
          {isAuthenticated && (
            <>
              <button
                type="button"
                onClick={onAddServerClick}
                title="Add node"
                className="flex h-8 w-8 items-center justify-center gap-2 rounded-lg border border-white bg-white text-xs font-semibold text-black transition-colors hover:bg-white/90 sm:w-auto sm:px-3"
              >
                <Plus className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Add node</span>
              </button>
              <button
                type="button"
                onClick={() => navigate('/profile')}
                className="flex h-8 w-8 items-center justify-center gap-2 rounded-full border border-white/10 bg-white/[0.035] text-xs font-medium text-muted-soft transition-colors hover:border-white/20 hover:bg-white/[0.07] hover:text-active md:w-auto md:px-3"
                title="Open profile"
              >
                <User className="h-3.5 w-3.5" />
                <span className="hidden md:inline">Profile</span>
              </button>
              <button
                type="button"
                onClick={() => navigate('/billing')}
                className="hidden items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 py-1.5 text-xs font-medium capitalize text-muted-soft transition-colors hover:border-white/20 hover:bg-white/[0.07] hover:text-active lg:flex"
              >
                <CreditCard className="h-3.5 w-3.5" />
                {plan}
              </button>
            </>
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
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setConfirmLogout(true)}
                  className="hidden h-8 w-8 p-0 text-muted hover:text-active sm:inline-flex"
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
