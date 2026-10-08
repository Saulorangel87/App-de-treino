'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { Bike, CalendarCheck2, CircleAlert, Clock3, HeartPulse, LineChart, LoaderCircle, MapPinned, MoonStar, Mountain, Target, Zap } from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { WeeklySummaryCard } from '@/components/weekly-summary-card';
import { useLocale, useMessages } from '@/components/locale-provider';
import { defineMessages, formatDecimal, INTL_LOCALE, type Locale } from '@/lib/i18n';
import { zoneLabel } from '@/lib/zones';

type User = { display_name: string };
type Week = { week_start: string; completed_sessions: number; cancelled_sessions: number; total_minutes: number; average_rpe: number; total_distance_km: number; total_elevation_m: number; average_power_watts: number; average_heart_rate: number; session_rpe_load: number; average_speed_kph: number };
type SessionComparison = { completed_on: string; name: string; planned_minutes: number; actual_minutes?: number; duration_delta_minutes?: number; planned_rpe: number; actual_rpe?: number; rpe_delta?: number; distance_km?: number; average_power_watts?: number; average_heart_rate?: number; fatigue_after?: number; pain_reported: boolean };
type RecoveryPoint = { recorded_on: string; sleep_minutes: number; sleep_quality: number; stress_level: number; fatigue_level: number; readiness: 'ready' | 'caution' | 'recovery_needed' };
type GoalProgress = { goal_type: string; priority: number; target_date?: string; details: string; window_days: number; completed_sessions: number; total_minutes: number; total_distance_km: number; session_rpe_load: number; average_speed_kph: number };
type Summary = { completed_sessions: number; cancelled_sessions: number; total_minutes: number; average_rpe: number; average_fatigue: number; completion_rate: number; total_distance_km: number; total_elevation_m: number; average_power_watts: number; average_heart_rate: number; session_rpe_load: number; average_speed_kph: number; goal_progress?: GoalProgress[]; weeks: Week[]; recent_sessions?: SessionComparison[]; recovery: RecoveryPoint[] };

