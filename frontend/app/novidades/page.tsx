'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  Bike,
  Check,
  ChevronDown,
  LoaderCircle,
  Sparkles,
} from 'lucide-react';
import { AccountActions } from '@/components/account-actions';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { APP_VERSION, UPDATE_NOTES, type UpdateNote } from '@/lib/release';

type User = { display_name: string };

function groupNotesByVersion(notes: readonly UpdateNote[]) {
  const groups = new Map<string, UpdateNote[]>();
  for (const note of notes) {
    const versionNotes = groups.get(note.version) || [];
    versionNotes.push(note);
    groups.set(note.version, versionNotes);
  }
  return [...groups.entries()];
}

export default function NoveltiesPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const versions = useMemo(() => groupNotesByVersion(UPDATE_NOTES), []);

  useEffect(() => {
    apiRequest<{ user: User }>('/v1/me')
      .then(({ user: account }) => setUser(account))
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, 'Não foi possível carregar as novidades.'));
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando novidades…
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || 'Não foi possível carregar as novidades.'} />;

  return (
    <main className="updates-shell">
      <header className="profile-topbar">
        <Link href="/" className="account-brand dark">
          <span>
            <Bike size={19} />
          </span>
          cadência
        </Link>
        <AccountActions label="NOVIDADES" name={user.display_name} />
      </header>
      <section className="updates-content">
        <Link href="/" className="back-link">
          <ArrowLeft size={15} />
          Voltar ao painel
        </Link>
        <header className="updates-heading">
          <p>NOVIDADES DO CADÊNCIA</p>
          <h1>O que mudou.</h1>
          <span>
            Veja as melhorias do app e entenda como elas ajudam sua rotina de
            ciclismo.
          </span>
        </header>

        <section
          className="updates-current"
          aria-labelledby="updates-current-title"
        >
          <span className="updates-current-icon" aria-hidden="true">
            <Sparkles size={22} />
          </span>
          <div>
            <p>VERSÃO ATUAL</p>
            <h2 id="updates-current-title">V{APP_VERSION}</h2>
            <span>
              As novidades mais recentes aparecem primeiro. O histórico anterior
              fica organizado abaixo.
            </span>
          </div>
        </section>

        <section
          className="updates-history"
          aria-label="Histórico de atualizações"
        >
          {versions.map(([version, notes]) => {
            const current = version === APP_VERSION;
            return (
              <details
                className={`updates-version${current ? ' current' : ''}`}
                key={version}
                open={current}
              >
                <summary className="updates-version-summary">
                  <span>
                    <strong>Versão {version}</strong>
                    <small>
                      {current
                        ? 'Atual'
                        : `${notes.length} ${notes.length === 1 ? 'novidade' : 'novidades'}`}
                    </small>
                  </span>
                  <ChevronDown size={17} aria-hidden="true" />
                </summary>
                <ul className="updates-version-list">
                  {notes.map((note) => (
                    <li key={note.title}>
                      <span aria-hidden="true">
                        <Check size={14} />
                      </span>
                      <div>
                        <strong>{note.title}</strong>
                        <p>{note.description}</p>
                      </div>
                    </li>
                  ))}
                </ul>
              </details>
            );
          })}
        </section>
        <p className="updates-footnote">
          As notas descrevem o comportamento publicado de cada versão. O motor
          de prescrição continua seguindo as regras de segurança e as evidências
          documentadas.
        </p>
      </section>
    </main>
  );
}
