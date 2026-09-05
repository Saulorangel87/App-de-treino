package planning

import "time"

const (
	rulesV2Version = "rules-v2"
	rulesV2Mode    = "shadow"
	rulesV2Scope   = "plan_generation_only"
)

// RulesV2ShadowAssessment evaluates the evidence gates that a future rules-v2
// prescription may use. It is intentionally not connected to workout
// generation: rules-v1 remains the only prescribing engine in this phase.
type RulesV2ShadowAssessment struct {
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

func assessRulesV2Shadow(input Context, now time.Time) RulesV2ShadowAssessment {
	result := RulesV2ShadowAssessment{
		Version:           rulesV2Version,
		Mode:              rulesV2Mode,
		Scope:             rulesV2Scope,
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "not_evaluated",
		RulesEvaluated: []string{
			"period_data_integrity_gate",
			"protective_signal_gate",
			"progression_evidence_gate",
		},
		RulesDeferred: []string{
			"change_duration",
			"change_target_rpe",
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
		MissingData:         []string{},
		DataIssues:          []string{},
		Reasons:             []ReadinessReason{},
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
		for _, existing := range result.MissingData {
			if existing == value {
				return
			}
		}
		result.MissingData = append(result.MissingData, value)
	}

	if !validReadinessHistory(input.Observed) {
		result.DataIssues = append(result.DataIssues, "inconsistent_observed_training")
	}
	comparison := buildTrainingHistoryPeriodComparison(input.TrainingHistoryPeriods)
	if len(input.TrainingHistoryPeriods) != 6 {
		addMissing("period_comparison")
	}
	result.DataIssues = append(result.DataIssues, comparison.DataIssues...)
	periods := comparison.Periods

	if len(input.Limitations) > 0 {
		addReason("active_limitation", "Há limitação ativa; uma futura versão deve manter a proteção antes de considerar carga.")
	}
	if input.Observed.PainReported {
		addReason("observed_pain", "Há dor relatada no histórico; o rules-v2 shadow não pode liberar progressão.")
	}
	if finiteInRange(input.Observed.AverageFatigue, 4, 5) || finiteInRange(input.Observed.AverageRecoveryFatigue, 4, 5) {
		addReason("observed_fatigue", "Há fadiga elevada em registros observados; a resposta candidata é protetiva.")
	}
	if len(periods) > 0 {
		recent := periods[0]
		if recent.PainReportedSessions > 0 {
			addReason("recent_pain", "O período mais recente contém dor após sessão; a carga não deve ser aumentada.")
		}
		if recent.HighFatigueSessions > 0 {
			addReason("recent_high_fatigue", "O período mais recente contém fadiga alta após sessão; a carga não deve ser aumentada.")
		}
		if recent.RecoveryNeededCheckins > 0 {
			addReason("recent_recovery_need", "O período mais recente contém check-in que já atende à proteção de recuperação.")
		}
	}
	if len(result.Reasons) > 0 {
		result.Status = "protective_signal"
		result.CandidateResponse = "prefer_recovery"
		return result
	}

	if len(result.DataIssues) > 0 {
		addReason("inconsistent_observation", "Há inconsistências nos dados observados; a avaliação shadow não produz resposta de carga.")
	}
	if len(periods) < 2 {
		addMissing("recent_period")
		addMissing("prior_period")
	} else {
		appendPeriodEvidenceMissing(&result.MissingData, periods[0], "recent")
		appendPeriodEvidenceMissing(&result.MissingData, periods[1], "prior")
		if periods[0].CompleteRecoveryCheckins+periods[1].CompleteRecoveryCheckins == 0 {
			addMissing("recent_complete_recovery_checkin_14d")
		}
	}
	if len(result.MissingData) > 0 || len(result.DataIssues) > 0 {
		addReason("insufficient_progression_evidence", "Ainda não há duas semanas recentes com carga, feedback e recuperação completos para avaliar progressão.")
		return result
	}

	if periods[0].AboveTargetRPESessions > 0 || periods[1].AboveTargetRPESessions > 0 {
		addReason("recent_above_target_rpe", "Houve RPE acima do alvo em período recente; a resposta candidata é manter e observar.")
	}
	addReason("shadow_only", "A resposta candidata não é aplicada; o rules-v1 continua gerando o plano.")
	result.Status = "observation_only"
	result.CandidateResponse = "maintain_observed"
	return result
}

func appendPeriodEvidenceMissing(missing *[]string, period TrainingHistoryPeriod, label string) {
	if period.PerformedSessions == 0 {
		*missing = appendUniqueString(*missing, label+"_performed_sessions")
	}
	if period.SessionsWithSessionRPELoad < period.PerformedSessions {
		*missing = appendUniqueString(*missing, label+"_session_rpe_load")
	}
	if period.SessionsWithCompleteFeedback < period.PerformedSessions {
		*missing = appendUniqueString(*missing, label+"_complete_feedback")
	}
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
