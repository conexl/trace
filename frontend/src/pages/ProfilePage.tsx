import * as React from 'react';
import { motion } from 'framer-motion';
import {
  Bell,
  Check,
  CreditCard,
  Crown,
  Lock,
  LogOut,
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

const featureRows = [
  ['remote_tasks', 'Remote tasks'],
  ['service_actions', 'Service actions'],
  ['ai_incident_analysis', 'AI analysis'],
  ['telegram_notifications', 'Telegram alerts'],
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
      <section className="mx-auto grid w-full max-w-6xl gap-5 lg:grid-cols-[320px_1fr]">
        <motion.aside
          initial={{ opacity: 0, y: 14 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.38, ease: [0.22, 1, 0.36, 1] }}
          className="lg:sticky lg:top-20 lg:self-start"
        >
          <Card hover={false} className="overflow-hidden border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.055),rgba(255,255,255,0.025))]">
            <div className="border-b border-white/10 p-5">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-black/30">
                  <img src="/logo.svg" alt="Trace" className="h-7 w-7 object-contain" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-semibold text-active">Trace account</p>
                  <p className="mt-1 truncate font-mono text-xs text-muted-soft">{user?.email}</p>
                </div>
              </div>
            </div>

            <div className="space-y-3 p-5">
              <SummaryItem icon={isPlus ? Crown : CreditCard} label="Plan" value={plan} active={isPlus} />
              <SummaryItem icon={Server} label="Node limit" value={`${user?.subscription.limits.max_servers ?? 1}`} />
              <SummaryItem icon={Bell} label="Metric retention" value={`${user?.subscription.limits.retention_hours ?? 24}h`} />
              <SummaryItem icon={ShieldCheck} label="Status" value="Active" active />
            </div>

            <div className="grid gap-2 border-t border-white/10 p-5">
              <Button variant="default" size="md" onClick={() => navigate('/billing')} className="w-full gap-2">
                <CreditCard className="h-4 w-4" />
                Manage plan
              </Button>
              <Button variant="ghost" size="md" onClick={logout} className="w-full gap-2 text-muted hover:text-red-300">
                <LogOut className="h-4 w-4" />
                Log out
              </Button>
            </div>
          </Card>
        </motion.aside>

        <motion.div
          initial={{ opacity: 0, y: 14 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.05, duration: 0.38, ease: [0.22, 1, 0.36, 1] }}
          className="min-w-0"
        >
          <div className="mb-5">
            <p className="text-xs uppercase tracking-[0.22em] text-muted">Settings</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-[-0.05em] text-active sm:text-4xl">Profile</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-soft">
              Account identity, plan limits and notification channels for this workspace.
            </p>
          </div>

          <Card hover={false} className="divide-y divide-white/10 overflow-hidden border-white/10">
            <SettingsRow
              icon={User}
              title="Account"
              description="The email used to sign in and receive workspace ownership context."
              action={<span className="truncate font-mono text-xs text-active">{user?.email}</span>}
            />

            <SettingsRow
              icon={CreditCard}
              title="Plan"
              description={isPlus ? 'Plus features are active for this workspace.' : 'Free plan is active. Upgrade when you need actions and alerts.'}
              action={
                <Button variant={isPlus ? 'default' : 'neon'} size="sm" onClick={() => navigate('/billing')} className="gap-2">
                  {isPlus ? <Crown className="h-4 w-4" /> : <Lock className="h-4 w-4" />}
                  {isPlus ? 'Plus active' : 'Upgrade'}
                </Button>
              }
            />

            <SettingsRow
              icon={MessageCircle}
              title="Telegram notifications"
              description="Connect a personal Telegram chat for incident notifications. Tokens are account-scoped."
              action={
                isPlus ? (
                  <TelegramConnectButton />
                ) : (
                  <Button variant="default" size="sm" onClick={() => navigate('/billing')} className="gap-2">
                    <Crown className="h-4 w-4" />
                    Plus only
                  </Button>
                )
              }
            />

            <div className="p-5">
              <div className="flex items-start gap-4">
                <SectionIcon icon={ShieldCheck} />
                <div className="min-w-0 flex-1">
                  <h2 className="text-sm font-semibold text-active">Feature access</h2>
                  <p className="mt-1 text-sm leading-6 text-muted-soft">
                    Current capabilities unlocked by the active plan.
                  </p>
                  <div className="mt-4 grid gap-2 sm:grid-cols-2">
                    {featureRows.map(([key, label]) => {
                      const enabled = Boolean(user?.subscription.features[key]);
                      return (
                        <div
                          key={key}
                          className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.025] px-3 py-2"
                        >
                          <span className="text-sm text-muted-soft">{label}</span>
                          <span className={enabled ? 'text-emerald-200' : 'text-muted'}>
                            {enabled ? <Check className="h-4 w-4" /> : <Lock className="h-4 w-4" />}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            </div>
          </Card>

          <div className="mt-5 rounded-2xl border border-white/10 bg-white/[0.025] p-4 text-xs leading-6 text-muted-soft">
            Trace keeps high-risk operational controls in the dashboard, while this page stays focused on account and notification settings.
          </div>
        </motion.div>
      </section>
    </main>
  );
}

function SummaryItem({
  icon: Icon,
  label,
  value,
  active = false,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  active?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.025] px-3 py-2">
      <div className="flex items-center gap-2">
        <Icon className={active ? 'h-4 w-4 text-accent' : 'h-4 w-4 text-muted-soft'} />
        <span className="text-xs text-muted">{label}</span>
      </div>
      <span className="truncate text-sm font-medium capitalize text-active">{value}</span>
    </div>
  );
}

function SettingsRow({
  icon,
  title,
  description,
  action,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description: string;
  action: React.ReactNode;
}) {
  return (
    <div className="grid gap-4 p-5 md:grid-cols-[1fr_auto] md:items-center">
      <div className="flex min-w-0 items-start gap-4">
        <SectionIcon icon={icon} />
        <div className="min-w-0">
          <h2 className="text-sm font-semibold text-active">{title}</h2>
          <p className="mt-1 max-w-xl text-sm leading-6 text-muted-soft">{description}</p>
        </div>
      </div>
      <div className="min-w-0 md:max-w-xs md:justify-self-end">{action}</div>
    </div>
  );
}

function SectionIcon({ icon: Icon }: { icon: React.ComponentType<{ className?: string }> }) {
  return (
    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04]">
      <Icon className="h-4 w-4 text-muted-soft" />
    </div>
  );
}
