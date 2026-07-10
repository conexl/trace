import type * as React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Activity,
  ArrowRight,
  BellRing,
  Check,
  Cpu,
  Crown,
  Globe2,
  LockKeyhole,
  Radar,
  Server,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
  Workflow,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

const heroStats = [
  { label: 'agent latency', value: '<1s' },
  { label: 'secure transport', value: 'mTLS' },
  { label: 'incident actions', value: '4' },
];

const productModules = [
  {
    icon: Radar,
    title: 'Observability cockpit',
    text: 'CPU, RAM, disks, network, DNS, ports and process health streamed from a tiny Go agent.',
  },
  {
    icon: Workflow,
    title: 'Control plane',
    text: 'Run allowlisted tasks, restart remote-controllable services and roll forward agent configs.',
  },
  {
    icon: Sparkles,
    title: 'AI incident copilot',
    text: 'Turn crashes and watchdog events into readable root-cause summaries and next actions.',
  },
  {
    icon: BellRing,
    title: 'SaaS notifications',
    text: 'Telegram links are bound to the user account, so alerts follow the right operator.',
  },
];

const pricing = [
  {
    name: 'Free',
    price: '$0',
    badge: 'Start here',
    text: 'Read-only monitoring for a single home server.',
    features: ['1 server', '24h metrics', 'Alerts and incidents', 'Demo dashboard'],
  },
  {
    name: 'Plus',
    price: '$12',
    badge: 'Control plane',
    text: 'For homelabs that need automation, AI triage and notifications.',
    features: ['10 servers', '30 day retention', 'Remote tasks', 'Service actions', 'AI analysis', 'Telegram alerts'],
  },
];

