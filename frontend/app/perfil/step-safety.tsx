import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { defineMessages } from '@/lib/i18n';
import { FormFeedback } from './profile-fields';
import { SAFETY_SYMPTOM_VALUES, profileText, questionIsVisible } from './profile-model';

const messages = defineMessages({
  pt: {
    current: 'Condição atual',
    none: 'Nenhuma limitação atual',
    noneHint: 'Posso pedalar sem dor ou restrição conhecida',
    some: 'Tenho algo a considerar',
    someHint: 'Dor, lesão, condição ou restrição de movimento',
    respect: 'O que devemos respeitar?',
    respectIntro: 'Esses detalhes são opcionais e ajudam a registrar o contexto com mais precisão. Não são um diagnóstico.',
    kind: 'Tipo',
    kinds: [
      ['pain', 'Dor ou desconforto'],
      ['injury', 'Lesão'],
      ['medical_condition', 'Condição de saúde'],
      ['mobility', 'Limitação de movimento'],
      ['other', 'Outro'],
    ],
    location: 'Localização',
    locationPlaceholder: 'Ex.: joelho direito',
    intensity: 'Intensidade percebida',
    notInformed: 'Não informar',
    outOf10: (value: number) => `${value} de 10`,
    started: 'Quando começou?',
    clearance: 'Orientação profissional recomendada',
    clearanceHint: 'Marque se um médico ou fisioterapeuta deve liberar o treino',
    restriction: 'Existe restrição médica atual',
    restrictionHint: 'O plano deve manter a carga protegida até nova orientação',
    surgery: 'Passei por uma cirurgia recentemente',
    surgeryHint: 'Não inicie ou avance o plano sem a orientação de quem acompanha sua recuperação.',
    prohibited: 'Recebi orientação para não me exercitar',
    prohibitedHint: 'O Cadência não substitui essa orientação e preserva uma leitura protegida.',
    condition: 'Tenho uma condição que afeta o exercício',
    conditionHint: 'Registre o contexto para que a segurança prevaleça sobre qualquer meta.',
    describe: 'Descreva brevemente',
    describePlaceholder: 'Ex.: desconforto no joelho direito ao subir…',
    aggravates: 'O que agrava?',
    aggravatesPlaceholder: 'Ex.: subir em pé ou pedalar forte',
    symptoms: 'Sintomas durante ou depois do treino',
    symptomsHint: 'Opcional. Se houver sinais importantes, interrompa e procure avaliação profissional.',
    back: 'Voltar',
    saving: 'Salvando…',
    save: 'Salvar e continuar',
  },
  en: {
    current: 'Current condition',
    none: 'No current limitation',
    noneHint: 'I can ride without pain or any known restriction',
    some: 'I have something to consider',
    someHint: 'Pain, injury, condition or movement restriction',
    respect: 'What should we respect?',
    respectIntro: 'These details are optional and help record the context more precisely. They are not a diagnosis.',
    kind: 'Type',
    kinds: [
      ['pain', 'Pain or discomfort'],
      ['injury', 'Injury'],
      ['medical_condition', 'Health condition'],
      ['mobility', 'Movement limitation'],
      ['other', 'Other'],
    ],
    location: 'Location',
    locationPlaceholder: 'E.g.: right knee',
    intensity: 'Perceived intensity',
    notInformed: "Don't say",
    outOf10: (value: number) => `${value} out of 10`,
    started: 'When did it start?',
    clearance: 'Professional guidance recommended',
    clearanceHint: 'Check this if a doctor or physiotherapist should clear you to train',
    restriction: 'There is a current medical restriction',
    restrictionHint: 'The plan should keep the load protected until new guidance',
    surgery: 'I had surgery recently',
    surgeryHint: "Don't start or advance the plan without guidance from whoever is following your recovery.",
    prohibited: 'I was told not to exercise',
    prohibitedHint: 'Cadência does not replace that guidance and keeps a protected reading.',
    condition: 'I have a condition that affects exercise',
    conditionHint: 'Record the context so safety comes before any goal.',
    describe: 'Describe it briefly',
    describePlaceholder: 'E.g.: discomfort in my right knee when climbing…',
    aggravates: 'What makes it worse?',
    aggravatesPlaceholder: 'E.g.: climbing out of the saddle or riding hard',
    symptoms: 'Symptoms during or after training',
    symptomsHint: 'Optional. If there are important signs, stop and get a professional assessment.',
    back: 'Back',
    saving: 'Saving…',
    save: 'Save and continue',
  },
});

