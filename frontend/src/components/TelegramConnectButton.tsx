import * as React from 'react';
import { MessageCircle } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { ConfirmationDialog } from '@/components/ConfirmationDialog';
import {
  createTelegramNotificationLink,
  deleteTelegramNotificationLink,
  getTelegramNotificationStatus,
} from '@/lib/api';
import type { TelegramNotificationStatus } from '@/lib/types';
import { cn } from '@/lib/utils';

export function TelegramConnectButton() {
  const [status, setStatus] = React.useState<TelegramNotificationStatus | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [confirmUnlink, setConfirmUnlink] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const pollRef = React.useRef<number | null>(null);

  const refresh = React.useCallback(async () => {
    try {
      const nextStatus = await getTelegramNotificationStatus();
      setStatus(nextStatus);
      if (nextStatus.connected) {
        setError(null);
      }
      return nextStatus;
    } catch {
      setStatus(null);
      return null;
    }
  }, []);

  const stopPolling = React.useCallback(() => {
    if (pollRef.current !== null) {
      window.clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  const startPolling = React.useCallback(() => {
    stopPolling();
    let attempts = 0;
    pollRef.current = window.setInterval(() => {
      attempts += 1;
      void refresh().then((nextStatus) => {
        if (nextStatus?.connected || attempts >= 10) {
          stopPolling();
        }
      });
    }, 2000);
  }, [refresh, stopPolling]);

  React.useEffect(() => {
    refresh();
    return stopPolling;
  }, [refresh, stopPolling]);

  const connect = async () => {
    setLoading(true);
    setError(null);
    try {
      const link = await createTelegramNotificationLink();
      window.open(link.start_url, '_blank', 'noopener,noreferrer');
      startPolling();
    } catch {
      setError('Telegram bot is not configured or temporarily unavailable.');
    } finally {
      setLoading(false);
    }
  };

  const unlink = async () => {
    setLoading(true);
    setError(null);
    try {
      await deleteTelegramNotificationLink();
      await refresh();
      stopPolling();
    } finally {
      setLoading(false);
    }
  };

  const connected = status?.connected ?? false;

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => (connected ? setConfirmUnlink(true) : connect())}
        disabled={loading}
        className={cn(
          'h-8 w-8 p-0 text-muted',
          error ? 'text-red-300 hover:text-red-200' : '',
          connected ? 'text-sky-300 hover:text-sky-200' : 'hover:text-sky-300'
        )}
        title={error ?? (connected ? 'Telegram connected. Click to unlink.' : 'Connect Telegram notifications')}
      >
        <MessageCircle className="h-4 w-4" />
      </Button>

      <ConfirmationDialog
        open={confirmUnlink}
        onOpenChange={setConfirmUnlink}
        title="Unlink Telegram"
        description="Stop sending incident notifications to your Telegram chat?"
        confirmLabel="Unlink"
        variant="danger"
        onConfirm={unlink}
      />
    </>
  );
}
