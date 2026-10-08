import type { Metadata, Viewport } from 'next';
import { Anybody, Schibsted_Grotesk } from 'next/font/google';
import { AppFooter } from '@/components/app-footer';
import { LocaleProvider } from '@/components/locale-provider';
import { TermsGate } from '@/components/terms-gate';
import { UpdateNotice } from '@/components/update-notice';
import { HTML_LANG, type Locale } from '@/lib/i18n';
import { requestLocale } from '@/lib/server-locale';
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
import './styles/zones.css';

// Anybody tem eixo de largura variável: títulos largos de carta topográfica e
// números estreitos usam a mesma família.
const display = Anybody({
  variable: '--font-display',
  subsets: ['latin'],
  axes: ['wdth'],
});
const body = Schibsted_Grotesk({ variable: '--font-body', subsets: ['latin'] });

const metadataText = {
  pt: {
    title: 'Cadência — Treino inteligente de ciclismo',
    description: 'Planejamento de ciclismo baseado no seu perfil, recuperação e evolução.',
    shareDescription: 'Um plano que combina seu perfil, recuperação e evolução contínua.',
    imageAlt: 'Cadência: treino de ciclismo que se adapta a você.',
  },
  en: {
    title: 'Cadência — Smart cycling training',
    description: 'Cycling plans built on your profile, recovery and progress.',
    shareDescription: 'A plan that combines your profile, recovery and steady progress.',
    imageAlt: 'Cadência: cycling training that adapts to you.',
  },
} satisfies Record<Locale, Record<string, string>>;

export async function generateMetadata(): Promise<Metadata> {
  const text = metadataText[await requestLocale()];
  return {
    title: text.title,
    description: text.description,
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
      title: text.title,
      description: text.shareDescription,
      images: [{ url: '/og.png', width: 1536, height: 1024, alt: text.imageAlt }],
    },
    twitter: {
      card: 'summary_large_image',
      title: text.title,
      description: text.shareDescription,
      images: ['/og.png'],
    },
  };
}

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

export default async function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const locale = await requestLocale();
  return (
    <html lang={HTML_LANG[locale]}>
      {!import.meta.env.PROD && (
        <head>
          <script dangerouslySetInnerHTML={{ __html: devPwaResetScript }} />
        </head>
      )}
      <body className={`${display.variable} ${body.variable}`}>
        <LocaleProvider locale={locale}>
          {children}
          <UpdateNotice />
          <TermsGate />
          <AppFooter />
        </LocaleProvider>
      </body>
    </html>
  );
}
