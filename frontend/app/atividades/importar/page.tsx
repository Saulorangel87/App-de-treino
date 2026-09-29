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

const dateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
});

const shortDateFormatter = new Intl.DateTimeFormat('pt-BR', { weekday: 'short', day: '2-digit', month: 'short' });

function shortDate(value: string): string {
  return shortDateFormatter.format(parseTrainingDate(value)).replace(/\./g, '');
}

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
        setLoadError(apiErrorMessage(caught, 'Não foi possível carregar suas atividades importadas.'));
      })
      .finally(() => setLoading(false));
  }, []);

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
      setUploadError(apiErrorMessage(caught, 'Não foi possível importar este arquivo.'));
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
      setUploadError(apiErrorMessage(caught, 'Não foi possível atualizar o vínculo.'));
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
      setUploadError(apiErrorMessage(caught, 'Não foi possível buscar treinos para esta atividade.'));
    }
  }

  async function handleDelete(activityID: string) {
    try {
      await apiRequest(`/v1/activities/imported/${activityID}`, { method: 'DELETE' });
      setActivities((current) => current.filter((item) => item.id !== activityID));
    } catch (caught) {
      setUploadError(apiErrorMessage(caught, 'Não foi possível remover esta atividade.'));
    }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando…</main>;
  if (!user) return <ApiErrorState message={loadError || 'Não foi possível carregar esta página.'} />;

  return (
    <main className="activities-shell">
      <AppHeader name={user?.display_name} />
      <section className="activities-content">
        <Link href="/atividades" className="back-link"><ArrowLeft size={15} />Voltar às atividades</Link>
        <header className="activities-heading">
          <p>DADOS REAIS DE EXECUÇÃO</p>
          <h1>Importar uma atividade.</h1>
          <span>
            Envie o arquivo <b>.fit</b> ou <b>.gpx</b> exportado do seu relógio ou ciclocomputador (Garmin, Wahoo,
            XOSS e outros). Isso não muda seu plano nem a prescrição de nenhum treino — os dados ficam disponíveis
            aqui para você conferir e, se quiser, usar ao concluir o treino do dia correspondente em{' '}
            <Link href="/plano">seu plano</Link>.
          </span>
        </header>

        <form
          className="import-form"
          onSubmit={(event) => event.preventDefault()}
        >
          <label htmlFor="activity-file">Arquivo da atividade</label>
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
          <small>Tamanho máximo de 20 MB. O arquivo não é guardado depois de lido; só o resumo da atividade é salvo.</small>
          {uploading && (
            <p className="form-notice" role="status">
              <LoaderCircle className="spin" size={14} /> Lendo o arquivo…
            </p>
          )}
          {uploadError && <p className="form-error" role="alert">{uploadError}</p>}
        </form>

        {lastImported && (
          <section className="import-result" aria-live="polite">
            <h2>Atividade importada.</h2>
            <p>
              {dateFormatter.format(new Date(lastImported.started_at))} · {formatDuration(lastImported.moving_seconds)}
              {lastImported.distance_km > 0 && ` · ${lastImported.distance_km.toFixed(1)} km`}
            </p>
            {candidates.length > 0 ? (
              <>
                <p>
                  Encontramos {candidates.length === 1 ? 'um treino planejado' : 'treinos planejados'} no mesmo dia.
                  Vincule para guardar a relação. Se abrir no plano, os dados acima já aparecem preenchidos no
                  formulário, e você decide se confirma.
                </p>
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
                            <span className="link-done"><Link2 size={14} aria-hidden="true" />Vinculada</span>
                          ) : (
                            <button
                              type="button"
                              className="btn btn-outline btn-sm"
                              disabled={busyActivity === lastImported.id}
                              onClick={() => void linkActivity(lastImported.id, candidate.id)}
                            >
                              {busyActivity === lastImported.id ? <LoaderCircle className="spin" size={14} /> : <Link2 size={14} />}
                              Vincular
                            </button>
                          )}
                          <Link href={planLinkFor(candidate.id, lastImported)}>Abrir no plano</Link>
                        </span>
                      </li>
                    );
                  })}
                </ul>
              </>
            ) : (
              <p>Não encontramos um treino planejado nesse dia. Os dados ficam guardados aqui, sem vínculo.</p>
            )}
          </section>
        )}

        <h2 className="activities-subheading">Atividades já importadas</h2>
        {activities.length === 0 && (
          <section className="activities-empty">
            <UploadCloud size={25} />
            <h2>Nenhuma atividade importada ainda.</h2>
            <p>Envie um arquivo .fit ou .gpx acima para começar.</p>
          </section>
        )}
        <div className="activities-list">
          {activities.map((item) => (
            <article className="activity-card completed" key={item.id}>
              <div className="activity-icon"><UploadCloud size={18} /></div>
              <div className="activity-main">
                <div className="activity-title">
                  <div>
                    <h2>{item.source === 'fit' ? 'Arquivo .fit' : 'Arquivo .gpx'}</h2>
                    <p>
                      {item.workout_id
                        ? `Vinculada a ${item.workout_name ?? 'um treino'}${item.workout_scheduled_on ? ` · ${shortDate(item.workout_scheduled_on)}` : ''}`
                        : 'Sem treino vinculado'}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="icon-button"
                    onClick={() => void handleDelete(item.id)}
                    aria-label="Remover atividade importada"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
                <time><CalendarDays size={14} />{dateFormatter.format(new Date(item.started_at))}</time>
                <div className="activity-link-row">
                  {item.workout_id ? (
                    <>
                      <Link href={planLinkFor(item.workout_id, item)}>Abrir no plano</Link>
                      <button
                        type="button"
                        className="btn btn-outline btn-sm"
                        disabled={busyActivity === item.id}
                        onClick={() => void linkActivity(item.id, null)}
                      >
                        {busyActivity === item.id ? <LoaderCircle className="spin" size={14} /> : <Link2Off size={14} />}
                        Desvincular
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
                      Vincular a um treino
                    </button>
                  )}
                </div>
                {pickerFor === item.id && (
                  <div className="link-picker" aria-live="polite">
                    {pickerCandidates === null ? (
                      <p><LoaderCircle className="spin" size={14} /> Buscando treinos perto dessa data…</p>
                    ) : pickerCandidates.length === 0 ? (
                      <p>Nenhum treino planejado perto dessa data. Você pode gerar ou ajustar o plano e voltar aqui.</p>
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
                              Vincular
                            </button>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )}
                <div className="activity-metrics">
                  <span><Clock3 size={14} /><b>{formatDuration(item.moving_seconds)}</b></span>
                  {item.distance_km > 0 && <span><MapPinned size={14} /><b>{item.distance_km.toFixed(1)} km</b></span>}
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
