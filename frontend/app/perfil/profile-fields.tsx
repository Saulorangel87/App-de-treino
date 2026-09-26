import { Check, CircleAlert } from 'lucide-react';

export function GoalOptions({ exclude = '' }: { exclude?: string }) {
  return (
    <>
      {[
        ['health', 'Melhorar saúde e bem-estar'],
        ['fitness', 'Ganhar condicionamento'],
        ['endurance', 'Pedalar por mais tempo'],
        ['performance', 'Aumentar meu desempenho'],
        ['event', 'Preparar para um evento'],
        ['weight_management', 'Apoiar o controle de peso'],
      ]
        .filter(([value]) => value !== exclude)
        .map(([value, label]) => (
          <option key={value} value={value}>
            {label}
          </option>
        ))}
    </>
  );
}

export function FormFeedback({
  error,
  message,
}: {
  error: string;
  message: string;
}) {
  return (
    <>
      {error && (
        <p className="profile-message error" role="alert">
          <CircleAlert size={15} />
          {error}
        </p>
      )}
      {message && (
        <p className="profile-message" role="status">
          <Check size={15} />
          {message}
        </p>
      )}
    </>
  );
}
