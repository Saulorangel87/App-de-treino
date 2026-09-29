'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import {
  ArrowRight,
  CalendarDays,
  Check,
  HeartPulse,
  ListTree,
  LoaderCircle,
  X,
} from 'lucide-react';
import { AdaptationCard } from '@/components/adaptation-card';
import { AppHeader } from '@/components/app-header';
import { RouteMap, RouteScale, stepsForWorkout } from '@/components/route-map';
import { RpeHelp } from '@/components/rpe-help';
import { TrailLegend, TrailSymbol, intensityLabels, intensityOf } from '@/components/trail-symbol';
import { WorkoutSessionActions } from '@/components/workout-session-actions';
import { WorkoutStructure } from '@/components/workout-structure';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import {
  parseTrainingDate,
  type TrainingPlan,
  type Workout,
} from '@/lib/planning';

type User = { display_name: string; email: string };
type Recovery = { readiness: 'ready' | 'caution' | 'recovery_needed' };

const dayFormatter = new Intl.DateTimeFormat('pt-BR', { weekday: 'short' });
const fullDateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
});
const sessionDateFormatter = new Intl.DateTimeFormat('pt-BR', {
  weekday: 'long',
  day: 'numeric',
  month: 'short',
});
const headerFormatter = new Intl.DateTimeFormat('pt-BR', {
  weekday: 'long',
  day: '2-digit',
  month: 'long',
});

const weekPhases = ['Progressão', 'Progressão', 'Maior carga', 'Recuperação'];

