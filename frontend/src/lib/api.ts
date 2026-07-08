import type {
  Alert,
  ServersResponse,
  ServerState,
  Task,
  PairingResponse,
  AgentDesiredConfig,
  AuditLog,
  Metric,
  Incident,
  IncidentAction,
  IncidentAnalysis,
  IncidentMetrics,
  TelegramNotificationLink,
  TelegramNotificationStatus,
} from './types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '';

export async function login(email: string, password: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
    credentials: 'include',
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error ?? `HTTP ${response.status}`);
  }

  await response.json().catch(() => ({}));
}

export async function getMe(): Promise<{ email: string; role: string }> {
  return request<{ email: string; role: string }>('/v1/auth/me');
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE_URL}/v1/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  }).catch(() => {});
}

export async function register(
  email: string,
  password: string,
  inviteToken?: string
): Promise<void> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (inviteToken) {
    headers.Authorization = `Bearer ${inviteToken}`;
  }
  const response = await fetch(`${API_BASE_URL}/v1/auth/register`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(body.error ?? `HTTP ${response.status}`);
  }

  await response.json().catch(() => ({}));
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${path}`;
  const headers = new Headers(options.headers);

  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(url, { ...options, headers, credentials: 'include' });

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

export function enqueueTask(
  serverId: string,
  taskName: string,
  payload?: { domains?: string[] }
): Promise<Task> {
  return request<Task>(`/v1/servers/${encodeURIComponent(serverId)}/tasks`, {
    method: 'POST',
    body: JSON.stringify({ task_name: taskName, ...payload }),
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

export function getServerConfig(serverId: string): Promise<AgentDesiredConfig> {
  return request<AgentDesiredConfig>(`/v1/servers/${encodeURIComponent(serverId)}/config`);
}

export function updateServerConfig(
  serverId: string,
  config: AgentDesiredConfig
): Promise<AgentDesiredConfig> {
  return request<AgentDesiredConfig>(`/v1/servers/${encodeURIComponent(serverId)}/config`, {
    method: 'POST',
    body: JSON.stringify(config),
  });
}

export function getAuditLogs(limit = 50): Promise<{ audit_logs: AuditLog[] }> {
  return request<{ audit_logs: AuditLog[] }>(`/v1/audit?limit=${limit}`);
}

export function getTask(taskId: string): Promise<Task> {
  return request<Task>(`/v1/tasks/${encodeURIComponent(taskId)}`);
}

export function listTasks(limit = 50): Promise<{ tasks: Task[] }> {
  return request<{ tasks: Task[] }>(`/v1/tasks?limit=${limit}`);
}

export function getMetrics(serverId: string, from?: string, to?: string): Promise<{ metrics: Metric[] }> {
  let url = `/v1/servers/${encodeURIComponent(serverId)}/metrics`;
  const params = new URLSearchParams();
  if (from) params.set('from', from);
  if (to) params.set('to', to);
  if (params.toString()) url += `?${params.toString()}`;
  return request<{ metrics: Metric[] }>(url);
}

export function subscribeToEvents(onEvent: (event: any) => void): () => void {
  const url = `${API_BASE_URL}/v1/events`;
  const source = new EventSource(url, { withCredentials: true });
  source.onmessage = (e) => {
    try {
      onEvent(JSON.parse(e.data));
    } catch (err) {
      console.error('Failed to parse event:', err);
    }
  };
  source.onerror = (err) => {
    console.error('EventSource failed:', err);
  };
  return () => source.close();
}

export function listIncidents(serverId?: string, limit = 50): Promise<{ incidents: Incident[] }> {
  let url = `/v1/incidents?limit=${limit}`;
  if (serverId) url += `&server_id=${encodeURIComponent(serverId)}`;
  return request<{ incidents: Incident[] }>(url);
}

export function getIncidentMetrics(window = '7d', serverId?: string): Promise<IncidentMetrics> {
  const params = new URLSearchParams({ window });
  if (serverId) params.set('server_id', serverId);
  return request<IncidentMetrics>(`/v1/incidents/metrics?${params.toString()}`);
}

export function getIncident(id: string): Promise<Incident> {
  return request<Incident>(`/v1/incidents/${encodeURIComponent(id)}`);
}

export function executeIncidentAction(incidentId: string, action: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/v1/incidents/${encodeURIComponent(incidentId)}/${encodeURIComponent(action)}`, {
    method: 'POST',
  });
}

export function analyzeIncident(incidentId: string): Promise<IncidentAnalysis> {
  return request<IncidentAnalysis>(`/v1/incidents/${encodeURIComponent(incidentId)}/analyze`, {
    method: 'POST',
  });
}

export const fallbackIncidentActions: IncidentAction[] = [
  {
    name: 'restart',
    label: 'Restart Service',
    description: 'Restart the failed service',
    enabled: true,
  },
  {
    name: 'disable-watchdog',
    label: 'Disable Watchdog',
    description: 'Stop automatic restart attempts',
    enabled: true,
  },
  {
    name: 'diagnostics',
    label: 'Run Diagnostics',
    description: 'Collect a safe host/service diagnostic bundle',
    enabled: true,
  },
  {
    name: 'rollback-config',
    label: 'Rollback Config',
    description: 'Restore the previous desired agent configuration',
    enabled: true,
  },
];

export function getIncidentActions(): Promise<{ actions: IncidentAction[] }> {
  return request<{ actions: IncidentAction[] }>('/v1/incidents/actions');
}

export function getTelegramNotificationStatus(): Promise<TelegramNotificationStatus> {
  return request<TelegramNotificationStatus>('/v1/notifications/telegram');
}

export function createTelegramNotificationLink(): Promise<TelegramNotificationLink> {
  return request<TelegramNotificationLink>('/v1/notifications/telegram/link', {
    method: 'POST',
  });
}

export function deleteTelegramNotificationLink(): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>('/v1/notifications/telegram', {
    method: 'DELETE',
  });
}

export { API_BASE_URL };
