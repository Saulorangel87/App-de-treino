package planning

import (
	"testing"
	"time"
)

func eventTaperTestContext(eventDate string) Context {
	return Context{
		ProfileID:        "profile-1",
		ExperienceLevel:  "advanced",
		PrimaryGoal:      "event",
		BaselineEligible: true,
		Availability: []AvailabilitySlot{
			{Weekday: 2, AvailableMinutes: 90},
			{Weekday: 6, AvailableMinutes: 180},
		},
		Cycling: CyclingContext{
			WeeklyRides:         3,
			RecentTrainingWeeks: 8,
			Discipline:          "road",
			EventGoal:           true,
			EventDate:           &eventDate,
		},
	}
}

func TestBuildPlanAppliesEventTaperOnlyInsideEligibleWindow(t *testing.T) {
	now := time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local)
	nearEvent := eventTaperTestContext("2026-09-19")
	plan, err := buildPlan(nearEvent, now)
	if err != nil {
		t.Fatalf("near event plan failed: %v", err)
	}

	assessment, ok := plan.PrescriptionSnapshot["event_taper"].(EventTaperAssessment)
	if !ok || !assessment.Applied || !assessment.UsedForPrescription || assessment.VolumeMultiplier != 0.5 {
		t.Fatalf("expected applied event taper assessment, got %#v", plan.PrescriptionSnapshot["event_taper"])
	}
	if assessment.DaysFromToday == nil || *assessment.DaysFromToday != 12 {
		t.Fatalf("expected 12 days to event, got %#v", assessment.DaysFromToday)
	}

	withoutEvent := nearEvent
	withoutEvent.Cycling.EventGoal = false
	withoutEvent.Cycling.EventDate = nil
	basePlan, err := buildPlan(withoutEvent, now)
	if err != nil {
		t.Fatalf("baseline plan failed: %v", err)
	}
	baseByDate := make(map[string]int, len(basePlan.Workouts))
	for _, workout := range basePlan.Workouts {
		baseByDate[workout.ScheduledOn] = workout.DurationMinutes
	}

	applied := 0
	for _, workout := range plan.Workouts {
		date, parseErr := time.ParseInLocation("2006-01-02", workout.ScheduledOn, time.Local)
		if parseErr != nil {
			t.Fatalf("invalid workout date: %v", parseErr)
		}
		tapered, _ := workout.Explanation["event_taper_applied"].(bool)
		if !date.Before(now) && date.Before(time.Date(2026, time.September, 19, 0, 0, 0, 0, time.Local)) && date.Before(nextMonday(now).AddDate(0, 0, 21)) {
			if !tapered {
				t.Fatalf("expected taper on eligible pre-event workout: %#v", workout)
			}
			expected := int(float64(baseByDate[workout.ScheduledOn]) * eventTaperVolumeMultiplier)
			if expected < 20 {
				expected = 20
			}
			if workout.DurationMinutes != expected {
				t.Fatalf("expected tapered duration %d, got %d for %s", expected, workout.DurationMinutes, workout.ScheduledOn)
			}
			keys := workout.Explanation["evidence_keys"].([]string)
			if len(keys) < 2 || keys[0] != "taper-cyclist-2025" || keys[1] != "taper-meta-2023" {
				t.Fatalf("expected taper evidence on affected workout, got %#v", keys)
			}
			applied++
		} else if tapered {
			t.Fatalf("taper must not apply outside the pre-event window: %#v", workout)
		}
	}
	if applied == 0 {
		t.Fatal("expected at least one tapered workout")
	}
}

func TestEventTaperRequiresTrainingBaseAndAdvancedLevel(t *testing.T) {
	eventDate := "2026-09-19"
	for _, test := range []struct {
		name   string
		mutate func(*Context)
	}{
		{name: "intermediate", mutate: func(input *Context) { input.ExperienceLevel = "intermediate" }},
		{name: "short training history", mutate: func(input *Context) { input.Cycling.RecentTrainingWeeks = 7 }},
		{name: "few weekly rides", mutate: func(input *Context) { input.Cycling.WeeklyRides = 2 }},
		{name: "baseline not eligible", mutate: func(input *Context) { input.BaselineEligible = false }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := eventTaperTestContext(eventDate)
			test.mutate(&input)
			plan, err := buildPlan(input, time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local))
			if err != nil {
				t.Fatalf("plan failed: %v", err)
			}
			assessment := plan.PrescriptionSnapshot["event_taper"].(EventTaperAssessment)
			if assessment.Applied || assessment.UsedForPrescription {
				t.Fatalf("ineligible context received taper: %#v", assessment)
			}
			for _, workout := range plan.Workouts {
				if applied, _ := workout.Explanation["event_taper_applied"].(bool); applied {
					t.Fatalf("ineligible workout received taper: %#v", workout)
				}
			}
		})
	}
}

func TestEventTaperYieldsToProtectionAndRecoveryWeek(t *testing.T) {
	input := eventTaperTestContext("2026-09-19")
	input.Observed = ObservedTrainingSummary{WindowDays: 28, PainReported: true}
	plan, err := buildPlan(input, time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("protected plan failed: %v", err)
	}
	assessment := plan.PrescriptionSnapshot["event_taper"].(EventTaperAssessment)
	if assessment.Applied || assessment.Status != "not_applicable" {
		t.Fatalf("protected context must not apply taper: %#v", assessment)
	}
	for _, workout := range plan.Workouts {
		if applied, _ := workout.Explanation["event_taper_applied"].(bool); applied {
			t.Fatalf("protected workout received taper: %#v", workout)
		}
	}

	input.Observed = ObservedTrainingSummary{}
	plan, err = buildPlan(input, time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("recovery-week plan failed: %v", err)
	}
	recoveryStart := nextMonday(time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local)).AddDate(0, 0, 21)
	for _, workout := range plan.Workouts {
		date, parseErr := time.ParseInLocation("2006-01-02", workout.ScheduledOn, time.Local)
		if parseErr != nil {
			t.Fatalf("invalid workout date: %v", parseErr)
		}
		if !date.Before(recoveryStart) {
			if applied, _ := workout.Explanation["event_taper_applied"].(bool); applied {
				t.Fatalf("recovery week must remain governed by recovery rules: %#v", workout)
			}
		}
	}
}

func TestEventTaperDoesNotApplyOutsideSevenToTwentyOneDayWindow(t *testing.T) {
	for _, eventDate := range []string{"2026-09-13", "2026-10-05"} {
		input := eventTaperTestContext(eventDate)
		plan, err := buildPlan(input, time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatalf("plan for %s failed: %v", eventDate, err)
		}
		assessment := plan.PrescriptionSnapshot["event_taper"].(EventTaperAssessment)
		if assessment.Applied || assessment.Status != "not_applicable" {
			t.Fatalf("event %s must be outside taper window: %#v", eventDate, assessment)
		}
		for _, workout := range plan.Workouts {
			if applied, _ := workout.Explanation["event_taper_applied"].(bool); applied {
				t.Fatalf("event %s received an unexpected taper: %#v", eventDate, workout)
			}
		}
	}
}

func TestEventSpecificPhaseDoesNotUsePastEventDate(t *testing.T) {
	eventDate := "2026-09-06"
	workoutDate := time.Date(2026, time.September, 7, 10, 0, 0, 0, time.Local)
	if eventSpecificPhase(CyclingContext{EventGoal: true, EventDate: &eventDate}, workoutDate) {
		t.Fatal("past event date must not activate the specific event phase")
	}
}
