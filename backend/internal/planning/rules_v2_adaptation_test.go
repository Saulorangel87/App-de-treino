package planning

import (
	"slices"
	"testing"
	"time"
)

func TestAssessRulesV2AdaptationShadowDefersSingleEasyResponse(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, nil, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("single easy response produced an unjustified candidate: %+v", assessment)
	}
	for _, missing := range []string{"period_comparison", "recent_period", "prior_period"} {
		if !slices.Contains(assessment.MissingData, missing) {
			t.Fatalf("missing evidence did not include %q: %+v", missing, assessment.MissingData)
		}
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("deferred progression became authoritative: %+v", assessment)
	}
}

func TestAssessRulesV2AdaptationShadowKeepsCompleteProgressionAsCandidate(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, validHistoryPeriods(), time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))

	if assessment.Status != "observation_only" || assessment.CandidateResponse != "progress_duration_5pct" {
		t.Fatalf("complete evidence did not produce the expected candidate: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("progression candidate became authoritative: %+v", assessment)
	}
	if len(assessment.MissingData) != 0 || len(assessment.DataIssues) != 0 {
		t.Fatalf("complete evidence was rejected: %+v", assessment)
	}
}

func TestAssessRulesV2AdaptationShadowPrioritizesPain(t *testing.T) {
	input := CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 1, PainReported: true}
	assessment := assessRulesV2AdaptationShadow(6, input, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.Status != "protective_signal" || assessment.CandidateResponse != "prefer_recovery" {
		t.Fatalf("pain did not produce a protective candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{Code: "post_workout_pain", Message: "Carga reduzida porque houve relato de dor após a sessão anterior."}) {
		t.Fatalf("missing pain reason: %+v", assessment.Reasons)
	}
}

func TestAssessRulesV2AdaptationShadowRejectsInconsistentEvidence(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PeriodDays = 6
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("inconsistent evidence produced a candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.DataIssues, "inconsistent_period_0") {
		t.Fatalf("inconsistent period was not reported: %+v", assessment.DataIssues)
	}
}

func TestAssessRulesV2AdaptationShadowDoesNotMaintainOnInconsistentEvidence(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PeriodDays = 6
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 6, Difficulty: "moderate", FatigueAfter: 3,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("neutral feedback produced a candidate with inconsistent evidence: %+v", assessment)
	}
}
