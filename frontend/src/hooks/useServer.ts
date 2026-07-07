import { useCallback } from 'react';
import { getServer } from '@/lib/api';
import { useAuth } from '@/lib/auth';
import type { ServerState } from '@/lib/types';
import { usePolling } from './usePolling';

export function useServer(id: string | undefined, enabled = true) {
  const { isAuthenticated } = useAuth();

  const fetcher = useCallback(async () => {
    if (!id) throw new Error('server id is required');
    return getServer(id);
  }, [id]);

  return usePolling<ServerState>({
    fetcher,
    interval: 2000,
    retryInterval: 5000,
    enabled: enabled && !!id && isAuthenticated,
  });
}
