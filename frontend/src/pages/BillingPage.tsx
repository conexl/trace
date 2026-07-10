import * as React from 'react';
import { motion } from 'framer-motion';
import { Check, CreditCard, Crown, Lock, Server, Sparkles, Zap } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { updateBillingPlan } from '@/lib/api';
import { useAuth } from '@/lib/auth';
import type { SubscriptionPlan } from '@/lib/types';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { cn } from '@/lib/utils';

const plans: {
  plan: SubscriptionPlan;
  name: string;
  price: string;
  description: string;
  badge: string;
  features: string[];
}[] = [
  {
    plan: 'free',
    name: 'Free',
    price: '$0',
    description: 'A clean read-only control plane for one node.',
    badge: 'Start',
    features: ['1 connected node', '24h metrics retention', 'Alerts and incident list', 'Read-only dashboard'],
  },
  {
    plan: 'plus',
    name: 'Plus',
    price: '$12',
    description: 'Automation, notifications and AI triage for serious homelabs.',
    badge: 'Most useful',
    features: [
      '10 connected nodes',
      '30 day metrics retention',
      'Remote tasks and service actions',
      'AI incident analysis',
      'Telegram incident notifications',
      'Agent config management',
    ],
  },
];

export function BillingPage() {
  const { user, isAuthenticated, refreshUser } = useAuth();
  const navigate = useNavigate();
  const [pendingPlan, setPendingPlan] = React.useState<SubscriptionPlan | null>(null);
  const [error, setError] = React.useState('');
  const currentPlan = user?.subscription.plan ?? 'free';

  const changePlan = async (plan: SubscriptionPlan) => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    setError('');
    setPendingPlan(plan);
    try {
      await updateBillingPlan(plan);
      await refreshUser();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Plan update failed');
    } finally {
      setPendingPlan(null);
    }
  };

  return (
    <main className="flex flex-1 flex-col px-4 py-5 sm:px-6 lg:px-8">
      <section className="mx-auto w-full max-w-6xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
          className="mb-5 overflow-hidden rounded-3xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.12),rgba(255,255,255,0.035)_38%,rgba(255,255,255,0.02))] p-5 sm:p-6"
        >
          <div className="grid gap-6 lg:grid-cols-[1fr_360px] lg:items-end">
            <div>
              <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-accent/25 bg-accent/10 px-3 py-1 text-xs font-medium text-accent">
                <Sparkles className="h-3.5 w-3.5" />
                Billing
              </div>
              <h1 className="max-w-3xl text-4xl font-semibold tracking-[-0.06em] text-active sm:text-5xl">
                Choose the amount of control this workspace needs.
              </h1>
              <p className="mt-4 max-w-2xl text-sm leading-7 text-muted-soft">
                Free is for observing one node. Plus unlocks actions that can change infrastructure state:
                remote execution, service actions, AI analysis and Telegram notifications.
              </p>
            </div>

            <Card hover={false} className="border-white/10 bg-black/20 p-4">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="text-xs uppercase tracking-[0.2em] text-muted">Current plan</p>
                  <h2 className="mt-2 text-2xl font-semibold capitalize text-active">{currentPlan}</h2>
                </div>
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.05]">
                  {currentPlan === 'plus' ? (
                    <Crown className="h-5 w-5 text-accent" />
                  ) : (
                    <Lock className="h-5 w-5 text-muted-soft" />
                  )}
                </div>
              </div>
              <div className="mt-4 grid grid-cols-2 gap-3">
                <div className="rounded-xl border border-white/10 bg-white/[0.025] p-3">
                  <Server className="h-4 w-4 text-muted-soft" />
                  <p className="mt-2 text-xs text-muted">Nodes</p>
                  <p className="mt-1 font-mono text-lg text-active">{user?.subscription.limits.max_servers ?? 1}</p>
                </div>
                <div className="rounded-xl border border-white/10 bg-white/[0.025] p-3">
                  <CreditCard className="h-4 w-4 text-muted-soft" />
                  <p className="mt-2 text-xs text-muted">Retention</p>
                  <p className="mt-1 font-mono text-lg text-active">{user?.subscription.limits.retention_hours ?? 24}h</p>
                </div>
              </div>
            </Card>
          </div>
        </motion.div>

        {error && (
          <div className="mb-5 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        )}

        <div className="grid gap-5 lg:grid-cols-2">
          {plans.map((plan, idx) => {
            const active = currentPlan === plan.plan;
            const plus = plan.plan === 'plus';
            return (
              <motion.div
                key={plan.plan}
                initial={{ opacity: 0, y: 18 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.08, duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
              >
                <Card
                  hover={false}
                  className={cn(
                    'flex min-h-[460px] flex-col p-5 sm:p-6',
                    plus && 'border-accent/35 bg-[linear-gradient(140deg,rgba(5,16,18,0.96),rgba(11,13,19,0.92))] shadow-accent-glow'
                  )}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="mb-4 inline-flex rounded-full border border-white/10 bg-white/[0.04] px-3 py-1 text-xs text-muted-soft">
                        {plan.badge}
                      </div>
                      <h2 className="text-3xl font-semibold tracking-[-0.05em] text-active">{plan.name}</h2>
                    </div>
                    {active && (
                      <span className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                        Active
                      </span>
                    )}
                  </div>

                  <div className="mt-7 flex items-end gap-2">
                    <span className="text-5xl font-semibold tracking-[-0.06em] text-active">{plan.price}</span>
                    <span className="pb-2 text-sm text-muted">/ month</span>
                  </div>
                  <p className="mt-4 text-sm leading-6 text-muted-soft">{plan.description}</p>

                  <div className="my-6 h-px bg-border" />
                  <div className="space-y-3">
                    {plan.features.map((feature) => (
                      <div key={feature} className="flex items-center gap-3 text-sm text-active">
                        <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-accent/10 text-accent">
                          <Check className="h-3 w-3" />
                        </span>
                        {feature}
                      </div>
                    ))}
                  </div>

                  <div className="mt-auto pt-8">
                    <Button
                      variant={plus ? 'neon' : 'default'}
                      size="lg"
                      className="w-full gap-2"
                      disabled={active || pendingPlan === plan.plan}
                      onClick={() => changePlan(plan.plan)}
                    >
                      {plus && <Zap className="h-4 w-4" />}
                      {active ? 'Current plan' : pendingPlan === plan.plan ? 'Updating...' : plus ? 'Upgrade to Plus' : 'Switch to Free'}
                    </Button>
                  </div>
                </Card>
              </motion.div>
            );
          })}
        </div>
      </section>
    </main>
  );
}
