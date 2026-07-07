import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Card } from '@/components/ui/Card';
import { UptimeDot } from '@/components/UptimeDot';
import type { ServerSummary } from '@/lib/types';
import { formatDuration } from '@/lib/utils';

interface ServerCardProps {
  server: ServerSummary;
  index?: number;
}

export function ServerCard({ server, index = 0 }: ServerCardProps) {
  const navigate = useNavigate();
  const hasPendingConfig = server.applied_config_revision < server.desired_config_revision;

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.35, ease: 'easeOut' }}
    >
      <Card
        className="h-40 p-5 flex flex-col justify-between"
        onClick={() => navigate(`/servers/${encodeURIComponent(server.id)}`)}
      >
        <div className="flex items-start justify-between">
          <div>
            <h3 className="text-lg font-medium tracking-tight text-active">
              {server.name || server.hostname}
            </h3>
            <div className="flex items-center gap-2">
              <p className="font-mono text-xs text-muted">{server.platform}</p>
              {server.version && (
                <span className="rounded bg-surface-elevated px-1 py-0.5 font-mono text-[10px] text-muted-soft">
                  v{server.version}
                </span>
              )}
            </div>
          </div>
          <UptimeDot status={server.status} />
        </div>

        {hasPendingConfig && (
          <div className="mt-2 px-2 py-0.5 bg-yellow-500/10 border border-yellow-500/20 rounded text-[10px] text-yellow-500 font-mono uppercase tracking-wider">
            Pending Config Update
          </div>
        )}

        <div className="space-y-2">
          <div className="flex items-center justify-between font-mono text-xs">
            <span className="text-muted-soft">uptime</span>
            <span className="text-active font-mono-nums">
              {server.status === 'online' ? '99.9%' : '—'}
            </span>
          </div>
          <div className="flex items-center justify-between font-mono text-xs">
            <span className="text-muted-soft">ip</span>
            <span className="text-active truncate max-w-[140px]">
              {server.public_ip || '—'}
            </span>
          </div>
          <div className="flex items-center justify-between font-mono text-xs">
            <span className="text-muted-soft">last seen</span>
            <span className="text-active">
              {server.last_seen ? formatDuration(Date.now() - new Date(server.last_seen).getTime()) + ' ago' : '—'}
            </span>
          </div>
        </div>
      </Card>
    </motion.div>
  );
}
