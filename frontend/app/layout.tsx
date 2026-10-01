import type { Metadata, Viewport } from 'next';
import { Anybody, Schibsted_Grotesk } from 'next/font/google';
import { AppFooter } from '@/components/app-footer';
import { TermsGate } from '@/components/terms-gate';
import { UpdateNotice } from '@/components/update-notice';
import './globals.css';
import './styles/base.css';
import './styles/shell.css';
import './styles/route.css';
import './styles/dashboard.css';
import './styles/plan.css';
import './styles/session.css';
import './styles/activities.css';
import './styles/evolution.css';
import './styles/checkin.css';
import './styles/profile.css';
import './styles/settings.css';
import './styles/auth.css';
import './styles/legal.css';

// Anybody tem eixo de largura variável: títulos largos de carta topográfica e
// números estreitos usam a mesma família.
const display = Anybody({
  variable: '--font-display',
  subsets: ['latin'],
  axes: ['wdth'],
});
const body = Schibsted_Grotesk({ variable: '--font-body', subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Cadência — Treino inteligente de ciclismo',
  description:
    'Planejamento de ciclismo baseado no seu perfil, recuperação e evolução.',
  applicationName: 'Cadência',
  manifest: '/app.webmanifest',
  icons: {
    icon: [
      { url: '/favicon.svg', type: 'image/svg+xml' },
      { url: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
    ],
    apple: [
      {
        url: '/icons/apple-touch-icon.png',
        sizes: '180x180',
        type: 'image/png',
      },
    ],
  },
  appleWebApp: {
    capable: true,
    statusBarStyle: 'default',
    title: 'Cadência',
  },
  openGraph: {
    title: 'Cadência — Treino inteligente de ciclismo',
    description:
      'Um plano que combina seu perfil, recuperação e evolução contínua.',
    images: [
      {
        url: '/og.png',
        width: 1536,
        height: 1024,
        alt: 'Cadência: treino de ciclismo que se adapta a você.',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Cadência — Treino inteligente de ciclismo',
    description:
      'Um plano que combina seu perfil, recuperação e evolução contínua.',
    images: ['/og.png'],
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover',
  themeColor: '#eef1ea',
  colorScheme: 'light',
};

const devPwaResetScript = `
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.getRegistrations()
      .then((registrations) => Promise.all(registrations.map((registration) => registration.unregister())))
      .catch(() => undefined);
  }
  if ('caches' in window) {
    caches.keys()
      .then((keys) => Promise.all(
        keys
          .filter((key) => key.startsWith('cadencia-static-'))
          .map((key) => caches.delete(key)),
      ))
      .catch(() => undefined);
  }
`;

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="pt-BR">
      {!import.meta.env.PROD && (
        <head>
          <script dangerouslySetInnerHTML={{ __html: devPwaResetScript }} />
        </head>
      )}
      <body className={`${display.variable} ${body.variable}`}>
        {children}
        <UpdateNotice />
        <TermsGate />
        <AppFooter />
      </body>
    </html>
  );
}
