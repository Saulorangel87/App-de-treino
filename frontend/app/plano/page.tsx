'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import {
  Check,
  CheckCircle2,
  Clock3,
  Gauge,
  LoaderCircle,
  RefreshCw,
  Shield,
  ShieldAlert,
  Sparkles,
  X,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { AppHeader } from '@/components/app-header';
import { AdaptationCard } from '@/components/adaptation-card';
import { ProtectionNotice } from '@/components/protection-notice';
import { ZoneHelp, ZoneSummary } from '@/components/zone-help';
import { zoneForRpe, zoneLabel, zoneReferenceFrom } from '@/lib/zones';
import { WorkoutSessionActions, type PrefillMetrics } from '@/components/workout-session-actions';
import { WorkoutStructure } from '@/components/workout-structure';
import { RouteMap, stepsForWorkout } from '@/components/route-map';
import { TrailSymbol } from '@/components/trail-symbol';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { currentLocale, defineMessages, formatDecimal, INTL_LOCALE } from '@/lib/i18n';
import { useLiveCyclingContext } from '@/lib/use-live-cycling-context';
import {
  activeProtection,
  parseTrainingDate,
  type TrainingPlan,
  type Workout,
  type WorkoutExplanationResponse,
} from '@/lib/planning';

type User = { display_name: string; email: string };

