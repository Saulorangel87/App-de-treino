package planning

import (
	"math"
	"time"
)

// Coverage counts distinguish absent/invalid fields from legitimate low values.
// They describe recorded data, not adherence or rides performed outside the app.
type ObservedDataCoverage struct {
	SessionsWithDuration int `json:"sessions_with_duration"`
	SessionsWithRPE      int `json:"sessions_with_rpe"`
	SessionsWithFeedback int `json:"sessions_with_feedback"`
	CompleteSessions     int `json:"complete_sessions"`
	RecoveryWithFatigue  int `json:"recovery_with_fatigue"`
}

type ReadinessReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ReadinessAssessment is an observational snapshot, NOT a prescription gate or
// the readiness of today's check-in. rules-v1 remains the only prescribing engine.
type ReadinessAssessment struct {
	ClassifierVersion   string                `json:"classifier_version"`
	Mode                string                `json:"mode"`
	Scope               string                `json:"scope"`
	AssessedAt          string                `json:"assessed_at"`
	Status              string                `json:"status"`
	Reasons             []ReadinessReason     `json:"reasons"`
	MissingData         []string              `json:"missing_data"`
	NotEvaluated        []string              `json:"not_evaluated"`
	ProgressionEligible bool                  `json:"progression_eligible"`
	ActiveLimitations   int                   `json:"active_limitations"`
	DataCoverage        *ObservedDataCoverage `json:"data_coverage,omitempty"`
}

func assessReadiness(input Context, now time.Time) ReadinessAssessment {
	observed := input.Observed
	result := ReadinessAssessment{
		ClassifierVersion: "readiness-v1",
		Mode:              "observation",
		Scope:             "observed_history_28d",
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "insufficient_data",
		Reasons:           []ReadinessReason{},
		MissingData:       []string{},
		NotEvaluated: []string{
			"adherence", "load_tolerance", "detraining", "trends_7_28_42d",
			"latest_sleep_stress_fatigue", "assessment_recency", "within_window_variation",
		},
		ActiveLimitations: len(input.Limitations),
	}
	// Copy rather than retain the caller's mutable coverage pointer.
	if observed.DataCoverage != nil {
		coverage := *observed.DataCoverage
		result.DataCoverage = &coverage
	}
	addReason := func(code, message string) {
		result.Reasons = append(result.Reasons, ReadinessReason{Code: code, Message: message})
	}
	if observed.WindowDays != 28 {
		result.MissingData = append(result.MissingData, "history_window_28d")
	}
	if observed.CompletedSessions == 0 {
		result.MissingData = append(result.MissingData, "completed_sessions")
	}
	if observed.RecoveryCheckins == 0 {
		result.MissingData = append(result.MissingData, "recovery_checkins")
	}
	coverage := result.DataCoverage
	if coverage == nil {
		result.MissingData = append(result.MissingData, "data_coverage")
	} else {
		if coverage.SessionsWithDuration < observed.CompletedSessions {
			result.MissingData = append(result.MissingData, "session_duration")
		}
		if coverage.SessionsWithRPE < observed.CompletedSessions {
			result.MissingData = append(result.MissingData, "session_rpe")
		}
		if coverage.SessionsWithFeedback < observed.CompletedSessions {
			result.MissingData = append(result.MissingData, "session_feedback")
		}
		if coverage.RecoveryWithFatigue < observed.RecoveryCheckins {
			result.MissingData = append(result.MissingData, "recovery_fatigue")
		}
	}
	valid := validReadinessHistory(observed)
	if !valid {
		result.MissingData = append(result.MissingData, "consistent_history")
	}
	// Preserve known protection signals even when other data is incomplete.
	if len(input.Limitations) > 0 {
		addReason("active_limitation", "Há limitação ativa no perfil; a proteção existente continua prioritária.")
	}
	if observed.PainReported {
		addReason("reported_pain", "Há dor relatada no histórico observado; este registro não indica se a dor persiste hoje.")
	}
	if finiteInRange(observed.AverageFatigue, 4, 5) {
		addReason("high_session_fatigue", "A fadiga média pós-treino atingiu o limiar de proteção já utilizado pelo rules-v1.")
	}
	if finiteInRange(observed.AverageRecoveryFatigue, 4, 5) {
		addReason("high_recovery_fatigue", "A fadiga média dos check-ins atingiu o limiar de proteção já utilizado pelo rules-v1.")
	}
	if len(result.Reasons) > 0 {
		result.Status = "recovery_needed"
		return result
	}
	if !valid || observed.WindowDays != 28 || coverage == nil ||
		(coverage.CompleteSessions == 0 && coverage.RecoveryWithFatigue == 0) {
		addReason("insufficient_observed_data", "Não há dados completos e consistentes suficientes para descrever o histórico disponível.")
		return result
	}
	if len(result.MissingData) > 0 {
		result.Status = "caution"
		addReason("partial_observed_data", "O histórico contém registros utilizáveis, mas há lacunas; ausência de registro não significa falta de treino.")
		return result
	}
	result.Status = "stable"
	addReason("no_aggregate_alerts", "Sem alertas nos agregados disponíveis; isso não comprova prontidão atual nem autoriza progressão.")
	return result
}

func finiteInRange(value, minimum, maximum float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= minimum && value <= maximum
}

func validReadinessHistory(observed ObservedTrainingSummary) bool {
	if observed.CompletedSessions < 0 || observed.CompletedMinutes < 0 || observed.RecoveryCheckins < 0 ||
		!finiteInRange(observed.AverageRPE, 0, 10) || !finiteInRange(observed.AverageFatigue, 0, 5) ||
		!finiteInRange(observed.AverageRecoveryFatigue, 0, 5) {
		return false
	}
	coverage := observed.DataCoverage
	if coverage == nil {
		return true // Unknown coverage is separately reported as missing, not invalid.
	}
	for _, count := range []int{coverage.SessionsWithDuration, coverage.SessionsWithRPE, coverage.SessionsWithFeedback, coverage.CompleteSessions} {
		if count < 0 || count > observed.CompletedSessions {
			return false
		}
	}
	if coverage.RecoveryWithFatigue < 0 || coverage.RecoveryWithFatigue > observed.RecoveryCheckins ||
		coverage.CompleteSessions > min(coverage.SessionsWithDuration, coverage.SessionsWithRPE, coverage.SessionsWithFeedback) ||
		coverage.CompleteSessions < max(0, coverage.SessionsWithDuration+coverage.SessionsWithRPE+coverage.SessionsWithFeedback-2*observed.CompletedSessions) ||
		observed.CompletedMinutes < coverage.SessionsWithDuration {
		return false
	}
	if observed.CompletedSessions == 0 && (observed.CompletedMinutes != 0 || observed.AverageRPE != 0 || observed.AverageFatigue != 0) ||
		observed.RecoveryCheckins == 0 && observed.AverageRecoveryFatigue != 0 {
		return false
	}
	return !(coverage.SessionsWithRPE == observed.CompletedSessions && observed.CompletedSessions > 0 && observed.AverageRPE < 1 ||
		coverage.SessionsWithFeedback == observed.CompletedSessions && observed.CompletedSessions > 0 && observed.AverageFatigue < 1 ||
		coverage.RecoveryWithFatigue == observed.RecoveryCheckins && observed.RecoveryCheckins > 0 && observed.AverageRecoveryFatigue < 1)
}
