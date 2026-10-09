'use client';

import { useEffect, useRef, useState } from 'react';
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
import { defineMessages, formatDecimal, INTL_LOCALE } from '@/lib/i18n';
import {
  MAX_LOGGED_DURATION_MINUTES,
  durationSourceLabel,
  formatTrainingDay,
  isFutureTrainingDate,
  loggableDateRange,
  undoConfirmation,
  type TrainingPlan,
  type Workout,
} from '@/lib/planning';
import {
  ZONES,
  formatHeartRateRange,
  formatPowerRange,
  zoneForRpe,
  zoneLabel,
  zoneName,
  zoneRanges,
  zoneTalk,
  type ZoneReference,
} from '@/lib/zones';
import { useLocale, useMessages } from './locale-provider';

export type PrefillMetrics = {
  distance_km?: number;
  elevation_gain_m?: number;
  average_heart_rate?: number;
  average_power_watts?: number;
  average_cadence_rpm?: number;
};

type Props = {
  workout: Workout;
  planStatus: TrainingPlan['status'];
  usesHeartRate?: boolean;
  usesPower?: boolean;
  /** Frequência máxima e FTP do atleta, para mostrar cada zona em bpm e watts. */
  zoneReference?: ZoneReference;
  onPlanUpdated: (plan: TrainingPlan, workoutID: string) => void;
  /**
   * Preenche o formulário de conclusão (ou de correção, se a sessão já
   * estiver concluída) com dados extraídos de uma atividade importada
   * (.fit/.gpx) e abre o formulário automaticamente. O atleta ainda revisa e
   * confirma manualmente — nada é enviado sozinho.
   */
  prefillMetrics?: PrefillMetrics;
};

type Difficulty = 'very_easy' | 'easy' | 'moderate' | 'hard' | 'very_hard';
type PartialReason = 'time_available_changed' | 'fatigue_or_recovery' | 'pain_or_discomfort' | 'equipment_or_conditions' | 'other';
type Terrain = 'flat' | 'rolling' | 'hilly' | 'mixed' | 'technical' | 'indoor';
type ExternalCondition = 'normal' | 'heat' | 'cold' | 'wind' | 'rain' | 'poor_visibility' | 'other';

