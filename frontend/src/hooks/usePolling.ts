import { useCallback, useEffect, useRef, useState } from 'react';

interface UsePollingOptions<T> {
  fetcher: () => Promise<T>;
  interval?: number;
  retryInterval?: number;
  enabled?: boolean;
}

interface UsePollingResult<T> {
  data: T | null;
  error: Error | null;
  loading: boolean;
  connected: boolean;
  reconnectIn: number;
  refresh: () => void;
}

export function usePolling<T>({
  fetcher,
  interval = 2000,
  retryInterval = 5000,
  enabled = true,
}: UsePollingOptions<T>): UsePollingResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [loading, setLoading] = useState(true);
  const [connected, setConnected] = useState(true);
  const [reconnectIn, setReconnectIn] = useState(0);
  const mounted = useRef(true);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearCurrentTimeout = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  const tick = useCallback(async () => {
    if (!mounted.current || !enabled) return;

    try {
      const next = await fetcher();
      if (!mounted.current) return;
      setData(next);
      setError(null);
      setConnected(true);
      setReconnectIn(0);
      timeoutRef.current = setTimeout(tick, interval);
    } catch (err) {
      if (!mounted.current) return;
      setError(err instanceof Error ? err : new Error(String(err)));
      setConnected(false);

      let remaining = Math.ceil(retryInterval / 1000);
      setReconnectIn(remaining);

      const countdown = setInterval(() => {
        remaining -= 1;
        if (mounted.current) {
          setReconnectIn(Math.max(0, remaining));
        }
      }, 1000);

      timeoutRef.current = setTimeout(() => {
        clearInterval(countdown);
        tick();
      }, retryInterval);
    } finally {
      if (mounted.current) {
        setLoading(false);
      }
    }
  }, [fetcher, interval, retryInterval, enabled]);

  const refresh = useCallback(() => {
    clearCurrentTimeout();
    tick();
  }, [clearCurrentTimeout, tick]);

  useEffect(() => {
    mounted.current = true;
    if (enabled) {
      setLoading(true);
      tick();
    }
    return () => {
      mounted.current = false;
      clearCurrentTimeout();
    };
  }, [enabled, tick, clearCurrentTimeout]);

  return { data, error, loading, connected, reconnectIn, refresh };
}
