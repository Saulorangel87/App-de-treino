import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { defineMessages } from '@/lib/i18n';
import { FormFeedback, GoalOptions } from './profile-fields';

const messages = defineMessages({
  pt: {
    primary: 'Objetivo principal',
    mainly: 'Quero principalmente',
    targetDate: 'Data-alvo (opcional)',
    details: 'Conte um pouco mais (opcional)',
    detailsPlaceholder: 'Ex.: completar meu primeiro pedal de 100 km com segurança…',
    secondary: 'Objetivo secundário',
    otherPriority: 'Outra prioridade (opcional)',
    none: 'Nenhuma por enquanto',
    secondaryNote: 'Ela só desempata estímulos já elegíveis. Segurança e objetivo principal sempre têm prioridade.',
    back: 'Voltar',
    saving: 'Salvando…',
    save: 'Salvar e continuar',
  },
  en: {
    primary: 'Main goal',
    mainly: 'Mainly, I want to',
    targetDate: 'Target date (optional)',
    details: 'Tell us a bit more (optional)',
    detailsPlaceholder: 'E.g.: finish my first 100 km ride safely…',
    secondary: 'Secondary goal',
    otherPriority: 'Another priority (optional)',
    none: 'None for now',
    secondaryNote: 'It only breaks ties between stimuli that are already eligible. Safety and the main goal always come first.',
    back: 'Back',
    saving: 'Saving…',
    save: 'Save and continue',
  },
});

export function ProfileStep3({ form }: { form: ProfileFormController }) {
  const t = useMessages(messages);
  const {
    error,
    goToStep,
    message,
    primaryGoal,
    saveGoals,
    saving,
    secondaryGoal,
    setPrimaryGoal,
    setSecondaryGoal,
  } = form;

  return (
    <form onSubmit={saveGoals} className="profile-form">
      <fieldset>
        <legend>{t.primary}</legend>
        <div className="form-grid">
          <div>
            <Label htmlFor="primary_goal">{t.mainly}</Label>
            <select
              id="primary_goal"
              value={primaryGoal.goal_type}
              onChange={(event) =>
                setPrimaryGoal((current) => ({
                  ...current,
                  goal_type: event.target.value,
                }))
              }
            >
              <GoalOptions />
            </select>
          </div>
          <div>
            <Label htmlFor="target_date">{t.targetDate}</Label>
            <Input
              id="target_date"
              type="date"
              value={primaryGoal.target_date}
              min={new Date().toISOString().slice(0, 10)}
              onChange={(event) =>
                setPrimaryGoal((current) => ({
                  ...current,
                  target_date: event.target.value,
                }))
              }
            />
          </div>
        </div>
        <div className="textarea-field">
          <Label htmlFor="goal_details">{t.details}</Label>
          <textarea
            id="goal_details"
            maxLength={500}
            value={primaryGoal.details}
            onChange={(event) =>
              setPrimaryGoal((current) => ({
                ...current,
                details: event.target.value,
              }))
            }
            placeholder={t.detailsPlaceholder}
          />
        </div>
      </fieldset>
      <fieldset>
        <legend>{t.secondary}</legend>
        <div className="activity-field">
          <Label htmlFor="secondary_goal">{t.otherPriority}</Label>
          <select
            id="secondary_goal"
            value={secondaryGoal}
            onChange={(event) => setSecondaryGoal(event.target.value)}
          >
            <option value="">{t.none}</option>
            <GoalOptions exclude={primaryGoal.goal_type} />
          </select>
          <small className="profile-field-note">{t.secondaryNote}</small>
        </div>
      </fieldset>
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(2)}>
          <ArrowLeft size={15} /> {t.back}
        </Button>
        <Button type="submit" disabled={saving} className="profile-submit">
          {saving ? t.saving : t.save}
          <ArrowRight size={16} />
        </Button>
      </div>
    </form>
  );
}
