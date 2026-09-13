package planning

import (
	"slices"
	"testing"
	"time"
)

func TestAssessPostWorkoutContextRecordsCompleteObservation(t *testing.T) {
	recoveryAfter := 4
	repeatConfidence := 5
	assessment := AssessPostWorkoutContext(CompletionInput{
		RecoveryAfter:    &recoveryAfter,
		RepeatConfidence: &repeatConfidence,
	}, time.Unix(0, 0))

	if assessment.Status != "observed" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("complete context = %+v", assessment)
	}
	if assessment.RecoveryAfter == nil || *assessment.RecoveryAfter != 4 || assessment.RepeatConfidence == nil || *assessment.RepeatConfidence != 5 {
		t.Fatalf("recorded values = recovery %v, confidence %v", assessment.RecoveryAfter, assessment.RepeatConfidence)
	}
	if !slices.Contains(assessment.ObservedFields, "recovery_after") || !slices.Contains(assessment.ObservedFields, "repeat_confidence") {
		t.Fatalf("observed fields = %#v", assessment.ObservedFields)
	}
	if len(assessment.MissingData) != 0 || len(assessment.DataIssues) != 0 {
		t.Fatalf("complete context has gaps = missing %#v, issues %#v", assessment.MissingData, assessment.DataIssues)
	}
	if assessment.ProgressionEligible || assessment.UsedForPrescription {
		t.Fatal("post-workout context must remain non-prescriptive")
	}
}

func TestAssessPostWorkoutContextMarksPartialObservation(t *testing.T) {
	recoveryAfter := 4
	assessment := AssessPostWorkoutContext(CompletionInput{RecoveryAfter: &recoveryAfter}, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("partial context produced a candidate = %+v", assessment)
	}
	if !slices.Contains(assessment.MissingData, "repeat_confidence") {
		t.Fatalf("missing partial field = %#v", assessment.MissingData)
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{
		Code:    "partial_post_workout_context",
		Message: "Somente parte dos sinais pós-treino foi registrada; o contexto permanece incompleto para análise conjunta.",
	}) {
		t.Fatalf("partial reason = %#v", assessment.Reasons)
	}
}

func TestAssessPostWorkoutContextRejectsInvalidValues(t *testing.T) {
	recoveryAfter := 0
	repeatConfidence := 6
	assessment := AssessPostWorkoutContext(CompletionInput{
		RecoveryAfter:    &recoveryAfter,
		RepeatConfidence: &repeatConfidence,
	}, time.Unix(0, 0))

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "not_evaluated" {
		t.Fatalf("invalid context produced a candidate = %+v", assessment)
	}
	for _, issue := range []string{"invalid_recovery_after", "invalid_repeat_confidence"} {
		if !slices.Contains(assessment.DataIssues, issue) {
			t.Fatalf("data issues did not include %q: %#v", issue, assessment.DataIssues)
		}
	}
	if !slices.Contains(assessment.Reasons, ReadinessReason{
		Code:    "invalid_post_workout_context",
		Message: "Os sinais pós-treino contêm valores fora da faixa e permanecem fora da interpretação observacional.",
	}) {
		t.Fatalf("invalid reason = %#v", assessment.Reasons)
	}
}

func TestAssessRulesV2AdaptationShadowAttachesPostWorkoutContext(t *testing.T) {
	recoveryAfter := 4
	repeatConfidence := 5
	assessment := assessRulesV2AdaptationShadow(6, CompletionInput{
		ActualRPE:        5,
		Difficulty:       "moderate",
		FatigueAfter:     3,
		RecoveryAfter:    &recoveryAfter,
		RepeatConfidence: &repeatConfidence,
	}, nil, time.Unix(0, 0))

	if assessment.PostWorkoutContext == nil {
		t.Fatal("post-workout context was not attached to the adaptation shadow")
	}
	if assessment.PostWorkoutContext.Status != "observed" || assessment.PostWorkoutContext.CandidateResponse != "maintain_observed" {
		t.Fatalf("attached context = %+v", assessment.PostWorkoutContext)
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatal("attaching context must not make adaptation authoritative")
	}
}
