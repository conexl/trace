import * as React from 'react';
import { useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Activity, ArrowUpRight, Cpu, Globe, HardDrive, MemoryStick, Server, Settings } from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { useServer } from '@/hooks/useServer';
import { BackLink } from '@/components/BackLink';
import { Card } from '@/components/ui/Card';
import { UptimeDot } from '@/components/UptimeDot';
import { MetricChart } from '@/components/MetricChart';
import { ServicesPanel } from '@/components/ServicesPanel';
import { LogsDrawer } from '@/components/LogsDrawer';
import { PolicyModal, type ServicePolicy } from '@/components/PolicyModal';
import { AgentSettingsModal, type AgentConfig } from '@/components/AgentSettingsModal';
import { FaultToleranceOverlay } from '@/components/FaultToleranceOverlay';
import { DnsManagementDrawer, statusOf } from '@/components/DnsManagementDrawer';
import { DEMO_STATE, generateDemoHistory, type HistoryPoint } from '@/lib/demo';
import type { ServerState, ProcessSnapshot } from '@/lib/types';
import { cn, formatBytes, formatDuration } from '@/lib/utils';

const MAX_HISTORY = 60;

const fadeInUp = {
  initial: { opacity: 0, y: 16 },
  animate: { opacity: 1, y: 0 },
};

