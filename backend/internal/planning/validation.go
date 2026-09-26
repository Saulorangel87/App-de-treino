package planning

import (
	"strings"
)

func validCompletion(input CompletionInput) bool {
	if !finiteInRange(input.ActualRPE, 1, 10) || input.FatigueAfter < 1 || input.FatigueAfter > 5 || len(input.Notes) > 1000 || len(input.EquipmentUsed) > 120 {
		return false
	}
	completionStatus := normalizeCompletionStatus(input.CompletionStatus)
	if completionStatus != "complete" && completionStatus != "partial" {
		return false
	}
	if completionStatus == "partial" && !validPartialReason(input.PartialReason) {
		return false
	}
	if completionStatus == "complete" && strings.TrimSpace(input.PartialReason) != "" {
		return false
	}
	if input.DistanceKM != nil && !finiteInRange(*input.DistanceKM, 0, 2000) {
		return false
	}
	if input.ElevationGainM != nil && (*input.ElevationGainM < 0 || *input.ElevationGainM > 20000) {
		return false
	}
	if input.AveragePowerW != nil && (*input.AveragePowerW < 0 || *input.AveragePowerW > 2000) {
		return false
	}
	if input.AverageHeartRate != nil && (*input.AverageHeartRate < 30 || *input.AverageHeartRate > 250) {
		return false
	}
	if input.AverageCadenceRPM != nil && (*input.AverageCadenceRPM < 1 || *input.AverageCadenceRPM > 300) {
		return false
	}
	if input.RecoveryAfter != nil && (*input.RecoveryAfter < 1 || *input.RecoveryAfter > 5) {
		return false
	}
	if input.RepeatConfidence != nil && (*input.RepeatConfidence < 1 || *input.RepeatConfidence > 5) {
		return false
	}
	if input.Satisfaction != nil && (*input.Satisfaction < 1 || *input.Satisfaction > 5) {
		return false
	}
	if input.Terrain != "" && !validFeedbackTerrain(input.Terrain) {
		return false
	}
	if input.ExternalConditions != "" && !validExternalConditions(input.ExternalConditions) {
		return false
	}
	switch input.Difficulty {
	case "very_easy", "easy", "moderate", "hard", "very_hard":
		return true
	default:
		return false
	}
}

func validWorkoutCorrection(input WorkoutCorrectionInput) bool {
	if input.DistanceKM != nil && !finiteInRange(*input.DistanceKM, 0, 2000) {
		return false
	}
	if input.ElevationGainM != nil && (*input.ElevationGainM < 0 || *input.ElevationGainM > 20000) {
		return false
	}
	if input.AveragePowerW != nil && (*input.AveragePowerW < 0 || *input.AveragePowerW > 2000) {
		return false
	}
	if input.AverageHeartRate != nil && (*input.AverageHeartRate < 30 || *input.AverageHeartRate > 250) {
		return false
	}
	if input.AverageCadenceRPM != nil && (*input.AverageCadenceRPM < 1 || *input.AverageCadenceRPM > 300) {
		return false
	}
	return true
}

func normalizeCompletionStatus(value string) string {
	if value == "" {
		return "complete"
	}
	return value
}

func validPartialReason(value string) bool {
	switch value {
	case "time_available_changed", "fatigue_or_recovery", "pain_or_discomfort", "equipment_or_conditions", "other":
		return true
	default:
		return false
	}
}

// WorkoutRequiresSafetyBlock keeps an already generated intense session from
// starting after a new active limitation was recorded. Protected sessions at
// RPE 4 or below remain startable.
func WorkoutRequiresSafetyBlock(targetRPE float64, hasActiveLimitation bool) bool {
	return hasActiveLimitation && targetRPE > 4 && finiteInRange(targetRPE, 1, 10)
}

func validFeedbackTerrain(value string) bool {
	switch value {
	case "flat", "rolling", "hilly", "mixed", "technical", "indoor":
		return true
	default:
		return false
	}
}

func validExternalConditions(value string) bool {
	switch value {
	case "normal", "heat", "cold", "wind", "rain", "poor_visibility", "other":
		return true
	default:
		return false
	}
}
