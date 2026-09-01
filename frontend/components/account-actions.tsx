'use client';

import { useState } from 'react';
import { LoaderCircle, LogOut } from 'lucide-react';
import { apiRequest } from '@/lib/api';

export function LogoutButton({ compact = false }: { compact?: boolean }) {
  const [leaving, setLeaving] = useState(false);
  const [error, setError] = useState('');

  async function logout() {
    setLeaving(true);
    setError('');
    try {
      await apiRequest<void>('/v1/auth/logout', { method: 'POST' });
      window.location.replace('/entrar');
    } catch (caught) {
      setError(
        caught instanceof Error ? caught.message : 'Não foi possível sair.',
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
        aria-label={leaving ? 'Encerrando sessão' : 'Sair da conta'}
      >
        {leaving ? (
          <LoaderCircle className="spin" size={16} />
        ) : (
          <LogOut size={16} />
        )}
        <span>{leaving ? 'Saindo…' : 'Sair'}</span>
      </button>
      {error && <small role="alert">{error}</small>}
    </div>
  );
}

export function AccountActions({
  label,
  name,
}: {
  label: string;
  name?: string;
}) {
  return (
    <div className="account-actions">
      <div className="account-identity">
        <small>{label}</small>
        <strong>{name}</strong>
      </div>
      <LogoutButton />
    </div>
  );
}
