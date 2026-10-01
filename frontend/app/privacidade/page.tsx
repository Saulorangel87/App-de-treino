import type { Metadata } from 'next';
import { LegalPage } from '@/components/legal-document';
import { PRIVACY_POLICY } from '@/lib/legal';

export const metadata: Metadata = {
  title: 'Política de Privacidade — Cadência',
  description:
    'Quais dados o Cadência guarda, para que usa, com quem compartilha e como você os controla.',
};

export default function PrivacyPage() {
  return (
    <LegalPage
      document={PRIVACY_POLICY}
      other={{ href: '/termos', label: 'Termos de Uso' }}
    />
  );
}
