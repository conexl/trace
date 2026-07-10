import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { NavLink, useNavigate } from 'react-router-dom';
import {
  Activity,
  AlertTriangle,
  CreditCard,
  Crown,
  LayoutDashboard,
  LogIn,
  LogOut,
  ShieldCheck,
  User,
  Workflow,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { TelegramConnectButton } from '@/components/TelegramConnectButton';
import { cn } from '@/lib/utils';

interface HeaderProps {
  onLoginClick: () => void;
}

const navItems = [
  { to: '/', label: 'Product', icon: LayoutDashboard },
  { to: '/servers', label: 'Servers', icon: Activity },
  { to: '/incidents', label: 'Incidents', icon: AlertTriangle },
  { to: '/tasks', label: 'Tasks', icon: Workflow },
  { to: '/billing', label: 'Pricing', icon: CreditCard },
];

export function Header({ onLoginClick: _onLoginClick }: HeaderProps) {
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const [confirmLogout, setConfirmLogout] = React.useState(false);
  const plan = user?.subscription.plan ?? 'free';
  const isPlus = plan === 'plus';

  return (
    <header className="fixed left-0 right-0 top-0 z-40 px-3 py-3 sm:px-6">
      <div className="mx-auto flex max-w-7xl items-center justify-between gap-3 rounded-2xl border border-white/10 bg-canvas/78 px-3 py-2 shadow-[0_18px_70px_rgba(0,0,0,0.35)] backdrop-blur-xl">
        <button
          type="button"
          onClick={() => navigate('/')}
          className="group flex items-center gap-3 rounded-xl px-2 py-1.5 text-left transition-colors hover:bg-white/[0.04]"
        >
          <div className="relative flex h-9 w-9 items-center justify-center rounded-xl border border-accent/30 bg-accent/10">
            <ShieldCheck className="h-4 w-4 text-accent" />
            <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-accent shadow-accent-glow" />
          </div>
          <div className="hidden sm:block">
            <p className="text-sm font-semibold tracking-tight text-active">Trace</p>
            <p className="text-[10px] uppercase tracking-[0.18em] text-muted">Homelab SaaS</p>
          </div>
        </button>

        <nav className="hidden items-center gap-1 rounded-xl border border-white/10 bg-white/[0.03] p-1 lg:flex">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition-all',
                  isActive ? 'bg-accent text-black shadow-accent-glow' : 'text-muted hover:bg-white/[0.04] hover:text-active'
                )
              }
            >
              <item.icon className="h-3.5 w-3.5" />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          {isAuthenticated && (
            <>
              <button
                type="button"
                onClick={() => navigate('/billing')}
                className={cn(
                  'hidden items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-semibold transition-colors sm:flex',
                  isPlus
                    ? 'border-accent/35 bg-accent/10 text-accent hover:bg-accent/15'
                    : 'border-amber-soft/35 bg-amber-soft/10 text-amber-soft hover:bg-amber-soft/15'
                )}
              >
                {isPlus ? <Crown className="h-3.5 w-3.5" /> : <CreditCard className="h-3.5 w-3.5" />}
                {isPlus ? 'Plus' : 'Free'}
              </button>
              {isPlus && <TelegramConnectButton />}
            </>
          )}

          <AnimatePresence mode="wait">
            {!isAuthenticated ? (
              <motion.div
                key="login"
                initial={{ opacity: 0, x: 12 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 12 }}
                transition={{ duration: 0.2 }}
              >
                <Button
                  variant="neon"
                  size="sm"
                  onClick={() => navigate('/login')}
                  className="gap-2"
                >
                  <LogIn className="h-4 w-4" />
                  <span className="hidden sm:inline">Login</span>
                </Button>
              </motion.div>
            ) : (
              <motion.div
                key="profile"
                initial={{ opacity: 0, scale: 0.96 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.96 }}
                transition={{ duration: 0.2 }}
                className="flex items-center gap-2"
              >
                <div className="hidden h-9 items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-3 md:flex">
                  <User className="h-3.5 w-3.5 text-accent" />
                  <span className="max-w-[160px] truncate font-mono text-xs text-active">
                    {user?.email ?? 'Session'}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setConfirmLogout(true)}
                  className="h-9 w-9 p-0 text-muted hover:text-accent"
                  title="Logout"
                >
                  <LogOut className="h-4 w-4" />
                </Button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>

      <nav className="mx-auto mt-2 flex max-w-7xl gap-2 overflow-x-auto rounded-2xl border border-white/10 bg-canvas/72 p-1 backdrop-blur-xl lg:hidden">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex shrink-0 items-center gap-2 rounded-xl px-3 py-2 text-xs font-medium transition-colors',
                isActive ? 'bg-accent text-black' : 'text-muted hover:bg-white/[0.04] hover:text-active'
              )
            }
          >
            <item.icon className="h-3.5 w-3.5" />
            {item.label}
          </NavLink>
        ))}
      </nav>

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
