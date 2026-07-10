import type * as React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  ArrowRight,
  BadgeCheck,
  Bot,
  Check,
  Cloud,
  LockKeyhole,
  RadioTower,
  Sparkles,
  TerminalSquare,
  Workflow,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

const operatingLayers = [
  {
    icon: RadioTower,
    title: 'Agent on the node',
    text: 'Collects process, service, network, hardware and watchdog signals close to the machine.',
  },
  {
    icon: Cloud,
    title: 'SaaS control plane',
    text: 'Stores telemetry, incidents, identities, plans and secure actions without exposing SSH.',
  },
  {
    icon: TerminalSquare,
    title: 'Dashboard for action',
    text: 'Operators open the dashboard only when they need live state, diagnostics or service control.',
  },
];

const homeVsDashboard = [
  {
    label: 'Home',
    title: 'Explain the product',
    text: 'Value proposition, security model, pricing and why Trace exists.',
  },
  {
    label: 'Dashboard',
    title: 'Operate live systems',
    text: 'Servers, incidents, service actions, logs, tasks, metrics and agent status.',
  },
  {
    label: 'Demo',
    title: 'Sponsor walkthrough',
    text: 'A safe preloaded node for showing the flow without touching production data.',
  },
];

const planCards = [
  {
    name: 'Free',
    price: '$0',
    text: 'One node, read-only monitoring and incident visibility for a small homelab.',
    features: ['1 connected server', '24h metric retention', 'Incidents list', 'Demo dashboard'],
  },
  {
    name: 'Plus',
    price: '$12',
    text: 'Remote control, AI incident analysis and notifications for serious home infrastructure.',
    features: ['10 connected servers', '30 day retention', 'Remote tasks', 'Service actions', 'AI analysis', 'Telegram alerts'],
  },
];

function FadeIn({ children, delay = 0, className }: { children: React.ReactNode; delay?: number; className?: string }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 18 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.25 }}
      transition={{ duration: 0.5, delay, ease: [0.22, 1, 0.36, 1] }}
      className={className}
    >
      {children}
    </motion.div>
  );
}

