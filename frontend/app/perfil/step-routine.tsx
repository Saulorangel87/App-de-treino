import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { FormFeedback } from './profile-fields';
import {
  DAYS,
  SESSION_PREFERENCES,
  questionIsVisible,
  type TrainingStatus,
} from './profile-model';

export function ProfileStep4({ form }: { form: ProfileFormController }) {
  const {
    availability,
    completed,
    cyclingContext,
    error,
    goToStep,
    message,
    questionnaire,
    saveAvailability,
    saving,
    setCyclingContext,
    toggleSessionPreference,
    totalMinutes,
    trainingDays,
    updateDay,
  } = form;

  return (
    <form
      onSubmit={saveAvailability}
      className="profile-form availability-form"
    >
      <fieldset>
        <legend>Seu ciclismo hoje</legend>
        <p className="fieldset-intro">
          Essas perguntas são opcionais e ajudam a tornar os próximos treinos
          mais específicos.
        </p>
        <div className="form-grid">
          <div>
            <Label>Modalidade principal</Label>
            <select
              value={cyclingContext.discipline}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  discipline: e.target.value,
                }))
              }
            >
              <option value="">Não informar</option>
              <option value="road">Estrada (speed)</option>
              <option value="mtb_xco">MTB cross-country (XCO)</option>
              <option value="mtb_xcm">MTB maratona (XCM)</option>
              <option value="gravel">Gravel</option>
              <option value="indoor">Indoor/rolo</option>
            </select>
          </div>
          <div>
            <Label>Horas por semana</Label>
            <Input
              type="number"
              min="0"
              max="80"
              step="0.5"
              value={cyclingContext.weekly_hours || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  weekly_hours: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label htmlFor="practice_duration_months">
              Há quanto tempo pedala? (meses)
            </Label>
            <Input
              id="practice_duration_months"
              type="number"
              min="0"
              max="1200"
              value={cyclingContext.practice_duration_months || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  practice_duration_months: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label>Pedais por semana</Label>
            <Input
              type="number"
              min="0"
              max="21"
              value={cyclingContext.weekly_rides || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  weekly_rides: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label>Distância semanal recente (km)</Label>
            <Input
              type="number"
              min="0"
              max="2000"
              step="1"
              value={cyclingContext.recent_weekly_distance_km || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  recent_weekly_distance_km: Number(e.target.value),
                }))
              }
            />
          </div>
          <div className="training-status-field">
            <Label htmlFor="training_status">Situação atual do treino</Label>
            <select
              id="training_status"
              value={cyclingContext.training_status}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  training_status: e.target.value as TrainingStatus,
                }))
              }
            >
              <option value="not_informed">Não informar</option>
              <option value="regular">Estou treinando regularmente</option>
              <option value="returning_after_break">
                Estou retornando após uma pausa
              </option>
            </select>
            <small>
              Esse contexto separa falta de informação, treino regular e
              retomada.
            </small>
          </div>
          <div>
            <Label htmlFor="recent_training_weeks">
              Semanas treinando com regularidade
            </Label>
            <Input
              id="recent_training_weeks"
              type="number"
              min="0"
              max="52"
              value={cyclingContext.recent_training_weeks || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  recent_training_weeks: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label>Maior distância recente (km)</Label>
            <Input
              type="number"
              min="0"
              max="2000"
              step="1"
              value={cyclingContext.recent_best_distance_km || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  recent_best_distance_km: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label>Maior pedal recente (min)</Label>
            <Input
              type="number"
              min="0"
              max="1440"
              value={cyclingContext.longest_ride_minutes || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  longest_ride_minutes: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label htmlFor="average_ride_minutes">
              Duração média do pedal (min)
            </Label>
            <Input
              id="average_ride_minutes"
              type="number"
              min="0"
              max="1440"
              value={cyclingContext.average_ride_minutes || ''}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  average_ride_minutes: Number(e.target.value),
                }))
              }
            />
          </div>
          <div>
            <Label>Tipo de bicicleta</Label>
            <select
              value={cyclingContext.bike_type}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  bike_type: e.target.value,
                }))
              }
            >
              <option value="">Não informar</option>
              <option value="road">Estrada</option>
              <option value="mtb">MTB</option>
              <option value="gravel">Gravel</option>
              <option value="indoor">Indoor/rolo</option>
            </select>
          </div>
          <div>
            <Label>Terreno predominante</Label>
            <select
              value={cyclingContext.terrain}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  terrain: e.target.value,
                }))
              }
            >
              <option value="">Não informar</option>
              <option value="flat">Plano</option>
              <option value="rolling">Misto</option>
              <option value="hilly">Com subidas</option>
            </select>
          </div>
        </div>
        <div className="preference-choice">
          <p>Que tipos de treino você gostaria de fazer?</p>
          <small>
            Opcional. Isso orienta futuras escolhas sem substituir os critérios
            de segurança.
          </small>
          <div>
            {SESSION_PREFERENCES.map((preference) => (
              <label key={preference.value}>
                <input
                  type="checkbox"
                  checked={cyclingContext.preferred_session_types.includes(
                    preference.value,
                  )}
                  onChange={() => toggleSessionPreference(preference.value)}
                />
                <span>{preference.label}</span>
              </label>
            ))}
          </div>
        </div>
        <div className="binary-choice">
          <label aria-label="Uso frequência cardíaca">
            <input
              type="checkbox"
              checked={cyclingContext.uses_heart_rate}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_heart_rate: e.target.checked,
                }))
              }
            />
            <span>
              <strong>Uso frequência cardíaca</strong>
            </span>
          </label>
          <label aria-label="Uso medidor de potência">
            <input
              type="checkbox"
              checked={cyclingContext.uses_power}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_power: e.target.checked,
                  ftp: e.target.checked ? c.ftp : undefined,
                  ftp_test_date: e.target.checked ? c.ftp_test_date : undefined,
                  ftp_protocol: e.target.checked ? c.ftp_protocol : undefined,
                  average_power_watts: e.target.checked
                    ? c.average_power_watts
                    : undefined,
                }))
              }
            />
            <span>
              <strong>Uso medidor de potência</strong>
            </span>
          </label>
          <label aria-label="Estou me preparando para uma prova">
            <input
              type="checkbox"
              checked={cyclingContext.event_goal}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  event_goal: e.target.checked,
                  event_distance_km: e.target.checked
                    ? c.event_distance_km
                    : undefined,
                  event_date: e.target.checked ? c.event_date : undefined,
                }))
              }
            />
            <span>
              <strong>Estou me preparando para uma prova</strong>
            </span>
          </label>
        </div>
        <div className="binary-choice equipment-choice">
          <label aria-label="Uso GPS no pedal">
            <input
              type="checkbox"
              checked={cyclingContext.uses_gps}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_gps: e.target.checked,
                }))
              }
            />
            <span>
              <strong>Uso GPS no pedal</strong>
              <small>Registro percurso, distância ou velocidade.</small>
            </span>
          </label>
          <label aria-label="Uso relógio esportivo">
            <input
              type="checkbox"
              checked={cyclingContext.uses_sports_watch}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_sports_watch: e.target.checked,
                }))
              }
            />
            <span>
              <strong>Uso relógio esportivo</strong>
              <small>Posso acompanhar as métricas durante o pedal.</small>
            </span>
          </label>
          <label aria-label="Uso rolo inteligente">
            <input
              type="checkbox"
              checked={cyclingContext.uses_smart_trainer}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_smart_trainer: e.target.checked,
                }))
              }
            />
            <span>
              <strong>Uso rolo inteligente</strong>
              <small>Tenho sessões indoor com carga controlada.</small>
            </span>
          </label>
        </div>
        {questionIsVisible(questionnaire, 'ftp', {
          uses_power: cyclingContext.uses_power,
        }) && (
          <div className="form-grid power-context-fields">
            <div>
              <Label htmlFor="ftp">FTP (watts)</Label>
              <Input
                id="ftp"
                type="number"
                min="50"
                max="600"
                value={cyclingContext.ftp || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    ftp: Number(e.target.value) || undefined,
                  }))
                }
              />
            </div>
            <div>
              <Label htmlFor="average_power_watts">
                Potência média recente (watts)
              </Label>
              <Input
                id="average_power_watts"
                type="number"
                min="0"
                max="2000"
                value={cyclingContext.average_power_watts || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    average_power_watts: Number(e.target.value) || undefined,
                  }))
                }
              />
            </div>
            <div>
              <Label htmlFor="ftp_test_date">Data do teste FTP</Label>
              <Input
                id="ftp_test_date"
                type="date"
                max={new Date().toISOString().slice(0, 10)}
                value={cyclingContext.ftp_test_date || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    ftp_test_date: e.target.value || undefined,
                  }))
                }
              />
            </div>
            <div>
              <Label htmlFor="ftp_protocol">Protocolo do teste</Label>
              <select
                id="ftp_protocol"
                value={cyclingContext.ftp_protocol || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    ftp_protocol: e.target.value || undefined,
                  }))
                }
              >
                <option value="">Não informar</option>
                <option value="20_minute">20 minutos</option>
                <option value="ramp">Ramp test</option>
                <option value="other">Outro protocolo</option>
              </select>
            </div>
          </div>
        )}
        {questionIsVisible(questionnaire, 'event_distance_km', {
          event_goal: cyclingContext.event_goal,
        }) && (
          <div className="form-grid">
            <div>
              <Label>Distância da prova (km)</Label>
              <Input
                type="number"
                min="1"
                max="2000"
                required
                value={cyclingContext.event_distance_km || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    event_distance_km: Number(e.target.value) || undefined,
                  }))
                }
              />
            </div>
            <div>
              <Label>Data da prova</Label>
              <Input
                type="date"
                required
                value={cyclingContext.event_date || ''}
                min={new Date().toISOString().slice(0, 10)}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    event_date: e.target.value || undefined,
                  }))
                }
              />
            </div>
          </div>
        )}
      </fieldset>
      <fieldset>
        <legend>Dias disponíveis</legend>
        <p className="fieldset-intro">
          Ative um dia e escolha quanto tempo você realmente consegue reservar.
        </p>
        <div className="availability-grid">
          {availability.map((day) => {
            const active = day.available_minutes > 0;
            return (
              <div
                className={`availability-day ${active ? 'active' : ''}`}
                key={day.weekday}
              >
                <button
                  type="button"
                  aria-pressed={active}
                  onClick={() =>
                    updateDay(
                      day.weekday,
                      active
                        ? {
                            available_minutes: 0,
                            preferred_time: null,
                            location: null,
                          }
                        : { available_minutes: 60 },
                    )
                  }
                >
                  <span>{DAYS[day.weekday]}</span>
                  <i>{active && <Check size={12} />}</i>
                </button>
                {active && (
                  <div>
                    <select
                      aria-label={`Duração de ${DAYS[day.weekday]}`}
                      value={day.available_minutes}
                      onChange={(event) =>
                        updateDay(day.weekday, {
                          available_minutes: Number(event.target.value),
                        })
                      }
                    >
                      <option value="30">30 min</option>
                      <option value="45">45 min</option>
                      <option value="60">1 hora</option>
                      <option value="90">1h30</option>
                      <option value="120">2 horas</option>
                      <option value="180">3 horas</option>
                      <option value="240">4 horas</option>
                      <option value="360">6 horas</option>
                      <option value="480">8 horas</option>
                    </select>
                    <select
                      aria-label={`Local de ${DAYS[day.weekday]}`}
                      value={day.location || ''}
                      onChange={(event) =>
                        updateDay(day.weekday, {
                          location: event.target.value || null,
                        })
                      }
                    >
                      <option value="">Qualquer local</option>
                      <option value="outdoor">Rua/estrada</option>
                      <option value="indoor">Rolo/indoor</option>
                      <option value="gym">Academia</option>
                    </select>
                    <Input
                      aria-label={`Horário preferido de ${DAYS[day.weekday]}`}
                      type="time"
                      value={day.preferred_time || ''}
                      onChange={(event) =>
                        updateDay(day.weekday, {
                          preferred_time: event.target.value || null,
                        })
                      }
                    />
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </fieldset>
      <div className="availability-summary">
        <div>
          <strong>{trainingDays}</strong>
          <span>dias possíveis</span>
        </div>
        <div>
          <strong>
            {Math.floor(totalMinutes / 60)}h
            {totalMinutes % 60 ? ` ${totalMinutes % 60}min` : ''}
          </strong>
          <span>por semana</span>
        </div>
        <p>
          O plano poderá usar menos tempo conforme sua recuperação e
          experiência.
        </p>
      </div>
      {completed && (
        <div className="completion-card">
          <span>
            <Check size={20} />
          </span>
          <div>
            <strong>Perfil inicial concluído</strong>
            <p>
              Seus dados estão prontos para orientar a próxima fase: a geração
              do plano.
            </p>
          </div>
        </div>
      )}
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(3)}>
          <ArrowLeft size={15} /> Voltar
        </Button>
        {completed && (
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              window.location.href = '/';
            }}
          >
            Ir para o painel
            <ArrowRight size={16} />
          </Button>
        )}
        <Button
          type="submit"
          disabled={saving || totalMinutes === 0}
          className="profile-submit"
        >
          {saving
            ? 'Salvando…'
            : completed
              ? 'Salvar alterações'
              : 'Concluir perfil'}
          <Check size={16} />
        </Button>
      </div>
    </form>
  );
}
