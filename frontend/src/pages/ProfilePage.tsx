import * as React from 'react';
import { motion } from 'framer-motion';
import { Bell, CreditCard, Crown, LogOut, Mail, MessageCircle, ShieldCheck, User } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { TelegramConnectButton } from '@/components/TelegramConnectButton';

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
    <main className="flex flex-1 flex-col px-6 py-8 sm:px-8">
      <section className="mx-auto w-full max-w-6xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
          className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"
        >
          <div>
            <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 py-1 text-xs font-medium text-muted-soft">
              <User className="h-3.5 w-3.5" />
              Account settings
            </div>
            <h1 className="text-3xl font-semibold tracking-[-0.05em] text-active sm:text-4xl">Profile</h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-soft">
              Manage your workspace identity, plan limits and notification channels.
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={logout} className="w-fit gap-2 text-muted hover:text-red-300">
            <LogOut className="h-4 w-4" />
            Log out
          </Button>
        </motion.div>

        <div className="grid gap-5 lg:grid-cols-[0.95fr_1.05fr]">
          <motion.div
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.05, duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
          >
            <Card hover={false} className="p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-xs uppercase tracking-[0.2em] text-muted">Signed in as</p>
                  <div className="mt-4 flex items-center gap-3">
                    <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.045]">
                      <Mail className="h-5 w-5 text-muted-soft" />
                    </div>
                    <div className="min-w-0">
                      <p className="truncate font-mono text-sm text-active">{user?.email}</p>
                      <p className="mt-1 text-xs text-muted">Member account</p>
                    </div>
                  </div>
                </div>
                <div className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                  Active
                </div>
              </div>

              <div className="mt-6 grid gap-3 sm:grid-cols-2">
                <div className="rounded-xl border border-border bg-canvas/70 p-4">
                  <p className="text-xs text-muted">Plan</p>
                  <p className="mt-2 flex items-center gap-2 text-lg font-semibold capitalize text-active">
                    {isPlus ? <Crown className="h-4 w-4 text-accent" /> : <CreditCard className="h-4 w-4 text-muted-soft" />}
                    {plan}
                  </p>
                </div>
                <div className="rounded-xl border border-border bg-canvas/70 p-4">
                  <p className="text-xs text-muted">Server limit</p>
                  <p className="mt-2 font-mono text-lg text-active">{user?.subscription.limits.max_servers ?? 1}</p>
                </div>
                <div className="rounded-xl border border-border bg-canvas/70 p-4">
                  <p className="text-xs text-muted">Metric retention</p>
                  <p className="mt-2 font-mono text-lg text-active">{user?.subscription.limits.retention_hours ?? 24}h</p>
                </div>
                <div className="rounded-xl border border-border bg-canvas/70 p-4">
                  <p className="text-xs text-muted">Remote actions</p>
                  <p className="mt-2 text-lg font-semibold text-active">{user?.subscription.features.remote_tasks ? 'Enabled' : 'Locked'}</p>
                </div>
              </div>

              <div className="mt-5 flex flex-wrap gap-2">
                <Button variant="default" size="sm" onClick={() => navigate('/billing')} className="gap-2">
                  <CreditCard className="h-4 w-4" />
                  Manage plan
                </Button>
                <Button variant="ghost" size="sm" onClick={() => navigate('/servers')} className="gap-2">
                  <ShieldCheck className="h-4 w-4" />
                  Back to nodes
                </Button>
              </div>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.12, duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
            className="space-y-5"
          >
            <Card hover={false} className="overflow-hidden p-5">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-sky-300/20 bg-sky-300/10">
                      <MessageCircle className="h-4 w-4 text-sky-200" />
                    </div>
                    <div>
                      <p className="text-xs uppercase tracking-[0.2em] text-muted">Notifications</p>
                      <h2 className="mt-1 text-xl font-semibold tracking-[-0.04em] text-active">Telegram</h2>
                    </div>
                  </div>
                  <p className="mt-4 max-w-xl text-sm leading-6 text-muted-soft">
                    Connect a Telegram chat to receive incident notifications when services go down,
                    watchdog restarts fail, or Trace needs operator attention.
                  </p>
                </div>

                {isPlus ? (
                  <div className="flex items-center gap-3 rounded-xl border border-white/10 bg-white/[0.035] px-3 py-2">
                    <span className="text-xs text-muted-soft">Status</span>
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
                <div className="mt-5 rounded-xl border border-amber-soft/25 bg-amber-soft/10 p-4 text-sm leading-6 text-amber-soft">
                  Telegram notifications are included in Plus because they depend on the external notification worker.
                </div>
              )}
            </Card>

            <Card hover={false} className="p-5">
              <div className="flex items-start gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 bg-white/[0.035]">
                  <Bell className="h-4 w-4 text-muted-soft" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold tracking-[-0.04em] text-active">Notification policy</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-soft">
                    Trace currently sends Telegram alerts for incident lifecycle events. Email and webhook
                    channels are ready to fit here later without crowding the dashboard header.
                  </p>
                </div>
              </div>
            </Card>
          </motion.div>
        </div>
      </section>
    </main>
  );
}
