package planning

import "testing"

func TestDecideAdaptationPrioritizesPain(t *testing.T) {
	decision := DecideAdaptation(6, CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 1, PainReported: true})
	if decision.Kind != "safety" || decision.Sessions != 2 || decision.TargetRPECap != 3 || decision.DurationFactor != 0.80 {
		t.Fatalf("unexpected pain adaptation: %#v", decision)
	}
}

func TestDecideAdaptationReducesLoadAfterHighStrain(t *testing.T) {
	decision := DecideAdaptation(5, CompletionInput{ActualRPE: 7, Difficulty: "hard", FatigueAfter: 4})
	if decision.Kind != "recovery" || decision.Sessions != 1 || decision.TargetRPEDelta != -1 || decision.DurationFactor != 0.90 {
		t.Fatalf("unexpected recovery adaptation: %#v", decision)
	}
}

func TestDecideAdaptationOnlyProgressesClearlyEasyResponse(t *testing.T) {
	decision := DecideAdaptation(6, CompletionInput{ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2})
	if decision.Kind != "progression" || decision.Sessions != 1 || decision.DurationFactor != 1.05 || decision.TargetRPEDelta != 0 {
		t.Fatalf("unexpected progression adaptation: %#v", decision)
	}

	neutral := DecideAdaptation(6, CompletionInput{ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3})
	if neutral.Sessions != 0 {
		t.Fatalf("a normal response must keep the plan unchanged: %#v", neutral)
	}
}
