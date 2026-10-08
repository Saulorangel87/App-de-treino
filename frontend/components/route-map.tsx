import type { Locale } from '@/lib/i18n';
import type { Workout, WorkoutStep } from '@/lib/planning';
import { intensityOf } from './trail-symbol';

// Desenha a sessão como um percurso de carta topográfica: cada etapa ocupa um
// trecho proporcional à sua duração e recebe a cor da escala de intensidade.
// Tudo é calculado de forma determinística (sem medir o DOM), então o SVG é o
// mesmo no servidor e no navegador.

const WIDTH = 560;
const HEIGHT = 300;
const SAMPLES = 160;

type Point = [number, number];

function routePoint(t: number): Point {
  const x = 40 + 480 * t + 16 * Math.sin(2 * Math.PI * 2 * t);
  const y =
    156 +
    68 * Math.sin(2 * Math.PI * 1.15 * t + 0.4) +
    26 * Math.sin(2 * Math.PI * 2.7 * t + 1.3);
  return [x, y];
}

const samples: Point[] = Array.from({ length: SAMPLES + 1 }, (_, i) => routePoint(i / SAMPLES));
const cumulative = samples.reduce<number[]>((acc, point, index) => {
  if (index === 0) return [0];
  const [px, py] = samples[index - 1];
  acc.push(acc[index - 1] + Math.hypot(point[0] - px, point[1] - py));
  return acc;
}, []);
const totalLength = cumulative[cumulative.length - 1];

function pointAt(fraction: number): Point {
  const target = Math.min(1, Math.max(0, fraction)) * totalLength;
  let index = cumulative.findIndex((length) => length >= target);
  if (index <= 0) return samples[0];
  if (index === -1) index = samples.length - 1;
  const span = cumulative[index] - cumulative[index - 1] || 1;
  const k = (target - cumulative[index - 1]) / span;
  const [ax, ay] = samples[index - 1];
  const [bx, by] = samples[index];
  return [ax + (bx - ax) * k, ay + (by - ay) * k];
}

function segmentPoints(from: number, to: number) {
  const points: Point[] = [pointAt(from)];
  cumulative.forEach((length, index) => {
    const fraction = length / totalLength;
    if (fraction > from && fraction < to) points.push(samples[index]);
  });
  points.push(pointAt(to));
  return points.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(' ');
}

function contour(cx: number, cy: number, rx: number, ry: number, seed: number, k: number) {
  const points: string[] = [];
  for (let i = 0; i <= 64; i += 1) {
    const t = (i / 64) * Math.PI * 2;
    const r =
      1 +
      0.18 * Math.sin(3 * t + seed) +
      0.1 * Math.sin(5 * t + seed * 2.3) +
      0.06 * Math.sin(7 * t + seed * 0.7);
    points.push(`${(cx + Math.cos(t) * rx * r * k).toFixed(1)} ${(cy + Math.sin(t) * ry * r * k).toFixed(1)}`);
  }
  return `M${points.join(' L')}Z`;
}

const hills: Array<[number, number, number, number, number, number]> = [
  [430, 70, 170, 110, 1.2, 7],
  [120, 250, 150, 90, 3.4, 6],
  [300, 180, 90, 60, 5.1, 3],
];

const contours = hills.flatMap(([cx, cy, rx, ry, seed, count]) =>
  Array.from({ length: count }, (_, i) => ({
    d: contour(cx, cy, rx, ry, seed + (i + 1) * 0.08, 1.3 - (i + 1) * 0.15),
    index: (i + 1) % 5 === 0,
  })),
);

const fallbackTitles = {
  pt: { warmup: 'Aquecimento', main: 'Parte principal', cooldown: 'Volta à calma' },
  en: { warmup: 'Warm-up', main: 'Main set', cooldown: 'Cool-down' },
};