const messages = defineMessages({
  pt: {
    startToPrefill: 'Inicie este treino para ver os dados importados preenchidos no formulário de conclusão.',
    loadFailed: 'Não foi possível carregar seu plano.',
    explanationFailed: 'Não foi possível carregar a explicação.',
    confirmUpdate: 'Criar um novo rascunho com sua disponibilidade atual? O plano ativo continuará preservado até você revisar e aceitar o novo plano.',
    confirmReplace: 'Substituir o rascunho atual por um novo plano calculado com seus dados mais recentes?',
    generateFailed: 'Não foi possível gerar o plano.',
    activated: 'Plano ativado. Seus próximos treinos já estão disponíveis no painel.',
    activateFailed: 'Não foi possível ativar o plano.',
    reassessed: 'Reavaliamos seus próximos treinos com a sua declaração.',
    recoveredFailed: 'Não foi possível registrar sua recuperação.',
    loading: 'Carregando seu plano…',
    emptyKicker: 'PLANEJAMENTO · REGRAS V1',
    emptyTitle: 'Seu contexto já pode virar um plano.',
    emptyText: 'O Cadência usará sua experiência, objetivo, limitações e disponibilidade para criar quatro semanas explicáveis.',
    calculating: 'Calculando…',
    generateFirst: 'Gerar meu primeiro plano',
    draftNote: 'O resultado será um rascunho. Nenhum treino substitui avaliação profissional.',
    kicker: 'MEU PLANO · 4 SEMANAS',
    title: 'Uma progressão que cabe na sua rotina.',
    until: 'até',
    statusActive: 'PLANO ATIVO',
    statusCompleted: 'CICLO CONCLUÍDO',
    statusDraft: 'RASCUNHO',
    activating: 'Ativando…',
    accept: 'Aceitar plano',
    generateAnother: 'Gerar outro',
    update: 'Atualizar plano',
    nextCycle: 'Gerar próximo ciclo',
    cycleDone: 'Ciclo concluído',
    cycleDoneText: 'Seu histórico foi preservado. O próximo plano começará depois deste ciclo.',
    safetyMode: 'Modo de segurança ativo',
    safetyModeText: 'As sessões foram limitadas a esforço leve por causa da condição informada no perfil.',
    seeDashboard: 'Ver painel',
    sessions: 'sessões',
    totalVolume: 'volume total',
    daysPerWeek: 'dias por semana',
    rulesV1: 'Regras V1',
    engineUsed: 'motor utilizado',
    week: (week: number) => `SEMANA ${week}`,
    phases: ['PROGRESSÃO', 'PROGRESSÃO', 'MAIOR CARGA', 'RECUPERAÇÃO'],
    workoutDone: 'Treino concluído',
    adjusted: 'Ajustado pelo feedback',
    protection: { light: 'Proteção leve', moderate: 'Proteção moderada', strong: 'Proteção forte' } as Record<string, string>,
    closeDetail: 'Fechar detalhes do treino',
    selected: 'Sessão selecionada',
    preparing: 'Preparando explicação…',
    explain: 'Explicar a escolha',
    localAssistant: 'Explicação por IA',
    rulesExplanation: 'Explicação das regras do plano',
    structure: 'Estrutura',
    routeOf: (name: string) => `Percurso da sessão ${name}`,
    why: 'Por que este treino?',
    auditDetails: 'Ver detalhes da decisão',
    auditIntro: 'Esta sessão foi definida pelo motor de regras. A confiança ainda não é calibrada com dados longitudinais individuais.',
    dataUsed: 'Dados considerados',
    constraints: 'Restrições aplicadas',
    noConstraints: 'Nenhuma restrição adicional foi aplicada.',
    rejected: 'Alternativas descartadas',
    noRejected: 'Nenhuma alternativa adicional foi descartada.',
    missing: 'Informações ausentes',
    noMissing: 'Não há lacunas registradas para esta decisão.',
    changes: 'O que pode mudar este treino',
    evidence: 'Base científica',
    auditLabels: {
      availability_minutes: 'Tempo disponível',
      experience_level: 'Experiência declarada',
      primary_goal: 'Objetivo principal',
      secondary_goal: 'Objetivo secundário',
      current_activity_level: 'Rotina de atividade atual',
      cycling_context: 'Contexto de ciclismo informado',
      availability_preferred_time: 'Horário preferido',
      availability_location: 'Local disponível',
      observed_training_28d: 'Histórico observado dos últimos 28 dias',
      event_goal: 'Objetivo de prova',
      event_date: 'Data do evento',
      heart_rate_sensor: 'Sensor de frequência cardíaca',
      power_meter: 'Medidor de potência',
      ftp: 'FTP informado',
      event_goal_or_date: 'Objetivo ou data de prova',
      power_meter_or_ftp: 'Medidor de potência ou FTP',
      eligible_submaximal_assessment: 'Avaliação submáxima apta',
      active_safety_limitation: 'Limitação de segurança ativa',
      recent_recovery_or_pain_signal: 'Sinal recente de recuperação ou dor',
      return_after_break: 'Retorno após pausa',
      recovery_week: 'Semana de recuperação',
      low_observed_adherence: 'Baixa aderência observada',
      low_current_activity: 'Rotina atual com baixa atividade',
      event_taper: 'Taper pré-prova',
      post_event_recovery: 'Recuperação pós-prova',
      higher_intensity_protocols: 'Protocolos de maior intensidade',
      quality_session: 'Sessão de qualidade',
      long_session_above_45_minutes: 'Sessão longa acima de 45 minutos',
      additional_quality_session: 'Sessão adicional de qualidade',
      volume_progression: 'Progressão de volume',
      long_session: 'Pedal longo',
      novo_feedback_valido: 'Novo feedback válido do treino',
      mudanca_de_disponibilidade: 'Mudança de disponibilidade',
      novo_sinal_de_seguranca: 'Novo sinal de segurança',
      mudanca_no_contexto_do_evento: 'Mudança no contexto do evento',
    } as Record<string, string>,
    observedKicker: 'CONTEXTO OBSERVADO',
    observedTitle: 'O plano considerou seus registros recentes.',
    lastDays: (days: number) => `Últimos ${days} dias`,
    observedText: 'Esses dados ajudam a manter a progressão compatível com o que você vem conseguindo realizar. Eles descrevem registros do app e não são um diagnóstico.',
    completedSessions: (count: number): string => (count === 1 ? 'sessão concluída' : 'sessões concluídas'),
    done: 'realizados',
    onAverage: 'em média',
    checkins: (count: number): string => (count === 1 ? 'check-in de recuperação' : 'check-ins de recuperação'),
    needsRecovery: (pain: boolean) =>
      `O motor identificou sinais de recuperação insuficiente${pain ? ' ou dor relatada' : ''} e manteve as próximas sessões mais conservadoras.`,
    averageFatigue: (value: string) => `Fadiga média registrada: ${value}/5.`,
  },
  en: {
    startToPrefill: 'Start this workout to see the imported data filled in on the completion form.',
    loadFailed: 'Your plan could not be loaded.',
    explanationFailed: 'The explanation could not be loaded.',
    confirmUpdate: 'Create a new draft with your current availability? The active plan stays in place until you review and accept the new plan.',
    confirmReplace: 'Replace the current draft with a new plan calculated from your latest data?',
    generateFailed: 'The plan could not be generated.',
    activated: 'Plan activated. Your next workouts are now on the dashboard.',
    activateFailed: 'The plan could not be activated.',
    reassessed: 'We reassessed your next workouts with what you told us.',
    recoveredFailed: 'Your recovery could not be recorded.',
    loading: 'Loading your plan…',
    emptyKicker: 'PLANNING · RULES V1',
    emptyTitle: 'Your context is ready to become a plan.',
    emptyText: 'Cadência will use your experience, goal, limitations and availability to build four explainable weeks.',
    calculating: 'Calculating…',
    generateFirst: 'Generate my first plan',
    draftNote: 'The result will be a draft. No workout replaces a professional assessment.',
    kicker: 'MY PLAN · 4 WEEKS',
    title: 'A progression that fits your routine.',
    until: 'to',
    statusActive: 'ACTIVE PLAN',
    statusCompleted: 'CYCLE COMPLETE',
    statusDraft: 'DRAFT',
    activating: 'Activating…',
    accept: 'Accept plan',
    generateAnother: 'Generate another',
    update: 'Update plan',
    nextCycle: 'Generate next cycle',
    cycleDone: 'Cycle complete',
    cycleDoneText: 'Your history was kept. The next plan will start after this cycle.',
    safetyMode: 'Safety mode on',
    safetyModeText: 'Sessions were limited to easy effort because of the condition entered in your profile.',
    seeDashboard: 'See dashboard',
    sessions: 'sessions',
    totalVolume: 'total volume',
    daysPerWeek: 'days per week',
    rulesV1: 'Rules V1',
    engineUsed: 'engine used',
    week: (week: number) => `WEEK ${week}`,
    phases: ['BUILD', 'BUILD', 'PEAK LOAD', 'RECOVERY'],
    workoutDone: 'Workout completed',
    adjusted: 'Adjusted from feedback',
    protection: { light: 'Light protection', moderate: 'Moderate protection', strong: 'Strong protection' },
    closeDetail: 'Close workout details',
    selected: 'Selected session',
    preparing: 'Preparing explanation…',
    explain: 'Explain the choice',
    localAssistant: 'AI explanation',
    rulesExplanation: "Explanation from the plan's rules",
    structure: 'Structure',
    routeOf: (name: string) => `Route for the ${name} session`,
    why: 'Why this workout?',
    auditDetails: 'See decision details',
    auditIntro: 'This session was set by the rules engine. Its confidence is not yet calibrated with individual long-term data.',
    dataUsed: 'Data considered',
    constraints: 'Constraints applied',
    noConstraints: 'No additional constraint was applied.',
    rejected: 'Alternatives ruled out',
    noRejected: 'No additional alternative was ruled out.',
    missing: 'Missing information',
    noMissing: 'There are no recorded gaps for this decision.',
    changes: 'What can change this workout',
    evidence: 'Scientific basis',
    auditLabels: {
      availability_minutes: 'Available time',
      experience_level: 'Stated experience',
      primary_goal: 'Main goal',
      secondary_goal: 'Secondary goal',
      current_activity_level: 'Current activity routine',
      cycling_context: 'Cycling context entered',
      availability_preferred_time: 'Preferred time',
      availability_location: 'Available location',
      observed_training_28d: 'History observed over the last 28 days',
      event_goal: 'Race goal',
      event_date: 'Event date',
      heart_rate_sensor: 'Heart rate sensor',
      power_meter: 'Power meter',
      ftp: 'FTP entered',
      event_goal_or_date: 'Race goal or date',
      power_meter_or_ftp: 'Power meter or FTP',
      eligible_submaximal_assessment: 'Passed submaximal assessment',
      active_safety_limitation: 'Active safety limitation',
      recent_recovery_or_pain_signal: 'Recent recovery or pain sign',
      return_after_break: 'Return after a break',
      recovery_week: 'Recovery week',
      low_observed_adherence: 'Low observed adherence',
      low_current_activity: 'Low current activity',
      event_taper: 'Pre-race taper',
      post_event_recovery: 'Post-race recovery',
      higher_intensity_protocols: 'Higher-intensity protocols',
      quality_session: 'Quality session',
      long_session_above_45_minutes: 'Long session over 45 minutes',
      additional_quality_session: 'Additional quality session',
      volume_progression: 'Volume progression',
      long_session: 'Long ride',
      novo_feedback_valido: 'New valid workout feedback',
      mudanca_de_disponibilidade: 'Change in availability',
      novo_sinal_de_seguranca: 'New safety sign',
      mudanca_no_contexto_do_evento: 'Change in event context',
    },
    observedKicker: 'OBSERVED CONTEXT',
    observedTitle: 'The plan took your recent records into account.',
    lastDays: (days: number) => `Last ${days} days`,
    observedText: "This data helps keep the progression in line with what you have been able to do. It describes the app's records and is not a diagnosis.",
    completedSessions: (count: number) => (count === 1 ? 'completed session' : 'completed sessions'),
    done: 'done',
    onAverage: 'on average',
    checkins: (count: number) => (count === 1 ? 'recovery check-in' : 'recovery check-ins'),
    needsRecovery: (pain: boolean) =>
      `The engine found signs of insufficient recovery${pain ? ' or reported pain' : ''} and kept your next sessions more conservative.`,
    averageFatigue: (value: string) => `Average fatigue recorded: ${value}/5.`,
  },
});