function FadeIn({ children, delay = 0, className }: { children: React.ReactNode; delay?: number; className?: string }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.25 }}
      transition={{ duration: 0.55, delay, ease: [0.22, 1, 0.36, 1] }}
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
    <main className="flex flex-1 flex-col overflow-hidden">
      <section className="relative px-6 pb-20 pt-12 sm:px-10 lg:pb-28">
        <div className="absolute left-1/2 top-10 h-72 w-72 -translate-x-1/2 rounded-full bg-accent/10 blur-3xl" />
        <div className="mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[1.05fr_0.95fr]">
          <motion.div
            initial={{ opacity: 0, y: 22 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
            className="relative z-10"
          >
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-accent/25 bg-accent/10 px-3 py-1 text-xs font-semibold text-accent">
              <ShieldCheck className="h-3.5 w-3.5" />
              SaaS control plane for home servers
            </div>
            <h1 className="max-w-4xl text-balance text-5xl font-bold tracking-[-0.055em] text-active sm:text-6xl lg:text-7xl">
              Your homelab, packaged like a production platform.
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-8 text-muted-soft">
              Trace watches the machine, explains incidents, and lets you act without SSH. Free gives you
              visibility. Plus turns the dashboard into an operator console.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers')} className="gap-2">
                Open dashboard <ArrowRight className="h-4 w-4" />
              </Button>
              <Button variant="default" size="lg" onClick={() => navigate('/billing')} className="gap-2">
                Compare plans <Crown className="h-4 w-4" />
              </Button>
            </div>
            <div className="mt-10 grid max-w-xl grid-cols-3 gap-3">
              {heroStats.map((stat) => (
                <div key={stat.label} className="rounded-2xl border border-white/10 bg-white/[0.035] p-4">
                  <p className="font-mono text-xl text-active">{stat.value}</p>
                  <p className="mt-1 text-[10px] uppercase tracking-[0.16em] text-muted">{stat.label}</p>
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, scale: 0.96, y: 18 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            transition={{ delay: 0.15, duration: 0.65, ease: [0.22, 1, 0.36, 1] }}
            className="relative z-10"
          >
            <div className="rounded-[2rem] border border-white/10 bg-[linear-gradient(145deg,rgba(19,34,41,0.92),rgba(7,16,19,0.96))] p-4 shadow-[0_40px_120px_rgba(0,0,0,0.48)]">
              <div className="rounded-[1.5rem] border border-white/10 bg-canvas/80 p-4">
                <div className="flex items-center justify-between border-b border-white/10 pb-4">
                  <div>
                    <p className="text-xs uppercase tracking-[0.18em] text-muted">Live node</p>
                    <h2 className="mt-1 text-xl font-semibold text-active">m1-homebase</h2>
                  </div>
                  <div className="rounded-full border border-accent/35 bg-accent/10 px-3 py-1 text-xs text-accent">online</div>
                </div>
                <div className="mt-4 grid grid-cols-2 gap-3">
                  <MetricCard icon={Cpu} label="CPU" value="18%" tone="cyan" />
                  <MetricCard icon={Activity} label="RAM" value="42%" tone="amber" />
                  <MetricCard icon={Globe2} label="DNS" value="8/8" tone="cyan" />
                  <MetricCard icon={TerminalSquare} label="Tasks" value="Plus" tone="amber" />
                </div>
                <div className="mt-4 rounded-2xl border border-white/10 bg-white/[0.035] p-4">
                  <div className="flex items-center gap-2 text-sm text-active">
                    <Sparkles className="h-4 w-4 text-accent" />
                    AI incident summary
                  </div>
                  <p className="mt-3 text-sm leading-6 text-muted-soft">
                    Watchdog detected a failed process. Likely root cause: config drift after restart.
                    Suggested action: run diagnostics, then restart the service.
                  </p>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.025] px-6 py-16 sm:px-10">
        <div className="mx-auto grid max-w-7xl gap-4 md:grid-cols-2 lg:grid-cols-4">
          {productModules.map((module, idx) => (
            <FadeIn key={module.title} delay={idx * 0.06}>
              <Card hover={false} className="h-full p-5">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-accent/20 bg-accent/10">
                  <module.icon className="h-5 w-5 text-accent" />
                </div>
                <h3 className="mt-5 text-lg font-semibold text-active">{module.title}</h3>
                <p className="mt-2 text-sm leading-6 text-muted-soft">{module.text}</p>
              </Card>
            </FadeIn>
          ))}
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <div className="mx-auto max-w-7xl">
          <FadeIn className="mb-10 flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
            <div>
              <p className="text-xs uppercase tracking-[0.22em] text-accent">Pricing</p>
              <h2 className="mt-3 max-w-3xl text-4xl font-semibold tracking-tight text-active">
                Free visibility. Paid control.
              </h2>
            </div>
            <p className="max-w-xl text-sm leading-6 text-muted-soft">
              Current account: <span className="font-semibold capitalize text-active">{plan}</span>. Upgrade is wired
              as a demo checkout endpoint, ready to be replaced by Stripe or Paddle.
            </p>
          </FadeIn>
          <div className="grid gap-5 lg:grid-cols-2">
            {pricing.map((tier, idx) => {
              const plus = tier.name === 'Plus';
              return (
                <FadeIn key={tier.name} delay={idx * 0.08}>
                  <Card hover={false} className={plus ? 'h-full border-accent/35 p-6 shadow-accent-glow' : 'h-full p-6'}>
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <div className="rounded-full border border-white/10 bg-white/[0.04] px-3 py-1 text-xs text-muted">
                          {tier.badge}
                        </div>
                        <h3 className="mt-5 text-3xl font-semibold text-active">{tier.name}</h3>
                      </div>
                      {plus ? <Crown className="h-6 w-6 text-accent" /> : <LockKeyhole className="h-6 w-6 text-muted" />}
                    </div>
                    <div className="mt-8 flex items-end gap-2">
                      <span className="text-5xl font-semibold text-active">{tier.price}</span>
                      <span className="pb-2 text-sm text-muted">/mo</span>
                    </div>
                    <p className="mt-4 text-sm leading-6 text-muted-soft">{tier.text}</p>
                    <div className="my-6 h-px bg-white/10" />
                    <div className="grid gap-3 sm:grid-cols-2">
                      {tier.features.map((feature) => (
                        <div key={feature} className="flex items-center gap-2 text-sm text-active">
                          <Check className="h-4 w-4 text-accent" />
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

      <section className="px-6 pb-24 sm:px-10">
        <FadeIn>
          <div className="mx-auto max-w-5xl rounded-[2rem] border border-white/10 bg-[linear-gradient(135deg,rgba(104,225,253,0.12),rgba(255,180,84,0.08))] p-8 text-center sm:p-12">
            <Server className="mx-auto h-8 w-8 text-accent" />
            <h2 className="mt-5 text-3xl font-semibold tracking-tight text-active">Ready to demo the product?</h2>
            <p className="mx-auto mt-3 max-w-2xl text-sm leading-6 text-muted-soft">
              Open the mock node for a sponsor-friendly walkthrough, or sign in and connect a real agent.
            </p>
            <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers/demo-server')}>Open demo node</Button>
              <Link to="/register" className="inline-flex h-12 items-center justify-center rounded-lg px-6 text-sm font-medium text-active hover:text-accent">
                Create account
              </Link>
            </div>
          </div>
        </FadeIn>
      </section>
    </main>
  );
}

function MetricCard({ icon: Icon, label, value, tone }: { icon: React.ElementType; label: string; value: string; tone: 'cyan' | 'amber' }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/[0.035] p-4">
      <div className="flex items-center justify-between">
        <Icon className={tone === 'cyan' ? 'h-4 w-4 text-accent' : 'h-4 w-4 text-amber-soft'} />
        <span className="text-[10px] uppercase tracking-[0.16em] text-muted">{label}</span>
      </div>
      <p className="mt-5 font-mono text-2xl text-active">{value}</p>
    </div>
  );
}
