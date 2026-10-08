'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Activity, CalendarDays, Clock3, Gauge, HeartPulse, LoaderCircle, MapPinned, UploadCloud, XCircle, Zap } from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { defineMessages, INTL_LOCALE } from '@/lib/i18n';
import { parseTrainingDate, type Activity as TrainingActivity } from '@/lib/planning';
import { zoneLabel } from '@/lib/zones';

type User = { display_name: string };
type Feedback = NonNullable<TrainingActivity['feedback']>;

const messages = defineMessages({
  pt: {
    difficulty: { very_easy: 'Muito fácil', easy: 'Fácil', moderate: 'Moderado', hard: 'Difícil', very_hard: 'Muito difícil' } as Record<Feedback['difficulty'], string>,
    partialReasons: {
      time_available_changed: 'sem tempo',
      fatigue_or_recovery: 'fadiga/recuperação',
      pain_or_discomfort: 'dor/desconforto',
      equipment_or_conditions: 'equipamento/clima/terreno',
      other: 'outro motivo',
    } as Record<NonNullable<Feedback['partial_reason']>, string>,
    terrain: { flat: 'plano', rolling: 'ondulado', hilly: 'montanhoso', mixed: 'misto', technical: 'técnico', indoor: 'indoor' } as Record<NonNullable<Feedback['terrain']>, string>,
    conditions: {
      normal: 'condições normais', heat: 'calor', cold: 'frio', wind: 'vento', rain: 'chuva', poor_visibility: 'baixa visibilidade', other: 'outra condição',
    } as Record<NonNullable<Feedback['external_conditions']>, string>,
    loadFailed: 'Não foi possível carregar suas atividades.',
    loading: 'Carregando suas atividades…',
    kicker: 'HISTÓRICO DE TREINOS',
    title: 'Suas atividades.',
    intro: 'Concluídas e canceladas, da mais recente para a mais antiga.',
    import: 'Importar de um arquivo .fit ou .gpx',
    emptyTitle: 'Ainda não há atividades registradas.',
    emptyText: 'Quando você concluir ou cancelar uma sessão, ela aparecerá aqui.',
    seePlan: 'Ver meu plano',
    completed: 'Concluída',
    cancelled: 'Cancelada',
    duration: 'duração',
    partial: 'Parcial',
    noReason: 'motivo não informado',
    complete: 'Completa',
    fatigue: 'Fadiga',
    recovery: 'Recuperação',
    confidence: 'Confiança',
    satisfaction: 'Satisfação',
    terrainLabel: 'Terreno',
    conditionsLabel: 'Condições',
    equipment: 'Equipamento',
    pain: 'Dor relatada',
    noPain: 'Sem dor',
  },
  en: {
    difficulty: { very_easy: 'Very easy', easy: 'Easy', moderate: 'Moderate', hard: 'Hard', very_hard: 'Very hard' },
    partialReasons: {
      time_available_changed: 'no time',
      fatigue_or_recovery: 'fatigue/recovery',
      pain_or_discomfort: 'pain/discomfort',
      equipment_or_conditions: 'equipment/weather/terrain',
      other: 'other reason',
    },
    terrain: { flat: 'flat', rolling: 'rolling', hilly: 'hilly', mixed: 'mixed', technical: 'technical', indoor: 'indoor' },
    conditions: {
      normal: 'normal conditions', heat: 'heat', cold: 'cold', wind: 'wind', rain: 'rain', poor_visibility: 'poor visibility', other: 'other condition',
    },
    loadFailed: 'Your activities could not be loaded.',
    loading: 'Loading your activities…',
    kicker: 'TRAINING HISTORY',
    title: 'Your activities.',
    intro: 'Completed and cancelled, newest first.',
    import: 'Import from a .fit or .gpx file',
    emptyTitle: 'No activities recorded yet.',
    emptyText: 'When you complete or cancel a session, it will show up here.',
    seePlan: 'See my plan',
    completed: 'Completed',
    cancelled: 'Cancelled',
    duration: 'duration',
    partial: 'Partial',
    noReason: 'no reason given',
    complete: 'Full',
    fatigue: 'Fatigue',
    recovery: 'Recovery',
    confidence: 'Confidence',
    satisfaction: 'Satisfaction',
    terrainLabel: 'Terrain',
    conditionsLabel: 'Conditions',
    equipment: 'Equipment',
    pain: 'Pain reported',
    noPain: 'No pain',
  },
});