export function ServerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { isAuthenticated } = useAuth();
  const { data: liveData, loading, connected, reconnectIn } = useServer(id);
  const isDemo = !isAuthenticated;
  const data = isDemo ? DEMO_STATE : liveData;
  const [history, setHistory] = React.useState<HistoryPoint[]>(() =>
    isDemo ? generateDemoHistory() : []
  );
  const [localServices, setLocalServices] = React.useState<string[]>([]);
  const [hiddenServices, setHiddenServices] = React.useState<Set<string>>(new Set());
  const [availableServices, setAvailableServices] = React.useState<string[]>([]);
  const [servicePolicies, setServicePolicies] = React.useState<Record<string, ServicePolicy>>({});
  const [watchdogExpanded, setWatchdogExpanded] = React.useState(false);
  const [drawerService, setDrawerService] = React.useState<string | null>(null);
  const drawerOpen = drawerService !== null;
  const [policyService, setPolicyService] = React.useState<string | null>(null);
  const policyModalOpen = policyService !== null;
  const [agentSettingsOpen, setAgentSettingsOpen] = React.useState(false);
  const [agentConfig, setAgentConfig] = React.useState<AgentConfig>(() => ({
    name: data?.snapshot.agent_name ?? '',
    intervalSeconds: 60,
    logPaths: '/var/log/system.log\n/var/log/nginx/access.log',
    loggingLevel: 'INFO',
    updatePolicy: 'check',
    watchdogPollingSeconds: 10,
    watchdogTimeoutSeconds: 30,
    performanceMode: 'balanced',
    fanCurve: 'auto',
    sleepSchedule: { enabled: false, sleepAt: '23:00', wakeAt: '07:00' },
  }));

  React.useEffect(() => {
    if (!data || isDemo) return;
    setHistory((prev) => {
      const snapshot = data.snapshot;
      const point: HistoryPoint = {
        ts: Date.now(),
        cpu: snapshot.system.cpu_percent,
        ram: snapshot.system.memory.used_percent,
        swap: snapshot.system.memory.swap_percent,
      };
      snapshot.system.per_cpu_percent.forEach((value, idx) => {
        point[`core${idx}`] = value;
      });
      const next = [...prev, point];
      if (next.length > MAX_HISTORY) next.shift();
      return next;
    });
  }, [data, isDemo]);

  React.useEffect(() => {
    if (!data) return;
    if (isDemo) {
      const demo = [
        'redis',
        'mysql',
        'docker',
        'prometheus',
        'grafana',
        'etcd',
        'rabbitmq',
        'kafka',
        'mongodb',
        'elasticsearch',
        'nginx',
        'postgres',
      ];
      setAvailableServices(demo.filter((s) => !data.snapshot.processes.some((p) => p.name === s)));
      return;
    }
    setAvailableServices([]);
  }, [data, isDemo]);

  const mergedProcesses = React.useMemo<ProcessSnapshot[]>(() => {
    const existing = new Set(data?.snapshot.processes.map((p) => p.name) ?? []);
    const local: ProcessSnapshot[] = localServices
      .filter((name) => !existing.has(name) && !hiddenServices.has(name))
      .map((name) => ({
        name,
        match: name,
        service: '',
        running: false,
        status: 'down',
      }));
    return [
      ...(data?.snapshot.processes ?? []).filter((p) => !hiddenServices.has(p.name)),
      ...local,
    ];
  }, [data, localServices, hiddenServices]);

  const handleAddService = (name: string) => {
    setLocalServices((prev) => (prev.includes(name) ? prev : [...prev, name]));
    setHiddenServices((prev) => {
      const next = new Set(prev);
      next.delete(name);
      return next;
    });
  };

  const handleRemoveService = (name: string) => {
    setLocalServices((prev) => prev.filter((n) => n !== name));
    setHiddenServices((prev) => {
      const next = new Set(prev);
      next.add(name);
      return next;
    });
  };

  const handleViewLogs = (name: string) => {
    setDrawerService(name);
  };

  const handleEditPolicy = (name: string, _policy: ServicePolicy) => {
    setPolicyService(name);
  };

  const handleSavePolicy = (policy: ServicePolicy) => {
    if (!policyService) return;
    setServicePolicies((prev) => ({ ...prev, [policyService]: policy }));
    setPolicyService(null);
  };

  if (loading && !data) {
    return (
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
      </main>
    );
  }

  if (!data) {
    return (
      <main className="flex flex-1 flex-col items-center justify-center px-6 text-muted">
        <p>Server not found</p>
        <BackLink to="/servers" className="mt-4">
          Назад к узлам
        </BackLink>
      </main>
    );
  }

  return (
    <FaultToleranceOverlay connected={isDemo ? true : connected} reconnectIn={isDemo ? 0 : reconnectIn}>
      <main className="flex flex-1 flex-col px-6 py-6 min-h-0">
        <div className="mb-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <BackLink to="/servers">Назад к узлам</BackLink>
            <div className="hidden h-4 w-px bg-border sm:block" />
            <div className="flex items-center gap-2">
              <Server className="h-4 w-4 text-accent" />
              <h1 className="text-sm font-medium tracking-tight text-active">
                {data.summary.name}
              </h1>
              <UptimeDot status={data.summary.status} />
            </div>
            <motion.button
              onClick={() => setAgentSettingsOpen(true)}
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.96 }}
              className="ml-1 flex h-7 w-7 items-center justify-center rounded-md border border-border bg-canvas text-muted transition-colors hover:border-accent hover:text-accent"
              title="Agent settings"
            >
              <Settings className="h-3.5 w-3.5" />
            </motion.button>
          </div>

          {isDemo && (
            <span className="rounded-full border border-accent/30 bg-accent/10 px-2.5 py-1 font-mono text-[10px] uppercase text-accent">
              Demo mode
            </span>
          )}
        </div>

        <div className="grid grid-cols-12 gap-4 lg:grid-rows-[auto_1fr_1fr] flex-1 min-h-0">
          <StatusBlock data={data} />
          <CpuBlock history={history} />
          <RamBlock history={history} />
          <ServicesBlock
            data={data}
            processes={mergedProcesses}
            localServices={localServices}
            availableServices={availableServices}
            servicePolicies={servicePolicies}
            expanded={watchdogExpanded}
            onExpandedChange={setWatchdogExpanded}
            onAddService={handleAddService}
            onRemoveService={handleRemoveService}
            onViewLogs={handleViewLogs}
            onPolicyChange={handleEditPolicy}
          />
          <NetworkBlock data={data} />
        </div>
      </main>

      <LogsDrawer
        open={drawerOpen}
        onOpenChange={(open) => {
          if (!open) setDrawerService(null);
        }}
        serviceName={drawerService ?? ''}
        logs={data.snapshot.logs}
      />

      <PolicyModal
        open={policyModalOpen}
        onOpenChange={(open) => {
          if (!open) setPolicyService(null);
        }}
        serviceName={policyService ?? ''}
        policy={policyService ? servicePolicies[policyService] ?? { autoRestart: false, pin: false, cpuThreshold: 80, memoryThreshold: 80 } : { autoRestart: false, pin: false, cpuThreshold: 80, memoryThreshold: 80 }}
        onSave={handleSavePolicy}
      />

      <AgentSettingsModal
        open={agentSettingsOpen}
        onOpenChange={setAgentSettingsOpen}
        config={agentConfig}
        onSave={setAgentConfig}
      />
    </FaultToleranceOverlay>
  );
}

