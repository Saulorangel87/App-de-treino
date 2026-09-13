package planning

import (
	"slices"
	"testing"
	"time"
)

func TestRulesV2AdaptationShadowNeverBecomesAuthoritative(t *testing.T) {
	invalidDuration := 0
	invalidInput := CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2}
	invalidIntegrity := AssessWorkoutDataIntegrity(WorkoutDataIntegrityInput{
		DurationMinutes: &invalidDuration,
		ActualRPE:       &invalidInput.ActualRPE,
		FeedbackPresent: true,
		Difficulty:      invalidInput.Difficulty,
		FatigueAfter:    &invalidInput.FatigueAfter,
	}, time.Unix(0, 0))

	cases := []struct {
		name       string
		assessment func() RulesV2AdaptationShadowAssessment
	}{
		{
			name: "invalid feedback",
			assessment: func() RulesV2AdaptationShadowAssessment {
				return assessRulesV2AdaptationShadow(0, CompletionInput{}, nil, time.Unix(0, 0))
			},
		},
		{
			name: "protective signal",
			assessment: func() RulesV2AdaptationShadowAssessment {
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2, PainReported: true,
				}, validHistoryPeriods(), time.Unix(0, 0))
			},
		},
		{
			name: "partial completion",
			assessment: func() RulesV2AdaptationShadowAssessment {
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
					CompletionStatus: "partial", PartialReason: "time_available_changed",
				}, validHistoryPeriods(), time.Unix(0, 0))
			},
		},
		{
			name: "incomplete evidence",
			assessment: func() RulesV2AdaptationShadowAssessment {
				periods := validHistoryPeriods()
				periods[0].CompleteRecoveryCheckins = 0
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
				}, periods, time.Unix(0, 0))
			},
		},
		{
			name: "low adherence",
			assessment: func() RulesV2AdaptationShadowAssessment {
				periods := validHistoryPeriods()
				periods[0].MissedSessions = 1
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
				}, periods, time.Unix(0, 0))
			},
		},
		{
			name: "inconsistent evidence",
			assessment: func() RulesV2AdaptationShadowAssessment {
				periods := validHistoryPeriods()
				periods[0].PeriodDays = 6
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
				}, periods, time.Unix(0, 0))
			},
		},
		{
			name: "current session integrity",
			assessment: func() RulesV2AdaptationShadowAssessment {
				return AssessRulesV2AdaptationShadowWithIntegrity(6, invalidInput, validHistoryPeriods(), invalidIntegrity, time.Unix(0, 0))
			},
		},
		{
			name: "complete candidate",
			assessment: func() RulesV2AdaptationShadowAssessment {
				return assessRulesV2AdaptationShadow(6, CompletionInput{
					ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
				}, validHistoryPeriods(), time.Unix(0, 0))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assessment := tc.assessment()
			if assessment.Mode != "shadow" || assessment.Scope != "post_workout_feedback" {
				t.Fatalf("unexpected shadow metadata: %+v", assessment)
			}
			if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
				t.Fatalf("shadow became authoritative: %+v", assessment)
			}
			if assessment.DecisionAudit == nil || assessment.DecisionAudit.UsedForPrescription {
				t.Fatalf("shadow audit became authoritative: %+v", assessment.DecisionAudit)
			}
			if !slices.Contains(assessment.RulesEvaluated, "prescription_isolation_gate") {
				t.Fatalf("prescription isolation was not evaluated: %#v", assessment.RulesEvaluated)
			}
		})
	}
}
