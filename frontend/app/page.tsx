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
import { ZoneHelp, ZoneSummary } from '@/components/zone-help';
import { zoneForRpe, zoneLabel, zoneName, zoneReferenceFrom } from '@/lib/zones';
import { useScrollLock } from '@/components/use-scroll-lock';
import { TrailLegend, TrailSymbol, intensityMessages, intensityOf } from '@/components/trail-symbol';
import { useLocale, useMessages } from '@/components/locale-provider';
import { WorkoutSessionActions } from '@/components/workout-session-actions';
import { WorkoutStructure } from '@/components/workout-structure';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { defineMessages, INTL_LOCALE } from '@/lib/i18n';
import {
  parseTrainingDate,
  type TrainingPlan,
  type Workout,
} from '@/lib/planning';

type User = { display_name: string; email: string };
type Recovery = { readiness: 'ready' | 'caution' | 'recovery_needed' };

const messages = defineMessages({
  pt: {
    weekPhases: ['Progressão', 'Progressão', 'Maior carga', 'Recuperação'],
    recovery: { ready: 'Recuperação ok', caution: 'Atenção à recuperação', recovery_needed: 'Priorize recuperação' } as Record<Recovery['readiness'], string>,
    loadFailed: 'Não foi possível carregar seu painel.',
    loading: 'Carregando seu painel…',
    checkin: 'Fazer check-in de hoje',
    yourPace: 'Seu treino, no seu ritmo',
    hello: (name: string) => `Olá, ${name}.`,
    cycleDone: 'Ciclo concluído',
    draftReady: 'Plano pronto para revisão',
    planning: 'Planejamento',
    cycleDoneTitle: 'Pronto para suas próximas quatro semanas.',
    draftTitle: 'Seu rascunho está esperando aprovação.',
    firstTitle: 'Crie seu primeiro ciclo de treinos.',
    cycleDoneText: 'Seu histórico foi preservado. Gere o próximo ciclo com seu perfil e sua disponibilidade atuais.',
    draftText: 'Revise as quatro semanas e aceite o plano para mostrar as sessões reais neste painel.',
    firstText: 'Conclua seu perfil para gerar um plano compatível com sua experiência e disponibilidade.',
    nextCycle: 'Gerar próximo ciclo',
    reviewDraft: 'Revisar e aceitar plano',
    createPlan: 'Criar meu plano',
    today: 'Hoje',
    nextSession: 'Próxima sessão',
    time: 'Tempo',
    zone: 'Zona',
    level: 'Nível',
    seeStructure: 'Ver estrutura',
    fullPlan: 'Plano completo',
    route: 'Percurso da sessão',
    routeLabel: (name: string, steps: number, minutes: number) => `Percurso da sessão ${name}: ${steps} etapas em ${minutes} minutos`,
    why: 'Por que este treino?',
    weekOf: (week: number, phase: string) => `Semana ${week} de 4 · ${phase}`,
    weekSheets: 'Folhas da semana',
    of: (done: number, total: number) => `${done} de ${total}`,
    sessionsDone: 'sessões concluídas',
    completed: 'Concluído',
    rest: 'Descanso',
    plannedVolume: 'Volume planejado',
    sessions: (count: number) => `${count} sessões`,
    activePlan: 'Plano ativo',
    currentCycle: 'Seu ciclo atual',
    weekLabel: (week: number) => `Semana ${week} de 4`,
    weekShort: (week: number) => `S${week}`,
    sessionsLabel: 'Sessões',
    in4Weeks: (count: number) => `${count} em 4 semanas`,
    start: 'Início',
    end: 'Término',
    reviewPlan: 'Revisar plano',
    explainable: 'Decisão explicável',
    explainableTitle: 'O plano mostra por que cada sessão foi escolhida.',
    rulesNote: 'As regras foram calculadas com os dados preenchidos no seu perfil.',
    understand: 'Entender a decisão',
    basedOn: 'Baseado em',
    rulesCount: (count: number) => `${count} regras do seu perfil`,
    engine: (engine: string) => `Motor ${engine}`,
    close: 'Fechar',
    plannedSession: 'Sessão planejada',
    routeOf: (name: string) => `Percurso da sessão ${name}`,
    openPlan: 'Abrir plano completo',
  },
  en: {
    weekPhases: ['Build', 'Build', 'Peak load', 'Recovery'],
    recovery: { ready: 'Recovery OK', caution: 'Watch your recovery', recovery_needed: 'Prioritize recovery' },
    loadFailed: 'Your dashboard could not be loaded.',
    loading: 'Loading your dashboard…',
    checkin: "Do today's check-in",
    yourPace: 'Your training, at your pace',
    hello: (name: string) => `Hi, ${name}.`,
    cycleDone: 'Cycle complete',
    draftReady: 'Plan ready for review',
    planning: 'Planning',
    cycleDoneTitle: 'Ready for your next four weeks.',
    draftTitle: 'Your draft is waiting for approval.',
    firstTitle: 'Create your first training cycle.',
    cycleDoneText: 'Your history was kept. Generate the next cycle with your current profile and availability.',
    draftText: 'Review the four weeks and accept the plan to show the real sessions on this dashboard.',
    firstText: 'Finish your profile to generate a plan that fits your experience and availability.',
    nextCycle: 'Generate next cycle',
    reviewDraft: 'Review and accept plan',
    createPlan: 'Create my plan',
    today: 'Today',
    nextSession: 'Next session',
    time: 'Time',
    zone: 'Zone',
    level: 'Level',
    seeStructure: 'See structure',
    fullPlan: 'Full plan',
    route: 'Session route',
    routeLabel: (name: string, steps: number, minutes: number) => `Route for the ${name} session: ${steps} steps in ${minutes} minutes`,
    why: 'Why this workout?',
    weekOf: (week: number, phase: string) => `Week ${week} of 4 · ${phase}`,
    weekSheets: "This week's sheets",
    of: (done: number, total: number) => `${done} of ${total}`,
    sessionsDone: 'sessions completed',
    completed: 'Completed',
    rest: 'Rest',
    plannedVolume: 'Planned volume',
    sessions: (count: number) => `${count} sessions`,
    activePlan: 'Active plan',
    currentCycle: 'Your current cycle',
    weekLabel: (week: number) => `Week ${week} of 4`,
    weekShort: (week: number) => `W${week}`,
    sessionsLabel: 'Sessions',
    in4Weeks: (count: number) => `${count} over 4 weeks`,
    start: 'Start',
    end: 'End',
    reviewPlan: 'Review plan',
    explainable: 'Explainable decision',
    explainableTitle: 'The plan shows why each session was chosen.',
    rulesNote: 'The rules were calculated from the data in your profile.',
    understand: 'Understand the decision',
    basedOn: 'Based on',
    rulesCount: (count: number) => `${count} rules from your profile`,
    engine: (engine: string) => `Engine ${engine}`,
    close: 'Close',
    plannedSession: 'Planned session',
    routeOf: (name: string) => `Route for the ${name} session`,
    openPlan: 'Open full plan',
  },
});

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

