import type {
  Alert,
  ServersResponse,
  ServerState,
  Task,
  PairingResponse,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

let currentAdminToken = import.meta.env.VITE_ADMIN_TOKEN ?? '';

export function setAdminToken(token: string) {
  currentAdminToken = token;
}

export function getAdminToken(): string {
  return currentAdminToken;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${path}`;
  const headers = new Headers(options.headers);

  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json');
  }

  const token = currentAdminToken;
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(url, { ...options, headers });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error ?? `HTTP ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export function getServers(): Promise<ServersResponse> {
  return request<ServersResponse>('/v1/servers');
}

export function getServer(id: string): Promise<ServerState> {
  return request<ServerState>(`/v1/servers/${encodeURIComponent(id)}`);
}

export function listAlerts(limit = 50): Promise<{ alerts: Alert[] }> {
  return request<{ alerts: Alert[] }>(`/v1/alerts?limit=${limit}`);
}

export function enqueueTask(serverId: string, taskName: string): Promise<Task> {
  return request<Task>(`/v1/servers/${encodeURIComponent(serverId)}/tasks`, {
    method: 'POST',
    body: JSON.stringify({ task_name: taskName }),
  });
}

export function enqueueServiceAction(
  serverId: string,
  service: string,
  action: 'start' | 'stop' | 'restart'
): Promise<Task> {
  return request<Task>(`/v1/servers/${encodeURIComponent(serverId)}/service-actions`, {
    method: 'POST',
    body: JSON.stringify({ service, action }),
  });
}

export function claimPairing(
  token: string,
  agentName: string,
  hostname: string
): Promise<PairingResponse> {
  return request<PairingResponse>('/v1/pairing/claim', {
    method: 'POST',
    body: JSON.stringify({ agent_name: agentName, hostname }),
    headers: { Authorization: `Bearer ${token}` },
  });
}

export { API_BASE_URL };
