package planning

import (
	"context"
	"errors"
	"testing"
	"time"
)

// reevaluationNow é uma terça-feira na primeira semana do plano gerado em 1/set
// (que começa na segunda, 7/set). A janela de reavaliação vai de hoje até o
// domingo da semana seguinte: 8 a 20/set.
var reevaluationNow = time.Date(2026, time.September, 8, 10, 0, 0, 0, time.UTC)

func workoutsWithIDs(t *testing.T, plan Plan) []Workout {
	t.Helper()
	workouts := make([]Workout, len(plan.Workouts))
	for index, workout := range plan.Workouts {
		workout.ID = "w-" + workout.ScheduledOn
		workout.Explanation = roundTrip(t, workout.Explanation) // como vem do banco
		workouts[index] = workout
	}
	return workouts
}

func reevaluationService(store *planStore, enabled bool) *Service {
	service := NewService(store, WithProtectionLevels(enabled))
	service.now = func() time.Time { return reevaluationNow }
	return service
}

func reevaluationStore(t *testing.T, stored Plan, signals []RecentSignal) *planStore {
	t.Helper()
	input := readinessContext()
	input.RecentSignals = signals
	return &planStore{input: input, planned: workoutsWithIDs(t, stored)}
}

func TestReevaluateIsANoOpWhenProtectionLevelsAreOff(t *testing.T) {
	store := reevaluationStore(t, planWithProtection(t, ""), []RecentSignal{{Date: reevaluationNow.AddDate(0, 0, -1), Source: "session", PainReported: true, Fatigue: 3}})
	changed, err := reevaluationService(store, false).Reevaluate(context.Background(), "user-1")
	if err != nil || changed != 0 || store.applyCalls != 0 {
		t.Fatalf("expected no change, got changed=%d err=%v apply=%d", changed, err, store.applyCalls)
	}
}

