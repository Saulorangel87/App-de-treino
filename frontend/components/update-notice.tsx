'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Check, Sparkles, X } from 'lucide-react';
import { PUBLIC_PATHS, apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';
import { APP_VERSION, updateNotes } from '@/lib/release';
import { useLocale, useMessages } from './locale-provider';
import { useScrollLock } from './use-scroll-lock';

const publicPaths = new Set(PUBLIC_PATHS);

const messages = defineMessages({
  pt: {
    close: 'Fechar novidades',
    kicker: 'NOVIDADES',
    title: 'O Cadência ganhou melhorias.',
    intro: 'Veja o que mudou para deixar seu planejamento mais claro e acompanhar melhor a sua rotina.',
    history: 'Ver histórico completo',
    continue: 'Entendi, continuar',
  },
  en: {
    close: "Close what's new",
    kicker: "WHAT'S NEW",
    title: 'Cadência got better.',
    intro: 'See what changed to make your planning clearer and follow your routine more closely.',
    history: 'See the full history',
    continue: 'Got it, continue',
  },
});

export function UpdateNotice() {
  const [visible, setVisible] = useState(false);
  const [storageKey, setStorageKey] = useState<string | null>(null);
  const locale = useLocale();
  const t = useMessages(messages);

  useScrollLock(visible);

  useEffect(() => {
    if (publicPaths.has(window.location.pathname)) return;

    let cancelled = false;
    apiRequest<{ user: { id: string } }>('/v1/me')
      .then(({ user }) => {
        try {
          const storageKey = `cadencia:update-notice:${user.id}:${APP_VERSION}`;
          if (window.localStorage.getItem(storageKey)) return;
          if (!cancelled) {
            setStorageKey(storageKey);
            setVisible(true);
          }
        } catch {
          // Storage may be unavailable in a private or restricted browser.
        }
      })
      .catch(() => undefined);

    return () => {
      cancelled = true;
    };
  }, []);

  function dismiss() {
    if (storageKey) {
      try {
        window.localStorage.setItem(storageKey, 'seen');
      } catch {
        // Storage may be unavailable in a private or restricted browser.
      }
    }
    setVisible(false);
  }

  if (!visible) return null;

  const currentNotes = updateNotes(locale).filter((note) => note.version === APP_VERSION);

  return (
    <dialog open className="modal-backdrop update-notice-backdrop" aria-labelledby="update-notice-title">
      <section className="update-notice">
        <button type="button" className="update-notice-close" onClick={dismiss} aria-label={t.close}>
          <X size={18} />
        </button>
        <div className="update-notice-icon" aria-hidden="true">
          <Sparkles size={21} />
        </div>
        <span className="update-notice-kicker">
          {t.kicker} · V{APP_VERSION}
        </span>
        <h2 id="update-notice-title">{t.title}</h2>
        <p className="update-notice-intro">{t.intro}</p>
        <ul className="update-notice-list">
          {currentNotes.map((note) => (
            <li key={note.title}>
              <span>
                <Check size={14} />
              </span>
              <div>
                <strong>{note.title}</strong>
                <p>{note.description}</p>
              </div>
            </li>
          ))}
        </ul>
        <div className="update-notice-actions">
          <Link className="update-notice-history" href="/novidades" onClick={dismiss}>
            {t.history}
          </Link>
          <button type="button" className="update-notice-action" onClick={dismiss}>
            {t.continue}
          </button>
        </div>
      </section>
    </dialog>
  );
}
