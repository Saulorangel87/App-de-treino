'use client';

import { useEffect, useState } from 'react';
import { Activity, ArrowLeft, Bike, CalendarDays, Clock3, Gauge, HeartPulse, LoaderCircle, MapPinned, XCircle, Zap } from 'lucide-react';
import { apiRequest } from '@/lib/api';
import { parseTrainingDate, type Activity as TrainingActivity } from '@/lib/planning';

type User = { display_name: string };

const dateFormatter = new Intl.DateTimeFormat('pt-BR', {
  day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
});

const difficultyLabels: Record<NonNullable<TrainingActivity['feedback']>['difficulty'], string> = {
  very_easy: 'Muito fácil', easy: 'Fácil', moderate: 'Moderado', hard: 'Difícil', very_hard: 'Muito difícil',
};

export default function ActivitiesPage() {
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
        if (caught instanceof Error && caught.message === 'Faça login para continuar.') {
          window.location.href = '/entrar';
          return;
        }
        setError(caught instanceof Error ? caught.message : 'Não foi possível carregar suas atividades.');
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando suas atividades…</main>;

  return (
    <main className="activities-shell">
      <header className="profile-topbar">
        <a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a>
        <div><small>ATLETA</small><strong>{user?.display_name}</strong></div>
      </header>
      <section className="activities-content">
        <a href="/" className="back-link"><ArrowLeft size={15} />Voltar ao painel</a>
        <header className="activities-heading">
          <p>HISTÓRICO DE TREINOS</p>
          <h1>Suas atividades.</h1>
          <span>Concluídas e canceladas, da mais recente para a mais antiga.</span>
        </header>
        {error && <p className="form-error" role="alert">{error}</p>}
        {!error && activities.length === 0 && (
          <section className="activities-empty">
            <Activity size={25} />
            <h2>Ainda não há atividades registradas.</h2>
            <p>Quando você concluir ou cancelar uma sessão, ela aparecerá aqui.</p>
            <a href="/plano">Ver meu plano</a>
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
                  <div className="activity-title"><div><h2>{item.name}</h2><p>{item.objective}</p></div><span className="activity-status">{completed ? 'Concluída' : 'Cancelada'}</span></div>
                  <time><CalendarDays size={14} />{terminalAt ? dateFormatter.format(new Date(terminalAt)) : parseTrainingDate(item.scheduled_on).toLocaleDateString('pt-BR')}</time>
                  <div className="activity-metrics">
                    <span><Clock3 size={14} /><b>{item.duration_minutes ?? '—'}</b> {item.duration_minutes === undefined ? 'duração' : 'min'}</span>
                    <span><Gauge size={14} /><b>{item.actual_rpe ? `RPE ${item.actual_rpe}` : '—'}</b></span>
                    {item.distance_km !== undefined && <span><MapPinned size={14} /><b>{item.distance_km} km</b></span>}
                    {item.elevation_gain_m !== undefined && <span><MapPinned size={14} /><b>{item.elevation_gain_m} m+</b></span>}
                    {item.average_heart_rate !== undefined && <span><HeartPulse size={14} /><b>{item.average_heart_rate} bpm</b></span>}
                    {item.average_power_watts !== undefined && <span><Zap size={14} /><b>{item.average_power_watts} W</b></span>}
                    {item.feedback && <><span><b>{difficultyLabels[item.feedback.difficulty]}</b></span><span>Fadiga <b>{item.feedback.fatigue_after}/5</b></span><span className={item.feedback.pain_reported ? 'pain' : ''}>{item.feedback.pain_reported ? 'Dor relatada' : 'Sem dor'}</span></>}
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
