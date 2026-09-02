'use client';

import { FormEvent, useState } from 'react';
import { ArrowRight, KeyRound } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';

export default function ForgotPasswordPage() {
  const [loading, setLoading] = useState(false); const [message, setMessage] = useState(''); const [error, setError] = useState(''); const [developmentResetURL, setDevelopmentResetURL] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) { event.preventDefault(); setLoading(true); setError(''); setDevelopmentResetURL(''); try { const data = new FormData(event.currentTarget); const result = await apiRequest<{ message: string; development_reset_url?: string }>('/v1/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email: data.get('email') }) }); setMessage(result.message); setDevelopmentResetURL(result.development_reset_url || ''); } catch (caught) { setError(caught instanceof Error ? caught.message : 'Não foi possível continuar.'); } finally { setLoading(false); } }
  return <main className="auth-action-shell"><section className="auth-action-card"><KeyRound size={34} /><p className="form-kicker">ACESSO À CONTA</p><h1>Redefina sua senha.</h1><p>Informe seu e-mail. Se houver uma conta, enviaremos um link seguro.</p><form className="account-form" onSubmit={submit}><div><Label htmlFor="email">E-mail</Label><Input id="email" name="email" type="email" required autoComplete="email" placeholder="voce@exemplo.com" /></div>{error && <p className="form-error">{error}</p>}{message && <p className="form-notice">{message}</p>}<Button type="submit" disabled={loading} className="account-submit">{loading ? 'Enviando…' : 'Enviar link'}<ArrowRight size={16} /></Button></form>{developmentResetURL && <a className="form-link" href={developmentResetURL}>Abrir redefinição local</a>}<a className="form-link" href="/entrar">Voltar para entrar</a></section></main>;
}
