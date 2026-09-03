import type { Workout } from '@/lib/planning';

type WorkoutStructureProps = {
  structure: Workout['structure'];
  durationMinutes?: number;
};

function formatRPE(value: number) {
  return Number.isInteger(value) ? String(value) : value.toFixed(1);
}

export function WorkoutStructure({ structure, durationMinutes }: WorkoutStructureProps) {
  const steps = structure.steps;

  if (!steps?.length) {
    return (
      <ol className="structured-workout structured-workout-fallback">
        <li>
          <span className="structured-workout-index">01</span>
          <div>
            <strong>Aquecimento</strong>
            <small>{structure.warmup_minutes} min · esforço progressivo</small>
          </div>
        </li>
        <li>
          <span className="structured-workout-index">02</span>
          <div>
            <strong>Parte principal</strong>
            <small>{structure.main}</small>
          </div>
        </li>
        <li>
          <span className="structured-workout-index">03</span>
          <div>
            <strong>Desaquecimento</strong>
            <small>{structure.cooldown_minutes} min · esforço leve</small>
          </div>
        </li>
      </ol>
    );
  }

  const displaySteps = fitStepsToDuration(steps, durationMinutes);

  return (
    <ol className="structured-workout">
      {displaySteps.map((step) => (
        <li className={`structured-workout-step structured-workout-${step.kind}`} key={`${step.order}-${step.title}`}>
          <span className="structured-workout-index">{String(step.order).padStart(2, '0')}</span>
          <div className="structured-workout-content">
            <strong>{step.title}</strong>
            <small>{step.duration_minutes} min · RPE {formatRPE(step.target_rpe)}</small>
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
