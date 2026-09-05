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
    protocol_key?: string;
    steps?: WorkoutStep[];
  };
  explanation: {
    summary?: string;
    rules?: string[];
    protocol_key?: string;
    evidence_scope?: string;
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

export type WorkoutExplanationResponse = {
  explanation: string;
  source: 'rules' | 'rules_fallback' | 'ollama';
  ai_enabled: boolean;
  warning?: string;
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

export type ReadinessAssessment = {
  classifier_version: 'readiness-v1';
  mode: 'observation';
  scope: 'observed_history_28d';
  assessed_at: string;
  status: 'insufficient_data' | 'caution' | 'recovery_needed' | 'stable';
  reasons: Array<{ code: string; message: string }>;
  missing_data: string[];
  not_evaluated: string[];
  progression_eligible: false;
  active_limitations: number;
  data_coverage?: {
    sessions_with_duration: number;
    sessions_with_rpe: number;
    sessions_with_feedback: number;
    complete_sessions: number;
    recovery_with_fatigue: number;
  };
};

export type TrainingHistoryWindow = {
  window_days: 7 | 28 | 42;
  expected_sessions: number;
  scheduled_completed_sessions: number;
  cancelled_sessions: number;
  missed_sessions: number;
  overdue_in_progress_sessions: number;
  completion_rate_percent: number | null;
  performed_sessions: number;
  performed_minutes: number;
  sessions_with_session_rpe_load: number;
  sessions_without_session_rpe_load: number;
  session_rpe_load: number;
  feedback_records?: number;
  sessions_with_complete_feedback?: number;
  pain_reported_sessions?: number;
  high_fatigue_sessions?: number;
  above_target_rpe_sessions?: number;
  recovery_checkins?: number;
  complete_recovery_checkins?: number;
  checkins_with_protective_signal?: number;
  recovery_needed_checkins?: number;
};

export type TrainingHistoryPeriod = {
  period_index: number;
  period_key: string;
  period_days: number;
  expected_sessions: number;
  scheduled_completed_sessions: number;
  cancelled_sessions: number;
  missed_sessions: number;
  overdue_in_progress_sessions: number;
  completion_rate_percent: number | null;
  performed_sessions: number;
  performed_minutes: number;
  sessions_with_session_rpe_load: number;
  sessions_without_session_rpe_load: number;
  session_rpe_load: number;
  feedback_records: number;
  sessions_with_complete_feedback: number;
  pain_reported_sessions: number;
  high_fatigue_sessions: number;
  above_target_rpe_sessions: number;
  recovery_checkins: number;
  complete_recovery_checkins: number;
  checkins_with_protective_signal: number;
  recovery_needed_checkins: number;
};

export type TrainingHistoryPeriodComparison = {
  version: 'period-comparison-v1';
  mode: 'observation';
  basis: 'six_non_overlapping_7_day_periods_by_database_clock';
  periods: TrainingHistoryPeriod[];
  missing_data: string[];
  data_issues: string[];
  used_for_prescription: false;
};

export type TrainingHistorySnapshot = {
  version: 'training-history-v1' | 'training-history-v2' | 'training-history-v3';
  mode: 'observation';
  captured_at: string;
  load_method: 'duration_minutes_x_actual_rpe';
  load_unit: 'session_rpe_arbitrary_units';
  adherence_basis: string;
  completion_time_basis: string;
  evidence_keys: string[];
  windows: TrainingHistoryWindow[];
  period_comparison?: TrainingHistoryPeriodComparison;
  temporal_quality?: {
    athlete_timezone_available: false;
    latest_completed_at: string | null;
    days_since_latest_completed: number | null;
    latest_session_rpe_load_at: string | null;
    days_since_latest_session_rpe_load: number | null;
    latest_recovery_recorded_on: string | null;
    days_since_latest_recovery_checkin: number | null;
    future_completed_sessions_excluded: number;
    future_recovery_checkins_excluded: number;
    app_recording_gap_interpretation: 'recorded_activity_gap_only_not_confirmed_training_cessation';
  };
  missing_data: string[];
  data_issues: string[];
  not_evaluated?: string[];
  used_for_prescription: false;
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
    readiness_assessment?: ReadinessAssessment;
    training_history?: TrainingHistorySnapshot;
    observed_training?: {
      window_days?: number;
      completed_sessions?: number;
      completed_minutes?: number;
      average_rpe?: number;
      average_fatigue?: number;
      pain_reported?: boolean;
      recovery_checkins?: number;
      average_recovery_fatigue?: number;
      requires_recovery?: boolean;
    };
    cycling_context?: {
      discipline?: string;
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
