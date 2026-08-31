package planning

import (
	"testing"
	"time"
)

func TestBuildPlanStartsAfterCompletedCycle(t *testing.T) {
	previousEnd := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	plan, err := buildPlan(Context{
		ProfileID:          "profile-1",
		ExperienceLevel:    "beginner",
		PrimaryGoal:        "health",
		PreviousPlanEndsOn: &previousEnd,
		Availability: []AvailabilitySlot{
			{Weekday: 2, AvailableMinutes: 45},
			{Weekday: 5, AvailableMinutes: 60},
		},
	}, time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.StartsOn != "2026-09-28" || plan.EndsOn != "2026-10-25" {
		t.Fatalf("next cycle overlaps the completed plan: %s to %s", plan.StartsOn, plan.EndsOn)
	}
}

func TestBuildPlanUsesCurrentNextMondayWhenPreviousCycleIsOld(t *testing.T) {
	previousEnd := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	plan, err := buildPlan(Context{
		ProfileID:          "profile-1",
		ExperienceLevel:    "beginner",
		PrimaryGoal:        "health",
		PreviousPlanEndsOn: &previousEnd,
		Availability:       []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 45}},
	}, time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.StartsOn != "2026-08-31" {
		t.Fatalf("expected current next Monday, got %s", plan.StartsOn)
	}
}
