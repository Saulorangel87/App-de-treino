package planning

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/i18n"
)

// The API translates workouts on the way out (internal/i18n). This test builds
// plans for many athletes and requires an English entry for every text the
// athlete reads: name, objective, main block, steps, summary, rules, evidence
// scope and protection reasons. A new phrase in the engine without a
// translation fails here instead of showing up in Portuguese to an English user.
func TestEveryWorkoutTextTheAthleteReadsHasAnEnglishEntry(t *testing.T) {
	now := time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local)
	missing := map[string]string{}
	seen := map[string]bool{}
	names := map[string]bool{}
	for name, input := range coverageContexts(now) {
		plan, err := buildPlan(input, now)
		if err != nil {
			t.Fatalf("%s: build plan: %v", name, err)
		}
		for _, workout := range plan.Workouts {
			names[workout.Name] = true
			for _, text := range visibleWorkoutTexts(workout) {
				seen[text] = true
				if text != "" && !i18n.Has(text) {
					missing[text] = name
				}
			}
		}
	}
	// Guards against the test passing because it read nothing: the matrix must
	// reach most protocols and produce steps, rules and protection reasons.
	if len(names) < 15 {
		t.Errorf("the matrix reached only %d workout names: %v", len(names), names)
	}
	for _, sample := range []string{
		"Aquecimento",
		"Progressão de três semanas seguida por uma semana de recuperação.",
		"Dor relatada nos últimos dias; proteção forte no início e reduzida nos dias seguintes.",
		"Progressão gradual e controle de carga; a referência não define minutos universais.",
	} {
		if !seen[sample] {
			t.Errorf("the matrix did not produce %q; visibleWorkoutTexts may be reading the wrong types", sample)
		}
	}
	for _, input := range []CompletionInput{
		{PainReported: true},
		{ActualRPE: 9, FatigueAfter: 5, Difficulty: "very_hard"},
		{ActualRPE: 7, FatigueAfter: 4, Difficulty: "hard"},
		{ActualRPE: 2, FatigueAfter: 1, Difficulty: "easy", CompletionStatus: "complete"},
	} {
		decision := DecideAdaptation(5, input)
		for _, text := range []string{decision.Reason, decision.SafetyNotice} {
			if text != "" && !i18n.Has(text) {
				missing[text] = "adaptation"
			}
		}
	}
	texts := make([]string, 0, len(missing))
	for text := range missing {
		texts = append(texts, text)
	}
	sort.Strings(texts)
	for _, text := range texts {
		t.Errorf("no English entry (%s): %q", missing[text], text)
	}
}

func visibleWorkoutTexts(workout Workout) []string {
	texts := []string{workout.Name, workout.Objective}
	if main, ok := workout.Structure["main"].(string); ok {
		texts = append(texts, main)
	}
	if steps, ok := workout.Structure["steps"].([]WorkoutStep); ok {
		for _, step := range steps {
			texts = append(texts, step.Title, step.Instruction)
		}
	}
	explanation := workout.Explanation
	for _, key := range []string{"summary", "evidence_scope"} {
		if text, ok := explanation[key].(string); ok {
			texts = append(texts, text)
		}
	}
	if rules, ok := explanation["rules"].([]string); ok {
		texts = append(texts, rules...)
	}
	if protection, ok := explanation["protection"].(map[string]any); ok {
		if reasons, ok := protection["reasons"].([]string); ok {
			texts = append(texts, reasons...)
		}
	}
	return texts
}

