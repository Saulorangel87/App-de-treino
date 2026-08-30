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
    if ('serviceWorker' in navigator && import.meta.env.PROD) {
      navigator.serviceWorker.register('/sw.js').catch(() => undefined);
    }

    const captureInstallPrompt = (event: Event) => {
      event.preventDefault();
      setInstallPrompt(event as InstallPromptEvent);
    };
    const markInstalled = () => {
      setInstalled(true);
      setInstallPrompt(null);
    };

    setInstalled(window.matchMedia('(display-mode: standalone)').matches);
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
    if (choice.outcome === 'accepted') setInstallPrompt(null);
  }

  return (
    <footer className="site-footer">
      <p>© 2026 DESENVOLVIDO POR SAULO RANGEL <span>— V0.3.0</span></p>
      <div className="footer-actions">
        {installPrompt && <button type="button" className="install-app" onClick={installApp}><Download size={14} />Instalar app</button>}
        {installed && <span className="installed-label"><span />App instalado</span>}
        <a className="footer-icon" href="https://www.linkedin.com/in/saulorangel87" target="_blank" rel="noreferrer" aria-label="LinkedIn de Saulo Rangel"><ContactRound size={15} /></a>
        <a className="footer-icon" href="https://github.com/Saulorangel87" target="_blank" rel="noreferrer" aria-label="GitHub de Saulo Rangel"><Code2 size={16} /></a>
        <a className="footer-icon" href="mailto:sauloleonardo1987@gmail.com" aria-label="Enviar e-mail para Saulo Rangel"><Mail size={16} /></a>
      </div>
    </footer>
  );
}
