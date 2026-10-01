import type { Metadata } from 'next';
import { LegalPage } from '@/components/legal-document';
import { TERMS_OF_USE } from '@/lib/legal';

export const metadata: Metadata = {
  title: 'Termos de Uso — Cadência',
  description: 'Regras de uso do Cadência, aplicativo de planejamento de treino de ciclismo.',
};

export default function TermsPage() {
  return (
    <LegalPage
      document={TERMS_OF_USE}
      other={{ href: '/privacidade', label: 'Política de Privacidade' }}
    />
  );
}