export function LandingPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const plan = user?.subscription.plan ?? 'free';

  return (
    <main className="relative flex flex-1 flex-col overflow-hidden">
      <section className="relative px-6 pb-20 pt-16 sm:px-10 lg:pb-28 lg:pt-20">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(255,255,255,0.12),transparent_34rem)]" />
        <div className="pointer-events-none absolute left-1/2 top-0 h-px w-[72rem] -translate-x-1/2 bg-gradient-to-r from-transparent via-white/30 to-transparent" />

        <div className="relative mx-auto max-w-6xl text-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.96 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1] }}
            className="mx-auto mb-7 flex h-16 w-16 items-center justify-center rounded-2xl border border-white/15 bg-white p-3 shadow-[0_20px_70px_rgba(255,255,255,0.12)]"
          >
            <img src="/logo.svg" alt="Trace logo" className="h-full w-full object-contain" />
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.08, duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
          >
            <div className="mx-auto mb-6 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-3 py-1 text-xs font-medium text-muted-soft shadow-[inset_0_1px_0_rgba(255,255,255,0.08)]">
              <Sparkles className="h-3.5 w-3.5 text-active" />
              Production-grade control for home servers
            </div>
            <h1 className="mx-auto max-w-5xl text-balance text-5xl font-semibold tracking-[-0.06em] text-active sm:text-6xl lg:text-7xl">
              A SaaS cockpit for infrastructure you actually own.
            </h1>
            <p className="mx-auto mt-6 max-w-2xl text-base leading-8 text-muted-soft sm:text-lg">
              Trace turns a Mac mini, NUC or Raspberry Pi into a managed node: secure telemetry,
              incidents, watchdog actions, AI analysis and remote operations without opening SSH.
            </p>
            <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers')} className="gap-2">
                Open dashboard <ArrowRight className="h-4 w-4" />
              </Button>
              <Button variant="default" size="lg" onClick={() => navigate('/servers/demo-server')} className="gap-2">
                View demo node
              </Button>
            </div>
          </motion.div>

          <FadeIn delay={0.18} className="mx-auto mt-14 max-w-5xl">
            <div className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-white/[0.035] p-2 shadow-[0_30px_120px_rgba(0,0,0,0.45)]">
              <div className="absolute inset-x-12 top-0 h-px bg-gradient-to-r from-transparent via-white/40 to-transparent" />
              <div className="grid gap-2 md:grid-cols-3">
                {operatingLayers.map((layer, index) => (
                  <div key={layer.title} className="rounded-[1.5rem] border border-white/10 bg-canvas/70 p-5 text-left">
                    <div className="mb-8 flex items-center justify-between">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04]">
                        <layer.icon className="h-4 w-4 text-active" />
                      </div>
                      <span className="font-mono text-xs text-muted">0{index + 1}</span>
                    </div>
                    <h2 className="text-base font-semibold text-active">{layer.title}</h2>
                    <p className="mt-2 text-sm leading-6 text-muted-soft">{layer.text}</p>
                  </div>
                ))}
              </div>
            </div>
          </FadeIn>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.02] px-6 py-20 sm:px-10">
        <div className="mx-auto grid max-w-6xl gap-8 lg:grid-cols-[0.85fr_1.15fr] lg:items-center">
          <FadeIn>
            <p className="text-xs uppercase tracking-[0.24em] text-muted">Clear separation</p>
            <h2 className="mt-4 max-w-xl text-4xl font-semibold tracking-[-0.04em] text-active">
              The homepage sells the promise. The dashboard runs the machines.
            </h2>
            <p className="mt-5 max-w-xl text-sm leading-7 text-muted-soft">
              Sponsors should understand the business in seconds. Operators should not hunt through a
              landing page to restart a service. Trace keeps those flows intentionally separate.
            </p>
          </FadeIn>

          <div className="grid gap-3">
            {homeVsDashboard.map((item, index) => (
              <FadeIn key={item.label} delay={index * 0.06}>
                <Card hover={false} className="group p-5 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-start gap-4">
                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04] font-mono text-xs text-muted-soft">
                        {item.label.slice(0, 1)}
                      </div>
                      <div>
                        <p className="text-xs uppercase tracking-[0.2em] text-muted">{item.label}</p>
                        <h3 className="mt-1 text-lg font-semibold text-active">{item.title}</h3>
                        <p className="mt-1 text-sm leading-6 text-muted-soft">{item.text}</p>
                      </div>
                    </div>
                    {index === 1 && <BadgeCheck className="hidden h-5 w-5 text-active sm:block" />}
                  </div>
                </Card>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <div className="mx-auto max-w-6xl">
          <FadeIn className="mb-10 flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
            <div>
              <p className="text-xs uppercase tracking-[0.24em] text-muted">Why it can become SaaS</p>
              <h2 className="mt-4 max-w-3xl text-4xl font-semibold tracking-[-0.04em] text-active">
                A small agent unlocks recurring value.
              </h2>
            </div>
            <p className="max-w-xl text-sm leading-7 text-muted-soft">
              The free plan proves visibility. Plus is where control, automation, notifications and AI triage
              become paid operator features.
            </p>
          </FadeIn>

          <div className="grid gap-4 lg:grid-cols-3">
            <FeatureCard icon={LockKeyhole} title="Secure by default" text="Account binding, mTLS-ready transport and plan-aware access keep remote operations controlled." />
            <FeatureCard icon={Workflow} title="Actionable operations" text="Watchdog events can become restart flows, diagnostics, disabled actions or incident timelines." />
            <FeatureCard icon={Bot} title="AI incident analyst" text="Trace can summarize crashes, correlate metrics and propose the next safe action for the operator." />
          </div>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.02] px-6 py-20 sm:px-10">
        <div className="mx-auto max-w-6xl">
          <FadeIn className="mb-10 flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
            <div>
              <p className="text-xs uppercase tracking-[0.24em] text-muted">Pricing</p>
              <h2 className="mt-4 text-4xl font-semibold tracking-[-0.04em] text-active">Free visibility. Paid control.</h2>
            </div>
            <p className="max-w-xl text-sm leading-7 text-muted-soft">
              Current account: <span className="font-semibold capitalize text-active">{plan}</span>. The billing page owns plan details and upgrades.
            </p>
          </FadeIn>

          <div className="grid gap-5 lg:grid-cols-2">
            {planCards.map((tier, index) => {
              const plus = tier.name === 'Plus';
              return (
                <FadeIn key={tier.name} delay={index * 0.08}>
                  <Card hover={false} className={plus ? 'h-full border-white/25 bg-white/[0.055] p-6' : 'h-full p-6'}>
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <p className="text-xs uppercase tracking-[0.22em] text-muted">{plus ? 'Control plane' : 'Starter'}</p>
                        <h3 className="mt-4 text-3xl font-semibold text-active">{tier.name}</h3>
                      </div>
                      {plus && <div className="rounded-full border border-white/15 bg-white px-3 py-1 text-xs font-semibold text-black">Best demo</div>}
                    </div>
                    <div className="mt-8 flex items-end gap-2">
                      <span className="text-5xl font-semibold tracking-[-0.05em] text-active">{tier.price}</span>
                      <span className="pb-2 text-sm text-muted">/mo</span>
                    </div>
                    <p className="mt-4 text-sm leading-6 text-muted-soft">{tier.text}</p>
                    <div className="my-6 h-px bg-white/10" />
                    <div className="grid gap-3 sm:grid-cols-2">
                      {tier.features.map((feature) => (
                        <div key={feature} className="flex items-center gap-2 text-sm text-active">
                          <Check className="h-4 w-4 text-muted-soft" />
                          {feature}
                        </div>
                      ))}
                    </div>
                  </Card>
                </FadeIn>
              );
            })}
          </div>
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <FadeIn>
          <div className="mx-auto max-w-5xl rounded-[2rem] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.08),rgba(255,255,255,0.025))] p-8 text-center shadow-[0_30px_110px_rgba(0,0,0,0.42)] sm:p-12">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl border border-white/15 bg-white p-2.5">
              <img src="/logo.svg" alt="Trace logo" className="h-full w-full object-contain" />
            </div>
            <h2 className="mt-6 text-3xl font-semibold tracking-[-0.04em] text-active">Ready for the operational view?</h2>
            <p className="mx-auto mt-3 max-w-2xl text-sm leading-7 text-muted-soft">
              Use the dashboard for real servers. Use the demo node when you need a clean sponsor walkthrough.
            </p>
            <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers')}>Open dashboard</Button>
              <Link to="/billing" className="inline-flex h-12 items-center justify-center rounded-lg px-6 text-sm font-medium text-muted-soft transition-colors hover:text-active">
                Compare plans
              </Link>
            </div>
          </div>
        </FadeIn>
      </section>
    </main>
  );
}

function FeatureCard({ icon: Icon, title, text }: { icon: React.ElementType; title: string; text: string }) {
  return (
    <FadeIn>
      <Card hover={false} className="h-full p-6 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
        <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04]">
          <Icon className="h-5 w-5 text-active" />
        </div>
        <h3 className="mt-6 text-lg font-semibold text-active">{title}</h3>
        <p className="mt-2 text-sm leading-6 text-muted-soft">{text}</p>
      </Card>
    </FadeIn>
  );
}
