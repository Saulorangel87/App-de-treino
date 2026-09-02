'use client';

import { useEffect, useState } from 'react';
import { CheckCircle2, MailCheck, TriangleAlert } from 'lucide-react';
import { apiRequest } from '@/lib/api';

export default function VerifyEmailPage() {
  const [state, setState] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('Confirmando seu e-mail…');

  useEffect(() => {
    const token = new URLSearchParams(window.location.search).get('token');
    if (!token) { setState('error'); setMessage('Este link de confirmação está incompleto.'); return; }
    apiRequest<{ message: string }>('/v1/auth/verify-email', { method: 'POST', body: JSON.stringify({ token }) })
      .then((result) => { setState('success'); setMessage(result.message); })
      .catch((error) => { setState('error'); setMessage(error instanceof Error ? error.message : 'Não foi possível confirmar seu e-mail.'); });
  }, []);

  return <main className="auth-action-shell"><section className="auth-action-card">
    {state === 'success' ? <CheckCircle2 size={34} /> : state === 'error' ? <TriangleAlert size={34} /> : <MailCheck size={34} />}
    <p className="form-kicker">SEGURANÇA DA CONTA</p><h1>{state === 'success' ? 'E-mail confirmado.' : state === 'error' ? 'Não foi possível confirmar.' : 'Só um instante.'}</h1><p>{message}</p>
    <a className="auth-action-link" href={state === 'success' ? '/perfil' : '/entrar'}>{state === 'success' ? 'Continuar para o perfil' : 'Voltar para entrar'}</a>
  </section></main>;
}
