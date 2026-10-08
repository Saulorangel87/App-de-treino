'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowRight, CheckCircle2, KeyRound } from 'lucide-react';
import { Brand } from '@/components/brand';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';

const messages = defineMessages({
  pt: {
    incomplete: 'Este link está incompleto.',
    mismatch: 'As senhas não coincidem.',
    failed: 'Não foi possível redefinir a senha.',
    kicker: 'ACESSO À CONTA',
    done: 'Senha redefinida.',
    title: 'Escolha uma nova senha.',
    signIn: 'Entrar no Cadência',
    password: 'Nova senha',
    confirmation: 'Confirme a nova senha',
    saving: 'Salvando…',
    save: 'Salvar nova senha',
  },
  en: {
    incomplete: 'This link is incomplete.',
    mismatch: "The passwords don't match.",
    failed: 'The password could not be reset.',
    kicker: 'ACCOUNT ACCESS',
    done: 'Password reset.',
    title: 'Choose a new password.',
    signIn: 'Sign in to Cadência',
    password: 'New password',
    confirmation: 'Confirm the new password',
    saving: 'Saving…',
    save: 'Save new password',
  },
});

export default function ResetPasswordPage() {
  const t = useMessages(messages);
  const [loading, setLoading] = useState(false); const [message, setMessage] = useState(''); const [error, setError] = useState('');
  async function submit(event: React.SubmitEvent<HTMLFormElement>) { event.preventDefault(); setLoading(true); setError(''); const token = new URLSearchParams(window.location.search).get('token'); if (!token) { setError(t.incomplete); setLoading(false); return; } try { const data = new FormData(event.currentTarget); const passwordField = data.get('password'); const password = typeof passwordField === 'string' ? passwordField : ''; if (password !== data.get('confirmation')) { setError(t.mismatch); return; } const result = await apiRequest<{ message: string }>('/v1/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }); setMessage(result.message); } catch (caught) { setError(caught instanceof Error ? caught.message : t.failed); } finally { setLoading(false); } }
  return <main className="auth-action-shell"><Brand /><section className="auth-action-card">{message ? <CheckCircle2 size={34} /> : <KeyRound size={34} />}<p className="form-kicker">{t.kicker}</p><h1>{message ? t.done : t.title}</h1>{message ? <><p>{message}</p><Link className="auth-action-link" href="/entrar">{t.signIn}</Link></> : <form method="post" className="account-form" onSubmit={submit}><div><Label htmlFor="password">{t.password}</Label><Input id="password" name="password" type="password" minLength={10} maxLength={72} required autoComplete="new-password" /></div><div><Label htmlFor="confirmation">{t.confirmation}</Label><Input id="confirmation" name="confirmation" type="password" minLength={10} maxLength={72} required autoComplete="new-password" /></div>{error && <p className="form-error">{error}</p>}<Button type="submit" disabled={loading} className="account-submit">{loading ? t.saving : t.save}<ArrowRight size={16} /></Button></form>}</section></main>;
}