export function ProfileStep2({ form }: { form: ProfileFormController }) {
  const t = useMessages(messages);
  const { safetySymptoms } = useMessages(profileText);
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
        <legend>{t.current}</legend>
        <div className="binary-choice">
          <label>
            <input
              type="radio"
              name="has_limitation"
              checked={!hasLimitation}
              onChange={() => setHasLimitation(false)}
            />
            <span>
              <strong>{t.none}</strong>
              <small>{t.noneHint}</small>
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
              <strong>{t.some}</strong>
              <small>{t.someHint}</small>
            </span>
          </label>
        </div>
      </fieldset>
      {hasLimitation &&
        questionIsVisible(questionnaire, 'limitation_kind', {
          has_limitation: true,
        }) && (
          <fieldset>
            <legend>{t.respect}</legend>
            <p className="fieldset-intro">{t.respectIntro}</p>
            <div className="form-grid">
              <div>
                <Label htmlFor="limitation_kind">{t.kind}</Label>
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
                  {t.kinds.map(([value, label]) => (
                    <option key={value} value={value}>
                      {label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <Label htmlFor="limitation_location">{t.location}</Label>
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
                  placeholder={t.locationPlaceholder}
                />
              </div>
              <div>
                <Label htmlFor="limitation_intensity">{t.intensity}</Label>
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
                  <option value="">{t.notInformed}</option>
                  {Array.from({ length: 10 }, (_, index) => (
                    <option key={index + 1} value={index + 1}>
                      {t.outOf10(index + 1)}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <Label htmlFor="limitation_started_on">{t.started}</Label>
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
                  <strong>{t.clearance}</strong>
                  <small>{t.clearanceHint}</small>
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
                  <strong>{t.restriction}</strong>
                  <small>{t.restrictionHint}</small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label={t.surgery}
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
                  <strong>{t.surgery}</strong>
                  <small>{t.surgeryHint}</small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label={t.prohibited}
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
                  <strong>{t.prohibited}</strong>
                  <small>{t.prohibitedHint}</small>
                </span>
              </label>
              <label
                className="clearance-check"
                aria-label={t.condition}
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
                  <strong>{t.condition}</strong>
                  <small>{t.conditionHint}</small>
                </span>
              </label>
            </div>
            <div className="form-grid">
              <div className="textarea-field">
                <Label htmlFor="limitation_description">{t.describe}</Label>
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
                  placeholder={t.describePlaceholder}
                />
              </div>
              <div className="textarea-field">
                <Label htmlFor="limitation_aggravating_movement">{t.aggravates}</Label>
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
                  placeholder={t.aggravatesPlaceholder}
                />
              </div>
            </div>
            <fieldset className="symptoms-field">
              <legend>{t.symptoms}</legend>
              <small>{t.symptomsHint}</small>
              <div className="preference-choice">
                {SAFETY_SYMPTOM_VALUES.map((symptom) => (
                  <label key={symptom}>
                    <input
                      type="checkbox"
                      checked={limitation.symptoms_during_after.includes(symptom)}
                      onChange={() => toggleSafetySymptom(symptom)}
                    />
                    <span>{safetySymptoms[symptom]}</span>
                  </label>
                ))}
              </div>
            </fieldset>
          </fieldset>
        )}
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(1)}>
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
