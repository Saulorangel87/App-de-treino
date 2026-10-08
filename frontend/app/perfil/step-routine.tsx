import type { ProfileFormController } from './use-profile-form';
import { ArrowLeft, ArrowRight, Check } from 'lucide-react';
import { useMessages } from '@/components/locale-provider';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { defineMessages } from '@/lib/i18n';
import { FormFeedback } from './profile-fields';
import {
  SESSION_PREFERENCE_VALUES,
  profileText,
  questionIsVisible,
  todayIso,
  type TrainingStatus,
} from './profile-model';

const messages = defineMessages({
  pt: {
    cycling: 'Seu ciclismo hoje',
    cyclingIntro: 'Essas perguntas são opcionais e ajudam a tornar os próximos treinos mais específicos.',
    discipline: 'Modalidade principal',
    disciplines: [
      ['', 'Não informar'],
      ['road', 'Estrada (speed)'],
      ['mtb_xco', 'MTB cross-country (XCO)'],
      ['mtb_xcm', 'MTB maratona (XCM)'],
      ['gravel', 'Gravel'],
      ['indoor', 'Indoor/rolo'],
    ],
    weeklyHours: 'Horas por semana',
    practiceMonths: 'Há quanto tempo pedala? (meses)',
    weeklyRides: 'Pedais por semana',
    weeklyDistance: 'Distância semanal recente (km)',
    trainingStatus: 'Situação atual do treino',
    trainingStatuses: [
      ['not_informed', 'Não informar'],
      ['regular', 'Estou treinando regularmente'],
      ['returning_after_break', 'Estou retornando após uma pausa'],
    ],
    trainingStatusHint: 'Esse contexto separa falta de informação, treino regular e retomada.',
    trainingWeeks: 'Semanas treinando com regularidade',
    bestDistance: 'Maior distância recente (km)',
    longestRide: 'Maior pedal recente (min)',
    averageRide: 'Duração média do pedal (min)',
    bikeType: 'Tipo de bicicleta',
    bikeTypes: [
      ['', 'Não informar'],
      ['road', 'Estrada'],
      ['mtb', 'MTB'],
      ['gravel', 'Gravel'],
      ['indoor', 'Indoor/rolo'],
    ],
    terrain: 'Terreno predominante',
    terrains: [
      ['', 'Não informar'],
      ['flat', 'Plano'],
      ['rolling', 'Misto'],
      ['hilly', 'Com subidas'],
    ],
    preferences: 'Que tipos de treino você gostaria de fazer?',
    preferencesHint: 'Opcional. Isso orienta futuras escolhas sem substituir os critérios de segurança.',
    usesHeartRate: 'Uso frequência cardíaca',
    usesPower: 'Uso medidor de potência',
    eventGoal: 'Estou me preparando para uma prova',
    usesGps: 'Uso GPS no pedal',
    usesGpsHint: 'Registro percurso, distância ou velocidade.',
    usesWatch: 'Uso relógio esportivo',
    usesWatchHint: 'Posso acompanhar as métricas durante o pedal.',
    usesTrainer: 'Uso rolo inteligente',
    usesTrainerHint: 'Tenho sessões indoor com carga controlada.',
    maxHeartRate: 'Frequência cardíaca máxima (bpm)',
    maxHeartRateHint:
      'Opcional. Com ela, mostramos cada zona de esforço em batimentos. Se você não sabe, deixe em branco: as zonas aparecem pela sensação (teste da conversa). A fórmula 220 menos a idade erra bastante; o melhor é o maior valor visto em um esforço máximo ou em um teste.',
    ftp: 'FTP (watts)',
    averagePower: 'Potência média recente (watts)',
    ftpDate: 'Data do teste FTP',
    ftpProtocol: 'Protocolo do teste',
    ftpProtocols: [
      ['', 'Não informar'],
      ['20_minute', '20 minutos'],
      ['ramp', 'Ramp test'],
      ['other', 'Outro protocolo'],
    ],
    eventDistance: 'Distância da prova (km)',
    eventDate: 'Data da prova',
    days: 'Dias disponíveis',
    daysIntro: 'Ative um dia e escolha quanto tempo você realmente consegue reservar.',
    durationOf: (day: string) => `Duração de ${day}`,
    durations: [
      ['30', '30 min'],
      ['45', '45 min'],
      ['60', '1 hora'],
      ['90', '1h30'],
      ['120', '2 horas'],
      ['180', '3 horas'],
      ['240', '4 horas'],
      ['360', '6 horas'],
      ['480', '8 horas'],
    ],
    locationOf: (day: string) => `Local de ${day}`,
    locations: [
      ['', 'Qualquer local'],
      ['outdoor', 'Rua/estrada'],
      ['indoor', 'Rolo/indoor'],
      ['gym', 'Academia'],
    ],
    timeOf: (day: string) => `Horário preferido de ${day}`,
    possibleDays: 'dias possíveis',
    perWeek: 'por semana',
    summaryNote: 'O plano poderá usar menos tempo conforme sua recuperação e experiência.',
    doneTitle: 'Perfil inicial concluído',
    doneText: 'Seus dados estão prontos para orientar a próxima fase: a geração do plano.',
    back: 'Voltar',
    toDashboard: 'Ir para o painel',
    saving: 'Salvando…',
    saveChanges: 'Salvar alterações',
    finish: 'Concluir perfil',
  },
  en: {
    cycling: 'Your cycling today',
    cyclingIntro: 'These questions are optional and help make your next workouts more specific.',
    discipline: 'Main discipline',
    disciplines: [
      ['', "Don't say"],
      ['road', 'Road'],
      ['mtb_xco', 'MTB cross-country (XCO)'],
      ['mtb_xcm', 'MTB marathon (XCM)'],
      ['gravel', 'Gravel'],
      ['indoor', 'Indoor/trainer'],
    ],
    weeklyHours: 'Hours per week',
    practiceMonths: 'How long have you been riding? (months)',
    weeklyRides: 'Rides per week',
    weeklyDistance: 'Recent weekly distance (km)',
    trainingStatus: 'Current training status',
    trainingStatuses: [
      ['not_informed', "Don't say"],
      ['regular', "I'm training regularly"],
      ['returning_after_break', "I'm coming back after a break"],
    ],
    trainingStatusHint: 'This context separates missing information, regular training and a comeback.',
    trainingWeeks: 'Weeks of regular training',
    bestDistance: 'Longest recent distance (km)',
    longestRide: 'Longest recent ride (min)',
    averageRide: 'Average ride duration (min)',
    bikeType: 'Bike type',
    bikeTypes: [
      ['', "Don't say"],
      ['road', 'Road'],
      ['mtb', 'MTB'],
      ['gravel', 'Gravel'],
      ['indoor', 'Indoor/trainer'],
    ],
    terrain: 'Main terrain',
    terrains: [
      ['', "Don't say"],
      ['flat', 'Flat'],
      ['rolling', 'Mixed'],
      ['hilly', 'Hilly'],
    ],
    preferences: 'What kinds of workouts would you like to do?',
    preferencesHint: 'Optional. It guides future choices without replacing the safety criteria.',
    usesHeartRate: 'I use heart rate',
    usesPower: 'I use a power meter',
    eventGoal: "I'm preparing for a race",
    usesGps: 'I use GPS on my rides',
    usesGpsHint: 'I record route, distance or speed.',
    usesWatch: 'I use a sports watch',
    usesWatchHint: 'I can follow metrics during the ride.',
    usesTrainer: 'I use a smart trainer',
    usesTrainerHint: 'I do indoor sessions with controlled resistance.',
    maxHeartRate: 'Maximum heart rate (bpm)',
    maxHeartRateHint:
      "Optional. With it, we show each effort zone in beats per minute. If you don't know it, leave it blank: zones are shown by feel (the talk test). The 220-minus-age formula is often far off; the best value is the highest you have seen in a maximal effort or a test.",
    ftp: 'FTP (watts)',
    averagePower: 'Recent average power (watts)',
    ftpDate: 'FTP test date',
    ftpProtocol: 'Test protocol',
    ftpProtocols: [
      ['', "Don't say"],
      ['20_minute', '20 minutes'],
      ['ramp', 'Ramp test'],
      ['other', 'Other protocol'],
    ],
    eventDistance: 'Race distance (km)',
    eventDate: 'Race date',
    days: 'Available days',
    daysIntro: 'Turn on a day and choose how much time you can really set aside.',
    durationOf: (day: string) => `Duration on ${day}`,
    durations: [
      ['30', '30 min'],
      ['45', '45 min'],
      ['60', '1 hour'],
      ['90', '1h30'],
      ['120', '2 hours'],
      ['180', '3 hours'],
      ['240', '4 hours'],
      ['360', '6 hours'],
      ['480', '8 hours'],
    ],
    locationOf: (day: string) => `Location on ${day}`,
    locations: [
      ['', 'Anywhere'],
      ['outdoor', 'Street/road'],
      ['indoor', 'Trainer/indoor'],
      ['gym', 'Gym'],
    ],
    timeOf: (day: string) => `Preferred time on ${day}`,
    possibleDays: 'possible days',
    perWeek: 'per week',
    summaryNote: 'The plan may use less time depending on your recovery and experience.',
    doneTitle: 'Starting profile complete',
    doneText: 'Your data is ready to guide the next phase: generating the plan.',
    back: 'Back',
    toDashboard: 'Go to the dashboard',
    saving: 'Saving…',
    saveChanges: 'Save changes',
    finish: 'Finish profile',
  },
});

