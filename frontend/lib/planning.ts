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
    event_taper_applied?: boolean;
    data_integrity?: {
      version: 'data-integrity-v1';
      mode: 'observation';
      scope: 'completed_workout';
      assessed_at: string;
      status: 'valid' | 'incomplete' | 'inconsistent' | 'not_evaluated';
      rules_evaluated: string[];
      reasons: { code: string; message: string }[];
      missing_data: string[];
      data_issues: string[];
      not_evaluated: string[];
      eligible_for_history: boolean;
      progression_eligible: false;
      used_for_prescription: false;
    };
    adaptation?: {
      kind: 'safety' | 'recovery' | 'progression';
      reason: string;
      safety_notice?: string;
      source_workout_id: string;
      previous_duration_minutes: number;
      previous_target_rpe: number;
    };
    adaptation_shadow?: {
      version: 'rules-v2-adaptation-v1';
      mode: 'shadow';
      scope: 'post_workout_feedback';
      assessed_at: string;
      status: 'not_evaluated' | 'protective_signal' | 'observation_only';
      candidate_response:
        | 'not_evaluated'
        | 'prefer_recovery'
        | 'maintain_observed'
        | 'defer_progression'
        | 'progress_duration_5pct';
      rules_evaluated: string[];
      rules_deferred: string[];
      reasons: { code: string; message: string }[];
      missing_data: string[];
      data_issues: string[];
      not_evaluated: string[];
      stimulus_distribution?: TrainingStimulusDistribution;
      post_workout_context?: {
        version: 'post-workout-context-v2';
        mode: 'observation';
        scope: 'post_workout_feedback';
        assessed_at: string;
        status: 'not_evaluated' | 'observed';
        candidate_response: 'not_evaluated' | 'maintain_observed';
        recovery_after?: number;
        repeat_confidence?: number;
        satisfaction?: number;
        terrain?: 'flat' | 'rolling' | 'hilly' | 'mixed' | 'technical' | 'indoor';
        external_conditions?: 'normal' | 'heat' | 'cold' | 'wind' | 'rain' | 'poor_visibility' | 'other';
        observed_fields: string[];
        reasons: { code: string; message: string }[];
        missing_data: string[];
        data_issues: string[];
        not_evaluated: string[];
        progression_eligible: false;
        used_for_prescription: false;
      };
      decision_audit?: {
        version: 'adaptation-audit-v1';
        mode: 'observation';
        scope: 'post_workout_feedback';
        assessed_at: string;
        data_used: string[];
        missing_data: string[];
        constraints_applied: string[];
        alternatives_rejected: string[];
        conditions_for_change: string[];
        confidence: 'not_calibrated';
        not_evaluated: string[];
        used_for_prescription: false;
      };
      planned_vs_actual?: {
        version: 'planned-vs-actual-v2';
        mode: 'observation';
        scope: 'completed_workout';
        assessed_at: string;
        status: 'not_evaluated' | 'observed';
      planned_duration_minutes: number;
      actual_duration_minutes: number;
      duration_delta_minutes?: number;
      duration_completion_percent?: number;
      completion_status: 'complete' | 'partial';
      partial_reason?:
        | 'time_available_changed'
        | 'fatigue_or_recovery'
        | 'pain_or_discomfort'
        | 'equipment_or_conditions'
        | 'other';
      target_rpe: number;
        actual_rpe: number;
        rpe_delta?: number;
      recovery_after?: number;
      repeat_confidence?: number;
      satisfaction?: number;
      terrain?: 'flat' | 'rolling' | 'hilly' | 'mixed' | 'technical' | 'indoor';
      external_conditions?: 'normal' | 'heat' | 'cold' | 'wind' | 'rain' | 'poor_visibility' | 'other';
        observed_fields: string[];
        reasons: { code: string; message: string }[];
        missing_data: string[];
        data_issues: string[];
        not_evaluated: string[];
        progression_eligible: false;
        used_for_prescription: false;
      };
      load_tolerance?: {
        version: 'load-tolerance-v1';
        mode: 'observation';
        scope: 'completed_workout';
        assessed_at: string;
        status: 'not_evaluated' | 'protective_signal' | 'observation_only';
        candidate_response: 'not_evaluated' | 'prefer_recovery' | 'maintain_observed';
        evidence_periods: string[];
        rules_evaluated: string[];
        reasons: { code: string; message: string }[];
        missing_data: string[];
        data_issues: string[];
        not_evaluated: string[];
        progression_eligible: false;
        applied: false;
        used_for_prescription: false;
      };
      progression_eligible: false;
      applied: false;
      used_for_prescription: false;
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
  completion_status: 'complete' | 'partial';
  partial_reason?:
    | 'time_available_changed'
    | 'fatigue_or_recovery'
    | 'pain_or_discomfort'
    | 'equipment_or_conditions'
    | 'other';
  difficulty: 'very_easy' | 'easy' | 'moderate' | 'hard' | 'very_hard';
  pain_reported: boolean;
  fatigue_after: number;
  recovery_after?: number;
  repeat_confidence?: number;
  satisfaction?: number;
  terrain?: 'flat' | 'rolling' | 'hilly' | 'mixed' | 'technical' | 'indoor';
  external_conditions?: 'normal' | 'heat' | 'cold' | 'wind' | 'rain' | 'poor_visibility' | 'other';
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

export type RulesV2ShadowAssessment = {
  version: 'rules-v2';
  mode: 'shadow';
  scope: 'plan_generation_only';
  assessed_at: string;
  status: 'not_evaluated' | 'protective_signal' | 'observation_only';
  candidate_response: 'not_evaluated' | 'prefer_recovery' | 'maintain_observed';
  rules_evaluated: string[];
  rules_deferred: string[];
  reasons: Array<{ code: string; message: string }>;
  missing_data: string[];
  data_issues: string[];
  not_evaluated: string[];
  stimulus_distribution?: TrainingStimulusDistribution;
  progression_eligible: false;
  applied: false;
  used_for_prescription: false;
};

export type PeriodizationWeekObservation = {
  week_index: number;
  phase: 'progression' | 'recovery';
  planned_sessions: number;
  total_planned_minutes: number;
  quality_sessions: number;
  high_intensity_sessions: number;
  recovery_sessions: number;
  long_sessions: number;
  tapered_sessions: number;
  average_target_rpe: number | null;
  minimum_days_between_quality_sessions: number | null;
};

export type PeriodizationShadowAssessment = {
  version: 'periodization-shadow-v1';
  mode: 'shadow';
  scope: 'plan_generation_only';
  assessed_at: string;
  status: 'observed' | 'not_evaluated';
  candidate_response: 'maintain_observed' | 'defer_evaluation';
  rules_evaluated: string[];
  reasons: Array<{ code: string; message: string }>;
  missing_data: string[];
  data_issues: string[];
  not_evaluated: string[];
  weeks: PeriodizationWeekObservation[];
  total_planned_minutes: number;
  quality_sessions: number;
  high_intensity_sessions: number;
  adjacent_quality_session_pairs: number;
  minimum_days_between_quality_sessions: number | null;
  progression_eligible: false;
  applied: false;
  used_for_prescription: false;
};

export type StimulusSelectionShadowAssessment = {
  version: 'stimulus-selection-shadow-v1';
  mode: 'shadow';
  scope: 'plan_generation_only';
  assessed_at: string;
  status: 'observed' | 'observed_mismatch' | 'not_evaluated';
  candidate_response: 'maintain_observed' | 'review_selection' | 'defer_evaluation';
  candidate_need: string;
  expected_stimuli: string[];
  selected_stimuli: string[];
  rules_evaluated: string[];
  reasons: Array<{ code: string; message: string }>;
  missing_data: string[];
  data_issues: string[];
  not_evaluated: string[];
  progression_eligible: false;
  applied: false;
  used_for_prescription: false;
};

export type PlanningCoherenceComponentObservation = {
  component: 'periodization' | 'stimulus_distribution' | 'stimulus_selection';
  version: string;
  mode: 'shadow' | 'observation';
  status: string;
  candidate_response?: string;
};

export type PlanningCoherenceShadowAssessment = {
  version: 'planning-coherence-shadow-v1';
  mode: 'shadow';
  scope: 'plan_generation_only';
  assessed_at: string;
  status: 'observed' | 'observed_mismatch' | 'not_evaluated';
  candidate_response: 'maintain_observed' | 'review_coherence' | 'defer_evaluation';
  components: PlanningCoherenceComponentObservation[];
  candidate_need: string;
  coherence_checks: string[];
  periodization_recovery_week_quality_sessions: number;
  periodization_recovery_week_high_intensity_sessions: number;
  distribution_quality_sessions_last_7d: number;
  distribution_adjacent_quality_session_pairs: number;
  rules_evaluated: string[];
  reasons: Array<{ code: string; message: string }>;
  missing_data: string[];
  data_issues: string[];
  not_evaluated: string[];
  progression_eligible: false;
  applied: false;
  used_for_prescription: false;
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
  sessions_with_satisfaction?: number;
  sessions_with_terrain?: number;
  sessions_with_external_conditions?: number;
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
  sessions_with_satisfaction: number;
  sessions_with_terrain: number;
  sessions_with_external_conditions: number;
  pain_reported_sessions: number;
  high_fatigue_sessions: number;
  above_target_rpe_sessions: number;
  quality_sessions: number;
  quality_sessions_with_load: number;
  quality_performed_minutes: number;
  quality_density_percent: number | null;
  high_intensity_sessions: number;
  recovery_checkins: number;
  complete_recovery_checkins: number;
  checkins_with_protective_signal: number;
  recovery_needed_checkins: number;
};

export type TrainingHistoryPeriodComparison = {
  version: 'period-comparison-v1' | 'period-comparison-v2' | 'period-comparison-v3';
  mode: 'observation';
  basis: 'six_non_overlapping_7_day_periods_by_database_clock';
  periods: TrainingHistoryPeriod[];
  missing_data: string[];
  data_issues: string[];
  used_for_prescription: false;
};

export type TrainingStimulusDistribution = {
  version: 'stimulus-distribution-v1';
  mode: 'observation';
  basis: 'eligible_completed_sessions_by_completed_at_utc_day';
  quality_target_rpe_threshold: number;
  high_intensity_rpe_threshold: number;
  quality_sessions_last_7d: number;
  quality_sessions_last_14d: number;
  quality_sessions_last_28d: number;
  quality_sessions_last_42d: number;
  high_intensity_sessions_last_42d: number;
  adjacent_quality_session_pairs: number;
  minimum_days_between_quality_sessions: number | null;
  latest_quality_session_at: string | null;
  days_since_latest_quality_session: number | null;
  missing_data: string[];
  data_issues: string[];
  used_for_prescription: false;
};

export type TrainingHistorySnapshot = {
  version: 'training-history-v1' | 'training-history-v2' | 'training-history-v3' | 'training-history-v4' | 'training-history-v5';
  mode: 'observation';
  captured_at: string;
  load_method: 'duration_minutes_x_actual_rpe';
  load_unit: 'session_rpe_arbitrary_units';
  adherence_basis: string;
  completion_time_basis: string;
  evidence_keys: string[];
  windows: TrainingHistoryWindow[];
  period_comparison?: TrainingHistoryPeriodComparison;
  stimulus_distribution?: TrainingStimulusDistribution;
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
    event_taper?: EventTaperAssessment;
    rules_v2_shadow?: RulesV2ShadowAssessment;
    periodization_shadow?: PeriodizationShadowAssessment;
    stimulus_selection_shadow?: StimulusSelectionShadowAssessment;
    planning_coherence_shadow?: PlanningCoherenceShadowAssessment;
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
      training_status?: 'not_informed' | 'regular' | 'returning_after_break';
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

export type EventTaperAssessment = {
  version: 'taper-v1';
  mode: 'prescriptive';
  scope: 'event_based_plan_volume';
  assessed_at: string;
  status: 'not_applicable' | 'eligible';
  applied: boolean;
  used_for_prescription: boolean;
  event_date?: string;
  days_from_today?: number;
  volume_multiplier: number;
  evidence_keys: string[];
  rules_evaluated: string[];
  reasons: Array<{ code: string; message: string }>;
};

export function parseTrainingDate(value: string) {
  return new Date(`${value}T12:00:00`);
}
