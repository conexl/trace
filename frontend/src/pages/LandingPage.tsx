import type * as React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  ArrowRight,
  BrainCircuit,
  Check,
  Cloud,
  Network,
  RadioTower,
  ShieldCheck,
  TerminalSquare,
  Workflow,
} from 'lucide-react';
import { useAuth } from '@/lib/auth';
import { useI18n, type TranslationKey } from '@/lib/i18n';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

const controlFlow = [
  { labelKey: 'landing.flowAgent', valueKey: 'landing.flowAgentValue', icon: RadioTower },
  { labelKey: 'landing.flowCloud', valueKey: 'landing.flowCloudValue', icon: Cloud },
  { labelKey: 'landing.flowAction', valueKey: 'landing.flowActionValue', icon: TerminalSquare },
] satisfies Array<{
  labelKey: TranslationKey;
  valueKey: TranslationKey;
  icon: React.ComponentType<{ className?: string }>;
}>;

const proofCards = [
  {
    labelKey: 'landing.flowAgent',
    valueKey: 'landing.proofAgentValue',
    textKey: 'landing.proofAgentText',
  },
  {
    labelKey: 'landing.proofBackend',
    valueKey: 'landing.proofBackendValue',
    textKey: 'landing.proofBackendText',
  },
  {
    labelKey: 'landing.aiTitle',
    valueKey: 'landing.proofAiValue',
    textKey: 'landing.proofAiText',
  },
] satisfies Array<{ labelKey: TranslationKey; valueKey: TranslationKey; textKey: TranslationKey }>;

const demoSteps = [
  {
    titleKey: 'landing.stepConnectTitle',
    textKey: 'landing.stepConnectText',
  },
  {
    titleKey: 'landing.stepDetectTitle',
    textKey: 'landing.stepDetectText',
  },
  {
    titleKey: 'landing.stepActTitle',
    textKey: 'landing.stepActText',
  },
] satisfies Array<{ titleKey: TranslationKey; textKey: TranslationKey }>;

const featureCards = [
  {
    icon: ShieldCheck,
    titleKey: 'landing.trustTitle',
    textKey: 'landing.trustText',
  },
  {
    icon: Workflow,
    titleKey: 'landing.actionsTitle',
    textKey: 'landing.actionsText',
  },
  {
    icon: BrainCircuit,
    titleKey: 'landing.aiTitle',
    textKey: 'landing.aiText',
  },
] satisfies Array<{ icon: React.ComponentType<{ className?: string }>; titleKey: TranslationKey; textKey: TranslationKey }>;

const planCards = [
  {
    name: 'Free',
    price: '$0',
    descriptionKey: 'landing.freeDescription',
    featureKeys: ['landing.freeFeatureServer', 'landing.freeFeatureRetention', 'landing.freeFeatureIncidents', 'landing.freeFeatureReadonly'],
  },
  {
    name: 'Plus',
    price: '$12',
    descriptionKey: 'landing.plusDescription',
    featureKeys: [
      'landing.plusFeatureServers',
      'landing.plusFeatureRetention',
      'landing.plusFeatureTasks',
      'landing.plusFeatureActions',
      'landing.plusFeatureAi',
      'landing.plusFeatureTelegram',
    ],
  },
] satisfies Array<{ name: 'Free' | 'Plus'; price: string; descriptionKey: TranslationKey; featureKeys: TranslationKey[] }>;

function FadeIn({ children, delay = 0, className }: { children: React.ReactNode; delay?: number; className?: string }) {
  return <div className={className} style={delay ? { animationDelay: `${delay}s` } : undefined}>{children}</div>;
}