type PlanMessages = (typeof messages)['pt'];

function numberParam(params: URLSearchParams, key: string): number | undefined {
  const raw = params.get(key);
  if (!raw) return undefined;
  const value = Number(raw);
  return Number.isFinite(value) ? value : undefined;
}

export default function PlanPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const liveCyclingContext = useLiveCyclingContext();
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<TrainingPlan | null>(null);
  const [selected, setSelected] = useState<Workout | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [activating, setActivating] = useState(false);
  const [recovering, setRecovering] = useState(false);
  const [mobileDetailOpen, setMobileDetailOpen] = useState(false);
  const [workoutExplanation, setWorkoutExplanation] = useState<WorkoutExplanationResponse | null>(null);
  const [explainingWorkout, setExplainingWorkout] = useState(false);
  const [explanationError, setExplanationError] = useState('');
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [prefillMetrics, setPrefillMetrics] = useState<PrefillMetrics | undefined>();

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ plan: TrainingPlan | null }>('/v1/plans/current'),
    ])
      .then(([account, current]) => {
        setUser(account.user);
        setPlan(current.plan);
        const params = new URLSearchParams(window.location.search);
        const target = current.plan?.workouts.find((workout) => workout.id === params.get('workoutID'));
        if (target) {
          setSelected(target);
          setPrefillMetrics({
            distance_km: numberParam(params, 'distance_km'),
            elevation_gain_m: numberParam(params, 'elevation_gain_m'),
            average_heart_rate: numberParam(params, 'average_heart_rate'),
            average_power_watts: numberParam(params, 'average_power_watts'),
            average_cadence_rpm: numberParam(params, 'average_cadence_rpm'),
          });
          if (target.status !== 'in_progress' && target.status !== 'completed') {
            setMessage(messages[currentLocale()].startToPrefill);
          }
          window.history.replaceState(null, '', window.location.pathname);
        } else if (current.plan?.workouts.length) {
          setSelected(current.plan.workouts[0]);
        }
      })
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, messages[currentLocale()].loadFailed));
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!mobileDetailOpen) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [mobileDetailOpen]);

  const weeks = useMemo(() => {
    if (!plan) return [];
    const start = parseTrainingDate(plan.starts_on).getTime();
    return [0, 1, 2, 3].map((week) =>
      plan.workouts.filter(
        (workout) =>
          Math.floor(
            (parseTrainingDate(workout.scheduled_on).getTime() - start) /
              604800000,
          ) === week,
      ),
    );
  }, [plan]);
  const totalMinutes = useMemo(
    () =>
      plan?.workouts.reduce(
        (sum, workout) => sum + workout.duration_minutes,
        0,
      ) || 0,
    [plan],
  );
  const evidenceByKey = useMemo(
    () => new Map((plan?.evidence || []).map((source) => [source.source_key, source])),
    [plan],
  );

  function updateSessionPlan(nextPlan: TrainingPlan, workoutID: string) {
    setPlan(nextPlan);
    setSelected(
      nextPlan.workouts.find((workout) => workout.id === workoutID) || null,
    );
    setWorkoutExplanation(null);
    setExplanationError('');
  }

  function selectWorkout(workout: Workout) {
    setSelected(workout);
    setWorkoutExplanation(null);
    setExplanationError('');
    if (window.matchMedia('(max-width: 900px)').matches) {
      setMobileDetailOpen(true);
    }
  }

  async function explainSelectedWorkout() {
    if (!selected || explainingWorkout) return;
    setExplainingWorkout(true);
    setExplanationError('');
    try {
      const result = await apiRequest<WorkoutExplanationResponse>(
        `/v1/workouts/${selected.id}/explanation`,
        { method: 'POST' },
      );
      setWorkoutExplanation(result);
    } catch (caught) {
      setExplanationError(
        caught instanceof Error
          ? caught.message
          : t.explanationFailed,
      );
    } finally {
      setExplainingWorkout(false);
    }
  }

  async function generate(replace = false) {
    if (replace) {
      const confirmation =
        plan?.status === 'active'
          ? t.confirmUpdate
          : t.confirmReplace;
      if (!window.confirm(confirmation)) return;
    }
    setGenerating(true);
    setError('');
    setMessage('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>(
        '/v1/plans/generate',
        { method: 'POST' },
      );
      setPlan(result.plan);
      setSelected(result.plan.workouts[0] || null);
      setWorkoutExplanation(null);
      setExplanationError('');
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : t.generateFailed,
      );
    } finally {
      setGenerating(false);
    }
  }

  async function activate() {
    if (!plan || plan.status !== 'draft') return;
    setActivating(true);
    setError('');
    setMessage('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>(
        `/v1/plans/${plan.id}/activate`,
        { method: 'POST' },
      );
      setPlan(result.plan);
      setSelected(
        (current) =>
          result.plan.workouts.find((workout) => workout.id === current?.id) ||
          result.plan.workouts[0] ||
          null,
      );
      setWorkoutExplanation(null);
      setExplanationError('');
      setMessage(t.activated);
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : t.activateFailed,
      );
    } finally {
      setActivating(false);
    }
  }

  // "Estou recuperado": grava a declaração e recebe o plano já reavaliado.
  async function reportRecovered() {
    setRecovering(true);
    setError('');
    setMessage('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>('/v1/protection/recovered', { method: 'POST' });
      setPlan(result.plan);
      setSelected(
        (current) =>
          result.plan.workouts.find((workout) => workout.id === current?.id) ||
          result.plan.workouts[0] ||
          null,
      );
      setWorkoutExplanation(null);
      setExplanationError('');
      setMessage(t.reassessed);
    } catch (caught) {
      setError(apiErrorMessage(caught, t.recoveredFailed));
    } finally {
      setRecovering(false);
    }
  }

  const protection = plan ? activeProtection(plan) : null;

  if (loading)
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        {t.loading}
      </main>
    );
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  const dateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: '2-digit', month: 'short' });
  const fullDateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: '2-digit', month: 'long', year: 'numeric' });

  return (
    <main className="plan-shell">
      <AppHeader name={user?.display_name} />
      <section className="plan-content">
        {!plan ? (
          <section className="plan-empty">
            <span>
              <Sparkles size={24} />
            </span>
            <p>{t.emptyKicker}</p>
            <h1>{t.emptyTitle}</h1>
            <div>{t.emptyText}</div>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <Button
              onClick={() => generate()}
              disabled={generating}
              className="plan-generate"
            >
              {generating ? (
                <LoaderCircle className="spin" />
              ) : (
                <Sparkles size={16} />
              )}
              {generating ? t.calculating : t.generateFirst}
            </Button>
            <small>{t.draftNote}</small>
          </section>
        ) : (
          <>
            <header className="plan-heading">
              <div>
                <p>{t.kicker}</p>
                <h1>{t.title}</h1>
                <span>
                  {fullDateFormatter.format(parseTrainingDate(plan.starts_on))}{' '}
                  {t.until}{' '}
                  {fullDateFormatter.format(parseTrainingDate(plan.ends_on))}
                </span>
              </div>
              <div className="plan-heading-actions">
                <span
                  className={
                    plan.status === 'active'
                      ? 'draft-pill active'
                      : 'draft-pill'
                  }
                >
                  {plan.status === 'active' ? t.statusActive : plan.status === 'completed' ? t.statusCompleted : t.statusDraft}
                </span>
                {plan.status === 'draft' && (
                  <>
                    <Button
                      onClick={activate}
                      disabled={activating || generating}
                      className="plan-activate"
                    >
                      {activating ? (
                        <LoaderCircle className="spin" size={14} />
                      ) : (
                        <CheckCircle2 size={14} />
                      )}
                      {activating ? t.activating : t.accept}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => generate(true)}
                      disabled={generating || activating}
                    >
                      <RefreshCw
                        className={generating ? 'spin' : ''}
                        size={14}
                      />{' '}
                      {t.generateAnother}
                    </Button>
                  </>
                )}
                {plan.status === 'active' && (
                  <Button
                    variant="outline"
                    onClick={() => generate(true)}
                    disabled={generating}
                  >
                    {generating ? (
                      <LoaderCircle className="spin" size={14} />
                    ) : (
                      <RefreshCw size={14} />
                    )}
                    {generating ? t.calculating : t.update}
                  </Button>
                )}
                {plan.status === 'completed' && (
                  <Button
                    onClick={() => generate()}
                    disabled={generating}
                    className="plan-activate"
                  >
                    {generating ? (
                      <LoaderCircle className="spin" size={14} />
                    ) : (
                      <Sparkles size={14} />
                    )}
                    {generating ? t.calculating : t.nextCycle}
                  </Button>
                )}
              </div>
            </header>
            {plan.status === 'completed' && (
              <div className="plan-safety">
                <CheckCircle2 size={18} />
                <div>
                  <strong>{t.cycleDone}</strong>
                  <p>{t.cycleDoneText}</p>
                </div>
              </div>
            )}
            {plan.prescription_snapshot.restricted && (
              <div className="plan-safety">
                <ShieldAlert size={18} />
                <div>
                  <strong>{t.safetyMode}</strong>
                  <p>{t.safetyModeText}</p>
                </div>
              </div>
            )}
            {protection && <ProtectionNotice protection={protection} busy={recovering} onRecovered={reportRecovered} />}
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            {message && (
              <output className="plan-message">
                <Check size={14} />
                {message}
                <Link href="/">{t.seeDashboard}</Link>
              </output>
            )}
            <ObservedTrainingCard observed={plan.prescription_snapshot.observed_training} t={t} />
            <div className="plan-stats">
              <div>
                <strong>{plan.workouts.length}</strong>
                <span>{t.sessions}</span>
              </div>
              <div>
                <strong>
                  {Math.floor(totalMinutes / 60)}h {totalMinutes % 60}min
                </strong>
                <span>{t.totalVolume}</span>
              </div>
              <div>
                <strong>{plan.prescription_snapshot.sessions_per_week}</strong>
                <span>{t.daysPerWeek}</span>
              </div>
              <div>
                <strong>{t.rulesV1}</strong>
                <span>{t.engineUsed}</span>
              </div>
            </div>
            <div className="plan-layout">
              <div className="plan-weeks">
                {weeks.map((workouts, index) => (
                  <section className="plan-week" key={index}>
                    <header>
                      <span>{t.week(index + 1)}</span>
                      <small>{t.phases[index]}</small>
                    </header>
                    <div>
                      {workouts.map((workout) => (
                        <button
                          type="button"
                          key={workout.id}
                          className={[
                            selected?.id === workout.id ? 'selected' : '',
                            workout.status === 'completed' ? 'completed' : '',
                          ].filter(Boolean).join(' ')}
                          onClick={() => selectWorkout(workout)}
                        >
                          <time>
                            {dateFormatter.format(
                              parseTrainingDate(workout.scheduled_on),
                            )}
                          </time>
                          <span>
                            <strong className="workout-name"><span>{workout.name}</span>{workout.status === 'completed' && <span className="workout-completion"><Check size={11} aria-label={t.workoutDone} /></span>}</strong>
                            <small>{workout.objective}</small>
                            {workout.explanation.adaptation && (
                              <small>
                                <Sparkles size={10} /> {t.adjusted}
                              </small>
                            )}
                            {workout.status === 'planned' &&
                              workout.explanation.protection &&
                              workout.explanation.protection.level !== 'none' && (
                                <small>
                                  <Shield size={10} /> {t.protection[workout.explanation.protection.level]}
                                </small>
                              )}
                          </span>
                          <em>
                            <TrailSymbol rpe={workout.target_rpe} />
                            {workout.duration_minutes} min · Z{zoneForRpe(workout.target_rpe).number}
                          </em>
                        </button>
                      ))}
                    </div>
                  </section>
                ))}
              </div>
              {selected && (
                <div
                  className={`workout-detail-shell${mobileDetailOpen ? ' mobile-open' : ''}`}
                  role={mobileDetailOpen ? 'dialog' : undefined}
                  aria-modal={mobileDetailOpen ? true : undefined}
                  aria-labelledby={
                    mobileDetailOpen ? 'selected-workout-title' : undefined
                  }
                >
                  <button
                    type="button"
                    className="workout-detail-backdrop"
                    aria-label={t.closeDetail}
                    onClick={() => setMobileDetailOpen(false)}
                  />
                  <aside className="workout-detail">
                  <button
                    type="button"
                    className="workout-detail-close"
                    aria-label={t.closeDetail}
                    onClick={() => setMobileDetailOpen(false)}
                  >
                    <X size={19} />
                  </button>
                  <span className="kicker">{t.selected}</span>
                  <h2 id="selected-workout-title">{selected.name}</h2>
                  <p>{selected.explanation.summary}</p>
                  <div className="workout-ai-explanation">
                    {!workoutExplanation ? (
                      <button
                        type="button"
                        className="workout-ai-trigger"
                        onClick={explainSelectedWorkout}
                        disabled={explainingWorkout}
                      >
                        {explainingWorkout ? (
                          <LoaderCircle className="spin" size={15} />
                        ) : (
                          <Sparkles size={15} />
                        )}
                        {explainingWorkout ? t.preparing : t.explain}
                      </button>
                    ) : (
                      <output className="workout-ai-result">
                        <div className="workout-ai-result-heading">
                          <Sparkles size={15} />
                          <strong>
                            {workoutExplanation.source === 'ollama' ? t.localAssistant : t.rulesExplanation}
                          </strong>
                        </div>
                        <p>{workoutExplanation.explanation}</p>
                        {workoutExplanation.warning && (
                          <small>{workoutExplanation.warning}</small>
                        )}
                      </output>
                    )}
                    {explanationError && (
                      <small className="workout-ai-error" role="alert">
                        {explanationError}
                      </small>
                    )}
                  </div>
                  <div className="detail-metrics">
                    <div>
                      <Clock3 size={15} />
                      <strong>{selected.duration_minutes} min</strong>
                    </div>
                    <div>
                      <Gauge size={15} />
                      <strong>{zoneLabel(selected.target_rpe, locale)}</strong>
                      <ZoneHelp compact />
                    </div>
                  </div>
                  <ZoneSummary
                    rpe={selected.target_rpe}
                    reference={zoneReferenceFrom(liveCyclingContext ?? plan.prescription_snapshot.cycling_context)}
                  />
                  <AdaptationCard workout={selected} />
                  <WorkoutSessionActions
                    workout={selected}
                    planStatus={plan.status}
                    usesHeartRate={Boolean(plan.prescription_snapshot.cycling_context?.uses_heart_rate)}
                    usesPower={Boolean(plan.prescription_snapshot.cycling_context?.uses_power)}
                    zoneReference={zoneReferenceFrom(liveCyclingContext ?? plan.prescription_snapshot.cycling_context)}
                    onPlanUpdated={updateSessionPlan}
                    prefillMetrics={prefillMetrics}
                  />
                  <h3>{t.structure}</h3>
                  <div className="workout-detail-map">
                    <RouteMap steps={stepsForWorkout(selected, locale)} label={t.routeOf(selected.name)} />
                  </div>
                  <WorkoutStructure structure={selected.structure} durationMinutes={selected.duration_minutes} />
                  <h3>{t.why}</h3>
                  <ul>
                    {selected.explanation.rules?.map((rule) => (
                      <li key={rule}>
                        <Check size={12} />
                        {rule}
                      </li>
                    ))}
                  </ul>
                  {selected.explanation.decision_audit && (
                    <details className="workout-decision-audit">
                      <summary>{t.auditDetails}</summary>
                      <p>{t.auditIntro}</p>
                      <DecisionAuditList
                        title={t.dataUsed}
                        values={selected.explanation.decision_audit.data_used}
                        labels={t.auditLabels}
                      />
                      <DecisionAuditList
                        title={t.constraints}
                        values={selected.explanation.decision_audit.constraints_applied}
                        labels={t.auditLabels}
                        emptyLabel={t.noConstraints}
                      />
                      <DecisionAuditList
                        title={t.rejected}
                        values={selected.explanation.decision_audit.alternatives_rejected}
                        labels={t.auditLabels}
                        emptyLabel={t.noRejected}
                      />
                      <DecisionAuditList
                        title={t.missing}
                        values={selected.explanation.decision_audit.missing_data}
                        labels={t.auditLabels}
                        emptyLabel={t.noMissing}
                      />
                      <DecisionAuditList
                        title={t.changes}
                        values={selected.explanation.decision_audit.conditions_for_change}
                        labels={t.auditLabels}
                      />
                    </details>
                  )}
                  {selected.explanation.evidence_keys?.length ? (
                    <>
                      <h3>{t.evidence}</h3>
                      <ul>
                        {selected.explanation.evidence_keys.map((key) => {
                          const source = evidenceByKey.get(key);
                          return source ? <li key={key}><a href={source.url} target="_blank" rel="noreferrer">{source.authors} ({source.published_year})</a></li> : null;
                        })}
                      </ul>
                      {selected.explanation.evidence_scope ? (
                        <p className="evidence-scope">{selected.explanation.evidence_scope}</p>
                      ) : null}
                    </>
                  ) : null}
                  </aside>
                </div>
              )}
            </div>
          </>
        )}
      </section>
    </main>
  );
}

