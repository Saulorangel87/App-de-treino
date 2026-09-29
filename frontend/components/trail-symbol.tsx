// Escala de intensidade do app inteiro, emprestada da sinalização de trilhas:
// círculo verde (leve), quadrado azul (moderado) e losango preto (intenso).

export type Intensity = 'easy' | 'moderate' | 'hard';

export function intensityOf(rpe: number): Intensity {
  if (rpe <= 3) return 'easy';
  if (rpe <= 6) return 'moderate';
  return 'hard';
}

export const intensityLabels: Record<Intensity, string> = {
  easy: 'Leve',
  moderate: 'Moderado',
  hard: 'Intenso',
};

export const intensityRanges: Record<Intensity, string> = {
  easy: 'RPE 1–3',
  moderate: 'RPE 4–6',
  hard: 'RPE 7+',
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
