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
	for _, field := range []string{"target_rpe", "actual_rpe", "recovery_after", "repeat_confidence"} {
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
