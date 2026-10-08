'use client';

// Escala de intensidade do app inteiro, emprestada da sinalização de trilhas:
// círculo verde (leve), quadrado azul (moderado) e losango preto (intenso).

import { defineMessages } from '@/lib/i18n';
import { zoneForRpe } from '@/lib/zones';
import { useMessages } from './locale-provider';

export type Intensity = 'easy' | 'moderate' | 'hard';

// Leve: zonas 1 e 2; moderado: zona 3; intenso: zonas 4 e 5.
export function intensityOf(rpe: number): Intensity {
  const zone = zoneForRpe(rpe).number;
  if (zone <= 2) return 'easy';
  if (zone === 3) return 'moderate';
  return 'hard';
}

export const intensityMessages = defineMessages({
  pt: {
    labels: { easy: 'Leve', moderate: 'Moderado', hard: 'Intenso' } as Record<Intensity, string>,
    ranges: { easy: 'Zonas 1–2', moderate: 'Zona 3', hard: 'Zonas 4–5' } as Record<Intensity, string>,
    rest: 'Descanso',
    intensity: (label: string) => `Intensidade ${label.toLowerCase()}`,
    legend: 'Legenda de intensidade',
  },
  en: {
    labels: { easy: 'Easy', moderate: 'Moderate', hard: 'Hard' },
    ranges: { easy: 'Zones 1–2', moderate: 'Zone 3', hard: 'Zones 4–5' },
    rest: 'Rest',
    intensity: (label: string) => `${label} intensity`,
    legend: 'Intensity legend',
  },
});

export function TrailSymbol({
  rpe,
  intensity,
  rest = false,
  label,
}: {
  rpe?: number;
  intensity?: Intensity;
  rest?: boolean;
  label?: string;
}) {
  const t = useMessages(intensityMessages);
  const level = rest ? 'rest' : intensity || intensityOf(rpe ?? 1);
  // label="" marca o símbolo como decorativo (o texto ao lado já descreve).
  const a11y =
    label === ''
      ? { 'aria-hidden': true as const }
      : {
          role: 'img',
          'aria-label':
            label || (level === 'rest' ? t.rest : t.intensity(t.labels[level])),
        };
  return <span className={`trail-sym trail-sym-${level}`} {...a11y} />;
}

export function TrailLegend() {
  const t = useMessages(intensityMessages);
  return (
    <ul className="trail-legend" aria-label={t.legend}>
      {(['easy', 'moderate', 'hard'] as const).map((level) => (
        <li key={level}>
          <TrailSymbol intensity={level} label="" />
          <span>
            {t.labels[level]} <small>{t.ranges[level]}</small>
          </span>
        </li>
      ))}
    </ul>
  );
}
