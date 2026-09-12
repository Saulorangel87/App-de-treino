package planning

import "time"

const (
	loadToleranceVersion = "load-tolerance-v1"
	loadToleranceMode    = "observation"
	loadToleranceScope   = "completed_workout"
)

// LoadToleranceAssessment records whether recent session load has enough
// coherent context to be observed as tolerated. It never authorizes a
// progression or changes the rules-v1 prescription.
type LoadToleranceAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	CandidateResponse   string            `json:"candidate_response"`
	EvidencePeriods     []string          `json:"evidence_periods"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	Reasons             []ReadinessReason `json:"reasons"`
	MissingData         []string          `json:"missing_data"`
	DataIssues          []string          `json:"data_issues"`
	NotEvaluated        []string          `json:"not_evaluated"`
	ProgressionEligible bool              `json:"progression_eligible"`
	Applied             bool              `json:"applied"`
	UsedForPrescription bool              `json:"used_for_prescription"`
}

// assessLoadTolerance observes two recent, non-overlapping periods with
// session-RPE load, complete feedback and recovery context. Absence of a
// protective signal is described as observed support only; it is not proof of
// physiological tolerance and is never used to increase load.
func assessLoadTolerance(targetRPE float64, input CompletionInput, periods []TrainingHistoryPeriod, now time.Time) LoadToleranceAssessment {
	result := LoadToleranceAssessment{
		Version:           loadToleranceVersion,
		Mode:              loadToleranceMode,
		Scope:             loadToleranceScope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "not_evaluated",
		EvidencePeriods:   []string{},
		RulesEvaluated: []string{
			"period_integrity_gate",
			"session_rpe_load_gate",
			"complete_feedback_gate",
			"recovery_context_gate",
			"protective_signal_gate",
			"prescription_isolation_gate",
		},
		Reasons:      []ReadinessReason{},
		MissingData:  []string{},
		DataIssues:   []string{},
		NotEvaluated: []string{"detraining", "fitness_change", "activities_outside_cadencia", "athlete_timezone", "prescription_effect"},
	}

	addReason := func(code, message string) {
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addProtectiveReason := func(code, message string) {
		addReason(code, message)
		result.Status = "protective_signal"
		result.CandidateResponse = "prefer_recovery"
	}

	if !finiteInRange(targetRPE, 1, 10) || !validCompletion(input) {
		result.DataIssues = append(result.DataIssues, "invalid_feedback_or_target_rpe")
		addReason("invalid_feedback", "O feedback ou o RPE planejado não passou pela validação mínima; a tolerância não é avaliada.")
		return result
	}

	comparison := buildTrainingHistoryPeriodComparison(periods)
	if len(periods) != 6 {
		addMissing("period_comparison")
	}
	result.DataIssues = append(result.DataIssues, comparison.DataIssues...)

	if input.PainReported {
		addProtectiveReason("current_pain", "O treino concluído teve dor relatada; a observação de tolerância não deve liberar aumento de carga.")
		return result
	}
	if input.FatigueAfter >= 4 {
		addProtectiveReason("current_high_fatigue", "O treino concluído terminou com fadiga alta; a resposta observada deve permanecer protetiva.")
		return result
	}
	if input.ActualRPE >= targetRPE+2 {
		addProtectiveReason("current_above_target_rpe", "O esforço percebido ficou pelo menos dois pontos acima do alvo; a tolerância não é considerada sustentada.")
		return result
	}

	if len(comparison.Periods) > 0 {
		recent := comparison.Periods[0]
		if recent.PainReportedSessions > 0 {
			addProtectiveReason("recent_pain", "O período mais recente contém dor após sessão; a tolerância observada não deve liberar aumento de carga.")
			return result
		}
		if recent.HighFatigueSessions > 0 {
			addProtectiveReason("recent_high_fatigue", "O período mais recente contém fadiga alta após sessão; a resposta observada deve permanecer protetiva.")
			return result
		}
		if recent.AboveTargetRPESessions > 0 {
			addProtectiveReason("recent_above_target_rpe", "O período mais recente contém esforço acima do alvo; a tolerância observada não é considerada sustentada.")
			return result
		}
		if recent.RecoveryNeededCheckins > 0 {
			addProtectiveReason("recent_recovery_need", "O período mais recente contém necessidade de recuperação; a carga não deve ser aumentada.")
			return result
		}
	}

	if len(result.DataIssues) > 0 {
		addReason("inconsistent_observation", "Há inconsistências nos períodos observados; a tolerância não é classificada.")
		return result
	}

	if len(comparison.Periods) < 2 {
		addMissing("recent_period")
		addMissing("prior_period")
	} else {
		recent, prior := comparison.Periods[0], comparison.Periods[1]
		appendLoadToleranceEvidenceMissing(&result.MissingData, recent, "recent")
		appendLoadToleranceEvidenceMissing(&result.MissingData, prior, "prior")
		if recent.PeriodKey != "" {
			result.EvidencePeriods = append(result.EvidencePeriods, recent.PeriodKey)
		}
		if prior.PeriodKey != "" {
			result.EvidencePeriods = append(result.EvidencePeriods, prior.PeriodKey)
		}
	}
	if len(result.MissingData) > 0 {
		addReason("insufficient_load_tolerance_evidence", "São necessários dois períodos recentes com carga, feedback e recuperação completos; a ausência de dados não prova baixa tolerância.")
		return result
	}

	addReason("load_tolerance_observed", "Os dois períodos recentes têm carga, feedback e recuperação completos sem sinal protetivo observado; isso descreve suporte observacional e não autoriza progressão.")
	result.Status = "observation_only"
	result.CandidateResponse = "maintain_observed"
	return result
}

func appendLoadToleranceEvidenceMissing(missing *[]string, period TrainingHistoryPeriod, label string) {
	if period.PerformedSessions == 0 {
		*missing = appendUniqueString(*missing, label+"_performed_sessions")
	}
	if period.SessionsWithSessionRPELoad < period.PerformedSessions {
		*missing = appendUniqueString(*missing, label+"_session_rpe_load")
	}
	if period.SessionsWithCompleteFeedback < period.PerformedSessions {
		*missing = appendUniqueString(*missing, label+"_complete_feedback")
	}
	if period.CompleteRecoveryCheckins == 0 {
		*missing = appendUniqueString(*missing, label+"_complete_recovery_checkin")
	}
}

// AssessLoadTolerance exposes the observational evaluator for focused tests
// and future audit tooling without making it part of prescription.
func AssessLoadTolerance(targetRPE float64, input CompletionInput, periods []TrainingHistoryPeriod, now time.Time) LoadToleranceAssessment {
	return assessLoadTolerance(targetRPE, input, periods, now)
}
