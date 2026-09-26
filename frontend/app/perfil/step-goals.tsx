import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FormFeedback, GoalOptions } from './profile-fields';

export function ProfileStep3({ form }: { form: ProfileFormController }) {
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
        <legend>Objetivo principal</legend>
        <div className="form-grid">
          <div>
            <Label htmlFor="primary_goal">Quero principalmente</Label>
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
            <Label htmlFor="target_date">Data-alvo (opcional)</Label>
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
          <Label htmlFor="goal_details">Conte um pouco mais (opcional)</Label>
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
            placeholder="Ex.: completar meu primeiro pedal de 100 km com segurança…"
          />
        </div>
      </fieldset>
      <fieldset>
        <legend>Objetivo secundário</legend>
        <div className="activity-field">
          <Label htmlFor="secondary_goal">Outra prioridade (opcional)</Label>
          <select
            id="secondary_goal"
            value={secondaryGoal}
            onChange={(event) => setSecondaryGoal(event.target.value)}
          >
            <option value="">Nenhuma por enquanto</option>
            <GoalOptions exclude={primaryGoal.goal_type} />
          </select>
          <small className="profile-field-note">
            Ela só desempata estímulos já elegíveis. Segurança e objetivo
            principal sempre têm prioridade.
          </small>
        </div>
      </fieldset>
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(2)}>
          <ArrowLeft size={15} /> Voltar
        </Button>
        <Button type="submit" disabled={saving} className="profile-submit">
          {saving ? 'Salvando…' : 'Salvar e continuar'}
          <ArrowRight size={16} />
        </Button>
      </div>
    </form>
  );
}
