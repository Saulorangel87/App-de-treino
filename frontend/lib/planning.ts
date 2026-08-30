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
  status: string;
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