func TestReevaluateProtectsUpcomingWorkoutsInTheWindowOnly(t *testing.T) {
	signals := []RecentSignal{{Date: reevaluationNow.AddDate(0, 0, -1), Source: "session", PainReported: true, Fatigue: 3}}
	store := reevaluationStore(t, planWithProtection(t, ""), signals)
	changed, err := reevaluationService(store, true).Reevaluate(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	wantDates := map[string]bool{"w-2026-09-09": true, "w-2026-09-12": true, "w-2026-09-14": true, "w-2026-09-16": true, "w-2026-09-19": true}
	if changed != len(wantDates) || len(store.revisions) != len(wantDates) {
		t.Fatalf("expected %d revisions, got changed=%d revisions=%d", len(wantDates), changed, len(store.revisions))
	}
	for _, revision := range store.revisions {
		if !wantDates[revision.WorkoutID] {
			t.Fatalf("workout %s is outside the current and next week", revision.WorkoutID)
		}
		if revision.Name != "Giro leve protegido" || revision.TargetRPE > 3.5 {
			t.Fatalf("expected the protected session: %+v", revision)
		}
		protection, ok := revision.Explanation["protection"].(map[string]any)
		if !ok || protection["level"] != ProtectionStrong {
			t.Fatalf("protection not recorded in the revised explanation: %#v", revision.Explanation["protection"])
		}
		if _, kept := revision.Explanation["prescription_inputs"]; !kept {
			t.Fatal("the revised workout must keep its prescription inputs so it can be restored later")
		}
	}
}

func TestReevaluateRestoresWorkoutsWhenTheSignalExpires(t *testing.T) {
	base := planWithProtection(t, "")
	store := reevaluationStore(t, planWithProtection(t, ProtectionStrong), nil)
	changed, err := reevaluationService(store, true).Reevaluate(context.Background(), "user-1")
	if err != nil || changed == 0 {
		t.Fatalf("expected the protected workouts to be restored, got changed=%d err=%v", changed, err)
	}
	byDate := map[string]Workout{}
	for _, workout := range base.Workouts {
		byDate["w-"+workout.ScheduledOn] = workout
	}
	for _, revision := range store.revisions {
		want := byDate[revision.WorkoutID]
		if revision.Name != want.Name || revision.DurationMinutes != want.DurationMinutes || revision.TargetRPE != want.TargetRPE {
			t.Fatalf("workout %s not restored: %+v want %+v", revision.WorkoutID, revision, want)
		}
		if protection, _ := revision.Explanation["protection"].(map[string]any); protection["level"] != ProtectionNone {
			t.Fatalf("the restored workout should record level none: %#v", revision.Explanation["protection"])
		}
	}
}

func TestReevaluateLeavesUntouchableWorkoutsAlone(t *testing.T) {
	signals := []RecentSignal{{Date: reevaluationNow.AddDate(0, 0, -1), Source: "session", PainReported: true, Fatigue: 3}}
	store := reevaluationStore(t, planWithProtection(t, ""), signals)
	for index := range store.planned {
		switch store.planned[index].ID {
		case "w-2026-09-09":
			store.planned[index].Status = "adapted" // ajustado pelo check-in
		case "w-2026-09-12":
			delete(store.planned[index].Explanation, "prescription_inputs") // gerado antes do campo existir
		case "w-2026-09-14":
			store.planned[index].Explanation["adaptation"] = map[string]any{"kind": "recovery"} // gatilho do feedback
		}
	}
	if _, err := reevaluationService(store, true).Reevaluate(context.Background(), "user-1"); err != nil {
		t.Fatal(err)
	}
	for _, revision := range store.revisions {
		switch revision.WorkoutID {
		case "w-2026-09-09", "w-2026-09-12", "w-2026-09-14":
			t.Fatalf("workout %s must not be revised", revision.WorkoutID)
		}
	}
	if len(store.revisions) != 2 {
		t.Fatalf("expected the 2 remaining workouts of the window, got %d", len(store.revisions))
	}
}

func TestReevaluateDoesNothingWhenNothingChanges(t *testing.T) {
	// Plano gerado com a mesma avaliação que a reavaliação vai produzir (sem sinais).
	input := readinessContext()
	assessment := assessProtection(nil, reevaluationNow)
	input.Protection = &assessment
	stored, err := buildPlan(input, protectionPlanNow)
	if err != nil {
		t.Fatal(err)
	}
	store := reevaluationStore(t, stored, nil)
	changed, err := reevaluationService(store, true).Reevaluate(context.Background(), "user-1")
	if err != nil || changed != 0 || store.applyCalls != 0 {
		t.Fatalf("an unchanged prescription must not be written, got changed=%d err=%v apply=%d", changed, err, store.applyCalls)
	}
}

func TestReevaluationWindowEndsOnTheSundayOfNextWeek(t *testing.T) {
	for _, test := range []struct{ now, from, to string }{
		{"2026-09-07", "2026-09-07", "2026-09-20"}, // segunda
		{"2026-09-08", "2026-09-08", "2026-09-20"}, // terça
		{"2026-09-13", "2026-09-13", "2026-09-20"}, // domingo fecha a semana atual
		{"2026-09-14", "2026-09-14", "2026-09-27"}, // a segunda seguinte abre outra janela
	} {
		now, _ := time.Parse("2006-01-02", test.now)
		if from, to := reevaluationWindow(now); from != test.from || to != test.to {
			t.Fatalf("%s: got %s..%s, want %s..%s", test.now, from, to, test.from, test.to)
		}
	}
}

func TestCompleteWorkoutReevaluatesAndIgnoresReevaluationFailures(t *testing.T) {
	signals := []RecentSignal{{Date: reevaluationNow.AddDate(0, 0, -1), Source: "session", PainReported: true, Fatigue: 3}}
	store := reevaluationStore(t, planWithProtection(t, ""), signals)
	input := CompletionInput{CompletionStatus: "complete", ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3}
	if _, err := reevaluationService(store, true).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", input); err != nil {
		t.Fatalf("completion failed: %v", err)
	}
	if store.applyCalls != 1 {
		t.Fatalf("completing a workout should re-evaluate upcoming ones, apply calls=%d", store.applyCalls)
	}

	failing := reevaluationStore(t, planWithProtection(t, ""), signals)
	failing.plannedErr = errors.New("db down")
	if _, err := reevaluationService(failing, true).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", input); err != nil {
		t.Fatalf("a re-evaluation failure must not fail the completion: %v", err)
	}
	if failing.completedID == "" {
		t.Fatal("the completion itself must still have been saved")
	}
}
