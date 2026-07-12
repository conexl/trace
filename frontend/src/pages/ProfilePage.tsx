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
import { PageHeader } from '@/components/PageHeader';
import { useI18n, type TranslationKey } from '@/lib/i18n';
import type { PlanFeatures } from '@/lib/types';

const featureRows = [
  ['remote_tasks', 'profile.remoteTasks'],
  ['service_actions', 'profile.serviceActions'],
  ['ai_incident_analysis', 'profile.aiAnalysis'],
  ['telegram_notifications', 'profile.telegramAlerts'],
  ['config_management', 'profile.agentConfig'],
  ['audit_log', 'profile.auditLog'],
] satisfies Array<[keyof PlanFeatures, TranslationKey]>;

export function ProfilePage() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const { t } = useI18n();
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
      <section className="mx-auto grid w-full max-w-5xl gap-5 lg:grid-cols-[280px_1fr]">
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
                  <p className="text-sm font-semibold text-active">{t('profile.accountTitle')}</p>
                  <p className="mt-1 truncate font-mono text-xs text-muted-soft">{user?.email}</p>
                </div>
              </div>
            </div>

            <div className="space-y-3 p-5">
              <SummaryItem icon={isPlus ? Crown : CreditCard} label={t('profile.plan')} value={plan} active={isPlus} />
              <SummaryItem icon={Server} label={t('profile.nodeLimit')} value={`${user?.subscription.limits.max_servers ?? 1}`} />
              <SummaryItem icon={Bell} label={t('profile.metricRetention')} value={`${user?.subscription.limits.retention_hours ?? 24}h`} />
              <SummaryItem icon={ShieldCheck} label={t('profile.status')} value={t('common.active')} active />
            </div>

            <div className="grid gap-2 border-t border-white/10 p-5">
              <Button variant="default" size="md" onClick={() => navigate('/billing')} className="w-full gap-2">
                <CreditCard className="h-4 w-4" />
                {t('common.managePlan')}
              </Button>
              <Button variant="ghost" size="md" onClick={logout} className="w-full gap-2 text-muted hover:text-red-300">
                <LogOut className="h-4 w-4" />
                {t('common.logout')}
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
          <PageHeader
            title={t('profile.settingsTitle')}
            description={t('profile.settingsDescription')}
            eyebrow={t('profile.eyebrow')}
            className="mb-5"
          />

          <Card hover={false} className="divide-y divide-white/10 overflow-hidden border-white/10">
            <SettingsRow
              icon={User}
              title={t('profile.account')}
              description={t('profile.accountDescription')}
              action={<span className="truncate font-mono text-xs text-active">{user?.email}</span>}
            />

            <SettingsRow
              icon={CreditCard}
              title={t('profile.plan')}
              description={isPlus ? t('profile.planDescriptionPlus') : t('profile.planDescriptionFree')}
              action={
                <Button variant={isPlus ? 'default' : 'neon'} size="sm" onClick={() => navigate('/billing')} className="gap-2">
                  {isPlus ? <Crown className="h-4 w-4" /> : <Lock className="h-4 w-4" />}
                  {isPlus ? t('common.plusActive') : t('common.upgrade')}
                </Button>
              }
            />

            <SettingsRow
              icon={MessageCircle}
              title={t('profile.telegramTitle')}
              description={t('profile.telegramDescription')}
              action={
                isPlus ? (
                  <TelegramConnectButton />
                ) : (
                  <Button variant="default" size="sm" onClick={() => navigate('/billing')} className="gap-2">
                    <Crown className="h-4 w-4" />
                    {t('common.plusOnly')}
                  </Button>
                )
              }
            />

            <div className="p-5">
              <div className="flex items-start gap-4">
                <SectionIcon icon={ShieldCheck} />
                <div className="min-w-0 flex-1">
                  <h2 className="text-sm font-semibold text-active">{t('profile.featureAccess')}</h2>
                  <p className="mt-1 text-sm leading-6 text-muted-soft">
                    {t('profile.featureAccessDescription')}
                  </p>
                  <div className="mt-4 grid gap-2 sm:grid-cols-2">
                    {featureRows.map(([key, labelKey]) => {
                      const enabled = Boolean(user?.subscription.features[key]);
                      return (
                        <div
                          key={key}
                          className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.025] px-3 py-2"
                        >
                          <span className="text-sm text-muted-soft">{t(labelKey)}</span>
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

          <div className="mt-5 rounded-xl border border-white/10 bg-white/[0.025] p-4 text-xs leading-6 text-muted-soft">
            {t('profile.footer')}
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
