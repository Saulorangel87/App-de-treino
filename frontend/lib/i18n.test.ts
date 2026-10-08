import { afterEach, describe, expect, it, vi } from 'vitest';
import { currentLocale, formatDecimal, hasLocaleCookie, localeFromCookieHeader, parseLocale } from './i18n';

describe('idioma', () => {
  it('português é o padrão; só "en" troca para inglês', () => {
    expect(parseLocale(undefined)).toBe('pt');
    expect(parseLocale('')).toBe('pt');
    expect(parseLocale('fr')).toBe('pt');
    expect(parseLocale('pt')).toBe('pt');
    expect(parseLocale('en')).toBe('en');
  });

  it('lê o cookie entre outros cookies', () => {
    expect(localeFromCookieHeader('a=1; cadencia_lang=en; b=2')).toBe('en');
    expect(localeFromCookieHeader('cadencia_lang=pt')).toBe('pt');
    expect(localeFromCookieHeader('outro_cadencia_lang=en')).toBe('pt');
    expect(localeFromCookieHeader(null)).toBe('pt');
  });

  it('formata números no padrão de cada idioma', () => {
    expect(formatDecimal(9.5, 1, 'pt')).toBe('9,5');
    expect(formatDecimal(9.5, 1, 'en')).toBe('9.5');
    expect(formatDecimal(1.2, 2, 'pt')).toBe('1,20');
  });
});

describe('idioma no navegador', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('fora do navegador (servidor), fica no padrão', () => {
    expect(currentLocale()).toBe('pt');
    expect(hasLocaleCookie()).toBe(false);
  });

  it('sem cookie, fica em português e não há escolha gravada', () => {
    vi.stubGlobal('document', { cookie: 'sessao=abc' });
    expect(currentLocale()).toBe('pt');
    expect(hasLocaleCookie()).toBe(false);
  });

  it('com o cookie em inglês, o app fala inglês', () => {
    vi.stubGlobal('document', { cookie: 'sessao=abc; cadencia_lang=en' });
    expect(currentLocale()).toBe('en');
    expect(hasLocaleCookie()).toBe(true);
  });
});
