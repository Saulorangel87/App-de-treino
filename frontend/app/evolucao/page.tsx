'use client';

import { useEffect, useMemo, useState } from 'react';
import { ArrowLeft, Bike, CalendarCheck2, CircleAlert, Clock3, HeartPulse, LineChart, LoaderCircle, MapPinned, MoonStar, Mountain, Target, Zap } from 'lucide-react';
import { apiRequest } from '@/lib/api';
import { AccountActions } from '@/components/account-actions';

type User = { display_name: string };
type Week = { week_start: string; completed_sessions: number; cancelled_sessions: number; total_minutes: number; average_rpe: number; total_distance_km: number; total_elevation_m: number; average_power_watts: number; average_heart_rate: number };
type SessionComparison = { completed_on: string; name: string; planned_minutes: number; actual_minutes?: number; duration_delta_minutes?: number; planned_rpe: number; actual_rpe?: number; rpe_delta?: number; distance_km?: number; average_power_watts?: number; average_heart_rate?: number; fatigue_after?: number; pain_reported: boolean };
type RecoveryPoint = { recorded_on: string; sleep_minutes: number; sleep_quality: number; stress_level: number; fatigue_level: number; readiness: 'ready' | 'caution' | 'recovery_needed' };
type Summary = { completed_sessions: number; cancelled_sessions: number; total_minutes: number; average_rpe: number; average_fatigue: number; completion_rate: number; total_distance_km: number; total_elevation_m: number; average_power_watts: number; average_heart_rate: number; weeks: Week[]; recent_sessions?: SessionComparison[]; recovery: RecoveryPoint[] };

const shortDate = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: 'short' });

function parseDate(value: string) { return new Date(`${value}T12:00:00`); }
function minutesLabel(value: number) { return value >= 60 ? `${Math.floor(value / 60)}h${value % 60 ? ` ${value % 60}min` : ''}` : `${value} min`; }
function distanceLabel(value: number) { return `${value >= 100 ? Math.round(value) : value.toFixed(1).replace('.', ',')} km`; }
function readinessLabel(value: RecoveryPoint['readiness']) { return value === 'ready' ? 'Adequada' : value === 'caution' ? 'Atenção' : 'Recuperação'; }
function signedMinutes(value?: number) { if (value === undefined) return 'Sem duração'; if (value === 0) return 'No tempo previsto'; return `${value > 0 ? '+' : ''}${value} min`; }
function responseLabel(value?: number) { if (value === undefined) return 'Sem RPE informado'; if (value >= 2) return 'Mais esforço'; if (value <= -2) return 'Menos esforço'; return 'Dentro do esperado'; }
function responseTone(value?: number) { if (value === undefined) return 'neutral'; if (value >= 2) return 'high'; if (value <= -2) return 'low'; return 'expected'; }

