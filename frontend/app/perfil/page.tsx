'use client';

import { FormEvent, useEffect, useMemo, useState } from 'react';
import { ArrowLeft, ArrowRight, Bike, CalendarDays, Check, CircleAlert, Flag, HeartPulse, LoaderCircle, MailCheck, ShieldAlert } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';
import { AccountActions } from '@/components/account-actions';

type User = { display_name: string; email: string; email_verified: boolean };
type Profile = { birth_date?: string | null; sex?: string | null; height_cm?: number | null; weight_kg?: number | null; experience_level: string; activity_level?: string | null };
type Limitation = { kind: string; description: string; is_active: boolean; professional_clearance_recommended: boolean };
type Goal = { goal_type: string; priority: number; target_date?: string | null; details: string };
type Availability = { weekday: number; available_minutes: number; preferred_time?: string | null; location?: string | null };
type CyclingContext = { weekly_hours: number; longest_ride_minutes: number; weekly_rides: number; recent_weekly_distance_km: number; recent_training_weeks: number; recent_best_distance_km: number; preferred_session_types: string[]; bike_type: string; terrain: string; uses_heart_rate: boolean; uses_power: boolean; ftp?: number; event_goal: boolean; event_distance_km?: number; event_date?: string };
type Onboarding = { limitations: Limitation[]; goals: Goal[]; availability: Availability[]; cycling_context: CyclingContext };

type ProfileForm = {
  birth_date: string;
  sex: string;
  height_cm: string;
  weight_kg: string;
  experience_level: string;
  activity_level: string;
};

const DAYS = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];
const initialProfile: ProfileForm = { birth_date: '', sex: '', height_cm: '', weight_kg: '', experience_level: 'beginner', activity_level: '' };
const initialAvailability = (): Availability[] => DAYS.map((_, weekday) => ({ weekday, available_minutes: 0, preferred_time: null, location: null }));
const initialCyclingContext: CyclingContext = { weekly_hours: 0, longest_ride_minutes: 0, weekly_rides: 0, recent_weekly_distance_km: 0, recent_training_weeks: 0, recent_best_distance_km: 0, preferred_session_types: [], bike_type: '', terrain: '', uses_heart_rate: false, uses_power: false, event_goal: false };
const SESSION_PREFERENCES = [{ value: 'base', label: 'Giro/base' }, { value: 'cadence', label: 'Cadência' }, { value: 'hills', label: 'Subidas' }, { value: 'intervals', label: 'Intervalos' }, { value: 'sweet_spot', label: 'Sweet spot' }, { value: 'recovery', label: 'Recuperação' }];

const stepCopy = [
  { kicker: 'PERFIL DO ATLETA · ETAPA 1', title: 'Conte-nos onde você está agora.', description: 'Esses dados definem os limites iniciais. Você poderá atualizá-los quando quiser.', icon: ShieldAlert, asideTitle: 'Uma base segura', aside: 'Experiência e rotina ajudam o Cadência a começar com uma carga compatível com seu momento.' },
  { kicker: 'SEGURANÇA · ETAPA 2', title: 'Existe algo que o treino deve respeitar?', description: 'Dor, lesões e limitações sempre têm prioridade sobre desempenho.', icon: HeartPulse, asideTitle: 'Segurança em primeiro lugar', aside: 'Uma limitação ativa restringe o que o motor poderá prescrever. O Cadência não realiza diagnóstico médico.' },
  { kicker: 'DIREÇÃO · ETAPA 3', title: 'Onde você quer chegar?', description: 'Defina um objetivo principal e, se desejar, uma prioridade secundária.', icon: Flag, asideTitle: 'Objetivos realistas', aside: 'O plano combinará sua meta com experiência, segurança e tempo disponível — nunca apenas com ambição.' },
  { kicker: 'ROTINA · ETAPA 4', title: 'Quanto tempo cabe na sua semana?', description: 'Marque os dias possíveis. Descanso também faz parte do plano.', icon: CalendarDays, asideTitle: 'Consistência vence excesso', aside: 'Usaremos somente os períodos que você informou e reservaremos espaço suficiente para recuperação.' },
];

