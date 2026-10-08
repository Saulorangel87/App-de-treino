'use client';

import { CalendarCheck2 } from 'lucide-react';
import { buildWeeklySummary, type WeeklySummaryInput } from '@/lib/weekly-summary';
import { useLocale } from './locale-provider';

/** Cartão "Sua semana" da tela de Evolução: leitura da última semana completa, por regras. */
export function WeeklySummaryCard({ summary, today }: { summary: WeeklySummaryInput; today: Date }) {
  const locale = useLocale();
  const weekly = buildWeeklySummary(summary, locale, today);
  if (!weekly) return null;
  return (
    <section className="evolution-card weekly-summary" aria-labelledby="weekly-summary-title">
      <div className="evolution-card-title">
        <div>
          <span>{weekly.range}</span>
          <h2 id="weekly-summary-title">{weekly.title}</h2>
        </div>
        <CalendarCheck2 size={20} aria-hidden="true" />
      </div>
      <ul>
        {weekly.lines.map((line) => (
          <li key={line.text} className={`tone-${line.tone}`}>
            {line.text}
          </li>
        ))}
      </ul>
      <p className="chart-caption">{weekly.footnote}</p>
    </section>
  );
}
