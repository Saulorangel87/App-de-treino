'use client';

import { useEffect, useState } from 'react';
import { Activity, AlertTriangle, ArrowLeft, Bike, CheckCircle2, Clock3, LoaderCircle } from 'lucide-react';
import { apiRequest } from '@/lib/api';
import { AccountActions } from '@/components/account-actions';

type User = { display_name: string };
type Assessment = { id: string; duration_minutes: number; actual_rpe: number; pain_reported: boolean; eligible_for_progression: boolean };

export default function AssessmentPage() {
  const [user, setUser] = useState<User | null>(null);
  const [assessment, setAssessment] = useState<Assessment | null>(null);
  const [duration, setDuration] = useState(20);
  const [actualRPE, setActualRPE] = useState(5);
  const [painReported, setPainReported] = useState(false);
  const [notes, setNotes] = useState('');
  const [confirmedSafe, setConfirmedSafe] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([apiRequest<{ user: User }>('/v1/me'), apiRequest<{ assessment: Assessment | null }>('/v1/assessments/current')])
      .then(([account, current]) => { setUser(account.user); setAssessment(current.assessment); })
      .catch(() => { window.location.href = '/entrar'; })
      .finally(() => setLoading(false));
  }, []);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault(); setSaving(true); setError('');
    try {
      const result = await apiRequest<{ assessment: Assessment }>('/v1/assessments/submaximal', { method: 'POST', body: JSON.stringify({ duration_minutes: duration, actual_rpe: actualRPE, pain_reported: painReported, notes }) });
      setAssessment(result.assessment);
    } catch (caught) { setError(caught instanceof Error ? caught.message : 'Não foi possível registrar a avaliação.'); } finally { setSaving(false); }
  }

  if (loading || !user) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando sua avaliação…</main>;
  return <main className="assessment-shell">
    <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><AccountActions label="ATLETA" name={user.display_name} /></header>
    <section className="assessment-content">
      <a href="/" className="back-link"><ArrowLeft size={15} />Voltar ao painel</a>
      <header className="assessment-heading"><p>AVALIAÇÃO INICIAL</p><h1>Seu pedal de referência.</h1><span>Uma referência submáxima para orientar a evolução do plano. Não é exame médico nem teste máximo.</span></header>
      {assessment && <section className={`assessment-result ${assessment.eligible_for_progression ? 'eligible' : ''}`}><CheckCircle2 size={21} /><div><strong>Avaliação registrada</strong><p>{assessment.pain_reported ? 'Você relatou dor. O app não usará este resultado para progredir intensidade; priorize recuperação e orientação profissional se a dor persistir.' : assessment.eligible_for_progression ? 'Referência concluída sem sinal de alerta. Ela ficará disponível para progressões futuras, sempre com regras de segurança.' : 'Resultado salvo como referência. O motor continuará com progressão conservadora.'}</p></div></section>}
      <div className="assessment-layout">
        <section className="assessment-card"><span className="assessment-icon"><Activity size={22} /></span><h2>Como realizar</h2><ol><li><b>1</b><span><strong>Aqueça por 5 minutos</strong>Pedale leve e confortável.</span></li><li><b>2</b><span><strong>Pedale de forma contínua</strong>Mantenha esforço controlado, perto de RPE 5/10, por até 20 minutos.</span></li><li><b>3</b><span><strong>Desaqueça</strong>Reduza o ritmo por 5 minutos antes de registrar como se sentiu.</span></li></ol><div className="assessment-warning"><AlertTriangle size={17} /><p><strong>Interrompa imediatamente</strong> se houver dor, tontura, falta de ar incomum, mal-estar ou outro sintoma preocupante.</p></div><p className="assessment-evidence">Base científica: <a href="https://pubmed.ncbi.nlm.nih.gov/8668467/" target="_blank" rel="noreferrer">Dunbar, Kalinski e Robertson (1996)</a>, sobre prescrição submáxima orientada por RPE.</p></section>
        <form className="assessment-form" onSubmit={submit}><h2>Registrar resultado</h2><p>Faça o registro após o pedal. Não tente compensar ou alcançar um número específico.</p><label><span>Duração contínua</span><select value={duration} onChange={(event) => setDuration(Number(event.target.value))}><option value="15">15 minutos</option><option value="20">20 minutos</option><option value="25">25 minutos</option><option value="30">30 minutos</option></select></label><label><span>Como foi o esforço? (RPE 1–10)</span><input type="number" min="1" max="10" step="0.5" value={actualRPE} onChange={(event) => setActualRPE(Number(event.target.value))} required /></label><label className="assessment-check"><input type="checkbox" checked={painReported} onChange={(event) => setPainReported(event.target.checked)} /><span>Relatei dor durante ou após o pedal</span></label><label><span>Observações opcionais</span><textarea maxLength={1000} value={notes} onChange={(event) => setNotes(event.target.value)} /></label><label className="assessment-check"><input type="checkbox" checked={confirmedSafe} onChange={(event) => setConfirmedSafe(event.target.checked)} required /><span>Li as orientações e não realizei esforço máximo.</span></label>{error && <p className="form-error" role="alert">{error}</p>}<button type="submit" disabled={saving}>{saving ? <LoaderCircle className="spin" size={16} /> : <Clock3 size={16} />}{saving ? 'Salvando…' : 'Salvar avaliação'}</button></form>
      </div>
    </section>
  </main>;
}
