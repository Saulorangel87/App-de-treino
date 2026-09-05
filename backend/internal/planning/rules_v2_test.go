package planning

import (
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestAssessRulesV2ShadowKeepsCompleteEvidenceObservational(t *testing.T) {
	assessment := assessRulesV2Shadow(Context{TrainingHistoryPeriods: validHistoryPeriods()}, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))

	if assessment.Version != rulesV2Version || assessment.Mode != rulesV2Mode ||
		assessment.Status != "observation_only" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("unexpected shadow assessment: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatalf("shadow assessment became authoritative: %+v", assessment)
	}
	if len(assessment.MissingData) != 0 || len(assessment.DataIssues) != 0 {
		t.Fatalf("complete evidence was not accepted: %+v", assessment)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{Code: "shadow_only", Message: "A resposta candidata não é aplicada; o rules-v1 continua gerando o plano."}) {
		t.Fatalf("missing shadow-only reason: %+v", assessment.Reasons)
	}
}

func TestAssessRulesV2ShadowPrioritizesRecentProtectiveSignals(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PainReportedSessions = 1
	assessment := assessRulesV2Shadow(Context{TrainingHistoryPeriods: periods}, time.Unix(0, 0))

	if assessment.Status != "protective_signal" || assessment.CandidateResponse != "prefer_recovery" {
		t.Fatalf("recent pain did not produce a protective shadow response: %+v", assessment)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{Code: "recent_pain", Message: "O período mais recente contém dor após sessão; a carga não deve ser aumentada."}) {
		t.Fatalf("missing recent pain reason: %+v", assessment.Reasons)
	}
	if assessment.Applied || assessment.UsedForPrescription || assessment.ProgressionEligible {
		t.Fatalf("protective response must remain unapplied: %+v", assessment)
	}
}

func TestAssessRulesV2ShadowDoesNotInferProgressionFromMissingPeriods(t *testing.T) {
	assessment := assessRulesV2Shadow(Context{}, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("missing periods produced an unjustified response: %+v", assessment)
	}
	for _, missing := range []string{"period_comparison", "recent_period", "prior_period"} {
		if !slices.Contains(assessment.MissingData, missing) {
			t.Fatalf("missing data did not include %q: %+v", missing, assessment.MissingData)
		}
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{Code: "insufficient_progression_evidence", Message: "Ainda não há duas semanas recentes com carga, feedback e recuperação completos para avaliar progressão."}) {
		t.Fatalf("missing insufficient-evidence reason: %+v", assessment.Reasons)
	}
}

func TestAssessRulesV2ShadowRejectsInconsistentPeriodData(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PeriodDays = 6
	assessment := assessRulesV2Shadow(Context{TrainingHistoryPeriods: periods}, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || !slices.Contains(assessment.DataIssues, "inconsistent_period_0") {
		t.Fatalf("inconsistent period was not rejected: %+v", assessment)
	}
}

func TestBuildPlanStoresRulesV2ShadowWithoutChangingRulesV1(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:              "profile-1",
		ExperienceLevel:        "intermediate",
		PrimaryGoal:            "endurance",
		Availability:           []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		TrainingHistoryPeriods: validHistoryPeriods(),
	}, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	shadow, ok := plan.PrescriptionSnapshot["rules_v2_shadow"].(RulesV2ShadowAssessment)
	if !ok || shadow.Version != rulesV2Version || shadow.Applied || shadow.UsedForPrescription {
		t.Fatalf("rules-v2 shadow missing or authoritative: %#v", plan.PrescriptionSnapshot["rules_v2_shadow"])
	}
	if plan.PrescriptionSnapshot["engine_version"] != "rules-v1" {
		t.Fatalf("rules-v1 was replaced: %#v", plan.PrescriptionSnapshot["engine_version"])
	}
}

func TestRulesV2ShadowDoesNotChangeRulesV1Workouts(t *testing.T) {
	input := Context{
		ProfileID:              "profile-1",
		ExperienceLevel:        "advanced",
		PrimaryGoal:            "performance",
		BaselineEligible:       true,
		Availability:           []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		TrainingHistoryPeriods: validHistoryPeriods(),
	}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	withShadow, err := buildPlan(input, now)
	if err != nil {
		t.Fatal(err)
	}
	input.TrainingHistoryPeriods = nil
	withoutShadow, err := buildPlan(input, now)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(withShadow.Workouts, withoutShadow.Workouts) {
		t.Fatalf("shadow evaluation changed rules-v1 workouts:\nwith=%#v\nwithout=%#v", withShadow.Workouts, withoutShadow.Workouts)
	}
}
