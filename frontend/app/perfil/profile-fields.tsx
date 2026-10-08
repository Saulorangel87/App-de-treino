import { Check, CircleAlert } from 'lucide-react';
import { useMessages } from '@/components/locale-provider';
import { profileText } from './profile-model';

export function GoalOptions({ exclude = '' }: { exclude?: string }) {
  const { goals } = useMessages(profileText);
  return (
    <>
      {Object.entries(goals)
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
