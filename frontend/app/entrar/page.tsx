'use client';

import { FormEvent, useState } from 'react';
import { ArrowRight, Bike, Check, ShieldCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';

export default function SignInPage() {
  const [mode, setMode] = useState<'login' | 'register'>('register');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [developmentVerificationURL, setDevelopmentVerificationURL] = useState('');

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError('');
    setNotice('');
    setDevelopmentVerificationURL('');
    const data = new FormData(event.currentTarget);
    try {
      const result = await apiRequest<{ message?: string; development_verification_url?: string }>(`/v1/auth/${mode === 'register' ? 'register' : 'login'}`, {
        method: 'POST',
        body: JSON.stringify({
          email: data.get('email'),
          password: data.get('password'),
          ...(mode === 'register' ? { display_name: data.get('display_name') } : {}),
        }),
      });
      if (mode === 'register') {
        setNotice(result.message || 'Enviamos um link para confirmar seu e-mail.');
        setDevelopmentVerificationURL(result.development_verification_url || '');
        return;
      }
      window.location.href = '/';
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Não foi possível continuar.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="account-shell">
      <section className="account-story">
        <a href="/" className="account-brand"><span><Bike size={21} /></span>cadência</a>
        <div className="story-copy">
          <p className="eyebrow-light">SEU TREINO, SEU CONTEXTO</p>
          <h1>Treinar melhor começa por conhecer você.</h1>
          <p>Seu plano respeita disponibilidade, experiência, recuperação e evolução — sem atalhos ou treinos inventados.</p>
          <ul>
            <li><Check size={15} /> Dados protegidos no seu próprio banco</li>
            <li><Check size={15} /> Decisões explicáveis e baseadas em regras</li>
            <li><Check size={15} /> Começamos apenas com ciclismo</li>
          </ul>
        </div>
        <div className="privacy-note"><ShieldCheck size={17} /><span><strong>Seus dados ficam protegidos</strong>A aplicação controla o acesso às suas informações com segurança.</span></div>
      </section>
      <section className="account-form-panel">
        <div className="account-form-wrap">
          <div className="mode-switch" aria-label="Escolha a forma de acesso">
            <button className={mode === 'register' ? 'active' : ''} onClick={() => { setMode('register'); setError(''); setNotice(''); }}>Criar conta</button>
            <button className={mode === 'login' ? 'active' : ''} onClick={() => { setMode('login'); setError(''); setNotice(''); }}>Entrar</button>
          </div>
          <p className="form-kicker">{mode === 'register' ? 'PRIMEIRO PASSO' : 'BEM-VINDO DE VOLTA'}</p>
          <h2>{mode === 'register' ? 'Vamos começar pelo básico.' : 'Continue sua evolução.'}</h2>
          <p className="form-intro">{mode === 'register' ? 'Poucas informações agora. Seu perfil será construído aos poucos.' : 'Entre para acessar seu plano e registrar seus treinos.'}</p>
          <form onSubmit={submit} className="account-form">
            {mode === 'register' && <div><Label htmlFor="display_name">Como podemos chamar você?</Label><Input id="display_name" name="display_name" minLength={2} maxLength={100} required placeholder="Seu nome" autoComplete="name" /></div>}
            <div><Label htmlFor="email">E-mail</Label><Input id="email" name="email" type="email" required placeholder="voce@exemplo.com" autoComplete="email" /></div>
            <div><Label htmlFor="password">Senha</Label><Input id="password" name="password" type="password" minLength={10} maxLength={72} required placeholder="No mínimo 10 caracteres" autoComplete={mode === 'register' ? 'new-password' : 'current-password'} /></div>
            {error && <p className="form-error" role="alert">{error}</p>}
            {notice && <p className="form-notice" role="status">{notice}</p>}
            <Button type="submit" disabled={loading} className="account-submit">{loading ? 'Aguarde…' : mode === 'register' ? 'Criar minha conta' : 'Entrar'}<ArrowRight size={16} /></Button>
          </form>
          {mode === 'login' && <a className="form-link" href="/esqueci-minha-senha">Esqueci minha senha</a>}
          {developmentVerificationURL && <a className="form-link" href={developmentVerificationURL}>Abrir confirmação local</a>}
          {notice && mode === 'register' && <a className="form-link" href="/perfil">Já confirmei meu e-mail</a>}
          <p className="form-legal">Ao continuar, você concorda em fornecer dados de treino para personalização. O Cadência não realiza diagnóstico clínico.</p>
        </div>
      </section>
    </main>
  );
}
