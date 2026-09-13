package planning

import "time"

const adaptationDecisionAuditVersion = "adaptation-audit-v1"

// AdaptationDecisionAudit records the provenance and limits of one shadow
// evaluation. It describes the decision surface without assigning a
// calibrated confidence or authorizing a prescription.
type AdaptationDecisionAudit struct {
	Version              string   `json:"version"`
	Mode                 string   `json:"mode"`
	Scope                string   `json:"scope"`
	AssessedAt           string   `json:"assessed_at"`
	DataUsed             []string `json:"data_used"`
	MissingData          []string `json:"missing_data"`
	ConstraintsApplied   []string `json:"constraints_applied"`
	AlternativesRejected []string `json:"alternatives_rejected"`
	ConditionsForChange  []string `json:"conditions_for_change"`
	Confidence           string   `json:"confidence"`
	NotEvaluated         []string `json:"not_evaluated"`
	UsedForPrescription  bool     `json:"used_for_prescription"`
}

func buildAdaptationDecisionAudit(result RulesV2AdaptationShadowAssessment, input CompletionInput, now time.Time) AdaptationDecisionAudit {
	audit := AdaptationDecisionAudit{
		Version:              adaptationDecisionAuditVersion,
		Mode:                 "observation",
		Scope:                "post_workout_feedback",
		AssessedAt:           now.UTC().Format(time.RFC3339Nano),
		DataUsed:             []string{"target_rpe", "actual_rpe", "difficulty", "fatigue_after", "pain_reported", "completion_status"},
		MissingData:          append([]string{}, result.MissingData...),
		ConstraintsApplied:   []string{"rules_v1_prescription_isolation"},
		AlternativesRejected: append([]string{}, result.RulesDeferred...),
		ConditionsForChange: []string{
			"revisar_e_calibrar_a_interpretacao_dos_sinais",
			"validar_o_efeito_da_prescricao_em_dados_reais",
		},
		Confidence:          "not_calibrated",
		NotEvaluated:        []string{"confidence_calibration", "prescription_effect", "longitudinal_effect"},
		UsedForPrescription: false,
	}

	addDataUsed := func(value string) {
		audit.DataUsed = appendUniqueString(audit.DataUsed, value)
	}
	addConstraint := func(value string) {
		audit.ConstraintsApplied = appendUniqueString(audit.ConstraintsApplied, value)
	}
	addCondition := func(value string) {
		audit.ConditionsForChange = appendUniqueString(audit.ConditionsForChange, value)
	}

	if input.RecoveryAfter != nil && *input.RecoveryAfter >= 1 && *input.RecoveryAfter <= 5 {
		addDataUsed("recovery_after")
	}
	if input.RepeatConfidence != nil && *input.RepeatConfidence >= 1 && *input.RepeatConfidence <= 5 {
		addDataUsed("repeat_confidence")
	}
	if input.CompletionStatus == "partial" && input.PartialReason != "" {
		addDataUsed("partial_reason")
		addConstraint("partial_completion")
	}
	if len(result.DataIssues) > 0 {
		addConstraint("data_integrity_gate")
		addCondition("corrigir_inconsistencias_antes_de_reavaliar")
	}
	if len(result.MissingData) > 0 {
		addConstraint("insufficient_evidence_gate")
		addCondition("obter_dados_minimos_de_carga_feedback_e_recuperacao")
	}
	if result.CandidateResponse == "prefer_recovery" {
		addConstraint("protective_signal_gate")
		addCondition("confirmar_recuperacao_adequada_antes_de_considerar_progressao")
	}
	if result.CandidateResponse == "progress_duration_5pct" {
		addConstraint("progression_remains_shadow_only")
	}

	return audit
}