export default function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<ProfileForm>(initialProfile);
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [hasLimitation, setHasLimitation] = useState(false);
  const [limitation, setLimitation] = useState({ kind: 'pain', description: '', professional_clearance_recommended: false });
  const [primaryGoal, setPrimaryGoal] = useState({ goal_type: 'health', target_date: '', details: '' });
  const [secondaryGoal, setSecondaryGoal] = useState('');
  const [availability, setAvailability] = useState<Availability[]>(initialAvailability);
  const [cyclingContext, setCyclingContext] = useState<CyclingContext>(initialCyclingContext);
  const [completed, setCompleted] = useState(false);
  const [verificationMessage, setVerificationMessage] = useState('');
  const [sendingVerification, setSendingVerification] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const [account, profileResult] = await Promise.all([
          apiRequest<{ user: User }>('/v1/me'),
          apiRequest<{ profile: Profile | null }>('/v1/profile'),
        ]);
        setUser(account.user);
        if (!profileResult.profile) return;

        const savedProfile = profileResult.profile;
        setProfile({
          birth_date: savedProfile.birth_date || '',
          sex: savedProfile.sex || '',
          height_cm: savedProfile.height_cm?.toString() || '',
          weight_kg: savedProfile.weight_kg?.toString() || '',
          experience_level: savedProfile.experience_level,
          activity_level: savedProfile.activity_level || '',
        });
        setStep(2);

        const { onboarding } = await apiRequest<{ onboarding: Onboarding }>('/v1/onboarding');
        const savedCyclingContext = onboarding.cycling_context || {};
        setCyclingContext({
          ...initialCyclingContext,
          ...savedCyclingContext,
          preferred_session_types: Array.isArray(savedCyclingContext.preferred_session_types) ? savedCyclingContext.preferred_session_types : [],
        });
        if (onboarding.limitations.length) {
          const saved = onboarding.limitations[0];
          setHasLimitation(true);
          setLimitation({ kind: saved.kind, description: saved.description, professional_clearance_recommended: saved.professional_clearance_recommended });
        }
        if (onboarding.goals.length) {
          const primary = onboarding.goals.find((goal) => goal.priority === 1) || onboarding.goals[0];
          const secondary = onboarding.goals.find((goal) => goal.priority === 2);
          setPrimaryGoal({ goal_type: primary.goal_type, target_date: primary.target_date || '', details: primary.details || '' });
          setSecondaryGoal(secondary?.goal_type || '');
          setStep(4);
        }
        if (onboarding.availability.length === 7) {
          setAvailability(onboarding.availability);
          setStep(4);
          setCompleted(onboarding.availability.some((day) => day.available_minutes > 0));
        }
      } catch {
        window.location.href = '/entrar';
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  useEffect(() => {
    setError('');
    setMessage('');
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, [step]);

  const totalMinutes = useMemo(() => availability.reduce((sum, day) => sum + day.available_minutes, 0), [availability]);
  const trainingDays = useMemo(() => availability.filter((day) => day.available_minutes > 0).length, [availability]);
  const copy = stepCopy[step - 1];
  const AsideIcon = copy.icon;

  function updateProfile(field: keyof ProfileForm, value: string) {
    setProfile((current) => ({ ...current, [field]: value }));
  }

  function updateDay(weekday: number, patch: Partial<Availability>) {
    setAvailability((current) => current.map((day) => day.weekday === weekday ? { ...day, ...patch } : day));
  }

  function toggleSessionPreference(value: string) {
    setCyclingContext((current) => ({
      ...current,
      preferred_session_types: (current.preferred_session_types || []).includes(value)
        ? (current.preferred_session_types || []).filter((item) => item !== value)
        : [...(current.preferred_session_types || []), value],
    }));
  }

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const result = await apiRequest<{ profile: Profile }>('/v1/profile', {
        method: 'PUT',
        body: JSON.stringify({
          birth_date: profile.birth_date || null,
          sex: profile.sex || null,
          height_cm: profile.height_cm ? Number(profile.height_cm) : null,
          weight_kg: profile.weight_kg ? Number(profile.weight_kg) : null,
          experience_level: profile.experience_level,
          activity_level: profile.activity_level || null,
        }),
      });
      setProfile((current) => ({ ...current, experience_level: result.profile.experience_level }));
      setStep(2);
    });
  }

  async function saveLimitations(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const limitations: Limitation[] = hasLimitation ? [{ ...limitation, is_active: true }] : [];
      await apiRequest('/v1/onboarding/limitations', { method: 'PUT', body: JSON.stringify({ limitations }) });
      setStep(3);
    });
  }

  async function saveGoals(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSave(async () => {
      const goals: Goal[] = [{ goal_type: primaryGoal.goal_type, priority: 1, target_date: primaryGoal.target_date || null, details: primaryGoal.details }];
      if (secondaryGoal) goals.push({ goal_type: secondaryGoal, priority: 2, target_date: null, details: '' });
      await apiRequest('/v1/onboarding/goals', { method: 'PUT', body: JSON.stringify({ goals }) });
      setStep(4);
    });
  }

  async function saveAvailability(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const updatingCompletedProfile = completed;
    await runSave(async () => {
      await apiRequest('/v1/onboarding/cycling-context', { method: 'PUT', body: JSON.stringify({ cycling_context: cyclingContext }) });
      await apiRequest('/v1/onboarding/availability', { method: 'PUT', body: JSON.stringify({ availability }) });
      setCompleted(true);
      setMessage(
        updatingCompletedProfile
          ? 'Disponibilidade salva. Abra Meu plano e escolha Atualizar plano para gerar um rascunho com esta nova rotina.'
          : 'Perfil concluído. Seu contexto inicial está salvo com segurança.',
      );
    });
  }

  async function runSave(action: () => Promise<void>) {
    setSaving(true);
    setError('');
    setMessage('');
    try {
      await action();
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Não foi possível salvar esta etapa.');
    } finally {
      setSaving(false);
    }
  }

  async function resendVerification() {
    setSendingVerification(true);
    setVerificationMessage('');
    try {
      const result = await apiRequest<{ message: string; development_verification_url?: string }>('/v1/auth/resend-verification', { method: 'POST' });
      setVerificationMessage(result.development_verification_url ? `${result.message} Abra o link local abaixo.` : result.message);
      if (result.development_verification_url) window.open(result.development_verification_url, '_blank', 'noopener,noreferrer');
    } catch (caught) {
      setVerificationMessage(caught instanceof Error ? caught.message : 'Não foi possível reenviar a confirmação.');
    } finally {
      setSendingVerification(false);
    }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando seu perfil…</main>;

  return (
    <main className="profile-shell">
      <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><AccountActions label="CONTA" name={user?.display_name} /></header>
      <section className="profile-content">
        <a href="/" className="back-link"><ArrowLeft size={15} /> Voltar ao painel</a>
        {user && !user.email_verified && <section className="email-verification-banner"><MailCheck size={19} /><div><strong>Confirme seu e-mail antes de gerar ou ativar um plano.</strong><p>{verificationMessage || `Enviamos um link para ${user.email}.`}</p></div><Button type="button" variant="outline" disabled={sendingVerification} onClick={resendVerification}>{sendingVerification ? 'Enviando…' : 'Reenviar link'}</Button></section>}
        <nav className="onboarding-progress" aria-label="Progresso do perfil">
          {stepCopy.map((_, index) => <span key={index} className={index + 1 <= step ? 'active' : ''}><i>{index + 1 < step || completed ? <Check size={11} /> : index + 1}</i></span>)}
        </nav>
        <div className="profile-heading"><div><p>{copy.kicker}</p><h1>{copy.title}</h1><span>{copy.description}</span></div><div className="step-indicator"><strong>0{step}</strong><span>de 04</span></div></div>
        <div className="profile-layout">
          {step === 1 && <form onSubmit={saveProfile} className="profile-form">
            <fieldset><legend>Informações básicas</legend><div className="form-grid"><div><Label htmlFor="birth_date">Data de nascimento</Label><Input id="birth_date" type="date" value={profile.birth_date} onChange={(event) => updateProfile('birth_date', event.target.value)} /></div><div><Label htmlFor="sex">Sexo</Label><select id="sex" value={profile.sex} onChange={(event) => updateProfile('sex', event.target.value)}><option value="">Prefiro não informar agora</option><option value="female">Feminino</option><option value="male">Masculino</option><option value="other">Outro</option><option value="prefer_not_to_say">Prefiro não dizer</option></select></div><div><Label htmlFor="height_cm">Altura</Label><div className="unit-input"><Input id="height_cm" type="number" min="100" max="250" step="0.1" value={profile.height_cm} onChange={(event) => updateProfile('height_cm', event.target.value)} /><span>cm</span></div></div><div><Label htmlFor="weight_kg">Peso atual</Label><div className="unit-input"><Input id="weight_kg" type="number" min="30" max="350" step="0.1" value={profile.weight_kg} onChange={(event) => updateProfile('weight_kg', event.target.value)} /><span>kg</span></div></div></div></fieldset>
            <fieldset><legend>Experiência no ciclismo</legend><div className="choice-cards">{[['beginner','Iniciante','Estou começando ou retomando'],['intermediate','Intermediário','Pedalo com regularidade'],['advanced','Avançado','Treino estruturado há anos']].map(([value,title,description]) => <label key={value}><input type="radio" name="experience_level" value={value} checked={profile.experience_level === value} onChange={(event) => updateProfile('experience_level', event.target.value)} /><span><strong>{title}</strong><small>{description}</small></span></label>)}</div><div className="activity-field"><Label htmlFor="activity_level">Como está sua rotina de atividade hoje?</Label><select id="activity_level" value={profile.activity_level} onChange={(event) => updateProfile('activity_level', event.target.value)}><option value="">Selecione</option><option value="sedentary">Quase não pratico atividade</option><option value="occasional">1–2 vezes por semana</option><option value="regular">3–4 vezes por semana</option><option value="frequent">5 ou mais vezes por semana</option></select></div></fieldset>
            <FormFeedback error={error} message={message} />
            <Button type="submit" disabled={saving} className="profile-submit">{saving ? 'Salvando…' : 'Salvar e continuar'}<ArrowRight size={16} /></Button>
          </form>}

          {step === 2 && <form onSubmit={saveLimitations} className="profile-form">
            <fieldset><legend>Condição atual</legend><div className="binary-choice"><label><input type="radio" name="has_limitation" checked={!hasLimitation} onChange={() => setHasLimitation(false)} /><span><strong>Nenhuma limitação atual</strong><small>Posso pedalar sem dor ou restrição conhecida</small></span></label><label><input type="radio" name="has_limitation" checked={hasLimitation} onChange={() => setHasLimitation(true)} /><span><strong>Tenho algo a considerar</strong><small>Dor, lesão, condição ou restrição de movimento</small></span></label></div></fieldset>
            {hasLimitation && <fieldset><legend>O que devemos respeitar?</legend><div className="form-grid"><div><Label htmlFor="limitation_kind">Tipo</Label><select id="limitation_kind" value={limitation.kind} onChange={(event) => setLimitation((current) => ({ ...current, kind: event.target.value }))}><option value="pain">Dor ou desconforto</option><option value="injury">Lesão</option><option value="medical_condition">Condição de saúde</option><option value="mobility">Limitação de movimento</option><option value="other">Outro</option></select></div><label className="clearance-check"><input type="checkbox" checked={limitation.professional_clearance_recommended} onChange={(event) => setLimitation((current) => ({ ...current, professional_clearance_recommended: event.target.checked }))} /><span><strong>Orientação profissional recomendada</strong><small>Marque se um médico ou fisioterapeuta deve liberar o treino</small></span></label></div><div className="textarea-field"><Label htmlFor="limitation_description">Descreva brevemente</Label><textarea id="limitation_description" minLength={3} maxLength={500} required value={limitation.description} onChange={(event) => setLimitation((current) => ({ ...current, description: event.target.value }))} placeholder="Ex.: desconforto no joelho direito ao subir…" /></div></fieldset>}
            <FormFeedback error={error} message={message} />
            <div className="form-actions"><Button type="button" variant="outline" onClick={() => setStep(1)}><ArrowLeft size={15} /> Voltar</Button><Button type="submit" disabled={saving} className="profile-submit">{saving ? 'Salvando…' : 'Salvar e continuar'}<ArrowRight size={16} /></Button></div>
          </form>}

          {step === 3 && <form onSubmit={saveGoals} className="profile-form">
            <fieldset><legend>Objetivo principal</legend><div className="form-grid"><div><Label htmlFor="primary_goal">Quero principalmente</Label><select id="primary_goal" value={primaryGoal.goal_type} onChange={(event) => setPrimaryGoal((current) => ({ ...current, goal_type: event.target.value }))}><GoalOptions /></select></div><div><Label htmlFor="target_date">Data-alvo (opcional)</Label><Input id="target_date" type="date" value={primaryGoal.target_date} min={new Date().toISOString().slice(0, 10)} onChange={(event) => setPrimaryGoal((current) => ({ ...current, target_date: event.target.value }))} /></div></div><div className="textarea-field"><Label htmlFor="goal_details">Conte um pouco mais (opcional)</Label><textarea id="goal_details" maxLength={500} value={primaryGoal.details} onChange={(event) => setPrimaryGoal((current) => ({ ...current, details: event.target.value }))} placeholder="Ex.: completar meu primeiro pedal de 100 km com segurança…" /></div></fieldset>
            <fieldset><legend>Objetivo secundário</legend><div className="activity-field"><Label htmlFor="secondary_goal">Outra prioridade (opcional)</Label><select id="secondary_goal" value={secondaryGoal} onChange={(event) => setSecondaryGoal(event.target.value)}><option value="">Nenhuma por enquanto</option><GoalOptions exclude={primaryGoal.goal_type} /></select></div></fieldset>
            <FormFeedback error={error} message={message} />
            <div className="form-actions"><Button type="button" variant="outline" onClick={() => setStep(2)}><ArrowLeft size={15} /> Voltar</Button><Button type="submit" disabled={saving} className="profile-submit">{saving ? 'Salvando…' : 'Salvar e continuar'}<ArrowRight size={16} /></Button></div>
          </form>}

          {step === 4 && <form onSubmit={saveAvailability} className="profile-form availability-form">
            <fieldset><legend>Seu ciclismo hoje</legend><p className="fieldset-intro">Essas perguntas são opcionais e ajudam a tornar os próximos treinos mais específicos.</p><div className="form-grid"><div><Label>Horas por semana</Label><Input type="number" min="0" max="80" step="0.5" value={cyclingContext.weekly_hours || ''} onChange={(e) => setCyclingContext(c => ({ ...c, weekly_hours: Number(e.target.value) }))} /></div><div><Label>Pedais por semana</Label><Input type="number" min="0" max="21" value={cyclingContext.weekly_rides || ''} onChange={(e) => setCyclingContext(c => ({ ...c, weekly_rides: Number(e.target.value) }))} /></div><div><Label>Distância semanal recente (km)</Label><Input type="number" min="0" max="2000" step="1" value={cyclingContext.recent_weekly_distance_km || ''} onChange={(e) => setCyclingContext(c => ({ ...c, recent_weekly_distance_km: Number(e.target.value) }))} /></div><div><Label>Semanas treinando com regularidade</Label><Input type="number" min="0" max="52" value={cyclingContext.recent_training_weeks || ''} onChange={(e) => setCyclingContext(c => ({ ...c, recent_training_weeks: Number(e.target.value) }))} /></div><div><Label>Maior distância recente (km)</Label><Input type="number" min="0" max="2000" step="1" value={cyclingContext.recent_best_distance_km || ''} onChange={(e) => setCyclingContext(c => ({ ...c, recent_best_distance_km: Number(e.target.value) }))} /></div><div><Label>Maior pedal recente (min)</Label><Input type="number" min="0" max="1440" value={cyclingContext.longest_ride_minutes || ''} onChange={(e) => setCyclingContext(c => ({ ...c, longest_ride_minutes: Number(e.target.value) }))} /></div><div><Label>Tipo de bicicleta</Label><select value={cyclingContext.bike_type} onChange={(e) => setCyclingContext(c => ({ ...c, bike_type: e.target.value }))}><option value="">Não informar</option><option value="road">Estrada</option><option value="mtb">MTB</option><option value="gravel">Gravel</option><option value="indoor">Indoor/rolo</option></select></div><div><Label>Terreno predominante</Label><select value={cyclingContext.terrain} onChange={(e) => setCyclingContext(c => ({ ...c, terrain: e.target.value }))}><option value="">Não informar</option><option value="flat">Plano</option><option value="rolling">Misto</option><option value="hilly">Com subidas</option></select></div></div><div className="preference-choice"><p>Que tipos de treino você gostaria de fazer?</p><small>Opcional. Isso orienta futuras escolhas sem substituir os critérios de segurança.</small><div>{SESSION_PREFERENCES.map((preference) => <label key={preference.value}><input type="checkbox" checked={cyclingContext.preferred_session_types.includes(preference.value)} onChange={() => toggleSessionPreference(preference.value)} /><span>{preference.label}</span></label>)}</div></div><div className="binary-choice"><label><input type="checkbox" checked={cyclingContext.uses_heart_rate} onChange={(e) => setCyclingContext(c => ({ ...c, uses_heart_rate: e.target.checked }))} /><span><strong>Uso frequência cardíaca</strong></span></label><label><input type="checkbox" checked={cyclingContext.uses_power} onChange={(e) => setCyclingContext(c => ({ ...c, uses_power: e.target.checked, ftp: e.target.checked ? c.ftp : undefined }))} /><span><strong>Uso medidor de potência</strong></span></label><label><input type="checkbox" checked={cyclingContext.event_goal} onChange={(e) => setCyclingContext(c => ({ ...c, event_goal: e.target.checked, event_distance_km: e.target.checked ? c.event_distance_km : undefined, event_date: e.target.checked ? c.event_date : undefined }))} /><span><strong>Estou me preparando para uma prova</strong></span></label></div>{cyclingContext.uses_power && <div className="activity-field"><Label>FTP (watts)</Label><Input type="number" min="50" max="600" value={cyclingContext.ftp || ''} onChange={(e) => setCyclingContext(c => ({ ...c, ftp: Number(e.target.value) || undefined }))} /></div>}{cyclingContext.event_goal && <div className="form-grid"><div><Label>Distância da prova (km)</Label><Input type="number" min="1" max="2000" required value={cyclingContext.event_distance_km || ''} onChange={(e) => setCyclingContext(c => ({ ...c, event_distance_km: Number(e.target.value) || undefined }))} /></div><div><Label>Data da prova</Label><Input type="date" required value={cyclingContext.event_date || ''} min={new Date().toISOString().slice(0, 10)} onChange={(e) => setCyclingContext(c => ({ ...c, event_date: e.target.value || undefined }))} /></div></div>}</fieldset>
            <fieldset><legend>Dias disponíveis</legend><p className="fieldset-intro">Ative um dia e escolha quanto tempo você realmente consegue reservar.</p><div className="availability-grid">{availability.map((day) => { const active = day.available_minutes > 0; return <div className={`availability-day ${active ? 'active' : ''}`} key={day.weekday}><button type="button" aria-pressed={active} onClick={() => updateDay(day.weekday, active ? { available_minutes: 0, preferred_time: null, location: null } : { available_minutes: 60 })}><span>{DAYS[day.weekday]}</span><i>{active && <Check size={12} />}</i></button>{active && <div><select aria-label={`Duração de ${DAYS[day.weekday]}`} value={day.available_minutes} onChange={(event) => updateDay(day.weekday, { available_minutes: Number(event.target.value) })}><option value="30">30 min</option><option value="45">45 min</option><option value="60">1 hora</option><option value="90">1h30</option><option value="120">2 horas</option><option value="180">3 horas</option><option value="240">4 horas</option><option value="360">6 horas</option><option value="480">8 horas</option></select><select aria-label={`Local de ${DAYS[day.weekday]}`} value={day.location || ''} onChange={(event) => updateDay(day.weekday, { location: event.target.value || null })}><option value="">Qualquer local</option><option value="outdoor">Rua/estrada</option><option value="indoor">Rolo/indoor</option><option value="gym">Academia</option></select></div>}</div>; })}</div></fieldset>
            <div className="availability-summary"><div><strong>{trainingDays}</strong><span>dias possíveis</span></div><div><strong>{Math.floor(totalMinutes / 60)}h{totalMinutes % 60 ? ` ${totalMinutes % 60}min` : ''}</strong><span>por semana</span></div><p>O plano poderá usar menos tempo conforme sua recuperação e experiência.</p></div>
            {completed && <div className="completion-card"><span><Check size={20} /></span><div><strong>Perfil inicial concluído</strong><p>Seus dados estão prontos para orientar a próxima fase: a geração do plano.</p></div></div>}
            <FormFeedback error={error} message={message} />
            <div className="form-actions"><Button type="button" variant="outline" onClick={() => setStep(3)}><ArrowLeft size={15} /> Voltar</Button>{completed && <Button type="button" variant="outline" onClick={() => { window.location.href = '/'; }}>Ir para o painel<ArrowRight size={16} /></Button>}<Button type="submit" disabled={saving || totalMinutes === 0} className="profile-submit">{saving ? 'Salvando…' : completed ? 'Salvar alterações' : 'Concluir perfil'}<Check size={16} /></Button></div>
          </form>}

          <aside className="profile-aside"><div className="aside-icon"><AsideIcon size={18} /></div><h2>{copy.asideTitle}</h2><p>{copy.aside}</p><hr /><span>Você poderá revisar todas essas informações quando sua rotina mudar.</span></aside>
        </div>
      </section>
    </main>
  );
}

function GoalOptions({ exclude = '' }: { exclude?: string }) {
  return <>{[
    ['health', 'Melhorar saúde e bem-estar'],
    ['fitness', 'Ganhar condicionamento'],
    ['endurance', 'Pedalar por mais tempo'],
    ['performance', 'Aumentar meu desempenho'],
    ['event', 'Preparar para um evento'],
    ['weight_management', 'Apoiar o controle de peso'],
  ].filter(([value]) => value !== exclude).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</>;
}

function FormFeedback({ error, message }: { error: string; message: string }) {
  return <>{error && <p className="profile-message error" role="alert"><CircleAlert size={15} />{error}</p>}{message && <p className="profile-message" role="status"><Check size={15} />{message}</p>}</>;
}
