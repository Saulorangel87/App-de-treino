package planning

// AdaptationDecision describes a small, deterministic change to future sessions.
// A zero Sessions value means that the completed workout should not change the plan.
type AdaptationDecision struct {
	Kind             string
	Sessions         int
	DurationFactor   float64
	TargetRPEDelta   float64
	TargetRPECap     float64
	MinimumTargetRPE float64
	Reason           string
	SafetyNotice     string
}

// DecideAdaptation converts subjective post-workout feedback into a conservative
// product rule. It never increases intensity: a clearly easy response may add a
// small amount of duration, while pain or high strain reduces upcoming load.
func DecideAdaptation(targetRPE float64, input CompletionInput) AdaptationDecision {
	if input.PainReported {
		return AdaptationDecision{
			Kind: "safety", Sessions: 2, DurationFactor: 0.80,
			TargetRPECap: 3, MinimumTargetRPE: 2,
			Reason:       "Carga reduzida porque houve relato de dor após a sessão anterior.",
			SafetyNotice: "Interrompa o treino se a dor reaparecer. Dor persistente ou intensa deve ser avaliada por um profissional de saúde.",
		}
	}

	if input.ActualRPE >= 9 || input.FatigueAfter == 5 || input.Difficulty == "very_hard" {
		return AdaptationDecision{
			Kind: "recovery", Sessions: 2, DurationFactor: 0.80,
			TargetRPEDelta: -1, TargetRPECap: 4, MinimumTargetRPE: 3,
			Reason: "Carga reduzida porque a sessão anterior gerou esforço ou fadiga muito altos.",
		}
	}

	if input.ActualRPE >= targetRPE+2 || input.FatigueAfter >= 4 || input.Difficulty == "hard" {
		return AdaptationDecision{
			Kind: "recovery", Sessions: 1, DurationFactor: 0.90,
			TargetRPEDelta: -1, MinimumTargetRPE: 3,
			Reason: "A próxima carga foi reduzida porque o esforço percebido ficou acima do esperado.",
		}
	}

	if input.ActualRPE <= targetRPE-2 && input.FatigueAfter <= 2 && (input.Difficulty == "easy" || input.Difficulty == "very_easy") {
		return AdaptationDecision{
			Kind: "progression", Sessions: 1, DurationFactor: 1.05,
			Reason: "A próxima sessão recebeu uma progressão leve porque o treino anterior foi claramente mais fácil que o previsto.",
		}
	}

	return AdaptationDecision{}
}
