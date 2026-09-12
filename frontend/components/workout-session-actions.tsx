'use client';

import { useEffect, useState } from 'react';
import {
  CheckCircle2,
  CircleStop,
  LoaderCircle,
  Play,
  RotateCcw,
  TriangleAlert,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { apiRequest } from '@/lib/api';
import type { TrainingPlan, Workout } from '@/lib/planning';

type Props = {
  workout: Workout;
  planStatus: TrainingPlan['status'];
  usesHeartRate?: boolean;
  usesPower?: boolean;
  onPlanUpdated: (plan: TrainingPlan, workoutID: string) => void;
};

const difficultyLabels = {
  very_easy: 'Muito fácil',
  easy: 'Fácil',
  moderate: 'Moderado',
  hard: 'Difícil',
  very_hard: 'Muito difícil',
} as const;

const statusLabels = {
  planned: 'Planejado',
  in_progress: 'Em andamento',
  completed: 'Concluído',
  skipped: 'Não realizado',
  adapted: 'Adaptado',
} as const;

const partialReasonLabels = {
  time_available_changed: 'Fiquei sem tempo',
  fatigue_or_recovery: 'Fadiga ou recuperação',
  pain_or_discomfort: 'Dor ou desconforto',
  equipment_or_conditions: 'Equipamento, clima ou terreno',
  other: 'Outro motivo',
} as const;

const timeFormatter = new Intl.DateTimeFormat('pt-BR', {
  hour: '2-digit',
  minute: '2-digit',
});

export function WorkoutSessionActions({
  workout,
  planStatus,
  usesHeartRate = false,
  usesPower = false,
  onPlanUpdated,
}: Props) {
  const [action, setAction] = useState('');
  const [todayKey, setTodayKey] = useState('');
  const [feedbackOpen, setFeedbackOpen] = useState(false);
  const [completionStatus, setCompletionStatus] = useState<'complete' | 'partial'>('complete');
  const [partialReason, setPartialReason] = useState<keyof typeof partialReasonLabels | ''>('');
  const [actualRPE, setActualRPE] = useState(
    Math.max(1, Math.round(workout.target_rpe)),
  );
  const [difficulty, setDifficulty] =
    useState<keyof typeof difficultyLabels>('moderate');
  const [fatigueAfter, setFatigueAfter] = useState(3);
  const [recoveryAfter, setRecoveryAfter] = useState(3);
  const [repeatConfidence, setRepeatConfidence] = useState(3);
  const [painReported, setPainReported] = useState(false);
  const [notes, setNotes] = useState('');
  const [distanceKM, setDistanceKM] = useState('');
  const [elevationGainM, setElevationGainM] = useState('');
  const [averageHeartRate, setAverageHeartRate] = useState('');
  const [averagePowerW, setAveragePowerW] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    queueMicrotask(() => setTodayKey(localDateKey()));
  }, []);

  async function mutate(path: string, body?: object) {
    setAction(path);
    setError('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>(
        `/v1/workouts/${workout.id}/${path}`,
        {
          method: 'POST',
          ...(body ? { body: JSON.stringify(body) } : {}),
        },
      );
      onPlanUpdated(result.plan, workout.id);
      setFeedbackOpen(false);
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : 'Não foi possível atualizar a sessão.',
      );
    } finally {
      setAction('');
    }
  }

  async function complete() {
    await mutate('complete', {
      completion_status: completionStatus,
      partial_reason: completionStatus === 'partial' ? partialReason : undefined,
      actual_rpe: actualRPE,
      difficulty,
      fatigue_after: fatigueAfter,
      recovery_after: recoveryAfter,
      repeat_confidence: repeatConfidence,
      pain_reported: painReported,
      notes,
      distance_km: optionalNumber(distanceKM),
      elevation_gain_m: optionalNumber(elevationGainM),
      average_heart_rate: usesHeartRate ? optionalNumber(averageHeartRate) : undefined,
      average_power_watts: usesPower ? optionalNumber(averagePowerW) : undefined,
    });
  }

  const session = workout.session;
  const feedback = session?.feedback;
  const busy = Boolean(action);
  const isPastWorkout = todayKey !== '' && workout.scheduled_on < todayKey;
  const canMarkMissed =
    planStatus === 'active' &&
    isPastWorkout &&
    (workout.status === 'planned' || workout.status === 'adapted');
  const canStart =
    planStatus === 'active' &&
    (workout.status === 'planned' || workout.status === 'adapted');
  const startLabel =
    workout.status === 'adapted' ? 'Iniciar treino adaptado' : 'Iniciar treino';

  return (
    <section className="session-actions" aria-label="Acompanhamento da sessão">
      <header>
        <span className={`session-status ${workout.status}`}>
          {statusLabels[workout.status]}
        </span>
        {session?.started_at && workout.status === 'in_progress' && todayKey !== '' && (
          <small>
            Iniciado às {timeFormatter.format(new Date(session.started_at))}
          </small>
        )}
      </header>

      {planStatus === 'draft' && (
        <p className="session-guidance">
          Aceite o plano antes de iniciar esta sessão.
        </p>
      )}

      {canStart && (
        <div className={canMarkMissed ? 'session-button-row' : undefined}>
          <Button
            type="button"
            className="session-primary"
            disabled={busy}
            onClick={() => mutate('start')}
          >
            {action === 'start' ? (
              <LoaderCircle className="spin" />
            ) : (
              <Play />
            )}
            {action === 'start' ? 'Iniciando…' : startLabel}
          </Button>
          {canMarkMissed && (
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={() => {
                if (
                  window.confirm(
                    'Marcar este treino como não realizado? Ele ficará registrado como perdido, sem criar uma sessão substituta ou aumentar a carga seguinte.',
                  )
                ) {
                  void mutate('missed');
                }
              }}
            >
              {action === 'missed' ? (
                <LoaderCircle className="spin" />
              ) : (
                <CircleStop />
              )}
              {action === 'missed' ? 'Registrando…' : 'Não realizei'}
            </Button>
          )}
        </div>
      )}

      {workout.status === 'in_progress' && !feedbackOpen && (
        <div className="session-button-row">
          <Button
            type="button"
            className="session-primary"
            disabled={busy}
            onClick={() => setFeedbackOpen(true)}
          >
            <CheckCircle2 />
            Concluir treino
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={busy}
            onClick={() => {
              if (
                window.confirm(
                  'Cancelar esta sessão? O treino será marcado como não realizado.',
                )
              ) {
                void mutate('cancel');
              }
            }}
          >
            {action === 'cancel' ? (
              <LoaderCircle className="spin" />
            ) : (
              <CircleStop />
            )}
            Cancelar
          </Button>
        </div>
      )}

      {workout.status === 'in_progress' && feedbackOpen && (
        <form
          className="session-feedback"
          onSubmit={(event) => {
            event.preventDefault();
            void complete();
          }}
        >
          <div className="feedback-heading">
            <div>
              <strong>Como foi o treino?</strong>
              <small>Seu relato será usado na adaptação futura.</small>
            </div>
            <button
              type="button"
              onClick={() => setFeedbackOpen(false)}
              aria-label="Voltar para as ações da sessão"
            >
              <RotateCcw />
            </button>
          </div>

          <fieldset className="completion-context">
            <legend>Quanto do treino você realizou?</legend>
            <div className="feedback-grid">
              <label>
                Conclusão
                <select
                  value={completionStatus}
                  onChange={(event) => {
                    const value = event.target.value as 'complete' | 'partial';
                    setCompletionStatus(value);
                    if (value === 'complete') {
                      setPartialReason('');
                    }
                  }}
                >
                  <option value="complete">Treino completo</option>
                  <option value="partial">Fiz apenas parte</option>
                </select>
              </label>
              {completionStatus === 'partial' && (
                <label>
                  Motivo
                  <select
                    required
                    value={partialReason}
                    onChange={(event) =>
                      setPartialReason(
                        event.target.value as keyof typeof partialReasonLabels,
                      )
                    }
                  >
                    <option value="">
                      Selecione o motivo
                    </option>
                    {Object.entries(partialReasonLabels).map(([value, label]) => (
                      <option value={value} key={value}>
                        {label}
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </div>
            {completionStatus === 'partial' && (
              <small className="completion-help">
                Esse contexto evita interpretar a sessão como tolerância ao treino completo.
              </small>
            )}
          </fieldset>

          <label htmlFor={`rpe-${workout.id}`}>
            RPE realizado
            <output>{actualRPE}</output>
          </label>
          <input
            id={`rpe-${workout.id}`}
            type="range"
            min="1"
            max="10"
            step="1"
            value={actualRPE}
            onChange={(event) => setActualRPE(Number(event.target.value))}
          />
          <div className="feedback-scale">
            <span>1 · muito leve</span>
            <span>10 · máximo</span>
          </div>

          <div className="feedback-grid">
            <label>
              Dificuldade
              <select
                value={difficulty}
                onChange={(event) =>
                  setDifficulty(
                    event.target.value as keyof typeof difficultyLabels,
                  )
                }
              >
                {Object.entries(difficultyLabels).map(([value, label]) => (
                  <option value={value} key={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Fadiga depois
              <select
                value={fatigueAfter}
                onChange={(event) =>
                  setFatigueAfter(Number(event.target.value))
                }
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {value} de 5
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="feedback-grid">
            <label>
              Recuperação percebida
              <select
                value={recoveryAfter}
                onChange={(event) => setRecoveryAfter(Number(event.target.value))}
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {value} de 5
                  </option>
                ))}
              </select>
            </label>
            <label>
              Confiança para repetir
              <select
                value={repeatConfidence}
                onChange={(event) => setRepeatConfidence(Number(event.target.value))}
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {value} de 5
                  </option>
                ))}
              </select>
            </label>
          </div>

          <fieldset className="session-metrics">
            <legend>Dados do pedal <small>opcionais</small></legend>
            <div className="feedback-grid">
              <label>
                Distância (km)
                <input type="number" min="0" max="2000" step="0.1" inputMode="decimal" value={distanceKM} onChange={(event) => setDistanceKM(event.target.value)} placeholder="Ex.: 32,5" />
              </label>
              <label>
                Ganho de elevação (m)
                <input type="number" min="0" max="20000" step="1" inputMode="numeric" value={elevationGainM} onChange={(event) => setElevationGainM(event.target.value)} placeholder="Ex.: 420" />
              </label>
              {usesHeartRate && <label>
                FC média (bpm)
                <input type="number" min="30" max="250" step="1" inputMode="numeric" value={averageHeartRate} onChange={(event) => setAverageHeartRate(event.target.value)} placeholder="Ex.: 142" />
              </label>}
              {usesPower && <label>
                Potência média (W)
                <input type="number" min="0" max="2000" step="1" inputMode="numeric" value={averagePowerW} onChange={(event) => setAveragePowerW(event.target.value)} placeholder="Ex.: 185" />
              </label>}
            </div>
          </fieldset>

          <label className="pain-check">
            <input
              type="checkbox"
              checked={painReported}
              onChange={(event) => setPainReported(event.target.checked)}
            />
            Senti dor durante ou depois do treino
          </label>
          {painReported && (
            <p className="pain-warning">
              <TriangleAlert />
              Dor será tratada como sinal de segurança nas próximas adaptações.
            </p>
          )}

          <label htmlFor={`notes-${workout.id}`}>Observações opcionais</label>
          <Textarea
            id={`notes-${workout.id}`}
            maxLength={1000}
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            placeholder="Terreno, clima, desconforto ou algo que influenciou o esforço."
          />

          <Button type="submit" className="session-primary" disabled={busy}>
            {action === 'complete' ? (
              <LoaderCircle className="spin" />
            ) : (
              <CheckCircle2 />
            )}
            {action === 'complete'
              ? 'Salvando feedback…'
              : 'Salvar e concluir'}
          </Button>
        </form>
      )}

      {workout.status === 'completed' && feedback && (
        <div className="session-result">
          <CheckCircle2 />
          <div>
            <strong>Treino concluído · RPE {session?.actual_rpe}</strong>
            <span>
              {feedback.completion_status === 'partial'
                ? `Conclusão parcial · ${feedback.partial_reason ? partialReasonLabels[feedback.partial_reason] : 'motivo não informado'}`
                : 'Treino completo'}{' '}
              ·{' '}
              {difficultyLabels[feedback.difficulty]} · fadiga{' '}
              {feedback.fatigue_after}/5
              {feedback.recovery_after !== undefined && ` · recuperação ${feedback.recovery_after}/5`}
              {feedback.repeat_confidence !== undefined && ` · confiança ${feedback.repeat_confidence}/5`}
              {feedback.pain_reported ? ' · dor relatada' : ' · sem dor'}
            </span>
            {feedback.notes && <p>{feedback.notes}</p>}
            {(session?.distance_km !== undefined || session?.elevation_gain_m !== undefined || session?.average_heart_rate !== undefined || session?.average_power_watts !== undefined) && <p className="session-metric-result">{session?.distance_km !== undefined && `${session.distance_km} km`}{session?.elevation_gain_m !== undefined && ` · ${session.elevation_gain_m} m+`}{session?.average_heart_rate !== undefined && ` · FC ${session.average_heart_rate} bpm`}{session?.average_power_watts !== undefined && ` · ${session.average_power_watts} W`}</p>}
          </div>
        </div>
      )}

      {workout.status === 'skipped' && (
        <p className="session-guidance">
          Esta sessão não foi realizada e ficou registrada no histórico.
        </p>
      )}

      {error && (
        <output className="session-error">
          <TriangleAlert />
          {error}
        </output>
      )}
    </section>
  );
}

function localDateKey(date = new Date()) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function optionalNumber(value: string) {
  const normalized = value.trim().replace(',', '.');
  return normalized === '' ? undefined : Number(normalized);
}
