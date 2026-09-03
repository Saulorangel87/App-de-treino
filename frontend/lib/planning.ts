export type Workout = {
  id: string;
  scheduled_on: string;
  name: string;
  objective: string;
  duration_minutes: number;
  target_rpe: number;
  structure: {
    warmup_minutes?: number;
    main?: string;
    cooldown_minutes?: number;
    steps?: WorkoutStep[];
  };
  explanation: {
    summary?: string;
    rules?: string[];
    evidence_keys?: string[];
    adaptation?: {
      kind: 'safety' | 'recovery' | 'progression';
      reason: string;
      safety_notice?: string;
      source_workout_id: string;
      previous_duration_minutes: number;
      previous_target_rpe: number;
    };
  };
  status: 'planned' | 'in_progress' | 'completed' | 'skipped' | 'adapted';
  session?: WorkoutSession;
};

export type WorkoutStep = {
  order: number;
  kind: string;
  title: string;
  duration_minutes: number;
  target_rpe: number;
  instruction: string;
};

export type WorkoutFeedback = {
  difficulty: 'very_easy' | 'easy' | 'moderate' | 'hard' | 'very_hard';
  pain_reported: boolean;
  fatigue_after: number;
  notes?: string;
};

export type WorkoutSession = {
  id: string;
  status: 'in_progress' | 'completed' | 'cancelled';
  started_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  duration_minutes?: number;
  actual_rpe?: number;
  distance_km?: number;
  elevation_gain_m?: number;
  average_power_watts?: number;
  average_heart_rate?: number;
  feedback?: WorkoutFeedback;
};

export type Activity = {
  id: string;
  workout_id: string;
  name: string;
  objective: string;
  scheduled_on: string;
  status: 'completed' | 'cancelled';
  started_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  duration_minutes?: number;
  actual_rpe?: number;
  distance_km?: number;
  elevation_gain_m?: number;
  average_power_watts?: number;
  average_heart_rate?: number;
  feedback?: WorkoutFeedback;
};

export type TrainingPlan = {
  id: string;
  starts_on: string;
  ends_on: string;
  status: 'draft' | 'active' | 'completed' | 'cancelled';
  prescription_snapshot: {
    engine_version?: string;
    experience_level?: string;
    primary_goal?: string;
    restricted?: boolean;
    sessions_per_week?: number;
    cycling_context?: {
      uses_heart_rate?: boolean;
      uses_power?: boolean;
    };
  };
  workouts: Workout[];
  evidence?: ScientificSource[];
};

export type ScientificSource = {
  source_key: string; title: string; authors: string; published_year: number;
  url: string; training_focus: string; evidence_level: string; summary: string;
};

export function parseTrainingDate(value: string) {
  return new Date(`${value}T12:00:00`);
}