function DecisionAuditList({
  title,
  values,
  labels,
  emptyLabel,
}: {
  title: string;
  values: string[];
  labels: Record<string, string>;
  emptyLabel?: string;
}) {
  return (
    <section className="workout-decision-audit-section">
      <strong>{title}</strong>
      {values.length ? (
        <ul>
          {values.map((value) => (
            <li key={value}>{labels[value] ?? value}</li>
          ))}
        </ul>
      ) : (
        <small>{emptyLabel}</small>
      )}
    </section>
  );
}

function ObservedTrainingCard({ observed, t }: { observed?: TrainingPlan['prescription_snapshot']['observed_training']; t: PlanMessages }) {
  const locale = useLocale();
  if (!observed || (!observed.completed_sessions && !observed.recovery_checkins)) {
    return null;
  }

  const completedSessions = observed.completed_sessions || 0;
  const completedMinutes = observed.completed_minutes || 0;
  const recoveryCheckins = observed.recovery_checkins || 0;
  const averageRPE = observed.average_rpe || 0;
  const averageFatigue = observed.average_fatigue || 0;
  const recoveryFatigue = observed.average_recovery_fatigue || 0;
  const needsRecovery = Boolean(observed.requires_recovery);

  return (
    <section className="plan-observed" aria-labelledby="plan-observed-title">
      <div className="plan-observed-heading">
        <div>
          <span>{t.observedKicker}</span>
          <h2 id="plan-observed-title">{t.observedTitle}</h2>
        </div>
        <small>{t.lastDays(observed.window_days || 28)}</small>
      </div>
      <p>{t.observedText}</p>
      <div className="plan-observed-metrics">
        {completedSessions > 0 && <span><strong>{completedSessions}</strong> {t.completedSessions(completedSessions)}</span>}
        {completedMinutes > 0 && <span><strong>{formatObservedMinutes(completedMinutes)}</strong> {t.done}</span>}
        {averageRPE > 0 && <span><strong>{zoneLabel(averageRPE, locale)}</strong> {t.onAverage}</span>}
        {recoveryCheckins > 0 && <span><strong>{recoveryCheckins}</strong> {t.checkins(recoveryCheckins)}</span>}
      </div>
      {needsRecovery && (
        <div className="plan-observed-alert">
          <ShieldAlert size={15} />
          <span>{t.needsRecovery(Boolean(observed.pain_reported))}</span>
        </div>
      )}
      {!needsRecovery && (averageFatigue > 0 || recoveryFatigue > 0) && (
        <small className="plan-observed-note">
          {t.averageFatigue(formatDecimal(averageFatigue || recoveryFatigue, 1, locale))}
        </small>
      )}
    </section>
  );
}

function formatObservedMinutes(value: number) {
  if (value < 60) return `${value} min`;
  const hours = Math.floor(value / 60);
  const minutes = value % 60;
  return minutes ? `${hours}h${minutes}min` : `${hours}h`;
}
