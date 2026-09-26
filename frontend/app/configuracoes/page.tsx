'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  ArrowRight,
  Bike,
  CircleAlert,
  KeyRound,
  LoaderCircle,
  LogOut,
  Mail,
  ShieldCheck,
  Trash2,
} from 'lucide-react';
import { AccountActions } from '@/components/account-actions';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';

type User = {
  display_name: string;
  email: string;
  email_verified: boolean;
};

const DELETE_CONFIRMATION = 'ENCERRAR CONTA';

export default function SettingsPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);
  const [password, setPassword] = useState('');
  const [confirmation, setConfirmation] = useState('');
  const [error, setError] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [newPasswordConfirmation, setNewPasswordConfirmation] = useState('');
  const [securityBusy, setSecurityBusy] = useState(false);
  const [securityError, setSecurityError] = useState('');
  const [securityMessage, setSecurityMessage] = useState('');

  useEffect(() => {
    apiRequest<{ user: User }>('/v1/me')
      .then(({ user: account }) => setUser(account))
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(
          apiErrorMessage(
            caught,
            'Não foi possível carregar as configurações.',
          ),
        );
      })
      .finally(() => setLoading(false));
  }, []);

  async function changePassword(event: { preventDefault: () => void }) {
    event.preventDefault();
    setSecurityError('');
    setSecurityMessage('');
    if (newPassword !== newPasswordConfirmation) {
      setSecurityError('A confirmação não coincide com a nova senha.');
      return;
    }
    setSecurityBusy(true);
    try {
      const result = await apiRequest<{ message: string }>(
        '/v1/auth/change-password',
        {
          method: 'POST',
          body: JSON.stringify({
            current_password: currentPassword,
            new_password: newPassword,
          }),
        },
      );
      setSecurityMessage(result.message);
      setCurrentPassword('');
      setNewPassword('');
      setNewPasswordConfirmation('');
    } catch (caught) {
      setSecurityError(
        apiErrorMessage(caught, 'Não foi possível alterar a senha.'),
      );
    } finally {
      setSecurityBusy(false);
    }
  }

  async function logoutOthers() {
    setSecurityError('');
    setSecurityMessage('');
    setSecurityBusy(true);
    try {
      const result = await apiRequest<{ revoked_sessions: number }>(
        '/v1/auth/logout-others',
        { method: 'POST' },
      );
      setSecurityMessage(
        result.revoked_sessions > 0
          ? `${result.revoked_sessions} outra(s) sessão(ões) encerrada(s).`
          : 'Não havia outras sessões ativas.',
      );
    } catch (caught) {
      setSecurityError(
        apiErrorMessage(caught, 'Não foi possível encerrar as outras sessões.'),
      );
    } finally {
      setSecurityBusy(false);
    }
  }

  async function deleteAccount(event: { preventDefault: () => void }) {
    event.preventDefault();
    setError('');
    const typedConfirmation = confirmation.trim();
    if (typedConfirmation.toUpperCase() !== DELETE_CONFIRMATION) {
      setError(`Digite ${DELETE_CONFIRMATION} para confirmar o encerramento.`);
      return;
    }
    setDeleting(true);
    try {
      await apiRequest<void>('/v1/auth/account', {
        method: 'DELETE',
        body: JSON.stringify({ password, confirmation: typedConfirmation }),
      });
      Object.keys(window.localStorage)
        .filter((key) => key.startsWith('cadencia:'))
        .forEach((key) => window.localStorage.removeItem(key));
      window.location.replace('/entrar');
    } catch (caught) {
      setError(
        apiErrorMessage(
          caught,
          'Não foi possível encerrar sua conta. Nenhum dado foi alterado.',
        ),
      );
      setDeleting(false);
    }
  }

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando configurações…
      </main>
    );
  }

  if (!user) {
    return (
      <ApiErrorState
        message={error || 'Não foi possível carregar as configurações.'}
      />
    );
  }

  return (
    <main className="settings-shell">
      <header className="profile-topbar">
        <Link href="/" className="account-brand dark">
          <span>
            <Bike size={19} />
          </span>
          cadência
        </Link>
        <AccountActions label="CONFIGURAÇÕES" name={user.display_name} />
      </header>
      <section className="settings-content">
        <Link href="/" className="back-link">
          <ArrowLeft size={15} />
          Voltar ao painel
        </Link>
        <header className="settings-heading">
          <p>CONTA E SEGURANÇA</p>
          <h1>Suas configurações.</h1>
          <span>
            Gerencie o acesso à conta e encontre rapidamente os dados do seu
            perfil de ciclismo.
          </span>
        </header>

        <div className="settings-layout">
          <div className="settings-column">
            <section className="settings-card settings-account-card">
              <span className="settings-icon">
                <ShieldCheck size={23} />
              </span>
              <h2>Dados da conta</h2>
              <p className="settings-card-intro">
                Estas informações identificam seu acesso ao Cadência.
              </p>
              <div className="settings-account-row">
                <Mail size={18} />
                <div>
                  <strong>E-mail</strong>
                  <span>{user.email}</span>
                </div>
                <small
                  className={
                    user.email_verified
                      ? 'settings-status verified'
                      : 'settings-status'
                  }
                >
                  {user.email_verified ? 'Confirmado' : 'Pendente'}
                </small>
              </div>
              <Link href="/perfil" className="settings-profile-link">
                Editar perfil de ciclismo
                <ArrowRight size={16} />
              </Link>
            </section>

            <section className="settings-card settings-security-card">
              <span className="settings-icon">
                <KeyRound size={23} />
              </span>
              <h2>Segurança do acesso</h2>
              <p className="settings-card-intro">
                Altere sua senha e encerre o acesso em outros aparelhos. Ao
                trocar a senha, todos os outros dispositivos são desconectados.
              </p>
              <form
                method="post"
                className="settings-form"
                onSubmit={changePassword}
              >
                <label htmlFor="current-password">Senha atual</label>
                <input
                  id="current-password"
                  type="password"
                  value={currentPassword}
                  onChange={(event) => setCurrentPassword(event.target.value)}
                  autoComplete="current-password"
                  required
                />
                <label htmlFor="new-password">Nova senha</label>
                <input
                  id="new-password"
                  type="password"
                  value={newPassword}
                  onChange={(event) => setNewPassword(event.target.value)}
                  autoComplete="new-password"
                  minLength={10}
                  maxLength={72}
                  required
                />
                <label htmlFor="new-password-confirmation">
                  Confirmar nova senha
                </label>
                <input
                  id="new-password-confirmation"
                  type="password"
                  value={newPasswordConfirmation}
                  onChange={(event) =>
                    setNewPasswordConfirmation(event.target.value)
                  }
                  autoComplete="new-password"
                  minLength={10}
                  maxLength={72}
                  required
                />
                <small>
                  Use de 10 a 72 caracteres, diferente da senha atual.
                </small>
                {securityError && (
                  <p className="form-error" role="alert">
                    {securityError}
                  </p>
                )}
                {securityMessage && (
                  <p className="form-notice" role="status">
                    {securityMessage}
                  </p>
                )}
                <button
                  type="submit"
                  className="settings-primary-button"
                  disabled={
                    securityBusy ||
                    !currentPassword ||
                    !newPassword ||
                    !newPasswordConfirmation
                  }
                >
                  {securityBusy ? (
                    <>
                      <LoaderCircle className="spin" size={16} /> Salvando…
                    </>
                  ) : (
                    <>
                      <KeyRound size={16} /> Alterar senha
                    </>
                  )}
                </button>
              </form>
              <button
                type="button"
                className="settings-secondary-button"
                onClick={logoutOthers}
                disabled={securityBusy}
              >
                <LogOut size={16} /> Sair dos outros dispositivos
              </button>
            </section>
          </div>

          <div className="settings-column">
            <section className="settings-card settings-danger-card">
              <span className="settings-danger-icon">
                <Trash2 size={22} />
              </span>
              <h2>Encerrar conta</h2>
              <p className="settings-danger-warning">
                <CircleAlert size={17} />
                <span>
                  <strong>Esta ação é permanente.</strong> Todos os seus dados
                  serão apagados do banco de dados: perfil, planos, treinos,
                  feedbacks, avaliações, recuperações e sessões. Cópias de
                  segurança podem permanecer apenas até o prazo de retenção
                  operacional.
                </span>
              </p>
              <p className="settings-card-intro">
                Para evitar um encerramento acidental, confirme sua senha atual
                e digite a frase abaixo.
              </p>
              <form
                method="post"
                className="settings-delete-form"
                onSubmit={deleteAccount}
              >
                <label htmlFor="account-delete-password">Senha atual</label>
                <input
                  id="account-delete-password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete="current-password"
                  required
                />
                <label htmlFor="account-delete-confirmation">Confirmação</label>
                <input
                  id="account-delete-confirmation"
                  type="text"
                  value={confirmation}
                  onChange={(event) => setConfirmation(event.target.value)}
                  placeholder={DELETE_CONFIRMATION}
                  autoComplete="off"
                  required
                  aria-describedby="account-delete-help"
                />
                <small id="account-delete-help">
                  Digite exatamente: <strong>{DELETE_CONFIRMATION}</strong>
                </small>
                {error && (
                  <p className="form-error" role="alert">
                    {error}
                  </p>
                )}
                <button
                  type="submit"
                  className="settings-delete-button"
                  disabled={
                    deleting ||
                    password.length === 0 ||
                    confirmation.trim().toUpperCase() !== DELETE_CONFIRMATION
                  }
                >
                  {deleting ? (
                    <>
                      <LoaderCircle className="spin" size={16} /> Apagando
                      dados…
                    </>
                  ) : (
                    <>
                      <Trash2 size={16} /> Encerrar minha conta
                    </>
                  )}
                </button>
              </form>
            </section>
          </div>
        </div>
      </section>
    </main>
  );
}
