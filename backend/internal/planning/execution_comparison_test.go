package planning

import (
	"math"
	"slices"
	"testing"
	"time"
)

func TestAssessPlannedVsActualRecordsCoreDifferences(t *testing.T) {
	distance := 42.5
	elevation := 380
	power := 210
	heartRate := 148
	recoveryAfter := 4
	repeatConfidence := 5
	satisfaction := 5
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 60,
		ActualDurationMinutes:  54,
		TargetRPE:              6,
		ActualRPE:              7,
		FeedbackPresent:        true,
		Difficulty:             "hard",
		PainReported:           false,
		FatigueAfter:           3,
		RecoveryAfter:          &recoveryAfter,
		RepeatConfidence:       &repeatConfidence,
		Satisfaction:           &satisfaction,
		Terrain:                "rolling",
		ExternalConditions:     "wind",
		EquipmentUsed:          "bike de estrada",
		DistanceKM:             &distance,
		ElevationGainM:         &elevation,
		AveragePowerW:          &power,
		AverageHeartRate:       &heartRate,
	}, time.Date(2026, time.September, 12, 20, 0, 0, 0, time.UTC))

	if assessment.Status != "observed" {
		t.Fatalf("status = %q, want observed", assessment.Status)
	}
	if assessment.DurationDeltaMinutes == nil || *assessment.DurationDeltaMinutes != -6 {
		t.Fatalf("duration delta = %v, want -6", assessment.DurationDeltaMinutes)
	}
	if assessment.DurationCompletionPercent == nil || math.Abs(*assessment.DurationCompletionPercent-90) > 0.001 {
		t.Fatalf("duration completion = %v, want 90", assessment.DurationCompletionPercent)
	}
	if assessment.RPEDelta == nil || math.Abs(*assessment.RPEDelta-1) > 0.001 {
		t.Fatalf("RPE delta = %v, want 1", assessment.RPEDelta)
	}
	if !slices.Contains(assessment.ObservedFields, "average_power_watts") || !slices.Contains(assessment.NotEvaluated, "sleep") {
		t.Fatalf("observed/not evaluated fields = %#v / %#v", assessment.ObservedFields, assessment.NotEvaluated)
	}
	if assessment.RecoveryAfter == nil || *assessment.RecoveryAfter != 4 || assessment.RepeatConfidence == nil || *assessment.RepeatConfidence != 5 || assessment.Satisfaction == nil || *assessment.Satisfaction != 5 {
		t.Fatalf("post-workout context = recovery %v, confidence %v, satisfaction %v", assessment.RecoveryAfter, assessment.RepeatConfidence, assessment.Satisfaction)
	}
	if assessment.EquipmentUsed != "bike de estrada" || !slices.Contains(assessment.ObservedFields, "recovery_after") || !slices.Contains(assessment.ObservedFields, "repeat_confidence") || !slices.Contains(assessment.ObservedFields, "satisfaction") || !slices.Contains(assessment.ObservedFields, "terrain") || !slices.Contains(assessment.ObservedFields, "external_conditions") || !slices.Contains(assessment.ObservedFields, "equipment_used") {
		t.Fatalf("post-workout context was not observed: %#v", assessment.ObservedFields)
	}
	if assessment.Version != "planned-vs-actual-v2" {
		t.Fatalf("version = %q, want planned-vs-actual-v2", assessment.Version)
	}
	if assessment.ProgressionEligible || assessment.UsedForPrescription {
		t.Fatal("comparison must remain non-prescriptive")
	}
}

func TestAssessPlannedVsActualKeepsIncompleteDataOutOfObservedState(t *testing.T) {
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 60,
		ActualDurationMinutes:  0,
		TargetRPE:              6,
		ActualRPE:              7,
		FeedbackPresent:        false,
	}, time.Now())

	if assessment.Status != "not_evaluated" {
		t.Fatalf("status = %q, want not_evaluated", assessment.Status)
	}
	if !slices.Contains(assessment.DataIssues, "invalid_actual_duration_minutes") || !slices.Contains(assessment.MissingData, "feedback") {
		t.Fatalf("issues/missing data = %#v / %#v", assessment.DataIssues, assessment.MissingData)
	}
	if assessment.DurationDeltaMinutes != nil || assessment.RPEDelta != nil {
		t.Fatal("invalid comparison must not produce deltas")
	}
}

func TestAssessPlannedVsActualDoesNotTreatMissingOptionalMetricsAsError(t *testing.T) {
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 45,
		ActualDurationMinutes:  45,
		TargetRPE:              5,
		ActualRPE:              5,
		FeedbackPresent:        true,
		Difficulty:             "moderate",
		FatigueAfter:           2,
	}, time.Now())

	if assessment.Status != "observed" {
		t.Fatalf("status = %q, want observed with optional coverage explicit", assessment.Status)
	}
	if !slices.Contains(assessment.MissingData, "distance_km") || !slices.Contains(assessment.MissingData, "average_heart_rate") {
		t.Fatalf("missing optional data = %#v", assessment.MissingData)
	}
	if len(assessment.DataIssues) != 0 {
		t.Fatalf("data issues = %#v, want none", assessment.DataIssues)
	}
}

func TestAssessPlannedVsActualRecordsPartialCompletionContext(t *testing.T) {
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 60,
		ActualDurationMinutes:  30,
		TargetRPE:              6,
		ActualRPE:              4,
		FeedbackPresent:        true,
		CompletionStatus:       "partial",
		PartialReason:          "fatigue_or_recovery",
		Difficulty:             "easy",
		FatigueAfter:           2,
	}, time.Now())

	if assessment.Status != "observed" || assessment.CompletionStatus != "partial" || assessment.PartialReason != "fatigue_or_recovery" {
		t.Fatalf("partial completion was not observed: %+v", assessment)
	}
	if !slices.Contains(assessment.ObservedFields, "partial_reason") || slices.Contains(assessment.NotEvaluated, "completion_reason") {
		t.Fatalf("partial context fields = %#v / %#v", assessment.ObservedFields, assessment.NotEvaluated)
	}
}

func TestAssessPlannedVsActualRejectsInvalidPostWorkoutContext(t *testing.T) {
	recoveryAfter := 0
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 45,
		ActualDurationMinutes:  45,
		TargetRPE:              5,
		ActualRPE:              5,
		FeedbackPresent:        true,
		Difficulty:             "moderate",
		FatigueAfter:           3,
		RecoveryAfter:          &recoveryAfter,
	}, time.Now())

	if assessment.Status != "not_evaluated" || !slices.Contains(assessment.DataIssues, "invalid_recovery_after") {
		t.Fatalf("invalid recovery context was not rejected: %+v", assessment)
	}
}

func TestAssessPlannedVsActualRejectsInvalidStructuredContext(t *testing.T) {
	satisfaction := 0
	assessment := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 45,
		ActualDurationMinutes:  45,
		TargetRPE:              5,
		ActualRPE:              5,
		FeedbackPresent:        true,
		Difficulty:             "moderate",
		FatigueAfter:           3,
		Satisfaction:           &satisfaction,
		Terrain:                "trail",
		ExternalConditions:     "storm",
	}, time.Now())

	if assessment.Status != "not_evaluated" {
		t.Fatalf("invalid structured context produced an observed comparison: %+v", assessment)
	}
	for _, issue := range []string{"invalid_satisfaction", "invalid_terrain", "invalid_external_conditions"} {
		if !slices.Contains(assessment.DataIssues, issue) {
			t.Fatalf("structured context issue %q was not reported: %+v", issue, assessment.DataIssues)
		}
	}
}
