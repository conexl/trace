import type * as React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  ArrowRight,
  BrainCircuit,
  Check,
  Cloud,
  Network,
  PlayCircle,
  RadioTower,
  ShieldCheck,
  TerminalSquare,
  Trophy,
  Workflow,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

const controlFlow = [
  { label: 'Agent', value: 'service control, watchdog, metrics', icon: RadioTower },
  { label: 'Cloud', value: 'identity, plans, incidents, AI', icon: Cloud },
  { label: 'Action', value: 'restart, diagnostics, tasks', icon: TerminalSquare },
];

const proofCards = [
  {
    label: 'Agent',
    value: 'Go binary',
    text: 'Collects host, network, hardware and service state close to the machine.',
  },
  {
    label: 'Backend',
    value: 'SaaS core',
    text: 'Auth, plans, incidents, metrics, remote actions and Telegram notification flow.',
  },
  {
    label: 'AI',
    value: 'Incident analyst',
    text: 'Turns watchdog failures and noisy telemetry into a readable operator summary.',
  },
];

const demoSteps = [
  {
    title: 'Connect a node',
    text: 'A one-time token binds a home server to the account, then the agent streams state securely.',
  },
  {
    title: 'Detect failure',
    text: 'Watchdog sees a service crash, records exit state and opens an incident with live context.',
  },
  {
    title: 'Act from browser',
    text: 'The operator restarts a service, runs diagnostics or escalates via Telegram without SSH.',
  },
];

const featureCards = [
  {
    icon: ShieldCheck,
    title: 'Trust model first',
    text: 'Account-bound agents, mTLS-ready transport and plan-aware permissions keep remote operations scoped.',
  },
  {
    icon: Workflow,
    title: 'Operational actions',
    text: 'Watchdog events can turn into restart flows, diagnostics, disabled actions or incident timelines.',
  },
  {
    icon: BrainCircuit,
    title: 'AI incident analyst',
    text: 'Trace summarizes crashes, correlates signals and proposes the next safe operator move.',
  },
];

