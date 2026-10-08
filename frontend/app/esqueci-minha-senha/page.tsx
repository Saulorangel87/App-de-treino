'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowRight, KeyRound } from 'lucide-react';
import { Brand } from '@/components/brand';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';

const messages = defineMessages({
  pt: {
    failed: 'Não foi possível continuar.',
    kicker: 'ACESSO À CONTA',
    title: 'Redefina sua senha.',
    intro: 'Informe seu e-mail. Se houver uma conta, enviaremos um link seguro.',
    email: 'E-mail',
    emailPlaceholder: 'voce@exemplo.com',
    sending: 'Enviando…',
    send: 'Enviar link',
    devLink: 'Abrir redefinição local',
    back: 'Voltar para entrar',
  },
  en: {
    failed: 'Something went wrong. Please try again.',
    kicker: 'ACCOUNT ACCESS',
    title: 'Reset your password.',
    intro: "Enter your email. If there's an account, we'll send you a secure link.",
    email: 'Email',
    emailPlaceholder: 'you@example.com',
    sending: 'Sending…',
    send: 'Send link',
    devLink: 'Open local reset',
    back: 'Back to sign in',
  },
});

export default function ForgotPasswordPage() {
  const t = useMessages(messages);
  const [loading, setLoading] = useState(false); const [message, setMessage] = useState(''); const [error, setError] = useState(''); const [developmentResetURL, setDevelopmentResetURL] = useState('');
  async function submit(event: React.SubmitEvent<HTMLFormElement>) { event.preventDefault(); setLoading(true); setError(''); setDevelopmentResetURL(''); try { const data = new FormData(event.currentTarget); const result = await apiRequest<{ message: string; development_reset_url?: string }>('/v1/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email: data.get('email') }) }); setMessage(result.message); setDevelopmentResetURL(result.development_reset_url || ''); } catch (caught) { setError(caught instanceof Error ? caught.message : t.failed); } finally { setLoading(false); } }
  return <main className="auth-action-shell"><Brand /><section className="auth-action-card"><KeyRound size={34} /><p className="form-kicker">{t.kicker}</p><h1>{t.title}</h1><p>{t.intro}</p><form className="account-form" onSubmit={submit}><div><Label htmlFor="email">{t.email}</Label><Input id="email" name="email" type="email" required autoComplete="email" placeholder={t.emailPlaceholder} /></div>{error && <p className="form-error">{error}</p>}{message && <p className="form-notice">{message}</p>}<Button type="submit" disabled={loading} className="account-submit">{loading ? t.sending : t.send}<ArrowRight size={16} /></Button></form>{developmentResetURL && <a className="form-link" href={developmentResetURL}>{t.devLink}</a>}<Link className="form-link" href="/entrar">{t.back}</Link></section></main>;
}
