import type { Metadata } from 'next';
import { LegalPage } from '@/components/legal-document';
import { termsOfUse } from '@/lib/legal-i18n';
import { requestLocale } from '@/lib/server-locale';

const text = {
  pt: {
    title: 'Termos de Uso — Cadência',
    description: 'Regras de uso do Cadência, aplicativo de planejamento de treino de ciclismo.',
    other: 'Política de Privacidade',
  },
  en: {
    title: 'Terms of Use — Cadência',
    description: 'Rules for using Cadência, a cycling training planning app.',
    other: 'Privacy Policy',
  },
};

export async function generateMetadata(): Promise<Metadata> {
  const { title, description } = text[await requestLocale()];
  return { title, description };
}

export default async function TermsPage() {
  const locale = await requestLocale();
  return <LegalPage locale={locale} document={termsOfUse(locale)} other={{ href: '/privacidade', label: text[locale].other }} />;
}