function dateKey(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function formatMinutes(total: number) {
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  if (!hours) return `${minutes} min`;
  return minutes ? `${hours}h${String(minutes).padStart(2, '0')}` : `${hours}h`;
}

const recoveryCopy = {
  ready: 'Recuperação ok',
  caution: 'Atenção à recuperação',
  recovery_needed: 'Priorize recuperação',
} as const;

export default function HomePage() {
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<TrainingPlan | null>(null);
  const [selected, setSelected] = useState<Workout | null>(null);
  const [recovery, setRecovery] = useState<Recovery | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const today = dateKey(new Date());
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ plan: TrainingPlan | null }>('/v1/plans/current'),
      apiRequest<{ recovery: Recovery | null }>(`/v1/recovery/today?date=${today}`),
    ])
      .then(([account, current, daily]) => {
        setUser(account.user);
        setPlan(current.plan);
        setRecovery(daily.recovery);
      })
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, 'Não foi possível carregar seu painel.'));
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
      activePlan.workouts.find((workout) => workout.status === 'in_progress') ||
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

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando seu painel…
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || 'Não foi possível carregar seu painel.'} />;

  const firstName = user.display_name.split(' ')[0];
  const recoveryLink = (
    <Link className={`recovery-pill ${recovery?.readiness || 'pending'}`} href="/recuperacao">
      <HeartPulse size={16} aria-hidden="true" />
      {recovery ? recoveryCopy[recovery.readiness] : 'Fazer check-in de hoje'}
    </Link>
  );

  if (!activePlan || !focusWorkout) {
    const completedCycle = plan?.status === 'completed';
    return (
      <main className="dashboard-shell">
        <AppHeader name={user.display_name} />
        <section className="workspace">
          <header className="topbar">
            <div>
              <p className="kicker">Seu treino, no seu ritmo</p>
              <h1>Olá, {firstName}.</h1>
            </div>
            {recoveryLink}
          </header>
          <section className="dashboard-empty sheet topo-surface">
            <span className="dashboard-empty-icon">
              <CalendarDays size={24} aria-hidden="true" />
            </span>
            <p className="kicker">
              {completedCycle
                ? 'Ciclo concluído'
                : plan?.status === 'draft'
                  ? 'Plano pronto para revisão'
                  : 'Planejamento'}
            </p>
            <h2>
              {completedCycle
                ? 'Pronto para suas próximas quatro semanas.'
                : plan?.status === 'draft'
                  ? 'Seu rascunho está esperando aprovação.'
                  : 'Crie seu primeiro ciclo de treinos.'}
            </h2>
            <p className="dashboard-empty-copy">
              {completedCycle
                ? 'Seu histórico foi preservado. Gere o próximo ciclo com seu perfil e sua disponibilidade atuais.'
                : plan?.status === 'draft'
                  ? 'Revise as quatro semanas e aceite o plano para mostrar as sessões reais neste painel.'
                  : 'Conclua seu perfil para gerar um plano compatível com sua experiência e disponibilidade.'}
            </p>
            <Link className="btn btn-primary" href="/plano">
              {completedCycle
                ? 'Gerar próximo ciclo'
                : plan?.status === 'draft'
                  ? 'Revisar e aceitar plano'
                  : 'Criar meu plano'}
              <ArrowRight size={16} aria-hidden="true" />
            </Link>
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
  const weekIndex = Math.min(
    3,
    Math.max(
      0,
      Math.floor(
        (parseTrainingDate(focusWorkout.scheduled_on).getTime() -
          parseTrainingDate(activePlan.starts_on).getTime()) /
          604800000,
      ),
    ),
  );
  const rules = focusWorkout.explanation.rules || [];
  const steps = stepsForWorkout(focusWorkout);
  const intensity = intensityOf(focusWorkout.target_rpe);
  const engine = activePlan.prescription_snapshot.engine_version || 'rules-v1';

  return (
    <main className="dashboard-shell">
      <AppHeader name={user.display_name} />
      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="kicker">{headerFormatter.format(today)}</p>
            <h1>Olá, {firstName}.</h1>
          </div>
          {recoveryLink}
        </header>

        <div className="dashboard-grid">
          <section className="today-card sheet" aria-labelledby="today-title">
            <p className="section-label">
              <span>{isToday ? 'Hoje' : 'Próxima sessão'}</span>
              {sessionDateFormatter.format(parseTrainingDate(focusWorkout.scheduled_on))}
            </p>
            <h2 id="today-title">{focusWorkout.name}</h2>
            <p className="today-objective">{focusWorkout.objective}</p>
            <dl className="today-facts">
              <div>
                <dt>Tempo</dt>
                <dd>{focusWorkout.duration_minutes}&prime;</dd>
              </div>
              <div>
                <dt>
                  RPE <RpeHelp compact />
                </dt>
                <dd>{focusWorkout.target_rpe}</dd>
              </div>
              <div>
                <dt>Nível</dt>
                <dd className="today-level">
                  <TrailSymbol intensity={intensity} label="" />
                  {intensityLabels[intensity]}
                </dd>
              </div>
            </dl>
            <AdaptationCard workout={focusWorkout} compact />
            <div className="today-actions">
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => setSelected(focusWorkout)}
              >
                <ListTree aria-hidden="true" />
                Ver estrutura
              </button>
              <Link className="details-button" href="/plano">
                Plano completo <ArrowRight size={16} aria-hidden="true" />
              </Link>
            </div>
          </section>

          <section className="today-map panel" aria-labelledby="today-map-title">
            <div className="today-map-head">
              <h3 id="today-map-title" className="kicker">
                Percurso da sessão
              </h3>
              <TrailLegend />
            </div>
            <RouteMap
              steps={steps}
              label={`Percurso da sessão ${focusWorkout.name}: ${steps.length} etapas em ${focusWorkout.duration_minutes} minutos`}
            />
            <div className="today-map-foot">
              <RouteScale minutes={focusWorkout.duration_minutes} />
              <p className="coach-tip">
                <strong>Por que este treino?</strong>
                {focusWorkout.explanation.summary}
              </p>
            </div>
          </section>

          <section className="week-card" aria-labelledby="week-title">
            <div className="week-header">
              <div>
                <p className="kicker">
                  Semana {weekIndex + 1} de 4 · {weekPhases[weekIndex]}
                </p>
                <h3 id="week-title">Folhas da semana</h3>
              </div>
              <p className="week-progress">
                <strong>
                  {completedInWeek} de {sessionsInWeek}
                </strong>
                sessões concluídas
              </p>
            </div>
            <ol className="week-days">
              {displayedWeek.map((item) => {
                const workout = item.workout;
                const done = workout?.status === 'completed';
                return (
                  <li
                    key={item.key}
                    className={[
                      'day-card',
                      item.key === focusWorkout.scheduled_on ? 'selected' : '',
                      done ? 'done' : '',
                      workout ? '' : 'rest',
                    ]
                      .filter(Boolean)
                      .join(' ')}
                  >
                    <span className="day-name">
                      {dayFormatter.format(item.date).replace('.', '')}
                    </span>
                    <strong className="day-number">
                      {String(item.date.getDate()).padStart(2, '0')}
                    </strong>
                    {done ? (
                      <span className="day-done" role="img" aria-label="Concluído">
                        <Check size={13} strokeWidth={3} />
                      </span>
                    ) : (
                      <TrailSymbol rpe={workout?.target_rpe} rest={!workout} />
                    )}
                    <span className="day-minutes">
                      {workout ? `${workout.duration_minutes}′` : '—'}
                    </span>
                    <span className="day-title">{workout?.name || 'Descanso'}</span>
                  </li>
                );
              })}
            </ol>
            <div className="load-summary">
              <span className="kicker">Volume planejado</span>
              <div className="load-track" aria-hidden="true">
                <i style={{ width: `${Math.min(100, (weekMinutes / 360) * 100)}%` }} />
              </div>
              <strong>{formatMinutes(weekMinutes)}</strong>
              <em>{sessionsInWeek} sessões</em>
            </div>
          </section>

          <aside className="plan-summary-card panel" aria-labelledby="cycle-title">
            <p className="kicker">Plano ativo</p>
            <h3 id="cycle-title">Seu ciclo atual</h3>
            <ol className="cycle-track" aria-label={`Semana ${weekIndex + 1} de 4`}>
              {weekPhases.map((phase, index) => (
                <li
                  key={index}
                  className={index < weekIndex ? 'past' : index === weekIndex ? 'current' : ''}
                >
                  <span>S{index + 1}</span>
                  <small>{phase}</small>
                </li>
              ))}
            </ol>
            <dl className="readiness-list">
              <div>
                <dt>Sessões</dt>
                <dd>{activePlan.workouts.length} em 4 semanas</dd>
              </div>
              <div>
                <dt>Início</dt>
                <dd>{fullDateFormatter.format(parseTrainingDate(activePlan.starts_on))}</dd>
              </div>
              <div>
                <dt>Término</dt>
                <dd>{fullDateFormatter.format(parseTrainingDate(activePlan.ends_on))}</dd>
              </div>
            </dl>
            <Link className="checkin-button" href="/plano">
              Revisar plano <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </aside>

          <section className="insight-card panel" aria-labelledby="insight-title">
            <div>
              <p className="kicker">Decisão explicável</p>
              <h3 id="insight-title">O plano mostra por que cada sessão foi escolhida.</h3>
              <p>
                {focusWorkout.explanation.summary} As regras foram calculadas
                com os dados preenchidos no seu perfil.
              </p>
              <button
                type="button"
                className="details-button"
                onClick={() => setSelected(focusWorkout)}
              >
                Entender a decisão <ArrowRight size={16} aria-hidden="true" />
              </button>
            </div>
            <div className="evidence-tag">
              <span className="kicker">Baseado em</span>
              <strong>{rules.length} regras do seu perfil</strong>
              <small>Motor {engine}</small>
            </div>
          </section>
        </div>
      </section>

      {selected && (
        <dialog open className="modal-backdrop" aria-labelledby="workout-title">
          <section className="workout-modal">
            <button
              type="button"
              className="modal-close"
              onClick={() => setSelected(null)}
              aria-label="Fechar"
            >
              <X size={20} />
            </button>
            <p className="kicker">
              Sessão planejada · {selected.duration_minutes} min · RPE {selected.target_rpe}
            </p>
            <h2 id="workout-title">{selected.name}</h2>
            <p className="workout-modal-summary">{selected.explanation.summary}</p>
            <AdaptationCard workout={selected} />
            <div className="workout-modal-map">
              <RouteMap
                steps={stepsForWorkout(selected)}
                label={`Percurso da sessão ${selected.name}`}
              />
            </div>
            <WorkoutStructure structure={selected.structure} durationMinutes={selected.duration_minutes} />
            <WorkoutSessionActions
              workout={selected}
              planStatus={activePlan.status}
              usesHeartRate={Boolean(activePlan.prescription_snapshot.cycling_context?.uses_heart_rate)}
              usesPower={Boolean(activePlan.prescription_snapshot.cycling_context?.uses_power)}
              onPlanUpdated={updateSessionPlan}
            />
            <Link className="modal-plan-link" href="/plano">
              Abrir plano completo <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </section>
        </dialog>
      )}
    </main>
  );
}
