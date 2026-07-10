import { useCallback } from 'react';
import { listAlerts } from '@/lib/api';
import { useAuth } from '@/lib/auth';
import type { Alert } from '@/lib/types';
import { usePolling } from './usePolling';

export function useAlerts(enabled = true) {
  const { isAuthenticated } = useAuth();

  const fetcher = useCallback(async () => {
    const response = await listAlerts(100);
    return Array.isArray(response.alerts) ? response.alerts : [];
  }, []);

  return usePolling<Alert[]>({
    fetcher,
    interval: 3000,
    retryInterval: 5000,
    enabled: enabled && isAuthenticated,
  });
}
