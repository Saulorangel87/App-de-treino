// Idiomas do Cadência: português (padrão) e inglês. A escolha fica num cookie
// para que o servidor já renderize a página no idioma certo, sem piscar em
// português. Os textos ficam junto de cada componente, num objeto com as duas
// línguas (defineMessages), e o inglês precisa ter exatamente as mesmas chaves.

export type Locale = 'pt' | 'en';

export const LOCALES: readonly Locale[] = ['pt', 'en'];
export const DEFAULT_LOCALE: Locale = 'pt';
export const LOCALE_COOKIE = 'cadencia_lang';

/** Nome de cada idioma escrito nele mesmo, como aparece no seletor. */
export const LOCALE_NAMES: Record<Locale, string> = { pt: 'Português', en: 'English' };

/** Valor do atributo lang do HTML e do cabeçalho Accept-Language enviado à API. */
export const HTML_LANG: Record<Locale, string> = { pt: 'pt-BR', en: 'en' };

/** Localidade usada pelo Intl para datas e números. */
export const INTL_LOCALE: Record<Locale, string> = { pt: 'pt-BR', en: 'en-US' };

export function parseLocale(value: string | null | undefined): Locale {
  return value === 'en' ? 'en' : DEFAULT_LOCALE;
}

export function localeFromCookieHeader(cookieHeader: string | null | undefined): Locale {
  const match = (cookieHeader || '').match(new RegExp(`(?:^|;\\s*)${LOCALE_COOKIE}=([^;]*)`));
  return parseLocale(match?.[1]);
}

/** Idioma atual no navegador; fora dele (renderização no servidor), o padrão. */
export function currentLocale(): Locale {
  if (typeof document === 'undefined') return DEFAULT_LOCALE;
  return localeFromCookieHeader(document.cookie);
}

/** Se o atleta já escolheu um idioma neste aparelho (há cookie gravado). */
export function hasLocaleCookie(): boolean {
  if (typeof document === 'undefined') return false;
  return new RegExp(`(?:^|;\\s*)${LOCALE_COOKIE}=(pt|en)(?:;|$)`).test(document.cookie);
}

export function writeLocaleCookie(locale: Locale) {
  const oneYear = 60 * 60 * 24 * 365;
  const secure = typeof location !== 'undefined' && location.protocol === 'https:' ? '; Secure' : '';
  document.cookie = `${LOCALE_COOKIE}=${locale}; Path=/; Max-Age=${oneYear}; SameSite=Lax${secure}`;
}

export type Messages<T> = { pt: T; en: T };

/** Textos de um componente nas duas línguas; o inglês segue a forma do português. */
export function defineMessages<T>(messages: { pt: T; en: NoInfer<T> }): Messages<T> {
  return messages;
}

/** Número com casas fixas no formato do idioma (vírgula em português, ponto em inglês). */
export function formatDecimal(value: number, digits: number, locale: Locale): string {
  return value.toLocaleString(INTL_LOCALE[locale], { minimumFractionDigits: digits, maximumFractionDigits: digits });
}

/** Número com até `digits` casas, sem zeros sobrando. */
export function formatNumber(value: number, locale: Locale, digits = 1): string {
  return value.toLocaleString(INTL_LOCALE[locale], { maximumFractionDigits: digits });
}
