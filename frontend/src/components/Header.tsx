import * as React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { AlertTriangle, Bell, LogIn, LogOut, User } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';

interface HeaderProps {
  onLoginClick: () => void;
}

export function Header({ onLoginClick: _onLoginClick }: HeaderProps) {
  const { isAuthenticated, logout, user } = useAuth();
  const navigate = useNavigate();
  const [confirmLogout, setConfirmLogout] = React.useState(false);

  return (
    <header className="fixed right-0 top-0 z-40 flex items-center gap-2 p-4 sm:p-6">
      {isAuthenticated && (
        <>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/incidents')}
            className="h-8 w-8 p-0 text-muted hover:text-red-400"
            title="Incidents"
          >
            <AlertTriangle className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/alerts')}
            className="h-8 w-8 p-0 text-muted hover:text-accent"
            title="Alerts"
          >
            <Bell className="h-4 w-4" />
          </Button>
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
              variant="ghost"
              size="sm"
              onClick={() => navigate('/login')}
              className="gap-2 text-muted hover:text-accent"
            >
              <LogIn className="h-4 w-4" />
              <span className="hidden sm:inline">Login</span>
            </Button>
          </motion.div>
        ) : (
          <motion.div
            key="profile"
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.9 }}
            transition={{ duration: 0.2 }}
            className="flex items-center gap-2"
          >
            <div className="flex h-8 items-center gap-2 rounded-full border border-border bg-surface px-3">
              <User className="h-3.5 w-3.5 text-accent" />
              <span className="max-w-[120px] truncate font-mono text-xs text-active">
                {user?.email ?? 'Session'}
              </span>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setConfirmLogout(true)}
              className="h-8 w-8 p-0 text-muted hover:text-accent"
              title="Logout"
            >
              <LogOut className="h-4 w-4" />
            </Button>
          </motion.div>
        )}
      </AnimatePresence>

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
