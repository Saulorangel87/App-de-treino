'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  Check,
  ChevronDown,
  LoaderCircle,
  Sparkles,
} from 'lucide-react';
import { AppHeader } from '@/components/app-header';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { defineMessages } from '@/lib/i18n';
import { APP_VERSION, updateNotes, type UpdateNote } from '@/lib/release';

type User = { display_name: string };

const messages = defineMessages({
  pt: {
    loadFailed: 'Não foi possível carregar as novidades.',
    loading: 'Carregando novidades…',
    kicker: 'NOVIDADES DO CADÊNCIA',
    title: 'O que mudou.',
    intro: 'Veja as melhorias do app e entenda como elas ajudam sua rotina de ciclismo.',
    currentKicker: 'VERSÃO ATUAL',
    currentText: 'As novidades mais recentes aparecem primeiro. O histórico anterior fica organizado abaixo.',
    history: 'Histórico de atualizações',
    version: (version: string) => `Versão ${version}`,
    current: 'Atual',
    count: (count: number) => `${count} ${count === 1 ? 'novidade' : 'novidades'}`,
    footnote:
      'As notas descrevem o comportamento publicado de cada versão. O motor de prescrição continua seguindo as regras de segurança e as evidências documentadas.',
  },
  en: {
    loadFailed: "What's new could not be loaded.",
    loading: "Loading what's new…",
    kicker: "WHAT'S NEW IN CADÊNCIA",
    title: 'What changed.',
    intro: 'See the improvements to the app and how they help your cycling routine.',
    currentKicker: 'CURRENT VERSION',
    currentText: 'The latest updates come first. The earlier history is organized below.',
    history: 'Update history',
    version: (version: string) => `Version ${version}`,
    current: 'Current',
    count: (count: number) => `${count} ${count === 1 ? 'update' : 'updates'}`,
    footnote:
      'The notes describe the published behavior of each version. The prescription engine keeps following the documented safety rules and evidence.',
  },
});

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
  const locale = useLocale();
  const t = useMessages(messages);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const versions = useMemo(() => groupNotesByVersion(updateNotes(locale)), [locale]);

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

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        {t.loading}
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  return (
    <main className="updates-shell">
      <AppHeader name={user.display_name} />
      <section className="updates-content">
        <header className="updates-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>{t.intro}</span>
        </header>

        <section className="updates-current" aria-labelledby="updates-current-title">
          <span className="updates-current-icon" aria-hidden="true">
            <Sparkles size={22} />
          </span>
          <div>
            <p>{t.currentKicker}</p>
            <h2 id="updates-current-title">V{APP_VERSION}</h2>
            <span>{t.currentText}</span>
          </div>
        </section>

        <section className="updates-history" aria-label={t.history}>
          {versions.map(([version, notes]) => {
            const current = version === APP_VERSION;
            return (
              <details className={`updates-version${current ? ' current' : ''}`} key={version} open={current}>
                <summary className="updates-version-summary">
                  <span>
                    <strong>{t.version(version)}</strong>
                    <small>{current ? t.current : t.count(notes.length)}</small>
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
        <p className="updates-footnote">{t.footnote}</p>
      </section>
    </main>
  );
}
