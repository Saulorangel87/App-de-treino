import type { ProfileFormController } from './use-profile-form';
import { ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FormFeedback } from './profile-fields';

export function ProfileStep1({ form }: { form: ProfileFormController }) {
  const { error, message, profile, saveProfile, saving, updateProfile } = form;

  return (
    <form onSubmit={saveProfile} className="profile-form">
      <fieldset>
        <legend>Informações básicas</legend>
        <div className="form-grid">
          <div>
            <Label htmlFor="birth_date">Data de nascimento</Label>
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
            <Label htmlFor="sex">Sexo</Label>
            <select
              id="sex"
              value={profile.sex}
              onChange={(event) => updateProfile('sex', event.target.value)}
            >
              <option value="">Prefiro não informar agora</option>
              <option value="female">Feminino</option>
              <option value="male">Masculino</option>
              <option value="other">Outro</option>
              <option value="prefer_not_to_say">Prefiro não dizer</option>
            </select>
          </div>
          <div>
            <Label htmlFor="height_cm">Altura</Label>
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
            <Label htmlFor="weight_kg">Peso atual</Label>
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
        <legend>Contexto opcional</legend>
        <p className="fieldset-intro">
          Estas medidas são apenas um registro de acompanhamento. Elas não geram
          diagnóstico nem alteram automaticamente a carga.
        </p>
        <div className="form-grid">
          <div>
            <Label htmlFor="waist_cm">Circunferência da cintura</Label>
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
            <Label htmlFor="body_fat_percent">
              Percentual de gordura corporal
            </Label>
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
            <Label htmlFor="weight_trend">Tendência recente do peso</Label>
            <select
              id="weight_trend"
              value={profile.weight_trend}
              onChange={(event) =>
                updateProfile('weight_trend', event.target.value)
              }
            >
              <option value="not_informed">Não informar</option>
              <option value="stable">Estável</option>
              <option value="increasing">Aumentando</option>
              <option value="decreasing">Diminuindo</option>
            </select>
          </div>
        </div>
      </fieldset>
      <fieldset>
        <legend>Experiência no ciclismo</legend>
        <div className="choice-cards">
          {[
            ['beginner', 'Iniciante', 'Estou começando ou retomando'],
            ['intermediate', 'Intermediário', 'Pedalo com regularidade'],
            ['advanced', 'Avançado', 'Treino estruturado há anos'],
          ].map(([value, title, description]) => (
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
          <Label htmlFor="activity_level">
            Como está sua rotina de atividade hoje?
          </Label>
          <select
            id="activity_level"
            value={profile.activity_level}
            onChange={(event) =>
              updateProfile('activity_level', event.target.value)
            }
          >
            <option value="">Selecione</option>
            <option value="sedentary">Quase não pratico atividade</option>
            <option value="occasional">1–2 vezes por semana</option>
            <option value="regular">3–4 vezes por semana</option>
            <option value="frequent">5 ou mais vezes por semana</option>
          </select>
        </div>
      </fieldset>
      <FormFeedback error={error} message={message} />
      <Button type="submit" disabled={saving} className="profile-submit">
        {saving ? 'Salvando…' : 'Salvar e continuar'}
        <ArrowRight size={16} />
      </Button>
    </form>
  );
}
