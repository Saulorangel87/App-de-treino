'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowRight, Check, ShieldCheck } from 'lucide-react';
import { Brand } from '@/components/brand';
import { LanguageSwitcher } from '@/components/language-switcher';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';
import { LEGAL_VERSION } from '@/lib/legal';

const messages = defineMessages({
  pt: {
    sent: 'Enviamos um link para confirmar seu e-mail.',
    failed: 'Não foi possível continuar.',
    storyKicker: 'SEU TREINO, SEU CONTEXTO',
    storyTitle: 'Treinar melhor começa por conhecer você.',
    storyIntro: 'Seu plano respeita disponibilidade, experiência, recuperação e evolução — sem atalhos ou treinos inventados.',
    points: ['Dados protegidos no seu próprio banco', 'Decisões explicáveis e baseadas em regras', 'Começamos apenas com ciclismo'],
    privacyTitle: 'Seus dados ficam protegidos',
    privacyText: 'A aplicação controla o acesso às suas informações com segurança.',
    modeLabel: 'Escolha a forma de acesso',
    register: 'Criar conta',
    login: 'Entrar',
    registerKicker: 'PRIMEIRO PASSO',
    loginKicker: 'BEM-VINDO DE VOLTA',
    registerTitle: 'Vamos começar pelo básico.',
    loginTitle: 'Continue sua evolução.',
    registerIntro: 'Poucas informações agora. Seu perfil será construído aos poucos.',
    loginIntro: 'Entre para acessar seu plano e registrar seus treinos.',
    name: 'Como podemos chamar você?',
    namePlaceholder: 'Seu nome',
    email: 'E-mail',
    emailPlaceholder: 'voce@exemplo.com',
    password: 'Senha',
    passwordPlaceholder: 'No mínimo 10 caracteres',
    acceptBefore: 'Li e aceito os ',
    terms: 'Termos de Uso',
    acceptMiddle: ' e a ',
    privacy: 'Política de Privacidade',
    acceptAfter: ', inclusive o tratamento dos dados de saúde que eu informar para montar meus treinos.',
    wait: 'Aguarde…',
    createAccount: 'Criar minha conta',
    forgot: 'Esqueci minha senha',
    devLink: 'Abrir confirmação local',
    confirmed: 'Já confirmei meu e-mail',
    legal: 'Ao continuar, você concorda em fornecer dados de treino para personalização. O Cadência não realiza diagnóstico clínico.',
  },
  en: {
    sent: 'We sent you a link to confirm your email.',
    failed: 'Something went wrong. Please try again.',
    storyKicker: 'YOUR TRAINING, YOUR CONTEXT',
    storyTitle: 'Training better starts with knowing you.',
    storyIntro: 'Your plan respects your availability, experience, recovery and progress — no shortcuts or made-up workouts.',
    points: ['Your data protected in your own database', 'Explainable, rule-based decisions', 'We start with cycling only'],
    privacyTitle: 'Your data stays protected',
    privacyText: 'The app controls access to your information securely.',
    modeLabel: 'Choose how to access',
    register: 'Create account',
    login: 'Sign in',
    registerKicker: 'FIRST STEP',
    loginKicker: 'WELCOME BACK',
    registerTitle: "Let's start with the basics.",
    loginTitle: 'Keep making progress.',
    registerIntro: 'Just a few details for now. Your profile is built step by step.',
    loginIntro: 'Sign in to see your plan and log your workouts.',
    name: 'What should we call you?',
    namePlaceholder: 'Your name',
    email: 'Email',
    emailPlaceholder: 'you@example.com',
    password: 'Password',
    passwordPlaceholder: 'At least 10 characters',
    acceptBefore: 'I have read and accept the ',
    terms: 'Terms of Use',
    acceptMiddle: ' and the ',
    privacy: 'Privacy Policy',
    acceptAfter: ', including the processing of the health data I enter to build my workouts.',
    wait: 'Please wait…',
    createAccount: 'Create my account',
    forgot: 'I forgot my password',
    devLink: 'Open local confirmation',
    confirmed: "I've confirmed my email",
    legal: 'By continuing, you agree to provide training data for personalization. Cadência does not provide clinical diagnosis.',
  },
});