export default function ActivitiesPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const [user, setUser] = useState<User | null>(null);
  const [activities, setActivities] = useState<TrainingActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ activities: TrainingActivity[] }>('/v1/activities'),
    ])
      .then(([account, result]) => {
        setUser(account.user);
        setActivities(result.activities);
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

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />{t.loading}</main>;
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  const dateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], {
    day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
  });

  return (
    <main className="activities-shell">
      <AppHeader name={user?.display_name} />
      <section className="activities-content">
        <header className="activities-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>{t.intro}</span>
          <Link href="/atividades/importar" className="back-link">
            <UploadCloud size={15} />{t.import}
          </Link>
        </header>
        {error && <p className="form-error" role="alert">{error}</p>}
        {!error && activities.length === 0 && (
          <section className="activities-empty">
            <Activity size={25} />
            <h2>{t.emptyTitle}</h2>
            <p>{t.emptyText}</p>
            <Link href="/plano">{t.seePlan}</Link>
          </section>
        )}
        <div className="activities-list">
          {activities.map((item) => {
            const completed = item.status === 'completed';
            const terminalAt = item.completed_at || item.cancelled_at;
            return (
              <article className={`activity-card ${item.status}`} key={item.id}>
                <div className="activity-icon">{completed ? <Activity size={18} /> : <XCircle size={18} />}</div>
                <div className="activity-main">
                  <div className="activity-title"><div><h2>{item.name}</h2><p>{item.objective}</p></div><span className="activity-status">{completed ? t.completed : t.cancelled}</span></div>
                  <time><CalendarDays size={14} />{terminalAt ? dateFormatter.format(new Date(terminalAt)) : parseTrainingDate(item.scheduled_on).toLocaleDateString(INTL_LOCALE[locale])}</time>
                  <div className="activity-metrics">
                    <span><Clock3 size={14} /><b>{item.duration_minutes ?? '—'}</b> {item.duration_minutes === undefined ? t.duration : 'min'}</span>
                    <span><Gauge size={14} /><b>{item.actual_rpe ? zoneLabel(item.actual_rpe, locale) : '—'}</b></span>
                    {item.distance_km !== undefined && <span><MapPinned size={14} /><b>{item.distance_km} km</b></span>}
                    {item.elevation_gain_m !== undefined && <span><MapPinned size={14} /><b>{item.elevation_gain_m} m+</b></span>}
                    {item.average_heart_rate !== undefined && <span><HeartPulse size={14} /><b>{item.average_heart_rate} bpm</b></span>}
                    {item.average_power_watts !== undefined && <span><Zap size={14} /><b>{item.average_power_watts} W</b></span>}
                    {item.average_cadence_rpm !== undefined && <span><Gauge size={14} /><b>{item.average_cadence_rpm} rpm</b></span>}
                    {item.feedback && <><span><b>{item.feedback.completion_status === 'partial' ? `${t.partial} · ${item.feedback.partial_reason ? t.partialReasons[item.feedback.partial_reason] : t.noReason}` : t.complete}</b></span><span><b>{t.difficulty[item.feedback.difficulty]}</b></span><span>{t.fatigue} <b>{item.feedback.fatigue_after}/5</b></span>{item.feedback.recovery_after !== undefined && <span>{t.recovery} <b>{item.feedback.recovery_after}/5</b></span>}{item.feedback.repeat_confidence !== undefined && <span>{t.confidence} <b>{item.feedback.repeat_confidence}/5</b></span>}{item.feedback.satisfaction !== undefined && <span>{t.satisfaction} <b>{item.feedback.satisfaction}/5</b></span>}{item.feedback.terrain && <span>{t.terrainLabel} <b>{t.terrain[item.feedback.terrain]}</b></span>}{item.feedback.external_conditions && <span>{t.conditionsLabel} <b>{t.conditions[item.feedback.external_conditions]}</b></span>}{item.feedback.equipment_used && <span>{t.equipment} <b>{item.feedback.equipment_used}</b></span>}<span className={item.feedback.pain_reported ? 'pain' : ''}>{item.feedback.pain_reported ? t.pain : t.noPain}</span></>}
                  </div>
                  {item.feedback?.notes && <p className="activity-notes">{item.feedback.notes}</p>}
                </div>
              </article>
            );
          })}
        </div>
      </section>
    </main>
  );
}
