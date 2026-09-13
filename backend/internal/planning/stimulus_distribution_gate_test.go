package planning

import (
	"slices"
	"testing"
	"time"
)

func qualityStimulusHistoryPeriods() []TrainingHistoryPeriod {
	periods := validHistoryPeriods()
	periods[0].ExpectedSessions = 3
	periods[0].ScheduledCompletedSessions = 3
	periods[0].CancelledSessions = 0
	periods[0].PerformedSessions = 3
	periods[0].PerformedMinutes = 120
	periods[0].SessionsWithSessionRPELoad = 3
	periods[0].SessionRPELoad = 600
	periods[0].FeedbackRecords = 3
	periods[0].SessionsWithCompleteFeedback = 3
	periods[0].QualitySessions = 2
	periods[0].QualitySessionsWithLoad = 2
	periods[0].QualityPerformedMinutes = 80
	periods[0].HighIntensitySessions = 1
	periods[0].QualitySessionDates = []string{"2026-09-12", "2026-09-13"}
	return periods
}

func TestRulesV2ShadowObservesStimulusDensityAndSpacing(t *testing.T) {
	assessment := assessRulesV2Shadow(Context{TrainingHistoryPeriods: qualityStimulusHistoryPeriods()}, time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))

	if assessment.Status != "protective_signal" || assessment.CandidateResponse != "prefer_recovery" {
		t.Fatalf("stimulus distribution did not produce a protective shadow response: %+v", assessment)
	}
	if assessment.StimulusDistribution == nil || assessment.StimulusDistribution.QualitySessionsLast7d != 2 || assessment.StimulusDistribution.AdjacentQualitySessionPairs != 1 {
		t.Fatalf("stimulus distribution was not attached: %+v", assessment.StimulusDistribution)
	}
	for _, code := range []string{"recent_quality_session_proximity", "recent_quality_session_density"} {
		if !slices.ContainsFunc(assessment.Reasons, func(reason ReadinessReason) bool { return reason.Code == code }) {
			t.Fatalf("missing stimulus reason %q: %+v", code, assessment.Reasons)
		}
	}
	if !slices.Contains(assessment.RulesEvaluated, "stimulus_distribution_gate") {
		t.Fatalf("stimulus gate was not evaluated: %#v", assessment.RulesEvaluated)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription || assessment.StimulusDistribution.UsedForPrescription {
		t.Fatalf("stimulus shadow became authoritative: %+v", assessment)
	}
}

func TestRulesV2ShadowDefersWhenStimulusDateCoverageIsIncomplete(t *testing.T) {
	periods := qualityStimulusHistoryPeriods()
	periods[0].QualitySessionDates = nil
	assessment := assessRulesV2Shadow(Context{TrainingHistoryPeriods: periods}, time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("incomplete stimulus distribution produced a candidate: %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "quality_session_date_coverage_42d") ||
		!slices.Contains(assessment.Reasons, ReadinessReason{Code: "insufficient_stimulus_distribution", Message: "A distribuição observada dos estímulos ainda não tem cobertura ou consistência suficiente para avaliar densidade e espaçamento."}) {
		t.Fatalf("stimulus coverage gap was not preserved: %+v", assessment)
	}
}

func TestRulesV2AdaptationShadowUsesStimulusGateWithoutApplyingIt(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, qualityStimulusHistoryPeriods(), time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))

	if assessment.Status != "protective_signal" || assessment.CandidateResponse != "prefer_recovery" {
		t.Fatalf("stimulus distribution did not block progression: %+v", assessment)
	}
	if assessment.DecisionAudit == nil || !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "stimulus_distribution_gate") ||
		!slices.Contains(assessment.DecisionAudit.DataUsed, "stimulus_distribution") {
		t.Fatalf("stimulus gate was not auditable: %+v", assessment.DecisionAudit)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("stimulus adaptation shadow became authoritative: %+v", assessment)
	}
}

func TestRulesV2AdaptationShadowDefersWhenStimulusDistributionIsIncomplete(t *testing.T) {
	periods := qualityStimulusHistoryPeriods()
	periods[0].QualitySessionDates = nil
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("incomplete stimulus distribution did not defer progression: %+v", assessment)
	}
	if assessment.DecisionAudit == nil || !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "stimulus_distribution_gate") ||
		!slices.Contains(assessment.DecisionAudit.ConditionsForChange, "obter_cobertura_completa_da_distribuicao_de_estimulos") {
		t.Fatalf("stimulus coverage gate was not auditable: %+v", assessment.DecisionAudit)
	}
}

func TestStimulusDistributionWithNoQualitySessionsDoesNotBlock(t *testing.T) {
	distribution := buildTrainingStimulusDistribution(validHistoryPeriods(), time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))
	if stimulusDistributionRequiresRecovery(distribution) || stimulusDistributionHasIncompleteData(distribution) {
		t.Fatalf("absence of quality sessions was treated as a gate: %+v", distribution)
	}
}
