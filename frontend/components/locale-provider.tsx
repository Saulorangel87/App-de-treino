'use client';

import { createContext, useContext } from 'react';
import { DEFAULT_LOCALE, writeLocaleCookie, type Locale, type Messages } from '@/lib/i18n';

const LocaleContext = createContext<Locale>(DEFAULT_LOCALE);

/** Recebe o idioma lido do cookie pelo servidor, para a primeira renderização já sair certa. */
export function LocaleProvider({ locale, children }: { locale: Locale; children: React.ReactNode }) {
  return <LocaleContext.Provider value={locale}>{children}</LocaleContext.Provider>;
}

export function useLocale(): Locale {
  return useContext(LocaleContext);
}

/** Os textos do componente no idioma atual. */
export function useMessages<T>(messages: Messages<T>): T {
  return messages[useContext(LocaleContext)];
}

/**
 * Troca o idioma: grava o cookie e recarrega a página, para que telas,
 * títulos e textos vindos da API saiam todos no idioma novo de uma vez.
 */
export function changeLocale(locale: Locale) {
  writeLocaleCookie(locale);
  window.location.reload();
}
