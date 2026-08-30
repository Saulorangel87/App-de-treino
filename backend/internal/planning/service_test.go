package planning

import (
	"context"
	"testing"
	"time"
)

type planStore struct {
	input       Context
	saved       Plan
	activatedID string
	activateErr error
}

func (s *planStore) PlanningContextByUserID(context.Context, string) (Context, error) {
	return s.input, nil
}
func (s *planStore) SaveDraftPlan(_ context.Context, _ string, plan Plan) (Plan, error) {
	s.saved = plan
	return plan, nil
}
func (s *planStore) CurrentPlanByUserID(context.Context, string) (Plan, error) { return s.saved, nil }
func (s *planStore) ActivatePlanByUserID(_ context.Context, _ string, planID string) error {
	s.activatedID = planID
	if s.activateErr != nil {
		return s.activateErr
	}
	s.saved.Status = "active"
	return nil
}

func TestGenerateBuildsFourWeeksAndRespectsAvailability(t *testing.T) {
	store := &planStore{input: Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 1, AvailableMinutes: 45}, {Weekday: 3, AvailableMinutes: 60}, {Weekday: 6, AvailableMinutes: 120}},
	}}
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC) }
	plan, err := service.Generate(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.StartsOn != "2026-08-31" || plan.EndsOn != "2026-09-27" {
		t.Fatalf("unexpected dates: %s to %s", plan.StartsOn, plan.EndsOn)
	}
	if len(plan.Workouts) != 12 {
		t.Fatalf("expected 12 workouts, got %d", len(plan.Workouts))
	}
	limits := map[int]int{1: 45, 3: 60, 6: 120}
	for _, workout := range plan.Workouts {
		date, _ := time.Parse("2006-01-02", workout.ScheduledOn)
		if workout.DurationMinutes > limits[int(date.Weekday())] {
			t.Fatalf("workout exceeded availability: %#v", workout)
		}
	}
}

func TestGenerateCapsIntensityWhenLimitationExists(t *testing.T) {
	store := &planStore{input: Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Limitations:  []LimitationContext{{Kind: "pain"}},
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 5, AvailableMinutes: 120}},
	}}
	service := NewService(store)
	service.now = func() time.Time { return time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC) }
	plan, err := service.Generate(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("unsafe restricted workout: %#v", workout)
		}
	}
}

func TestGenerateRequiresCompletedOnboarding(t *testing.T) {
	_, err := buildPlan(Context{ProfileID: "profile-1", ExperienceLevel: "beginner"}, time.Now())
	if err != ErrIncompleteOnboarding {
		t.Fatalf("expected incomplete onboarding, got %v", err)
	}
}

func TestActivateReturnsTheActivePlan(t *testing.T) {
	store := &planStore{saved: Plan{ID: "9a1eead7-6168-4d50-8c7c-451301e29d85", Status: "draft"}}
	service := NewService(store)
	plan, err := service.Activate(context.Background(), "user-1", store.saved.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.activatedID != store.saved.ID || plan.Status != "active" {
		t.Fatalf("plan was not activated: %#v", plan)
	}
}

func TestActivateRejectsInvalidPlanID(t *testing.T) {
	store := &planStore{}
	_, err := NewService(store).Activate(context.Background(), "user-1", "not-a-uuid")
	if err != ErrInvalidPlanID {
		t.Fatalf("expected invalid plan id, got %v", err)
	}
	if store.activatedID != "" {
		t.Fatal("store must not be called with an invalid id")
	}
}
