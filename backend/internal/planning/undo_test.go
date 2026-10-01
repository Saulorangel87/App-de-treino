package planning

import (
	"context"
	"errors"
	"testing"
	"time"
)

const undoWorkoutID = "9a1eead7-6168-4d50-8c7c-451301e29d85"

func TestStartWorkoutPassesTheNormalizedLocalDate(t *testing.T) {
	store := &planStore{saved: Plan{Status: "active"}}
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 10, 2, 1, 30, 0, 0, time.UTC) }

	if _, err := service.StartWorkout(context.Background(), "user-1", undoWorkoutID, "2026-10-01"); err != nil {
		t.Fatal(err)
	}
	if store.startedToday != "2026-10-01" {
		t.Fatalf("the athlete's evening date must reach the store, got %q", store.startedToday)
	}
	if _, err := service.StartWorkout(context.Background(), "user-1", undoWorkoutID, "2030-01-01"); err != nil {
		t.Fatal(err)
	}
	if store.startedToday != "2026-10-02" {
		t.Fatalf("a forged date must fall back to the UTC date, got %q", store.startedToday)
	}
	if _, err := service.StartWorkout(context.Background(), "user-1", "nao-e-uuid", ""); !errors.Is(err, ErrInvalidWorkoutID) {
		t.Fatalf("invalid id returned %v", err)
	}
}

func TestUndoWorkoutDelegatesAndValidates(t *testing.T) {
	store := &planStore{saved: Plan{ID: "plan-1", Status: "active"}}
	service := NewService(store)

	plan, err := service.UndoWorkout(context.Background(), "user-1", undoWorkoutID)
	if err != nil || plan.ID != "plan-1" || store.undoneID != undoWorkoutID {
		t.Fatalf("undo: plan=%+v err=%v undone=%q", plan, err, store.undoneID)
	}

	store.undoneID = ""
	if _, err := service.UndoWorkout(context.Background(), "user-1", "nao-e-uuid"); !errors.Is(err, ErrInvalidWorkoutID) || store.undoneID != "" {
		t.Fatalf("an invalid id must be refused before the store: %v", err)
	}

	store.undoErr = ErrInvalidTransition
	if _, err := service.UndoWorkout(context.Background(), "user-1", undoWorkoutID); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("store errors must reach the caller: %v", err)
	}
}

func TestUndoWorkoutReevaluatesProtectionOnlyAfterASuccessfulUndo(t *testing.T) {
	input := readinessContext()
	store := &planStore{input: input, saved: Plan{ID: "plan-1"}}
	service := NewService(store, WithProtectionLevels(true))
	service.now = func() time.Time { return reevaluationNow }

	if _, err := service.UndoWorkout(context.Background(), "user-1", undoWorkoutID); err != nil {
		t.Fatal(err)
	}
	if store.plannedCalls == 0 {
		t.Fatal("a successful undo must re-evaluate the upcoming workouts")
	}
	calls := store.plannedCalls
	store.undoErr = ErrWorkoutMissing
	if _, err := service.UndoWorkout(context.Background(), "user-1", undoWorkoutID); err == nil {
		t.Fatal("expected the store error")
	}
	if store.plannedCalls != calls {
		t.Fatal("a failed undo must not re-evaluate anything")
	}
}
