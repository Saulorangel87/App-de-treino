'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  CalendarDays,
  Clock3,
  Gauge,
  HeartPulse,
  Link2,
  Link2Off,
  LoaderCircle,
  MapPinned,
  Trash2,
  UploadCloud,
  Zap,
} from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { currentLocale, defineMessages, formatDecimal, INTL_LOCALE } from '@/lib/i18n';
import { parseTrainingDate } from '@/lib/planning';

type User = { display_name: string };

type WorkoutCandidate = {
  id: string;
  scheduled_on: string;
  name: string;
  status: string;
};

type ImportedActivity = {
  id: string;
  source: 'fit' | 'gpx';
  started_at: string;
  moving_seconds: number;
  distance_km: number;
  elevation_gain_m?: number;
  average_heart_rate?: number;
  max_heart_rate?: number;
  average_power_watts?: number;
  normalized_power_watts?: number;
  average_cadence_rpm?: number;
  workout_id?: string;
  workout_name?: string;
  workout_scheduled_on?: string;
  imported_at: string;
};

const messages = defineMessages({
  pt: {
    loadFailed: 'Não foi possível carregar suas atividades importadas.',
    importFailed: 'Não foi possível importar este arquivo.',
    linkFailed: 'Não foi possível atualizar o vínculo.',
    candidatesFailed: 'Não foi possível buscar treinos para esta atividade.',
    deleteFailed: 'Não foi possível remover esta atividade.',
    pageFailed: 'Não foi possível carregar esta página.',
    loading: 'Carregando…',
    back: 'Voltar às atividades',
    kicker: 'DADOS REAIS DE EXECUÇÃO',
    title: 'Importar uma atividade.',
    introBefore: 'Envie o arquivo ',
    introOr: ' ou ',
    introAfter:
      ' exportado do seu relógio ou ciclocomputador (Garmin, Wahoo, XOSS e outros). Isso não muda seu plano nem a prescrição de nenhum treino — os dados ficam disponíveis aqui para você conferir e, se quiser, usar ao concluir o treino do dia correspondente em ',
    yourPlan: 'seu plano',
    fileLabel: 'Arquivo da atividade',
    fileHint: 'Tamanho máximo de 20 MB. O arquivo não é guardado depois de lido; só o resumo da atividade é salvo.',
    reading: 'Lendo o arquivo…',
    imported: 'Atividade importada.',
    found: (count: number) =>
      `Encontramos ${count === 1 ? 'um treino planejado' : 'treinos planejados'} no mesmo dia. Vincule para guardar a relação. Se abrir no plano, os dados acima já aparecem preenchidos no formulário, e você decide se confirma.`,
    linked: 'Vinculada',
    link: 'Vincular',
    openInPlan: 'Abrir no plano',
    noneFound: 'Não encontramos um treino planejado nesse dia. Os dados ficam guardados aqui, sem vínculo.',
    alreadyImported: 'Atividades já importadas',
    emptyTitle: 'Nenhuma atividade importada ainda.',
    emptyText: 'Envie um arquivo .fit ou .gpx acima para começar.',
    file: (source: string) => `Arquivo .${source}`,
    linkedTo: (name: string) => `Vinculada a ${name}`,
    aWorkout: 'um treino',
    notLinked: 'Sem treino vinculado',
    remove: 'Remover atividade importada',
    unlink: 'Desvincular',
    linkToWorkout: 'Vincular a um treino',
    searching: 'Buscando treinos perto dessa data…',
    noneNearby: 'Nenhum treino planejado perto dessa data. Você pode gerar ou ajustar o plano e voltar aqui.',
  },
  en: {
    loadFailed: 'Your imported activities could not be loaded.',
    importFailed: 'This file could not be imported.',
    linkFailed: 'The link could not be updated.',
    candidatesFailed: 'Workouts for this activity could not be found.',
    deleteFailed: 'This activity could not be removed.',
    pageFailed: 'This page could not be loaded.',
    loading: 'Loading…',
    back: 'Back to activities',
    kicker: 'REAL RIDE DATA',
    title: 'Import an activity.',
    introBefore: 'Upload the ',
    introOr: ' or ',
    introAfter:
      ' file exported from your watch or bike computer (Garmin, Wahoo, XOSS and others). This does not change your plan or the prescription of any workout — the data stays here for you to check and, if you want, use when completing the matching day\'s workout in ',
    yourPlan: 'your plan',
    fileLabel: 'Activity file',
    fileHint: 'Maximum size 20 MB. The file is not kept after it is read; only the activity summary is saved.',
    reading: 'Reading the file…',
    imported: 'Activity imported.',
    found: (count: number) =>
      `We found ${count === 1 ? 'a planned workout' : 'planned workouts'} on the same day. Link it to keep the relationship. If you open it in the plan, the data above is already filled in on the form, and you decide whether to confirm.`,
    linked: 'Linked',
    link: 'Link',
    openInPlan: 'Open in plan',
    noneFound: "We didn't find a planned workout on that day. The data stays here, unlinked.",
    alreadyImported: 'Imported activities',
    emptyTitle: 'No imported activities yet.',
    emptyText: 'Upload a .fit or .gpx file above to get started.',
    file: (source: string) => `.${source} file`,
    linkedTo: (name: string) => `Linked to ${name}`,
    aWorkout: 'a workout',
    notLinked: 'No linked workout',
    remove: 'Remove imported activity',
    unlink: 'Unlink',
    linkToWorkout: 'Link to a workout',
    searching: 'Looking for workouts near that date…',
    noneNearby: 'No planned workouts near that date. You can generate or adjust the plan and come back here.',
  },
});

