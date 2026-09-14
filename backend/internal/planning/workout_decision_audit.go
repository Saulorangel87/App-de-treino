package planning

const workoutDecisionAuditVersion = "workout-decision-audit-v1"

// WorkoutDecisionAudit is the structured, per-workout explanation produced
// together with rules-v1. It records the context that shaped a prescription;
// it is not a second prescribing engine.
type WorkoutDecisionAudit struct {
	Version              string   `json:"version"`
	Mode                 string   `json:"mode"`
	Scope                string   `json:"scope"`
	RulesEvaluated       []string `json:"rules_evaluated"`
	RulesApplied         []string `json:"rules_applied"`
	AlternativesRejected []string `json:"alternatives_rejected"`
	DataUsed             []string `json:"data_used"`
	MissingData          []string `json:"missing_data"`
	ConstraintsApplied   []string `json:"constraints_applied"`
	ConditionsForChange  []string `json:"conditions_for_change"`
	Confidence           string   `json:"confidence"`
	UsedForPrescription  bool     `json:"used_for_prescription"`
}

func buildWorkoutDecisionAudit(input Context, kind string, rules []string, restricted, observedProtected, returningAfterPause, recoveryWeek, eventTaperApplied, postEventRecoveryApplied bool) WorkoutDecisionAudit {
	audit := WorkoutDecisionAudit{
		Version: workoutDecisionAuditVersion,
		Mode:    "deterministic",
		Scope:   "plan_generation_workout",
		RulesEvaluated: []string{
			"availability_gate",
			"experience_complexity_gate",
			"safety_precedence_gate",
			"recovery_precedence_gate",
			"weekly_periodization_gate",
			"protocol_eligibility_gate",
		},
		RulesApplied:         append([]string(nil), rules...),
		AlternativesRejected: []string{},
		DataUsed:             []string{"availability_minutes", "experience_level", "primary_goal", "cycling_context"},
		MissingData:          []string{},
		ConstraintsApplied:   []string{},
		ConditionsForChange:  []string{"novo_feedback_valido", "mudanca_de_disponibilidade", "novo_sinal_de_seguranca", "mudanca_no_contexto_do_evento"},
		Confidence:           "rule_based_not_calibrated",
		UsedForPrescription:  true,
	}

	if input.Observed.HasData() {
		audit.DataUsed = append(audit.DataUsed, "observed_training_28d")
	} else {
		audit.MissingData = append(audit.MissingData, "observed_training_28d")
	}
	if input.Cycling.EventGoal && input.Cycling.EventDate != nil {
		audit.DataUsed = append(audit.DataUsed, "event_goal", "event_date")
	} else {
		audit.MissingData = append(audit.MissingData, "event_goal_or_date")
	}
	if input.Cycling.UsesHeartRate {
		audit.DataUsed = append(audit.DataUsed, "heart_rate_sensor")
	} else {
		audit.MissingData = append(audit.MissingData, "heart_rate_sensor")
	}
	if input.Cycling.UsesPower && input.Cycling.FTP != nil {
		audit.DataUsed = append(audit.DataUsed, "power_meter", "ftp")
	} else {
		audit.MissingData = append(audit.MissingData, "power_meter_or_ftp")
	}
	if !input.BaselineEligible {
		audit.MissingData = append(audit.MissingData, "eligible_submaximal_assessment")
	}

	if restricted {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "active_safety_limitation")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "higher_intensity_protocols")
	}
	if observedProtected {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "recent_recovery_or_pain_signal")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "quality_session")
	}
	if returningAfterPause {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "return_after_break")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "quality_session", "long_session_above_45_minutes")
	}
	if recoveryWeek {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "recovery_week")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "additional_quality_session")
	}
	if hasLowObservedAdherence(input.TrainingHistory) {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "low_observed_adherence")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "quality_session")
	}
	if eventTaperApplied {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "event_taper")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "volume_progression")
	}
	if postEventRecoveryApplied {
		audit.ConstraintsApplied = append(audit.ConstraintsApplied, "post_event_recovery")
		audit.AlternativesRejected = append(audit.AlternativesRejected, "quality_session", "long_session")
	}
	if kind != "quality" {
		audit.AlternativesRejected = append(audit.AlternativesRejected, "additional_quality_session")
	}
	audit.DataUsed = uniqueWorkoutDecisionAuditValues(audit.DataUsed)
	audit.MissingData = uniqueWorkoutDecisionAuditValues(audit.MissingData)
	audit.ConstraintsApplied = uniqueWorkoutDecisionAuditValues(audit.ConstraintsApplied)
	audit.AlternativesRejected = uniqueWorkoutDecisionAuditValues(audit.AlternativesRejected)
	return audit
}

func uniqueWorkoutDecisionAuditValues(values []string) []string {
	unique := make([]string, 0, len(values))
	for _, value := range values {
		unique = appendUniqueString(unique, value)
	}
	return unique
}
