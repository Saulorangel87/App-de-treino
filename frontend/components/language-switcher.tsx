'use client';

import { Languages } from 'lucide-react';
import { LOCALES, LOCALE_NAMES } from '@/lib/i18n';
import { changeLocale, useLocale } from './locale-provider';

/** Escolha entre português e inglês; a página recarrega já no idioma novo. */
export function LanguageSwitcher({ className = '', compact = false }: { className?: string; compact?: boolean }) {
  const locale = useLocale();
  return (
    <div className={`language-switcher${compact ? ' compact' : ''} ${className}`.trim()} role="group" aria-label={locale === 'en' ? 'Language' : 'Idioma'}>
      {!compact && <Languages size={15} aria-hidden="true" />}
      {LOCALES.map((option) => (
        <button
          key={option}
          type="button"
          lang={option === 'en' ? 'en' : 'pt-BR'}
          aria-pressed={option === locale}
          aria-label={compact ? LOCALE_NAMES[option] : undefined}
          className={option === locale ? 'active' : undefined}
          onClick={() => option !== locale && changeLocale(option)}
        >
          {compact ? option.toUpperCase() : LOCALE_NAMES[option]}
        </button>
      ))}
    </div>
  );
}
