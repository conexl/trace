import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { NavLink, useNavigate } from 'react-router-dom';
import {
  Activity,
  CreditCard,
  Crown,
  LayoutDashboard,
  LogIn,
  LogOut,
  Menu,
  User,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import { cn } from '@/lib/utils';

interface HeaderProps {
  onLoginClick: () => void;
}

const navItems = [
  { to: '/', label: 'Home', icon: LayoutDashboard },
  { to: '/servers', label: 'Dashboard', icon: Activity },
  { to: '/billing', label: 'Pricing', icon: CreditCard },
];

export function Header({ onLoginClick: _onLoginClick }: HeaderProps) {
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const [confirmLogout, setConfirmLogout] = React.useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const plan = user?.subscription.plan ?? 'free';
  const isPlus = plan === 'plus';

  return (
    <header className="fixed left-0 right-0 top-0 z-40 px-3 py-3 sm:px-6">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between gap-2 rounded-2xl border border-white/10 bg-black/72 px-3 shadow-[0_18px_70px_rgba(0,0,0,0.42)] backdrop-blur-xl sm:h-auto sm:py-2">
        <button
          type="button"
          onClick={() => navigate('/')}
          className="group flex min-w-0 items-center gap-2 rounded-xl px-1 py-1.5 text-left transition-colors hover:bg-white/[0.05] sm:gap-3 sm:px-2"
        >
          <img
            src="/logo.svg"
            alt="Trace"
            className="h-8 w-8 shrink-0 object-contain drop-shadow-[0_0_14px_rgba(255,255,255,0.20)] sm:h-10 sm:w-10"
          />
          <div className="hidden sm:block">
            <p className="text-sm font-semibold tracking-tight text-active">Trace</p>
            <p className="text-[10px] uppercase tracking-[0.18em] text-muted">Server control plane</p>
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
                  isActive ? 'bg-white text-black shadow-[0_8px_24px_rgba(255,255,255,0.12)]' : 'text-muted-soft hover:bg-white/[0.05] hover:text-active'
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
                    ? 'border-white/20 bg-white text-black hover:bg-white/90'
                    : 'border-white/10 bg-white/[0.04] text-muted-soft hover:bg-white/[0.08] hover:text-active'
                )}
              >
                {isPlus ? <Crown className="h-3.5 w-3.5" /> : <CreditCard className="h-3.5 w-3.5" />}
                {isPlus ? 'Plus' : 'Free'}
              </button>
            </>
          )}

          <Button variant="default" size="sm" onClick={() => navigate('/servers/demo-server')} className="hidden sm:inline-flex">
            Demo
          </Button>

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
                <button
                  type="button"
                  onClick={() => navigate('/profile')}
                  className="flex h-9 w-9 items-center justify-center gap-2 rounded-full border border-white/10 bg-white/[0.04] transition-colors hover:border-white/20 hover:bg-white/[0.08] md:w-auto md:px-3"
                  title="Open profile"
                >
                  <User className="h-3.5 w-3.5 text-muted-soft" />
                  <span className="hidden max-w-[160px] truncate font-mono text-xs text-active md:inline">
                    {user?.email ?? 'Session'}
                  </span>
                </button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setConfirmLogout(true)}
                  className="h-9 w-9 p-0 text-muted hover:text-active"
                  title="Logout"
                >
                  <LogOut className="h-4 w-4" />
                </Button>
              </motion.div>
            )}
          </AnimatePresence>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => setMobileMenuOpen((open) => !open)}
            className="h-9 w-9 p-0 lg:hidden"
            title="Menu"
          >
            <Menu className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <nav
        className={cn(
          'mx-auto mt-2 max-w-7xl gap-2 overflow-x-auto rounded-2xl border border-white/10 bg-black/82 p-1 backdrop-blur-xl lg:hidden',
          mobileMenuOpen ? 'flex' : 'hidden'
        )}
      >
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex shrink-0 items-center gap-2 rounded-xl px-3 py-2 text-xs font-medium transition-colors',
                isActive ? 'bg-white text-black' : 'text-muted-soft hover:bg-white/[0.05] hover:text-active'
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
