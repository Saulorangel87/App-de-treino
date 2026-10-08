'use client';

import Link from 'next/link';
import { useLocale } from './locale-provider';

// Marca do Cadência: curvas de nível abertas formando um "C" em volta de um
// cume, o ponto de chegada marcado em vermelho como nas cartas de trilha.
export function BrandMark({ size = 28 }: { size?: number }) {
  return (
    <svg
      className="brand-mark"
      width={size}
      height={size}
      viewBox="0 0 32 32"
      aria-hidden="true"
      focusable="false"
    >
      <path
        d="M23.78 8.22A11 11 0 1 0 23.78 23.78"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.6"
        strokeLinecap="round"
      />
      <path
        d="M20.18 11.02A6.5 6.5 0 1 0 20.18 20.98"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.7"
        strokeLinecap="round"
      />
      <circle className="brand-mark-summit" cx="16" cy="16" r="2.6" />
    </svg>
  );
}

export function Brand({ href = '/' }: { href?: string }) {
  const label = useLocale() === 'en' ? 'Cadência, go to the dashboard' : 'Cadência, ir para o painel';
  return (
    <Link href={href} className="brand" aria-label={label}>
      <BrandMark />
      <span className="brand-name">Cadência</span>
    </Link>
  );
}
