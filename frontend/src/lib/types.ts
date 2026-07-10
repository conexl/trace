export interface HostSnapshot {
  hostname: string;
  os: string;
  platform: string;
  kernel: string;
  version?: string;
  uptime: number; // nanoseconds in Go, but JSON encodes duration as number of nanoseconds
}

export interface MemorySnapshot {
  total: number;
  available: number;
  used: number;
  used_percent: number;
  swap_total: number;
  swap_used: number;
  swap_percent: number;
}

export interface DiskSnapshot {
  mountpoint: string;
  filesystem: string;
  total: number;
  free: number;
  used_percent: number;
}

export interface SystemSnapshot {
  cpu_percent: number;
  per_cpu_percent: number[];
  memory: MemorySnapshot;
  disks: DiskSnapshot[];
}

export interface DNSHistoryPoint {
  ts: number;
  latency_ms: number;
  ok: boolean;
}

export interface DNSResult {
  name: string;
  domain: string;
  records: string[];
  matches_public_ip: boolean;
  group?: string;
  latency_ms?: number;
  status?: 'ok' | 'error' | 'slow';
  history?: DNSHistoryPoint[];
  critical?: boolean;
  error?: string;
}

export interface PortResult {
  name: string;
  address: string;
  reachable: boolean;
  latency: number; // nanoseconds
  error?: string;
}

export interface TrafficCounter {
  interface: string;
  bytes_sent: number;
  bytes_recv: number;
}

export interface ListeningPort {
  protocol: string;
  address: string;
  port: number;
  pid?: number;
  process?: string;
}

export interface SpeedResult {
  name: string;
  url: string;
  bytes_read: number;
  duration: number; // nanoseconds
  mbps: number;
  error?: string;
}

export interface NetworkSnapshot {
  public_ip: string;
  dns: DNSResult[];
  ports: PortResult[];
  traffic: TrafficCounter[];
  listening_ports: ListeningPort[];
  speed_tests: SpeedResult[];
}

export interface ProcessSnapshot {
  name: string;
  match: string;
  service: string;
  remote_control?: boolean;
  running: boolean;
  pid?: number;
  status?: string;
  last_exit_code?: number;
  cpu_percent?: number;
  memory_rss?: number;
  error?: string;
}

export interface LogChunk {
  name: string;
  path: string;
  data: string;
  offset: number;
  truncated: boolean;
  collected_at: string;
  error?: string;
}

export interface AgentEvent {
  type: string;
  severity: string;
  message: string;
  subject?: string;
  action?: string;
  exit_code?: number;
  timestamp: string;
}

export interface AgentSnapshot {
  agent_name: string;
  host: HostSnapshot;
  system: SystemSnapshot;
  network: NetworkSnapshot;
  hardware?: unknown;
  processes: ProcessSnapshot[];
  logs: LogChunk[];
  events: AgentEvent[];
  applied_config_revision: number;
  available_services?: string[];
  health: AgentHealth;
  collected_at: string;
}

export interface AgentHealth {
  config_age_seconds: number;
  buffered_events_count: number;
  last_upload_success: boolean;
}

export interface Metric {
  server_id: string;
  timestamp: string;
  cpu: number;
  memory: number;
  net_in?: number;
  net_out?: number;
}

export interface ServerSummary {
  id: string;
  name: string;
  hostname: string;
  platform: string;
  public_ip: string;
  version?: string;
  status: 'online' | 'offline' | 'unknown';
  last_seen: string;
  cpu_percent: number;
  memory_used_percent: number;
  process_count: number;
  event_count: number;
  applied_config_revision: number;
  desired_config_revision: number;
}

export interface ServerState {
  summary: ServerSummary;
  snapshot: AgentSnapshot;
  events: AgentEvent[];
}

export interface ServersResponse {
  servers: ServerSummary[];
}

export type SubscriptionPlan = 'free' | 'plus';

export interface PlanLimits {
  max_servers: number;
  retention_hours: number;
}

export interface PlanFeatures {
  remote_tasks: boolean;
  service_actions: boolean;
  ai_incident_analysis: boolean;
  telegram_notifications: boolean;
  config_management: boolean;
  audit_log: boolean;
}

export interface Subscription {
  plan: SubscriptionPlan;
  status: 'active' | string;
  limits: PlanLimits;
  features: PlanFeatures;
}

export interface AuthUser {
  email: string;
  role: string;
  subscription: Subscription;
}

