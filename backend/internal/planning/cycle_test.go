package planning

import (
	"testing"
	"time"
)

func TestBuildPlanStartsInCurrentWeekEvenWhenPreviousCycleEndsLater(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		ExperienceLevel: "beginner",
		PrimaryGoal:     "health",
		Availability: []AvailabilitySlot{
			{Weekday: 2, AvailableMinutes: 45},
			{Weekday: 5, AvailableMinutes: 60},
		},
	}, time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.StartsOn != "2026-08-31" || plan.EndsOn != "2026-09-27" {
		t.Fatalf("expected a current-week plan, got %s to %s", plan.StartsOn, plan.EndsOn)
	}
}

func TestBuildPlanSkipsDatesAlreadyPastInCurrentWeek(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		ExperienceLevel: "beginner",
		PrimaryGoal:     "health",
		Availability: []AvailabilitySlot{
			{Weekday: 2, AvailableMinutes: 45},
			{Weekday: 4, AvailableMinutes: 60},
			{Weekday: 6, AvailableMinutes: 90},
		},
	}, time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.StartsOn != "2026-08-31" || len(plan.Workouts) != 11 {
		t.Fatalf("unexpected current-week plan: %#v", plan)
	}
	for _, workout := range plan.Workouts {
		if workout.ScheduledOn < "2026-09-03" {
			t.Fatalf("plan contains a date already past: %#v", workout)
		}
	}
}
