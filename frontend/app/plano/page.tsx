'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  Bike,
  Check,
  CheckCircle2,
  Clock3,
  Gauge,
  LoaderCircle,
  RefreshCw,
  ShieldAlert,
  Sparkles,
  X,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { AccountActions } from '@/components/account-actions';
import { AdaptationCard } from '@/components/adaptation-card';
import { RpeHelp } from '@/components/rpe-help';
import { WorkoutSessionActions } from '@/components/workout-session-actions';
import { WorkoutStructure } from '@/components/workout-structure';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import {
  parseTrainingDate,
  type TrainingPlan,
  type Workout,
  type WorkoutExplanationResponse,
} from '@/lib/planning';

type User = { display_name: string; email: string };

const dateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'short',
});
const fullDateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit',
  month: 'long',
  year: 'numeric',
});

export default function PlanPage() {
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<TrainingPlan | null>(null);
  const [selected, setSelected] = useState<Workout | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [activating, setActivating] = useState(false);
  const [mobileDetailOpen, setMobileDetailOpen] = useState(false);
  const [workoutExplanation, setWorkoutExplanation] = useState<WorkoutExplanationResponse | null>(null);
  const [explainingWorkout, setExplainingWorkout] = useState(false);
  const [explanationError, setExplanationError] = useState('');
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ plan: TrainingPlan | null }>('/v1/plans/current'),
    ])
      .then(([account, current]) => {
        setUser(account.user);
        setPlan(current.plan);
        if (current.plan?.workouts.length)
          setSelected(current.plan.workouts[0]);
      })
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, 'Não foi possível carregar seu plano.'));
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
          : 'Não foi possível carregar a explicação.',
      );
    } finally {
      setExplainingWorkout(false);
    }
  }

  async function generate(replace = false) {
    if (replace) {
      const confirmation =
        plan?.status === 'active'
          ? 'Criar um novo rascunho com sua disponibilidade atual? O plano ativo continuará preservado até você revisar e aceitar o novo plano.'
          : 'Substituir o rascunho atual por um novo plano calculado com seus dados mais recentes?';
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
          : 'Não foi possível gerar o plano.',
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
      setMessage(
        'Plano ativado. Seus próximos treinos já estão disponíveis no painel.',
      );
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : 'Não foi possível ativar o plano.',
      );
    } finally {
      setActivating(false);
    }
  }

  if (loading)
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando seu plano…
      </main>
    );
  if (!user) return <ApiErrorState message={error || 'Não foi possível carregar seu plano.'} />;

  return (
    <main className="plan-shell">
      <header className="profile-topbar">
        <Link href="/" className="account-brand dark">
          <span>
            <Bike size={19} />
          </span>
          cadência
        </Link>
        <AccountActions label="ATLETA" name={user?.display_name} />
      </header>
      <section className="plan-content">
        <Link href="/" className="back-link">
          <ArrowLeft size={15} /> Voltar ao painel
        </Link>
        {!plan ? (
          <section className="plan-empty">
            <span>
              <Sparkles size={24} />
            </span>
            <p>PLANEJAMENTO · REGRAS V1</p>
            <h1>Seu contexto já pode virar um plano.</h1>
            <div>
              O Cadência usará sua experiência, objetivo, limitações e
              disponibilidade para criar quatro semanas explicáveis.
            </div>
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
              {generating ? 'Calculando…' : 'Gerar meu primeiro plano'}
            </Button>
            <small>
              O resultado será um rascunho. Nenhum treino substitui avaliação
              profissional.
            </small>
          </section>
        ) : (
          <>
            <header className="plan-heading">
              <div>
                <p>MEU PLANO · 4 SEMANAS</p>
                <h1>Uma progressão que cabe na sua rotina.</h1>
                <span>
                  {fullDateFormatter.format(parseTrainingDate(plan.starts_on))}{' '}
                  até{' '}
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
                  {plan.status === 'active'
                    ? 'PLANO ATIVO'
                    : plan.status === 'completed'
                      ? 'CICLO CONCLUÍDO'
                      : 'RASCUNHO'}
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
                      {activating ? 'Ativando…' : 'Aceitar plano'}
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
                      Gerar outro
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
                    {generating ? 'Calculando…' : 'Atualizar plano'}
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
                    {generating ? 'Calculando…' : 'Gerar próximo ciclo'}
                  </Button>
                )}
              </div>
            </header>
            {plan.status === 'completed' && (
              <div className="plan-safety">
                <CheckCircle2 size={18} />
                <div>
                  <strong>Ciclo concluído</strong>
                  <p>
                    Seu histórico foi preservado. O próximo plano começará
                    depois deste ciclo.
                  </p>
                </div>
              </div>
            )}
            {plan.prescription_snapshot.restricted && (
              <div className="plan-safety">
                <ShieldAlert size={18} />
                <div>
                  <strong>Modo de segurança ativo</strong>
                  <p>
                    As sessões foram limitadas a esforço leve por causa da
                    condição informada no perfil.
                  </p>
                </div>
              </div>
            )}
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            {message && (
              <output className="plan-message">
                <Check size={14} />
                {message}
                <Link href="/">Ver painel</Link>
              </output>
            )}
            <ObservedTrainingCard observed={plan.prescription_snapshot.observed_training} />
            <div className="plan-stats">
              <div>
                <strong>{plan.workouts.length}</strong>
                <span>sessões</span>
              </div>
              <div>
                <strong>
                  {Math.floor(totalMinutes / 60)}h {totalMinutes % 60}min
                </strong>
                <span>volume total</span>
              </div>
              <div>
                <strong>{plan.prescription_snapshot.sessions_per_week}</strong>
                <span>dias por semana</span>
              </div>
              <div>
                <strong>Regras V1</strong>
                <span>motor utilizado</span>
              </div>
            </div>
            <div className="plan-layout">
              <div className="plan-weeks">
                {weeks.map((workouts, index) => (
                  <section className="plan-week" key={index}>
                    <header>
                      <span>SEMANA {index + 1}</span>
                      <small>
                        {index === 3
                          ? 'RECUPERAÇÃO'
                          : index === 2
                            ? 'MAIOR CARGA'
                            : 'PROGRESSÃO'}
                      </small>
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
                            <strong className="workout-name"><span>{workout.name}</span>{workout.status === 'completed' && <span className="workout-completion"><Check size={11} aria-label="Treino concluído" /></span>}</strong>
                            <small>{workout.objective}</small>
                            {workout.explanation.adaptation && (
                              <small>
                                <Sparkles size={10} /> Ajustado pelo feedback
                              </small>
                            )}
                          </span>
                          <em>
                            <Clock3 size={12} />
                            {workout.duration_minutes} min
                          </em>
                          <em>
                            <Gauge size={12} />
                            RPE {workout.target_rpe}
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
                    aria-label="Fechar detalhes do treino"
                    onClick={() => setMobileDetailOpen(false)}
                  />
                  <aside className="workout-detail">
                  <button
                    type="button"
                    className="workout-detail-close"
                    aria-label="Fechar detalhes do treino"
                    onClick={() => setMobileDetailOpen(false)}
                  >
                    <X size={19} />
                  </button>
                  <span>SESSÃO SELECIONADA</span>
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
                        {explainingWorkout
                          ? 'Preparando explicação…'
                          : 'Explicar a escolha'}
                      </button>
                    ) : (
                      <output className="workout-ai-result">
                        <div className="workout-ai-result-heading">
                          <Sparkles size={15} />
                          <strong>
                            {workoutExplanation.source === 'ollama'
                              ? 'Explicação do assistente local'
                              : 'Explicação das regras do plano'}
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
                      <strong>RPE {selected.target_rpe}</strong>
                      <RpeHelp compact />
                    </div>
                  </div>
                  <AdaptationCard workout={selected} />
                  <WorkoutSessionActions
                    workout={selected}
                    planStatus={plan.status}
                    usesHeartRate={Boolean(plan.prescription_snapshot.cycling_context?.uses_heart_rate)}
                    usesPower={Boolean(plan.prescription_snapshot.cycling_context?.uses_power)}
                    onPlanUpdated={updateSessionPlan}
                  />
                  <h3>Estrutura</h3>
                  <WorkoutStructure structure={selected.structure} durationMinutes={selected.duration_minutes} />
                  <h3>Por que este treino?</h3>
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
                      <summary>Ver detalhes da decisão</summary>
                      <p>
                        Esta sessão foi definida pelo motor de regras. A confiança
                        ainda não é calibrada com dados longitudinais individuais.
                      </p>
                      <DecisionAuditList
                        title="Dados considerados"
                        values={selected.explanation.decision_audit.data_used}
                      />
                      <DecisionAuditList
                        title="Restrições aplicadas"
                        values={selected.explanation.decision_audit.constraints_applied}
                        emptyLabel="Nenhuma restrição adicional foi aplicada."
                      />
                      <DecisionAuditList
                        title="Alternativas descartadas"
                        values={selected.explanation.decision_audit.alternatives_rejected}
                        emptyLabel="Nenhuma alternativa adicional foi descartada."
                      />
                      <DecisionAuditList
                        title="Informações ausentes"
                        values={selected.explanation.decision_audit.missing_data}
                        emptyLabel="Não há lacunas registradas para esta decisão."
                      />
                      <DecisionAuditList
                        title="O que pode mudar este treino"
                        values={selected.explanation.decision_audit.conditions_for_change}
                      />
                    </details>
                  )}
                  {selected.explanation.evidence_keys?.length ? (
                    <>
                      <h3>Base científica</h3>
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
  emptyLabel,
}: {
  title: string;
  values: string[];
  emptyLabel?: string;
}) {
  return (
    <section className="workout-decision-audit-section">
      <strong>{title}</strong>
      {values.length ? (
        <ul>
          {values.map((value) => (
            <li key={value}>{formatDecisionAuditValue(value)}</li>
          ))}
        </ul>
      ) : (
        <small>{emptyLabel}</small>
      )}
    </section>
  );
}

function formatDecisionAuditValue(value: string) {
  const labels: Record<string, string> = {
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
  };
  return labels[value] ?? value;
}

function ObservedTrainingCard({ observed }: { observed?: TrainingPlan['prescription_snapshot']['observed_training'] }) {
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
          <span>CONTEXTO OBSERVADO</span>
          <h2 id="plan-observed-title">O plano considerou seus registros recentes.</h2>
        </div>
        <small>Últimos {observed.window_days || 28} dias</small>
      </div>
      <p>
        Esses dados ajudam a manter a progressão compatível com o que você vem conseguindo realizar. Eles descrevem registros do app e não são um diagnóstico.
      </p>
      <div className="plan-observed-metrics">
        {completedSessions > 0 && <span><strong>{completedSessions}</strong> {completedSessions === 1 ? 'sessão concluída' : 'sessões concluídas'}</span>}
        {completedMinutes > 0 && <span><strong>{formatObservedMinutes(completedMinutes)}</strong> realizados</span>}
        {averageRPE > 0 && <span><strong>RPE {averageRPE.toFixed(1)}</strong> médio</span>}
        {recoveryCheckins > 0 && <span><strong>{recoveryCheckins}</strong> {recoveryCheckins === 1 ? 'check-in de recuperação' : 'check-ins de recuperação'}</span>}
      </div>
      {needsRecovery && (
        <div className="plan-observed-alert">
          <ShieldAlert size={15} />
          <span>
            O motor identificou sinais de recuperação insuficiente{observed.pain_reported ? ' ou dor relatada' : ''} e manteve as próximas sessões mais conservadoras.
          </span>
        </div>
      )}
      {!needsRecovery && (averageFatigue > 0 || recoveryFatigue > 0) && (
        <small className="plan-observed-note">
          Fadiga média registrada: {(averageFatigue || recoveryFatigue).toFixed(1)}/5.
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
