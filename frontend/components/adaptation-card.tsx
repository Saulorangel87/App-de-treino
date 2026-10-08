'use client';

import {
  ArrowDownRight,
  ArrowUpRight,
  ShieldAlert,
  Sparkles,
} from 'lucide-react';
import { defineMessages } from '@/lib/i18n';
import type { Workout } from '@/lib/planning';
import { zoneForRpe, zoneLabel } from '@/lib/zones';
import { useLocale, useMessages } from './locale-provider';
import styles from './adaptation-card.module.css';

type AdaptationCardProps = {
  workout: Workout;
  compact?: boolean;
};

const messages = defineMessages({
  pt: {
    kinds: { safety: 'AJUSTE DE SEGURANÇA', recovery: 'CARGA AJUSTADA', progression: 'PROGRESSÃO LEVE' },
    label: 'Adaptação automática do treino',
    title: 'Plano ajustado com seu feedback',
    duration: 'DURAÇÃO',
    zone: 'ZONA',
  },
  en: {
    kinds: { safety: 'SAFETY ADJUSTMENT', recovery: 'LOAD ADJUSTED', progression: 'GENTLE PROGRESSION' },
    label: 'Automatic workout adaptation',
    title: 'Plan adjusted with your feedback',
    duration: 'DURATION',
    zone: 'ZONE',
  },
});

export function AdaptationCard({
  workout,
  compact = false,
}: AdaptationCardProps) {
  const locale = useLocale();
  const t = useMessages(messages);
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
    <section className={classes} aria-label={t.label}>
      <div className={styles.heading}>
        <span className={styles.icon}>
          <Icon size={15} />
        </span>
        <div>
          <small>{t.kinds[adaptation.kind]}</small>
          <strong>{t.title}</strong>
        </div>
        <Sparkles className={styles.spark} size={15} aria-hidden="true" />
      </div>
      <p className={styles.reason}>{adaptation.reason}</p>
      <div className={styles.comparison}>
        <span>
          <small>{t.duration}</small>
          <s>{adaptation.previous_duration_minutes} min</s>
          <b>{workout.duration_minutes} min</b>
        </span>
        {zoneForRpe(adaptation.previous_target_rpe).number !==
          zoneForRpe(workout.target_rpe).number && (
          <span>
            <small>{t.zone}</small>
            <s>{zoneLabel(adaptation.previous_target_rpe, locale)}</s>
            <b>{zoneLabel(workout.target_rpe, locale)}</b>
          </span>
        )}
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
