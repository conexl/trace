import { useNavigate, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Activity,
  Shield,
  Workflow,
  BarChart3,
  Globe,
  Server,
  Settings,
  Zap,
  Lock,
  Terminal,
  Cpu,
  HardDrive,
  MemoryStick,
  ChevronRight,
} from 'lucide-react';
import { NeonButton } from '@/components/NeonButton';
import { Card } from '@/components/ui/Card';

const features = [
  {
    icon: Activity,
    title: 'Легковесный Go-агент',
    description: 'Сбор метрик без нагрузки на систему. Работает в фоне и буферизирует данные.',
  },
  {
    icon: Shield,
    title: 'Безопасность mTLS',
    description: 'Защищенный канал связи, работающий за NAT. One-time pairing token.',
  },
  {
    icon: Workflow,
    title: 'Контроль процессов',
    description: 'Управление службами и логами в реальном времени. systemd и launchd.',
  },
];

const highlights = [
  {
    icon: BarChart3,
    title: 'CPU, RAM и диски в одном окне',
    description:
      'Bento-дашборд собирает ключевые метрики узла: загрузка процессора, потребление памяти, сетевой трафик и состояние дисков. Всё обновляется в реальном времени через WebSocket.',
    stats: [
      { label: 'CPU cores', value: 'up to 64' },
      { label: 'Update latency', value: '< 1s' },
    ],
  },
  {
    icon: Globe,
    title: 'DNS Management Hub',
    description:
      'Управляй десятками доменов: группируй по проектам, фильтруй по статусу, отслеживай latency и помечай критичные домены для алертов.',
    stats: [
      { label: 'Groups', value: 'unlimited' },
      { label: 'History', value: '24h' },
    ],
  },
  {
    icon: Settings,
    title: 'Power & Agent Configuration',
    description:
      'Настраивай режимы производительности, расписание сна, уровень логирования, политику обновлений и стратегию Watchdog прямо из интерфейса.',
    stats: [
      { label: 'Modes', value: '3' },
      { label: 'Log levels', value: '4' },
    ],
  },
];

const steps = [
  { icon: Terminal, title: '1. Установи агента', description: 'Один бинарник и pairing token — без зависимостей.' },
  { icon: Server, title: '2. Добавь узел', description: 'Введи имя сервера и получи сертификат mTLS.' },
  { icon: Zap, title: '3. Мониторь', description: 'Смотри метрики, логи и статус служб в дашборде.' },
];

const stats = [
  { value: '∞', label: 'Nodes' },
  { value: '<1s', label: 'Latency' },
  { value: 'mTLS', label: 'By default' },
];

