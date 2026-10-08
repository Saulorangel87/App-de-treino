'use client';

import { useEffect, useState } from 'react';
import { LoaderCircle, ScrollText } from 'lucide-react';
import { PUBLIC_PATHS, apiErrorMessage, apiRequest } from '@/lib/api';
import { defineMessages } from '@/lib/i18n';
import { LEGAL_VERSION } from '@/lib/legal';
import { legalUpdatedLabel } from '@/lib/legal-i18n';
import { useLocale, useMessages } from './locale-provider';
import { useScrollLock } from './use-scroll-lock';

// Nestas telas o aceite não bloqueia: os textos precisam poder ser lidos e o
// atleta precisa conseguir exportar ou apagar os dados sem aceitar nada.
const exemptPaths = new Set([...PUBLIC_PATHS, '/configuracoes']);

type Me = { legal?: { terms_version: string; accepted: boolean } };

const messages = defineMessages({
  pt: {
    failed: 'Não foi possível registrar o aceite agora.',
    kicker: 'TERMOS',
    title: 'Antes de continuar.',
    intro:
      'O Cadência passou a registrar o aceite dos Termos de Uso e da Política de Privacidade, que explica quais dados guardamos, inclusive os de saúde que você informa, e como você os controla.',
    read: 'Leia a',
    privacy: 'Política de Privacidade',
    and: 'e os',
    terms: 'Termos de Uso',
    newTab: '(abrem em outra aba).',
    check:
      'Li e aceito os Termos de Uso e a Política de Privacidade, inclusive o tratamento dos meus dados de saúde para montar meus treinos.',
    leave: 'Prefiro sair',
    saving: 'Registrando…',
    accept: 'Aceitar e continuar',
  },
  en: {
    failed: 'Your acceptance could not be recorded right now.',
    kicker: 'TERMS',
    title: 'Before you continue.',
    intro:
      'Cadência now records your acceptance of the Terms of Use and the Privacy Policy, which explains what data we keep, including the health data you enter, and how you control it.',
    read: 'Read the',
    privacy: 'Privacy Policy',
    and: 'and the',
    terms: 'Terms of Use',
    newTab: '(they open in a new tab). The Portuguese version is the one that legally applies.',
    check:
      'I have read and accept the Terms of Use and the Privacy Policy, including the processing of my health data to build my workouts.',
    leave: "I'd rather leave",
    saving: 'Saving…',
    accept: 'Accept and continue',
  },
});

// Pede o aceite dos Termos de Uso e da Política de Privacidade a quem ainda não
// aceitou a versão atual: contas criadas antes do registro do aceite e qualquer
// mudança futura dos textos.
export function TermsGate() {
  const [visible, setVisible] = useState(false);
  const [checked, setChecked] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const locale = useLocale();
  const t = useMessages(messages);

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
      setError(apiErrorMessage(caught, t.failed));
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
    <dialog open className="modal-backdrop terms-gate-backdrop" aria-labelledby="terms-gate-title">
      <section className="update-notice terms-gate">
        <div className="update-notice-icon" aria-hidden="true">
          <ScrollText size={21} />
        </div>
        <span className="update-notice-kicker">
          {t.kicker} · {legalUpdatedLabel(locale).toUpperCase()}
        </span>
        <h2 id="terms-gate-title">{t.title}</h2>
        <p className="update-notice-intro">{t.intro}</p>
        <p className="terms-gate-links">
          {t.read}{' '}
          <a href="/privacidade" target="_blank" rel="noreferrer">
            {t.privacy}
          </a>{' '}
          {t.and}{' '}
          <a href="/termos" target="_blank" rel="noreferrer">
            {t.terms}
          </a>{' '}
          {t.newTab}
        </p>
        <label className="terms-check">
          <input type="checkbox" checked={checked} onChange={(event) => setChecked(event.target.checked)} />
          <span>{t.check}</span>
        </label>
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
        <div className="update-notice-actions">
          <button type="button" className="terms-gate-leave" onClick={leave}>
            {t.leave}
          </button>
          <button type="button" className="update-notice-action" onClick={accept} disabled={!checked || busy}>
            {busy ? (
              <>
                <LoaderCircle className="spin" size={16} /> {t.saving}
              </>
            ) : (
              t.accept
            )}
          </button>
        </div>
      </section>
    </dialog>
  );
}
