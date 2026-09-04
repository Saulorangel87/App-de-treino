'use client';

import { useEffect, useState } from 'react';
import { Check, Sparkles, X } from 'lucide-react';
import { apiRequest } from '@/lib/api';
import { APP_VERSION, UPDATE_NOTES } from '@/lib/release';

const publicPaths = new Set([
  '/entrar',
  '/esqueci-minha-senha',
  '/redefinir-senha',
  '/verificar-email',
]);

export function UpdateNotice() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (publicPaths.has(window.location.pathname)) return;

    let cancelled = false;
    apiRequest<{ user: { id: string } }>('/v1/me')
      .then(({ user }) => {
        try {
          const storageKey = `cadencia:update-notice:${user.id}:${APP_VERSION}`;
          if (window.localStorage.getItem(storageKey)) return;
          window.localStorage.setItem(storageKey, 'seen');
          if (!cancelled) setVisible(true);
        } catch {
          // Storage may be unavailable in a private or restricted browser.
        }
      })
      .catch(() => undefined);

    return () => {
      cancelled = true;
    };
  }, []);

  if (!visible) return null;

  return (
    <dialog open className="modal-backdrop update-notice-backdrop" aria-labelledby="update-notice-title">
      <section className="update-notice">
        <button
          type="button"
          className="update-notice-close"
          onClick={() => setVisible(false)}
          aria-label="Fechar novidades"
        >
          <X size={18} />
        </button>
        <div className="update-notice-icon" aria-hidden="true">
          <Sparkles size={21} />
        </div>
        <span className="update-notice-kicker">NOVIDADES · V{APP_VERSION}</span>
        <h2 id="update-notice-title">O Cadência ganhou melhorias.</h2>
        <p className="update-notice-intro">
          Veja o que mudou para deixar seu planejamento mais claro e acompanhar melhor a sua rotina.
        </p>
        <ul className="update-notice-list">
          {UPDATE_NOTES.map((note) => (
            <li key={note.title}>
              <span><Check size={14} /></span>
              <div>
                <strong>{note.title}</strong>
                <p>{note.description}</p>
              </div>
            </li>
          ))}
        </ul>
        <button type="button" className="update-notice-action" onClick={() => setVisible(false)}>
          Entendi, continuar
        </button>
      </section>
    </dialog>
  );
}