func coverageContexts(now time.Time) map[string]Context {
	text := func(value string) *string { return &value }
	number := func(value int) *int { return &value }
	eventSoon := now.AddDate(0, 0, 12).Format("2006-01-02")
	eventPast := now.AddDate(0, 0, -3).Format("2006-01-02")
	eventFar := now.AddDate(0, 0, 50).Format("2006-01-02")
	availability := []AvailabilitySlot{
		{Weekday: 1, AvailableMinutes: 60, PreferredTime: text("06:30"), Location: text("indoor")},
		{Weekday: 3, AvailableMinutes: 90},
		{Weekday: 6, AvailableMinutes: 180},
	}
	base := func(level, goal string) Context {
		return Context{
			ProfileID: "coverage", ExperienceLevel: level, PrimaryGoal: goal,
			Availability: availability,
			Cycling:      CyclingContext{WeeklyRides: 4, RecentTrainingWeeks: 12, Discipline: "road"},
		}
	}

	contexts := map[string]Context{}
	for _, level := range []string{"beginner", "intermediate", "advanced"} {
		for _, goal := range []string{"health", "fitness", "endurance", "performance", "event", "weight_management"} {
			contexts[level+"/"+goal] = base(level, goal)
		}
	}
	preferences := []string{"base", "cadence", "hills", "intervals", "threshold", "sweet_spot", "vo2max", "short_intervals", "recovery"}
	for _, discipline := range []string{"road", "mtb_xco", "mtb_xcm", "gravel", "indoor"} {
		for _, preference := range preferences {
			for rotation := 0; rotation < 2; rotation++ {
				input := base("advanced", "performance")
				input.BaselineEligible = true
				input.RotationIndex = rotation
				input.SecondaryGoal = "endurance"
				input.Cycling.Discipline = discipline
				input.Cycling.PreferredSessionTypes = []string{preference}
				input.Cycling.Terrain = "hilly"
				input.Cycling.UsesPower = preference == "sweet_spot"
				input.Cycling.FTP = number(250)
				contexts[fmt.Sprintf("%s/%s/%d", discipline, preference, rotation)] = input
			}
		}
	}

	eventSoonContext := base("advanced", "event")
	eventSoonContext.BaselineEligible = true
	eventSoonContext.Cycling.EventGoal = true
	eventSoonContext.Cycling.EventDate = &eventSoon
	contexts["event-taper"] = eventSoonContext

	eventFarContext := eventSoonContext
	eventFarContext.Cycling.EventDate = &eventFar
	contexts["event-specific"] = eventFarContext

	eventPastContext := eventSoonContext
	eventPastContext.Cycling.EventDate = &eventPast
	contexts["post-event"] = eventPastContext

	returning := base("intermediate", "fitness")
	returning.Cycling.TrainingStatus = "returning_after_break"
	contexts["returning"] = returning

	limited := base("intermediate", "fitness")
	limited.Limitations = []LimitationContext{{Kind: "injury", MedicalRestriction: true}}
	contexts["limited"] = limited

	lowActivity := base("intermediate", "performance")
	lowActivity.BaselineEligible = true
	lowActivity.Profile.ActivityLevel = text("sedentary")
	contexts["low-activity"] = lowActivity

	observed := base("advanced", "performance")
	observed.BaselineEligible = true
	observed.Observed = ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 6, CompletedMinutes: 360, AverageRPE: 5, AverageFatigue: 2}
	contexts["observed"] = observed

	protected := base("advanced", "performance")
	protected.Observed = ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 3, AverageFatigue: 4.5, PainReported: true}
	contexts["observed-pain"] = protected

	lowAdherence := base("advanced", "performance")
	lowAdherence.BaselineEligible = true
	lowAdherence.TrainingHistory = []TrainingHistoryWindow{{WindowDays: 28, MissedSessions: 2}}
	contexts["low-adherence"] = lowAdherence

	for name, signals := range map[string][]RecentSignal{
		"pain-twice":    {{Date: now.AddDate(0, 0, -1), Source: "session", PainReported: true}, {Date: now.AddDate(0, 0, -5), Source: "session", PainReported: true}},
		"pain-recent":   {{Date: now.AddDate(0, 0, -1), Source: "session", PainReported: true}},
		"pain-older":    {{Date: now.AddDate(0, 0, -5), Source: "session", PainReported: true}},
		"pain-checkin":  {{Date: now.AddDate(0, 0, -3), Source: "session", PainReported: true}, {Date: now.AddDate(0, 0, -1), Source: "checkin", Fatigue: 2}},
		"fatigue-twice": {{Date: now.AddDate(0, 0, -1), Source: "checkin", Fatigue: 5}, {Date: now.AddDate(0, 0, -3), Source: "checkin", Fatigue: 5}},
		"fatigue-once":  {{Date: now.AddDate(0, 0, -2), Source: "checkin", Fatigue: 5}},
	} {
		input := base("advanced", "performance")
		input.BaselineEligible = true
		input.RecentSignals = signals
		assessment := assessProtection(signals, now)
		input.Protection = &assessment
		contexts["protection/"+name] = input
	}
	return contexts
}
