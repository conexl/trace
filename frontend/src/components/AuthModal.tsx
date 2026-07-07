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

  React.useEffect(() => {
    if (open) {
      login();
    }
  }, [open, login]);

  React.useEffect(() => {
    if (isAuthenticated) {
      onOpenChange(false);
    }
  }, [isAuthenticated, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <KeyRound className="h-4 w-4 text-accent" />
            Авторизация
          </DialogTitle>
          <DialogDescription>
            Сессия будет обновлена автоматически, если кука доступна.
          </DialogDescription>
        </DialogHeader>

        <div className="pt-2 text-xs text-muted">
          Если вы не вошли, перейдите на страницу входа.
        </div>
        <div className="mt-4 flex justify-end">
          <Button variant="ghost" size="sm" type="button" onClick={() => onOpenChange(false)}>
            Закрыть
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
