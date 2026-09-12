package planning

import "time"

const (
	workoutDataIntegrityVersion = "data-integrity-v1"
	workoutDataIntegrityMode    = "observation"
	workoutDataIntegrityScope   = "completed_workout"
)

// WorkoutDataIntegrityInput is the completed session and feedback as stored by
// the repository. Pointers distinguish an absent value from a valid zero.
type WorkoutDataIntegrityInput struct {
	DurationMinutes  *int
	ActualRPE        *float64
	DistanceKM       *float64
	ElevationGainM   *int
	AveragePowerW    *int
	AverageHeartRate *int
	FeedbackPresent  bool
	CompletionStatus string
	PartialReason    string
	Difficulty       string
	PainReported     bool
	FatigueAfter     *int
}

// WorkoutDataIntegrityAssessment records whether a completed session has the
// minimum coherent data to enter observational history. It is not a medical
// assessment and does not prescribe a workout.
type WorkoutDataIntegrityAssessment struct {
	Version             string            `json:"version"`
	Mode                string            `json:"mode"`
	Scope               string            `json:"scope"`
	AssessedAt          string            `json:"assessed_at"`
	Status              string            `json:"status"`
	RulesEvaluated      []string          `json:"rules_evaluated"`
	Reasons             []ReadinessReason `json:"reasons"`
	MissingData         []string          `json:"missing_data"`
	DataIssues          []string          `json:"data_issues"`
	NotEvaluated        []string          `json:"not_evaluated"`
	EligibleForHistory  bool              `json:"eligible_for_history"`
	ProgressionEligible bool              `json:"progression_eligible"`
	UsedForPrescription bool              `json:"used_for_prescription"`
}

// AssessWorkoutDataIntegrity applies format, minimum-data and consistency
// gates without rejecting the original completed record. The result is used
// to keep incomplete observations out of future shadow comparisons.
func AssessWorkoutDataIntegrity(input WorkoutDataIntegrityInput, now time.Time) WorkoutDataIntegrityAssessment {
	result := WorkoutDataIntegrityAssessment{
		Version:    workoutDataIntegrityVersion,
		Mode:       workoutDataIntegrityMode,
		Scope:      workoutDataIntegrityScope,
		AssessedAt: now.UTC().Format(time.RFC3339Nano),
		Status:     "not_evaluated",
		RulesEvaluated: []string{
			"required_session_data_gate",
			"required_feedback_gate",
			"completion_context_gate",
			"metric_range_gate",
			"measurement_consistency_gate",
			"history_eligibility_gate",
		},
		NotEvaluated: []string{
			"prescription_effect",
			"activities_outside_cadencia",
		},
		Reasons:             []ReadinessReason{},
		MissingData:         []string{},
		DataIssues:          []string{},
		ProgressionEligible: false,
		UsedForPrescription: false,
	}

	addMissing := func(value string) {
		result.MissingData = appendUniqueString(result.MissingData, value)
	}
	addIssue := func(value string) {
		result.DataIssues = appendUniqueString(result.DataIssues, value)
	}

	if input.DurationMinutes == nil {
		addMissing("duration_minutes")
	} else if *input.DurationMinutes < 0 {
		addIssue("negative_duration_minutes")
	} else if *input.DurationMinutes == 0 {
		addMissing("positive_duration_minutes")
	}
	if input.ActualRPE == nil {
		addMissing("actual_rpe")
	} else if !finiteInRange(*input.ActualRPE, 1, 10) {
		addIssue("invalid_actual_rpe")
	}

	if !input.FeedbackPresent {
		addMissing("feedback")
	} else {
		completionStatus := normalizeCompletionStatus(input.CompletionStatus)
		addObservedCompletionContext := func() {
			if completionStatus == "partial" {
				if !validPartialReason(input.PartialReason) {
					if input.PartialReason == "" {
						addMissing("partial_reason")
					} else {
						addIssue("invalid_partial_reason")
					}
				}
			} else if completionStatus == "complete" && input.PartialReason != "" {
				addIssue("unexpected_partial_reason")
			} else if completionStatus != "complete" {
				addIssue("invalid_completion_status")
			}
		}
		addObservedCompletionContext()
		switch input.Difficulty {
		case "very_easy", "easy", "moderate", "hard", "very_hard":
		default:
			addIssue("invalid_difficulty")
		}
		if input.FatigueAfter == nil {
			addMissing("fatigue_after")
		} else if *input.FatigueAfter < 1 || *input.FatigueAfter > 5 {
			addIssue("invalid_fatigue_after")
		}
	}

	if input.DistanceKM != nil && !finiteInRange(*input.DistanceKM, 0, 2000) {
		addIssue("invalid_distance_km")
	}
	if input.ElevationGainM != nil && (*input.ElevationGainM < 0 || *input.ElevationGainM > 20000) {
		addIssue("invalid_elevation_gain_m")
	}
	if input.AveragePowerW != nil && (*input.AveragePowerW < 0 || *input.AveragePowerW > 2000) {
		addIssue("invalid_average_power_watts")
	}
	if input.AverageHeartRate != nil && (*input.AverageHeartRate < 30 || *input.AverageHeartRate > 250) {
		addIssue("invalid_average_heart_rate")
	}

	noElapsedTime := input.DurationMinutes == nil || *input.DurationMinutes <= 0
	if noElapsedTime && input.DistanceKM != nil && *input.DistanceKM > 0 {
		if input.DurationMinutes == nil {
			addIssue("distance_without_duration")
		} else {
			addIssue("duration_zero_with_distance")
		}
	}
	if noElapsedTime && input.ElevationGainM != nil && *input.ElevationGainM > 0 {
		if input.DurationMinutes == nil {
			addIssue("elevation_without_duration")
		} else {
			addIssue("duration_zero_with_elevation")
		}
	}

	if len(result.DataIssues) > 0 {
		result.Status = "inconsistent"
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "inconsistent_completed_workout",
			Message: "Há valores impossíveis ou incompatíveis no treino concluído; ele fica fora da elegibilidade observacional até ser revisado.",
		})
	} else if len(result.MissingData) > 0 {
		result.Status = "incomplete"
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "incomplete_completed_workout",
			Message: "Faltam dados mínimos do treino concluído; ele não deve influenciar comparações ou progressão.",
		})
	} else {
		result.Status = "valid"
		result.EligibleForHistory = true
		result.Reasons = append(result.Reasons, ReadinessReason{
			Code:    "eligible_for_observation",
			Message: "Os dados mínimos do treino estão completos e coerentes para observação; isso não autoriza progressão.",
		})
	}

	return result
}
