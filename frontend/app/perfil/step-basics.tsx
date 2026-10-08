import type { ProfileFormController } from './use-profile-form';
import { ArrowRight } from 'lucide-react';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { defineMessages } from '@/lib/i18n';
import { FormFeedback } from './profile-fields';

const messages = defineMessages({
  pt: {
    basics: 'Informações básicas',
    birthDate: 'Data de nascimento',
    sex: 'Sexo',
    sexOptions: [
      ['', 'Prefiro não informar agora'],
      ['female', 'Feminino'],
      ['male', 'Masculino'],
      ['other', 'Outro'],
      ['prefer_not_to_say', 'Prefiro não dizer'],
    ],
    height: 'Altura',
    weight: 'Peso atual',
    optional: 'Contexto opcional',
    optionalIntro: 'Estas medidas são apenas um registro de acompanhamento. Elas não geram diagnóstico nem alteram automaticamente a carga.',
    waist: 'Circunferência da cintura',
    bodyFat: 'Percentual de gordura corporal',
    weightTrend: 'Tendência recente do peso',
    weightTrendOptions: [
      ['not_informed', 'Não informar'],
      ['stable', 'Estável'],
      ['increasing', 'Aumentando'],
      ['decreasing', 'Diminuindo'],
    ],
    experience: 'Experiência no ciclismo',
    levels: [
      ['beginner', 'Iniciante', 'Estou começando ou retomando'],
      ['intermediate', 'Intermediário', 'Pedalo com regularidade'],
      ['advanced', 'Avançado', 'Treino estruturado há anos'],
    ],
    activity: 'Como está sua rotina de atividade hoje?',
    activityOptions: [
      ['', 'Selecione'],
      ['sedentary', 'Quase não pratico atividade'],
      ['occasional', '1–2 vezes por semana'],
      ['regular', '3–4 vezes por semana'],
      ['frequent', '5 ou mais vezes por semana'],
    ],
    saving: 'Salvando…',
    save: 'Salvar e continuar',
  },
  en: {
    basics: 'Basic information',
    birthDate: 'Date of birth',
    sex: 'Sex',
    sexOptions: [
      ['', "I'd rather not say for now"],
      ['female', 'Female'],
      ['male', 'Male'],
      ['other', 'Other'],
      ['prefer_not_to_say', 'Prefer not to say'],
    ],
    height: 'Height',
    weight: 'Current weight',
    optional: 'Optional context',
    optionalIntro: 'These measurements are only a tracking record. They do not produce a diagnosis or change the load automatically.',
    waist: 'Waist circumference',
    bodyFat: 'Body fat percentage',
    weightTrend: 'Recent weight trend',
    weightTrendOptions: [
      ['not_informed', "Don't say"],
      ['stable', 'Stable'],
      ['increasing', 'Increasing'],
      ['decreasing', 'Decreasing'],
    ],
    experience: 'Cycling experience',
    levels: [
      ['beginner', 'Beginner', "I'm starting out or getting back into it"],
      ['intermediate', 'Intermediate', 'I ride regularly'],
      ['advanced', 'Advanced', "I've trained in a structured way for years"],
    ],
    activity: 'How active is your routine today?',
    activityOptions: [
      ['', 'Select'],
      ['sedentary', 'I hardly do any activity'],
      ['occasional', '1–2 times a week'],
      ['regular', '3–4 times a week'],
      ['frequent', '5 or more times a week'],
    ],
    saving: 'Saving…',
    save: 'Save and continue',
  },
});

export function ProfileStep1({ form }: { form: ProfileFormController }) {
  const t = useMessages(messages);
  const { error, message, profile, saveProfile, saving, updateProfile } = form;

  return (
    <form onSubmit={saveProfile} className="profile-form">
      <fieldset>
        <legend>{t.basics}</legend>
        <div className="form-grid">
          <div>
            <Label htmlFor="birth_date">{t.birthDate}</Label>
            <Input
              id="birth_date"
              type="date"
              value={profile.birth_date}
              onChange={(event) =>
                updateProfile('birth_date', event.target.value)
              }
            />
          </div>
          <div>
            <Label htmlFor="sex">{t.sex}</Label>
            <select
              id="sex"
              value={profile.sex}
              onChange={(event) => updateProfile('sex', event.target.value)}
            >
              {t.sexOptions.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <Label htmlFor="height_cm">{t.height}</Label>
            <div className="unit-input">
              <Input
                id="height_cm"
                type="number"
                min="100"
                max="250"
                step="0.1"
                value={profile.height_cm}
                onChange={(event) =>
                  updateProfile('height_cm', event.target.value)
                }
              />
              <span>cm</span>
            </div>
          </div>
          <div>
            <Label htmlFor="weight_kg">{t.weight}</Label>
            <div className="unit-input">
              <Input
                id="weight_kg"
                type="number"
                min="30"
                max="350"
                step="0.1"
                value={profile.weight_kg}
                onChange={(event) =>
                  updateProfile('weight_kg', event.target.value)
                }
              />
              <span>kg</span>
            </div>
          </div>
        </div>
      </fieldset>
      <fieldset>
        <legend>{t.optional}</legend>
        <p className="fieldset-intro">{t.optionalIntro}</p>
        <div className="form-grid">
          <div>
            <Label htmlFor="waist_cm">{t.waist}</Label>
            <div className="unit-input">
              <Input
                id="waist_cm"
                type="number"
                min="30"
                max="250"
                step="0.1"
                value={profile.waist_cm}
                onChange={(event) =>
                  updateProfile('waist_cm', event.target.value)
                }
              />
              <span>cm</span>
            </div>
          </div>
          <div>
            <Label htmlFor="body_fat_percent">{t.bodyFat}</Label>
            <div className="unit-input">
              <Input
                id="body_fat_percent"
                type="number"
                min="2"
                max="70"
                step="0.1"
                value={profile.body_fat_percent}
                onChange={(event) =>
                  updateProfile('body_fat_percent', event.target.value)
                }
              />
              <span>%</span>
            </div>
          </div>
          <div>
            <Label htmlFor="weight_trend">{t.weightTrend}</Label>
            <select
              id="weight_trend"
              value={profile.weight_trend}
              onChange={(event) =>
                updateProfile('weight_trend', event.target.value)
              }
            >
              {t.weightTrendOptions.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </fieldset>
      <fieldset>
        <legend>{t.experience}</legend>
        <div className="choice-cards">
          {t.levels.map(([value, title, description]) => (
            <label key={value}>
              <input
                type="radio"
                name="experience_level"
                value={value}
                checked={profile.experience_level === value}
                onChange={(event) =>
                  updateProfile('experience_level', event.target.value)
                }
              />
              <span>
                <strong>{title}</strong>
                <small>{description}</small>
              </span>
            </label>
          ))}
        </div>
        <div className="activity-field">
          <Label htmlFor="activity_level">{t.activity}</Label>
          <select
            id="activity_level"
            value={profile.activity_level}
            onChange={(event) =>
              updateProfile('activity_level', event.target.value)
            }
          >
            {t.activityOptions.map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </div>
      </fieldset>
      <FormFeedback error={error} message={message} />
      <Button type="submit" disabled={saving} className="profile-submit">
        {saving ? t.saving : t.save}
        <ArrowRight size={16} />
      </Button>
    </form>
  );
}
