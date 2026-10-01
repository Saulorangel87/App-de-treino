// Tipos, constantes e textos da tela de perfil.
import { CalendarDays, Flag, HeartPulse, ShieldAlert } from 'lucide-react';

export type User = {
  display_name: string;
  email: string;
  email_verified: boolean;
};
export type Profile = {
  birth_date?: string | null;
  sex?: string | null;
  height_cm?: number | null;
  weight_kg?: number | null;
  waist_cm?: number | null;
  body_fat_percent?: number | null;
  weight_trend?: string | null;
  experience_level: string;
  activity_level?: string | null;
};
export type Limitation = {
  kind: string;
  description: string;
  location?: string;
  intensity?: number | null;
  aggravating_movement?: string;
  started_on?: string | null;
  symptoms_during_after?: string[];
  medical_restriction: boolean;
  recent_surgery: boolean;
  exercise_prohibited: boolean;
  condition_affecting_exercise: boolean;
  is_active: boolean;
  professional_clearance_recommended: boolean;
};
export type Goal = {
  goal_type: string;
  priority: number;
  target_date?: string | null;
  details: string;
};
export type Availability = {
  weekday: number;
  available_minutes: number;
  preferred_time?: string | null;
  location?: string | null;
};
export type TrainingStatus =
  | 'not_informed'
  | 'regular'
  | 'returning_after_break';
export type CyclingContext = {
  weekly_hours: number;
  practice_duration_months: number;
  average_ride_minutes: number;
  longest_ride_minutes: number;
  weekly_rides: number;
  recent_weekly_distance_km: number;
  recent_training_weeks: number;
  training_status: TrainingStatus;
  recent_best_distance_km: number;
  preferred_session_types: string[];
  discipline: string;
  bike_type: string;
  terrain: string;
  uses_heart_rate: boolean;
  uses_power: boolean;
  max_heart_rate?: number;
  uses_gps: boolean;
  uses_sports_watch: boolean;
  uses_smart_trainer: boolean;
  ftp?: number;
  ftp_test_date?: string;
  ftp_protocol?: string;
  average_power_watts?: number;
  event_goal: boolean;
  event_distance_km?: number;
  event_date?: string;
};
export type Onboarding = {
  limitations: Limitation[];
  goals: Goal[];
  availability: Availability[];
  cycling_context: CyclingContext;
};
export type QuestionnaireStep = {
  id: string;
  title: string;
  description: string;
  question_ids: string[];
};
export type QuestionnaireQuestion = {
  id: string;
  condition?: { question_id: string; equals: unknown };
};
export type Questionnaire = {
  version: string;
  scope: string;
  steps: QuestionnaireStep[];
  questions: QuestionnaireQuestion[];
};

export function questionIsVisible(
  questionnaire: Questionnaire | null,
  questionID: string,
  answers: Record<string, unknown>,
) {
  const question = questionnaire?.questions.find(
    (item) => item.id === questionID,
  );
  if (!question?.condition) return true;
  return answers[question.condition.question_id] === question.condition.equals;
}

export type ProfileForm = {
  birth_date: string;
  sex: string;
  height_cm: string;
  weight_kg: string;
  waist_cm: string;
  body_fat_percent: string;
  weight_trend: string;
  experience_level: string;
  activity_level: string;
};

export const DAYS = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'];
export const initialProfile: ProfileForm = {
  birth_date: '',
  sex: '',
  height_cm: '',
  weight_kg: '',
  waist_cm: '',
  body_fat_percent: '',
  weight_trend: 'not_informed',
  experience_level: 'beginner',
  activity_level: '',
};
export const initialAvailability = (): Availability[] =>
  DAYS.map((_, weekday) => ({
    weekday,
    available_minutes: 0,
    preferred_time: null,
    location: null,
  }));
export const initialCyclingContext: CyclingContext = {
  weekly_hours: 0,
  practice_duration_months: 0,
  average_ride_minutes: 0,
  longest_ride_minutes: 0,
  weekly_rides: 0,
  recent_weekly_distance_km: 0,
  recent_training_weeks: 0,
  training_status: 'not_informed',
  recent_best_distance_km: 0,
  preferred_session_types: [],
  discipline: '',
  bike_type: '',
  terrain: '',
  uses_heart_rate: false,
  uses_power: false,
  uses_gps: false,
  uses_sports_watch: false,
  uses_smart_trainer: false,
  event_goal: false,
};
export const TRAINING_STATUSES = new Set<TrainingStatus>([
  'not_informed',
  'regular',
  'returning_after_break',
]);
export const EXCLUDED_DISCIPLINES = new Set(['dh_enduro', 'track_sprint']);
export const SESSION_PREFERENCES = [
  { value: 'base', label: 'Giro/base' },
  { value: 'cadence', label: 'Cadência' },
  { value: 'hills', label: 'Subidas' },
  { value: 'intervals', label: 'Intervalos' },
  { value: 'threshold', label: 'Limiar' },
  { value: 'sweet_spot', label: 'Sweet spot' },
  { value: 'vo2max', label: 'VO₂max' },
  { value: 'short_intervals', label: 'Intervalos curtos' },
  { value: 'recovery', label: 'Recuperação' },
];
export const SAFETY_SYMPTOMS = [
  { value: 'dizziness', label: 'Tontura' },
  { value: 'unusual_shortness_of_breath', label: 'Falta de ar incomum' },
  { value: 'malaise', label: 'Mal-estar' },
  { value: 'extreme_fatigue', label: 'Fadiga extrema' },
  { value: 'other', label: 'Outro sintoma' },
];

export const stepCopy = [
  {
    kicker: 'PERFIL DO ATLETA · ETAPA 1',
    title: 'Conte-nos onde você está agora.',
    description:
      'Esses dados definem os limites iniciais. Você poderá atualizá-los quando quiser.',
    icon: ShieldAlert,
    asideTitle: 'Uma base segura',
    aside:
      'Experiência e rotina ajudam o Cadência a começar com uma carga compatível com seu momento.',
  },
  {
    kicker: 'SEGURANÇA · ETAPA 2',
    title: 'Existe algo que o treino deve respeitar?',
    description:
      'Dor, lesões e limitações sempre têm prioridade sobre desempenho.',
    icon: HeartPulse,
    asideTitle: 'Segurança em primeiro lugar',
    aside:
      'Uma limitação ativa restringe o que o motor poderá prescrever. O Cadência não realiza diagnóstico médico.',
  },
  {
    kicker: 'DIREÇÃO · ETAPA 3',
    title: 'Onde você quer chegar?',
    description:
      'Defina um objetivo principal e, se desejar, uma prioridade secundária.',
    icon: Flag,
    asideTitle: 'Objetivos realistas',
    aside:
      'O plano combinará sua meta com experiência, segurança e tempo disponível — nunca apenas com ambição.',
  },
  {
    kicker: 'ROTINA · ETAPA 4',
    title: 'Quanto tempo cabe na sua semana?',
    description:
      'Marque os dias possíveis. Descanso também faz parte do plano.',
    icon: CalendarDays,
    asideTitle: 'Consistência vence excesso',
    aside:
      'Usaremos somente os períodos que você informou e reservaremos espaço suficiente para recuperação.',
  },
];
