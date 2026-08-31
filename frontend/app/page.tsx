'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  Activity,
  ArrowRight,
  Bike,
  CalendarDays,
  Check,
  ChevronRight,
  Clock3,
  Gauge,
  Home,
  LineChart,
  ListTree,
  LoaderCircle,
  Settings,
  Sparkles,
} from 'lucide-react';
import { AdaptationCard } from '@/components/adaptation-card';
import { RpeHelp } from '@/components/rpe-help';
import { WorkoutSessionActions } from '@/components/workout-session-actions';
import { apiRequest } from '@/lib/api';
import {
  parseTrainingDate,
  type TrainingPlan,
  type Workout,
} from '@/lib/planning';

type User = { display_name: string; email: string };

const dayFormatter = new Intl.DateTimeFormat('pt-BR', { weekday: 'short' });
const fullDateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
});
const headerFormatter = new Intl.DateTimeFormat('pt-BR', {
  weekday: 'long',
  day: '2-digit',
  month: 'long',
});

function dateKey(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('');
}

function experienceLabel(value?: string) {
  return (
    {
      beginner: 'Iniciante',
      intermediate: 'Intermediário',
      advanced: 'Avançado',
    }[value || ''] || 'Ciclista'
  );
}

function Sidebar({ user, plan }: { user: User; plan: TrainingPlan | null }) {
  return (
    <aside className="sidebar">
      <div className="brand" aria-label="Cadência">
        <span className="brand-mark">
          <Bike size={21} strokeWidth={2.4} />
        </span>
        <span>cadência</span>
      </div>
      <nav className="main-nav" aria-label="Navegação principal">
        <a className="nav-item active" href="/">
          <Home size={18} />
          Visão geral
        </a>
        <a className="nav-item" href="/plano">
          <CalendarDays size={18} />
          Meu plano
        </a>
        <button className="nav-item" disabled>
          <Activity size={18} />
          Atividades
        </button>
        <button className="nav-item" disabled>
          <LineChart size={18} />
          Evolução
        </button>
      </nav>
      <div className="sidebar-bottom">
        <div className="coach-note">
          <span className="coach-icon">
            <Sparkles size={15} />
          </span>
          <p>
            <strong>Plano explicável</strong>Cada sessão mostra as regras usadas
            na decisão.
          </p>
        </div>
        <button className="nav-item" disabled>
          <Settings size={18} />
          Configurações
        </button>
        <button
          className="profile-mini"
          onClick={() => {
            window.location.href = '/perfil';
          }}
        >
          <span className="avatar">{initials(user.display_name)}</span>
          <span>
            <strong>{user.display_name}</strong>
            <small>
              Ciclismo ·{' '}
              {experienceLabel(plan?.prescription_snapshot.experience_level)}
            </small>
          </span>
          <ChevronRight size={15} />
        </button>
      </div>
    </aside>
  );
}

