'use client';

import { useEffect, useMemo, useState } from 'react';
import { ArrowLeft, Bike, CheckCircle2, HeartPulse, LoaderCircle, MoonStar, ShieldAlert } from 'lucide-react';
import { apiRequest } from '@/lib/api';

type User = { display_name: string };
type AdaptedWorkout = { id: string; scheduled_on: string; name: string; duration_minutes: number; target_rpe: number };
type Recovery = {
  id: string; recorded_on: string; sleep_minutes: number; sleep_quality: number;
  stress_level: number; fatigue_level: number; notes?: string; readiness: 'ready' | 'caution' | 'recovery_needed';
  adaptation_applied: boolean; adapted_workout?: AdaptedWorkout;
};

function localDateKey() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}

const readinessCopy = {
  ready: ['Recuperação dentro do esperado', 'Seu check-in foi salvo. O plano não recebeu aumento automático de carga.'],
  caution: ['Hoje pede atenção', 'Um sinal ficou abaixo do habitual e o próximo treino foi ajustado com cautela, quando havia uma sessão futura.'],
  recovery_needed: ['Priorize recuperação', 'A combinação dos sinais indicou necessidade de reduzir a próxima carga, quando havia uma sessão futura.'],
};

export default function RecoveryPage() {
  const today = useMemo(localDateKey, []);
  const [user, setUser] = useState<User | null>(null);
  const [recovery, setRecovery] = useState<Recovery | null>(null);
  const [sleepMinutes, setSleepMinutes] = useState(480);
  const [sleepQuality, setSleepQuality] = useState(3);
  const [stressLevel, setStressLevel] = useState(3);
  const [fatigueLevel, setFatigueLevel] = useState(3);
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ recovery: Recovery | null }>(`/v1/recovery/today?date=${today}`),
    ]).then(([account, result]) => {
      setUser(account.user);
      setRecovery(result.recovery);
      if (result.recovery) {
        setSleepMinutes(result.recovery.sleep_minutes);
        setSleepQuality(result.recovery.sleep_quality);
        setStressLevel(result.recovery.stress_level);
        setFatigueLevel(result.recovery.fatigue_level);
        setNotes(result.recovery.notes || '');
      }
    }).catch(() => { window.location.href = '/entrar'; }).finally(() => setLoading(false));
  }, [today]);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault(); setSaving(true); setError('');
    try {
      const result = await apiRequest<{ recovery: Recovery }>('/v1/recovery/today', {
        method: 'PUT', body: JSON.stringify({ recorded_on: today, sleep_minutes: sleepMinutes, sleep_quality: sleepQuality, stress_level: stressLevel, fatigue_level: fatigueLevel, notes }),
      });
      setRecovery(result.recovery);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Não foi possível salvar o check-in.');
    } finally { setSaving(false); }
  }

  if (loading || !user) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando sua recuperação…</main>;
  const resultCopy = recovery ? readinessCopy[recovery.readiness] : null;
  return <main className="recovery-shell">
    <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><div><small>ATLETA</small><strong>{user.display_name}</strong></div></header>
    <section className="recovery-content">
      <a href="/" className="back-link"><ArrowLeft size={15} />Voltar ao painel</a>
      <header className="recovery-heading"><p>CHECK-IN DIÁRIO</p><h1>Como você chega para hoje?</h1><span>Registre sono, estresse e fadiga percebida. O app usa esses sinais apenas para manter ou reduzir a próxima carga — nunca para aumentá-la automaticamente.</span></header>
      {recovery && resultCopy && <section className={`recovery-result ${recovery.readiness}`}><CheckCircle2 size={22} /><div><strong>{resultCopy[0]}</strong><p>{resultCopy[1]}</p>{recovery.adapted_workout && <p className="adapted-recovery-workout"><b>{recovery.adapted_workout.name}</b>: {recovery.adapted_workout.duration_minutes} min · RPE {recovery.adapted_workout.target_rpe}. Por segurança, editar o check-in depois não aumenta novamente essa sessão.</p>}</div></section>}
      <div className="recovery-layout">
        <section className="recovery-guide"><span className="recovery-icon"><HeartPulse size={23} /></span><h2>Uma leitura simples</h2><p>Responda como você realmente se sente. Um único sinal ruim gera cautela; fadiga máxima ou uma combinação de sinais reduz a próxima sessão ainda não iniciada.</p><ul><li><MoonStar size={17} /><span><strong>Sono</strong>Duração e qualidade percebida da última noite.</span></li><li><HeartPulse size={17} /><span><strong>Estresse e fadiga</strong>Escalas subjetivas de 1 a 5, considerando o momento atual.</span></li></ul><div className="recovery-warning"><ShieldAlert size={18} /><p>Este check-in não diagnostica condições de saúde. Dor, tontura, falta de ar incomum ou mal-estar são motivos para não iniciar o treino e buscar orientação profissional quando necessário.</p></div><p className="recovery-evidence">Referência: <a href="https://pubmed.ncbi.nlm.nih.gov/28253038/" target="_blank" rel="noreferrer">Bourdon et al. (2017)</a>, consenso sobre monitoramento de carga e resposta do atleta.</p></section>
        <form className="recovery-form" onSubmit={submit}><h2>Registro de hoje</h2><label><span>Quanto você dormiu?</span><select value={sleepMinutes} onChange={(event) => setSleepMinutes(Number(event.target.value))}>{Array.from({ length: 13 }, (_, index) => 240 + index * 30).map((minutes) => <option key={minutes} value={minutes}>{Math.floor(minutes / 60)}h{minutes % 60 ? '30' : ''}</option>)}</select></label><Scale label="Qualidade do sono" value={sleepQuality} onChange={setSleepQuality} low="Muito ruim" high="Muito boa" /><Scale label="Nível de estresse" value={stressLevel} onChange={setStressLevel} low="Muito baixo" high="Muito alto" /><Scale label="Fadiga percebida" value={fatigueLevel} onChange={setFatigueLevel} low="Muito baixa" high="Muito alta" /><label><span>Observações opcionais</span><textarea maxLength={1000} value={notes} onChange={(event) => setNotes(event.target.value)} placeholder="Ex.: dormi interrompido, dia mais exigente…" /></label>{error && <p className="form-error" role="alert">{error}</p>}<button type="submit" disabled={saving}>{saving ? <LoaderCircle className="spin" size={17} /> : <HeartPulse size={17} />}{saving ? 'Salvando…' : recovery ? 'Atualizar check-in' : 'Salvar check-in'}</button></form>
      </div>
    </section>
  </main>;
}

function Scale({ label, value, onChange, low, high }: { label: string; value: number; onChange: (value: number) => void; low: string; high: string }) {
  return <fieldset className="recovery-scale"><legend>{label}</legend><div>{[1, 2, 3, 4, 5].map((item) => <button type="button" key={item} className={value === item ? 'selected' : ''} onClick={() => onChange(item)} aria-pressed={value === item}>{item}</button>)}</div><small><span>{low}</span><span>{high}</span></small></fieldset>;
}
