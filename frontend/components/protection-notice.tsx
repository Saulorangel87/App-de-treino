import { HeartPulse, LoaderCircle, ShieldAlert } from 'lucide-react';
import { parseTrainingDate, type WorkoutProtection } from '@/lib/planning';

const levelLabel: Record<Exclude<WorkoutProtection['level'], 'none'>, string> = {
  light: 'leve',
  moderate: 'moderada',
  strong: 'forte',
};

const dateFormatter = new Intl.DateTimeFormat('pt-BR', { day: 'numeric', month: 'long' });

type ProtectionNoticeProps = {
  protection: WorkoutProtection;
  busy: boolean;
  onRecovered: () => void;
};

// Explica por que os próximos treinos estão protegidos, até quando e como o atleta
// pode pedir uma nova avaliação.
export function ProtectionNotice({ protection, busy, onRecovered }: ProtectionNoticeProps) {
  if (protection.level === 'none') return null;
  const reasons = protection.reasons.slice(0, 2);
  return (
    <section className={`protection-notice level-${protection.level}`} aria-labelledby="protection-notice-title">
      <ShieldAlert size={18} aria-hidden="true" />
      <div>
        <strong id="protection-notice-title">Proteção {levelLabel[protection.level]} nos próximos treinos</strong>
        {reasons.map((reason) => (
          <p key={reason}>{reason}</p>
        ))}
        <p className="protection-notice-when">
          Reavaliamos a cada treino concluído e a cada check-in.
          {protection.expires_on &&
            ` Sem novos registros, a proteção termina em ${dateFormatter.format(parseTrainingDate(protection.expires_on))}.`}
        </p>
        {protection.suggest_professional && (
          <p className="protection-notice-professional">
            Se a dor persistir, procure avaliação de um profissional de saúde.
          </p>
        )}
        <div className="protection-notice-actions">
          <button type="button" className="btn btn-outline btn-sm" onClick={onRecovered} disabled={busy}>
            {busy ? <LoaderCircle className="spin" size={14} aria-hidden="true" /> : <HeartPulse size={14} aria-hidden="true" />}
            Estou recuperado
          </button>
          <small>Reavalia seus próximos treinos agora. Se houve dor nos últimos dias, a proteção é mantida por segurança.</small>
        </div>
      </div>
    </section>
  );
}
