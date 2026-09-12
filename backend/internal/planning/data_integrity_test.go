package planning

import (
	"math"
	"slices"
	"testing"
	"time"
)

func validWorkoutDataIntegrityInput() WorkoutDataIntegrityInput {
	duration := 45
	actualRPE := 6.0
	fatigue := 3
	recoveryAfter := 4
	repeatConfidence := 5
	distance := 32.5
	elevation := 420
	power := 185
	heartRate := 142
	return WorkoutDataIntegrityInput{
		DurationMinutes:  &duration,
		ActualRPE:        &actualRPE,
		DistanceKM:       &distance,
		ElevationGainM:   &elevation,
		AveragePowerW:    &power,
		AverageHeartRate: &heartRate,
		FeedbackPresent:  true,
		Difficulty:       "moderate",
		PainReported:     false,
		FatigueAfter:     &fatigue,
		RecoveryAfter:    &recoveryAfter,
		RepeatConfidence: &repeatConfidence,
	}
}

func TestAssessWorkoutDataIntegrityAcceptsCompleteCoherentSession(t *testing.T) {
	assessment := AssessWorkoutDataIntegrity(validWorkoutDataIntegrityInput(), time.Unix(0, 0))

	if assessment.Status != "valid" || !assessment.EligibleForHistory {
		t.Fatalf("complete session was not eligible: %+v", assessment)
	}
	if len(assessment.MissingData) != 0 || len(assessment.DataIssues) != 0 {
		t.Fatalf("complete session reported data problems: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.UsedForPrescription {
		t.Fatalf("integrity gate became prescriptive: %+v", assessment)
	}
}

func TestAssessWorkoutDataIntegrityMarksZeroDurationAsIncomplete(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	zero := 0
	input.DurationMinutes = &zero
	input.DistanceKM = nil
	input.ElevationGainM = nil
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "incomplete" || assessment.EligibleForHistory {
		t.Fatalf("zero duration was not treated as incomplete: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "positive_duration_minutes") {
		t.Fatalf("positive duration gap was not reported: %+v", assessment.MissingData)
	}
}

func TestAssessWorkoutDataIntegrityRejectsDistanceWithoutElapsedTime(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	zero := 0
	input.DurationMinutes = &zero
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "inconsistent" || assessment.EligibleForHistory {
		t.Fatalf("distance without elapsed time was not rejected: %+v", assessment)
	}
	if !slices.Contains(assessment.DataIssues, "duration_zero_with_distance") {
		t.Fatalf("inconsistent duration/distance pair was not reported: %+v", assessment.DataIssues)
	}
}

func TestAssessWorkoutDataIntegritySeparatesMissingFeedbackFromInvalidMetrics(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	input.FeedbackPresent = false
	input.ActualRPE = nil
	invalidPower := 2001
	input.AveragePowerW = &invalidPower
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "inconsistent" {
		t.Fatalf("invalid metric must take precedence over missing feedback: %+v", assessment)
	}
	for _, value := range []string{"feedback", "actual_rpe"} {
		if !slices.Contains(assessment.MissingData, value) {
			t.Fatalf("missing data did not include %q: %+v", value, assessment.MissingData)
		}
	}
	if !slices.Contains(assessment.DataIssues, "invalid_average_power_watts") {
		t.Fatalf("invalid metric was not reported: %+v", assessment.DataIssues)
	}
}

func TestAssessWorkoutDataIntegrityRejectsNonFiniteRPE(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	notANumber := math.NaN()
	input.ActualRPE = &notANumber
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "inconsistent" || !slices.Contains(assessment.DataIssues, "invalid_actual_rpe") {
		t.Fatalf("non-finite RPE was not rejected: %+v", assessment)
	}
}

func TestAssessWorkoutDataIntegrityAcceptsValidPartialCompletion(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	input.CompletionStatus = "partial"
	input.PartialReason = "time_available_changed"
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "valid" || !assessment.EligibleForHistory {
		t.Fatalf("valid partial completion was not eligible for observation: %+v", assessment)
	}
}

func TestAssessWorkoutDataIntegrityRejectsPartialCompletionWithoutReason(t *testing.T) {
	input := validWorkoutDataIntegrityInput()
	input.CompletionStatus = "partial"
	assessment := AssessWorkoutDataIntegrity(input, time.Unix(0, 0))

	if assessment.Status != "incomplete" || !slices.Contains(assessment.MissingData, "partial_reason") {
		t.Fatalf("missing partial reason was not reported: %+v", assessment)
	}
}
