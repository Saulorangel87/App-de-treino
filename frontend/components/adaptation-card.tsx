import {
  ArrowDownRight,
  ArrowUpRight,
  ShieldAlert,
  Sparkles,
} from 'lucide-react';
import type { Workout } from '@/lib/planning';
import styles from './adaptation-card.module.css';

type AdaptationCardProps = {
  workout: Workout;
  compact?: boolean;
};

const labels = {
  safety: 'AJUSTE DE SEGURANÇA',
  recovery: 'CARGA AJUSTADA',
  progression: 'PROGRESSÃO LEVE',
} as const;

export function AdaptationCard({
  workout,
  compact = false,
}: AdaptationCardProps) {
  const adaptation = workout.explanation.adaptation;
  if (!adaptation) return null;

  const Icon =
    adaptation.kind === 'safety'
      ? ShieldAlert
      : adaptation.kind === 'progression'
        ? ArrowUpRight
        : ArrowDownRight;
  const classes = [
    styles.card,
    styles[adaptation.kind],
    compact ? styles.compact : '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <section className={classes} aria-label="Adaptação automática do treino">
      <div className={styles.heading}>
        <span className={styles.icon}>
          <Icon size={15} />
        </span>
        <div>
          <small>{labels[adaptation.kind]}</small>
          <strong>Plano ajustado com seu feedback</strong>
        </div>
        <Sparkles className={styles.spark} size={15} aria-hidden="true" />
      </div>
      <p className={styles.reason}>{adaptation.reason}</p>
      <div className={styles.comparison}>
        <span>
          <small>DURAÇÃO</small>
          <s>{adaptation.previous_duration_minutes} min</s>
          <b>{workout.duration_minutes} min</b>
        </span>
        <span>
          <small>ESFORÇO</small>
          <s>RPE {adaptation.previous_target_rpe}</s>
          <b>RPE {workout.target_rpe}</b>
        </span>
      </div>
      {adaptation.safety_notice && (
        <div className={styles.warning} role="note">
          <ShieldAlert size={15} />
          <span>{adaptation.safety_notice}</span>
        </div>
      )}
    </section>
  );
}
