package planning

import (
	"slices"
	"testing"
	"time"
)

func TestPlanningCoherenceShadowObservesAlignedComponents(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	workouts := coherentPeriodizationWorkouts()
	periodization := assessPeriodizationShadow(workouts, now, now)
	distribution := buildTrainingStimulusDistribution(validHistoryPeriods(), now)
	selection := assessStimulusSelectionShadow(Context{
		PrimaryGoal:      "performance",
		ExperienceLevel:  "advanced",
		BaselineEligible: true,
		TrainingHistory:  selectionHistory(4, 0),
	}, workouts, now, false)

	assessment := assessPlanningCoherenceShadow(periodization, distribution, selection, now)

	if assessment.Status != "observed" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("unexpected coherent planning assessment: %+v", assessment)
	}
	if assessment.CandidateNeed != "quality_progression" || !slices.Contains(assessment.CoherenceChecks, "quality_need_matches_periodization") {
		t.Fatalf("quality coherence was not recorded: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("planning coherence became authoritative: %+v", assessment)
	}
	if !slices.Contains(assessment.RulesEvaluated, "prescription_isolation_gate") {
		t.Fatalf("prescription isolation was not evaluated: %+v", assessment.RulesEvaluated)
	}
}

func TestPlanningCoherenceShadowFlagsRecoveryConflict(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	workouts := coherentPeriodizationWorkouts()
	workouts[9] = periodizationWorkout("2026-10-05", "Tempo controlado", "controlled_tempo", 60, 6)
	periodization := assessPeriodizationShadow(workouts, now, now)
	distribution := buildTrainingStimulusDistribution(validHistoryPeriods(), now)
	selection := assessStimulusSelectionShadow(Context{
		Observed: ObservedTrainingSummary{AverageFatigue: 4},
	}, workouts, now, false)

	assessment := assessPlanningCoherenceShadow(periodization, distribution, selection, now)

	if assessment.Status != "observed_mismatch" || assessment.CandidateResponse != "review_coherence" {
		t.Fatalf("recovery conflict was not flagged: %+v", assessment)
	}
	if !slices.Contains(assessment.DataIssues, "periodization_shadow_inconsistent") {
		t.Fatalf("periodization conflict was not preserved: %+v", assessment.DataIssues)
	}
	if assessment.UsedForPrescription {
		t.Fatal("recovery conflict became authoritative")
	}
}

func TestPlanningCoherenceShadowDefersIncompleteDistribution(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	workouts := coherentPeriodizationWorkouts()
	periodization := assessPeriodizationShadow(workouts, now, now)
	distribution := buildTrainingStimulusDistribution(nil, now)
	selection := assessStimulusSelectionShadow(Context{
		PrimaryGoal:     "health",
		ExperienceLevel: "beginner",
		TrainingHistory: selectionHistory(4, 0),
	}, workouts, now, false)

	assessment := assessPlanningCoherenceShadow(periodization, distribution, selection, now)

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_evaluation" {
		t.Fatalf("incomplete distribution produced a candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "stimulus_distribution_stimulus_distribution_periods") {
		t.Fatalf("distribution gap was not preserved: %+v", assessment.MissingData)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatal("incomplete coherence became authoritative")
	}
}

func TestBuildPlanAttachesPlanningCoherenceWithoutReplacingRulesV1(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		PrimaryGoal:     "health",
		ExperienceLevel: "beginner",
		Availability: []AvailabilitySlot{
			{Weekday: 1, AvailableMinutes: 45},
			{Weekday: 3, AvailableMinutes: 60},
			{Weekday: 6, AvailableMinutes: 90},
		},
	}, now)
	if err != nil {
		t.Fatalf("buildPlan returned error: %v", err)
	}
	shadow, ok := plan.PrescriptionSnapshot["planning_coherence_shadow"].(PlanningCoherenceShadowAssessment)
	if !ok || shadow.Mode != planningCoherenceShadowMode || shadow.UsedForPrescription {
		t.Fatalf("planning coherence shadow missing or authoritative: %#v", plan.PrescriptionSnapshot["planning_coherence_shadow"])
	}
	if plan.PrescriptionSnapshot["engine_version"] != "rules-v1" {
		t.Fatalf("rules-v1 was replaced: %#v", plan.PrescriptionSnapshot["engine_version"])
	}
}
