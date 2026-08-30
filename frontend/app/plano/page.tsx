'use client';

import { useEffect, useMemo, useState } from 'react';
import { ArrowLeft, Bike, CalendarDays, Check, Clock3, Gauge, LoaderCircle, RefreshCw, ShieldAlert, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { apiRequest } from '@/lib/api';

type User = { display_name: string; email: string };
type Workout = {
  id: string;
  scheduled_on: string;
  name: string;
  objective: string;
  duration_minutes: number;
  target_rpe: number;
  structure: { warmup_minutes?: number; main?: string; cooldown_minutes?: number };
  explanation: { summary?: string; rules?: string[] };
  status: string;
};
type Plan = {
  id: string;
  starts_on: string;
  ends_on: string;
  status: string;
  prescription_snapshot: { engine_version?: string; experience_level?: string; primary_goal?: string; restricted?: boolean; sessions_per_week?: number };
  workouts: Workout[];
};

const dateFormatter = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: 'short' });
const fullDateFormatter = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: 'long', year: 'numeric' });

function parseDate(value: string) { return new Date(`${value}T12:00:00`); }

export default function PlanPage() {
  const [user, setUser] = useState<User | null>(null);
  const [plan, setPlan] = useState<Plan | null>(null);
  const [selected, setSelected] = useState<Workout | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ plan: Plan | null }>('/v1/plans/current'),
    ]).then(([account, current]) => {
      setUser(account.user);
      setPlan(current.plan);
      if (current.plan?.workouts.length) setSelected(current.plan.workouts[0]);
    }).catch(() => { window.location.href = '/entrar'; }).finally(() => setLoading(false));
  }, []);

  const weeks = useMemo(() => {
    if (!plan) return [];
    const start = parseDate(plan.starts_on).getTime();
    return [0, 1, 2, 3].map((week) => plan.workouts.filter((workout) => Math.floor((parseDate(workout.scheduled_on).getTime() - start) / 604800000) === week));
  }, [plan]);
  const totalMinutes = useMemo(() => plan?.workouts.reduce((sum, workout) => sum + workout.duration_minutes, 0) || 0, [plan]);

  async function generate(replace = false) {
    if (replace && !window.confirm('Substituir o rascunho atual por um novo plano calculado com seus dados mais recentes?')) return;
    setGenerating(true);
    setError('');
    try {
      const result = await apiRequest<{ plan: Plan }>('/v1/plans/generate', { method: 'POST' });
      setPlan(result.plan);
      setSelected(result.plan.workouts[0] || null);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Não foi possível gerar o plano.');
    } finally {
      setGenerating(false);
    }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando seu plano…</main>;

  return (
    <main className="plan-shell">
      <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><div><small>ATLETA</small><strong>{user?.display_name}</strong></div></header>
      <section className="plan-content">
        <a href="/" className="back-link"><ArrowLeft size={15} /> Voltar ao painel</a>
        {!plan ? <section className="plan-empty">
          <span><Sparkles size={24} /></span><p>PLANEJAMENTO · REGRAS V1</p><h1>Seu contexto já pode virar um plano.</h1><div>O Cadência usará sua experiência, objetivo, limitações e disponibilidade para criar quatro semanas explicáveis.</div>
          {error && <p className="form-error" role="alert">{error}</p>}
          <Button onClick={() => generate()} disabled={generating} className="plan-generate">{generating ? <LoaderCircle className="spin" /> : <Sparkles size={16} />}{generating ? 'Calculando…' : 'Gerar meu primeiro plano'}</Button>
          <small>O resultado será um rascunho. Nenhum treino substitui avaliação profissional.</small>
        </section> : <>
          <header className="plan-heading"><div><p>MEU PLANO · 4 SEMANAS</p><h1>Uma progressão que cabe na sua rotina.</h1><span>{fullDateFormatter.format(parseDate(plan.starts_on))} até {fullDateFormatter.format(parseDate(plan.ends_on))}</span></div><div className="plan-heading-actions"><span className="draft-pill">RASCUNHO</span><Button variant="outline" onClick={() => generate(true)} disabled={generating}><RefreshCw className={generating ? 'spin' : ''} size={14} /> Recalcular</Button></div></header>
          {plan.prescription_snapshot.restricted && <div className="plan-safety"><ShieldAlert size={18} /><div><strong>Modo de segurança ativo</strong><p>As sessões foram limitadas a esforço leve por causa da condição informada no perfil.</p></div></div>}
          {error && <p className="form-error" role="alert">{error}</p>}
          <div className="plan-stats"><div><strong>{plan.workouts.length}</strong><span>sessões</span></div><div><strong>{Math.floor(totalMinutes / 60)}h {totalMinutes % 60}min</strong><span>volume total</span></div><div><strong>{plan.prescription_snapshot.sessions_per_week}</strong><span>dias por semana</span></div><div><strong>Regras V1</strong><span>motor utilizado</span></div></div>
          <div className="plan-layout">
            <div className="plan-weeks">{weeks.map((workouts, index) => <section className="plan-week" key={index}><header><span>SEMANA {index + 1}</span><small>{index === 3 ? 'RECUPERAÇÃO' : index === 2 ? 'MAIOR CARGA' : 'PROGRESSÃO'}</small></header><div>{workouts.map((workout) => <button type="button" key={workout.id} className={selected?.id === workout.id ? 'selected' : ''} onClick={() => setSelected(workout)}><time>{dateFormatter.format(parseDate(workout.scheduled_on))}</time><span><strong>{workout.name}</strong><small>{workout.objective}</small></span><em><Clock3 size={12} />{workout.duration_minutes} min</em><em><Gauge size={12} />RPE {workout.target_rpe}</em></button>)}</div></section>)}</div>
            {selected && <aside className="workout-detail"><span>SESSÃO SELECIONADA</span><h2>{selected.name}</h2><p>{selected.explanation.summary}</p><div className="detail-metrics"><div><Clock3 size={15} /><strong>{selected.duration_minutes} min</strong></div><div><Gauge size={15} /><strong>RPE {selected.target_rpe}</strong></div></div><h3>Estrutura</h3><ol><li><b>{selected.structure.warmup_minutes} min</b>Aquecimento progressivo</li><li><b>Principal</b>{selected.structure.main}</li><li><b>{selected.structure.cooldown_minutes} min</b>Desaquecimento leve</li></ol><h3>Por que este treino?</h3><ul>{selected.explanation.rules?.map((rule) => <li key={rule}><Check size={12} />{rule}</li>)}</ul></aside>}
          </div>
        </>}
      </section>
    </main>
  );
}
