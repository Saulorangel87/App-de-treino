'use client';

import { FormEvent, useState } from 'react';
import { ArrowRight, CheckCircle2, KeyRound } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiRequest } from '@/lib/api';

export default function ResetPasswordPage() {
  const [loading, setLoading] = useState(false); const [message, setMessage] = useState(''); const [error, setError] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) { event.preventDefault(); setLoading(true); setError(''); const token = new URLSearchParams(window.location.search).get('token'); if (!token) { setError('Este link está incompleto.'); setLoading(false); return; } try { const data = new FormData(event.currentTarget); const password = String(data.get('password') || ''); if (password !== data.get('confirmation')) { setError('As senhas não coincidem.'); return; } const result = await apiRequest<{ message: string }>('/v1/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }); setMessage(result.message); } catch (caught) { setError(caught instanceof Error ? caught.message : 'Não foi possível redefinir a senha.'); } finally { setLoading(false); } }
  return <main className="auth-action-shell"><section className="auth-action-card">{message ? <CheckCircle2 size={34} /> : <KeyRound size={34} />}<p className="form-kicker">ACESSO À CONTA</p><h1>{message ? 'Senha redefinida.' : 'Escolha uma nova senha.'}</h1>{message ? <><p>{message}</p><a className="auth-action-link" href="/entrar">Entrar no Cadência</a></> : <form className="account-form" onSubmit={submit}><div><Label htmlFor="password">Nova senha</Label><Input id="password" name="password" type="password" minLength={10} maxLength={72} required autoComplete="new-password" /></div><div><Label htmlFor="confirmation">Confirme a nova senha</Label><Input id="confirmation" name="confirmation" type="password" minLength={10} maxLength={72} required autoComplete="new-password" /></div>{error && <p className="form-error">{error}</p>}<Button type="submit" disabled={loading} className="account-submit">{loading ? 'Salvando…' : 'Salvar nova senha'}<ArrowRight size={16} /></Button></form>}</section></main>;
}
