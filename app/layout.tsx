import type { Metadata } from 'next';
import { Geist, Newsreader } from 'next/font/google';
import './globals.css';

const geist = Geist({ variable: '--font-geist-sans', subsets: ['latin'] });
const newsreader = Newsreader({ variable: '--font-newsreader', subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Cadência — Treino inteligente de ciclismo',
  description: 'Planejamento de ciclismo baseado no seu perfil, recuperação e evolução.',
  openGraph: {
    title: 'Cadência — Treino inteligente de ciclismo',
    description: 'Um plano que combina seu perfil, recuperação e evolução contínua.',
    images: [{ url: '/og.png', width: 1536, height: 1024, alt: 'Cadência — Treino inteligente. Evolução contínua.' }],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Cadência — Treino inteligente de ciclismo',
    description: 'Um plano que combina seu perfil, recuperação e evolução contínua.',
    images: ['/og.png'],
  },
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="pt-BR"><body className={`${geist.variable} ${newsreader.variable}`}>{children}</body></html>;
}
