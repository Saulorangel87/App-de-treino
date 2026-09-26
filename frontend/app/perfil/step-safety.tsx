import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FormFeedback } from './profile-fields';
import { SAFETY_SYMPTOMS, questionIsVisible } from './profile-model';

export function ProfileStep2({ form }: { form: ProfileFormController }) {
  const {
    error,
    goToStep,
    hasLimitation,
    limitation,
    message,
    questionnaire,
    saveLimitations,
    saving,
    setHasLimitation,
    setLimitation,
    toggleSafetySymptom,
  } = form;

  return (
    <form onSubmit={saveLimitations} className="profile-form">
      <fieldset>
        <legend>Condição atual</legend>
        <div className="binary-choice">
          <label>
            <input
              type="radio"
              name="has_limitation"
              checked={!hasLimitation}
              onChange={() => setHasLimitation(false)}
            />
            <span>
              <strong>Nenhuma limitação atual</strong>
              <small>Posso pedalar sem dor ou restrição conhecida</small>
            </span>
          </label>
          <label>
            <input
              type="radio"
              name="has_limitation"
              checked={hasLimitation}
              onChange={() => setHasLimitation(true)}
            />
            <span>
              <strong>Tenho algo a considerar</strong>
              <small>Dor, lesão, condição ou restrição de movimento</small>
            </span>
          </label>
        </div>
      </fieldset>
      {hasLimitation &&
        questionIsVisible(questionnaire, 'limitation_kind', {
          has_limitation: true,
        }) && (
          <fieldset>
            <legend>O que devemos respeitar?</legend>
            <p className="fieldset-intro">
              Esses detalhes são opcionais e ajudam a registrar o contexto com
              mais precisão. Não são um diagnóstico.
            </p>
            <div className="form-grid">
              <div>
                <Label htmlFor="limitation_kind">Tipo</Label>
                <select
                  id="limitation_kind"
                  value={limitation.kind}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      kind: event.target.value,
                    }))
                  }
                >
                  <option value="pain">Dor ou desconforto</option>
                  <option value="injury">Lesão</option>
                  <option value="medical_condition">Condição de saúde</option>
                  <option value="mobility">Limitação de movimento</option>
                  <option value="other">Outro</option>
                </select>
              </div>
              <div>
                <Label htmlFor="limitation_location">Localização</Label>
                <Input
                  id="limitation_location"
                  maxLength={120}
                  value={limitation.location}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      location: event.target.value,
                    }))
                  }
                  placeholder="Ex.: joelho direito"
                />
              </div>
              <div>
                <Label htmlFor="limitation_intensity">
                  Intensidade percebida
                </Label>
                <select
                  id="limitation_intensity"
                  value={limitation.intensity}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      intensity: event.target.value,
                    }))
                  }
                >
                  <option value="">Não informar</option>
                  {Array.from({ length: 10 }, (_, index) => (
                    <option key={index + 1} value={index + 1}>
                      {index + 1} de 10
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <Label htmlFor="limitation_started_on">Quando começou?</Label>
                <Input
                  id="limitation_started_on"
                  type="date"
                  value={limitation.started_on}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      started_on: event.target.value,
                    }))
                  }
                />
              </div>
              <label className="clearance-check">
                <input
                  type="checkbox"
                  checked={limitation.professional_clearance_recommended}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      professional_clearance_recommended: event.target.checked,
                    }))
                  }
                />
                <span>
                  <strong>Orientação profissional recomendada</strong>
                  <small>
                    Marque se um médico ou fisioterapeuta deve liberar o treino
                  </small>
                </span>
              </label>
              <label className="clearance-check">
                <input
                  type="checkbox"
                  checked={limitation.medical_restriction}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      medical_restriction: event.target.checked,
                    }))
                  }
                />
                <span>
                  <strong>Existe restrição médica atual</strong>
                  <small>
                    O plano deve manter a carga protegida até nova orientação
                  </small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label="Passei por uma cirurgia recentemente"
              >
                <input
                  type="checkbox"
                  checked={limitation.recent_surgery}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      recent_surgery: event.target.checked,
                    }))
                  }
                />
                <span>
                  <strong>Passei por uma cirurgia recentemente</strong>
                  <small>
                    Não inicie ou avance o plano sem a orientação de quem
                    acompanha sua recuperação.
                  </small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label="Recebi orientação para não me exercitar"
              >
                <input
                  type="checkbox"
                  checked={limitation.exercise_prohibited}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      exercise_prohibited: event.target.checked,
                    }))
                  }
                />
                <span>
                  <strong>Recebi orientação para não me exercitar</strong>
                  <small>
                    O Cadência não substitui essa orientação e preserva uma
                    leitura protegida.
                  </small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label="Tenho uma condição que afeta o exercício"
              >
                <input
                  type="checkbox"
                  checked={limitation.condition_affecting_exercise}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      condition_affecting_exercise: event.target.checked,
                    }))
                  }
                />
                <span>
                  <strong>Tenho uma condição que afeta o exercício</strong>
                  <small>
                    Registre o contexto para que a segurança prevaleça sobre
                    qualquer meta.
                  </small>
                </span>
              </label>
            </div>
            <div className="form-grid">
              <div className="textarea-field">
                <Label htmlFor="limitation_description">
                  Descreva brevemente
                </Label>
                <textarea
                  id="limitation_description"
                  minLength={3}
                  maxLength={500}
                  required
                  value={limitation.description}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      description: event.target.value,
                    }))
                  }
                  placeholder="Ex.: desconforto no joelho direito ao subir…"
                />
              </div>
              <div className="textarea-field">
                <Label htmlFor="limitation_aggravating_movement">
                  O que agrava?
                </Label>
                <textarea
                  id="limitation_aggravating_movement"
                  maxLength={200}
                  value={limitation.aggravating_movement}
                  onChange={(event) =>
                    setLimitation((current) => ({
                      ...current,
                      aggravating_movement: event.target.value,
                    }))
                  }
                  placeholder="Ex.: subir em pé ou pedalar forte"
                />
              </div>
            </div>
            <fieldset className="symptoms-field">
              <legend>Sintomas durante ou depois do treino</legend>
              <small>
                Opcional. Se houver sinais importantes, interrompa e procure
                avaliação profissional.
              </small>
              <div className="preference-choice">
                {SAFETY_SYMPTOMS.map((symptom) => (
                  <label key={symptom.value}>
                    <input
                      type="checkbox"
                      checked={limitation.symptoms_during_after.includes(
                        symptom.value,
                      )}
                      onChange={() => toggleSafetySymptom(symptom.value)}
                    />
                    <span>{symptom.label}</span>
                  </label>
                ))}
              </div>
            </fieldset>
          </fieldset>
        )}
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(1)}>
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
