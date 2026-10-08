'use client';

import { useState } from 'react';
import { LoaderCircle, LogOut } from 'lucide-react';
import { apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';
import { useMessages } from './locale-provider';

const messages = defineMessages({
  pt: { failed: 'Não foi possível sair.', leavingLabel: 'Encerrando sessão', logoutLabel: 'Sair da conta', leaving: 'Saindo…', logout: 'Sair' },
  en: { failed: 'Could not sign out.', leavingLabel: 'Signing out', logoutLabel: 'Sign out of your account', leaving: 'Signing out…', logout: 'Sign out' },
});

export function LogoutButton({ compact = false }: { compact?: boolean }) {
  const [leaving, setLeaving] = useState(false);
  const [error, setError] = useState('');
  const t = useMessages(messages);

  async function logout() {
    setLeaving(true);
    setError('');
    try {
      await apiRequest<void>('/v1/auth/logout', { method: 'POST' });
      window.location.replace('/entrar');
    } catch (caught) {
      setError(
        caught instanceof Error ? caught.message : t.failed,
      );
      setLeaving(false);
    }
  }

  return (
    <div className="logout-wrap">
      <button
        type="button"
        className={`logout-button${compact ? ' compact' : ''}`}
        onClick={logout}
        disabled={leaving}
        aria-label={leaving ? t.leavingLabel : t.logoutLabel}
      >
        {leaving ? (
          <LoaderCircle className="spin" size={16} />
        ) : (
          <LogOut size={16} />
        )}
        <span>{leaving ? t.leaving : t.logout}</span>
      </button>
      {error && <small role="alert">{error}</small>}
    </div>
  );
}