function FadeIn({
  children,
  delay = 0,
  className,
}: {
  children: React.ReactNode;
  delay?: number;
  className?: string;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 28 }}
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

  return (
    <main className="flex flex-1 flex-col">
      {/* Hero */}
      <section className="flex min-h-[90vh] flex-col items-center justify-center px-6 py-16">
        <motion.div
          initial={{ opacity: 0, y: 20, filter: 'blur(4px)' }}
          animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
          transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
          className="max-w-3xl text-center"
        >
          <h1 className="text-balance text-4xl font-medium tracking-tight text-active sm:text-5xl md:text-6xl">
            Управляй домашним сервером без боли
          </h1>
          <p className="mx-auto mt-6 max-w-xl text-base text-muted sm:text-lg">
            Минималистичный дашборд для мониторинга узлов, метрик и служб. Всё в одном Bento-интерфейсе.
          </p>
        </motion.div>

        <div className="mt-12 grid w-full max-w-4xl grid-cols-1 gap-4 sm:grid-cols-3">
          {features.map((feature, idx) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 + idx * 0.1, duration: 0.4, ease: 'easeOut' }}
            >
              <Card hover className="h-full p-5 text-left">
                <feature.icon className="h-5 w-5 text-accent" />
                <h3 className="mt-4 text-sm font-medium tracking-tight text-active">{feature.title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-muted">{feature.description}</p>
              </Card>
            </motion.div>
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.4, duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
          className="mt-14"
        >
          <NeonButton onClick={() => navigate('/servers')}>Начать мониторинг</NeonButton>
        </motion.div>
      </section>

      {/* Stats strip */}
      <section className="border-y border-border bg-canvas/30 px-6 py-10">
        <div className="mx-auto flex max-w-4xl flex-wrap items-center justify-center gap-8 sm:gap-16">
          {stats.map((s, idx) => (
            <FadeIn key={s.label} delay={idx * 0.1}>
              <div className="text-center">
                <div className="text-3xl font-medium tracking-tight text-active">{s.value}</div>
                <div className="mt-1 text-xs font-mono uppercase text-muted">{s.label}</div>
              </div>
            </FadeIn>
          ))}
        </div>
      </section>

      {/* Feature highlights */}
      <section className="px-6 py-24">
        <div className="mx-auto max-w-5xl space-y-24">
          {highlights.map((item, idx) => (
            <div
              key={item.title}
              className={`flex flex-col items-center gap-8 md:flex-row ${
                idx % 2 === 1 ? 'md:flex-row-reverse' : ''
              }`}
            >
              <FadeIn className="flex-1" delay={0}>
                <div className="relative overflow-hidden rounded-2xl border border-border bg-canvas p-6 shadow-xl">
                  <div className="absolute -right-10 -top-10 h-32 w-32 rounded-full bg-accent/5 blur-2xl" />
                  <div className="relative space-y-3">
                    <div className="flex items-center gap-3 text-active">
                      <item.icon className="h-5 w-5 text-accent" />
                      <span className="text-sm font-medium tracking-tight">{item.title}</span>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      {item.stats.map((stat) => (
                        <Card key={stat.label} className="p-3">
                          <div className="text-xs text-muted">{stat.label}</div>
                          <div className="mt-1 font-mono text-sm text-active">{stat.value}</div>
                        </Card>
                      ))}
                    </div>
                    <div className="flex gap-2 pt-2">
                      <div className="h-1.5 flex-1 rounded-full bg-accent/20" />
                      <div className="h-1.5 flex-1 rounded-full bg-accent/40" />
                      <div className="h-1.5 flex-1 rounded-full bg-accent/60" />
                    </div>
                  </div>
                </div>
              </FadeIn>
              <FadeIn className="flex-1" delay={0.15}>
                <div className="space-y-4">
                  <div className="inline-flex items-center gap-2 rounded-full border border-border bg-canvas px-3 py-1 text-[10px] font-mono uppercase text-muted">
                    <Lock className="h-3 w-3 text-accent" />
                    {idx === 0 ? 'Real-time' : idx === 1 ? 'Management' : 'Control'}
                  </div>
                  <h3 className="text-2xl font-medium tracking-tight text-active">{item.title}</h3>
                  <p className="leading-relaxed text-muted">{item.description}</p>
                  <Link
                    to="/servers/demo-server"
                    className="inline-flex items-center gap-1 text-sm text-accent transition-opacity hover:opacity-80"
                  >
                    Посмотреть демо <ChevronRight className="h-3.5 w-3.5" />
                  </Link>
                </div>
              </FadeIn>
            </div>
          ))}
        </div>
      </section>

      {/* Dashboard preview */}
      <section className="bg-canvas/30 px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <FadeIn>
            <div className="mb-10 text-center">
              <h2 className="text-2xl font-medium tracking-tight text-active">Bento-дашборд</h2>
              <p className="mt-2 text-sm text-muted">Все метрики, службы и сеть — в одном окне.</p>
            </div>
          </FadeIn>
          <FadeIn delay={0.15}>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <Card className="col-span-1 flex flex-col justify-between p-4 lg:col-span-2">
                <div className="flex items-center gap-2 text-active">
                  <Cpu className="h-4 w-4 text-accent" />
                  <span className="text-xs font-medium">CPU load</span>
                </div>
                <div className="mt-8 h-16 rounded-md bg-canvas" />
              </Card>
              <Card className="col-span-1 flex flex-col justify-between p-4">
                <div className="flex items-center gap-2 text-active">
                  <MemoryStick className="h-4 w-4 text-accent" />
                  <span className="text-xs font-medium">RAM</span>
                </div>
                <div className="mt-4 text-2xl font-medium text-active">45%</div>
              </Card>
              <Card className="col-span-1 flex flex-col justify-between p-4">
                <div className="flex items-center gap-2 text-active">
                  <HardDrive className="h-4 w-4 text-accent" />
                  <span className="text-xs font-medium">Disk</span>
                </div>
                <div className="mt-4 text-2xl font-medium text-active">58%</div>
              </Card>
              <Card className="col-span-1 p-4 lg:col-span-2">
                <div className="mb-3 flex items-center gap-2 text-active">
                  <Workflow className="h-4 w-4 text-accent" />
                  <span className="text-xs font-medium">Watchdog</span>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between rounded-md bg-canvas px-3 py-2 text-xs">
                    <span className="text-muted">nginx</span>
                    <span className="text-accent">running</span>
                  </div>
                  <div className="flex items-center justify-between rounded-md bg-canvas px-3 py-2 text-xs">
                    <span className="text-muted">postgres</span>
                    <span className="text-accent">running</span>
                  </div>
                </div>
              </Card>
              <Card className="col-span-1 flex flex-col justify-between p-4 lg:col-span-2">
                <div className="flex items-center gap-2 text-active">
                  <Globe className="h-4 w-4 text-accent" />
                  <span className="text-xs font-medium">Network telemetry</span>
                </div>
                <div className="mt-6 flex items-center justify-center gap-2 text-2xl font-medium text-active">
                  DNS: 6/8 OK
                </div>
              </Card>
            </div>
          </FadeIn>
        </div>
      </section>

      {/* Getting started */}
      <section className="px-6 py-24">
        <div className="mx-auto max-w-4xl">
          <FadeIn>
            <div className="mb-12 text-center">
              <h2 className="text-2xl font-medium tracking-tight text-active">Начало работы</h2>
              <p className="mt-2 text-sm text-muted">Три простых шага до полного контроля.</p>
            </div>
          </FadeIn>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            {steps.map((step, idx) => (
              <FadeIn key={step.title} delay={idx * 0.1}>
                <Card className="h-full p-5">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-border bg-canvas">
                    <step.icon className="h-4 w-4 text-accent" />
                  </div>
                  <h3 className="mt-4 text-sm font-medium tracking-tight text-active">{step.title}</h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-muted">{step.description}</p>
                </Card>
              </FadeIn>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-border px-6 py-24">
        <FadeIn>
          <div className="mx-auto max-w-2xl rounded-2xl border border-border bg-canvas p-8 text-center sm:p-12">
            <h2 className="text-2xl font-medium tracking-tight text-active">Готов к мониторингу?</h2>
            <p className="mx-auto mt-3 max-w-md text-sm text-muted">
              Попробуй демо-режим без авторизации или войди, чтобы управлять своими узлами.
            </p>
            <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <NeonButton onClick={() => navigate('/servers/demo-server')}>Открыть демо</NeonButton>
              <Link
                to="/login"
                className="text-sm text-muted transition-colors hover:text-active"
              >
                Войти в аккаунт →
              </Link>
            </div>
          </div>
        </FadeIn>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-6 py-8">
        <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 sm:flex-row">
          <div className="flex items-center gap-2 text-active">
            <Activity className="h-4 w-4 text-accent" />
            <span className="text-sm font-medium tracking-tight">Homelytics</span>
          </div>
          <div className="flex items-center gap-6 text-xs text-muted">
            <Link to="/login" className="transition-colors hover:text-active">
              Login
            </Link>
            <Link to="/register" className="transition-colors hover:text-active">
              Register
            </Link>
            <Link to="/servers/demo-server" className="transition-colors hover:text-active">
              Demo
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}
