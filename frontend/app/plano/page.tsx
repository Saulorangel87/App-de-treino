'use client';

import { useEffect, useMemo, useState } from 'react';
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
} from 'lucide-react';
import { Button } from '@/components/ui/button';
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
      .catch(() => {
        window.location.href = '/entrar';
      })
      .finally(() => setLoading(false));
  }, []);

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

  function updateSessionPlan(nextPlan: TrainingPlan, workoutID: string) {
    setPlan(nextPlan);
    setSelected(
      nextPlan.workouts.find((workout) => workout.id === workoutID) || null,
    );
  }

  async function generate(replace = false) {
    if (
      replace &&
      !window.confirm(
        'Substituir o rascunho atual por um novo plano calculado com seus dados mais recentes?',
      )
    )
      return;
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

  return (
    <main className="plan-shell">
      <header className="profile-topbar">
        <a href="/" className="account-brand dark">
          <span>
            <Bike size={19} />
          </span>
          cadência
        </a>
        <div>
          <small>ATLETA</small>
          <strong>{user?.display_name}</strong>
        </div>
      </header>
      <section className="plan-content">
        <a href="/" className="back-link">
          <ArrowLeft size={15} /> Voltar ao painel
        </a>
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
                  {plan.status === 'active' ? 'PLANO ATIVO' : 'RASCUNHO'}
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
              </div>
            </header>
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
                <a href="/">Ver painel</a>
              </output>
            )}
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
                          className={
                            selected?.id === workout.id ? 'selected' : ''
                          }
                          onClick={() => setSelected(workout)}
                        >
                          <time>
                            {dateFormatter.format(
                              parseTrainingDate(workout.scheduled_on),
                            )}
                          </time>
                          <span>
                            <strong>{workout.name}</strong>
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
                <aside className="workout-detail">
                  <span>SESSÃO SELECIONADA</span>
                  <h2>{selected.name}</h2>
                  <p>{selected.explanation.summary}</p>
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
                    onPlanUpdated={updateSessionPlan}
                  />
                  <h3>Estrutura</h3>
                  <ol>
                    <li>
                      <b>{selected.structure.warmup_minutes} min</b>Aquecimento
                      progressivo
                    </li>
                    <li>
                      <b>Principal</b>
                      {selected.structure.main}
                    </li>
                    <li>
                      <b>{selected.structure.cooldown_minutes} min</b>
                      Desaquecimento leve
                    </li>
                  </ol>
                  <h3>Por que este treino?</h3>
                  <ul>
                    {selected.explanation.rules?.map((rule) => (
                      <li key={rule}>
                        <Check size={12} />
                        {rule}
                      </li>
                    ))}
                  </ul>
                </aside>
              )}
            </div>
          </>
        )}
      </section>
    </main>
  );
}
