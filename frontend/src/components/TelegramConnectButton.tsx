import * as React from 'react';
import { CheckCircle2, Loader2, MessageCircle } from 'lucide-react';
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
  const chatLabel = status?.chat?.username
    ? `@${status.chat.username}`
    : status?.chat?.title || status?.chat?.first_name || 'Connected';

  return (
    <div className="space-y-2">
      <Button
        variant={connected ? 'default' : 'neon'}
        size="sm"
        onClick={() => (connected ? setConfirmUnlink(true) : connect())}
        disabled={loading}
        className={cn('w-full gap-2 sm:w-auto', error ? 'border-red-400/30 text-red-300 hover:text-red-200' : '')}
        title={error ?? (connected ? 'Telegram connected. Click to unlink.' : 'Connect Telegram notifications')}
      >
        {loading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : connected ? (
          <CheckCircle2 className="h-4 w-4 text-sky-300" />
        ) : (
          <MessageCircle className="h-4 w-4" />
        )}
        {loading ? 'Checking...' : connected ? chatLabel : 'Connect Telegram'}
      </Button>
      {error && <p className="max-w-xs text-xs leading-5 text-red-300">{error}</p>}

      <ConfirmationDialog
        open={confirmUnlink}
        onOpenChange={setConfirmUnlink}
        title="Unlink Telegram"
        description="Stop sending incident notifications to your Telegram chat?"
        confirmLabel="Unlink"
        variant="danger"
        onConfirm={unlink}
      />
    </div>
  );
}