const messages = defineMessages({
  pt: {
    goals: { health: 'Saúde', fitness: 'Condicionamento', endurance: 'Resistência', performance: 'Desempenho', event: 'Prova', weight_management: 'Gestão de peso' } as Record<string, string>,
    readiness: { ready: 'Adequada', caution: 'Atenção', recovery_needed: 'Recuperação' } as Record<RecoveryPoint['readiness'], string>,
    noDuration: 'Sem duração',
    onTime: 'No tempo previsto',
    noZone: 'Sem zona informada',
    moreEffort: 'Mais esforço',
    lessEffort: 'Menos esforço',
    asExpected: 'Dentro do esperado',
    loadFailed: 'Não foi possível carregar sua evolução.',
    loading: 'Carregando sua evolução…',
    kicker: 'EVOLUÇÃO',
    title: 'Seu histórico, com contexto.',
    intro: 'Registros observados ao longo do tempo. Eles ajudam você a acompanhar consistência e resposta percebida, mas não substituem avaliação profissional nem representam diagnóstico.',
    emptyTitle: 'Os primeiros dados aparecerão após seus treinos.',
    emptyText: 'Ao concluir ou cancelar sessões, o Cadência passará a organizar sua consistência, duração e esforço percebido aqui.',
    seePlan: 'Ver meu plano',
    summaryLabel: 'Resumo do histórico',
    completedSessions: 'sessões concluídas',
    recordedTime: 'tempo registrado',
    completionRate: 'conclusão das sessões',
    averageZone: 'zona média registrada',
    last8Weeks: 'ÚLTIMAS 8 SEMANAS',
    completedTime: 'Tempo concluído',
    sessionsInHistory: (count: number) => `${count} sessões no histórico`,
    minutesPerWeek: 'Minutos concluídos por semana',
    weekOf: (range: string) => `Semana de ${range}`,
    volumeCaption: 'Cada coluna representa o intervalo de sete dias indicado abaixo. Sem treino concluído, a semana aparece sem volume. Cancelamentos não entram no tempo registrado.',
    distancePerWeek: 'DISTÂNCIA POR SEMANA',
    inHistory: (value: string) => `${value} no histórico`,
    distanceChart: 'Distância registrada por semana',
    consistency: 'CONSISTÊNCIA',
    sessionLog: 'Registro das sessões',
    completed: 'concluídas',
    cancelled: (count: number) => `${count} canceladas`,
    completedPercent: (percent: number) => `${percent}% concluídas`,
    consistencyNote: 'Esta taxa considera somente sessões que já foram encerradas.',
    goalsKicker: 'ACOMPANHAMENTO DE OBJETIVOS',
    goalsTitle: 'O que foi registrado para sua direção atual',
    last28Days: 'Últimos 28 dias',
    goalsIntro: 'O Cadência mostra os registros ligados ao seu contexto atual sem inventar uma porcentagem de conclusão para metas que não têm uma medida final definida.',
    primaryGoal: 'OBJETIVO PRINCIPAL',
    secondaryGoal: 'OBJETIVO SECUNDÁRIO',
    targetDate: 'Data-alvo:',
    sessions: 'sessões',
    ridden: 'pedalados',
    recorded: 'registrados',
    trainingLoad: 'carga de treino',
    average: 'média',
    responseKicker: 'RESPOSTA AO TREINO',
    responseTitle: 'Planejado e realizado',
    lastSessions: (count: number) => `Últimas ${count} sessões concluídas`,
    responseIntro: 'Compare o que estava previsto com o que você registrou. Qualquer adaptação segue as regras de segurança do motor.',
    planned: 'Planejado',
    done: 'Realizado',
    time: 'Tempo',
    hr: 'FC',
    fatigue: 'Fadiga',
    pain: 'Dor relatada',
    rideKicker: 'DADOS DO PEDAL',
    rideTitle: 'Métricas registradas',
    totalDistance: 'distância acumulada',
    totalElevation: 'elevação acumulada',
    averagePower: 'potência média registrada',
    averageHr: 'frequência cardíaca média',
    averageSpeed: 'velocidade média registrada',
    rideCaption: 'A carga de treino é a duração em minutos × o esforço da sessão (a zona informada, convertida para a escala de 1 a 10). Velocidade e métricas físicas aparecem apenas quando foram registradas; elas descrevem registros, não estimam desempenho.',
    recoveryKicker: 'RECUPERAÇÃO',
    recoveryTitle: 'Check-ins recentes',
    newCheckin: 'Novo check-in',
    recoveryEmpty: 'Ainda não há check-ins registrados. Eles aparecerão aqui conforme você preencher a recuperação diária.',
    sleep: 'de sono',
    stress: 'Estresse',
  },
  en: {
    goals: { health: 'Health', fitness: 'Fitness', endurance: 'Endurance', performance: 'Performance', event: 'Race', weight_management: 'Weight management' },
    readiness: { ready: 'Good', caution: 'Caution', recovery_needed: 'Recovery' },
    noDuration: 'No duration',
    onTime: 'On planned time',
    noZone: 'No zone entered',
    moreEffort: 'More effort',
    lessEffort: 'Less effort',
    asExpected: 'As expected',
    loadFailed: 'Your progress could not be loaded.',
    loading: 'Loading your progress…',
    kicker: 'PROGRESS',
    title: 'Your history, in context.',
    intro: 'Records observed over time. They help you follow your consistency and perceived response, but they do not replace a professional assessment or represent a diagnosis.',
    emptyTitle: 'Your first data will show up after your workouts.',
    emptyText: 'Once you complete or cancel sessions, Cadência will organize your consistency, duration and perceived effort here.',
    seePlan: 'See my plan',
    summaryLabel: 'History summary',
    completedSessions: 'completed sessions',
    recordedTime: 'time recorded',
    completionRate: 'session completion',
    averageZone: 'average zone recorded',
    last8Weeks: 'LAST 8 WEEKS',
    completedTime: 'Time completed',
    sessionsInHistory: (count: number) => `${count} sessions in your history`,
    minutesPerWeek: 'Minutes completed per week',
    weekOf: (range: string) => `Week of ${range}`,
    volumeCaption: 'Each column represents the seven-day range shown below. With no completed workout, the week shows no volume. Cancellations do not count toward recorded time.',
    distancePerWeek: 'DISTANCE PER WEEK',
    inHistory: (value: string) => `${value} in your history`,
    distanceChart: 'Distance recorded per week',
    consistency: 'CONSISTENCY',
    sessionLog: 'Session log',
    completed: 'completed',
    cancelled: (count: number) => `${count} cancelled`,
    completedPercent: (percent: number) => `${percent}% completed`,
    consistencyNote: 'This rate only counts sessions that are already closed.',
    goalsKicker: 'GOAL TRACKING',
    goalsTitle: 'What was recorded toward your current direction',
    last28Days: 'Last 28 days',
    goalsIntro: 'Cadência shows the records tied to your current context without making up a completion percentage for goals that have no defined end measure.',
    primaryGoal: 'MAIN GOAL',
    secondaryGoal: 'SECONDARY GOAL',
    targetDate: 'Target date:',
    sessions: 'sessions',
    ridden: 'ridden',
    recorded: 'recorded',
    trainingLoad: 'training load',
    average: 'average',
    responseKicker: 'TRAINING RESPONSE',
    responseTitle: 'Planned and done',
    lastSessions: (count: number) => `Last ${count} completed sessions`,
    responseIntro: "Compare what was planned with what you logged. Any adaptation follows the engine's safety rules.",
    planned: 'Planned',
    done: 'Done',
    time: 'Time',
    hr: 'HR',
    fatigue: 'Fatigue',
    pain: 'Pain reported',
    rideKicker: 'RIDE DATA',
    rideTitle: 'Recorded metrics',
    totalDistance: 'total distance',
    totalElevation: 'total elevation',
    averagePower: 'average power recorded',
    averageHr: 'average heart rate',
    averageSpeed: 'average speed recorded',
    rideCaption: 'Training load is the duration in minutes × the session effort (the zone you entered, converted to the 1 to 10 scale). Speed and physical metrics only appear when they were recorded; they describe records and do not estimate performance.',
    recoveryKicker: 'RECOVERY',
    recoveryTitle: 'Recent check-ins',
    newCheckin: 'New check-in',
    recoveryEmpty: 'No check-ins recorded yet. They will show up here as you fill in your daily recovery.',
    sleep: 'of sleep',
    stress: 'Stress',
  },
});

