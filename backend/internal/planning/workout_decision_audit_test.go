package planning

import (
	"slices"
	"testing"
)

func TestBuildWorkoutDecisionAuditExplainsSafetyAndMissingContext(t *testing.T) {
	audit := buildWorkoutDecisionAudit(
		Context{
			ExperienceLevel: "advanced",
			PrimaryGoal:     "performance",
			Cycling:         CyclingContext{UsesHeartRate: true},
		},
		"base",
		[]string{"Intensidade limitada por uma condição de segurança ativa."},
		true,
		false,
		false,
		true,
		false,
		false,
	)

	if audit.Version != workoutDecisionAuditVersion || audit.Mode != "deterministic" || !audit.UsedForPrescription {
		t.Fatalf("unexpected audit identity: %#v", audit)
	}
	for _, expected := range []string{"availability_gate", "safety_precedence_gate", "protocol_eligibility_gate"} {
		if !slices.Contains(audit.RulesEvaluated, expected) {
			t.Fatalf("missing evaluated rule %q: %#v", expected, audit.RulesEvaluated)
		}
	}
	for _, expected := range []string{"active_safety_limitation", "recovery_week"} {
		if !slices.Contains(audit.ConstraintsApplied, expected) {
			t.Fatalf("missing constraint %q: %#v", expected, audit.ConstraintsApplied)
		}
	}
	for _, expected := range []string{"observed_training_28d", "power_meter_or_ftp", "eligible_submaximal_assessment"} {
		if !slices.Contains(audit.MissingData, expected) {
			t.Fatalf("missing data indicator %q: %#v", expected, audit.MissingData)
		}
	}
	for _, expected := range []string{"higher_intensity_protocols", "additional_quality_session"} {
		if !slices.Contains(audit.AlternativesRejected, expected) {
			t.Fatalf("missing rejected alternative %q: %#v", expected, audit.AlternativesRejected)
		}
	}
}
