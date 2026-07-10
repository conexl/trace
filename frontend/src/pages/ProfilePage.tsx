import * as React from 'react';
import { motion } from 'framer-motion';
import {
  Bell,
  Check,
  CreditCard,
  Crown,
  ExternalLink,
  Lock,
  LogOut,
  Mail,
  MessageCircle,
  Server,
  ShieldCheck,
  User,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { TelegramConnectButton } from '@/components/TelegramConnectButton';

const featureLabels = [
  ['remote_tasks', 'Remote tasks'],
  ['service_actions', 'Service actions'],
  ['ai_incident_analysis', 'AI incident analysis'],
  ['telegram_notifications', 'Telegram notifications'],
  ['config_management', 'Agent config'],
  ['audit_log', 'Audit log'],
] as const;

export function ProfilePage() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const plan = user?.subscription.plan ?? 'free';
  const isPlus = plan === 'plus';

  React.useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  if (!isAuthenticated) {
    return (
      <main className="flex flex-1 items-center justify-center px-6">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-accent" />
      </main>
    );
  }

  return (
    <main className="flex flex-1 flex-col px-4 py-5 sm:px-6 lg:px-8">
      <section className="mx-auto w-full max-w-6xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
          className="overflow-hidden rounded-3xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.12),rgba(255,255,255,0.035)_36%,rgba(255,255,255,0.02))]"
        >
          <div className="grid gap-6 p-5 sm:p-6 lg:grid-cols-[1fr_auto] lg:items-end">
            <div className="min-w-0">
              <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-white/10 bg-black/20 px-3 py-1 text-xs font-medium text-muted-soft">
                <User className="h-3.5 w-3.5" />
                Account
              </div>
              <h1 className="text-3xl font-semibold tracking-[-0.06em] text-active sm:text-5xl">Profile settings</h1>
              <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-soft">
                Keep account, subscription and notification settings in one place, away from the operational dashboard.
              </p>
              <div className="mt-5 flex min-w-0 items-center gap-3 rounded-2xl border border-white/10 bg-black/20 p-3 sm:w-fit sm:pr-5">
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.055]">
                  <Mail className="h-5 w-5 text-muted-soft" />
                </div>
                <div className="min-w-0">
                  <p className="truncate font-mono text-sm text-active">{user?.email}</p>
                  <p className="mt-1 text-xs text-muted">Member workspace</p>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:w-[440px]">
              <Metric label="Plan" value={plan} icon={isPlus ? Crown : CreditCard} active={isPlus} />
              <Metric label="Nodes" value={`${user?.subscription.limits.max_servers ?? 1}`} icon={Server} />
              <Metric label="Retention" value={`${user?.subscription.limits.retention_hours ?? 24}h`} icon={Bell} />
              <Metric label="Status" value="Active" icon={ShieldCheck} active />
            </div>
          </div>
        </motion.div>

        <div className="mt-5 grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
          <motion.div
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.06, duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
            className="space-y-5"
          >
            <Card hover={false} className="p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-xs uppercase tracking-[0.2em] text-muted">Subscription</p>
                  <h2 className="mt-2 text-2xl font-semibold capitalize tracking-[-0.04em] text-active">{plan}</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-soft">
                    {isPlus
                      ? 'Plus is active. Automation, AI analysis and notifications are unlocked.'
                      : 'Free is active. Upgrade when this workspace needs automation and notifications.'}
                  </p>
                </div>
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04]">
                  {isPlus ? <Crown className="h-5 w-5 text-accent" /> : <Lock className="h-5 w-5 text-muted-soft" />}
                </div>
              </div>
              <Button variant="default" size="md" onClick={() => navigate('/billing')} className="mt-5 w-full gap-2">
                <CreditCard className="h-4 w-4" />
                Manage billing
              </Button>
            </Card>

            <Card hover={false} className="p-5">
              <p className="text-xs uppercase tracking-[0.2em] text-muted">Feature access</p>
              <div className="mt-4 space-y-2">
                {featureLabels.map(([key, label]) => {
                  const enabled = Boolean(user?.subscription.features[key]);
                  return (
                    <div key={key} className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.025] px-3 py-2">
                      <span className="text-sm text-muted-soft">{label}</span>
                      <span className={enabled ? 'text-emerald-200' : 'text-muted'}>
                        {enabled ? <Check className="h-4 w-4" /> : <Lock className="h-4 w-4" />}
                      </span>
                    </div>
                  );
                })}
              </div>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.12, duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
            className="space-y-5"
          >
            <Card hover={false} className="p-5">
              <div className="flex flex-col gap-5 md:flex-row md:items-start md:justify-between">
                <div>
                  <div className="flex items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-sky-300/20 bg-sky-300/10">
                      <MessageCircle className="h-5 w-5 text-sky-200" />
                    </div>
                    <div>
                      <p className="text-xs uppercase tracking-[0.2em] text-muted">Notifications</p>
                      <h2 className="mt-1 text-2xl font-semibold tracking-[-0.04em] text-active">Telegram</h2>
                    </div>
                  </div>
                  <p className="mt-4 max-w-xl text-sm leading-6 text-muted-soft">
                    Link a personal Telegram chat through a one-time bot start token. Incident notifications
                    are routed per account, not to a global chat.
                  </p>
                </div>

                {isPlus ? (
                  <div className="flex shrink-0 items-center gap-3 rounded-2xl border border-white/10 bg-white/[0.035] px-3 py-2">
                    <span className="text-xs text-muted-soft">Connection</span>
                    <TelegramConnectButton />
                  </div>
                ) : (
                  <Button variant="neon" size="sm" onClick={() => navigate('/billing')} className="shrink-0 gap-2">
                    <Crown className="h-4 w-4" />
                    Upgrade
                  </Button>
                )}
              </div>

              {!isPlus && (
                <div className="mt-5 rounded-2xl border border-amber-soft/25 bg-amber-soft/10 p-4 text-sm leading-6 text-amber-soft">
                  Telegram notifications are a Plus feature because they use the external notifications worker.
                </div>
              )}
            </Card>

            <Card hover={false} className="p-5">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-xs uppercase tracking-[0.2em] text-muted">Session</p>
                  <h2 className="mt-2 text-xl font-semibold tracking-[-0.04em] text-active">Signed in on this device</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-soft">
                    Log out here when presenting Trace on shared machines.
                  </p>
                </div>
                <Button variant="ghost" size="sm" onClick={logout} className="gap-2 text-muted hover:text-red-300">
                  <LogOut className="h-4 w-4" />
                  Log out
                </Button>
              </div>
            </Card>

            <button
              type="button"
              onClick={() => navigate('/servers')}
              className="flex w-full items-center justify-between rounded-2xl border border-white/10 bg-white/[0.025] px-4 py-3 text-left text-sm text-muted-soft transition-colors hover:border-white/20 hover:bg-white/[0.05] hover:text-active"
            >
              Back to operational dashboard
              <ExternalLink className="h-4 w-4" />
            </button>
          </motion.div>
        </div>
      </section>
    </main>
  );
}

function Metric({
  label,
  value,
  icon: Icon,
  active = false,
}: {
  label: string;
  value: string;
  icon: React.ComponentType<{ className?: string }>;
  active?: boolean;
}) {
  return (
    <div className="rounded-2xl border border-white/10 bg-black/20 p-3">
      <Icon className={active ? 'h-4 w-4 text-accent' : 'h-4 w-4 text-muted-soft'} />
      <p className="mt-3 text-[10px] uppercase tracking-[0.18em] text-muted">{label}</p>
      <p className="mt-1 truncate text-sm font-semibold capitalize text-active">{value}</p>
    </div>
  );
}
