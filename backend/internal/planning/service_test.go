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
	startedID   string
	completedID string
	cancelledID string
	completion  CompletionInput
	activities  []Activity
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
func (s *planStore) StartWorkoutByUserID(_ context.Context, _ string, workoutID string) error {
	s.startedID = workoutID
	return nil
}
func (s *planStore) CompleteWorkoutByUserID(_ context.Context, _ string, workoutID string, input CompletionInput) error {
	s.completedID = workoutID
	s.completion = input
	return nil
}
func (s *planStore) CancelWorkoutByUserID(_ context.Context, _ string, workoutID string) error {
	s.cancelledID = workoutID
	return nil
}
func (s *planStore) ActivitiesByUserID(context.Context, string) ([]Activity, error) {
	return s.activities, nil
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

func TestBuildPlanUsesCyclingContextForSpecificQualitySession(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{UsesPower: true, FTP: intPointer(250)},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Sweet spot por potência" {
			return
		}
	}
	t.Fatalf("expected a power-guided quality session, got %#v", plan.Workouts)
}

func TestBuildPlanUsesHillyTerrainForIntermediateQualitySession(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		Cycling:      CyclingContext{Terrain: "hilly"},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Subidas controladas" {
			return
		}
	}
	t.Fatalf("expected a hilly-terrain quality session, got %#v", plan.Workouts)
}

func TestBuildPlanIncludesActionableStepsForSpecificSessions(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{Terrain: "hilly"},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		steps, ok := workout.Structure["steps"].([]WorkoutStep)
		if !ok || len(steps) == 0 {
			t.Fatalf("expected actionable steps, got %#v", workout.Structure)
		}
		total := 0
		for index, step := range steps {
			if step.Order != index+1 || step.DurationMinutes <= 0 {
				t.Fatalf("invalid step ordering or duration: %#v", steps)
			}
			total += step.DurationMinutes
		}
		if total != workout.DurationMinutes {
			t.Fatalf("step total %d does not match workout duration %d: %#v", total, workout.DurationMinutes, steps)
		}
		if workout.Name == "Subidas controladas" {
			if steps[1].Title != "Subida controlada 1 de 4" || steps[2].Kind != "recovery" {
				t.Fatalf("expected explicit hill blocks and recovery: %#v", steps)
			}
			return
		}
	}
	t.Fatal("expected a specific hill workout")
}

func TestBuildPlanRestrictionOverridesSpecificCyclingContext(t *testing.T) {
	ftp := 250
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Limitations:  []LimitationContext{{Kind: "pain", ProfessionalClearanceRecommended: true}},
		Cycling:      CyclingContext{UsesPower: true, FTP: &ftp, Terrain: "hilly", EventGoal: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Giro leve protegido" || workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("expected protected session to override cycling context, got %#v", workout)
		}
	}
}

func TestBuildPlanUsesAssessmentForControlledIntervalsOnlyInBuildWeeks(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 180}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intervals := 0
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos controlados" {
			intervals++
			if workout.TargetRPE != 7 || workout.DurationMinutes > 75 {
				t.Fatalf("unexpected interval workout: %#v", workout)
			}
		}
	}
	if intervals != 2 {
		t.Fatalf("expected intervals only in two build weeks, got %d", intervals)
	}
}

func TestBuildPlanRotatesAdvancedQualityAcrossCycles(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true, RotationIndex: 1,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 180}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos controlados" {
			t.Fatalf("expected rotation to use a different quality session, got %#v", workout)
		}
	}
}

func intPointer(value int) *int { return &value }

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

func TestWorkoutLifecycleValidatesAndDelegates(t *testing.T) {
	const workoutID = "9a1eead7-6168-4d50-8c7c-451301e29d85"
	store := &planStore{saved: Plan{Status: "active"}}
	service := NewService(store)

	if _, err := service.StartWorkout(context.Background(), "user-1", workoutID); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	input := CompletionInput{ActualRPE: 7, Difficulty: "hard", FatigueAfter: 4, PainReported: false, Notes: "Sessão consistente."}
	if _, err := service.CompleteWorkout(context.Background(), "user-1", workoutID, input); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if _, err := service.CancelWorkout(context.Background(), "user-1", workoutID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if store.startedID != workoutID || store.completedID != workoutID || store.cancelledID != workoutID || store.completion.ActualRPE != 7 {
		t.Fatalf("unexpected delegated lifecycle: %#v", store)
	}
}

func TestCompleteWorkoutRejectsInvalidFeedback(t *testing.T) {
	store := &planStore{}
	_, err := NewService(store).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", CompletionInput{
		ActualRPE: 11, Difficulty: "extreme", FatigueAfter: 0,
	})
	if err != ErrInvalidFeedback {
		t.Fatalf("expected invalid feedback, got %v", err)
	}
	if store.completedID != "" {
		t.Fatal("store must not receive invalid feedback")
	}
}

func TestCompleteWorkoutRejectsInvalidOptionalMetrics(t *testing.T) {
	store := &planStore{}
	tooHighPower := 2001
	_, err := NewService(store).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3, AveragePowerW: &tooHighPower,
	})
	if err != ErrInvalidFeedback || store.completedID != "" {
		t.Fatalf("invalid optional metrics must be rejected before persistence: %#v, %v", store, err)
	}
}

func TestActivitiesReturnsOnlyWhatTheStoreProvidesForTheUser(t *testing.T) {
	store := &planStore{activities: []Activity{{ID: "session-1", Name: "Giro de base", Status: "completed"}}}
	activities, err := NewService(store).Activities(context.Background(), "user-1")
	if err != nil || len(activities) != 1 || activities[0].ID != "session-1" {
		t.Fatalf("unexpected activities: %#v, error: %v", activities, err)
	}
}
