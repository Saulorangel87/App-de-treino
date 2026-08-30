import type { Metadata, Viewport } from 'next';
import { Geist, Newsreader } from 'next/font/google';
import { AppFooter } from '@/components/app-footer';
import './globals.css';

const geist = Geist({ variable: '--font-geist-sans', subsets: ['latin'] });
const newsreader = Newsreader({ variable: '--font-newsreader', subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Cadência — Treino inteligente de ciclismo',
  description: 'Planejamento de ciclismo baseado no seu perfil, recuperação e evolução.',
  applicationName: 'Cadência',
  manifest: '/app.webmanifest',
  icons: {
    icon: [{ url: '/favicon.svg', type: 'image/svg+xml' }, { url: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' }],
    apple: [{ url: '/icons/apple-touch-icon.png', sizes: '180x180', type: 'image/png' }],
  },
  appleWebApp: { capable: true, statusBarStyle: 'black-translucent', title: 'Cadência' },
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

export const viewport: Viewport = { themeColor: '#102d24', colorScheme: 'light' };

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

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="pt-BR">
      {!import.meta.env.PROD && (
        <head>
          <script dangerouslySetInnerHTML={{ __html: devPwaResetScript }} />
        </head>
      )}
      <body className={`${geist.variable} ${newsreader.variable}`}>
        {children}
        <AppFooter />
      </body>
    </html>
  );
}
