import Link from 'next/link';
import { Brand } from '@/components/brand';
import { LanguageSwitcher } from '@/components/language-switcher';
import type { Locale } from '@/lib/i18n';
import { LEGAL_VERSION, type LegalDocument } from '@/lib/legal';
import { legalUpdatedLabel } from '@/lib/legal-i18n';

const text = {
  pt: {
    back: 'Voltar ao app',
    meta: (version: string, date: string) => `Versão ${version} · atualizada em ${date}`,
    index: 'Nesta página',
    alsoRead: 'Leia também:',
  },
  en: {
    back: 'Back to the app',
    meta: (version: string, date: string) => `Version ${version} · updated on ${date}`,
    index: 'On this page',
    alsoRead: 'Also read:',
  },
};

export function LegalPage({
  locale,
  document,
  other,
}: {
  locale: Locale;
  document: LegalDocument;
  other: { href: string; label: string };
}) {
  const t = text[locale];
  return (
    <main className="legal-shell">
      <header className="legal-top">
        <Brand />
        <Link href="/" className="legal-back">
          {t.back}
        </Link>
      </header>
      <article className="legal-content">
        <p className="legal-kicker">{document.kicker}</p>
        <h1>{document.title}</h1>
        <p className="legal-meta">{t.meta(LEGAL_VERSION, legalUpdatedLabel(locale))}</p>
        {locale === 'en' && (
          <p className="legal-translation-note" role="note">
            This English version is a translation provided for convenience. The Portuguese version is the one that
            legally applies and prevails in case of any difference between them.
          </p>
        )}
        <LanguageSwitcher className="legal-language" />
        <p className="legal-intro">{document.intro}</p>
        <nav className="legal-index" aria-label={t.index}>
          {document.sections.map((section) => (
            <a key={section.id} href={`#${section.id}`}>
              {section.title}
            </a>
          ))}
        </nav>
        {document.sections.map((section) => (
          <section key={section.id} id={section.id} className="legal-section">
            <h2>{section.title}</h2>
            {section.paragraphs?.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
            {section.items && (
              <ul>
                {section.items.map((item) => <li key={item}>{item}</li>)}
              </ul>
            )}
            {section.after?.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
          </section>
        ))}
        <p className="legal-cross">
          {t.alsoRead} <Link href={other.href}>{other.label}</Link>
        </p>
      </article>
    </main>
  );
}
