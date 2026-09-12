package planning

import (
	"slices"
	"testing"
	"time"
)

func TestAssessLoadToleranceRequiresCompleteRecentEvidence(t *testing.T) {
	assessment := AssessLoadTolerance(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 2,
	}, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.Status != "observation_only" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("complete recent evidence was not observed: %+v", assessment)
	}
	if got, want := assessment.EvidencePeriods, []string{"last_7d", "days_8_14"}; !slices.Equal(got, want) {
		t.Fatalf("evidence periods = %v, want %v", got, want)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("load tolerance became prescriptive: %+v", assessment)
	}
}

func TestAssessLoadToleranceReportsMissingEvidence(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].SessionsWithSessionRPELoad = 0
	periods[0].SessionsWithoutSessionRPELoad = 1
	assessment := AssessLoadTolerance(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("missing evidence produced a tolerance classification: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "recent_session_rpe_load") {
		t.Fatalf("missing session-RPE evidence was not reported: %+v", assessment.MissingData)
	}
}

func TestAssessLoadToleranceProtectsCurrentHighEffort(t *testing.T) {
	assessment := AssessLoadTolerance(6, CompletionInput{
		ActualRPE: 8, Difficulty: "very_hard", FatigueAfter: 3,
	}, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.Status != "protective_signal" || assessment.CandidateResponse != "prefer_recovery" {
		t.Fatalf("current high effort did not protect tolerance: %+v", assessment)
	}
	if !slices.ContainsFunc(assessment.Reasons, func(reason ReadinessReason) bool {
		return reason.Code == "current_above_target_rpe"
	}) {
		t.Fatalf("missing current high-effort reason: %+v", assessment.Reasons)
	}
}

func TestAssessLoadToleranceRejectsInconsistentPeriods(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PeriodDays = 6
	assessment := AssessLoadTolerance(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("inconsistent periods produced a tolerance classification: %+v", assessment)
	}
	if !slices.Contains(assessment.DataIssues, "inconsistent_period_0") {
		t.Fatalf("inconsistent period was not reported: %+v", assessment.DataIssues)
	}
}

func TestAdaptationShadowIncludesNonAuthoritativeLoadTolerance(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 2,
	}, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.LoadTolerance == nil || assessment.LoadTolerance.Version != loadToleranceVersion {
		t.Fatalf("load tolerance was not attached to adaptation shadow: %+v", assessment)
	}
	if assessment.LoadTolerance.UsedForPrescription || assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("observational fields became authoritative: %+v", assessment)
	}
	if slices.Contains(assessment.NotEvaluated, "load_tolerance") {
		t.Fatalf("load tolerance remained globally unevaluated after the shadow was attached: %+v", assessment.NotEvaluated)
	}
}
