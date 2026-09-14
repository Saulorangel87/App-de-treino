package planning

import (
	"testing"
	"time"
)

func strPtr(value string) *string { return &value }

func TestBuildPlanUsesPostEventRecoveryAfterRecentEvent(t *testing.T) {
	now := time.Date(2026, time.September, 14, 10, 0, 0, 0, time.Local)
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "event", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling: CyclingContext{
			EventGoal: true, EventDate: strPtr("2026-09-12"), Discipline: "road",
			WeeklyRides: 3, RecentTrainingWeeks: 8,
		},
	}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Workouts) == 0 {
		t.Fatal("expected workouts")
	}
	recoveryWorkouts := 0
	regularWorkoutsAfterWindow := 0
	for _, workout := range plan.Workouts {
		workoutDate, parseErr := time.ParseInLocation("2006-01-02", workout.ScheduledOn, time.Local)
		if parseErr != nil {
			t.Fatalf("unexpected workout date: %v", parseErr)
		}
		if postEventRecoveryAppliesToWorkout(planContextCycling(), PostEventRecoveryAssessment{Applied: true}, workoutDate) {
			recoveryWorkouts++
			if workout.Name != "Recuperação pós-prova" || workout.TargetRPE != postEventRecoveryTargetRPE || workout.DurationMinutes > postEventRecoveryMaxMinutes {
				t.Fatalf("recent event must produce a conservative recovery workout: %#v", workout)
			}
			if workout.Structure["protocol_key"] != "post_event_recovery" {
				t.Fatalf("expected post-event protocol metadata: %#v", workout.Structure)
			}
		} else if workout.ScheduledOn > "2026-09-19" {
			regularWorkoutsAfterWindow++
		}
	}
	if recoveryWorkouts == 0 || regularWorkoutsAfterWindow == 0 {
		t.Fatalf("expected recovery only inside the seven-day window: recovery=%d regular_after=%d", recoveryWorkouts, regularWorkoutsAfterWindow)
	}
	assessment, ok := plan.PrescriptionSnapshot["post_event_recovery"].(PostEventRecoveryAssessment)
	if !ok || !assessment.Applied || !assessment.UsedForPrescription {
		t.Fatalf("expected applied post-event assessment: %#v", plan.PrescriptionSnapshot["post_event_recovery"])
	}
}

func planContextCycling() CyclingContext {
	return CyclingContext{EventGoal: true, EventDate: strPtr("2026-09-12")}
}

func TestPostEventRecoveryWindowExpires(t *testing.T) {
	assessment := assessPostEventRecovery(Context{
		Cycling: CyclingContext{EventGoal: true, EventDate: strPtr("2026-09-01")},
	}, time.Date(2026, time.September, 14, 10, 0, 0, 0, time.Local))
	if assessment.Applied || assessment.Status != "not_applicable" {
		t.Fatalf("expired event must not keep reducing plans: %#v", assessment)
	}
}

func TestPostEventRecoveryDoesNotClaimAuthorityOverExistingProtection(t *testing.T) {
	assessment := assessPostEventRecovery(Context{
		Limitations: []LimitationContext{{Kind: "knee", ProfessionalClearanceRecommended: true}},
		Cycling:     CyclingContext{EventGoal: true, EventDate: strPtr("2026-09-12")},
	}, time.Date(2026, time.September, 14, 10, 0, 0, 0, time.Local))
	if assessment.Applied || assessment.UsedForPrescription || assessment.Status != "protective_signal" {
		t.Fatalf("post-event recovery must yield to an active protection: %#v", assessment)
	}
}
