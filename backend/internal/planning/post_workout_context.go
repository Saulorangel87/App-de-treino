package planning

import "time"

const postWorkoutContextVersion = "post-workout-context-v1"

// PostWorkoutContextAssessment records the optional subjective context added
// after a workout. It describes data coverage only; it does not interpret
// physiology or authorize a change to the active prescription.
type PostWorkoutContextAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	CandidateResponse   string            `json:"candidate_response"`
	RecoveryAfter       *int              `json:"recovery_after,omitempty"`
	RepeatConfidence    *int              `json:"repeat_confidence,omitempty"`
	ObservedFields      []string          `json:"observed_fields"`
	Reasons             []ReadinessReason `json:"reasons"`
	MissingData         []string          `json:"missing_data"`
	DataIssues          []string          `json:"data_issues"`
	NotEvaluated        []string          `json:"not_evaluated"`
	ProgressionEligible bool              `json:"progression_eligible"`
	UsedForPrescription bool              `json:"used_for_prescription"`
}

// AssessPostWorkoutContext classifies the presence and range of the optional
// post-workout signals. A partial or invalid context is never converted into
// a candidate prescription.
func AssessPostWorkoutContext(input CompletionInput, now time.Time) PostWorkoutContextAssessment {
	result := PostWorkoutContextAssessment{
		Version:           postWorkoutContextVersion,
		Mode:              "observation",
		Scope:             "post_workout_feedback",
		AssessedAt:        now.UTC().Format(time.RFC3339Nano),
		Status:            "not_evaluated",
		CandidateResponse: "not_evaluated",
		ObservedFields:    []string{},
		Reasons:           []ReadinessReason{},
		MissingData:       []string{},
		DataIssues:        []string{},
		NotEvaluated: []string{
			"physiological_meaning",
			"longitudinal_trend",
			"prescription_effect",
		},
		ProgressionEligible: false,
		UsedForPrescription: false,
	}

	addObserved := func(value string) {
		result.ObservedFields = appendUniqueString(result.ObservedFields, value)
	}
	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addIssue := func(value string) {
		result.DataIssues = appendUniqueString(result.DataIssues, value)
	}

	if input.RecoveryAfter == nil {
		addMissing("recovery_after")
	} else if *input.RecoveryAfter < 1 || *input.RecoveryAfter > 5 {
		addIssue("invalid_recovery_after")
	} else {
		value := *input.RecoveryAfter
		result.RecoveryAfter = &value
		addObserved("recovery_after")
	}

	if input.RepeatConfidence == nil {
		addMissing("repeat_confidence")
	} else if *input.RepeatConfidence < 1 || *input.RepeatConfidence > 5 {
		addIssue("invalid_repeat_confidence")
	} else {
		value := *input.RepeatConfidence
		result.RepeatConfidence = &value
		addObserved("repeat_confidence")
	}

	if len(result.DataIssues) > 0 {
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "invalid_post_workout_context",
			Message: "Os sinais pós-treino contêm valores fora da faixa e permanecem fora da interpretação observacional.",
		})
		return result
	}

	if len(result.ObservedFields) == 0 {
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "post_workout_context_not_recorded",
			Message: "Os sinais pós-treino não foram registrados; nenhuma interpretação é feita a partir da ausência.",
		})
		return result
	}

	if len(result.MissingData) > 0 {
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "partial_post_workout_context",
			Message: "Somente parte dos sinais pós-treino foi registrada; o contexto permanece incompleto para análise conjunta.",
		})
		return result
	}

	result.Status = "observed"
	result.CandidateResponse = "maintain_observed"
	result.Reasons = append(result.Reasons, ReadinessReason{
		Code:    "post_workout_context_recorded",
		Message: "Os dois sinais pós-treino foram registrados para observação; nenhum deles altera a prescrição.",
	})
	return result
}
