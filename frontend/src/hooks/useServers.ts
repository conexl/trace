import { useCallback } from 'react';
import { getServers } from '@/lib/api';
import { useAuth } from '@/lib/auth';
import type { ServerSummary } from '@/lib/types';
import { usePolling } from './usePolling';

export function useServers(enabled = true) {
  const { isAuthenticated, token } = useAuth();

  const fetcher = useCallback(async () => {
    const response = await getServers();
    return response.servers;
  }, [token]);

  return usePolling<ServerSummary[]>({
    fetcher,
    interval: 3000,
    retryInterval: 5000,
    enabled: enabled && isAuthenticated,
  });
}