export function ProfileStep4({ form }: { form: ProfileFormController }) {
  const t = useMessages(messages);
  const { days, sessionPreferences } = useMessages(profileText);
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
        <legend>{t.cycling}</legend>
        <p className="fieldset-intro">{t.cyclingIntro}</p>
        <div className="form-grid">
          <div>
            <Label>{t.discipline}</Label>
            <select
              value={cyclingContext.discipline}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  discipline: e.target.value,
                }))
              }
            >
              {t.disciplines.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <Label>{t.weeklyHours}</Label>
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
            <Label htmlFor="practice_duration_months">{t.practiceMonths}</Label>
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
            <Label>{t.weeklyRides}</Label>
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
            <Label>{t.weeklyDistance}</Label>
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
            <Label htmlFor="training_status">{t.trainingStatus}</Label>
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
              {t.trainingStatuses.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
            <small>{t.trainingStatusHint}</small>
          </div>
          <div>
            <Label htmlFor="recent_training_weeks">{t.trainingWeeks}</Label>
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
            <Label>{t.bestDistance}</Label>
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
            <Label>{t.longestRide}</Label>
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
            <Label htmlFor="average_ride_minutes">{t.averageRide}</Label>
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
            <Label>{t.bikeType}</Label>
            <select
              value={cyclingContext.bike_type}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  bike_type: e.target.value,
                }))
              }
            >
              {t.bikeTypes.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <Label>{t.terrain}</Label>
            <select
              value={cyclingContext.terrain}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  terrain: e.target.value,
                }))
              }
            >
              {t.terrains.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="preference-choice">
          <p>{t.preferences}</p>
          <small>{t.preferencesHint}</small>
          <div>
            {SESSION_PREFERENCE_VALUES.map((preference) => (
              <label key={preference}>
                <input
                  type="checkbox"
                  checked={cyclingContext.preferred_session_types.includes(preference)}
                  onChange={() => toggleSessionPreference(preference)}
                />
                <span>{sessionPreferences[preference]}</span>
              </label>
            ))}
          </div>
        </div>
        <div className="binary-choice">
          <label aria-label={t.usesHeartRate}>
            <input
              type="checkbox"
              checked={cyclingContext.uses_heart_rate}
              onChange={(e) =>
                setCyclingContext((c) => ({
                  ...c,
                  uses_heart_rate: e.target.checked,
                  max_heart_rate: e.target.checked ? c.max_heart_rate : undefined,
                }))
              }
            />
            <span>
              <strong>{t.usesHeartRate}</strong>
            </span>
          </label>
          <label aria-label={t.usesPower}>
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
              <strong>{t.usesPower}</strong>
            </span>
          </label>
          <label aria-label={t.eventGoal}>
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
              <strong>{t.eventGoal}</strong>
            </span>
          </label>
        </div>
        <div className="binary-choice equipment-choice">
          <label aria-label={t.usesGps}>
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
              <strong>{t.usesGps}</strong>
              <small>{t.usesGpsHint}</small>
            </span>
          </label>
          <label aria-label={t.usesWatch}>
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
              <strong>{t.usesWatch}</strong>
              <small>{t.usesWatchHint}</small>
            </span>
          </label>
          <label aria-label={t.usesTrainer}>
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
              <strong>{t.usesTrainer}</strong>
              <small>{t.usesTrainerHint}</small>
            </span>
          </label>
        </div>
        {cyclingContext.uses_heart_rate &&
          questionIsVisible(questionnaire, 'max_heart_rate', {
            uses_heart_rate: cyclingContext.uses_heart_rate,
          }) && (
            <div className="max-heart-rate-field">
              <Label htmlFor="max_heart_rate">{t.maxHeartRate}</Label>
              <Input
                id="max_heart_rate"
                type="number"
                min="100"
                max="230"
                value={cyclingContext.max_heart_rate || ''}
                onChange={(e) =>
                  setCyclingContext((c) => ({
                    ...c,
                    max_heart_rate: Number(e.target.value) || undefined,
                  }))
                }
              />
              <small>{t.maxHeartRateHint}</small>
            </div>
          )}
        {questionIsVisible(questionnaire, 'ftp', {
          uses_power: cyclingContext.uses_power,
        }) && (
          <div className="form-grid power-context-fields">
            <div>
              <Label htmlFor="ftp">{t.ftp}</Label>
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
              <Label htmlFor="average_power_watts">{t.averagePower}</Label>
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
              <Label htmlFor="ftp_test_date">{t.ftpDate}</Label>
              <Input
                id="ftp_test_date"
                type="date"
                max={todayIso()}
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
              <Label htmlFor="ftp_protocol">{t.ftpProtocol}</Label>
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
              {t.ftpProtocols.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
              </select>
            </div>
          </div>
        )}
        {questionIsVisible(questionnaire, 'event_distance_km', {
          event_goal: cyclingContext.event_goal,
        }) && (
          <div className="form-grid">
            <div>
              <Label>{t.eventDistance}</Label>
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
              <Label>{t.eventDate}</Label>
              <Input
                type="date"
                required
                value={cyclingContext.event_date || ''}
                min={todayIso()}
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
        <legend>{t.days}</legend>
        <p className="fieldset-intro">{t.daysIntro}</p>
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
                  <span>{days[day.weekday]}</span>
                  <i>{active && <Check size={12} />}</i>
                </button>
                {active && (
                  <div>
                    <select
                      aria-label={t.durationOf(days[day.weekday])}
                      value={day.available_minutes}
                      onChange={(event) =>
                        updateDay(day.weekday, {
                          available_minutes: Number(event.target.value),
                        })
                      }
                    >
              {t.durations.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
                    </select>
                    <select
                      aria-label={t.locationOf(days[day.weekday])}
                      value={day.location || ''}
                      onChange={(event) =>
                        updateDay(day.weekday, {
                          location: event.target.value || null,
                        })
                      }
                    >
              {t.locations.map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
                    </select>
                    <Input
                      aria-label={t.timeOf(days[day.weekday])}
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
          <span>{t.possibleDays}</span>
        </div>
        <div>
          <strong>
            {Math.floor(totalMinutes / 60)}h
            {totalMinutes % 60 ? ` ${totalMinutes % 60}min` : ''}
          </strong>
          <span>{t.perWeek}</span>
        </div>
        <p>{t.summaryNote}</p>
      </div>
      {completed && (
        <div className="completion-card">
          <span>
            <Check size={20} />
          </span>
          <div>
            <strong>{t.doneTitle}</strong>
            <p>{t.doneText}</p>
          </div>
        </div>
      )}
      <FormFeedback error={error} message={message} />
      <div className="form-actions">
        <Button type="button" variant="outline" onClick={() => goToStep(3)}>
          <ArrowLeft size={15} /> {t.back}
        </Button>
        {completed && (
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              window.location.href = '/';
            }}
          >
            {t.toDashboard}
            <ArrowRight size={16} />
          </Button>
        )}
        <Button
          type="submit"
          disabled={saving || totalMinutes === 0}
          className="profile-submit"
        >
          {saving ? t.saving : completed ? t.saveChanges : t.finish}
          <Check size={16} />
        </Button>
      </div>
    </form>
  );
}