function parseDate(value: string) { return new Date(`${value}T12:00:00`); }
function minutesLabel(value: number) { return value >= 60 ? `${Math.floor(value / 60)}h${value % 60 ? ` ${value % 60}min` : ''}` : `${value} min`; }
function distanceLabel(value: number, locale: Locale) { return `${value >= 100 ? Math.round(value) : formatDecimal(value, 1, locale)} km`; }
function speedLabel(value: number, locale: Locale) { return `${formatDecimal(value, 1, locale)} km/h`; }
function responseTone(value?: number) { if (value === undefined) return 'neutral'; if (value >= 2) return 'high'; if (value <= -2) return 'low'; return 'expected'; }

export default function EvolutionPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const [user, setUser] = useState<User | null>(null);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [today] = useState(() => new Date());

  useEffect(() => {
    Promise.all([apiRequest<{ user: User }>('/v1/me'), apiRequest<{ summary: Summary }>('/v1/evolution/summary')])
      .then(([account, result]) => { setUser(account.user); setSummary(result.summary); })
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

  const maxMinutes = useMemo(() => Math.max(...(summary?.weeks.map((week) => week.total_minutes) || [1]), 1), [summary]);
  const maxDistance = useMemo(() => Math.max(...(summary?.weeks.map((week) => week.total_distance_km) || [1]), 1), [summary]);
  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />{t.loading}</main>;
  if (!user || !summary) return <ApiErrorState message={error || t.loadFailed} />;
  const hasActivities = summary.completed_sessions + summary.cancelled_sessions > 0;
  const hasCyclingMetrics = summary.total_distance_km > 0 || summary.total_elevation_m > 0 || summary.average_power_watts > 0 || summary.average_heart_rate > 0;
  const recentSessions = summary.recent_sessions || [];
  const goalProgress = summary.goal_progress || [];

  const shortDate = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: '2-digit', month: 'short' });
  const dayLabel = (value: string) => shortDate.format(parseDate(value)).replace('.', '');
  const compactDate = (value: Date) => shortDate.format(value).replace(/\./g, '').replace(' de ', ' ');
  const weekRangeLabel = (value: string) => { const start = parseDate(value); const end = new Date(start); end.setDate(end.getDate() + 6); return `${compactDate(start)} – ${compactDate(end)}`; };
  const goalLabel = (value: string) => t.goals[value] || value;
  const signedMinutes = (value?: number) => { if (value === undefined) return t.noDuration; if (value === 0) return t.onTime; return `${value > 0 ? '+' : ''}${value} min`; };
  const responseLabel = (value?: number) => { if (value === undefined) return t.noZone; if (value >= 2) return t.moreEffort; if (value <= -2) return t.lessEffort; return t.asExpected; };
  const distance = (value: number) => distanceLabel(value, locale);

  return <main className="evolution-shell">
    <AppHeader name={user.display_name} />
    <section className="evolution-content">
      <header className="evolution-heading"><p>{t.kicker}</p><h1>{t.title}</h1><span>{t.intro}</span></header>
      {!hasActivities ? <section className="evolution-empty"><LineChart size={28} /><h2>{t.emptyTitle}</h2><p>{t.emptyText}</p><Link href="/plano">{t.seePlan}</Link></section> : <>
        <WeeklySummaryCard summary={summary} today={today} />
        <section className="evolution-metrics" aria-label={t.summaryLabel}><Metric icon={<CalendarCheck2 size={19} />} value={String(summary.completed_sessions)} label={t.completedSessions} /><Metric icon={<Clock3 size={19} />} value={minutesLabel(summary.total_minutes)} label={t.recordedTime} /><Metric icon={<Target size={19} />} value={`${Math.round(summary.completion_rate)}%`} label={t.completionRate} /><Metric icon={<CircleAlert size={19} />} value={summary.average_rpe ? zoneLabel(summary.average_rpe, locale) : '—'} label={t.averageZone} /></section>
        <section className="evolution-grid">
          <section className="evolution-card volume-card"><div className="evolution-card-title"><div><span>{t.last8Weeks}</span><h2>{t.completedTime}</h2></div><small>{t.sessionsInHistory(summary.completed_sessions)}</small></div><div className="volume-chart" aria-label={t.minutesPerWeek}>{summary.weeks.map((week) => <div className="volume-column" key={week.week_start}><strong>{week.total_minutes ? minutesLabel(week.total_minutes) : '—'}</strong><i style={{ height: `${Math.max(4, week.total_minutes / maxMinutes * 100)}%` }} /><span className="week-range-label" title={t.weekOf(weekRangeLabel(week.week_start))}>{weekRangeLabel(week.week_start)}</span></div>)}</div><p className="chart-caption">{t.volumeCaption}</p>{summary.total_distance_km > 0 && <><div className="cycling-chart-title"><span>{t.distancePerWeek}</span><strong>{t.inHistory(distance(summary.total_distance_km))}</strong></div><div className="distance-chart" aria-label={t.distanceChart}>{summary.weeks.map((week) => <div className="volume-column" key={week.week_start}><strong>{week.total_distance_km ? distance(week.total_distance_km) : '—'}</strong><i style={{ height: `${Math.max(4, week.total_distance_km / maxDistance * 100)}%` }} /><span className="week-range-label" title={t.weekOf(weekRangeLabel(week.week_start))}>{weekRangeLabel(week.week_start)}</span></div>)}</div></>}</section>
          <section className="evolution-card consistency-card"><div className="evolution-card-title"><div><span>{t.consistency}</span><h2>{t.sessionLog}</h2></div></div><div className="consistency-count"><strong>{summary.completed_sessions}</strong><span>{t.completed}</span></div><div className="consistency-line"><i style={{ width: `${summary.completion_rate}%` }} /></div><div className="consistency-labels"><span>{t.cancelled(summary.cancelled_sessions)}</span><span>{t.completedPercent(Math.round(summary.completion_rate))}</span></div><p>{t.consistencyNote}</p></section>
        </section>
        {goalProgress.length > 0 && <section className="evolution-card goal-progress-card"><div className="evolution-card-title"><div><span>{t.goalsKicker}</span><h2>{t.goalsTitle}</h2></div><small>{t.last28Days}</small></div><p className="goal-progress-intro">{t.goalsIntro}</p><div className="goal-progress-list">{goalProgress.map((goal) => <article key={`${goal.priority}-${goal.goal_type}`}><div><span>{goal.priority === 1 ? t.primaryGoal : t.secondaryGoal}</span><strong>{goalLabel(goal.goal_type)}</strong>{goal.target_date && <small>{t.targetDate} {dayLabel(goal.target_date)}</small>}</div><div className="goal-progress-metrics"><span><b>{goal.completed_sessions}</b> {t.sessions}</span><span><b>{minutesLabel(goal.total_minutes)}</b> {t.ridden}</span>{goal.total_distance_km > 0 && <span><b>{distance(goal.total_distance_km)}</b> {t.recorded}</span>}{goal.session_rpe_load > 0 && <span><b>{Math.round(goal.session_rpe_load)}</b> {t.trainingLoad}</span>}{goal.average_speed_kph > 0 && <span><b>{speedLabel(goal.average_speed_kph, locale)}</b> {t.average}</span>}</div>{goal.details && <p>{goal.details}</p>}</article>)}</div></section>}
        {recentSessions.length > 0 && <section className="evolution-card response-card"><div className="evolution-card-title"><div><span>{t.responseKicker}</span><h2>{t.responseTitle}</h2></div><small>{t.lastSessions(recentSessions.length)}</small></div><p className="response-intro">{t.responseIntro}</p><div className="session-comparison-list">{recentSessions.map((session, index) => <article className="session-comparison-item" key={`${session.completed_on}-${session.name}-${index}`}><div className="session-comparison-heading"><div><time>{dayLabel(session.completed_on)}</time><strong>{session.name}</strong></div><span className={`response-tag ${responseTone(session.rpe_delta)}`}>{responseLabel(session.rpe_delta)}</span></div><div className="session-comparison-metrics"><span><small>{t.planned}</small><b>{session.planned_minutes} min · {zoneLabel(session.planned_rpe, locale)}</b></span><span><small>{t.done}</small><b>{session.actual_minutes !== undefined ? `${session.actual_minutes} min` : '—'}{session.actual_rpe !== undefined ? ` · ${zoneLabel(session.actual_rpe, locale)}` : ''}</b></span><span><small>{t.time}</small><b>{signedMinutes(session.duration_delta_minutes)}</b></span></div><div className="session-comparison-foot">{session.distance_km !== undefined && <span>{distance(session.distance_km)}</span>}{session.average_heart_rate !== undefined && <span>{t.hr} {session.average_heart_rate} bpm</span>}{session.average_power_watts !== undefined && <span>{session.average_power_watts} W</span>}{session.fatigue_after !== undefined && <span>{t.fatigue} {session.fatigue_after}/5</span>}{session.pain_reported && <span className="pain">{t.pain}</span>}</div></article>)}</div></section>}
        {hasCyclingMetrics && <section className="evolution-card cycling-overview"><div className="evolution-card-title"><div><span>{t.rideKicker}</span><h2>{t.rideTitle}</h2></div></div><div className="cycling-overview-grid">{summary.total_distance_km > 0 && <Metric icon={<MapPinned size={18} />} value={distance(summary.total_distance_km)} label={t.totalDistance} />}{summary.total_elevation_m > 0 && <Metric icon={<Mountain size={18} />} value={`${Math.round(summary.total_elevation_m)} m+`} label={t.totalElevation} />}{summary.average_power_watts > 0 && <Metric icon={<Zap size={18} />} value={`${Math.round(summary.average_power_watts)} W`} label={t.averagePower} />}{summary.average_heart_rate > 0 && <Metric icon={<HeartPulse size={18} />} value={`${Math.round(summary.average_heart_rate)} bpm`} label={t.averageHr} />}{summary.average_speed_kph > 0 && <Metric icon={<Bike size={18} />} value={speedLabel(summary.average_speed_kph, locale)} label={t.averageSpeed} />}{summary.session_rpe_load > 0 && <Metric icon={<Target size={18} />} value={String(Math.round(summary.session_rpe_load))} label={t.trainingLoad} />}</div><p className="chart-caption">{t.rideCaption}</p></section>}
      </>}
      <section className="evolution-card recovery-history"><div className="evolution-card-title"><div><span>{t.recoveryKicker}</span><h2>{t.recoveryTitle}</h2></div><Link href="/recuperacao"><MoonStar size={15} />{t.newCheckin}</Link></div>{summary.recovery.length === 0 ? <p className="recovery-empty">{t.recoveryEmpty}</p> : <div className="recovery-history-list">{summary.recovery.map((point) => <div className="recovery-history-item" key={point.recorded_on}><time>{dayLabel(point.recorded_on)}</time><span className={`readiness-tag ${point.readiness}`}>{t.readiness[point.readiness]}</span><span>{Math.floor(point.sleep_minutes / 60)}h{point.sleep_minutes % 60 ? '30' : ''} {t.sleep}</span><span>{t.stress} {point.stress_level}/5</span><span>{t.fatigue} {point.fatigue_level}/5</span></div>)}</div>}</section>
    </section>
  </main>;
}

function Metric({ icon, value, label }: { icon: React.ReactNode; value: string; label: string }) { return <article><span>{icon}</span><strong>{value}</strong><small>{label}</small></article>; }