const planCards = [
  {
    name: 'Free',
    price: '$0',
    description: 'Visibility for one home server.',
    features: ['1 connected server', '24h metric retention', 'Incidents list', 'Demo dashboard'],
  },
  {
    name: 'Plus',
    price: '$12',
    description: 'Control, AI and notifications for serious homelabs.',
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
      <section className="relative px-6 pb-20 pt-14 sm:px-10 lg:pb-24 lg:pt-20">
        <div className="pointer-events-none absolute inset-x-0 top-0 h-[34rem] bg-[radial-gradient(circle_at_50%_0%,rgba(255,255,255,0.14),transparent_34rem)]" />
        <div className="pointer-events-none absolute left-1/2 top-0 h-px w-[78rem] -translate-x-1/2 bg-gradient-to-r from-transparent via-white/40 to-transparent" />

        <div className="relative mx-auto grid max-w-7xl gap-10 lg:grid-cols-12 lg:items-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
            className="lg:col-span-7"
          >
            <div className="mb-8 flex items-center gap-4">
              <img src="/logo.svg" alt="Trace logo" className="h-14 w-14 object-contain drop-shadow-[0_0_20px_rgba(255,255,255,0.22)]" />
              <div className="h-10 w-px bg-white/12" />
              <div>
                <p className="text-xs uppercase tracking-[0.24em] text-muted">Hackathon product</p>
                <p className="mt-1 text-sm text-muted-soft">Micro SaaS for self-hosted infrastructure</p>
              </div>
            </div>

            <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 py-1 text-xs font-medium text-muted-soft">
              <Trophy className="h-3.5 w-3.5 text-active" />
              Jury-ready story: detect, explain, act
            </div>
            <h1 className="max-w-5xl text-balance text-5xl font-semibold tracking-[-0.07em] text-active sm:text-6xl lg:text-7xl">
              Turn any home server into a managed SaaS node.
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-8 text-muted-soft sm:text-lg">
              Trace gives hobby infrastructure the control loop teams expect in production: agent telemetry,
              service watchdog, incident intelligence, secure actions and plan-based SaaS access.
            </p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers/demo-server')} className="gap-2">
                Run jury demo <PlayCircle className="h-4 w-4" />
              </Button>
              <Button variant="default" size="lg" onClick={() => navigate('/servers')} className="gap-2">
                Open dashboard <ArrowRight className="h-4 w-4" />
              </Button>
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            transition={{ delay: 0.12, duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
            className="lg:col-span-5"
          >
            <div className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.085),rgba(255,255,255,0.025))] p-2 shadow-[0_30px_120px_rgba(0,0,0,0.48)]">
              <div className="absolute inset-x-10 top-0 h-px bg-gradient-to-r from-transparent via-white/50 to-transparent" />
              <div className="rounded-[1.55rem] border border-white/10 bg-black/55 p-5">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-xs uppercase tracking-[0.24em] text-muted">Live product loop</p>
                    <h2 className="mt-2 text-2xl font-semibold tracking-[-0.04em] text-active">Failure to fix, in one screen.</h2>
                  </div>
                  <div className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                    demo-safe
                  </div>
                </div>

                <div className="mt-8 grid gap-3">
                  {controlFlow.map((item, index) => (
                    <div key={item.label} className="group rounded-2xl border border-white/10 bg-white/[0.035] p-4 transition-colors hover:border-white/20 hover:bg-white/[0.055]">
                      <div className="flex items-center justify-between gap-4">
                        <div className="flex items-center gap-3">
                          <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04]">
                            <item.icon className="h-4 w-4 text-active" />
                          </div>
                          <div>
                            <p className="text-sm font-semibold text-active">{item.label}</p>
                            <p className="mt-1 text-xs text-muted-soft">{item.value}</p>
                          </div>
                        </div>
                        <span className="font-mono text-xs text-muted">0{index + 1}</span>
                      </div>
                    </div>
                  ))}
                </div>

                <div className="mt-5 rounded-2xl border border-white/10 bg-white/[0.035] p-4">
                  <div className="flex items-center justify-between text-xs uppercase tracking-[0.2em] text-muted">
                    <span>AI summary</span>
                    <span>37s ago</span>
                  </div>
                  <p className="mt-4 text-sm leading-6 text-muted-soft">
                    Redis crashed after config drift. Suggested action: restart service, then run diagnostics.
                  </p>
                  <div className="mt-4 grid grid-cols-2 gap-2 text-xs">
                    <span className="rounded-xl border border-white/20 bg-white px-2 py-2 text-center font-semibold text-black">Restart</span>
                    <span className="rounded-xl border border-white/10 bg-white/[0.04] px-2 py-2 text-center text-muted-soft">Diagnostics</span>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.018] px-6 py-16 sm:px-10">
        <div className="mx-auto grid max-w-7xl gap-4 lg:grid-cols-3">
          {proofCards.map((card, index) => (
            <FadeIn key={card.label} delay={index * 0.06}>
              <Card hover={false} className="h-full p-6 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                <p className="font-mono text-xs uppercase tracking-[0.2em] text-muted">{card.label}</p>
                <h2 className="mt-5 text-3xl font-semibold tracking-[-0.05em] text-active">{card.value}</h2>
                <p className="mt-3 text-sm leading-7 text-muted-soft">{card.text}</p>
              </Card>
            </FadeIn>
          ))}
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-12 lg:items-start">
          <FadeIn className="lg:col-span-5">
            <p className="text-xs uppercase tracking-[0.24em] text-muted">Demo narrative</p>
            <h2 className="mt-4 max-w-xl text-4xl font-semibold tracking-[-0.05em] text-active sm:text-5xl">
              A judge can understand it in three clicks.
            </h2>
            <p className="mt-5 max-w-xl text-sm leading-7 text-muted-soft">
              The homepage explains why Trace exists. The demo node proves the workflow. The dashboard
              is where a real user operates their machines.
            </p>
          </FadeIn>

          <div className="grid gap-4 lg:col-span-7">
            {demoSteps.map((step, index) => (
              <FadeIn key={step.title} delay={index * 0.06}>
                <Card hover={false} className="p-5 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                  <div className="flex flex-col gap-5 sm:flex-row sm:items-start">
                    <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04] font-mono text-xs text-muted-soft">
                      0{index + 1}
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-active">{step.title}</h3>
                      <p className="mt-1 text-sm leading-6 text-muted-soft">{step.text}</p>
                    </div>
                  </div>
                </Card>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.018] px-6 py-20 sm:px-10">
        <div className="mx-auto max-w-7xl">
          <FadeIn className="mb-10 flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
            <div>
              <p className="text-xs uppercase tracking-[0.24em] text-muted">Micro SaaS engine</p>
              <h2 className="mt-4 text-4xl font-semibold tracking-[-0.05em] text-active sm:text-5xl">Free visibility. Paid control.</h2>
            </div>
            <p className="max-w-xl text-sm leading-7 text-muted-soft">
              Current account: <span className="font-semibold capitalize text-active">{plan}</span>. Plus unlocks remote actions, AI analysis and notifications.
            </p>
          </FadeIn>

          <div className="grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="grid gap-4">
              {featureCards.map((feature, index) => (
                <FadeIn key={feature.title} delay={index * 0.06}>
                  <Card hover={false} className="p-5 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                    <div className="flex items-start gap-4">
                      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04]">
                        <feature.icon className="h-5 w-5 text-active" />
                      </div>
                      <div>
                        <h3 className="text-lg font-semibold text-active">{feature.title}</h3>
                        <p className="mt-1 text-sm leading-6 text-muted-soft">{feature.text}</p>
                      </div>
                    </div>
                  </Card>
                </FadeIn>
              ))}
            </div>

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
                      <p className="mt-4 text-sm leading-6 text-muted-soft">{tier.description}</p>
                      <div className="my-6 h-px bg-white/10" />
                      <div className="grid gap-3">
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
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <FadeIn>
          <div className="mx-auto grid max-w-7xl gap-8 rounded-[2rem] border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.085),rgba(255,255,255,0.025))] p-8 shadow-[0_30px_110px_rgba(0,0,0,0.42)] sm:p-10 lg:grid-cols-[0.65fr_0.35fr] lg:items-center">
            <div>
              <div className="flex items-center gap-3">
                <Network className="h-5 w-5 text-active" />
                <p className="text-xs uppercase tracking-[0.24em] text-muted">Product boundary</p>
              </div>
              <h2 className="mt-6 max-w-2xl text-3xl font-semibold tracking-[-0.05em] text-active sm:text-4xl">
                Home sells the vision. Dashboard runs the machines.
              </h2>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-soft">
                That separation keeps the pitch sharp for judges and the product usable for operators.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
              <Button variant="neon" size="lg" onClick={() => navigate('/servers/demo-server')}>Run jury demo</Button>
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
