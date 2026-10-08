'use client';

import { useEffect, useMemo, useState } from 'react';
import { CheckCircle2, HeartPulse, LoaderCircle, MoonStar, ShieldAlert } from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { defineMessages } from '@/lib/i18n';
import { zoneLabel } from '@/lib/zones';

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

const messages = defineMessages({
  pt: {
    readiness: {
      ready: ['Recuperação dentro do esperado', 'Seu check-in foi salvo. O plano não recebeu aumento automático de carga.'],
      caution: ['Hoje pede atenção', 'Um sinal ficou abaixo do habitual e o próximo treino foi ajustado com cautela, quando havia uma sessão futura.'],
      recovery_needed: ['Priorize recuperação', 'A combinação dos sinais indicou necessidade de reduzir a próxima carga, quando havia uma sessão futura.'],
    } as Record<Recovery['readiness'], [string, string]>,
    loadFailed: 'Não foi possível carregar sua recuperação.',
    saveFailed: 'Não foi possível salvar o check-in.',
    loading: 'Carregando sua recuperação…',
    kicker: 'CHECK-IN DIÁRIO',
    title: 'Como você chega para hoje?',
    intro: 'Registre sono, estresse e fadiga percebida. O app usa esses sinais apenas para manter ou reduzir a próxima carga — nunca para aumentá-la automaticamente.',
    adaptedNote: 'Por segurança, editar o check-in depois não aumenta novamente essa sessão.',
    guideTitle: 'Uma leitura simples',
    guideText: 'Responda como você realmente se sente. Um único sinal ruim gera cautela; fadiga máxima ou uma combinação de sinais reduz a próxima sessão ainda não iniciada.',
    sleep: 'Sono',
    sleepText: 'Duração e qualidade percebida da última noite.',
    stressFatigue: 'Estresse e fadiga',
    stressFatigueText: 'Escalas subjetivas de 1 a 5, considerando o momento atual.',
    warning: 'Este check-in não diagnostica condições de saúde. Dor, tontura, falta de ar incomum ou mal-estar são motivos para não iniciar o treino e buscar orientação profissional quando necessário.',
    reference: 'Referência:',
    referenceText: ', consenso sobre monitoramento de carga e resposta do atleta.',
    formTitle: 'Registro de hoje',
    slept: 'Quanto você dormiu?',
    sleepQuality: 'Qualidade do sono',
    veryBadF: 'Muito ruim',
    veryGoodF: 'Muito boa',
    stress: 'Nível de estresse',
    veryLow: 'Muito baixo',
    veryHigh: 'Muito alto',
    fatigue: 'Fadiga percebida',
    veryLowF: 'Muito baixa',
    veryHighF: 'Muito alta',
    notes: 'Observações opcionais',
    notesPlaceholder: 'Ex.: dormi interrompido, dia mais exigente…',
    saving: 'Salvando…',
    update: 'Atualizar check-in',
    save: 'Salvar check-in',
  },
  en: {
    readiness: {
      ready: ['Recovery as expected', 'Your check-in was saved. The plan got no automatic load increase.'],
      caution: ['Take it easy today', 'One sign was below your usual and the next workout was adjusted carefully, if there was an upcoming session.'],
      recovery_needed: ['Prioritize recovery', 'The combination of signs called for reducing the next load, if there was an upcoming session.'],
    },
    loadFailed: 'Your recovery could not be loaded.',
    saveFailed: 'The check-in could not be saved.',
    loading: 'Loading your recovery…',
    kicker: 'DAILY CHECK-IN',
    title: 'How are you feeling today?',
    intro: 'Record sleep, stress and perceived fatigue. The app only uses these signs to keep or reduce the next load — never to increase it automatically.',
    adaptedNote: 'For safety, editing the check-in later does not increase this session again.',
    guideTitle: 'A simple reading',
    guideText: 'Answer how you really feel. A single bad sign calls for caution; maximum fatigue or a combination of signs reduces the next session that has not started yet.',
    sleep: 'Sleep',
    sleepText: "Duration and perceived quality of last night's sleep.",
    stressFatigue: 'Stress and fatigue',
    stressFatigueText: 'Subjective scales from 1 to 5, based on how you feel right now.',
    warning: 'This check-in does not diagnose health conditions. Pain, dizziness, unusual shortness of breath or feeling unwell are reasons not to start the workout and to seek professional advice when needed.',
    reference: 'Reference:',
    referenceText: ', consensus statement on monitoring athlete training loads and responses.',
    formTitle: "Today's log",
    slept: 'How long did you sleep?',
    sleepQuality: 'Sleep quality',
    veryBadF: 'Very poor',
    veryGoodF: 'Very good',
    stress: 'Stress level',
    veryLow: 'Very low',
    veryHigh: 'Very high',
    fatigue: 'Perceived fatigue',
    veryLowF: 'Very low',
    veryHighF: 'Very high',
    notes: 'Optional notes',
    notesPlaceholder: 'E.g.: broken sleep, a more demanding day…',
    saving: 'Saving…',
    update: 'Update check-in',
    save: 'Save check-in',
  },
});

