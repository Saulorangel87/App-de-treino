package planning

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func TestAdaptationDecisionAuditExplainsShadowLimits(t *testing.T) {
	recoveryAfter := 4
	repeatConfidence := 5
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE:        5,
		Difficulty:       "moderate",
		FatigueAfter:     3,
		RecoveryAfter:    &recoveryAfter,
		RepeatConfidence: &repeatConfidence,
	}, validHistoryPeriods(), time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("shadow assessment did not include decision audit")
	}
	audit := assessment.DecisionAudit
	if audit.Version != "adaptation-audit-v1" || audit.Confidence != "not_calibrated" {
		t.Fatalf("audit metadata = %+v", audit)
	}
	for _, field := range []string{"target_rpe", "actual_rpe", "recovery_after", "repeat_confidence", "training_history_periods"} {
		if !slices.Contains(audit.DataUsed, field) {
			t.Fatalf("audit data used did not include %q: %#v", field, audit.DataUsed)
		}
	}
	if !slices.Contains(audit.ConstraintsApplied, "rules_v1_prescription_isolation") || !slices.Contains(audit.AlternativesRejected, "progression") {
		t.Fatalf("audit constraints/alternatives = %#v / %#v", audit.ConstraintsApplied, audit.AlternativesRejected)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription || audit.UsedForPrescription {
		t.Fatal("audit must not authorize the shadow assessment")
	}
}

func TestAdaptationDecisionAuditPreservesMissingEvidence(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, nil, time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("shadow assessment did not include decision audit")
	}
	if !slices.Contains(assessment.DecisionAudit.MissingData, "period_comparison") {
		t.Fatalf("audit missing data = %#v", assessment.DecisionAudit.MissingData)
	}
	if !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "insufficient_evidence_gate") {
		t.Fatalf("audit constraints = %#v", assessment.DecisionAudit.ConstraintsApplied)
	}
	if !slices.Contains(assessment.DecisionAudit.ConditionsForChange, "obter_dados_minimos_de_carga_feedback_e_recuperacao") {
		t.Fatalf("audit conditions = %#v", assessment.DecisionAudit.ConditionsForChange)
	}
	if !slices.Contains(assessment.DecisionAudit.MissingData, "recovery_after") || !slices.Contains(assessment.DecisionAudit.MissingData, "repeat_confidence") {
		t.Fatalf("nested context gaps were not preserved: %#v", assessment.DecisionAudit.MissingData)
	}
}

func TestAdaptationDecisionAuditRecordsLoadToleranceGate(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].AboveTargetRPESessions = 1
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("shadow assessment did not include decision audit")
	}
	if !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "load_tolerance_gate") {
		t.Fatalf("load tolerance gate was not recorded: %#v", assessment.DecisionAudit.ConstraintsApplied)
	}
}

func TestAdaptationDecisionAuditRecordsIncompleteLoadToleranceGate(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].CompleteRecoveryCheckins = 0
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2,
	}, periods, time.Unix(0, 0))

	if assessment.LoadTolerance == nil || assessment.LoadTolerance.Status != "not_evaluated" {
		t.Fatalf("incomplete load tolerance was not preserved: %+v", assessment.LoadTolerance)
	}
	if assessment.CandidateResponse != "defer_progression" {
		t.Fatalf("incomplete load tolerance did not defer progression: %+v", assessment)
	}
	if assessment.DecisionAudit == nil || !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "load_tolerance_gate") {
		t.Fatalf("incomplete load tolerance gate was not recorded: %+v", assessment.DecisionAudit)
	}
}

func TestAdaptationDecisionAuditDoesNotClaimInvalidFeedbackWasUsed(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(0, CompletionInput{
		ActualRPE: 0, Difficulty: "", FatigueAfter: 0,
	}, nil, time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("shadow assessment did not include decision audit")
	}
	for _, field := range []string{"target_rpe", "actual_rpe", "difficulty", "fatigue_after", "pain_reported", "completion_status", "training_history_periods"} {
		if slices.Contains(assessment.DecisionAudit.DataUsed, field) {
			t.Fatalf("invalid feedback was reported as used field %q: %#v", field, assessment.DecisionAudit.DataUsed)
		}
	}
}

func TestAdaptationDecisionAuditUsesStableJSONField(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
	}, nil, time.Unix(0, 0))
	payload, err := json.Marshal(assessment)
	if err != nil {
		t.Fatalf("marshal shadow assessment: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal shadow assessment: %v", err)
	}
	audit, ok := decoded["decision_audit"].(map[string]any)
	if !ok {
		t.Fatalf("decision_audit JSON = %#v", decoded["decision_audit"])
	}
	if audit["version"] != "adaptation-audit-v1" || audit["used_for_prescription"] != false {
		t.Fatalf("decision_audit JSON = %#v", audit)
	}
}

func TestRefreshAdaptationDecisionAuditIncludesAttachedObservations(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
	}, validHistoryPeriods(), time.Unix(0, 0))
	planned := AssessPlannedVsActual(PlannedVsActualInput{
		PlannedDurationMinutes: 35,
		ActualDurationMinutes:  3,
		TargetRPE:              6,
		ActualRPE:              5,
		FeedbackPresent:        true,
		CompletionStatus:       "complete",
		Difficulty:             "moderate",
		FatigueAfter:           3,
	}, time.Unix(0, 0))
	assessment.PlannedVsActual = &planned
	RefreshAdaptationDecisionAudit(&assessment, 6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
	}, time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("refreshed audit was not attached")
	}
	for _, missing := range []string{"average_power_watts", "average_heart_rate"} {
		if !slices.Contains(assessment.DecisionAudit.MissingData, missing) {
			t.Fatalf("attached planned-vs-actual gap %q was not preserved: %#v", missing, assessment.DecisionAudit.MissingData)
		}
	}
}

func TestRefreshAdaptationDecisionAuditDoesNotClaimHistoryAfterQueryFailure(t *testing.T) {
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
	}, validHistoryPeriods(), time.Unix(0, 0))
	assessment.DataIssues = append(assessment.DataIssues, "history_query_failed")
	assessment.Status = "not_evaluated"
	assessment.CandidateResponse = "not_evaluated"
	RefreshAdaptationDecisionAudit(&assessment, 6, CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
	}, time.Unix(0, 0))

	if assessment.DecisionAudit == nil {
		t.Fatal("refreshed audit was not attached")
	}
	if slices.Contains(assessment.DecisionAudit.DataUsed, "training_history_periods") {
		t.Fatalf("failed history query was reported as used data: %#v", assessment.DecisionAudit.DataUsed)
	}
	if !slices.Contains(assessment.DecisionAudit.ConstraintsApplied, "history_query_gate") {
		t.Fatalf("history query gate was not recorded: %#v", assessment.DecisionAudit.ConstraintsApplied)
	}
}