export interface AgentDesiredConfig {
  agent: {
    name: string;
    interval: number; // nanoseconds
  };
  logging: {
    level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  };
  watchdog: {
    polling_seconds: number;
    timeout_seconds: number;
  };
  performance: {
    mode: 'high' | 'balanced' | 'power-save';
    fan_curve: 'auto' | 'quiet' | 'max';
  };
  network: {
    public_ip_url: string;
    dns_checks: { name: string; domain: string; group?: string; critical?: boolean }[];
    port_checks: { name: string; address: string; timeout: number }[];
    speed_tests: { name: string; url: string; max_bytes: number; timeout: number }[];
  };
  processes: {
    name: string;
    match: string;
    service: string;
    critical: boolean;
    restart: boolean;
    remote_control: boolean;
    restart_command: string[];
    grace_period: number;
    max_restarts: number;
    restart_window: number;
    restart_cooldown: number;
    cpu_threshold: number;
    memory_threshold: number;
  }[];
  log_streams: { name: string; path: string; max_bytes: number }[];
  remote: {
    tasks_enabled: boolean;
    shell_enabled: boolean;
    audit_path: string;
    poll_every: number;
  };
  update: {
    policy: 'manual' | 'check' | 'auto';
    url: string;
    sha256: string;
    signature_url: string;
    ed25519_public_key: string;
  };
  hardware: {
    smart_devices: string[];
  };
  power: {
    prevent_sleep: boolean;
    sleep_at?: string;
    wake_at?: string;
  };
  buffer: {
    path: string;
    max_events: number;
    mirror_to_stdout: boolean;
  };
  revision: number;
  updated_at?: string;
}

export interface Alert {
  id: string;
  server_id: string;
  type: string;
  severity: string;
  subject?: string;
  action?: string;
  exit_code?: number;
  message: string;
  created_at: string;
}

export interface AlertsResponse {
  alerts: Alert[];
}

export interface PairingResponse {
  agent_id: string;
  certificate_pem: string;
  private_key_pem: string;
  ca_cert_pem: string;
  expires_at: string;
}

export interface Task {
  id: string;
  server_id: string;
  name: string;
  payload?: {
    service?: string;
    action?: 'start' | 'stop' | 'restart';
    domains?: string[];
  };
  status: 'pending' | 'running' | 'completed' | 'failed' | 'canceled';
  created_at: string;
  created_by?: string;
  claimed_at?: string;
  completed_at?: string;
  retries: number;
  max_retries: number;
  timeout_seconds?: number;
  result?: {
    exit_code: number;
    stdout: string;
    stderr: string;
    duration_ms: number;
    started_at: string;
    error?: string;
  };
}

export interface AuditLog {
  id: string;
  timestamp: string;
  user_email: string;
  action: string;
  target: string;
  details?: string;
}

export interface TimelineEvent {
  id: string;
  type: 'crash' | 'restart' | 'action' | 'log' | 'resolved';
  timestamp: string;
  title: string;
  message?: string;
  exit_code?: number;
  action?: string;
  actor?: string;
  result?: string;
  metadata?: Record<string, unknown>;
}

export interface Incident {
  id: string;
  server_id: string;
  service_name: string;
  status: 'open' | 'investigating' | 'resolved' | 'suppressed';
  severity: 'critical' | 'warning';
  title: string;
  summary: string;
  timeline: TimelineEvent[];
  created_at: string;
  updated_at: string;
  resolved_at?: string;
}

export interface IncidentAction {
  name: string;
  label: string;
  description: string;
  enabled: boolean;
  coming_soon?: boolean;
}

export interface IncidentsResponse {
  incidents: Incident[];
}

export interface IncidentActionsResponse {
  actions: IncidentAction[];
}

export interface ServiceIncidentMetrics {
  total: number;
  open: number;
  resolved: number;
  critical: number;
  warning: number;
  mttr_seconds: number;
  frequency_per_day: number;
}

export interface IncidentMetrics {
  window: string;
  total: number;
  open: number;
  resolved: number;
  critical: number;
  warning: number;
  mttr_seconds: number;
  frequency_per_day: number;
  by_service: Record<string, ServiceIncidentMetrics>;
}

export interface IncidentAnalysis {
  summary: string;
  root_cause: string;
  severity: 'critical' | 'warning' | 'info';
  suggestions: string[];
  confidence: number;
}

export interface TelegramChat {
  id: number;
  type?: string;
  username?: string;
  title?: string;
  first_name?: string;
}

export interface TelegramNotificationStatus {
  connected: boolean;
  chat?: TelegramChat;
  linked_at?: string;
}

export interface TelegramNotificationLink {
  bot_username: string;
  start_url: string;
  expires_at: string;
}