export function stepsForWorkout(workout: Pick<Workout, 'structure' | 'duration_minutes' | 'target_rpe'>, locale: Locale): WorkoutStep[] {
  const steps = workout.structure.steps;
  const titles = fallbackTitles[locale];
  if (steps?.length) return steps;
  const warmup = workout.structure.warmup_minutes || 0;
  const cooldown = workout.structure.cooldown_minutes || 0;
  const main = Math.max(1, workout.duration_minutes - warmup - cooldown);
  return [
    warmup && { order: 1, kind: 'warmup', title: titles.warmup, duration_minutes: warmup, target_rpe: Math.min(3, workout.target_rpe), instruction: '' },
    { order: 2, kind: 'main', title: titles.main, duration_minutes: main, target_rpe: workout.target_rpe, instruction: '' },
    cooldown && { order: 3, kind: 'cooldown', title: titles.cooldown, duration_minutes: cooldown, target_rpe: 2, instruction: '' },
  ].filter(Boolean) as WorkoutStep[];
}

export function RouteMap({ steps, label }: { steps: WorkoutStep[]; label: string }) {
  const total = steps.reduce((sum, step) => sum + step.duration_minutes, 0) || 1;
  const segments = steps.map((step, index) => {
    const before = steps.slice(0, index).reduce((sum, item) => sum + item.duration_minutes, 0);
    return {
      step,
      from: before / total,
      to: (before + step.duration_minutes) / total,
      intensity: intensityOf(step.target_rpe),
    };
  });
  const numbered = steps.length <= 9;
  const start = pointAt(0);
  const finish = pointAt(1);

  return (
    <svg
      className="route-map"
      viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
      role="img"
      aria-label={label}
      preserveAspectRatio="xMidYMid meet"
    >
      <g className="route-terrain" aria-hidden="true">
        <path className="route-forest" d={contour(430, 70, 170, 110, 1.2, 1.35)} />
        {contours.map((line, index) => (
          <path key={index} className={line.index ? 'route-contour index' : 'route-contour'} d={line.d} />
        ))}
        <path className="route-water" d="M-10 214 C 90 196, 150 262, 260 246 S 440 196, 570 236" />
      </g>
      <g aria-hidden="true">
        <polyline className="route-halo" points={segmentPoints(0, 1)} />
        {segments.map(({ step, from, to, intensity }) => (
          <polyline
            key={`${step.order}-${from}`}
            className={`route-line route-${intensity}`}
            points={segmentPoints(from, to)}
          />
        ))}
        {numbered &&
          segments.map(({ step, from, to, intensity }) => {
            const [x, y] = pointAt((from + to) / 2);
            return (
              <g key={`m-${step.order}-${from}`} className={`route-marker route-marker-${intensity}`}>
                {intensity === 'hard' ? (
                  <rect x={x - 9} y={y - 9} width="18" height="18" transform={`rotate(45 ${x} ${y})`} />
                ) : intensity === 'moderate' ? (
                  <rect x={x - 10} y={y - 10} width="20" height="20" />
                ) : (
                  <circle cx={x} cy={y} r="10.5" />
                )}
                <text x={x} y={y + 4.5} textAnchor="middle">
                  {step.order}
                </text>
              </g>
            );
          })}
        <circle className="route-start" cx={start[0]} cy={start[1]} r="8" />
        <path
          className="route-finish"
          d={`M${finish[0]} ${finish[1]} l0 -28 l18 7 l-18 7`}
        />
      </g>
    </svg>
  );
}

export function RouteScale({ minutes }: { minutes: number }) {
  const marks = [0, 0.25, 0.5, 0.75, 1].map((f) => Math.round(minutes * f));
  return (
    <div className="route-scale" aria-hidden="true">
      <div className="route-scale-bar">
        <i />
        <i />
        <i />
        <i />
      </div>
      <div className="route-scale-marks">
        {marks.map((mark, index) => (
          <span key={index}>{mark}&prime;</span>
        ))}
      </div>
    </div>
  );
}
