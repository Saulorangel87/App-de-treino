'use client';

import { FormEvent, useEffect, useState } from 'react';
import { ArrowLeft, ArrowRight, Bike, Check, LoaderCircle, ShieldAlert } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';

type User = { display_name: string; email: string };
type Profile = { birth_date?: string | null; sex?: string | null; height_cm?: number | null; weight_kg?: number | null; experience_level: string; activity_level?: string | null };

export default function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<Profile>({ experience_level: 'beginner' });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ profile: Profile }>('/v1/profile').catch(() => null),
    ]).then(([account, saved]) => {
      setUser(account.user);
      if (saved) setProfile(saved.profile);
    }).catch(() => { window.location.href = '/entrar'; }).finally(() => setLoading(false));
  }, []);

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setMessage('');
    const data = new FormData(event.currentTarget);
    try {
      const result = await apiRequest<{ profile: Profile }>('/v1/profile', {
        method: 'PUT',
        body: JSON.stringify({
          birth_date: data.get('birth_date') || null,
          sex: data.get('sex') || null,
          height_cm: data.get('height_cm') ? Number(data.get('height_cm')) : null,
          weight_kg: data.get('weight_kg') ? Number(data.get('weight_kg')) : null,
          experience_level: data.get('experience_level'),
          activity_level: data.get('activity_level') || null,
        }),
      });
      setProfile(result.profile);
      setMessage('Perfil salvo. Já temos uma base segura para personalizar seu plano.');
    } catch (caught) {
      setMessage(caught instanceof Error ? caught.message : 'Não foi possível salvar.');
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <main className="profile-loading"><LoaderCircle className="spin" />Carregando seu perfil…</main>;

  return (
    <main className="profile-shell">
      <header className="profile-topbar"><a href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</a><div><small>CONTA</small><strong>{user?.display_name}</strong></div></header>
      <section className="profile-content">
        <a href="/" className="back-link"><ArrowLeft size={15} /> Voltar ao painel</a>
        <div className="profile-heading"><div><p>PERFIL DO ATLETA · ETAPA 1</p><h1>Conte-nos onde você está agora.</h1><span>Esses dados definem os limites iniciais. Você poderá atualizá-los quando quiser.</span></div><div className="step-indicator"><strong>01</strong><span>de 04</span></div></div>
        <div className="profile-layout">
          <form onSubmit={save} className="profile-form">
            <fieldset><legend>Informações básicas</legend><div className="form-grid"><div><Label htmlFor="birth_date">Data de nascimento</Label><Input id="birth_date" name="birth_date" type="date" defaultValue={profile.birth_date || ''} /></div><div><Label htmlFor="sex">Sexo</Label><select id="sex" name="sex" defaultValue={profile.sex || ''}><option value="">Prefiro não informar agora</option><option value="female">Feminino</option><option value="male">Masculino</option><option value="other">Outro</option><option value="prefer_not_to_say">Prefiro não dizer</option></select></div><div><Label htmlFor="height_cm">Altura</Label><div className="unit-input"><Input id="height_cm" name="height_cm" type="number" min="100" max="250" step="0.1" defaultValue={profile.height_cm || ''} /><span>cm</span></div></div><div><Label htmlFor="weight_kg">Peso atual</Label><div className="unit-input"><Input id="weight_kg" name="weight_kg" type="number" min="30" max="350" step="0.1" defaultValue={profile.weight_kg || ''} /><span>kg</span></div></div></div></fieldset>
            <fieldset><legend>Experiência no ciclismo</legend><div className="choice-cards">{[['beginner','Iniciante','Estou começando ou retomando'],['intermediate','Intermediário','Pedalo com regularidade'],['advanced','Avançado','Treino estruturado há anos']].map(([value,title,description]) => <label key={value}><input type="radio" name="experience_level" value={value} defaultChecked={profile.experience_level === value} /><span><strong>{title}</strong><small>{description}</small></span></label>)}</div><div className="activity-field"><Label htmlFor="activity_level">Como está sua rotina de atividade hoje?</Label><select id="activity_level" name="activity_level" defaultValue={profile.activity_level || ''}><option value="">Selecione</option><option value="sedentary">Quase não pratico atividade</option><option value="occasional">1–2 vezes por semana</option><option value="regular">3–4 vezes por semana</option><option value="frequent">5 ou mais vezes por semana</option></select></div></fieldset>
            {message && <p className="profile-message" role="status"><Check size={15} />{message}</p>}
            <Button type="submit" disabled={saving} className="profile-submit">{saving ? 'Salvando…' : 'Salvar e continuar'}<ArrowRight size={16} /></Button>
          </form>
          <aside className="profile-aside"><div className="aside-icon"><ShieldAlert size={18} /></div><h2>Segurança em primeiro lugar</h2><p>Na próxima etapa perguntaremos sobre lesões, dor e limitações. Essas respostas restringem o que o motor poderá prescrever.</p><hr /><span>O Cadência personaliza treinos, mas não realiza diagnóstico médico.</span></aside>
        </div>
      </section>
    </main>
  );
}
