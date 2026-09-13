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

func TestAssessRulesV2AdaptationShadowRequiresRecoveryInEachTolerancePeriod(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].CompleteRecoveryCheckins = 0
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("missing recent recovery evidence produced a progression candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "recent_complete_recovery_checkin") {
		t.Fatalf("load-tolerance recovery gap was not propagated: %#v", assessment.MissingData)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{
		Code:    "progression_deferred_load_tolerance",
		Message: "A avaliação de tolerância à carga ainda não está completa em cada período exigido; a progressão permanece adiada.",
	}) {
		t.Fatalf("missing load-tolerance deferral reason: %#v", assessment.Reasons)
	}
}

func TestAssessRulesV2AdaptationShadowDefersAfterRecentMissedWorkout(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].ScheduledCompletedSessions = 1
	periods[0].CancelledSessions = 0
	periods[0].MissedSessions = 1
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("recent missed workout produced an unexpected response: %+v", assessment)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{
		Code:    "low_adherence",
		Message: "Há treino previsto perdido ou em andamento vencido nos períodos usados; a progressão permanece adiada até a aderência ser observada com mais consistência.",
	}) {
		t.Fatalf("missing low-adherence reason: %#v", assessment.Reasons)
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

func TestAssessRulesV2AdaptationShadowDefersPartialCompletion(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
		CompletionStatus: "partial", PartialReason: "time_available_changed",
	}, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("partial completion produced an adaptation candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{
		Code:    "partial_completion",
		Message: "A sessão foi concluída parcialmente; ela pode ser observada, mas não é evidência de tolerância ao treino completo e não libera progressão.",
	}) {
		t.Fatalf("missing partial-completion reason: %+v", assessment.Reasons)
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

func TestAssessRulesV2AdaptationShadowDefersWhenCurrentSessionIsNotEligible(t *testing.T) {
	input := CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2}
	duration := 0
	integrity := AssessWorkoutDataIntegrity(WorkoutDataIntegrityInput{
		DurationMinutes: &duration,
		ActualRPE:       &input.ActualRPE,
		FeedbackPresent: true,
		Difficulty:      input.Difficulty,
		FatigueAfter:    &input.FatigueAfter,
	}, time.Unix(0, 0))
	assessment := AssessRulesV2AdaptationShadowWithIntegrity(6, input, validHistoryPeriods(), integrity, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("ineligible current session produced an adaptation candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.DataIssues, "current_session_data_integrity") {
		t.Fatalf("current-session integrity issue was not propagated: %+v", assessment.DataIssues)
	}
}