export default function RecoveryPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const today = useMemo(() => localDateKey(), []);
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
    }).catch((caught) => {
      if (caught instanceof ApiError && caught.status === 401) {
        window.location.href = '/entrar';
        return;
      }
      setError(apiErrorMessage(caught, t.loadFailed));
    }).finally(() => setLoading(false));
  }, [today, t]);

  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault(); setSaving(true); setError('');
    try {
      const result = await apiRequest<{ recovery: Recovery }>('/v1/recovery/today', {
        method: 'PUT', body: JSON.stringify({ recorded_on: today, sleep_minutes: sleepMinutes, sleep_quality: sleepQuality, stress_level: stressLevel, fatigue_level: fatigueLevel, notes }),
      });
      setRecovery(result.recovery);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : t.saveFailed);
    } finally { setSaving(false); }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />{t.loading}</main>;
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;
  const resultCopy = recovery ? t.readiness[recovery.readiness] : null;
  return <main className="recovery-shell">
    <AppHeader name={user.display_name} />
    <section className="recovery-content">
      <header className="recovery-heading"><p>{t.kicker}</p><h1>{t.title}</h1><span>{t.intro}</span></header>
      {recovery && resultCopy && <section className={`recovery-result ${recovery.readiness}`}><CheckCircle2 size={22} /><div><strong>{resultCopy[0]}</strong><p>{resultCopy[1]}</p>{recovery.adapted_workout && <p className="adapted-recovery-workout"><b>{recovery.adapted_workout.name}</b>: {recovery.adapted_workout.duration_minutes} min · {zoneLabel(recovery.adapted_workout.target_rpe, locale)}. {t.adaptedNote}</p>}</div></section>}
      <div className="recovery-layout">
        <section className="recovery-guide"><span className="recovery-icon"><HeartPulse size={23} /></span><h2>{t.guideTitle}</h2><p>{t.guideText}</p><ul><li><MoonStar size={17} /><span><strong>{t.sleep}</strong>{t.sleepText}</span></li><li><HeartPulse size={17} /><span><strong>{t.stressFatigue}</strong>{t.stressFatigueText}</span></li></ul><div className="recovery-warning"><ShieldAlert size={18} /><p>{t.warning}</p></div><p className="recovery-evidence">{t.reference} <a href="https://pubmed.ncbi.nlm.nih.gov/28253038/" target="_blank" rel="noreferrer">Bourdon et al. (2017)</a>{t.referenceText}</p></section>
        <form className="recovery-form" onSubmit={submit}><h2>{t.formTitle}</h2><label><span>{t.slept}</span><select value={sleepMinutes} onChange={(event) => setSleepMinutes(Number(event.target.value))}>{Array.from({ length: 13 }, (_, index) => 240 + index * 30).map((minutes) => <option key={minutes} value={minutes}>{Math.floor(minutes / 60)}h{minutes % 60 ? '30' : ''}</option>)}</select></label><Scale label={t.sleepQuality} value={sleepQuality} onChange={setSleepQuality} low={t.veryBadF} high={t.veryGoodF} /><Scale label={t.stress} value={stressLevel} onChange={setStressLevel} low={t.veryLow} high={t.veryHigh} /><Scale label={t.fatigue} value={fatigueLevel} onChange={setFatigueLevel} low={t.veryLowF} high={t.veryHighF} /><label><span>{t.notes}</span><textarea maxLength={1000} value={notes} onChange={(event) => setNotes(event.target.value)} placeholder={t.notesPlaceholder} /></label>{error && <p className="form-error" role="alert">{error}</p>}<button type="submit" disabled={saving}>{saving ? <LoaderCircle className="spin" size={17} /> : <HeartPulse size={17} />}{saving ? t.saving : recovery ? t.update : t.save}</button></form>
      </div>
    </section>
  </main>;
}

function Scale({ label, value, onChange, low, high }: { label: string; value: number; onChange: (value: number) => void; low: string; high: string }) {
  return <fieldset className="recovery-scale"><legend>{label}</legend><div>{[1, 2, 3, 4, 5].map((item) => <button type="button" key={item} className={value === item ? 'selected' : ''} onClick={() => onChange(item)} aria-pressed={value === item}>{item}</button>)}</div><small><span>{low}</span><span>{high}</span></small></fieldset>;
}
