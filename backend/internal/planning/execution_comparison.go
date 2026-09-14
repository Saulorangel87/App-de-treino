package planning

import "time"

const (
	plannedVsActualVersion = "planned-vs-actual-v2"
	plannedVsActualMode    = "observation"
	plannedVsActualScope   = "completed_workout"
)

// PlannedVsActualInput contains the prescription and execution facts that
// are available when a workout is completed. It is descriptive only and does
// not decide how a future workout should be prescribed.
type PlannedVsActualInput struct {
	PlannedDurationMinutes int
	ActualDurationMinutes  int
	TargetRPE              float64
	ActualRPE              float64
	FeedbackPresent        bool
	CompletionStatus       string
	PartialReason          string
	Difficulty             string
	PainReported           bool
	FatigueAfter           int
	RecoveryAfter          *int
	RepeatConfidence       *int
	Satisfaction           *int
	Terrain                string
	ExternalConditions     string
	EquipmentUsed          string
	DistanceKM             *float64
	ElevationGainM         *int
	AveragePowerW          *int
	AverageHeartRate       *int
}

// PlannedVsActualAssessment records the observable difference between one
// prescription and its completed execution. Missing fields are explicit so
// the comparison cannot be mistaken for a complete physiological assessment.
type PlannedVsActualAssessment struct {
	Version                   string            `json:"version"`
	Mode                      string            `json:"mode"`
	Scope                     string            `json:"scope"`
	AssessedAt                string            `json:"assessed_at"`
	Status                    string            `json:"status"`
	PlannedDurationMinutes    int               `json:"planned_duration_minutes"`
	ActualDurationMinutes     int               `json:"actual_duration_minutes"`
	DurationDeltaMinutes      *int              `json:"duration_delta_minutes,omitempty"`
	DurationCompletionPercent *float64          `json:"duration_completion_percent,omitempty"`
	CompletionStatus          string            `json:"completion_status"`
	PartialReason             string            `json:"partial_reason,omitempty"`
	TargetRPE                 float64           `json:"target_rpe"`
	ActualRPE                 float64           `json:"actual_rpe"`
	RPEDelta                  *float64          `json:"rpe_delta,omitempty"`
	RecoveryAfter             *int              `json:"recovery_after,omitempty"`
	RepeatConfidence          *int              `json:"repeat_confidence,omitempty"`
	Satisfaction              *int              `json:"satisfaction,omitempty"`
	Terrain                   string            `json:"terrain,omitempty"`
	ExternalConditions        string            `json:"external_conditions,omitempty"`
	EquipmentUsed             string            `json:"equipment_used,omitempty"`
	ObservedFields            []string          `json:"observed_fields"`
	Reasons                   []ReadinessReason `json:"reasons"`
	MissingData               []string          `json:"missing_data"`
	DataIssues                []string          `json:"data_issues"`
	NotEvaluated              []string          `json:"not_evaluated"`
	ProgressionEligible       bool              `json:"progression_eligible"`
	UsedForPrescription       bool              `json:"used_for_prescription"`
}