export function LandingPage() {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const { t } = useI18n();
  const plan = user?.subscription.plan ?? 'free';

  return (
    <main className="relative flex flex-1 flex-col overflow-hidden">
      <section className="relative px-6 pb-20 pt-14 sm:px-10 lg:pb-24 lg:pt-20">
        <div className="pointer-events-none absolute inset-x-0 top-0 h-[34rem] bg-[radial-gradient(circle_at_50%_0%,rgba(255,255,255,0.14),transparent_34rem)]" />
        <div className="pointer-events-none absolute left-1/2 top-0 h-px w-[78rem] -translate-x-1/2 bg-gradient-to-r from-transparent via-white/40 to-transparent" />

        <div className="relative mx-auto grid max-w-7xl gap-10 lg:grid-cols-12 lg:items-center">
          <div className="animate-page-in lg:col-span-7">
            <div className="mb-8 flex items-center gap-4">
              <img src="/logo.svg" alt="Trace logo" className="h-14 w-14 object-contain drop-shadow-[0_0_20px_rgba(255,255,255,0.22)]" />
              <div className="h-10 w-px bg-white/12" />
              <div>
                <p className="text-xs uppercase tracking-[0.24em] text-muted">{t('landing.product')}</p>
                <p className="mt-1 text-sm text-muted-soft">{t('landing.subproduct')}</p>
              </div>
            </div>

            <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.035] px-3 py-1 text-xs font-medium text-muted-soft">
              <ShieldCheck className="h-3.5 w-3.5 text-active" />
              {t('landing.badge')}
            </div>
            <h1 className="max-w-5xl text-balance text-5xl font-semibold tracking-[-0.07em] text-active sm:text-6xl lg:text-7xl">
              {t('landing.heroTitle')}
            </h1>
            <p className="mt-6 max-w-2xl text-base leading-8 text-muted-soft sm:text-lg">
              {t('landing.heroText')}
            </p>
            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Button variant="neon" size="lg" onClick={() => navigate(isAuthenticated ? '/servers' : '/servers/demo-server')} className="gap-2">
                {isAuthenticated ? t('common.openDashboard') : t('landing.demoCta')} <ArrowRight className="h-4 w-4" />
              </Button>
              <Button variant="default" size="lg" onClick={() => navigate(isAuthenticated ? '/billing' : '/register')} className="gap-2">
                {isAuthenticated ? t('common.managePlanCta') : t('common.createFreeAccount')} <ArrowRight className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div className="animate-page-in lg:col-span-5" style={{ animationDelay: '120ms' }}>
            <div className="relative overflow-hidden rounded-xl border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.065),rgba(255,255,255,0.02))] p-2 shadow-sm">
              <div className="absolute inset-x-10 top-0 h-px bg-gradient-to-r from-transparent via-white/50 to-transparent" />
              <div className="rounded-lg border border-white/10 bg-black/55 p-5">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-xs uppercase tracking-[0.24em] text-muted">{t('landing.liveLoop')}</p>
                    <h2 className="mt-2 text-2xl font-semibold tracking-[-0.04em] text-active">{t('landing.failureTitle')}</h2>
                  </div>
                  <div className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                    {t('landing.safePreview')}
                  </div>
                </div>

                <div className="mt-8 grid gap-3">
                  {controlFlow.map((item, index) => (
                    <div key={item.labelKey} className="group rounded-xl border border-white/10 bg-white/[0.035] p-4 transition-colors hover:border-white/20 hover:bg-white/[0.055]">
                      <div className="flex items-center justify-between gap-4">
                        <div className="flex items-center gap-3">
                          <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04]">
                            <item.icon className="h-4 w-4 text-active" />
                          </div>
                          <div>
                            <p className="text-sm font-semibold text-active">{t(item.labelKey)}</p>
                            <p className="mt-1 text-xs text-muted-soft">{t(item.valueKey)}</p>
                          </div>
                        </div>
                        <span className="font-mono text-xs text-muted">0{index + 1}</span>
                      </div>
                    </div>
                  ))}
                </div>

                <div className="mt-5 rounded-xl border border-white/10 bg-white/[0.035] p-4">
                  <div className="flex items-center justify-between text-xs uppercase tracking-[0.2em] text-muted">
                    <span>{t('landing.aiSummary')}</span>
                    <span>{t('landing.ago')}</span>
                  </div>
                  <p className="mt-4 text-sm leading-6 text-muted-soft">
                    {t('landing.aiSummaryText')}
                  </p>
                  <div className="mt-4 grid grid-cols-2 gap-2 text-xs">
                    <span className="rounded-lg border border-white/20 bg-white px-2 py-2 text-center font-semibold text-black">{t('landing.restart')}</span>
                    <span className="rounded-lg border border-white/10 bg-white/[0.04] px-2 py-2 text-center text-muted-soft">{t('landing.diagnostics')}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="border-y border-white/10 bg-white/[0.018] px-6 py-16 sm:px-10">
        <div className="mx-auto grid max-w-7xl gap-4 lg:grid-cols-3">
          {proofCards.map((card, index) => (
            <FadeIn key={card.labelKey} delay={index * 0.06}>
              <Card hover={false} className="h-full p-6 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                <p className="font-mono text-xs uppercase tracking-[0.2em] text-muted">{t(card.labelKey)}</p>
                <h2 className="mt-5 text-3xl font-semibold tracking-[-0.05em] text-active">{t(card.valueKey)}</h2>
                <p className="mt-3 text-sm leading-7 text-muted-soft">{t(card.textKey)}</p>
              </Card>
            </FadeIn>
          ))}
        </div>
      </section>

      <section className="px-6 py-24 sm:px-10">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-12 lg:items-start">
          <FadeIn className="lg:col-span-5">
            <p className="text-xs uppercase tracking-[0.24em] text-muted">{t('landing.workflowEyebrow')}</p>
            <h2 className="mt-4 max-w-xl text-4xl font-semibold tracking-[-0.05em] text-active sm:text-5xl">
              {t('landing.workflowTitle')}
            </h2>
            <p className="mt-5 max-w-xl text-sm leading-7 text-muted-soft">
              {t('landing.workflowText')}
            </p>
          </FadeIn>

          <div className="grid gap-4 lg:col-span-7">
            {demoSteps.map((step, index) => (
              <FadeIn key={step.titleKey} delay={index * 0.06}>
                <Card hover={false} className="p-5 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                  <div className="flex flex-col gap-5 sm:flex-row sm:items-start">
                    <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04] font-mono text-xs text-muted-soft">
                      0{index + 1}
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-active">{t(step.titleKey)}</h3>
                      <p className="mt-1 text-sm leading-6 text-muted-soft">{t(step.textKey)}</p>
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
              <p className="text-xs uppercase tracking-[0.24em] text-muted">{t('landing.engineEyebrow')}</p>
              <h2 className="mt-4 text-4xl font-semibold tracking-[-0.05em] text-active sm:text-5xl">{t('landing.engineTitle')}</h2>
            </div>
            <p className="max-w-xl text-sm leading-7 text-muted-soft">
              {isAuthenticated ? <>{t('landing.currentPlan')} <span className="font-semibold capitalize text-active">{plan}</span>. </> : t('landing.startVisibility')}
              {t('landing.plusUnlocks')}
            </p>
          </FadeIn>

          <div className="grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="grid gap-4">
              {featureCards.map((feature, index) => (
                <FadeIn key={feature.titleKey} delay={index * 0.06}>
                  <Card hover={false} className="p-5 transition-colors hover:border-white/20 hover:bg-white/[0.045]">
                    <div className="flex items-start gap-4">
                      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04]">
                        <feature.icon className="h-5 w-5 text-active" />
                      </div>
                      <div>
                        <h3 className="text-lg font-semibold text-active">{t(feature.titleKey)}</h3>
                        <p className="mt-1 text-sm leading-6 text-muted-soft">{t(feature.textKey)}</p>
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
                          <p className="text-xs uppercase tracking-[0.22em] text-muted">{plus ? t('landing.controlPlane') : t('landing.starter')}</p>
                          <h3 className="mt-4 text-3xl font-semibold text-active">{plus ? t('common.plus') : t('common.free')}</h3>
                        </div>
                        {plus && <div className="rounded-full border border-white/15 bg-white px-3 py-1 text-xs font-semibold text-black">{t('landing.bestValue')}</div>}
                      </div>
                      <div className="mt-8 flex items-end gap-2">
                        <span className="text-5xl font-semibold tracking-[-0.05em] text-active">{tier.price}</span>
                        <span className="pb-2 text-sm text-muted">{t('landing.perMonth')}</span>
                      </div>
                      <p className="mt-4 text-sm leading-6 text-muted-soft">{t(tier.descriptionKey)}</p>
                      <div className="my-6 h-px bg-white/10" />
                      <div className="grid gap-3">
                        {tier.featureKeys.map((featureKey) => (
                          <div key={featureKey} className="flex items-center gap-2 text-sm text-active">
                            <Check className="h-4 w-4 text-muted-soft" />
                            {t(featureKey)}
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
          <div className="mx-auto grid max-w-7xl gap-8 rounded-xl border border-white/10 bg-[linear-gradient(180deg,rgba(255,255,255,0.065),rgba(255,255,255,0.02))] p-8 shadow-sm sm:p-10 lg:grid-cols-[0.65fr_0.35fr] lg:items-center">
            <div>
              <div className="flex items-center gap-3">
                <Network className="h-5 w-5 text-active" />
                <p className="text-xs uppercase tracking-[0.24em] text-muted">{t('landing.boundaryEyebrow')}</p>
              </div>
              <h2 className="mt-6 max-w-2xl text-3xl font-semibold tracking-[-0.05em] text-active sm:text-4xl">
                {t('landing.boundaryTitle')}
              </h2>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-muted-soft">
                {t('landing.boundaryText')}
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row lg:flex-col">
              <Button variant="neon" size="lg" onClick={() => navigate(isAuthenticated ? '/servers' : '/register')}>
                {isAuthenticated ? t('common.openDashboard') : t('common.createFreeAccount')}
              </Button>
              <Link to="/billing" className="inline-flex h-12 items-center justify-center rounded-lg px-6 text-sm font-medium text-muted-soft transition-colors hover:text-active">
                {t('landing.comparePlans')}
              </Link>
            </div>
          </div>
        </FadeIn>
      </section>
    </main>
  );
}
