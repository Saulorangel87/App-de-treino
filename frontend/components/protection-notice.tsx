'use client';

import { HeartPulse, LoaderCircle, ShieldAlert } from 'lucide-react';
import { defineMessages, INTL_LOCALE } from '@/lib/i18n';
import { parseTrainingDate, type WorkoutProtection } from '@/lib/planning';
import { useLocale, useMessages } from './locale-provider';

type Level = Exclude<WorkoutProtection['level'], 'none'>;

const messages = defineMessages({
  pt: {
    title: (level: Level) => `Proteção ${{ light: 'leve', moderate: 'moderada', strong: 'forte' }[level]} nos próximos treinos`,
    when: 'Reavaliamos a cada treino concluído e a cada check-in.',
    ends: (date: string) => ` Sem novos registros, a proteção termina em ${date}.`,
    scope: 'Os níveis e os prazos são critérios do Cadência, não recomendações de estudos científicos.',
    professional: 'Se a dor persistir, procure avaliação de um profissional de saúde.',
    recovered: 'Estou recuperado',
    recoveredHelp: 'Reavalia seus próximos treinos agora. Se houve dor nos últimos dias, a proteção é mantida por segurança.',
  },
  en: {
    title: (level: Level) => `${{ light: 'Light', moderate: 'Moderate', strong: 'Strong' }[level]} protection on your next workouts`,
    when: 'We reassess after every completed workout and every check-in.',
    ends: (date: string) => ` With no new records, the protection ends on ${date}.`,
    scope: 'The levels and time frames are Cadência criteria, not recommendations from scientific studies.',
    professional: 'If the pain persists, see a health professional.',
    recovered: "I've recovered",
    recoveredHelp: 'Reassesses your next workouts now. If you had pain in the last few days, the protection stays on for safety.',
  },
});

type ProtectionNoticeProps = {
  protection: WorkoutProtection;
  busy: boolean;
  onRecovered: () => void;
};

// Explica por que os próximos treinos estão protegidos, até quando e como o atleta
// pode pedir uma nova avaliação.
export function ProtectionNotice({ protection, busy, onRecovered }: ProtectionNoticeProps) {
  const locale = useLocale();
  const t = useMessages(messages);
  if (protection.level === 'none') return null;
  const reasons = protection.reasons.slice(0, 2);
  const dateFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: 'numeric', month: 'long' });
  return (
    <section className={`protection-notice level-${protection.level}`} aria-labelledby="protection-notice-title">
      <ShieldAlert size={18} aria-hidden="true" />
      <div>
        <strong id="protection-notice-title">{t.title(protection.level)}</strong>
        {reasons.map((reason) => (
          <p key={reason}>{reason}</p>
        ))}
        <p className="protection-notice-when">
          {t.when}
          {protection.expires_on && t.ends(dateFormatter.format(parseTrainingDate(protection.expires_on)))}
        </p>
        <p className="protection-notice-scope">{t.scope}</p>
        {protection.suggest_professional && <p className="protection-notice-professional">{t.professional}</p>}
        <div className="protection-notice-actions">
          <button type="button" className="btn btn-outline btn-sm" onClick={onRecovered} disabled={busy}>
            {busy ? <LoaderCircle className="spin" size={14} aria-hidden="true" /> : <HeartPulse size={14} aria-hidden="true" />}
            {t.recovered}
          </button>
          <small>{t.recoveredHelp}</small>
        </div>
      </div>
    </section>
  );
}
