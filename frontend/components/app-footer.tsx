'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Code2, ContactRound, Download, Mail } from 'lucide-react';
import { APP_VERSION } from '@/lib/release';
import { defineMessages } from '@/lib/i18n';
import { LanguageSwitcher } from './language-switcher';
import { useMessages } from './locale-provider';

const messages = defineMessages({
  pt: {
    developedBy: 'DESENVOLVIDO POR ',
    legal: 'Documentos legais',
    privacy: 'Privacidade',
    terms: 'Termos de uso',
    install: 'Instalar app',
    installed: 'App instalado',
    linkedin: 'LinkedIn de Saulo Rangel',
    github: 'GitHub de Saulo Rangel',
    email: 'Enviar e-mail para Saulo Rangel',
  },
  en: {
    developedBy: 'DEVELOPED BY ',
    legal: 'Legal documents',
    privacy: 'Privacy',
    terms: 'Terms of use',
    install: 'Install app',
    installed: 'App installed',
    linkedin: "Saulo Rangel's LinkedIn",
    github: "Saulo Rangel's GitHub",
    email: 'Email Saulo Rangel',
  },
});

type InstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
};

export function AppFooter() {
  const [installPrompt, setInstallPrompt] = useState<InstallPromptEvent | null>(null);
  const [installed, setInstalled] = useState(false);
  const t = useMessages(messages);

  useEffect(() => {
    if ('serviceWorker' in navigator) {
      if (import.meta.env.PROD) {
        navigator.serviceWorker.register('/sw.js').catch(() => undefined);
      } else {
        // A preview PWA may have registered a worker for localhost earlier.
        // Development must always use the current Vite modules, not that cache.
        navigator.serviceWorker
          .getRegistrations()
          .then((registrations) =>
            Promise.all(
              registrations.map((registration) => registration.unregister()),
            ),
          )
          .catch(() => undefined);

        if ('caches' in window) {
          window.caches
            .keys()
            .then((keys) =>
              Promise.all(
                keys
                  .filter((key) => key.startsWith('cadencia-static-'))
                  .map((key) => window.caches.delete(key)),
              ),
            )
            .catch(() => undefined);
        }
      }
    }

    const standalone = window.matchMedia('(display-mode: standalone)').matches || Boolean((navigator as Navigator & { standalone?: boolean }).standalone);
    const rememberedInstall = window.localStorage.getItem('cadencia:pwa-installed') === 'true';
    queueMicrotask(() => setInstalled(standalone || rememberedInstall));

    const captureInstallPrompt = (event: Event) => {
      event.preventDefault();
      window.localStorage.removeItem('cadencia:pwa-installed');
      setInstalled(false);
      setInstallPrompt(event as InstallPromptEvent);
    };
    const markInstalled = () => {
      window.localStorage.setItem('cadencia:pwa-installed', 'true');
      setInstalled(true);
      setInstallPrompt(null);
    };

    window.addEventListener('beforeinstallprompt', captureInstallPrompt);
    window.addEventListener('appinstalled', markInstalled);
    return () => {
      window.removeEventListener('beforeinstallprompt', captureInstallPrompt);
      window.removeEventListener('appinstalled', markInstalled);
    };
  }, []);

  async function installApp() {
    if (!installPrompt) return;
    await installPrompt.prompt();
    const choice = await installPrompt.userChoice;
    if (choice.outcome === 'accepted') {
      window.localStorage.setItem('cadencia:pwa-installed', 'true');
      setInstalled(true);
      setInstallPrompt(null);
    }
  }

  return (
    <footer className="site-footer">
      <p>
        © 2026 <span className="footer-credit">{t.developedBy}</span>SAULO RANGEL{' '}
        <span className="footer-version">— V{APP_VERSION}</span>
      </p>
      <nav className="legal-links" aria-label={t.legal}>
        <Link href="/privacidade">{t.privacy}</Link>
        <Link href="/termos">{t.terms}</Link>
      </nav>
      <LanguageSwitcher className="footer-language" compact />
      <div className="footer-actions">
        {installPrompt && !installed && <button type="button" className="install-app" onClick={installApp}><Download size={14} />{t.install}</button>}
        {installed && <span className="installed-label"><span />{t.installed}</span>}
        <div className="footer-icons">
          <a className="footer-icon" href="https://www.linkedin.com/in/saulorangel87" target="_blank" rel="noreferrer" aria-label={t.linkedin}><ContactRound size={15} /></a>
          <a className="footer-icon" href="https://github.com/Saulorangel87" target="_blank" rel="noreferrer" aria-label={t.github}><Code2 size={16} /></a>
          <a className="footer-icon" href="mailto:sauloleonardo1987@gmail.com" aria-label={t.email}><Mail size={16} /></a>
        </div>
      </div>
    </footer>
  );
}
