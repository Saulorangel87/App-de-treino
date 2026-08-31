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
  };
  explanation: {
    summary?: string;
    rules?: string[];
  };
  status: 'planned' | 'in_progress' | 'completed' | 'skipped' | 'adapted';
  session?: WorkoutSession;
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
  };
  workouts: Workout[];
};

export function parseTrainingDate(value: string) {
  return new Date(`${value}T12:00:00`);
}
