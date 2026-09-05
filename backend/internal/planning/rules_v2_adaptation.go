package planning

import "time"

const (
	rulesV2AdaptationVersion = "rules-v2-adaptation-v1"
	rulesV2AdaptationMode    = "shadow"
	rulesV2AdaptationScope   = "post_workout_feedback"
)

// RulesV2AdaptationShadowAssessment describes a candidate response to a
// completed workout without changing the active plan. The current SQL trigger
// remains the authoritative rules-v1 implementation until this assessment is
// reviewed and integrated explicitly.
type RulesV2AdaptationShadowAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	CandidateResponse   string            `json:"candidate_response"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	RulesDeferred       []string          `json:"rules_deferred"`
	Reasons             []ReadinessReason `json:"reasons"`
	MissingData         []string          `json:"missing_data"`
	DataIssues          []string          `json:"data_issues"`
	NotEvaluated        []string          `json:"not_evaluated"`
	ProgressionEligible bool              `json:"progression_eligible"`
	Applied             bool              `json:"applied"`
	UsedForPrescription bool              `json:"used_for_prescription"`
}

// assessRulesV2AdaptationShadow compares one completed workout with the
// evidence gates required by the future adaptation engine. A single easy
// response is never enough for a progression candidate; protective feedback
// can still produce a protective candidate without waiting for history.
func assessRulesV2AdaptationShadow(targetRPE float64, input CompletionInput, periods []TrainingHistoryPeriod, now time.Time) RulesV2AdaptationShadowAssessment {
	result := RulesV2AdaptationShadowAssessment{
		Version:           rulesV2AdaptationVersion,
		Mode:              rulesV2AdaptationMode,
		Scope:             rulesV2AdaptationScope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "not_evaluated",
		RulesEvaluated: []string{
			"feedback_integrity_gate",
			"protective_signal_gate",
			"progression_evidence_gate",
			"prescription_isolation_gate",
		},
		RulesDeferred: []string{
			"apply_duration",
			"apply_target_rpe",
			"select_alternative_stimulus",
			"progression",
		},
		NotEvaluated: []string{
			"load_tolerance",
			"detraining",
			"fitness_change",
			"activities_outside_cadencia",
			"athlete_timezone",
			"prescription_effect",
		},
		Reasons:             []ReadinessReason{},
		MissingData:         []string{},
		DataIssues:          []string{},
		ProgressionEligible: false,
		Applied:             false,
		UsedForPrescription: false,
	}

	seenReasons := map[string]bool{}
	addReason := func(code, message string) {
		if seenReasons[code] {
			return
		}
		seenReasons[code] = true
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}

	if targetRPE < 1 || targetRPE > 10 || !validCompletion(input) {
		result.DataIssues = append(result.DataIssues, "invalid_feedback_or_target_rpe")
		addReason("invalid_feedback", "O feedback ou o RPE planejado não passou pela validação mínima; nenhuma resposta de adaptação é produzida.")
		return result
	}

	comparison := buildTrainingHistoryPeriodComparison(periods)
	if len(periods) != 6 {
		addMissing("period_comparison")
	}
	result.DataIssues = append(result.DataIssues, comparison.DataIssues...)
	recentProtective := false
	if len(comparison.Periods) > 0 {
		recent := comparison.Periods[0]
		recentProtective = recent.PainReportedSessions > 0 || recent.HighFatigueSessions > 0 || recent.RecoveryNeededCheckins > 0
		if recent.PainReportedSessions > 0 {
			addReason("recent_pain", "O período mais recente contém dor após sessão; a progressão deve permanecer bloqueada.")
		}
		if recent.HighFatigueSessions > 0 {
			addReason("recent_high_fatigue", "O período mais recente contém fadiga alta após sessão; a resposta candidata deve ser protetiva.")
		}
		if recent.RecoveryNeededCheckins > 0 {
			addReason("recent_recovery_need", "O período mais recente contém necessidade de recuperação; a carga não deve ser aumentada.")
		}
	}

	decision := DecideAdaptation(targetRPE, input)
	if decision.Kind == "safety" || decision.Kind == "recovery" || recentProtective {
		if decision.Kind == "safety" {
			addReason("post_workout_pain", decision.Reason)
		} else if decision.Kind == "recovery" {
			addReason("post_workout_strain", decision.Reason)
		}
		result.Status = "protective_signal"
		result.CandidateResponse = "prefer_recovery"
		return result
	}

	if len(result.DataIssues) > 0 {
		addReason("inconsistent_observation", "Há inconsistências no histórico observado; a avaliação não produz candidata de adaptação.")
		return result
	}

	if decision.Kind != "progression" {
		addReason("shadow_only", "A resposta observada não justifica mudança; o rules-v1 permanece responsável pela adaptação ativa.")
		result.Status = "observation_only"
		result.CandidateResponse = "maintain_observed"
		return result
	}

	if len(comparison.Periods) < 2 {
		addMissing("recent_period")
		addMissing("prior_period")
	} else {
		appendPeriodEvidenceMissing(&result.MissingData, comparison.Periods[0], "recent")
		appendPeriodEvidenceMissing(&result.MissingData, comparison.Periods[1], "prior")
		if comparison.Periods[0].CompleteRecoveryCheckins+comparison.Periods[1].CompleteRecoveryCheckins == 0 {
			addMissing("recent_complete_recovery_checkin_14d")
		}
	}
	if len(result.DataIssues) > 0 || len(result.MissingData) > 0 {
		addReason("progression_deferred_insufficient_evidence", "Uma resposta fácil isolada não basta para considerar progressão; são necessários períodos recentes com carga, feedback e recuperação completos.")
		result.CandidateResponse = "defer_progression"
		return result
	}

	addReason("progression_candidate", "A resposta fácil coincide com evidência recente completa; a proposta de aumento é pequena, somente na duração, e continua sem aplicação no shadow.")
	result.Status = "observation_only"
	result.CandidateResponse = "progress_duration_5pct"
	return result
}
