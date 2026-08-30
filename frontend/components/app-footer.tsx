'use client';

import { useEffect, useState } from 'react';
import { Code2, ContactRound, Download, Mail } from 'lucide-react';

type InstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
};

export function AppFooter() {
  const [installPrompt, setInstallPrompt] = useState<InstallPromptEvent | null>(null);
  const [installed, setInstalled] = useState(false);

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
      <p>© 2026 DESENVOLVIDO POR SAULO RANGEL <span>— V0.4.0</span></p>
      <div className="footer-actions">
        {installPrompt && !installed && <button type="button" className="install-app" onClick={installApp}><Download size={14} />Instalar app</button>}
        {installed && <span className="installed-label"><span />App instalado</span>}
        <a className="footer-icon" href="https://www.linkedin.com/in/saulorangel87" target="_blank" rel="noreferrer" aria-label="LinkedIn de Saulo Rangel"><ContactRound size={15} /></a>
        <a className="footer-icon" href="https://github.com/Saulorangel87" target="_blank" rel="noreferrer" aria-label="GitHub de Saulo Rangel"><Code2 size={16} /></a>
        <a className="footer-icon" href="mailto:sauloleonardo1987@gmail.com" aria-label="Enviar e-mail para Saulo Rangel"><Mail size={16} /></a>
      </div>
    </footer>
  );
}