const messages = defineMessages({
  pt: {
    difficulty: { very_easy: 'Muito fácil', easy: 'Fácil', moderate: 'Moderado', hard: 'Difícil', very_hard: 'Muito difícil' } as Record<Difficulty, string>,
    status: { planned: 'Planejado', in_progress: 'Em andamento', completed: 'Concluído', skipped: 'Não realizado', adapted: 'Adaptado' } as Record<Workout['status'], string>,
    partialReasons: {
      time_available_changed: 'Fiquei sem tempo',
      fatigue_or_recovery: 'Fadiga ou recuperação',
      pain_or_discomfort: 'Dor ou desconforto',
      equipment_or_conditions: 'Equipamento, clima ou terreno',
      other: 'Outro motivo',
    } as Record<PartialReason, string>,
    terrain: { flat: 'Plano', rolling: 'Ondulado', hilly: 'Montanhoso', mixed: 'Misto', technical: 'Técnico', indoor: 'Indoor' } as Record<Terrain, string>,
    conditions: {
      normal: 'Condições normais', heat: 'Calor', cold: 'Frio', wind: 'Vento', rain: 'Chuva', poor_visibility: 'Baixa visibilidade', other: 'Outra condição',
    } as Record<ExternalCondition, string>,
    updateFailed: 'Não foi possível atualizar a sessão.',
    durationRange: (max: number) => `Informe a duração em minutos, de 1 a ${max}.`,
    correctionSaved: 'Correção salva. O registro foi reavaliado sem recalcular o plano.',
    correctionFailed: 'Não foi possível corrigir os dados do treino.',
    startAdapted: 'Iniciar treino adaptado',
    start: 'Iniciar treino',
    label: 'Acompanhamento da sessão',
    startedAt: (time: string) => `Iniciado às ${time}`,
    acceptFirst: 'Aceite o plano antes de iniciar esta sessão.',
    futureWorkout: (day: string) => `Este treino é de ${day}. Ele fica disponível no dia planejado.`,
    starting: 'Iniciando…',
    markDone: 'Marcar como feito',
    confirmMissed: 'Marcar este treino como não realizado? Ele ficará registrado como perdido, sem criar uma sessão substituta ou aumentar a carga seguinte.',
    recording: 'Registrando…',
    missed: 'Não realizei',
    complete: 'Concluir treino',
    confirmCancel: 'Cancelar esta sessão? O treino será marcado como não realizado.',
    cancel: 'Cancelar',
    logTitle: 'Registrar treino feito',
    howWasIt: 'Como foi o treino?',
    usedForAdaptation: 'Seu relato será usado na adaptação futura.',
    backToActions: 'Voltar para as ações da sessão',
    howLongWhen: 'Quanto tempo e quando?',
    durationMin: 'Duração (min)',
    workoutDay: 'Dia do treino',
    logHelp: (minutes: number) =>
      `Planejado: ${minutes} min. Você pode registrar até 7 dias depois. Para o próximo treino subir de carga, o app considera pelo menos 80% do tempo planejado.`,
    howMuch: 'Quanto do treino você realizou?',
    completion: 'Conclusão',
    fullWorkout: 'Treino completo',
    partOnly: 'Fiz apenas parte',
    reason: 'Motivo',
    selectReason: 'Selecione o motivo',
    partialHelp: 'Esse contexto evita interpretar a sessão como tolerância ao treino completo.',
    zoneQuestion: 'Em que zona você pedalou?',
    zoneHelp: 'Pense na parte principal do treino. Sem sensor, use o teste da conversa: ele diz qual zona combina com o que você sentiu.',
    planned: 'planejada',
    difficultyLabel: 'Dificuldade',
    fatigueAfter: 'Fadiga depois',
    outOf5: (value: number) => `${value} de 5`,
    satisfaction: 'Satisfação com a sessão',
    satisfactionOptions: ['Muito baixa', 'Baixa', 'Neutra', 'Boa', 'Muito boa'],
    terrainLabel: 'Terreno',
    notInformed: 'Não informado',
    conditionsLabel: 'Condições externas',
    equipment: 'Equipamento utilizado',
    optional: 'opcional',
    equipmentPlaceholder: 'Ex.: bike de estrada, rolo ou sensor',
    recoveryAfter: 'Recuperação percebida',
    repeatConfidence: 'Confiança para repetir',
    rideData: 'Dados do pedal',
    optionals: 'opcionais',
    distance: 'Distância (km)',
    elevation: 'Ganho de elevação (m)',
    averageHr: 'FC média (bpm)',
    averagePower: 'Potência média (W)',
    averageCadence: 'Cadência média (rpm)',
    example: (value: string) => `Ex.: ${value}`,
    pain: 'Senti dor durante ou depois do treino',
    painWarning: 'Dor será tratada como sinal de segurança nas próximas adaptações.',
    notes: 'Observações opcionais',
    notesPlaceholder: 'Terreno, clima, desconforto ou algo que influenciou o esforço.',
    saving: 'Salvando…',
    saveLog: 'Salvar treino feito',
    saveComplete: 'Salvar e concluir',
    completed: 'Treino concluído',
    noZone: 'zona não informada',
    durationSource: (minutes: number, source: string) => `${minutes} min · duração ${source}`,
    partialSummary: (reason: string) => `Conclusão parcial · ${reason}`,
    noReason: 'motivo não informado',
    fatigue: 'fadiga',
    recovery: 'recuperação',
    confidence: 'confiança',
    satisfactionShort: 'satisfação',
    equipmentShort: 'equipamento',
    painReported: 'dor relatada',
    noPain: 'sem dor',
    hr: 'FC',
    reviewTitle: 'Registro salvo para revisão',
    inconsistent: 'Há dados incompatíveis no registro. Ele foi preservado, mas não entra na observação do histórico.',
    incomplete: 'Faltam dados mínimos no registro. Ele foi preservado, mas não entra na observação do histórico.',
    correct: 'Corrigir dados do pedal',
    correctHelp: 'Apague um campo para removê-lo. Duração, zona e feedback não serão alterados.',
    cancelCorrection: 'Cancelar correção',
    savingCorrection: 'Salvando correção…',
    saveCorrection: 'Salvar correção',
    skipped: 'Esta sessão não foi realizada e ficou registrada no histórico.',
    undoing: 'Desfazendo…',
    undo: 'Desfazer registro',
    reopen: 'Reabrir treino',
  },
  en: {
    difficulty: { very_easy: 'Very easy', easy: 'Easy', moderate: 'Moderate', hard: 'Hard', very_hard: 'Very hard' },
    status: { planned: 'Planned', in_progress: 'In progress', completed: 'Completed', skipped: 'Missed', adapted: 'Adapted' },
    partialReasons: {
      time_available_changed: 'I ran out of time',
      fatigue_or_recovery: 'Fatigue or recovery',
      pain_or_discomfort: 'Pain or discomfort',
      equipment_or_conditions: 'Equipment, weather or terrain',
      other: 'Other reason',
    },
    terrain: { flat: 'Flat', rolling: 'Rolling', hilly: 'Hilly', mixed: 'Mixed', technical: 'Technical', indoor: 'Indoor' },
    conditions: {
      normal: 'Normal conditions', heat: 'Heat', cold: 'Cold', wind: 'Wind', rain: 'Rain', poor_visibility: 'Poor visibility', other: 'Other condition',
    },
    updateFailed: 'The session could not be updated.',
    durationRange: (max: number) => `Enter the duration in minutes, from 1 to ${max}.`,
    correctionSaved: 'Correction saved. The log was reassessed without recalculating the plan.',
    correctionFailed: 'The workout data could not be corrected.',
    startAdapted: 'Start adapted workout',
    start: 'Start workout',
    label: 'Session tracking',
    startedAt: (time: string) => `Started at ${time}`,
    acceptFirst: 'Accept the plan before starting this session.',
    futureWorkout: (day: string) => `This workout is for ${day}. It becomes available on the planned day.`,
    starting: 'Starting…',
    markDone: 'Mark as done',
    confirmMissed: 'Mark this workout as missed? It will be logged as missed, without creating a make-up session or increasing the next load.',
    recording: 'Saving…',
    missed: "I didn't do it",
    complete: 'Complete workout',
    confirmCancel: 'Cancel this session? The workout will be marked as missed.',
    cancel: 'Cancel',
    logTitle: 'Log a completed workout',
    howWasIt: 'How was the workout?',
    usedForAdaptation: 'Your report will be used for future adaptation.',
    backToActions: 'Back to the session actions',
    howLongWhen: 'How long and when?',
    durationMin: 'Duration (min)',
    workoutDay: 'Workout day',
    logHelp: (minutes: number) =>
      `Planned: ${minutes} min. You can log it up to 7 days later. For the next workout to increase the load, the app needs at least 80% of the planned time.`,
    howMuch: 'How much of the workout did you do?',
    completion: 'Completion',
    fullWorkout: 'Full workout',
    partOnly: 'Only part of it',
    reason: 'Reason',
    selectReason: 'Select the reason',
    partialHelp: 'This context keeps the session from being read as tolerance to the full workout.',
    zoneQuestion: 'Which zone did you ride in?',
    zoneHelp: 'Think about the main set. Without a sensor, use the talk test: it tells you which zone matches what you felt.',
    planned: 'planned',
    difficultyLabel: 'Difficulty',
    fatigueAfter: 'Fatigue afterward',
    outOf5: (value: number) => `${value} out of 5`,
    satisfaction: 'Satisfaction with the session',
    satisfactionOptions: ['Very low', 'Low', 'Neutral', 'Good', 'Very good'],
    terrainLabel: 'Terrain',
    notInformed: 'Not entered',
    conditionsLabel: 'Outside conditions',
    equipment: 'Equipment used',
    optional: 'optional',
    equipmentPlaceholder: 'E.g.: road bike, trainer or sensor',
    recoveryAfter: 'Perceived recovery',
    repeatConfidence: 'Confidence to repeat',
    rideData: 'Ride data',
    optionals: 'optional',
    distance: 'Distance (km)',
    elevation: 'Elevation gain (m)',
    averageHr: 'Average HR (bpm)',
    averagePower: 'Average power (W)',
    averageCadence: 'Average cadence (rpm)',
    example: (value: string) => `E.g.: ${value}`,
    pain: 'I felt pain during or after the workout',
    painWarning: 'Pain will be treated as a safety sign in the next adaptations.',
    notes: 'Optional notes',
    notesPlaceholder: 'Terrain, weather, discomfort or anything that affected the effort.',
    saving: 'Saving…',
    saveLog: 'Save completed workout',
    saveComplete: 'Save and complete',
    completed: 'Workout completed',
    noZone: 'zone not entered',
    durationSource: (minutes: number, source: string) => `${minutes} min · duration ${source}`,
    partialSummary: (reason: string) => `Partial completion · ${reason}`,
    noReason: 'no reason given',
    fatigue: 'fatigue',
    recovery: 'recovery',
    confidence: 'confidence',
    satisfactionShort: 'satisfaction',
    equipmentShort: 'equipment',
    painReported: 'pain reported',
    noPain: 'no pain',
    hr: 'HR',
    reviewTitle: 'Log saved for review',
    inconsistent: 'The log has data that does not match. It was kept, but it does not count toward your history.',
    incomplete: 'The log is missing minimum data. It was kept, but it does not count toward your history.',
    correct: 'Correct ride data',
    correctHelp: 'Clear a field to remove it. Duration, zone and feedback will not change.',
    cancelCorrection: 'Cancel correction',
    savingCorrection: 'Saving correction…',
    saveCorrection: 'Save correction',
    skipped: 'This session was not done and is recorded in your history.',
    undoing: 'Undoing…',
    undo: 'Undo log',
    reopen: 'Reopen workout',
  },
});