export default function EvolutionPage() {
  const [user, setUser] = useState<User | null>(null);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([apiRequest<{ user: User }>('/v1/me'), apiRequest<{ summary: Summary }>('/v1/evolution/summary')])
      .then(([account, result]) => { setUser(account.user); setSummary(result.summary); })
      .catch(() => { window.location.href = '/entrar'; })
      .finally(() => setLoading(false));
  }, []);

  const maxMinutes = useMemo(() => Math.max(...(summary?.weeks.map((week) => week.total_minutes) || [1]), 1), [summary]);
  const maxDistance = useMemo(() => Math.max(...(summary?.weeks.map((week) => week.total_distance_km) || [1]), 1), [summary]);
  if (loading || !user || !summary) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando sua evolução…</main>;
  const hasActivities = summary.completed_sessions + summary.cancelled_sessions > 0;
  const hasCyclingMetrics = summary.total_distance_km > 0 || summary.total_elevation_m > 0 || summary.average_power_watts > 0 || summary.average_heart_rate > 0;
  const recentSessions = summary.recent_sessions || [];

  return <main className="evolution-shell">
    <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><AccountActions label="ATLETA" name={user.display_name} /></header>
    <section className="evolution-content">
      <a href="/" className="back-link"><ArrowLeft size={15} />Voltar ao painel</a>
      <header className="evolution-heading"><p>EVOLUÇÃO</p><h1>Seu histórico, com contexto.</h1><span>Registros observados ao longo do tempo. Eles ajudam você a acompanhar consistência e resposta percebida, mas não substituem avaliação profissional nem representam diagnóstico.</span></header>
      {!hasActivities ? <section className="evolution-empty"><LineChart size={28} /><h2>Os primeiros dados aparecerão após seus treinos.</h2><p>Ao concluir ou cancelar sessões, o Cadência passará a organizar sua consistência, duração e esforço percebido aqui.</p><a href="/plano">Ver meu plano</a></section> : <>
        <section className="evolution-metrics" aria-label="Resumo do histórico"><Metric icon={<CalendarCheck2 size={19} />} value={String(summary.completed_sessions)} label="sessões concluídas" /><Metric icon={<Clock3 size={19} />} value={minutesLabel(summary.total_minutes)} label="tempo registrado" /><Metric icon={<Target size={19} />} value={`${Math.round(summary.completion_rate)}%`} label="conclusão das sessões" /><Metric icon={<CircleAlert size={19} />} value={summary.average_rpe ? `RPE ${summary.average_rpe.toFixed(1)}` : '—'} label="esforço médio registrado" /></section>
        <section className="evolution-grid">
          <section className="evolution-card volume-card"><div className="evolution-card-title"><div><span>ÚLTIMAS 8 SEMANAS</span><h2>Tempo concluído</h2></div><small>{summary.completed_sessions} sessões no histórico</small></div><div className="volume-chart" aria-label="Minutos concluídos por semana">{summary.weeks.map((week) => <div className="volume-column" key={week.week_start}><strong>{week.total_minutes ? minutesLabel(week.total_minutes) : '—'}</strong><i style={{ height: `${Math.max(4, week.total_minutes / maxMinutes * 100)}%` }} /><span>{shortDate.format(parseDate(week.week_start)).replace('.', '')}</span></div>)}</div><p className="chart-caption">Sem treino concluído, a semana aparece sem volume. Cancelamentos não entram no tempo registrado.</p>{summary.total_distance_km > 0 && <><div className="cycling-chart-title"><span>DISTÂNCIA POR SEMANA</span><strong>{distanceLabel(summary.total_distance_km)} no histórico</strong></div><div className="distance-chart" aria-label="Distância registrada por semana">{summary.weeks.map((week) => <div className="volume-column" key={week.week_start}><strong>{week.total_distance_km ? distanceLabel(week.total_distance_km) : '—'}</strong><i style={{ height: `${Math.max(4, week.total_distance_km / maxDistance * 100)}%` }} /><span>{shortDate.format(parseDate(week.week_start)).replace('.', '')}</span></div>)}</div></>}</section>
          <section className="evolution-card consistency-card"><div className="evolution-card-title"><div><span>CONSISTÊNCIA</span><h2>Registro das sessões</h2></div></div><div className="consistency-count"><strong>{summary.completed_sessions}</strong><span>concluídas</span></div><div className="consistency-line"><i style={{ width: `${summary.completion_rate}%` }} /></div><div className="consistency-labels"><span>{summary.cancelled_sessions} canceladas</span><span>{Math.round(summary.completion_rate)}% concluídas</span></div><p>Esta taxa considera somente sessões que já foram encerradas.</p></section>
        </section>
        {recentSessions.length > 0 && <section className="evolution-card response-card"><div className="evolution-card-title"><div><span>RESPOSTA AO TREINO</span><h2>Planejado e realizado</h2></div><small>Últimas {recentSessions.length} sessões concluídas</small></div><p className="response-intro">Compare o que estava previsto com o que você registrou. Qualquer adaptação segue as regras de segurança do motor.</p><div className="session-comparison-list">{recentSessions.map((session, index) => <article className="session-comparison-item" key={`${session.completed_on}-${session.name}-${index}`}><div className="session-comparison-heading"><div><time>{shortDate.format(parseDate(session.completed_on)).replace('.', '')}</time><strong>{session.name}</strong></div><span className={`response-tag ${responseTone(session.rpe_delta)}`}>{responseLabel(session.rpe_delta)}</span></div><div className="session-comparison-metrics"><span><small>Planejado</small><b>{session.planned_minutes} min · RPE {session.planned_rpe}</b></span><span><small>Realizado</small><b>{session.actual_minutes !== undefined ? `${session.actual_minutes} min` : '—'}{session.actual_rpe !== undefined ? ` · RPE ${session.actual_rpe}` : ''}</b></span><span><small>Tempo</small><b>{signedMinutes(session.duration_delta_minutes)}</b></span></div><div className="session-comparison-foot">{session.distance_km !== undefined && <span>{distanceLabel(session.distance_km)}</span>}{session.average_heart_rate !== undefined && <span>FC {session.average_heart_rate} bpm</span>}{session.average_power_watts !== undefined && <span>{session.average_power_watts} W</span>}{session.fatigue_after !== undefined && <span>Fadiga {session.fatigue_after}/5</span>}{session.pain_reported && <span className="pain">Dor relatada</span>}</div></article>)}</div></section>}
        {hasCyclingMetrics && <section className="evolution-card cycling-overview"><div className="evolution-card-title"><div><span>DADOS DO PEDAL</span><h2>Métricas registradas</h2></div></div><div className="cycling-overview-grid">{summary.total_distance_km > 0 && <Metric icon={<MapPinned size={18} />} value={distanceLabel(summary.total_distance_km)} label="distância acumulada" />}{summary.total_elevation_m > 0 && <Metric icon={<Mountain size={18} />} value={`${Math.round(summary.total_elevation_m)} m+`} label="elevação acumulada" />}{summary.average_power_watts > 0 && <Metric icon={<Zap size={18} />} value={`${Math.round(summary.average_power_watts)} W`} label="potência média registrada" />}{summary.average_heart_rate > 0 && <Metric icon={<HeartPulse size={18} />} value={`${Math.round(summary.average_heart_rate)} bpm`} label="frequência cardíaca média" />}</div><p className="chart-caption">Cada medida considera somente as sessões concluídas em que ela foi preenchida. Ela descreve registros, não estima desempenho.</p></section>}
      </>}
      <section className="evolution-card recovery-history"><div className="evolution-card-title"><div><span>RECUPERAÇÃO</span><h2>Check-ins recentes</h2></div><a href="/recuperacao"><MoonStar size={15} />Novo check-in</a></div>{summary.recovery.length === 0 ? <p className="recovery-empty">Ainda não há check-ins registrados. Eles aparecerão aqui conforme você preencher a recuperação diária.</p> : <div className="recovery-history-list">{summary.recovery.map((point) => <div className="recovery-history-item" key={point.recorded_on}><time>{shortDate.format(parseDate(point.recorded_on)).replace('.', '')}</time><span className={`readiness-tag ${point.readiness}`}>{readinessLabel(point.readiness)}</span><span>{Math.floor(point.sleep_minutes / 60)}h{point.sleep_minutes % 60 ? '30' : ''} de sono</span><span>Estresse {point.stress_level}/5</span><span>Fadiga {point.fatigue_level}/5</span></div>)}</div>}</section>
    </section>
  </main>;
}

function Metric({ icon, value, label }: { icon: React.ReactNode; value: string; label: string }) { return <article><span>{icon}</span><strong>{value}</strong><small>{label}</small></article>; }
