import Link from 'next/link';
import { Brand } from '@/components/brand';
import {
  LEGAL_UPDATED_LABEL,
  LEGAL_VERSION,
  type LegalDocument,
} from '@/lib/legal';

export function LegalPage({
  document,
  other,
}: {
  document: LegalDocument;
  other: { href: string; label: string };
}) {
  return (
    <main className="legal-shell">
      <header className="legal-top">
        <Brand />
        <Link href="/entrar" className="legal-back">
          Voltar ao app
        </Link>
      </header>
      <article className="legal-content">
        <p className="legal-kicker">{document.kicker}</p>
        <h1>{document.title}</h1>
        <p className="legal-meta">
          Versão {LEGAL_VERSION} · atualizada em {LEGAL_UPDATED_LABEL}
        </p>
        <p className="legal-intro">{document.intro}</p>
        <nav className="legal-index" aria-label="Nesta página">
          {document.sections.map((section) => (
            <a key={section.id} href={`#${section.id}`}>
              {section.title}
            </a>
          ))}
        </nav>
        {document.sections.map((section) => (
          <section key={section.id} id={section.id} className="legal-section">
            <h2>{section.title}</h2>
            {section.paragraphs?.map((text) => <p key={text}>{text}</p>)}
            {section.items && (
              <ul>
                {section.items.map((text) => <li key={text}>{text}</li>)}
              </ul>
            )}
            {section.after?.map((text) => <p key={text}>{text}</p>)}
          </section>
        ))}
        <p className="legal-cross">
          Leia também: <Link href={other.href}>{other.label}</Link>
        </p>
      </article>
    </main>
  );
}
