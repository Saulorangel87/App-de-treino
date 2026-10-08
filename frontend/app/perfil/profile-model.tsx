// Tipos, constantes e textos da tela de perfil.
import { CalendarDays, Flag, HeartPulse, ShieldAlert } from 'lucide-react';
import { defineMessages } from '@/lib/i18n';

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
  Array.from({ length: 7 }, (_, weekday) => ({
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
export const SESSION_PREFERENCE_VALUES = [
  'base',
  'cadence',
  'hills',
  'intervals',
  'threshold',
  'sweet_spot',
  'vo2max',
  'short_intervals',
  'recovery',
] as const;
export const SAFETY_SYMPTOM_VALUES = [
  'dizziness',
  'unusual_shortness_of_breath',
  'malaise',
  'extreme_fatigue',
  'other',
] as const;

export const STEP_ICONS = [ShieldAlert, HeartPulse, Flag, CalendarDays];

type StepCopy = { kicker: string; title: string; description: string; asideTitle: string; aside: string };

// Textos fixos do perfil. Título e descrição de cada etapa podem vir do
// questionário da API, que já responde no idioma pedido.
export const profileText = defineMessages({
  pt: {
    days: ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb'],
    sessionPreferences: {
      base: 'Giro/base',
      cadence: 'Cadência',
      hills: 'Subidas',
      intervals: 'Intervalos',
      threshold: 'Limiar',
      sweet_spot: 'Sweet spot',
      vo2max: 'VO₂max',
      short_intervals: 'Intervalos curtos',
      recovery: 'Recuperação',
    } as Record<(typeof SESSION_PREFERENCE_VALUES)[number], string>,
    safetySymptoms: {
      dizziness: 'Tontura',
      unusual_shortness_of_breath: 'Falta de ar incomum',
      malaise: 'Mal-estar',
      extreme_fatigue: 'Fadiga extrema',
      other: 'Outro sintoma',
    } as Record<(typeof SAFETY_SYMPTOM_VALUES)[number], string>,
    goals: {
      health: 'Melhorar saúde e bem-estar',
      fitness: 'Ganhar condicionamento',
      endurance: 'Pedalar por mais tempo',
      performance: 'Aumentar meu desempenho',
      event: 'Preparar para um evento',
      weight_management: 'Apoiar o controle de peso',
    } as Record<string, string>,
    steps: [
      {
        kicker: 'PERFIL DO ATLETA · ETAPA 1',
        title: 'Conte-nos onde você está agora.',
        description: 'Esses dados definem os limites iniciais. Você poderá atualizá-los quando quiser.',
        asideTitle: 'Uma base segura',
        aside: 'Experiência e rotina ajudam o Cadência a começar com uma carga compatível com seu momento.',
      },
      {
        kicker: 'SEGURANÇA · ETAPA 2',
        title: 'Existe algo que o treino deve respeitar?',
        description: 'Dor, lesões e limitações sempre têm prioridade sobre desempenho.',
        asideTitle: 'Segurança em primeiro lugar',
        aside: 'Uma limitação ativa restringe o que o motor poderá prescrever. O Cadência não realiza diagnóstico médico.',
      },
      {
        kicker: 'DIREÇÃO · ETAPA 3',
        title: 'Onde você quer chegar?',
        description: 'Defina um objetivo principal e, se desejar, uma prioridade secundária.',
        asideTitle: 'Objetivos realistas',
        aside: 'O plano combinará sua meta com experiência, segurança e tempo disponível — nunca apenas com ambição.',
      },
      {
        kicker: 'ROTINA · ETAPA 4',
        title: 'Quanto tempo cabe na sua semana?',
        description: 'Marque os dias possíveis. Descanso também faz parte do plano.',
        asideTitle: 'Consistência vence excesso',
        aside: 'Usaremos somente os períodos que você informou e reservaremos espaço suficiente para recuperação.',
      },
    ] as StepCopy[],
    loadFailed: 'Não foi possível carregar seu perfil. Verifique se a API está em execução e tente novamente.',
    availabilitySaved: 'Disponibilidade salva. Abra Meu plano e escolha Atualizar plano para gerar um rascunho com esta nova rotina.',
    profileDone: 'Perfil concluído. Seu contexto inicial está salvo com segurança.',
    saveFailed: 'Não foi possível salvar esta etapa.',
    openLocalLink: 'Abra o link local abaixo.',
    resendFailed: 'Não foi possível reenviar a confirmação.',
  },
  en: {
    days: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
    sessionPreferences: {
      base: 'Easy/base',
      cadence: 'Cadence',
      hills: 'Climbs',
      intervals: 'Intervals',
      threshold: 'Threshold',
      sweet_spot: 'Sweet spot',
      vo2max: 'VO₂max',
      short_intervals: 'Short intervals',
      recovery: 'Recovery',
    },
    safetySymptoms: {
      dizziness: 'Dizziness',
      unusual_shortness_of_breath: 'Unusual shortness of breath',
      malaise: 'Feeling unwell',
      extreme_fatigue: 'Extreme fatigue',
      other: 'Other symptom',
    },
    goals: {
      health: 'Improve health and well-being',
      fitness: 'Get fitter',
      endurance: 'Ride for longer',
      performance: 'Boost my performance',
      event: 'Prepare for an event',
      weight_management: 'Support weight management',
    },
    steps: [
      {
        kicker: 'ATHLETE PROFILE · STEP 1',
        title: 'Tell us where you are right now.',
        description: 'This data sets the starting limits. You can update it whenever you want.',
        asideTitle: 'A safe base',
        aside: 'Experience and routine help Cadência start with a load that fits where you are now.',
      },
      {
        kicker: 'SAFETY · STEP 2',
        title: 'Is there anything your training should respect?',
        description: 'Pain, injuries and limitations always take priority over performance.',
        asideTitle: 'Safety first',
        aside: 'An active limitation restricts what the engine can prescribe. Cadência does not provide medical diagnosis.',
      },
      {
        kicker: 'DIRECTION · STEP 3',
        title: 'Where do you want to get to?',
        description: 'Set a main goal and, if you want, a secondary priority.',
        asideTitle: 'Realistic goals',
        aside: 'The plan will combine your goal with experience, safety and available time — never with ambition alone.',
      },
      {
        kicker: 'ROUTINE · STEP 4',
        title: 'How much time fits in your week?',
        description: 'Mark the days you can ride. Rest is part of the plan too.',
        asideTitle: 'Consistency beats excess',
        aside: 'We will only use the times you entered and leave enough room for recovery.',
      },
    ],
    loadFailed: 'Your profile could not be loaded. Check that the API is running and try again.',
    availabilitySaved: 'Availability saved. Open My plan and choose Update plan to generate a draft with this new routine.',
    profileDone: 'Profile complete. Your starting context is saved securely.',
    saveFailed: 'This step could not be saved.',
    openLocalLink: 'Open the local link below.',
    resendFailed: 'The confirmation could not be resent.',
  },
});

/** Hoje no formato AAAA-MM-DD, para limites de campos de data. */
export function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}
