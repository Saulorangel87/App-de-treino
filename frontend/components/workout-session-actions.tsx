'use client';

import { useState } from 'react';
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
  skipped: 'Cancelado',
  adapted: 'Adaptado',
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
  const [feedbackOpen, setFeedbackOpen] = useState(false);
  const [actualRPE, setActualRPE] = useState(
    Math.max(1, Math.round(workout.target_rpe)),
  );
  const [difficulty, setDifficulty] =
    useState<keyof typeof difficultyLabels>('moderate');
  const [fatigueAfter, setFatigueAfter] = useState(3);
  const [painReported, setPainReported] = useState(false);
  const [notes, setNotes] = useState('');
  const [distanceKM, setDistanceKM] = useState('');
  const [elevationGainM, setElevationGainM] = useState('');
  const [averageHeartRate, setAverageHeartRate] = useState('');
  const [averagePowerW, setAveragePowerW] = useState('');
  const [error, setError] = useState('');

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
      actual_rpe: actualRPE,
      difficulty,
      fatigue_after: fatigueAfter,
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

  return (
    <section className="session-actions" aria-label="Acompanhamento da sessão">
      <header>
        <span className={`session-status ${workout.status}`}>
          {statusLabels[workout.status]}
        </span>
        {session?.started_at && workout.status === 'in_progress' && (
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

      {planStatus === 'active' && workout.status === 'planned' && (
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
          {action === 'start' ? 'Iniciando…' : 'Iniciar treino'}
        </Button>
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
              {difficultyLabels[feedback.difficulty]} · fadiga{' '}
              {feedback.fatigue_after}/5
              {feedback.pain_reported ? ' · dor relatada' : ' · sem dor'}
            </span>
            {feedback.notes && <p>{feedback.notes}</p>}
            {(session?.distance_km !== undefined || session?.elevation_gain_m !== undefined || session?.average_heart_rate !== undefined || session?.average_power_watts !== undefined) && <p className="session-metric-result">{session?.distance_km !== undefined && `${session.distance_km} km`}{session?.elevation_gain_m !== undefined && ` · ${session.elevation_gain_m} m+`}{session?.average_heart_rate !== undefined && ` · FC ${session.average_heart_rate} bpm`}{session?.average_power_watts !== undefined && ` · ${session.average_power_watts} W`}</p>}
          </div>
        </div>
      )}

      {workout.status === 'skipped' && (
        <p className="session-guidance">
          Esta sessão foi cancelada e ficou registrada no histórico.
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

function optionalNumber(value: string) {
  const normalized = value.trim().replace(',', '.');
  return normalized === '' ? undefined : Number(normalized);
}