export function WorkoutSessionActions({
  workout,
  planStatus,
  usesHeartRate = false,
  usesPower = false,
  zoneReference = {},
  onPlanUpdated,
  prefillMetrics,
}: Props) {
  const locale = useLocale();
  const t = useMessages(messages);
  const [action, setAction] = useState('');
  const [todayKey, setTodayKey] = useState('');
  const [feedbackOpen, setFeedbackOpen] = useState(false);
  const [logOpen, setLogOpen] = useState(false);
  const [logDuration, setLogDuration] = useState('');
  const [logDate, setLogDate] = useState('');
  const [completionStatus, setCompletionStatus] = useState<'complete' | 'partial'>('complete');
  const [partialReason, setPartialReason] = useState<PartialReason | ''>('');
  const plannedZone = zoneForRpe(workout.target_rpe).number;
  const [actualZone, setActualZone] = useState<number>(plannedZone);
  const [difficulty, setDifficulty] = useState<Difficulty>('moderate');
  const [fatigueAfter, setFatigueAfter] = useState(3);
  const [recoveryAfter, setRecoveryAfter] = useState(3);
  const [repeatConfidence, setRepeatConfidence] = useState(3);
  const [satisfaction, setSatisfaction] = useState(3);
  const [terrain, setTerrain] = useState<Terrain | ''>('');
  const [externalConditions, setExternalConditions] = useState<ExternalCondition | ''>('');
  const [equipmentUsed, setEquipmentUsed] = useState('');
  const [painReported, setPainReported] = useState(false);
  const [notes, setNotes] = useState('');
  const [distanceKM, setDistanceKM] = useState('');
  const [elevationGainM, setElevationGainM] = useState('');
  const [averageHeartRate, setAverageHeartRate] = useState('');
  const [averagePowerW, setAveragePowerW] = useState('');
  const [averageCadenceRPM, setAverageCadenceRPM] = useState('');
  const [correctionOpen, setCorrectionOpen] = useState(false);
  const [correctionDistanceKM, setCorrectionDistanceKM] = useState('');
  const [correctionElevationGainM, setCorrectionElevationGainM] = useState('');
  const [correctionAverageHeartRate, setCorrectionAverageHeartRate] = useState('');
  const [correctionAveragePowerW, setCorrectionAveragePowerW] = useState('');
  const [correctionAverageCadenceRPM, setCorrectionAverageCadenceRPM] = useState('');
  const [correctionNotice, setCorrectionNotice] = useState('');
  const [error, setError] = useState('');
  const appliedPrefillRef = useRef(false);

  useEffect(() => {
    queueMicrotask(() => setTodayKey(localDateKey()));
  }, []);

  useEffect(() => {
    if (!prefillMetrics || appliedPrefillRef.current) return;
    appliedPrefillRef.current = true;
    const asText = (value?: number) => (value !== undefined ? String(value) : '');
    queueMicrotask(() => {
      if (workout.status === 'completed') {
        setCorrectionDistanceKM(asText(prefillMetrics.distance_km));
        setCorrectionElevationGainM(asText(prefillMetrics.elevation_gain_m));
        setCorrectionAverageHeartRate(asText(prefillMetrics.average_heart_rate));
        setCorrectionAveragePowerW(asText(prefillMetrics.average_power_watts));
        setCorrectionAverageCadenceRPM(asText(prefillMetrics.average_cadence_rpm));
        setCorrectionOpen(true);
      } else {
        setDistanceKM(asText(prefillMetrics.distance_km));
        setElevationGainM(asText(prefillMetrics.elevation_gain_m));
        setAverageHeartRate(asText(prefillMetrics.average_heart_rate));
        setAveragePowerW(asText(prefillMetrics.average_power_watts));
        setAverageCadenceRPM(asText(prefillMetrics.average_cadence_rpm));
        setFeedbackOpen(true);
      }
    });
  }, [prefillMetrics, workout.status]);

  async function mutate(path: string, body?: object, query = ''): Promise<boolean> {
    setAction(path);
    setError('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>(
        `/v1/workouts/${workout.id}/${path}${query}`,
        {
          method: 'POST',
          ...(body ? { body: JSON.stringify(body) } : {}),
        },
      );
      onPlanUpdated(result.plan, workout.id);
      setFeedbackOpen(false);
      return true;
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : t.updateFailed,
      );
      return false;
    } finally {
      setAction('');
    }
  }

  function feedbackPayload() {
    return {
      completion_status: completionStatus,
      partial_reason: completionStatus === 'partial' ? partialReason : undefined,
      actual_zone: actualZone,
      difficulty,
      fatigue_after: fatigueAfter,
      recovery_after: recoveryAfter,
      repeat_confidence: repeatConfidence,
      satisfaction,
      terrain: terrain || undefined,
      external_conditions: externalConditions || undefined,
      equipment_used: equipmentUsed.trim() || undefined,
      pain_reported: painReported,
      notes,
      distance_km: optionalNumber(distanceKM),
      elevation_gain_m: optionalNumber(elevationGainM),
      average_heart_rate: usesHeartRate ? optionalNumber(averageHeartRate) : undefined,
      average_power_watts: usesPower ? optionalNumber(averagePowerW) : undefined,
      average_cadence_rpm: optionalNumber(averageCadenceRPM),
    };
  }

  async function complete() {
    await mutate('complete', feedbackPayload());
  }

  function openLog() {
    setError('');
    setLogDuration(String(workout.duration_minutes));
    setLogDate(todayKey);
    setLogOpen(true);
  }

  // Marca o treino como feito sem o cronômetro: o atleta informa duração e dia.
  async function logDone() {
    const minutes = Number(logDuration);
    if (!Number.isInteger(minutes) || minutes < 1 || minutes > MAX_LOGGED_DURATION_MINUTES) {
      setError(t.durationRange(MAX_LOGGED_DURATION_MINUTES));
      return;
    }
    const saved = await mutate(
      'log',
      { ...feedbackPayload(), duration_minutes: minutes, performed_on: logDate },
      `?date=${todayKey}`,
    );
    if (saved) setLogOpen(false);
  }

  function closeForm() {
    if (logOpen) setLogOpen(false);
    else setFeedbackOpen(false);
  }

  function openCorrection() {
    setCorrectionDistanceKM(workout.session?.distance_km?.toString() ?? '');
    setCorrectionElevationGainM(workout.session?.elevation_gain_m?.toString() ?? '');
    setCorrectionAverageHeartRate(workout.session?.average_heart_rate?.toString() ?? '');
    setCorrectionAveragePowerW(workout.session?.average_power_watts?.toString() ?? '');
    setCorrectionAverageCadenceRPM(workout.session?.average_cadence_rpm?.toString() ?? '');
    setCorrectionNotice('');
    setError('');
    setCorrectionOpen(true);
  }

  async function correctWorkoutData() {
    setAction('correct');
    setError('');
    setCorrectionNotice('');
    try {
      const result = await apiRequest<{ plan: TrainingPlan }>(
        `/v1/workouts/${workout.id}/correct`,
        {
          method: 'POST',
          body: JSON.stringify({
            distance_km: optionalNumber(correctionDistanceKM),
            elevation_gain_m: optionalNumber(correctionElevationGainM),
            average_heart_rate: correctionHeartRateVisible ? optionalNumber(correctionAverageHeartRate) : undefined,
            average_power_watts: correctionPowerVisible ? optionalNumber(correctionAveragePowerW) : undefined,
            average_cadence_rpm: optionalNumber(correctionAverageCadenceRPM),
          }),
        },
      );
      onPlanUpdated(result.plan, workout.id);
      setCorrectionOpen(false);
      setCorrectionNotice(t.correctionSaved);
    } catch (caught) {
      setError(
        caught instanceof Error
          ? caught.message
          : t.correctionFailed,
      );
    } finally {
      setAction('');
    }
  }

  const session = workout.session;
  const feedback = session?.feedback;
  const busy = Boolean(action);
  const isPastWorkout = todayKey !== '' && workout.scheduled_on < todayKey;
  const canMarkMissed =
    planStatus === 'active' &&
    isPastWorkout &&
    (workout.status === 'planned' || workout.status === 'adapted');
  const isFutureWorkout = isFutureTrainingDate(workout.scheduled_on, todayKey);
  const logRange = loggableDateRange(workout.scheduled_on, todayKey);
  const awaitingStart =
    planStatus === 'active' &&
    (workout.status === 'planned' || workout.status === 'adapted');
  // Só depois de saber a data local (todayKey) é que dá para decidir; um treino
  // futuro não pode ser iniciado (o servidor também recusa).
  const canStart = awaitingStart && todayKey !== '' && !isFutureWorkout;
  const canUndo =
    (planStatus === 'active' || planStatus === 'completed') &&
    (workout.status === 'completed' || workout.status === 'skipped');
  const startLabel = workout.status === 'adapted' ? t.startAdapted : t.start;
  const timeFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { hour: '2-digit', minute: '2-digit' });
  const dataIntegrityIssue =
    workout.status === 'completed' &&
    workout.explanation.data_integrity &&
    workout.explanation.data_integrity.status !== 'valid';
  const correctionHeartRateVisible =
    usesHeartRate || workout.session?.average_heart_rate !== undefined;
  const correctionPowerVisible =
    usesPower || workout.session?.average_power_watts !== undefined;

  return (
    <section className="session-actions" aria-label={t.label}>
      <header>
        <span className={`session-status ${workout.status}`}>
          {t.status[workout.status]}
        </span>
        {session?.started_at && workout.status === 'in_progress' && todayKey !== '' && (
          <small>{t.startedAt(timeFormatter.format(new Date(session.started_at)))}</small>
        )}
      </header>

      {planStatus === 'draft' && (
        <p className="session-guidance">{t.acceptFirst}</p>
      )}

      {awaitingStart && isFutureWorkout && (
        <p className="session-guidance">{t.futureWorkout(formatTrainingDay(workout.scheduled_on, locale))}</p>
      )}

      {canStart && !logOpen && (
        <div className="session-button-row">
          <Button
            type="button"
            className="session-primary"
            disabled={busy}
            onClick={() => mutate('start', undefined, `?date=${todayKey}`)}
          >
            {action === 'start' ? (
              <LoaderCircle className="spin" />
            ) : (
              <Play />
            )}
            {action === 'start' ? t.starting : startLabel}
          </Button>
          <Button type="button" variant="outline" disabled={busy} onClick={openLog}>
            <CheckCircle2 />
            {t.markDone}
          </Button>
          {canMarkMissed && (
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={() => {
                if (
                  window.confirm(t.confirmMissed)
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
              {action === 'missed' ? t.recording : t.missed}
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
            {t.complete}
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={busy}
            onClick={() => {
              if (
                window.confirm(t.confirmCancel)
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
            {t.cancel}
          </Button>
        </div>
      )}

      {(logOpen || (workout.status === 'in_progress' && feedbackOpen)) && (
        <form
          className="session-feedback"
          onSubmit={(event) => {
            event.preventDefault();
            if (logOpen) void logDone();
            else void complete();
          }}
        >
          <div className="feedback-heading">
            <div>
              <strong>{logOpen ? t.logTitle : t.howWasIt}</strong>
              <small>{t.usedForAdaptation}</small>
            </div>
            <button
              type="button"
              onClick={closeForm}
              aria-label={t.backToActions}
            >
              <RotateCcw />
            </button>
          </div>

          {logOpen && (
            <fieldset className="log-fields">
              <legend>{t.howLongWhen}</legend>
              <div className="feedback-grid">
                <label>
                  {t.durationMin}
                  <input
                    type="number"
                    required
                    min="1"
                    max={MAX_LOGGED_DURATION_MINUTES}
                    step="1"
                    inputMode="numeric"
                    value={logDuration}
                    onChange={(event) => setLogDuration(event.target.value)}
                  />
                </label>
                <label>
                  {t.workoutDay}
                  <input
                    type="date"
                    required
                    min={logRange.min}
                    max={logRange.max}
                    value={logDate}
                    onChange={(event) => setLogDate(event.target.value)}
                  />
                </label>
              </div>
              <small className="completion-help">{t.logHelp(workout.duration_minutes)}</small>
            </fieldset>
          )}

          <fieldset className="completion-context">
            <legend>{t.howMuch}</legend>
            <div className="feedback-grid">
              <label>
                {t.completion}
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
                  <option value="complete">{t.fullWorkout}</option>
                  <option value="partial">{t.partOnly}</option>
                </select>
              </label>
              {completionStatus === 'partial' && (
                <label>
                  {t.reason}
                  <select
                    required
                    value={partialReason}
                    onChange={(event) =>
                      setPartialReason(
                        event.target.value as PartialReason,
                      )
                    }
                  >
                    <option value="">{t.selectReason}</option>
                    {Object.entries(t.partialReasons).map(([value, label]) => (
                      <option value={value} key={value}>
                        {label}
                      </option>
                    ))}
                  </select>
                </label>
              )}
            </div>
            {completionStatus === 'partial' && (
              <small className="completion-help">{t.partialHelp}</small>
            )}
          </fieldset>

          <fieldset className="zone-choice">
            <legend>{t.zoneQuestion}</legend>
            <small>{t.zoneHelp}</small>
            <div className="zone-cards">
              {ZONES.map((zone) => {
                const ranges = zoneRanges(zone, zoneReference);
                return (
                  <label
                    key={zone.number}
                    className={[
                      'zone-card',
                      actualZone === zone.number ? 'selected' : '',
                      plannedZone === zone.number ? 'planned' : '',
                    ].filter(Boolean).join(' ')}
                  >
                    <input
                      type="radio"
                      name={`zone-${workout.id}`}
                      checked={actualZone === zone.number}
                      onChange={() => setActualZone(zone.number)}
                    />
                    <strong>
                      Z{zone.number} · {zoneName(zone, locale)}
                      {ranges.heartRate && <small>{formatHeartRateRange(ranges.heartRate, locale)}</small>}
                      {ranges.power && <small>{formatPowerRange(ranges.power, locale)}</small>}
                      {plannedZone === zone.number && <em className="zone-planned">{t.planned}</em>}
                    </strong>
                    <span>{zoneTalk(zone, locale)}</span>
                  </label>
                );
              })}
            </div>
          </fieldset>

          <div className="feedback-grid">
            <label>
              {t.difficultyLabel}
              <select
                value={difficulty}
                onChange={(event) => setDifficulty(event.target.value as Difficulty)}
              >
                {Object.entries(t.difficulty).map(([value, label]) => (
                  <option value={value} key={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              {t.fatigueAfter}
              <select
                value={fatigueAfter}
                onChange={(event) =>
                  setFatigueAfter(Number(event.target.value))
                }
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {t.outOf5(value)}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="feedback-grid">
            <label>
              {t.satisfaction}
              <select
                value={satisfaction}
                onChange={(event) => setSatisfaction(Number(event.target.value))}
              >
                {t.satisfactionOptions.map((label, index) => (
                  <option value={index + 1} key={label}>
                    {t.outOf5(index + 1)} · {label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              {t.terrainLabel}
              <select
                value={terrain}
                onChange={(event) => setTerrain(event.target.value as Terrain | '')}
              >
                <option value="">{t.notInformed}</option>
                {Object.entries(t.terrain).map(([value, label]) => (
                  <option value={value} key={value}>{label}</option>
                ))}
              </select>
            </label>
          </div>

          <label>
            {t.conditionsLabel}
            <select
              value={externalConditions}
              onChange={(event) => setExternalConditions(event.target.value as ExternalCondition | '')}
            >
              <option value="">{t.notInformed}</option>
              {Object.entries(t.conditions).map(([value, label]) => (
                <option value={value} key={value}>{label}</option>
              ))}
            </select>
          </label>

          <label>
            {t.equipment} <small>{t.optional}</small>
            <input
              type="text"
              maxLength={120}
              value={equipmentUsed}
              onChange={(event) => setEquipmentUsed(event.target.value)}
              placeholder={t.equipmentPlaceholder}
            />
          </label>

          <div className="feedback-grid">
            <label>
              {t.recoveryAfter}
              <select
                value={recoveryAfter}
                onChange={(event) => setRecoveryAfter(Number(event.target.value))}
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {t.outOf5(value)}
                  </option>
                ))}
              </select>
            </label>
            <label>
              {t.repeatConfidence}
              <select
                value={repeatConfidence}
                onChange={(event) => setRepeatConfidence(Number(event.target.value))}
              >
                {[1, 2, 3, 4, 5].map((value) => (
                  <option value={value} key={value}>
                    {t.outOf5(value)}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <fieldset className="session-metrics">
            <legend>{t.rideData} <small>{t.optionals}</small></legend>
            <div className="feedback-grid">
              <label>
                {t.distance}
                <input type="number" min="0" max="2000" step="0.1" inputMode="decimal" value={distanceKM} onChange={(event) => setDistanceKM(event.target.value)} placeholder={t.example(formatDecimal(32.5, 1, locale))} />
              </label>
              <label>
                {t.elevation}
                <input type="number" min="0" max="20000" step="1" inputMode="numeric" value={elevationGainM} onChange={(event) => setElevationGainM(event.target.value)} placeholder={t.example('420')} />
              </label>
              {usesHeartRate && <label>
                {t.averageHr}
                <input type="number" min="30" max="250" step="1" inputMode="numeric" value={averageHeartRate} onChange={(event) => setAverageHeartRate(event.target.value)} placeholder={t.example('142')} />
              </label>}
              {usesPower && <label>
                {t.averagePower}
                <input type="number" min="0" max="2000" step="1" inputMode="numeric" value={averagePowerW} onChange={(event) => setAveragePowerW(event.target.value)} placeholder={t.example('185')} />
              </label>}
              <label>
                {t.averageCadence}
                <input type="number" min="1" max="300" step="1" inputMode="numeric" value={averageCadenceRPM} onChange={(event) => setAverageCadenceRPM(event.target.value)} placeholder={t.example('88')} />
              </label>
            </div>
          </fieldset>

          <label className="pain-check">
            <input
              type="checkbox"
              checked={painReported}
              onChange={(event) => setPainReported(event.target.checked)}
            />
            {t.pain}
          </label>
          {painReported && (
            <p className="pain-warning">
              <TriangleAlert />
              {t.painWarning}
            </p>
          )}

          <label htmlFor={`notes-${workout.id}`}>{t.notes}</label>
          <Textarea
            id={`notes-${workout.id}`}
            maxLength={1000}
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            placeholder={t.notesPlaceholder}
          />

          <Button type="submit" className="session-primary" disabled={busy}>
            {action === (logOpen ? 'log' : 'complete') ? (
              <LoaderCircle className="spin" />
            ) : (
              <CheckCircle2 />
            )}
            {action === (logOpen ? 'log' : 'complete') ? t.saving : logOpen ? t.saveLog : t.saveComplete}
          </Button>
        </form>
      )}

      {workout.status === 'completed' && feedback && (
        <div className="session-result">
          <CheckCircle2 />
          <div>
            <strong>{t.completed} · {session?.actual_rpe !== undefined ? zoneLabel(session.actual_rpe, locale) : t.noZone}</strong>
            {session?.duration_minutes !== undefined && (
              <small className="duration-source">
                {t.durationSource(session.duration_minutes, durationSourceLabel(session.duration_source, locale))}
              </small>
            )}
            <span>
              {feedback.completion_status === 'partial'
                ? t.partialSummary(feedback.partial_reason ? t.partialReasons[feedback.partial_reason] : t.noReason)
                : t.fullWorkout}{' '}
              ·{' '}
              {t.difficulty[feedback.difficulty]} · {t.fatigue}{' '}
              {feedback.fatigue_after}/5
              {feedback.recovery_after !== undefined && ` · ${t.recovery} ${feedback.recovery_after}/5`}
              {feedback.repeat_confidence !== undefined && ` · ${t.confidence} ${feedback.repeat_confidence}/5`}
              {feedback.satisfaction !== undefined && ` · ${t.satisfactionShort} ${feedback.satisfaction}/5`}
              {feedback.terrain && ` · ${t.terrain[feedback.terrain]}`}
              {feedback.external_conditions && ` · ${t.conditions[feedback.external_conditions]}`}
              {feedback.equipment_used && ` · ${t.equipmentShort}: ${feedback.equipment_used}`}
              {feedback.pain_reported ? ` · ${t.painReported}` : ` · ${t.noPain}`}
            </span>
            {feedback.notes && <p>{feedback.notes}</p>}
            {(session?.distance_km !== undefined || session?.elevation_gain_m !== undefined || session?.average_heart_rate !== undefined || session?.average_power_watts !== undefined || session?.average_cadence_rpm !== undefined) && <p className="session-metric-result">{session?.distance_km !== undefined && `${session.distance_km} km`}{session?.elevation_gain_m !== undefined && ` · ${session.elevation_gain_m} m+`}{session?.average_heart_rate !== undefined && ` · ${t.hr} ${session.average_heart_rate} bpm`}{session?.average_power_watts !== undefined && ` · ${session.average_power_watts} W`}{session?.average_cadence_rpm !== undefined && ` · ${session.average_cadence_rpm} rpm`}</p>}
          </div>
        </div>
      )}

      {workout.status === 'completed' && workout.explanation.data_integrity && workout.explanation.data_integrity.status !== 'valid' && (
        <div className="session-data-warning" aria-live="polite">
          <TriangleAlert />
          <div>
            <strong>{t.reviewTitle}</strong>
            <p>{workout.explanation.data_integrity.status === 'inconsistent' ? t.inconsistent : t.incomplete}</p>
          </div>
        </div>
      )}

      {correctionNotice && (
        <p className="session-correction-success" aria-live="polite">
          {correctionNotice}
        </p>
      )}

      {dataIntegrityIssue && !correctionOpen && (
        <Button type="button" variant="outline" disabled={busy} onClick={openCorrection}>
          {t.correct}
        </Button>
      )}

      {dataIntegrityIssue && correctionOpen && (
        <form
          className="session-feedback session-correction-form"
          onSubmit={(event) => {
            event.preventDefault();
            void correctWorkoutData();
          }}
        >
          <div className="feedback-heading">
            <div>
              <strong>{t.correct}</strong>
              <small>{t.correctHelp}</small>
            </div>
            <button type="button" onClick={() => setCorrectionOpen(false)} aria-label={t.cancelCorrection}>
              <RotateCcw />
            </button>
          </div>
          <div className="feedback-grid">
            <label>
              {t.distance}
              <input type="number" min="0" max="2000" step="0.1" inputMode="decimal" value={correctionDistanceKM} onChange={(event) => setCorrectionDistanceKM(event.target.value)} />
            </label>
            <label>
              {t.elevation}
              <input type="number" min="0" max="20000" step="1" inputMode="numeric" value={correctionElevationGainM} onChange={(event) => setCorrectionElevationGainM(event.target.value)} />
            </label>
            {correctionHeartRateVisible && <label>
              {t.averageHr}
              <input type="number" min="30" max="250" step="1" inputMode="numeric" value={correctionAverageHeartRate} onChange={(event) => setCorrectionAverageHeartRate(event.target.value)} />
            </label>}
            {correctionPowerVisible && <label>
              {t.averagePower}
              <input type="number" min="0" max="2000" step="1" inputMode="numeric" value={correctionAveragePowerW} onChange={(event) => setCorrectionAveragePowerW(event.target.value)} />
            </label>}
            <label>
              {t.averageCadence}
              <input type="number" min="1" max="300" step="1" inputMode="numeric" value={correctionAverageCadenceRPM} onChange={(event) => setCorrectionAverageCadenceRPM(event.target.value)} />
            </label>
          </div>
          <Button type="submit" className="session-primary" disabled={busy}>
            {action === 'correct' ? <LoaderCircle className="spin" /> : <CheckCircle2 />}
            {action === 'correct' ? t.savingCorrection : t.saveCorrection}
          </Button>
        </form>
      )}

      {workout.status === 'skipped' && (
        <p className="session-guidance">{t.skipped}</p>
      )}

      {canUndo && (
        <Button
          type="button"
          variant="outline"
          disabled={busy}
          onClick={() => {
            if (window.confirm(undoConfirmation(workout.status, locale))) {
              void mutate('undo');
            }
          }}
        >
          {action === 'undo' ? <LoaderCircle className="spin" /> : <RotateCcw />}
          {action === 'undo' ? t.undoing : workout.status === 'completed' ? t.undo : t.reopen}
        </Button>
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
