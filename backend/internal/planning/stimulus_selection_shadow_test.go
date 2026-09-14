package planning

import (
	"testing"
	"time"
)

func selectionWorkout(key string, targetRPE float64) Workout {
	return Workout{
		ScheduledOn:     "2026-09-14",
		Name:            key,
		DurationMinutes: 45,
		TargetRPE:       targetRPE,
		Structure:       map[string]any{"protocol_key": key},
		Explanation:     map[string]any{"protocol_key": key},
	}
}

func selectionHistory(expected, missed int) []TrainingHistoryWindow {
	return []TrainingHistoryWindow{{
		WindowDays:                   28,
		ExpectedSessions:             expected,
		MissedSessions:               missed,
		PerformedSessions:            4,
		PerformedMinutes:             180,
		SessionRPELoad:               720,
		FeedbackRecords:              4,
		SessionsWithCompleteFeedback: 4,
	}}
}

func TestStimulusSelectionShadowRecordsAlignedNeed(t *testing.T) {
	assessment := assessStimulusSelectionShadow(Context{
		PrimaryGoal:      "performance",
		ExperienceLevel:  "advanced",
		BaselineEligible: true,
		TrainingHistory:  selectionHistory(4, 0),
	}, []Workout{
		selectionWorkout("base_endurance", 4),
		selectionWorkout("controlled_tempo", 6),
		selectionWorkout("long_endurance", 5),
	}, time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC), false)

	if assessment.Status != "observed" || assessment.CandidateNeed != "quality_progression" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("unexpected aligned selection assessment: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("stimulus selection shadow became authoritative: %+v", assessment)
	}
	if len(assessment.SelectedStimuli) != 3 || len(assessment.MissingData) != 0 || len(assessment.DataIssues) != 0 {
		t.Fatalf("unexpected selection evidence: %+v", assessment)
	}
}

func TestStimulusSelectionShadowFlagsLowAdherenceAndRecoveryMismatch(t *testing.T) {
	lowAdherence := assessStimulusSelectionShadow(Context{
		PrimaryGoal:     "performance",
		ExperienceLevel: "advanced",
		TrainingHistory: selectionHistory(4, 1),
	}, []Workout{selectionWorkout("road_vo2_intervals", 8)}, time.Now(), false)
	if lowAdherence.Status != "observed_mismatch" || !containsString(lowAdherence.DataIssues, "high_intensity_with_low_adherence") {
		t.Fatalf("low adherence mismatch was not recorded: %+v", lowAdherence)
	}

	recovery := assessStimulusSelectionShadow(Context{}, []Workout{selectionWorkout("road_vo2_intervals", 8)}, time.Now(), true)
	if recovery.Status != "observed_mismatch" || !containsString(recovery.DataIssues, "quality_in_recovery_need") || !containsString(recovery.DataIssues, "missing_recovery_stimulus") {
		t.Fatalf("recovery mismatch was not recorded: %+v", recovery)
	}
}

func TestStimulusSelectionShadowRecognizesGradualReturn(t *testing.T) {
	assessment := assessStimulusSelectionShadow(Context{
		ExperienceLevel: "advanced",
		Cycling:         CyclingContext{RecentTrainingWeeks: 2},
		TrainingHistory: selectionHistory(4, 0),
	}, []Workout{selectionWorkout("return_after_break", 3.5)}, time.Now(), false)
	if assessment.Status != "observed" || assessment.CandidateNeed != "return_to_training" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("unexpected gradual-return assessment: %+v", assessment)
	}
	if !containsString(assessment.ExpectedStimuli, "return_after_break") || len(assessment.DataIssues) != 0 || len(assessment.MissingData) != 0 {
		t.Fatalf("unexpected gradual-return evidence: %+v", assessment)
	}
}

func TestBuildPlanAttachesNonAuthoritativeStimulusSelectionShadow(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		ExperienceLevel: "advanced",
		PrimaryGoal:     "performance",
		Availability: []AvailabilitySlot{
			{Weekday: 2, AvailableMinutes: 60},
			{Weekday: 4, AvailableMinutes: 60},
			{Weekday: 6, AvailableMinutes: 90},
		},
	}, time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("buildPlan returned error: %v", err)
	}
	shadow, ok := plan.PrescriptionSnapshot["stimulus_selection_shadow"].(StimulusSelectionShadowAssessment)
	if !ok || shadow.Version != stimulusSelectionShadowVersion || shadow.Mode != stimulusSelectionShadowMode {
		t.Fatalf("stimulus selection shadow missing: %#v", plan.PrescriptionSnapshot["stimulus_selection_shadow"])
	}
	if shadow.Applied || shadow.UsedForPrescription || shadow.ProgressionEligible {
		t.Fatalf("stimulus selection shadow is authoritative: %+v", shadow)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
