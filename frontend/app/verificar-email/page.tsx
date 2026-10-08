'use client';

import { useEffect, useState } from 'react';
import { CheckCircle2, MailCheck, TriangleAlert } from 'lucide-react';
import { Brand } from '@/components/brand';
import { useMessages } from '@/components/locale-provider';
import { apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';

const messages = defineMessages({
  pt: {
    confirming: 'Confirmando seu e-mail…',
    incomplete: 'Este link de confirmação está incompleto.',
    failed: 'Não foi possível confirmar seu e-mail.',
    kicker: 'SEGURANÇA DA CONTA',
    success: 'E-mail confirmado.',
    error: 'Não foi possível confirmar.',
    loading: 'Só um instante.',
    toProfile: 'Continuar para o perfil',
    back: 'Voltar para entrar',
  },
  en: {
    confirming: 'Confirming your email…',
    incomplete: 'This confirmation link is incomplete.',
    failed: 'Your email could not be confirmed.',
    kicker: 'ACCOUNT SECURITY',
    success: 'Email confirmed.',
    error: 'Confirmation failed.',
    loading: 'Just a moment.',
    toProfile: 'Continue to your profile',
    back: 'Back to sign in',
  },
});

export default function VerifyEmailPage() {
  const t = useMessages(messages);
  const [state, setState] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const token = new URLSearchParams(window.location.search).get('token');
    // Todo setState ocorre de forma assíncrona, depois da montagem.
    const request = token
      ? apiRequest<{ message: string }>('/v1/auth/verify-email', { method: 'POST', body: JSON.stringify({ token }) })
      : Promise.reject(new Error(t.incomplete));
    request
      .then((result) => { setState('success'); setMessage(result.message); })
      .catch((error) => { setState('error'); setMessage(error instanceof Error ? error.message : t.failed); });
  }, [t]);

  return <main className="auth-action-shell"><Brand /><section className="auth-action-card">
    {state === 'success' ? <CheckCircle2 size={34} /> : state === 'error' ? <TriangleAlert size={34} /> : <MailCheck size={34} />}
    <p className="form-kicker">{t.kicker}</p><h1>{state === 'success' ? t.success : state === 'error' ? t.error : t.loading}</h1><p>{message || t.confirming}</p>
    <a className="auth-action-link" href={state === 'success' ? '/perfil' : '/entrar'}>{state === 'success' ? t.toProfile : t.back}</a>
  </section></main>;
}
