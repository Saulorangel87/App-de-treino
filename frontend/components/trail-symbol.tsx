// Escala de intensidade do app inteiro, emprestada da sinalização de trilhas:
// círculo verde (leve), quadrado azul (moderado) e losango preto (intenso).

import { zoneForRpe } from '@/lib/zones';

export type Intensity = 'easy' | 'moderate' | 'hard';

// Leve: zonas 1 e 2; moderado: zona 3; intenso: zonas 4 e 5.
export function intensityOf(rpe: number): Intensity {
  const zone = zoneForRpe(rpe).number;
  if (zone <= 2) return 'easy';
  if (zone === 3) return 'moderate';
  return 'hard';
}

export const intensityLabels: Record<Intensity, string> = {
  easy: 'Leve',
  moderate: 'Moderado',
  hard: 'Intenso',
};

export const intensityRanges: Record<Intensity, string> = {
  easy: 'Zonas 1–2',
  moderate: 'Zona 3',
  hard: 'Zonas 4–5',
};

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
  const level = rest ? 'rest' : intensity || intensityOf(rpe ?? 1);
  // label="" marca o símbolo como decorativo (o texto ao lado já descreve).
  const a11y =
    label === ''
      ? { 'aria-hidden': true as const }
      : {
          role: 'img',
          'aria-label':
            label || (level === 'rest' ? 'Descanso' : `Intensidade ${intensityLabels[level].toLowerCase()}`),
        };
  return <span className={`trail-sym trail-sym-${level}`} {...a11y} />;
}

export function TrailLegend() {
  return (
    <ul className="trail-legend" aria-label="Legenda de intensidade">
      {(['easy', 'moderate', 'hard'] as const).map((level) => (
        <li key={level}>
          <TrailSymbol intensity={level} label="" />
          <span>
            {intensityLabels[level]} <small>{intensityRanges[level]}</small>
          </span>
        </li>
      ))}
    </ul>
  );
}
