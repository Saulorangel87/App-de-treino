import { cookies } from 'next/headers';
import { LOCALE_COOKIE, parseLocale, type Locale } from './i18n';

/** Idioma da requisição atual, lido do cookie (só em componentes de servidor). */
export async function requestLocale(): Promise<Locale> {
  return parseLocale((await cookies()).get(LOCALE_COOKIE)?.value);
}
