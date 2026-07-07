import type { DNSHistoryPoint, ServerState } from './types';

export interface HistoryPoint {
  ts: number;
  cpu: number;
  ram: number;
  swap: number;
  [core: string]: number;
}

function generateDNSHistory(baseLatency: number, errorSpike = false): DNSHistoryPoint[] {
  const points = 48;
  return Array.from({ length: points }, (_, i) => {
    const noise = Math.random() * 40 - 20;
    const spike = errorSpike && i > 38 && i < 44 ? 400 + Math.random() * 300 : 0;
    const latency = Math.max(10, baseLatency + noise + spike);
    return {
      ts: Date.now() - (points - i) * 1_800_000,
      latency_ms: Math.round(latency),
      ok: latency < 500,
    };
  });
}

export const DEMO_STATE: ServerState = {
  summary: {
    id: 'demo-server',
    name: 'demo-server',
    hostname: 'mac-mini-m4',
    platform: 'darwin',
    public_ip: '203.0.113.42',
    status: 'online',
    last_seen: new Date().toISOString(),
    cpu_percent: 23.4,
    memory_used_percent: 45.2,
    process_count: 4,
    event_count: 0,
    applied_config_revision: 12,
    desired_config_revision: 12,
  },

  snapshot: {
    agent_name: 'demo-server',
    host: {
      hostname: 'mac-mini-m4',
      os: 'darwin',
      platform: 'macOS',
      kernel: '24.0.0',
      uptime: 172800000000000,
    },
    system: {
      cpu_percent: 23.4,
      per_cpu_percent: [12, 34, 56, 28, 19, 41, 8, 33],
      memory: {
        total: 34359738368,
        available: 18874368000,
        used: 15489372160,
        used_percent: 45.2,
        swap_total: 2147483648,
        swap_used: 104857600,
        swap_percent: 4.9,
      },
      disks: [
        { mountpoint: '/', filesystem: 'apfs', total: 994662965248, free: 412719226880, used_percent: 58.5 },
        { mountpoint: '/Data', filesystem: 'apfs', total: 2000000000000, free: 1500000000000, used_percent: 25.0 },
      ],
    },
    network: {
      public_ip: '203.0.113.42',
      dns: [
        { name: 'api-prod', domain: 'api.example.com', records: ['203.0.113.42'], matches_public_ip: true, group: 'Production', latency_ms: 28, status: 'ok', critical: true, history: generateDNSHistory(30) },
        { name: 'app-prod', domain: 'app.example.com', records: ['203.0.113.42'], matches_public_ip: true, group: 'Production', latency_ms: 34, status: 'ok', critical: true, history: generateDNSHistory(35) },
        { name: 'cdn-prod', domain: 'cdn.example.com', records: ['203.0.113.43'], matches_public_ip: false, group: 'Production', latency_ms: 18, status: 'error', history: generateDNSHistory(20, true) },
        { name: 'api-staging', domain: 'api.staging.example.com', records: ['198.51.100.10'], matches_public_ip: false, group: 'Staging', latency_ms: 62, status: 'slow', history: generateDNSHistory(65) },
        { name: 'git-staging', domain: 'git.staging.example.com', records: ['198.51.100.11'], matches_public_ip: true, group: 'Staging', latency_ms: 22, status: 'ok', history: generateDNSHistory(25) },
        { name: 'vpn-internal', domain: 'vpn.internal.example.com', records: ['10.0.0.5'], matches_public_ip: false, group: 'Internal', latency_ms: 14, status: 'ok', history: generateDNSHistory(15) },
        { name: 'ldap-internal', domain: 'ldap.internal.example.com', records: ['10.0.0.12'], matches_public_ip: false, group: 'Internal', latency_ms: 120, status: 'slow', history: generateDNSHistory(110) },
        { name: 'monitoring', domain: 'grafana.internal.example.com', records: ['10.0.0.20'], matches_public_ip: false, group: 'Internal', latency_ms: 41, status: 'ok', history: generateDNSHistory(40) },
      ],
      ports: [
        { name: 'local-web', address: '127.0.0.1:8080', reachable: true, latency: 405501 },
      ],
      traffic: [
        { interface: 'en0', bytes_sent: 1240000000, bytes_recv: 7800000000 },
      ],
      listening_ports: [
        { protocol: 'tcp4', address: '0.0.0.0', port: 22, pid: 42, process: 'sshd' },
        { protocol: 'tcp4', address: '0.0.0.0', port: 8080, pid: 101, process: 'homelytics-agent' },
        { protocol: 'tcp4', address: '127.0.0.1', port: 3000, pid: 88, process: 'node' },
      ],
      speed_tests: [
        { name: 'cloudflare-small', url: 'https://speed.cloudflare.com/__down?bytes=1000000', bytes_read: 1000000, duration: 800000000, mbps: 95.4, error: '' },
      ],
    },
    processes: [
      { name: 'sshd', match: 'sshd', service: 'sshd', remote_control: false, running: true, pid: 42, status: 'running', cpu_percent: 0.1, memory_rss: 2048000 },
      { name: 'homelytics-agent', match: 'homelytics-agent', service: '', running: true, pid: 101, status: 'running', cpu_percent: 2.3, memory_rss: 16777216 },
      { name: 'nginx', match: 'nginx', service: 'nginx', remote_control: true, running: false, status: 'down', cpu_percent: 0, memory_rss: 0, error: 'process not found' },
      { name: 'postgres', match: 'postgres', service: 'postgres', remote_control: true, running: true, pid: 205, status: 'running', cpu_percent: 1.1, memory_rss: 67108864 },
    ],
    logs: [
      { name: 'system', path: '/var/log/demo.log', data: '2026-07-02T12:00:00 [INFO] agent started\n2026-07-02T12:00:01 [INFO] snapshot uploaded\n2026-07-02T12:00:02 [WARN] high memory usage: 45%\n2026-07-02T12:00:03 [INFO] all processes healthy\n', offset: 0, truncated: false, collected_at: new Date().toISOString() },
    ],
    events: [],
    applied_config_revision: 12,
    health: {
      config_age_seconds: 120,
      buffered_events_count: 0,
      last_upload_success: true,
    },
    collected_at: new Date().toISOString(),
  },
  events: [],
};

export function generateDemoHistory(points = 60): HistoryPoint[] {
  return Array.from({ length: points }, (_, i) => {
    const t = i / points;
    return {
      ts: Date.now() - (points - i) * 2000,
      cpu: 15 + Math.sin(t * Math.PI * 4) * 10 + Math.random() * 5,
      ram: 40 + Math.cos(t * Math.PI * 2) * 8 + Math.random() * 4,
      swap: 3 + Math.random() * 3,
      core0: 5 + Math.random() * 20,
      core1: 10 + Math.random() * 30,
      core2: 20 + Math.random() * 40,
      core3: 15 + Math.random() * 25,
      core4: 8 + Math.random() * 18,
      core5: 12 + Math.random() * 28,
      core6: 6 + Math.random() * 15,
      core7: 18 + Math.random() * 35,
    };
  });
}