export default function SignInPage() {
  const t = useMessages(messages);
  const [mode, setMode] = useState<'login' | 'register'>('register');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [developmentVerificationURL, setDevelopmentVerificationURL] = useState('');

  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
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
          ...(mode === 'register' ? { display_name: data.get('display_name'), accept_terms: data.get('accept_terms') === 'on', terms_version: LEGAL_VERSION } : {}),
        }),
      });
      if (mode === 'register') {
        setNotice(result.message || t.sent);
        setDevelopmentVerificationURL(result.development_verification_url || '');
        return;
      }
      window.location.href = '/';
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : t.failed);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="account-shell">
      <LanguageSwitcher compact className="account-language" />
      <section className="account-story topo-surface">
        <Brand />
        <div className="story-copy">
          <p className="eyebrow-light">{t.storyKicker}</p>
          <h1>{t.storyTitle}</h1>
          <p>{t.storyIntro}</p>
          <ul>
            {t.points.map((point) => <li key={point}><Check size={15} /> {point}</li>)}
          </ul>
        </div>
        <div className="privacy-note"><ShieldCheck size={17} /><span><strong>{t.privacyTitle}</strong>{t.privacyText}</span></div>
      </section>
      <section className="account-form-panel">
        <div className="account-form-wrap">
          <div className="mode-switch" aria-label={t.modeLabel}>
            <button className={mode === 'register' ? 'active' : ''} onClick={() => { setMode('register'); setError(''); setNotice(''); }}>{t.register}</button>
            <button className={mode === 'login' ? 'active' : ''} onClick={() => { setMode('login'); setError(''); setNotice(''); }}>{t.login}</button>
          </div>
          <p className="form-kicker">{mode === 'register' ? t.registerKicker : t.loginKicker}</p>
          <h2>{mode === 'register' ? t.registerTitle : t.loginTitle}</h2>
          <p className="form-intro">{mode === 'register' ? t.registerIntro : t.loginIntro}</p>
          <form method="post" onSubmit={submit} className="account-form">
            {mode === 'register' && <div><Label htmlFor="display_name">{t.name}</Label><Input id="display_name" name="display_name" minLength={2} maxLength={100} required placeholder={t.namePlaceholder} autoComplete="name" /></div>}
            <div><Label htmlFor="email">{t.email}</Label><Input id="email" name="email" type="email" required placeholder={t.emailPlaceholder} autoComplete="email" /></div>
            <div><Label htmlFor="password">{t.password}</Label><Input id="password" name="password" type="password" minLength={10} maxLength={72} required placeholder={t.passwordPlaceholder} autoComplete={mode === 'register' ? 'new-password' : 'current-password'} /></div>
            {mode === 'register' && <label className="terms-check"><input type="checkbox" name="accept_terms" required /><span>{t.acceptBefore}<Link href="/termos" target="_blank">{t.terms}</Link>{t.acceptMiddle}<Link href="/privacidade" target="_blank">{t.privacy}</Link>{t.acceptAfter}</span></label>}
            {error && <p className="form-error" role="alert">{error}</p>}
            {notice && <p className="form-notice" role="status">{notice}</p>}
            <Button type="submit" disabled={loading} className="account-submit">{loading ? t.wait : mode === 'register' ? t.createAccount : t.login}<ArrowRight size={16} /></Button>
          </form>
          {mode === 'login' && <Link className="form-link" href="/esqueci-minha-senha">{t.forgot}</Link>}
          {developmentVerificationURL && <a className="form-link" href={developmentVerificationURL}>{t.devLink}</a>}
          {notice && mode === 'register' && <Link className="form-link" href="/perfil">{t.confirmed}</Link>}
          <p className="form-legal">{t.legal}</p>
        </div>
      </section>
    </main>
  );
}
