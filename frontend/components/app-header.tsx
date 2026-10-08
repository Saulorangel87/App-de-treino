'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Activity,
  CalendarDays,
  ChevronDown,
  Gauge,
  HeartPulse,
  LineChart,
  Map as MapIcon,
  MessageSquareHeart,
  Settings,
  Sparkles,
  UserRound,
} from 'lucide-react';
import { defineMessages } from '@/lib/i18n';
import { Brand } from './brand';
import { LogoutButton } from './account-actions';
import { useMessages } from './locale-provider';

const sections = [
  { href: '/', key: 'today', icon: MapIcon },
  { href: '/plano', key: 'plan', icon: CalendarDays },
  { href: '/atividades', key: 'activities', icon: Activity },
  { href: '/evolucao', key: 'progress', icon: LineChart },
  { href: '/recuperacao', key: 'checkin', icon: HeartPulse },
] as const;

const accountLinks = [
  { href: '/perfil', key: 'profile', icon: UserRound },
  { href: '/avaliacao', key: 'assessment', icon: Gauge },
  { href: '/novidades', key: 'news', icon: Sparkles },
  { href: '/feedback', key: 'feedback', icon: MessageSquareHeart },
  { href: '/configuracoes', key: 'settings', icon: Settings },
] as const;

const messages = defineMessages({
  pt: {
    links: {
      today: 'Hoje',
      plan: 'Plano',
      activities: 'Atividades',
      progress: 'Evolução',
      checkin: 'Check-in',
      profile: 'Perfil do atleta',
      assessment: 'Avaliação',
      news: 'Novidades',
      feedback: 'Enviar feedback',
      settings: 'Configurações',
    },
    sections: 'Seções',
    account: 'Conta',
    openMenu: 'Abrir menu da conta',
    athlete: 'Atleta',
  },
  en: {
    links: {
      today: 'Today',
      plan: 'Plan',
      activities: 'Activities',
      progress: 'Progress',
      checkin: 'Check-in',
      profile: 'Athlete profile',
      assessment: 'Assessment',
      news: "What's new",
      feedback: 'Send feedback',
      settings: 'Settings',
    },
    sections: 'Sections',
    account: 'Account',
    openMenu: 'Open account menu',
    athlete: 'Athlete',
  },
});

function isActive(pathname: string, href: string) {
  return href === '/' ? pathname === '/' : pathname === href || pathname.startsWith(`${href}/`);
}

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('');
}

export function AppHeader({ name }: { name?: string }) {
  const pathname = usePathname() || '/';
  const [menuOpen, setMenuOpen] = useState(false);
  const t = useMessages(messages);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    function close(event: MouseEvent | KeyboardEvent) {
      if (event instanceof KeyboardEvent) {
        if (event.key === 'Escape') setMenuOpen(false);
        return;
      }
      if (!menuRef.current?.contains(event.target as Node)) setMenuOpen(false);
    }
    document.addEventListener('mousedown', close);
    document.addEventListener('keydown', close);
    return () => {
      document.removeEventListener('mousedown', close);
      document.removeEventListener('keydown', close);
    };
  }, [menuOpen]);

  return (
    <>
      <header className="app-header">
        <Brand />
        <nav className="app-nav" aria-label={t.sections}>
          {sections.map(({ href, key }) => (
            <Link
              key={href}
              href={href}
              className={isActive(pathname, href) ? 'active' : undefined}
              aria-current={isActive(pathname, href) ? 'page' : undefined}
            >
              {t.links[key]}
            </Link>
          ))}
        </nav>
        <div className="app-menu" ref={menuRef}>
          <button
            type="button"
            className="app-menu-trigger"
            aria-expanded={menuOpen}
            aria-controls="app-menu-panel"
            onClick={() => setMenuOpen((open) => !open)}
          >
            <span className="app-menu-avatar" aria-hidden="true">
              {name ? initials(name) : <UserRound size={15} />}
            </span>
            <span className="app-menu-name">{name?.split(' ')[0] || t.account}</span>
            <ChevronDown size={15} aria-hidden="true" />
            <span className="sr-only">{t.openMenu}</span>
          </button>
          {menuOpen && (
            <div className="app-menu-panel" id="app-menu-panel">
              {name && (
                <p className="app-menu-identity">
                  <small>{t.athlete}</small>
                  <strong>{name}</strong>
                </p>
              )}
              <ul>
                {accountLinks.map(({ href, key, icon: Icon }) => (
                  <li key={href}>
                    <Link
                      href={href}
                      className={isActive(pathname, href) ? 'active' : undefined}
                      aria-current={isActive(pathname, href) ? 'page' : undefined}
                      onClick={() => setMenuOpen(false)}
                    >
                      <Icon size={17} aria-hidden="true" />
                      {t.links[key]}
                    </Link>
                  </li>
                ))}
              </ul>
              <LogoutButton />
            </div>
          )}
        </div>
      </header>
      <nav className="tab-bar" aria-label={t.sections}>
        {sections.map(({ href, key, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={isActive(pathname, href) ? 'active' : undefined}
            aria-current={isActive(pathname, href) ? 'page' : undefined}
          >
            <Icon size={20} aria-hidden="true" />
            <span>{t.links[key]}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
