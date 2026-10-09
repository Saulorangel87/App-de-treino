'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import {
  ArrowRight,
  CircleAlert,
  KeyRound,
  Languages,
  LoaderCircle,
  LogOut,
  Download,
  Mail,
  ShieldCheck,
  Trash2,
} from 'lucide-react';
import { AppHeader } from '@/components/app-header';
import { LanguageSwitcher } from '@/components/language-switcher';
import { ReferenceNumbersCard } from '@/components/reference-numbers-card';
import { useLocale, useMessages } from '@/components/locale-provider';
import { ApiError, apiDownload, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { defineMessages } from '@/lib/i18n';
import { LEGAL_CONTACT_EMAIL } from '@/lib/legal';

type User = {
  display_name: string;
  email: string;
  email_verified: boolean;
};

// A frase muda com o idioma; o backend aceita as duas.
const DELETE_CONFIRMATION = { pt: 'ENCERRAR CONTA', en: 'DELETE ACCOUNT' } as const;

const messages = defineMessages({
  pt: {
    loadFailed: 'Não foi possível carregar as configurações.',
    mismatch: 'A confirmação não coincide com a nova senha.',
    passwordFailed: 'Não foi possível alterar a senha.',
    sessionsEnded: (count: number) => `${count} outra(s) sessão(ões) encerrada(s).`,
    noSessions: 'Não havia outras sessões ativas.',
    sessionsFailed: 'Não foi possível encerrar as outras sessões.',
    exportFailed: 'Não foi possível exportar seus dados agora.',
    exportFilename: 'cadencia-dados',
    typeToConfirm: (phrase: string) => `Digite ${phrase} para confirmar o encerramento.`,
    deleteFailed: 'Não foi possível encerrar sua conta. Nenhum dado foi alterado.',
    loading: 'Carregando configurações…',
    kicker: 'CONTA E SEGURANÇA',
    title: 'Suas configurações.',
    intro: 'Gerencie o acesso à conta e encontre rapidamente os dados do seu perfil de ciclismo.',
    accountTitle: 'Dados da conta',
    accountIntro: 'Estas informações identificam seu acesso ao Cadência.',
    email: 'E-mail',
    verified: 'Confirmado',
    pending: 'Pendente',
    editProfile: 'Editar perfil de ciclismo',
    languageTitle: 'Idioma',
    languageIntro: 'Escolha o idioma do app. Os treinos, os avisos e os e-mails passam a vir neste idioma.',
    securityTitle: 'Segurança do acesso',
    securityIntro: 'Altere sua senha e encerre o acesso em outros aparelhos. Ao trocar a senha, todos os outros dispositivos são desconectados.',
    currentPassword: 'Senha atual',
    newPassword: 'Nova senha',
    confirmPassword: 'Confirmar nova senha',
    passwordHint: 'Use de 10 a 72 caracteres, diferente da senha atual.',
    saving: 'Salvando…',
    changePassword: 'Alterar senha',
    logoutOthers: 'Sair dos outros dispositivos',
    dataTitle: 'Seus dados',
    dataIntro: (email: string) =>
      `Baixe uma planilha (.xlsx) com os dados da sua conta, os treinos que você realizou e as atividades que importou. Ela não traz dados de saúde nem de acesso. Para uma cópia completa, escreva para ${email}.`,
    preparing: 'Preparando…',
    download: 'Baixar minha planilha',
    privacyBefore: 'Veja como tratamos seus dados na ',
    privacy: 'Política de Privacidade',
    privacyMiddle: ' e nos ',
    terms: 'Termos de Uso',
    deleteTitle: 'Encerrar conta',
    permanent: 'Esta ação é permanente.',
    permanentText:
      ' Todos os seus dados serão apagados do banco de dados: perfil, planos, treinos, feedbacks, avaliações, recuperações e sessões. Cópias de segurança podem permanecer apenas até o prazo de retenção operacional.',
    deleteIntro: 'Para evitar um encerramento acidental, confirme sua senha atual e digite a frase abaixo.',
    confirmation: 'Confirmação',
    typeExactly: 'Digite exatamente:',
    deleting: 'Apagando dados…',
    deleteAccount: 'Encerrar minha conta',
  },
  en: {
    loadFailed: 'Your settings could not be loaded.',
    mismatch: "The confirmation doesn't match the new password.",
    passwordFailed: 'The password could not be changed.',
    sessionsEnded: (count: number) => `${count} other session${count === 1 ? '' : 's'} ended.`,
    noSessions: 'There were no other active sessions.',
    sessionsFailed: 'The other sessions could not be ended.',
    exportFailed: 'Your data could not be exported right now.',
    exportFilename: 'cadencia-data',
    typeToConfirm: (phrase: string) => `Type ${phrase} to confirm the deletion.`,
    deleteFailed: 'Your account could not be deleted. No data was changed.',
    loading: 'Loading settings…',
    kicker: 'ACCOUNT AND SECURITY',
    title: 'Your settings.',
    intro: 'Manage access to your account and quickly find your cycling profile data.',
    accountTitle: 'Account details',
    accountIntro: 'This information identifies your access to Cadência.',
    email: 'Email',
    verified: 'Confirmed',
    pending: 'Pending',
    editProfile: 'Edit cycling profile',
    languageTitle: 'Language',
    languageIntro: 'Choose the app language. Workouts, notices and emails will come in this language.',
    securityTitle: 'Access security',
    securityIntro: 'Change your password and sign out of other devices. When you change the password, all other devices are signed out.',
    currentPassword: 'Current password',
    newPassword: 'New password',
    confirmPassword: 'Confirm new password',
    passwordHint: 'Use 10 to 72 characters, different from your current password.',
    saving: 'Saving…',
    changePassword: 'Change password',
    logoutOthers: 'Sign out of other devices',
    dataTitle: 'Your data',
    dataIntro: (email: string) =>
      `Download a spreadsheet (.xlsx) with your account data, the workouts you completed and the activities you imported. It does not include health or access data. For a full copy, write to ${email}.`,
    preparing: 'Preparing…',
    download: 'Download my spreadsheet',
    privacyBefore: 'See how we handle your data in the ',
    privacy: 'Privacy Policy',
    privacyMiddle: ' and the ',
    terms: 'Terms of Use',
    deleteTitle: 'Delete account',
    permanent: 'This action is permanent.',
    permanentText:
      ' All of your data will be erased from the database: profile, plans, workouts, feedback, assessments, recovery records and sessions. Backups may remain only until the operational retention period ends.',
    deleteIntro: 'To avoid an accidental deletion, confirm your current password and type the phrase below.',
    confirmation: 'Confirmation',
    typeExactly: 'Type exactly:',
    deleting: 'Erasing data…',
    deleteAccount: 'Delete my account',
  },
});

export default function SettingsPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const deletePhrase = DELETE_CONFIRMATION[locale];
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
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState('');

  useEffect(() => {
    apiRequest<{ user: User }>('/v1/me')
      .then(({ user: account }) => setUser(account))
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

  async function changePassword(event: { preventDefault: () => void }) {
    event.preventDefault();
    setSecurityError('');
    setSecurityMessage('');
    if (newPassword !== newPasswordConfirmation) {
      setSecurityError(t.mismatch);
      return;
    }
    setSecurityBusy(true);
    try {
      const result = await apiRequest<{ message: string }>('/v1/auth/change-password', {
        method: 'POST',
        body: JSON.stringify({
          current_password: currentPassword,
          new_password: newPassword,
        }),
      });
      setSecurityMessage(result.message);
      setCurrentPassword('');
      setNewPassword('');
      setNewPasswordConfirmation('');
    } catch (caught) {
      setSecurityError(apiErrorMessage(caught, t.passwordFailed));
    } finally {
      setSecurityBusy(false);
    }
  }

  async function logoutOthers() {
    setSecurityError('');
    setSecurityMessage('');
    setSecurityBusy(true);
    try {
      const result = await apiRequest<{ revoked_sessions: number }>('/v1/auth/logout-others', { method: 'POST' });
      setSecurityMessage(result.revoked_sessions > 0 ? t.sessionsEnded(result.revoked_sessions) : t.noSessions);
    } catch (caught) {
      setSecurityError(apiErrorMessage(caught, t.sessionsFailed));
    } finally {
      setSecurityBusy(false);
    }
  }

  async function exportData() {
    setExportError('');
    setExporting(true);
    try {
      const today = new Date().toISOString().slice(0, 10);
      await apiDownload('/v1/auth/account/export', `${t.exportFilename}-${today}.xlsx`);
    } catch (caught) {
      setExportError(apiErrorMessage(caught, t.exportFailed));
    } finally {
      setExporting(false);
    }
  }

  async function deleteAccount(event: { preventDefault: () => void }) {
    event.preventDefault();
    setError('');
    const typedConfirmation = confirmation.trim();
    if (typedConfirmation.toUpperCase() !== deletePhrase) {
      setError(t.typeToConfirm(deletePhrase));
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
      setError(apiErrorMessage(caught, t.deleteFailed));
      setDeleting(false);
    }
  }

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        {t.loading}
      </main>
    );
  }

  if (!user) {
    return <ApiErrorState message={error || t.loadFailed} />;
  }

  return (
    <main className="settings-shell">
      <AppHeader name={user.display_name} />
      <section className="settings-content">
        <header className="settings-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>{t.intro}</span>
        </header>

        <div className="settings-layout">
          <div className="settings-column">
            <section className="settings-card settings-account-card">
              <span className="settings-icon">
                <ShieldCheck size={23} />
              </span>
              <h2>{t.accountTitle}</h2>
              <p className="settings-card-intro">{t.accountIntro}</p>
              <div className="settings-account-row">
                <Mail size={18} />
                <div>
                  <strong>{t.email}</strong>
                  <span>{user.email}</span>
                </div>
                <small className={user.email_verified ? 'settings-status verified' : 'settings-status'}>
                  {user.email_verified ? t.verified : t.pending}
                </small>
              </div>
              <Link href="/perfil" className="settings-profile-link">
                {t.editProfile}
                <ArrowRight size={16} />
              </Link>
            </section>

            <section className="settings-card settings-language-card">
              <span className="settings-icon">
                <Languages size={23} />
              </span>
              <h2>{t.languageTitle}</h2>
              <p className="settings-card-intro">{t.languageIntro}</p>
              <LanguageSwitcher />
            </section>

            <ReferenceNumbersCard />

            <section className="settings-card settings-security-card">
              <span className="settings-icon">
                <KeyRound size={23} />
              </span>
              <h2>{t.securityTitle}</h2>
              <p className="settings-card-intro">{t.securityIntro}</p>
              <form method="post" className="settings-form" onSubmit={changePassword}>
                <label htmlFor="current-password">{t.currentPassword}</label>
                <input
                  id="current-password"
                  type="password"
                  value={currentPassword}
                  onChange={(event) => setCurrentPassword(event.target.value)}
                  autoComplete="current-password"
                  required
                />
                <label htmlFor="new-password">{t.newPassword}</label>
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
                <label htmlFor="new-password-confirmation">{t.confirmPassword}</label>
                <input
                  id="new-password-confirmation"
                  type="password"
                  value={newPasswordConfirmation}
                  onChange={(event) => setNewPasswordConfirmation(event.target.value)}
                  autoComplete="new-password"
                  minLength={10}
                  maxLength={72}
                  required
                />
                <small>{t.passwordHint}</small>
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
                  disabled={securityBusy || !currentPassword || !newPassword || !newPasswordConfirmation}
                >
                  {securityBusy ? (
                    <>
                      <LoaderCircle className="spin" size={16} /> {t.saving}
                    </>
                  ) : (
                    <>
                      <KeyRound size={16} /> {t.changePassword}
                    </>
                  )}
                </button>
              </form>
              <button type="button" className="settings-secondary-button" onClick={logoutOthers} disabled={securityBusy}>
                <LogOut size={16} /> {t.logoutOthers}
              </button>
            </section>

            <section className="settings-card settings-data-card">
              <span className="settings-icon">
                <Download size={23} />
              </span>
              <h2>{t.dataTitle}</h2>
              <p className="settings-card-intro">{t.dataIntro(LEGAL_CONTACT_EMAIL)}</p>
              {exportError && (
                <p className="form-error" role="alert">
                  {exportError}
                </p>
              )}
              <button type="button" className="settings-secondary-button" onClick={exportData} disabled={exporting}>
                {exporting ? (
                  <>
                    <LoaderCircle className="spin" size={16} /> {t.preparing}
                  </>
                ) : (
                  <>
                    <Download size={16} /> {t.download}
                  </>
                )}
              </button>
              <p className="settings-card-intro">
                {t.privacyBefore}
                <Link href="/privacidade">{t.privacy}</Link>
                {t.privacyMiddle}
                <Link href="/termos">{t.terms}</Link>.
              </p>
            </section>
          </div>

          <div className="settings-column">
            <section className="settings-card settings-danger-card">
              <span className="settings-danger-icon">
                <Trash2 size={22} />
              </span>
              <h2>{t.deleteTitle}</h2>
              <p className="settings-danger-warning">
                <CircleAlert size={17} />
                <span>
                  <strong>{t.permanent}</strong>
                  {t.permanentText}
                </span>
              </p>
              <p className="settings-card-intro">{t.deleteIntro}</p>
              <form method="post" className="settings-delete-form" onSubmit={deleteAccount}>
                <label htmlFor="account-delete-password">{t.currentPassword}</label>
                <input
                  id="account-delete-password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete="current-password"
                  required
                />
                <label htmlFor="account-delete-confirmation">{t.confirmation}</label>
                <input
                  id="account-delete-confirmation"
                  type="text"
                  value={confirmation}
                  onChange={(event) => setConfirmation(event.target.value)}
                  placeholder={deletePhrase}
                  autoComplete="off"
                  required
                  aria-describedby="account-delete-help"
                />
                <small id="account-delete-help">
                  {t.typeExactly} <strong>{deletePhrase}</strong>
                </small>
                {error && (
                  <p className="form-error" role="alert">
                    {error}
                  </p>
                )}
                <button
                  type="submit"
                  className="settings-delete-button"
                  disabled={deleting || password.length === 0 || confirmation.trim().toUpperCase() !== deletePhrase}
                >
                  {deleting ? (
                    <>
                      <LoaderCircle className="spin" size={16} /> {t.deleting}
                    </>
                  ) : (
                    <>
                      <Trash2 size={16} /> {t.deleteAccount}
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