function StatusBlock({ data }: { data: ServerState }) {
  const { summary, snapshot } = data;
  const dnsMatches = snapshot.network.dns.filter((d) => d.matches_public_ip).length;

  const cards = [
    {
      icon: Server,
      label: 'Host',
      value: snapshot.host.hostname,
      sub: `${snapshot.host.os} · ${snapshot.host.platform}`,
    },
    {
      icon: Activity,
      label: 'Uptime',
      value: formatDuration(snapshot.host.uptime / 1_000_000),
      sub: summary.status,
      status: true,
    },
    {
      icon: Globe,
      label: 'Network',
      value: snapshot.network.public_ip || '—',
      sub: `DNS matches ${dnsMatches}/${snapshot.network.dns.length}`,
    },
    {
      icon: HardDrive,
      label: 'Disks',
      value: `${snapshot.system.disks.length} mounted`,
      sub: snapshot.system.disks
        .slice(0, 2)
        .map((d) => `${d.mountpoint} ${d.used_percent.toFixed(0)}%`)
        .join(' · '),
    },
  ];

  return (
    <div className="col-span-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {cards.map((card, idx) => (
        <motion.div
          key={card.label}
          variants={fadeInUp}
          initial="initial"
          animate="animate"
          transition={{ delay: idx * 0.05, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
        >
          <Card hover={false} className="p-4">
            <div className="flex items-center gap-2 text-muted">
              <card.icon className="h-4 w-4 text-accent" />
              <span className="text-xs font-mono uppercase">{card.label}</span>
            </div>
            <div className="mt-2 text-lg font-medium tracking-tight text-active truncate">
              {card.value}
            </div>
            <div className="mt-1 flex items-center gap-2 font-mono text-xs text-muted-soft">
              {card.status && <UptimeDot status={summary.status} />}
              <span className="truncate">{card.sub}</span>
            </div>
          </Card>
        </motion.div>
      ))}
    </div>
  );
}

