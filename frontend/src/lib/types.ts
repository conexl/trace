export interface HostSnapshot {
  hostname: string;
  os: string;
  platform: string;
  kernel: string;
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
  collected_at: string;
}

export interface ServerSummary {
  id: string;
  name: string;
  hostname: string;
  platform: string;
  public_ip: string;
  status: 'online' | 'offline' | 'unknown';
  last_seen: string;
  cpu_percent: number;
  memory_used_percent: number;
  process_count: number;
  event_count: number;
}

export interface ServerState {
  summary: ServerSummary;
  snapshot: AgentSnapshot;
  events: AgentEvent[];
}

export interface ServersResponse {
  servers: ServerSummary[];
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
  };
  status: 'pending' | 'running' | 'completed' | 'failed';
  created_at: string;
  claimed_at?: string;
  completed_at?: string;
  result?: {
    exit_code: number;
    stdout: string;
    stderr: string;
    duration_ms: number;
    started_at: string;
    error?: string;
  };
}
