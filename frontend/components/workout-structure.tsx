'use client';

import { defineMessages } from '@/lib/i18n';
import type { Workout } from '@/lib/planning';
import { zoneLabel } from '@/lib/zones';
import { useLocale, useMessages } from './locale-provider';
import { intensityOf } from './trail-symbol';

const messages = defineMessages({
  pt: {
    warmup: 'Aquecimento',
    warmupEffort: 'esforço progressivo',
    main: 'Parte principal',
    cooldown: 'Desaquecimento',
    cooldownEffort: 'esforço leve',
  },
  en: {
    warmup: 'Warm-up',
    warmupEffort: 'progressive effort',
    main: 'Main set',
    cooldown: 'Cool-down',
    cooldownEffort: 'easy effort',
  },
});

type WorkoutStructureProps = {
  structure: Workout['structure'];
  durationMinutes?: number;
};

export function WorkoutStructure({ structure, durationMinutes }: WorkoutStructureProps) {
  const locale = useLocale();
  const t = useMessages(messages);
  const steps = structure.steps;

  if (!steps?.length) {
    return (
      <ol className="structured-workout structured-workout-fallback">
        <li>
          <span className="structured-workout-index">01</span>
          <div>
            <strong>{t.warmup}</strong>
            <small>{structure.warmup_minutes} min · {t.warmupEffort}</small>
          </div>
        </li>
        <li>
          <span className="structured-workout-index">02</span>
          <div>
            <strong>{t.main}</strong>
            <small>{structure.main}</small>
          </div>
        </li>
        <li>
          <span className="structured-workout-index">03</span>
          <div>
            <strong>{t.cooldown}</strong>
            <small>{structure.cooldown_minutes} min · {t.cooldownEffort}</small>
          </div>
        </li>
      </ol>
    );
  }

  const displaySteps = fitStepsToDuration(steps, durationMinutes);

  return (
    <ol className="structured-workout">
      {displaySteps.map((step) => (
        <li
          className={`structured-workout-step structured-workout-${step.kind}`}
          data-intensity={intensityOf(step.target_rpe)}
          key={`${step.order}-${step.title}`}
        >
          <span className="structured-workout-index">{String(step.order).padStart(2, '0')}</span>
          <div className="structured-workout-content">
            <strong>{step.title}</strong>
            <small>{step.duration_minutes} min · {zoneLabel(step.target_rpe, locale)}</small>
            <p>{step.instruction}</p>
          </div>
        </li>
      ))}
    </ol>
  );
}

function fitStepsToDuration(steps: NonNullable<Workout['structure']['steps']>, durationMinutes?: number) {
  if (!durationMinutes || durationMinutes <= 0) return steps;
  const total = steps.reduce((sum, step) => sum + step.duration_minutes, 0);
  if (total === durationMinutes) return steps;

  const scaled = steps.map((step) => ({
    ...step,
    duration_minutes: Math.max(1, Math.round((step.duration_minutes / total) * durationMinutes)),
  }));
  let difference = durationMinutes - scaled.reduce((sum, step) => sum + step.duration_minutes, 0);
  for (let index = scaled.length - 1; index >= 0 && difference !== 0; index -= 1) {
    if (difference > 0) {
      scaled[index].duration_minutes += difference;
      difference = 0;
    } else {
      const reduction = Math.min(scaled[index].duration_minutes - 1, -difference);
      scaled[index].duration_minutes -= reduction;
      difference += reduction;
    }
  }
  return scaled;
}