// AssessPlannedVsActual captures the fields available at workout completion.
// It deliberately reports differences instead of assigning training meaning
// to them; later adaptation rules must evaluate the context separately.
func AssessPlannedVsActual(input PlannedVsActualInput, now time.Time) PlannedVsActualAssessment {
	result := PlannedVsActualAssessment{
		Version:                plannedVsActualVersion,
		Mode:                   plannedVsActualMode,
		Scope:                  plannedVsActualScope,
		AssessedAt:             now.UTC().Format(time.RFC3339Nano),
		Status:                 "not_evaluated",
		PlannedDurationMinutes: input.PlannedDurationMinutes,
		ActualDurationMinutes:  input.ActualDurationMinutes,
		CompletionStatus:       normalizeCompletionStatus(input.CompletionStatus),
		TargetRPE:              input.TargetRPE,
		ActualRPE:              input.ActualRPE,
		ObservedFields: []string{
			"planned_duration_minutes",
			"actual_duration_minutes",
			"target_rpe",
			"actual_rpe",
		},
		Reasons:             []ReadinessReason{},
		MissingData:         []string{},
		DataIssues:          []string{},
		NotEvaluated:        []string{"cadence", "sleep", "stress", "recovery"},
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

	if input.PlannedDurationMinutes <= 0 {
		addIssue("invalid_planned_duration_minutes")
	}
	if input.ActualDurationMinutes <= 0 {
		addIssue("invalid_actual_duration_minutes")
	}
	if !finiteInRange(input.TargetRPE, 1, 10) {
		addIssue("invalid_target_rpe")
	}
	if !finiteInRange(input.ActualRPE, 1, 10) {
		addIssue("invalid_actual_rpe")
	}
	if !input.FeedbackPresent {
		addMissing("feedback")
	} else {
		addObserved("completion_status")
		if result.CompletionStatus == "partial" {
			if validPartialReason(input.PartialReason) {
				result.PartialReason = input.PartialReason
				addObserved("partial_reason")
			} else if input.PartialReason == "" {
				addMissing("partial_reason")
			} else {
				addIssue("invalid_partial_reason")
			}
		} else if result.CompletionStatus == "complete" {
			if input.PartialReason != "" {
				addIssue("unexpected_partial_reason")
			}
		} else {
			addIssue("invalid_completion_status")
		}
		addObserved("difficulty")
		addObserved("pain_reported")
		addObserved("fatigue_after")
		switch input.Difficulty {
		case "very_easy", "easy", "moderate", "hard", "very_hard":
		default:
			addIssue("invalid_difficulty")
		}
		if input.FatigueAfter < 1 || input.FatigueAfter > 5 {
			addIssue("invalid_fatigue_after")
		}
		if input.RecoveryAfter != nil {
			if *input.RecoveryAfter < 1 || *input.RecoveryAfter > 5 {
				addIssue("invalid_recovery_after")
			} else {
				result.RecoveryAfter = input.RecoveryAfter
				addObserved("recovery_after")
			}
		}
		if input.RepeatConfidence != nil {
			if *input.RepeatConfidence < 1 || *input.RepeatConfidence > 5 {
				addIssue("invalid_repeat_confidence")
			} else {
				result.RepeatConfidence = input.RepeatConfidence
				addObserved("repeat_confidence")
			}
		}
		if input.Satisfaction != nil {
			if *input.Satisfaction < 1 || *input.Satisfaction > 5 {
				addIssue("invalid_satisfaction")
			} else {
				value := *input.Satisfaction
				result.Satisfaction = &value
				addObserved("satisfaction")
			}
		} else {
			addMissing("satisfaction")
		}
		if input.Terrain == "" {
			addMissing("terrain")
		} else if !validFeedbackTerrain(input.Terrain) {
			addIssue("invalid_terrain")
		} else {
			result.Terrain = input.Terrain
			addObserved("terrain")
		}
		if input.ExternalConditions == "" {
			addMissing("external_conditions")
		} else if !validExternalConditions(input.ExternalConditions) {
			addIssue("invalid_external_conditions")
		} else {
			result.ExternalConditions = input.ExternalConditions
			addObserved("external_conditions")
		}
		if input.EquipmentUsed == "" {
			addMissing("equipment_used")
		} else if len(input.EquipmentUsed) > 120 {
			addIssue("invalid_equipment_used")
		} else {
			result.EquipmentUsed = input.EquipmentUsed
			addObserved("equipment_used")
		}
	}

	optionalFields := []struct {
		name    string
		present bool
	}{
		{name: "distance_km", present: input.DistanceKM != nil},
		{name: "elevation_gain_m", present: input.ElevationGainM != nil},
		{name: "average_power_watts", present: input.AveragePowerW != nil},
		{name: "average_heart_rate", present: input.AverageHeartRate != nil},
	}
	for _, field := range optionalFields {
		if field.present {
			addObserved(field.name)
		} else {
			addMissing(field.name)
		}
	}

	if len(result.DataIssues) > 0 || (!input.FeedbackPresent && len(result.MissingData) > 0) {
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "comparison_not_complete",
			Message: "A comparação registra os dados disponíveis, mas ainda contém campos inválidos ou ausentes.",
		})
		return result
	}

	durationDelta := input.ActualDurationMinutes - input.PlannedDurationMinutes
	durationPercent := float64(input.ActualDurationMinutes) / float64(input.PlannedDurationMinutes) * 100
	rpeDelta := input.ActualRPE - input.TargetRPE
	result.DurationDeltaMinutes = &durationDelta
	result.DurationCompletionPercent = &durationPercent
	result.RPEDelta = &rpeDelta
	result.Status = "observed"
	result.Reasons = append(result.Reasons, ReadinessReason{
		Code:    "comparison_recorded",
		Message: "A diferença entre o planejado e o realizado foi registrada apenas para observação; ela não altera a próxima sessão.",
	})
	return result
}