function planLinkFor(workoutID: string, activity: ImportedActivity): string {
  const params = new URLSearchParams({ workoutID });
  if (activity.distance_km > 0) params.set('distance_km', String(activity.distance_km));
  if (activity.elevation_gain_m !== undefined) params.set('elevation_gain_m', String(activity.elevation_gain_m));
  if (activity.average_heart_rate !== undefined) params.set('average_heart_rate', String(activity.average_heart_rate));
  if (activity.average_power_watts !== undefined) params.set('average_power_watts', String(activity.average_power_watts));
  if (activity.average_cadence_rpm !== undefined) params.set('average_cadence_rpm', String(activity.average_cadence_rpm));
  return `/plano?${params.toString()}`;
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.round((seconds % 3600) / 60);
  return hours > 0 ? `${hours}h${minutes.toString().padStart(2, '0')}` : `${minutes} min`;
}

export default function ImportActivityPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const [user, setUser] = useState<User | null>(null);
  const [activities, setActivities] = useState<ImportedActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState('');
  const [lastImported, setLastImported] = useState<ImportedActivity | null>(null);
  const [candidates, setCandidates] = useState<WorkoutCandidate[]>([]);
  // Vínculo de atividades já importadas: qual está com o seletor aberto, as
  // sugestões carregadas e qual pedido está em andamento.
  const [pickerFor, setPickerFor] = useState<string | null>(null);
  const [pickerCandidates, setPickerCandidates] = useState<WorkoutCandidate[] | null>(null);
  const [busyActivity, setBusyActivity] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ activities: ImportedActivity[] }>('/v1/activities/imported'),
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
        setLoadError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

  // Valores espelhados de public/sw.js (SHARE_CACHE_NAME / SHARED_FILE_URL):
  // o service worker intercepta o POST do menu "Compartilhar" do Android,
  // guarda o arquivo aqui e redireciona para esta página com ?compartilhado=1.
  useEffect(() => {
    if (new URLSearchParams(window.location.search).get('compartilhado') !== '1') return;
    if (!('caches' in window)) return;
    caches
      .open('cadencia-shared-file')
      .then((cache) => cache.match('/__shared-activity').then((response) => ({ cache, response })))
      .then(async ({ cache, response }) => {
        if (!response) return;
        const blob = await response.blob();
        const filename = decodeURIComponent(response.headers.get('X-Shared-Filename') || 'atividade.fit');
        await cache.delete('/__shared-activity');
        await handleFileChosen(new File([blob], filename, { type: blob.type }));
      })
      .catch(() => undefined);
  }, []);

  async function handleFileChosen(file: File) {
    setUploadError('');
    setLastImported(null);
    setCandidates([]);
    setUploading(true);
    try {
      const form = new FormData();
      form.append('file', file, file.name);
      const result = await apiRequest<{ activity: ImportedActivity; candidate_workouts: WorkoutCandidate[] }>(
        '/v1/activities/import',
        { method: 'POST', body: form },
      );
      setLastImported(result.activity);
      setCandidates(result.candidate_workouts || []);
      setActivities((current) => [result.activity, ...current]);
    } catch (caught) {
      setUploadError(apiErrorMessage(caught, messages[currentLocale()].importFailed));
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }

  // Grava (ou remove, com workoutID nulo) só a associação entre a atividade e o
  // treino; nada é copiado para o treino nem muda no plano.
  async function linkActivity(activityID: string, workoutID: string | null) {
    setUploadError('');
    setBusyActivity(activityID);
    try {
      const result = await apiRequest<{ activity: ImportedActivity }>(
        `/v1/activities/imported/${activityID}/workout`,
        { method: 'PUT', body: JSON.stringify({ workout_id: workoutID }) },
      );
      setActivities((current) => current.map((item) => (item.id === activityID ? result.activity : item)));
      setPickerFor(null);
      setPickerCandidates(null);
    } catch (caught) {
      setUploadError(apiErrorMessage(caught, t.linkFailed));
    } finally {
      setBusyActivity(null);
    }
  }

  async function togglePicker(activityID: string) {
    if (pickerFor === activityID) {
      setPickerFor(null);
      setPickerCandidates(null);
      return;
    }
    setUploadError('');
    setPickerFor(activityID);
    setPickerCandidates(null);
    try {
      const result = await apiRequest<{ candidate_workouts: WorkoutCandidate[] }>(
        `/v1/activities/imported/${activityID}/candidates`,
      );
      setPickerCandidates(result.candidate_workouts || []);
    } catch (caught) {
      setPickerFor(null);
      setUploadError(apiErrorMessage(caught, t.candidatesFailed));
    }
  }

  async function handleDelete(activityID: string) {
    try {
      await apiRequest(`/v1/activities/imported/${activityID}`, { method: 'DELETE' });
      setActivities((current) => current.filter((item) => item.id !== activityID));
    } catch (caught) {
      setUploadError(apiErrorMessage(caught, t.deleteFailed));
    }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />{t.loading}</main>;
  if (!user) return <ApiErrorState message={loadError || t.pageFailed} />;

  const dateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], {
    day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
  });
  const shortDateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { weekday: 'short', day: '2-digit', month: 'short' });
  const shortDate = (value: string) => shortDateFormatter.format(parseTrainingDate(value)).replace(/\./g, '');

  return (
    <main className="activities-shell">
      <AppHeader name={user?.display_name} />
      <section className="activities-content">
        <Link href="/atividades" className="back-link"><ArrowLeft size={15} />{t.back}</Link>
        <header className="activities-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>
            {t.introBefore}<b>.fit</b>{t.introOr}<b>.gpx</b>{t.introAfter}
            <Link href="/plano">{t.yourPlan}</Link>.
          </span>
        </header>

        <form
          className="import-form"
          onSubmit={(event) => event.preventDefault()}
        >
          <label htmlFor="activity-file">{t.fileLabel}</label>
          <input
            id="activity-file"
            ref={fileInputRef}
            type="file"
            accept=".fit,.gpx"
            disabled={uploading}
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (file) void handleFileChosen(file);
            }}
          />
          <small>{t.fileHint}</small>
          {uploading && (
            <p className="form-notice" role="status">
              <LoaderCircle className="spin" size={14} /> {t.reading}
            </p>
          )}
          {uploadError && <p className="form-error" role="alert">{uploadError}</p>}
        </form>

        {lastImported && (
          <section className="import-result" aria-live="polite">
            <h2>{t.imported}</h2>
            <p>
              {dateFormatter.format(new Date(lastImported.started_at))} · {formatDuration(lastImported.moving_seconds)}
              {lastImported.distance_km > 0 && ` · ${formatDecimal(lastImported.distance_km, 1, locale)} km`}
            </p>
            {candidates.length > 0 ? (
              <>
                <p>{t.found(candidates.length)}</p>
                <ul className="link-candidates">
                  {candidates.map((candidate) => {
                    const current = activities.find((item) => item.id === lastImported.id);
                    const linkedHere = current?.workout_id === candidate.id;
                    return (
                      <li key={candidate.id}>
                        <span>
                          <strong>{candidate.name}</strong>
                          <small>{shortDate(candidate.scheduled_on)}</small>
                        </span>
                        <span className="link-actions">
                          {linkedHere ? (
                            <span className="link-done"><Link2 size={14} aria-hidden="true" />{t.linked}</span>
                          ) : (
                            <button
                              type="button"
                              className="btn btn-outline btn-sm"
                              disabled={busyActivity === lastImported.id}
                              onClick={() => void linkActivity(lastImported.id, candidate.id)}
                            >
                              {busyActivity === lastImported.id ? <LoaderCircle className="spin" size={14} /> : <Link2 size={14} />}
                              {t.link}
                            </button>
                          )}
                          <Link href={planLinkFor(candidate.id, lastImported)}>{t.openInPlan}</Link>
                        </span>
                      </li>
                    );
                  })}
                </ul>
              </>
            ) : (
              <p>{t.noneFound}</p>
            )}
          </section>
        )}

        <h2 className="activities-subheading">{t.alreadyImported}</h2>
        {activities.length === 0 && (
          <section className="activities-empty">
            <UploadCloud size={25} />
            <h2>{t.emptyTitle}</h2>
            <p>{t.emptyText}</p>
          </section>
        )}
        <div className="activities-list">
          {activities.map((item) => (
            <article className="activity-card completed" key={item.id}>
              <div className="activity-icon"><UploadCloud size={18} /></div>
              <div className="activity-main">
                <div className="activity-title">
                  <div>
                    <h2>{t.file(item.source)}</h2>
                    <p>
                      {item.workout_id
                        ? `${t.linkedTo(item.workout_name ?? t.aWorkout)}${item.workout_scheduled_on ? ` · ${shortDate(item.workout_scheduled_on)}` : ''}`
                        : t.notLinked}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="icon-button"
                    onClick={() => void handleDelete(item.id)}
                    aria-label={t.remove}
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
                <time><CalendarDays size={14} />{dateFormatter.format(new Date(item.started_at))}</time>
                <div className="activity-link-row">
                  {item.workout_id ? (
                    <>
                      <Link href={planLinkFor(item.workout_id, item)}>{t.openInPlan}</Link>
                      <button
                        type="button"
                        className="btn btn-outline btn-sm"
                        disabled={busyActivity === item.id}
                        onClick={() => void linkActivity(item.id, null)}
                      >
                        {busyActivity === item.id ? <LoaderCircle className="spin" size={14} /> : <Link2Off size={14} />}
                        {t.unlink}
                      </button>
                    </>
                  ) : (
                    <button
                      type="button"
                      className="btn btn-outline btn-sm"
                      aria-expanded={pickerFor === item.id}
                      onClick={() => void togglePicker(item.id)}
                    >
                      <Link2 size={14} />
                      {t.linkToWorkout}
                    </button>
                  )}
                </div>
                {pickerFor === item.id && (
                  <div className="link-picker" aria-live="polite">
                    {pickerCandidates === null ? (
                      <p><LoaderCircle className="spin" size={14} /> {t.searching}</p>
                    ) : pickerCandidates.length === 0 ? (
                      <p>{t.noneNearby}</p>
                    ) : (
                      <ul className="link-candidates">
                        {pickerCandidates.map((candidate) => (
                          <li key={candidate.id}>
                            <span>
                              <strong>{candidate.name}</strong>
                              <small>{shortDate(candidate.scheduled_on)}</small>
                            </span>
                            <button
                              type="button"
                              className="btn btn-primary btn-sm"
                              disabled={busyActivity === item.id}
                              onClick={() => void linkActivity(item.id, candidate.id)}
                            >
                              {t.link}
                            </button>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )}
                <div className="activity-metrics">
                  <span><Clock3 size={14} /><b>{formatDuration(item.moving_seconds)}</b></span>
                  {item.distance_km > 0 && <span><MapPinned size={14} /><b>{formatDecimal(item.distance_km, 1, locale)} km</b></span>}
                  {item.elevation_gain_m !== undefined && <span><MapPinned size={14} /><b>{item.elevation_gain_m} m+</b></span>}
                  {item.average_heart_rate !== undefined && <span><HeartPulse size={14} /><b>{item.average_heart_rate} bpm</b></span>}
                  {item.average_power_watts !== undefined && <span><Zap size={14} /><b>{item.average_power_watts} W</b></span>}
                  {item.average_cadence_rpm !== undefined && <span><Gauge size={14} /><b>{item.average_cadence_rpm} rpm</b></span>}
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