export default function HomePage() {
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<TrainingPlan | null>(null);
  const [selected, setSelected] = useState<Workout | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ plan: TrainingPlan | null }>('/v1/plans/current'),
    ])
      .then(([account, current]) => {
        setUser(account.user);
        setPlan(current.plan);
      })
      .catch(() => {
        window.location.href = '/entrar';
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!selected) return;

    const root = document.documentElement;
    const body = document.body;
    const previousRootOverflow = root.style.overflow;
    const previousBodyOverflow = body.style.overflow;
    const previousBodyPaddingRight = body.style.paddingRight;
    const scrollbarWidth = window.innerWidth - root.clientWidth;

    root.style.overflow = 'hidden';
    body.style.overflow = 'hidden';
    if (scrollbarWidth > 0) {
      body.style.paddingRight = `${scrollbarWidth}px`;
    }

    return () => {
      root.style.overflow = previousRootOverflow;
      body.style.overflow = previousBodyOverflow;
      body.style.paddingRight = previousBodyPaddingRight;
    };
  }, [selected]);

  const activePlan = plan?.status === 'active' ? plan : null;
  const today = useMemo(() => new Date(), []);
  const todayKey = dateKey(today);
  const focusWorkout = useMemo(() => {
    if (!activePlan?.workouts.length) return null;
    return (
      activePlan.workouts.find(
        (workout) =>
          workout.status === 'planned' && workout.scheduled_on >= todayKey,
      ) ||
      activePlan.workouts.find((workout) => workout.status === 'planned') ||
      activePlan.workouts[activePlan.workouts.length - 1]
    );
  }, [activePlan, todayKey]);

  function updateSessionPlan(nextPlan: TrainingPlan, workoutID: string) {
    setPlan(nextPlan);
    setSelected(
      nextPlan.workouts.find((workout) => workout.id === workoutID) || null,
    );
  }

  const displayedWeek = useMemo(() => {
    if (!focusWorkout || !activePlan) return [];
    const focusDate = parseTrainingDate(focusWorkout.scheduled_on);
    const monday = new Date(focusDate);
    monday.setDate(focusDate.getDate() - ((focusDate.getDay() + 6) % 7));
    return Array.from({ length: 7 }, (_, index) => {
      const date = new Date(monday);
      date.setDate(monday.getDate() + index);
      const key = dateKey(date);
      return {
        key,
        date,
        workout:
          activePlan.workouts.find((item) => item.scheduled_on === key) || null,
      };
    });
  }, [activePlan, focusWorkout]);

  if (loading || !user) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando seu painel…
      </main>
    );
  }

  if (!activePlan || !focusWorkout) {
    const completedCycle = plan?.status === 'completed';
    return (
      <main className="app-shell">
        <Sidebar user={user} plan={plan} />
        <section className="workspace">
          <header className="topbar">
            <div>
              <p className="eyebrow">SEU TREINO, NO SEU RITMO</p>
              <h1>Olá, {user.display_name.split(' ')[0]}.</h1>
            </div>
          </header>
          <section className="dashboard-empty">
            <span>
              <CalendarDays size={25} />
            </span>
            <p>
              {completedCycle
                ? 'CICLO CONCLUÍDO'
                : plan?.status === 'draft'
                  ? 'PLANO PRONTO PARA REVISÃO'
                  : 'PLANEJAMENTO'}
            </p>
            <h2>
              {completedCycle
                ? 'Pronto para suas próximas quatro semanas.'
                : plan?.status === 'draft'
                  ? 'Seu rascunho está esperando aprovação.'
                  : 'Crie seu primeiro ciclo de treinos.'}
            </h2>
            <div>
              {completedCycle
                ? 'Seu histórico foi preservado. Gere o próximo ciclo com seu perfil e sua disponibilidade atuais.'
                : plan?.status === 'draft'
                  ? 'Revise as quatro semanas e aceite o plano para mostrar as sessões reais neste painel.'
                  : 'Conclua seu perfil para gerar um plano compatível com sua experiência e disponibilidade.'}
            </div>
            <a href="/plano">
              {completedCycle
                ? 'Gerar próximo ciclo'
                : plan?.status === 'draft'
                  ? 'Revisar e aceitar plano'
                  : 'Criar meu plano'}
              <ArrowRight size={15} />
            </a>
          </section>
        </section>
      </main>
    );
  }

  const isToday = focusWorkout.scheduled_on === todayKey;
  const completedInWeek = displayedWeek.filter(
    (item) => item.workout?.status === 'completed',
  ).length;
  const sessionsInWeek = displayedWeek.filter((item) => item.workout).length;
  const weekMinutes = displayedWeek.reduce(
    (total, item) => total + (item.workout?.duration_minutes || 0),
    0,
  );
  const rules = focusWorkout.explanation.rules || [];
  const effortHeight = Math.max(34, Math.min(82, focusWorkout.target_rpe * 10));
  const effortBars = [
    25,
    30,
    36,
    effortHeight,
    effortHeight,
    effortHeight,
    34,
    effortHeight,
    effortHeight,
    effortHeight,
    30,
    24,
  ];

  return (
    <main className="app-shell">
      <Sidebar user={user} plan={activePlan} />
      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">
              {headerFormatter.format(today).toUpperCase()}
            </p>
            <h1>Olá, {user.display_name.split(' ')[0]}.</h1>
          </div>
          <a className="active-plan-pill" href="/plano">
            <span className="status-dot" />
            Plano ativo<strong>4 semanas</strong>
          </a>
        </header>
        <div className="dashboard-grid">
          <section className="today-card">
            <div className="today-copy">
              <div className="section-label">
                <span>{isToday ? 'HOJE' : 'PRÓXIMA SESSÃO'}</span>
                <i>•</i>
                {fullDateFormatter
                  .format(parseTrainingDate(focusWorkout.scheduled_on))
                  .toUpperCase()}
              </div>
              <AdaptationCard workout={focusWorkout} compact />
              <h2>{focusWorkout.name}</h2>
              <p>{focusWorkout.objective}</p>
              <div className="metric-row">
                <span>
                  <Clock3 size={17} />
                  <b>{focusWorkout.duration_minutes}</b> min
                </span>
                <span>
                  <Gauge size={17} />
                  <b>RPE {focusWorkout.target_rpe}</b> / 10 <RpeHelp compact />
                </span>
              </div>
              <button
                className="start-button"
                onClick={() => setSelected(focusWorkout)}
              >
                <ListTree size={16} />
                Ver estrutura
              </button>
              <a className="details-button" href="/plano">
                Ver plano completo <ArrowRight size={15} />
              </a>
            </div>
            <div
              className="effort-visual"
              aria-label="Representação da percepção de esforço da sessão"
            >
              <div className="effort-label">
                <span>PERFIL DE ESFORÇO</span>
                <strong>alvo RPE {focusWorkout.target_rpe}</strong>
              </div>
              <div className="bars">
                {effortBars.map((height, index) => (
                  <span
                    key={index}
                    style={{ height }}
                    className={index >= 3 && index <= 9 ? 'work' : ''}
                  />
                ))}
              </div>
              <div className="zone-scale">
                <span>LEVE</span>
                <span>MODERADO</span>
                <span>FORTE</span>
              </div>
              <div className="coach-tip">
                <Sparkles size={16} />
                <p>
                  <strong>Por que este treino?</strong>
                  {focusWorkout.explanation.summary}
                </p>
              </div>
            </div>
          </section>

          <aside className="readiness-card plan-summary-card">
            <div className="card-heading">
              <div>
                <span>PLANO ATIVO</span>
                <h3>Seu ciclo atual.</h3>
              </div>
              <Check size={18} />
            </div>
            <div className="plan-summary-number">
              <strong>{activePlan.workouts.length}</strong>
              <span>sessões em quatro semanas</span>
            </div>
            <div className="readiness-list">
              <div>
                <span>Início</span>
                <b>
                  {fullDateFormatter.format(
                    parseTrainingDate(activePlan.starts_on),
                  )}
                </b>
              </div>
              <div>
                <span>Término</span>
                <b>
                  {fullDateFormatter.format(
                    parseTrainingDate(activePlan.ends_on),
                  )}
                </b>
              </div>
              <div>
                <span>Motor</span>
                <b>
                  {activePlan.prescription_snapshot.engine_version ||
                    'rules-v1'}
                </b>
              </div>
            </div>
            <a className="checkin-button" href="/plano">
              Revisar plano <ArrowRight size={15} />
            </a>
          </aside>

          <section className="week-card">
            <div className="week-header">
              <div>
                <span className="section-kicker">SEMANA EM FOCO</span>
                <h3>Consistência vence intensidade isolada.</h3>
              </div>
              <div className="week-progress">
                <strong>
                  {completedInWeek} de {sessionsInWeek}
                </strong>
                <span>sessões concluídas</span>
              </div>
            </div>
            <div className="week-days">
              {displayedWeek.map((item) => (
                <article
                  key={item.key}
                  className={`day-card ${item.key === focusWorkout.scheduled_on ? 'selected' : ''}`}
                >
                  <div>
                    <span>
                      {dayFormatter
                        .format(item.date)
                        .replace('.', '')
                        .toUpperCase()}
                    </span>
                    <strong>
                      {String(item.date.getDate()).padStart(2, '0')}
                    </strong>
                  </div>
                  <i
                    className={`workout-dot ${item.workout ? (item.workout.status === 'completed' ? 'mint' : 'lime') : ''}`}
                  >
                    {item.workout?.status === 'completed' ? (
                      <Check size={13} />
                    ) : item.workout ? (
                      <Bike size={14} />
                    ) : (
                      ''
                    )}
                  </i>
                  <p>{item.workout?.name || 'Descanso'}</p>
                  {item.workout && (
                    <small>
                      {item.workout.duration_minutes} min · RPE{' '}
                      {item.workout.target_rpe}
                    </small>
                  )}
                </article>
              ))}
            </div>
            <div className="load-summary">
              <span>VOLUME PLANEJADO</span>
              <div className="load-track">
                <i
                  style={{
                    width: `${Math.min(100, (weekMinutes / 360) * 100)}%`,
                  }}
                />
              </div>
              <strong>
                {Math.floor(weekMinutes / 60)}h {weekMinutes % 60}min
              </strong>
              <em>{sessionsInWeek} sessões</em>
            </div>
          </section>

          <section className="insight-card">
            <div className="insight-icon">
              <Sparkles size={18} />
            </div>
            <div>
              <span>DECISÃO EXPLICÁVEL</span>
              <h3>O plano mostra por que cada sessão foi escolhida.</h3>
              <p>
                {focusWorkout.explanation.summary} As regras abaixo foram
                calculadas com os dados preenchidos no seu perfil.
              </p>
              <button onClick={() => setSelected(focusWorkout)}>
                Entender a decisão <ArrowRight size={15} />
              </button>
            </div>
            <div className="evidence-tag">
              <span>BASEADO EM</span>
              <strong>{rules.length} regras do seu perfil</strong>
              <small>
                Motor{' '}
                {activePlan.prescription_snapshot.engine_version || 'rules-v1'}
              </small>
            </div>
          </section>
        </div>
      </section>

      {selected && (
        <dialog open className="modal-backdrop" aria-labelledby="workout-title">
          <section className="workout-modal">
            <button
              className="modal-close"
              onClick={() => setSelected(null)}
              aria-label="Fechar"
            >
              ×
            </button>
            <span className="modal-kicker">
              SESSÃO PLANEJADA · {selected.duration_minutes} MIN
            </span>
            <h2 id="workout-title">{selected.name}</h2>
            <p>{selected.explanation.summary}</p>
            <AdaptationCard workout={selected} />
            <ol className="workout-steps">
              <li>
                <b>01</b>
                <span>
                  <strong>Aquecimento</strong>
                  <small>
                    {selected.structure.warmup_minutes} min · esforço
                    progressivo
                  </small>
                </span>
                <em>Cadência confortável</em>
              </li>
              <li>
                <b>02</b>
                <span>
                  <strong>Bloco principal</strong>
                  <small>
                    RPE {selected.target_rpe} <RpeHelp compact />
                  </small>
                </span>
                <em>{selected.structure.main}</em>
              </li>
              <li>
                <b>03</b>
                <span>
                  <strong>Desaquecimento</strong>
                  <small>
                    {selected.structure.cooldown_minutes} min · esforço leve
                  </small>
                </span>
                <em>Giro confortável</em>
              </li>
            </ol>
            <WorkoutSessionActions
              workout={selected}
              planStatus={activePlan.status}
              onPlanUpdated={updateSessionPlan}
            />
            <a className="modal-plan-link" href="/plano">
              Abrir plano completo <ArrowRight size={15} />
            </a>
          </section>
        </dialog>
      )}
    </main>
  );
}
