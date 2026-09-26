package planning

import (
	"fmt"
	"time"
)

func hasMedicalRestriction(limitations []LimitationContext) bool {
	for _, limitation := range limitations {
		if limitation.MedicalRestriction {
			return true
		}
	}
	return false
}

func hasRecentSurgery(limitations []LimitationContext) bool {
	for _, limitation := range limitations {
		if limitation.RecentSurgery {
			return true
		}
	}
	return false
}

func hasExerciseProhibited(limitations []LimitationContext) bool {
	for _, limitation := range limitations {
		if limitation.ExerciseProhibited {
			return true
		}
	}
	return false
}

func hasConditionAffectingExercise(limitations []LimitationContext) bool {
	for _, limitation := range limitations {
		if limitation.ConditionAffectingExercise {
			return true
		}
	}
	return false
}

func isLowCurrentActivity(value *string) bool {
	return value != nil && (*value == "sedentary" || *value == "occasional")
}

func qualityGoalFor(input Context) string {
	if input.PrimaryGoal == "performance" || input.PrimaryGoal == "event" {
		return input.PrimaryGoal
	}
	if input.SecondaryGoal == "performance" || input.SecondaryGoal == "event" {
		return input.SecondaryGoal
	}
	return input.PrimaryGoal
}

func preferredQualityPreference(context CyclingContext) string {
	if len(context.PreferredSessionTypes) == 0 || len(context.PreferredSessionTypes) >= 7 {
		return ""
	}
	for _, preference := range []string{"vo2max", "short_intervals", "threshold", "intervals", "sweet_spot", "hills", "cadence"} {
		for _, selected := range context.PreferredSessionTypes {
			if selected == preference {
				return preference
			}
		}
	}
	return ""
}

func sessionPreferenceLabel(preference string) string {
	labels := map[string]string{"cadence": "cadência", "hills": "subidas", "intervals": "intervalos", "threshold": "limiar", "sweet_spot": "sweet spot", "vo2max": "VO₂max", "short_intervals": "intervalos curtos"}
	return labels[preference]
}

func isReturningAfterPause(cycling CyclingContext) bool {
	return cycling.TrainingStatus == "returning_after_break"
}

func buildStructure(duration int, targetRPE float64, name, mainBlock string) map[string]any {
	warmup := minInt(10, maxInt(5, duration/6))
	cooldown := minInt(10, maxInt(5, duration/8))
	mainMinutes := maxInt(1, duration-warmup-cooldown)
	steps := []WorkoutStep{
		{Order: 1, Kind: "warmup", Title: "Aquecimento", DurationMinutes: warmup, TargetRPE: 3, Instruction: "Pedale de forma confortável e aumente o ritmo aos poucos."},
	}

	protocol := protocolForWorkout(name)
	var mainSteps []WorkoutStep
	if protocol.Repetitions > 1 {
		mainSteps = repeatedSteps(mainMinutes, protocol.WorkMinutes, protocol.RecoveryMinutes, protocol.Repetitions, targetRPE, protocol.WorkTitle, protocol.WorkInstruction)
	} else {
		mainSteps = []WorkoutStep{{Kind: "main", Title: "Parte principal", DurationMinutes: mainMinutes, TargetRPE: targetRPE, Instruction: mainBlock + "."}}
	}

	for index := range mainSteps {
		mainSteps[index].Order = len(steps) + 1
		steps = append(steps, mainSteps[index])
	}
	steps = append(steps, WorkoutStep{
		Order: len(steps) + 1, Kind: "cooldown", Title: "Desaquecimento", DurationMinutes: cooldown,
		TargetRPE: 2.5, Instruction: "Reduza o ritmo gradualmente e termine pedalando leve.",
	})

	return map[string]any{
		"warmup_minutes":   warmup,
		"main":             mainBlock,
		"cooldown_minutes": cooldown,
		"protocol_key":     protocol.Key,
		"steps":            steps,
	}
}

func repeatedSteps(mainMinutes, workMinutes, recoveryMinutes, repetitions int, targetRPE float64, title, instruction string) []WorkoutStep {
	if mainMinutes <= 0 {
		return nil
	}
	for repetitions > 1 && repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		repetitions--
	}
	for workMinutes > 3 && repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		workMinutes--
	}
	if repetitions < 2 || repetitions*workMinutes+(repetitions-1)*recoveryMinutes > mainMinutes {
		return []WorkoutStep{{Kind: "main", Title: "Parte principal", DurationMinutes: mainMinutes, TargetRPE: targetRPE, Instruction: instruction}}
	}

	steps := make([]WorkoutStep, 0, repetitions*2+1)
	used := 0
	for index := 0; index < repetitions; index++ {
		steps = append(steps, WorkoutStep{Kind: "work", Title: fmt.Sprintf("%s %d de %d", title, index+1, repetitions), DurationMinutes: workMinutes, TargetRPE: targetRPE, Instruction: instruction})
		used += workMinutes
		if index < repetitions-1 {
			steps = append(steps, WorkoutStep{Kind: "recovery", Title: "Recuperação leve", DurationMinutes: recoveryMinutes, TargetRPE: 2.5, Instruction: "Pedale bem leve até recuperar a respiração antes do próximo bloco."})
			used += recoveryMinutes
		}
	}
	if remaining := mainMinutes - used; remaining > 0 {
		steps = append(steps, WorkoutStep{Kind: "easy", Title: "Pedal leve contínuo", DurationMinutes: remaining, TargetRPE: 3, Instruction: "Complete o tempo restante em ritmo confortável, sem forçar."})
	}
	return steps
}

func nextMonday(value time.Time) time.Time {
	local := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	if local.Weekday() == time.Sunday {
		return local.AddDate(0, 0, 1)
	}
	return local.AddDate(0, 0, -((int(local.Weekday()) + 6) % 7))
}

func weekdayOffset(weekday int) int { return (weekday + 6) % 7 }

func eventSpecificPhase(cycling CyclingContext, workoutDate time.Time) bool {
	daysUntilEvent, ok := daysUntilEvent(cycling, workoutDate)
	return ok && daysUntilEvent >= 0 && daysUntilEvent <= 42
}

func longestSlot(slots []AvailabilitySlot) int {
	longest := 0
	for index := range slots {
		if slots[index].AvailableMinutes > slots[longest].AvailableMinutes {
			longest = index
		}
	}
	return longest
}

func intensitySlot(slots []AvailabilitySlot, longIndex int) int {
	for index := range slots {
		if index != longIndex {
			return index
		}
	}
	return -1
}

func objectiveFor(goal string) string {
	labels := map[string]string{
		"health":            "Construir saúde cardiovascular com consistência",
		"fitness":           "Desenvolver condicionamento geral",
		"endurance":         "Aumentar resistência para pedais mais longos",
		"performance":       "Elevar a capacidade de sustentar esforço",
		"event":             "Criar base específica para o evento-alvo",
		"weight_management": "Apoiar gasto energético com carga sustentável",
	}
	return labels[goal]
}

func explanationFor(kind string, restricted bool) string {
	if restricted {
		return "Sessão deliberadamente leve para respeitar a limitação informada."
	}
	if kind == "long" {
		return "O maior período disponível da semana recebe o estímulo principal de resistência."
	}
	if kind == "quality" {
		return "Uma única sessão de qualidade oferece estímulo sem concentrar carga excessiva."
	}
	return "Sessão de base para acumular consistência com baixo custo de recuperação."
}

func experienceLabel(value string) string {
	return map[string]string{"beginner": "iniciante", "intermediate": "intermediária", "advanced": "avançada"}[value]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
