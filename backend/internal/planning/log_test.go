package planning

import (
	"context"
	"errors"
	"testing"
	"time"
)

func validLog(duration int, performedOn string) LogWorkoutInput {
	return LogWorkoutInput{
		CompletionInput: CompletionInput{CompletionStatus: "complete", ActualRPE: 4.5, Difficulty: "moderate", FatigueAfter: 3},
		DurationMinutes: duration,
		PerformedOn:     performedOn,
	}
}

// logService fixes the clock at 01:30 UTC on Oct 2: still the evening of Oct 1 in São Paulo.
func logService(store *planStore) *Service {
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 10, 2, 1, 30, 0, 0, time.UTC) }
	return service
}

func TestLogWorkoutDelegatesWithTheAthletesToday(t *testing.T) {
	store := &planStore{saved: Plan{ID: "plan-1"}}
	service := logService(store)

	plan, err := service.LogWorkout(context.Background(), "user-1", undoWorkoutID, validLog(75, ""), "2026-10-01")
	if err != nil || plan.ID != "plan-1" {
		t.Fatalf("log: plan=%+v err=%v", plan, err)
	}
	if store.loggedID != undoWorkoutID || store.loggedToday != "2026-10-01" || store.logged.PerformedOn != "2026-10-01" || store.logged.DurationMinutes != 75 {
		t.Fatalf("unexpected delegation: id=%q today=%q performed=%q duration=%d", store.loggedID, store.loggedToday, store.logged.PerformedOn, store.logged.DurationMinutes)
	}
}

func TestLogWorkoutAcceptsTheLastWeekAndNothingElse(t *testing.T) {
	cases := map[string]bool{
		"2026-10-01": true,  // today for the athlete
		"2026-09-30": true,  // yesterday
		"2026-09-24": true,  // exactly 7 days back
		"2026-09-23": false, // 8 days back
		"2026-10-02": false, // tomorrow for the athlete
		"2026-10-03": false,
		"ontem":      false,
		"2026-9-30":  false,
		"2026-13-01": false,
	}
	for performedOn, valid := range cases {
		store := &planStore{saved: Plan{}}
		_, err := logService(store).LogWorkout(context.Background(), "user-1", undoWorkoutID, validLog(60, performedOn), "2026-10-01")
		if valid && err != nil {
			t.Errorf("%s must be accepted, got %v", performedOn, err)
		}
		if !valid && !errors.Is(err, ErrInvalidLog) {
			t.Errorf("%s returned %v, want ErrInvalidLog", performedOn, err)
		}
		if !valid && store.loggedID != "" {
			t.Errorf("%s reached the store", performedOn)
		}
	}
}

func TestLogWorkoutBoundsTheReportedDuration(t *testing.T) {
	for duration, valid := range map[int]bool{-5: false, 0: false, 1: true, 90: true, 720: true, 721: false, 5000: false} {
		store := &planStore{saved: Plan{}}
		_, err := logService(store).LogWorkout(context.Background(), "user-1", undoWorkoutID, validLog(duration, ""), "2026-10-01")
		if valid != (err == nil) {
			t.Errorf("duration %d: err=%v, valid=%t", duration, err, valid)
		}
		if !valid && !errors.Is(err, ErrInvalidLog) {
			t.Errorf("duration %d returned %v, want ErrInvalidLog", duration, err)
		}
	}
}

func TestLogWorkoutValidatesTheFeedbackAndTheId(t *testing.T) {
	store := &planStore{saved: Plan{}}
	service := logService(store)

	bad := validLog(60, "")
	bad.ActualRPE = 0
	if _, err := service.LogWorkout(context.Background(), "user-1", undoWorkoutID, bad, ""); !errors.Is(err, ErrInvalidFeedback) {
		t.Errorf("feedback without effort returned %v, want ErrInvalidFeedback", err)
	}
	if _, err := service.LogWorkout(context.Background(), "user-1", "nao-e-uuid", validLog(60, ""), ""); !errors.Is(err, ErrInvalidWorkoutID) {
		t.Errorf("invalid id returned %v", err)
	}
	if store.loggedID != "" {
		t.Error("invalid input reached the store")
	}
}

func TestLogWorkoutReevaluatesOnlyAfterSuccess(t *testing.T) {
	store := &planStore{input: readinessContext(), saved: Plan{ID: "plan-1"}}
	service := NewService(store, WithProtectionLevels(true))
	service.now = func() time.Time { return reevaluationNow }
	today := reevaluationNow.Format("2006-01-02")

	if _, err := service.LogWorkout(context.Background(), "user-1", undoWorkoutID, validLog(60, ""), today); err != nil {
		t.Fatal(err)
	}
	if store.plannedCalls == 0 {
		t.Fatal("a logged ride must re-evaluate the upcoming workouts")
	}
	calls := store.plannedCalls
	store.logErr = ErrWorkoutInFuture
	if _, err := service.LogWorkout(context.Background(), "user-1", undoWorkoutID, validLog(60, ""), today); !errors.Is(err, ErrWorkoutInFuture) {
		t.Fatalf("store error must reach the caller, got %v", err)
	}
	if store.plannedCalls != calls {
		t.Fatal("a refused log must not re-evaluate anything")
	}
}
