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
import { Brand } from './brand';
import { LogoutButton } from './account-actions';

const sections = [
  { href: '/', label: 'Hoje', icon: MapIcon },
  { href: '/plano', label: 'Plano', icon: CalendarDays },
  { href: '/atividades', label: 'Atividades', icon: Activity },
  { href: '/evolucao', label: 'Evolução', icon: LineChart },
  { href: '/recuperacao', label: 'Check-in', icon: HeartPulse },
] as const;

const accountLinks = [
  { href: '/perfil', label: 'Perfil do atleta', icon: UserRound },
  { href: '/avaliacao', label: 'Avaliação', icon: Gauge },
  { href: '/novidades', label: 'Novidades', icon: Sparkles },
  { href: '/feedback', label: 'Enviar feedback', icon: MessageSquareHeart },
  { href: '/configuracoes', label: 'Configurações', icon: Settings },
] as const;

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
        <nav className="app-nav" aria-label="Seções">
          {sections.map(({ href, label }) => (
            <Link
              key={href}
              href={href}
              className={isActive(pathname, href) ? 'active' : undefined}
              aria-current={isActive(pathname, href) ? 'page' : undefined}
            >
              {label}
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
            <span className="app-menu-name">{name?.split(' ')[0] || 'Conta'}</span>
            <ChevronDown size={15} aria-hidden="true" />
            <span className="sr-only">Abrir menu da conta</span>
          </button>
          {menuOpen && (
            <div className="app-menu-panel" id="app-menu-panel">
              {name && (
                <p className="app-menu-identity">
                  <small>Atleta</small>
                  <strong>{name}</strong>
                </p>
              )}
              <ul>
                {accountLinks.map(({ href, label, icon: Icon }) => (
                  <li key={href}>
                    <Link
                      href={href}
                      className={isActive(pathname, href) ? 'active' : undefined}
                      aria-current={isActive(pathname, href) ? 'page' : undefined}
                      onClick={() => setMenuOpen(false)}
                    >
                      <Icon size={17} aria-hidden="true" />
                      {label}
                    </Link>
                  </li>
                ))}
              </ul>
              <LogoutButton />
            </div>
          )}
        </div>
      </header>
      <nav className="tab-bar" aria-label="Seções">
        {sections.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={isActive(pathname, href) ? 'active' : undefined}
            aria-current={isActive(pathname, href) ? 'page' : undefined}
          >
            <Icon size={20} aria-hidden="true" />
            <span>{label}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
