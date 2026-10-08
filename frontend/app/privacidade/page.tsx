import type { Metadata } from 'next';
import { LegalPage } from '@/components/legal-document';
import { privacyPolicy } from '@/lib/legal-i18n';
import { requestLocale } from '@/lib/server-locale';

const text = {
  pt: {
    title: 'Política de Privacidade — Cadência',
    description: 'Quais dados o Cadência guarda, para que usa, com quem compartilha e como você os controla.',
    other: 'Termos de Uso',
  },
  en: {
    title: 'Privacy Policy — Cadência',
    description: 'What data Cadência keeps, what it is used for, who it is shared with and how you control it.',
    other: 'Terms of Use',
  },
};

export async function generateMetadata(): Promise<Metadata> {
  const { title, description } = text[await requestLocale()];
  return { title, description };
}

export default async function PrivacyPage() {
  const locale = await requestLocale();
  return <LegalPage locale={locale} document={privacyPolicy(locale)} other={{ href: '/termos', label: text[locale].other }} />;
}
