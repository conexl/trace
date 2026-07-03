import * as React from 'react';
import { KeyRound } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog';
import { Button } from '@/components/ui/Button';
import { useAuth } from '@/lib/auth';

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuthModal({ open, onOpenChange }: AuthModalProps) {
  const { login, isAuthenticated } = useAuth();
  const [token, setToken] = React.useState('');

  React.useEffect(() => {
    if (open) setToken('');
  }, [open]);

  React.useEffect(() => {
    if (isAuthenticated) {
      onOpenChange(false);
    }
  }, [isAuthenticated, onOpenChange]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (token.trim()) {
      login(token.trim());
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-accent" />
            Авторизация
          </DialogTitle>
          <DialogDescription>
            Введите admin token для управления узлами.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          <div className="space-y-1.5">
            <label className="text-xs font-mono uppercase text-muted">Admin token</label>
            <input
              type="password"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="dev-admin-token"
              className="w-full rounded-md border border-border bg-canvas px-3 py-2 text-sm text-active placeholder:text-muted/50 focus:border-border-focus focus:outline-none font-mono"
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="ghost" size="sm" type="button" onClick={() => onOpenChange(false)}>
              Отмена
            </Button>
            <Button variant="neon" size="sm" type="submit" disabled={!token.trim()}>
              Войти
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
