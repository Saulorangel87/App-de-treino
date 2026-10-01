'use client';

import { useEffect, useState } from 'react';
import { LoaderCircle, ScrollText } from 'lucide-react';
import { PUBLIC_PATHS, apiErrorMessage, apiRequest } from '@/lib/api';
import { LEGAL_UPDATED_LABEL, LEGAL_VERSION } from '@/lib/legal';
import { useScrollLock } from './use-scroll-lock';

// Nestas telas o aceite não bloqueia: os textos precisam poder ser lidos e o
// atleta precisa conseguir exportar ou apagar os dados sem aceitar nada.
const exemptPaths = new Set([...PUBLIC_PATHS, '/configuracoes']);

type Me = { legal?: { terms_version: string; accepted: boolean } };

// Pede o aceite dos Termos de Uso e da Política de Privacidade a quem ainda não
// aceitou a versão atual: contas criadas antes do registro do aceite e qualquer
// mudança futura dos textos.
export function TermsGate() {
  const [visible, setVisible] = useState(false);
  const [checked, setChecked] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useScrollLock(visible);

  useEffect(() => {
    if (exemptPaths.has(window.location.pathname)) return;
    let cancelled = false;
    apiRequest<Me>('/v1/me')
      .then((me) => {
        if (!cancelled && me.legal && !me.legal.accepted) setVisible(true);
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);

  async function accept() {
    setBusy(true);
    setError('');
    try {
      await apiRequest<void>('/v1/auth/accept-terms', {
        method: 'POST',
        body: JSON.stringify({ terms_version: LEGAL_VERSION }),
      });
      setVisible(false);
    } catch (caught) {
      setError(
        apiErrorMessage(caught, 'Não foi possível registrar o aceite agora.'),
      );
    } finally {
      setBusy(false);
    }
  }

  async function leave() {
    try {
      await apiRequest<void>('/v1/auth/logout', { method: 'POST' });
    } finally {
      window.location.replace('/entrar');
    }
  }

  if (!visible) return null;

  return (
    <dialog
      open
      className="modal-backdrop terms-gate-backdrop"
      aria-labelledby="terms-gate-title"
    >
      <section className="update-notice terms-gate">
        <div className="update-notice-icon" aria-hidden="true">
          <ScrollText size={21} />
        </div>
        <span className="update-notice-kicker">
          TERMOS · {LEGAL_UPDATED_LABEL.toUpperCase()}
        </span>
        <h2 id="terms-gate-title">Antes de continuar.</h2>
        <p className="update-notice-intro">
          O Cadência passou a registrar o aceite dos Termos de Uso e da Política
          de Privacidade, que explica quais dados guardamos, inclusive os de
          saúde que você informa, e como você os controla.
        </p>
        <p className="terms-gate-links">
          Leia a{' '}
          <a href="/privacidade" target="_blank" rel="noreferrer">
            Política de Privacidade
          </a>{' '}
          e os{' '}
          <a href="/termos" target="_blank" rel="noreferrer">
            Termos de Uso
          </a>{' '}
          (abrem em outra aba).
        </p>
        <label className="terms-check">
          <input
            type="checkbox"
            checked={checked}
            onChange={(event) => setChecked(event.target.checked)}
          />
          <span>
            Li e aceito os Termos de Uso e a Política de Privacidade, inclusive
            o tratamento dos meus dados de saúde para montar meus treinos.
          </span>
        </label>
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
        <div className="update-notice-actions">
          <button type="button" className="terms-gate-leave" onClick={leave}>
            Prefiro sair
          </button>
          <button
            type="button"
            className="update-notice-action"
            onClick={accept}
            disabled={!checked || busy}
          >
            {busy ? (
              <>
                <LoaderCircle className="spin" size={16} /> Registrando…
              </>
            ) : (
              'Aceitar e continuar'
            )}
          </button>
        </div>
      </section>
    </dialog>
  );
}