export default function HomePage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const intensityText = useMessages(intensityMessages);
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
        setError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

  useScrollLock(Boolean(selected));

  const activePlan = plan?.status === 'active' ? plan : null;
  const [today] = useState(() => new Date());
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
        {t.loading}
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  const intl = INTL_LOCALE[locale];
  const dayFormatter = new Intl.DateTimeFormat(intl, { weekday: 'short' });
  const fullDateFormatter = new Intl.DateTimeFormat(intl, { day: '2-digit', month: 'long', year: 'numeric' });
  const sessionDateFormatter = new Intl.DateTimeFormat(intl, { weekday: 'long', day: 'numeric', month: 'short' });
  const headerFormatter = new Intl.DateTimeFormat(intl, { weekday: 'long', day: '2-digit', month: 'long' });

  const firstName = user.display_name.split(' ')[0];
  const recoveryLink = (
    <Link className={`recovery-pill ${recovery?.readiness || 'pending'}`} href="/recuperacao">
      <HeartPulse size={16} aria-hidden="true" />
      {recovery ? t.recovery[recovery.readiness] : t.checkin}
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
              <p className="kicker">{t.yourPace}</p>
              <h1>{t.hello(firstName)}</h1>
            </div>
            {recoveryLink}
          </header>
          <section className="dashboard-empty sheet topo-surface">
            <span className="dashboard-empty-icon">
              <CalendarDays size={24} aria-hidden="true" />
            </span>
            <p className="kicker">
              {completedCycle ? t.cycleDone : plan?.status === 'draft' ? t.draftReady : t.planning}
            </p>
            <h2>
              {completedCycle ? t.cycleDoneTitle : plan?.status === 'draft' ? t.draftTitle : t.firstTitle}
            </h2>
            <p className="dashboard-empty-copy">
              {completedCycle ? t.cycleDoneText : plan?.status === 'draft' ? t.draftText : t.firstText}
            </p>
            <Link className="btn btn-primary" href="/plano">
              {completedCycle ? t.nextCycle : plan?.status === 'draft' ? t.reviewDraft : t.createPlan}
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
  const steps = stepsForWorkout(focusWorkout, locale);
  const intensity = intensityOf(focusWorkout.target_rpe);
  const engine = activePlan.prescription_snapshot.engine_version || 'rules-v1';

  return (
    <main className="dashboard-shell">
      <AppHeader name={user.display_name} />
      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="kicker">{headerFormatter.format(today)}</p>
            <h1>{t.hello(firstName)}</h1>
          </div>
          {recoveryLink}
        </header>

        <div className="dashboard-grid">
          <section className="today-card sheet" aria-labelledby="today-title">
            <p className="section-label">
              <span>{isToday ? t.today : t.nextSession}</span>
              {sessionDateFormatter.format(parseTrainingDate(focusWorkout.scheduled_on))}
            </p>
            <h2 id="today-title">{focusWorkout.name}</h2>
            <p className="today-objective">{focusWorkout.objective}</p>
            <dl className="today-facts">
              <div>
                <dt>{t.time}</dt>
                <dd>{focusWorkout.duration_minutes}&prime;</dd>
              </div>
              <div>
                <dt>
                  {t.zone} <ZoneHelp compact />
                </dt>
                <dd className="today-zone">
                  <b>Z{zoneForRpe(focusWorkout.target_rpe).number}</b>
                  <span>{zoneName(zoneForRpe(focusWorkout.target_rpe), locale)}</span>
                </dd>
              </div>
              <div>
                <dt>{t.level}</dt>
                <dd className="today-level">
                  <TrailSymbol intensity={intensity} label="" />
                  {intensityText.labels[intensity]}
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
                {t.seeStructure}
              </button>
              <Link className="details-button" href="/plano">
                {t.fullPlan} <ArrowRight size={16} aria-hidden="true" />
              </Link>
            </div>
          </section>

          <section className="today-map panel" aria-labelledby="today-map-title">
            <div className="today-map-head">
              <h3 id="today-map-title" className="kicker">
                {t.route}
              </h3>
              <TrailLegend />
            </div>
            <RouteMap
              steps={steps}
              label={t.routeLabel(focusWorkout.name, steps.length, focusWorkout.duration_minutes)}
            />
            <div className="today-map-foot">
              <RouteScale minutes={focusWorkout.duration_minutes} />
              <p className="coach-tip">
                <strong>{t.why}</strong>
                {focusWorkout.explanation.summary}
              </p>
            </div>
          </section>

          <section className="week-card" aria-labelledby="week-title">
            <div className="week-header">
              <div>
                <p className="kicker">
                  {t.weekOf(weekIndex + 1, t.weekPhases[weekIndex])}
                </p>
                <h3 id="week-title">{t.weekSheets}</h3>
              </div>
              <p className="week-progress">
                <strong>{t.of(completedInWeek, sessionsInWeek)}</strong>
                {t.sessionsDone}
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
                      <span className="day-done" role="img" aria-label={t.completed}>
                        <Check size={13} strokeWidth={3} />
                      </span>
                    ) : (
                      <TrailSymbol rpe={workout?.target_rpe} rest={!workout} />
                    )}
                    <span className="day-minutes">
                      {workout ? `${workout.duration_minutes}′` : '—'}
                    </span>
                    <span className="day-title">{workout?.name || t.rest}</span>
                  </li>
                );
              })}
            </ol>
            <div className="load-summary">
              <span className="kicker">{t.plannedVolume}</span>
              <div className="load-track" aria-hidden="true">
                <i style={{ width: `${Math.min(100, (weekMinutes / 360) * 100)}%` }} />
              </div>
              <strong>{formatMinutes(weekMinutes)}</strong>
              <em>{t.sessions(sessionsInWeek)}</em>
            </div>
          </section>

          <aside className="plan-summary-card panel" aria-labelledby="cycle-title">
            <p className="kicker">{t.activePlan}</p>
            <h3 id="cycle-title">{t.currentCycle}</h3>
            <ol className="cycle-track" aria-label={t.weekLabel(weekIndex + 1)}>
              {t.weekPhases.map((phase, index) => (
                <li
                  key={index}
                  className={index < weekIndex ? 'past' : index === weekIndex ? 'current' : ''}
                >
                  <span>{t.weekShort(index + 1)}</span>
                  <small>{phase}</small>
                </li>
              ))}
            </ol>
            <dl className="readiness-list">
              <div>
                <dt>{t.sessionsLabel}</dt>
                <dd>{t.in4Weeks(activePlan.workouts.length)}</dd>
              </div>
              <div>
                <dt>{t.start}</dt>
                <dd>{fullDateFormatter.format(parseTrainingDate(activePlan.starts_on))}</dd>
              </div>
              <div>
                <dt>{t.end}</dt>
                <dd>{fullDateFormatter.format(parseTrainingDate(activePlan.ends_on))}</dd>
              </div>
            </dl>
            <Link className="checkin-button" href="/plano">
              {t.reviewPlan} <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </aside>

          <section className="insight-card panel" aria-labelledby="insight-title">
            <div>
              <p className="kicker">{t.explainable}</p>
              <h3 id="insight-title">{t.explainableTitle}</h3>
              <p>
                {focusWorkout.explanation.summary} {t.rulesNote}
              </p>
              <button
                type="button"
                className="details-button"
                onClick={() => setSelected(focusWorkout)}
              >
                {t.understand} <ArrowRight size={16} aria-hidden="true" />
              </button>
            </div>
            <div className="evidence-tag">
              <span className="kicker">{t.basedOn}</span>
              <strong>{t.rulesCount(rules.length)}</strong>
              <small>{t.engine(engine)}</small>
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
              aria-label={t.close}
            >
              <X size={20} />
            </button>
            <p className="kicker">
              {t.plannedSession} · {selected.duration_minutes} min · {zoneLabel(selected.target_rpe, locale)}
            </p>
            <h2 id="workout-title">{selected.name}</h2>
            <p className="workout-modal-summary">{selected.explanation.summary}</p>
            <AdaptationCard workout={selected} />
            <ZoneSummary
              rpe={selected.target_rpe}
              reference={zoneReferenceFrom(activePlan.prescription_snapshot.cycling_context)}
            />
            <div className="workout-modal-map">
              <RouteMap
                steps={stepsForWorkout(selected, locale)}
                label={t.routeOf(selected.name)}
              />
            </div>
            <WorkoutStructure structure={selected.structure} durationMinutes={selected.duration_minutes} />
            <WorkoutSessionActions
              workout={selected}
              planStatus={activePlan.status}
              usesHeartRate={Boolean(activePlan.prescription_snapshot.cycling_context?.uses_heart_rate)}
              usesPower={Boolean(activePlan.prescription_snapshot.cycling_context?.uses_power)}
              zoneReference={zoneReferenceFrom(activePlan.prescription_snapshot.cycling_context)}
              onPlanUpdated={updateSessionPlan}
            />
            <Link className="modal-plan-link" href="/plano">
              {t.openPlan} <ArrowRight size={16} aria-hidden="true" />
            </Link>
          </section>
        </dialog>
      )}
    </main>
  );
}