function CpuBlock({ history }: { history: HistoryPoint[] }) {
  const cpuSeries = [
    { key: 'cpu', name: 'CPU %', color: '#00F576', fill: true },
    ...Array.from(
      {
        length: Math.max(
          0,
          history[0] ? Object.keys(history[0]).filter((k) => k.startsWith('core')).length : 0
        ),
      },
      (_, i) => ({
        key: `core${i}`,
        name: `Core ${i}`,
        color: `hsl(${140 + i * 25}, 80%, 60%)`,
        fill: false,
      })
    ),
  ];

  return (
    <motion.div
      variants={fadeInUp}
      initial="initial"
      animate="animate"
      transition={{ delay: 0.1, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
      className="col-span-12 lg:col-start-1 lg:col-span-6 lg:row-start-2 h-full min-h-[14rem]"
    >
      <Card hover={false} className="h-full p-4">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2 text-active">
            <Cpu className="h-4 w-4 text-accent" />
            <span className="text-sm font-medium tracking-tight">CPU load</span>
          </div>
          <span className="font-mono text-xs text-muted">
            {history.length > 0 ? `${history[history.length - 1].cpu.toFixed(1)}%` : '—'}
          </span>
        </div>
        <MetricChart
          data={history}
          series={cpuSeries}
          yDomain={[0, 100]}
          className="h-[calc(100%-2rem)]"
        />
      </Card>
    </motion.div>
  );
}

function RamBlock({ history }: { history: HistoryPoint[] }) {
  const ramSeries = [
    { key: 'ram', name: 'RAM %', color: '#00F576', fill: true },
    { key: 'swap', name: 'Swap %', color: '#38BDF8', fill: false },
  ];

  return (
    <motion.div
      variants={fadeInUp}
      initial="initial"
      animate="animate"
      transition={{ delay: 0.15, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
      className="col-span-12 lg:col-start-7 lg:col-span-6 lg:row-start-2 h-full min-h-[14rem]"
    >
      <Card hover={false} className="h-full p-4">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2 text-active">
            <MemoryStick className="h-4 w-4 text-accent" />
            <span className="text-sm font-medium tracking-tight">Memory / Swap</span>
          </div>
          <span className="font-mono text-xs text-muted">
            {history.length > 0 ? `${history[history.length - 1].ram.toFixed(1)}%` : '—'}
          </span>
        </div>
        <MetricChart
          data={history}
          series={ramSeries}
          yDomain={[0, 100]}
          className="h-[calc(100%-2rem)]"
        />
      </Card>
    </motion.div>
  );
}

function ServicesBlock({
  data,
  processes,
  localServices,
  availableServices,
  servicePolicies,
  expanded,
  onExpandedChange,
  onAddService,
  onRemoveService,
  onViewLogs,
  onPolicyChange,
}: {
  data: ServerState;
  processes: ProcessSnapshot[];
  localServices: string[];
  availableServices: string[];
  servicePolicies: Record<string, ServicePolicy>;
  expanded: boolean;
  onExpandedChange: (expanded: boolean) => void;
  onAddService: (name: string) => void;
  onRemoveService: (name: string) => void;
  onViewLogs: (name: string) => void;
  onPolicyChange: (name: string, policy: ServicePolicy) => void;
}) {
  return (
    <motion.div
      layout
      variants={fadeInUp}
      initial="initial"
      animate="animate"
      transition={{ layout: { duration: 0.4, ease: 'easeInOut' } }}
      style={{ willChange: 'transform, width, height' }}
      className={cn(
        'col-span-12 lg:col-start-1 h-full min-h-[14rem]',
        expanded ? 'lg:col-span-12 lg:row-start-2 lg:row-span-2 z-20' : 'lg:col-span-8 lg:row-start-3'
      )}
    >
      <Card hover={false} className="h-full p-4">
        <ServicesPanel
          serverId={data.summary.id}
          processes={processes}
          localServices={localServices}
          availableServices={availableServices}
          servicePolicies={servicePolicies}
          expanded={expanded}
          onExpandedChange={onExpandedChange}
          onAddService={onAddService}
          onRemoveService={onRemoveService}
          onViewLogs={onViewLogs}
          onPolicyChange={onPolicyChange}
        />
      </Card>
    </motion.div>
  );
}

function NetworkBlock({ data }: { data: ServerState }) {
  const { network } = data.snapshot;
  const [drawerOpen, setDrawerOpen] = React.useState(false);
  const mainIface = network.traffic.find((t) => !t.interface.startsWith('lo')) ?? network.traffic[0];
  const okCount = network.dns.filter((d) => statusOf(d) === 'ok').length;

  return (
    <>
      <motion.div
        variants={fadeInUp}
        initial="initial"
        animate="animate"
        transition={{ delay: 0.25, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
        className="col-span-12 lg:col-start-9 lg:col-span-4 lg:row-start-3 h-full min-h-[14rem]"
      >
        <Card hover={false} className="h-full p-4 flex flex-col">
          <div className="mb-4 flex items-center justify-between text-active">
            <div className="flex items-center gap-2">
              <Globe className="h-4 w-4 text-accent" />
              <span className="text-sm font-medium tracking-tight">Network telemetry</span>
            </div>
            <motion.button
              onClick={() => setDrawerOpen(true)}
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.98 }}
              className="group flex items-center gap-1.5 rounded-md border border-border bg-canvas px-2 py-1 text-[10px] text-muted transition-colors hover:border-accent hover:text-accent"
            >
              <span>DNS {okCount}/{network.dns.length}</span>
              <motion.span
                className="inline-block"
                initial={{ x: 0, y: 0 }}
                whileHover={{ x: 1, y: -1 }}
                transition={{ duration: 0.2 }}
              >
                <ArrowUpRight className="h-3 w-3 transition-transform group-hover:-translate-y-px group-hover:translate-x-px" />
              </motion.span>
            </motion.button>
          </div>

          <div className="space-y-4 flex-1 overflow-auto pr-1">
            <div className="grid grid-cols-2 gap-3">
              <motion.div
                whileHover={{ scale: 1.02 }}
                className="rounded-lg border border-border bg-canvas p-3"
              >
                <div className="text-[10px] font-mono uppercase text-muted">Rx</div>
                <div className="mt-1 font-mono text-sm text-active">
                  {mainIface ? formatBytes(mainIface.bytes_recv) : '—'}
                </div>
              </motion.div>
              <motion.div
                whileHover={{ scale: 1.02 }}
                className="rounded-lg border border-border bg-canvas p-3"
              >
                <div className="text-[10px] font-mono uppercase text-muted">Tx</div>
                <div className="mt-1 font-mono text-sm text-active">
                  {mainIface ? formatBytes(mainIface.bytes_sent) : '—'}
                </div>
              </motion.div>
            </div>

            <div>
              <div className="mb-2 text-[10px] font-mono uppercase text-muted">Speed tests</div>
              <div className="space-y-1.5">
                {network.speed_tests.length === 0 && (
                  <div className="text-xs text-muted">No speed tests configured</div>
                )}
                {network.speed_tests.map((test) => (
                  <div
                    key={test.name}
                    className="flex items-center justify-between font-mono text-xs"
                  >
                    <span className="truncate max-w-[120px] text-muted-soft">{test.name}</span>
                    <span className={cn('text-active', test.error && 'text-amber-muted')}>
                      {test.error ? 'failed' : `${test.mbps.toFixed(1)} Mbps`}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <div className="mb-2 text-[10px] font-mono uppercase text-muted">Listening ports</div>
              <div className="space-y-1 pr-1">
                {network.listening_ports.slice(0, 20).map((port, idx) => (
                  <div
                    key={`${port.address}-${port.port}-${idx}`}
                    className="flex items-center justify-between font-mono text-xs"
                  >
                    <span className="text-muted-soft">
                      {port.port}/{port.protocol}
                    </span>
                    <span className="truncate max-w-[120px] text-active">
                      {port.process || `pid ${port.pid}` || '—'}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Card>
      </motion.div>

      <DnsManagementDrawer
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
        dns={network.dns}
      />
    </>
  );
}
