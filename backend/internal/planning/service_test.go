package planning

import (
	"context"
	"math"
	"slices"
	"testing"
	"time"
)

func TestValidCompletionRejectsNonFiniteMetrics(t *testing.T) {
	if validCompletion(CompletionInput{ActualRPE: math.NaN(), Difficulty: "moderate", FatigueAfter: 3}) {
		t.Fatal("NaN actual RPE must not pass completion validation")
	}
	distance := math.NaN()
	if validCompletion(CompletionInput{ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3, DistanceKM: &distance}) {
		t.Fatal("NaN distance must not pass completion validation")
	}
}

func TestWorkoutRequiresSafetyBlockOnlyForIntenseSessionWithLimitation(t *testing.T) {
	if !WorkoutRequiresSafetyBlock(5, true) {
		t.Fatal("an intense workout with an active limitation must be blocked")
	}
	if WorkoutRequiresSafetyBlock(4, true) {
		t.Fatal("a protected workout at RPE 4 must remain startable")
	}
	if WorkoutRequiresSafetyBlock(8, false) {
		t.Fatal("an intense workout without an active limitation must not be blocked by this gate")
	}
}

type planStore struct {
	input       Context
	saved       Plan
	activatedID string
	activateErr error
	startedID   string
	completedID string
	correctedID string
	cancelledID string
	missedID    string
	completion  CompletionInput
	correction  WorkoutCorrectionInput
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
func (s *planStore) CorrectWorkoutDataByUserID(_ context.Context, _ string, workoutID string, input WorkoutCorrectionInput) error {
	s.correctedID = workoutID
	s.correction = input
	return nil
}
func (s *planStore) CancelWorkoutByUserID(_ context.Context, _ string, workoutID string) error {
	s.cancelledID = workoutID
	return nil
}
func (s *planStore) MarkWorkoutMissedByUserID(_ context.Context, _ string, workoutID string) error {
	s.missedID = workoutID
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

func TestGenerateRecordsMedicalRestrictionAndKeepsPrescriptionProtected(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		ExperienceLevel: "advanced",
		PrimaryGoal:     "performance",
		Limitations:     []LimitationContext{{Kind: "medical_condition", MedicalRestriction: true}},
		Availability:    []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}},
	}, time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	safety, ok := plan.PrescriptionSnapshot["safety_context"].(map[string]any)
	if !ok || safety["medical_restriction"] != true || safety["prescription_protected"] != true {
		t.Fatalf("medical restriction was not preserved in safety context: %#v", plan.PrescriptionSnapshot["safety_context"])
	}
	for _, workout := range plan.Workouts {
		if workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("medical restriction did not protect workout: %#v", workout)
		}
		rules, ok := workout.Explanation["rules"].([]string)
		if !ok || !slices.Contains(rules, "Restrição médica informada: a carga permanece protegida e não substitui orientação profissional.") {
			t.Fatalf("medical restriction rule was not explained: %#v", workout.Explanation["rules"])
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
			cyclingContext, ok := plan.PrescriptionSnapshot["cycling_context"].(map[string]any)
			if !ok || cyclingContext["discipline"] != "" {
				t.Fatalf("expected empty discipline to be preserved in snapshot, got %#v", plan.PrescriptionSnapshot)
			}
			return
		}
	}
	t.Fatalf("expected a power-guided quality session, got %#v", plan.Workouts)
}

func TestBuildPlanUsesRoadModerateIntervalsForEligibleRoadContext(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos moderados de estrada" {
			if workout.Structure["protocol_key"] != "road_moderate_intervals" || workout.Explanation["protocol_key"] != "road_moderate_intervals" {
				t.Fatalf("expected road protocol metadata, got %#v", workout)
			}
			if workout.Explanation["evidence_keys"].([]string)[0] != "road-mit-block-2025" {
				t.Fatalf("expected road evidence mapping, got %#v", workout.Explanation)
			}
			steps := workout.Structure["steps"].([]WorkoutStep)
			if steps[1].Title != "Intervalo moderado 1 de 3" || steps[1].DurationMinutes != 10 || steps[2].Kind != "recovery" {
				t.Fatalf("expected conservative road interval structure, got %#v", steps)
			}
			return
		}
	}
	t.Fatalf("expected eligible road context to receive the road pilot, got %#v", plan.Workouts)
}

func TestBuildPlanUsesRoadVO2IntervalsForExplicitEligiblePreference(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"vo2max"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Intervalos VO₂max de estrada" {
			continue
		}
		if workout.Structure["protocol_key"] != "road_vo2_intervals" || workout.Explanation["protocol_key"] != "road_vo2_intervals" {
			t.Fatalf("expected road VO₂max protocol metadata, got %#v", workout)
		}
		if workout.TargetRPE != 8.0 || workout.DurationMinutes > 90 {
			t.Fatalf("unexpected road VO₂max load: %#v", workout)
		}
		if workout.Explanation["evidence_keys"].([]string)[0] != "road-vo2-intervention-2024" {
			t.Fatalf("expected recent VO₂max evidence mapping, got %#v", workout.Explanation)
		}
		steps := workout.Structure["steps"].([]WorkoutStep)
		if steps[1].Title != "Intervalo VO₂max de estrada 1 de 4" || steps[1].DurationMinutes != 4 || steps[2].Kind != "recovery" || steps[2].DurationMinutes != 4 {
			t.Fatalf("expected conservative VO₂max interval structure, got %#v", steps)
		}
		return
	}
	t.Fatalf("expected explicit eligible road context to receive the VO₂max pilot, got %#v", plan.Workouts)
}

func TestBuildPlanUsesShortSelfRegulatedIntervalsForExplicitEligiblePreference(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"short_intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Intervalos curtos autorregulados" {
			continue
		}
		if workout.Structure["protocol_key"] != "short_self_regulated_intervals" || workout.Explanation["protocol_key"] != "short_self_regulated_intervals" {
			t.Fatalf("expected short interval protocol metadata, got %#v", workout)
		}
		if workout.TargetRPE != 7.5 || workout.DurationMinutes > 90 {
			t.Fatalf("unexpected short interval load: %#v", workout)
		}
		if workout.Explanation["evidence_keys"].([]string)[0] != "short-self-paced-2025" {
			t.Fatalf("expected short interval evidence mapping, got %#v", workout.Explanation)
		}
		steps := workout.Structure["steps"].([]WorkoutStep)
		if steps[1].Title != "Intervalo curto autorregulado 1 de 6" || steps[1].DurationMinutes != 1 || steps[2].Kind != "recovery" || steps[2].DurationMinutes != 1 {
			t.Fatalf("expected short self-regulated interval structure, got %#v", steps)
		}
		return
	}
	t.Fatalf("expected explicit eligible context to receive the short interval pilot, got %#v", plan.Workouts)
}

func TestBuildPlanUsesControlledThresholdForExplicitEligiblePreference(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"threshold"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Limiar controlado" {
			continue
		}
		if workout.Structure["protocol_key"] != "controlled_threshold" || workout.Explanation["protocol_key"] != "controlled_threshold" {
			t.Fatalf("expected controlled threshold protocol metadata, got %#v", workout)
		}
		if workout.TargetRPE != 7.5 || workout.DurationMinutes > 90 {
			t.Fatalf("unexpected controlled threshold load: %#v", workout)
		}
		if workout.Explanation["evidence_keys"].([]string)[0] != "road-block-comparison-2025" {
			t.Fatalf("expected threshold evidence mapping, got %#v", workout.Explanation)
		}
		steps := workout.Structure["steps"].([]WorkoutStep)
		if steps[1].Title != "Bloco de limiar 1 de 3" || steps[1].DurationMinutes != 8 || steps[2].Kind != "recovery" || steps[2].DurationMinutes != 4 {
			t.Fatalf("expected controlled threshold interval structure, got %#v", steps)
		}
		return
	}
	t.Fatalf("expected explicit eligible context to receive the controlled threshold pilot, got %#v", plan.Workouts)
}

func TestBuildPlanDoesNotUseControlledThresholdOutsideEligibleContext(t *testing.T) {
	cases := []Context{
		{
			ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"threshold"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 7, Discipline: "road", PreferredSessionTypes: []string{"threshold"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 45}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "indoor", PreferredSessionTypes: []string{"threshold"}},
		},
	}
	for index, input := range cases {
		plan, err := buildPlan(input, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatalf("case %d returned unexpected error: %v", index, err)
		}
		for _, workout := range plan.Workouts {
			if workout.Name == "Limiar controlado" || workout.Structure["protocol_key"] == "controlled_threshold" {
				t.Fatalf("case %d must not use controlled threshold protocol: %#v", index, workout)
			}
		}
	}
}

func TestBuildPlanControlledThresholdYieldsToPainProtection(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"threshold"}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 2, PainReported: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Limiar controlado" || workout.Structure["protocol_key"] == "controlled_threshold" {
			t.Fatalf("pain must override the controlled threshold protocol: %#v", workout)
		}
		if workout.Name == "Giro leve protegido" && workout.Structure["protocol_key"] != "protected_recovery" {
			t.Fatalf("expected protected fallback, got %#v", workout)
		}
	}
}

func TestBuildPlanDoesNotUseShortSelfRegulatedIntervalsOutsideEligibleContext(t *testing.T) {
	cases := []Context{
		{
			ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"short_intervals"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "mtb_xco", PreferredSessionTypes: []string{"short_intervals"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 45}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "indoor", PreferredSessionTypes: []string{"short_intervals"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 7, Discipline: "road", PreferredSessionTypes: []string{"short_intervals"}},
		},
	}
	for index, input := range cases {
		plan, err := buildPlan(input, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatalf("case %d returned unexpected error: %v", index, err)
		}
		for _, workout := range plan.Workouts {
			if workout.Name == "Intervalos curtos autorregulados" || workout.Structure["protocol_key"] == "short_self_regulated_intervals" {
				t.Fatalf("case %d must not use short self-regulated interval protocol: %#v", index, workout)
			}
		}
	}
}

func TestBuildPlanShortSelfRegulatedIntervalsYieldToPainProtection(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"short_intervals"}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 2, PainReported: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos curtos autorregulados" || workout.Structure["protocol_key"] == "short_self_regulated_intervals" {
			t.Fatalf("pain must override the short interval protocol: %#v", workout)
		}
		if workout.Name == "Giro leve protegido" && workout.Structure["protocol_key"] != "protected_recovery" {
			t.Fatalf("expected protected fallback, got %#v", workout)
		}
	}
}

func TestBuildPlanDoesNotUseRoadVO2OutsideEligibleContext(t *testing.T) {
	cases := []Context{
		{
			ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"vo2max"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 7, Discipline: "road", PreferredSessionTypes: []string{"vo2max"}},
		},
		{
			ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
			Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
			Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "mtb_xco", PreferredSessionTypes: []string{"vo2max"}},
		},
	}
	for index, input := range cases {
		plan, err := buildPlan(input, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatalf("case %d returned unexpected error: %v", index, err)
		}
		for _, workout := range plan.Workouts {
			if workout.Name == "Intervalos VO₂max de estrada" || workout.Structure["protocol_key"] == "road_vo2_intervals" {
				t.Fatalf("case %d must not use road VO₂max protocol: %#v", index, workout)
			}
		}
	}
}

func TestBuildPlanDoesNotUseRoadProtocolOutsideRoadDiscipline(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{Discipline: "mtb_xco", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos moderados de estrada" || workout.Name == "Intervalos intensos de estrada" || workout.Structure["protocol_key"] == "road_moderate_intervals" || workout.Structure["protocol_key"] == "road_high_intensity_intervals" {
			t.Fatalf("road protocol must not be selected for MTB: %#v", workout)
		}
	}
}

func TestBuildPlanRoadProtocolStillYieldsToPainProtection(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true, RotationIndex: 1,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 2, PainReported: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Giro leve protegido" || workout.Structure["protocol_key"] != "protected_recovery" {
			t.Fatalf("pain must override the road protocol: %#v", workout)
		}
	}
}

func TestBuildPlanPreservesCyclingDisciplineInSnapshot(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		Cycling:      CyclingContext{Discipline: "mtb_xco"},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cyclingContext, ok := plan.PrescriptionSnapshot["cycling_context"].(map[string]any)
	if !ok || cyclingContext["discipline"] != "mtb_xco" {
		t.Fatalf("expected discipline in plan snapshot, got %#v", plan.PrescriptionSnapshot)
	}
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

func TestBuildPlanUsesSelectedCadencePreference(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		Cycling:      CyclingContext{PreferredSessionTypes: []string{"cadence"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Cadência técnica" {
			if workout.Explanation["protocol_key"] != "technical_cadence" {
				t.Fatalf("expected cadence protocol metadata, got %#v", workout.Explanation)
			}
			return
		}
	}
	t.Fatalf("expected selected cadence preference to guide quality session, got %#v", plan.Workouts)
}

func TestBuildPlanUsesObservedRecoverySignalsConservatively(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 3, CompletedMinutes: 180, AverageRPE: 7.5, AverageFatigue: 4},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	observed, ok := plan.PrescriptionSnapshot["observed_training"].(map[string]any)
	if !ok || observed["requires_recovery"] != true {
		t.Fatalf("expected observed summary in snapshot, got %#v", plan.PrescriptionSnapshot)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Giro leve protegido" {
			if workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
				t.Fatalf("observed recovery signal must keep the quality session conservative: %#v", workout)
			}
			if workout.Explanation["summary"] == "" {
				t.Fatal("protected session should explain the observed signal")
			}
			return
		}
	}
	t.Fatal("expected the quality session to be protected by the observed signal")
}

func TestBuildPlanUsesObservedPainForAllSessions(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 75}, {Weekday: 6, AvailableMinutes: 150}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 1, PainReported: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Giro leve protegido" || workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("observed pain must protect every future session: %#v", workout)
		}
	}
}

func TestBuildPlanDoesNotBypassAssessmentForIntervalPreference(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos controlados" {
			t.Fatalf("interval preference must not bypass assessment eligibility: %#v", workout)
		}
	}
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
			if workout.Structure["protocol_key"] != "controlled_hills" || workout.Explanation["protocol_key"] != "controlled_hills" {
				t.Fatalf("expected explicit protocol mapping: %#v / %#v", workout.Structure, workout.Explanation)
			}
			if steps[1].Title != "Subida controlada 1 de 4" || steps[2].Kind != "recovery" {
				t.Fatalf("expected explicit hill blocks and recovery: %#v", steps)
			}
			return
		}
	}
	t.Fatal("expected a specific hill workout")
}

func TestSessionProtocolsKeepEvidenceMapping(t *testing.T) {
	for _, name := range []string{
		"Giro de base", "Recuperação ativa", "Retorno gradual", "Endurance contínuo", "Pedal longo", "Giro leve protegido", "Tempo controlado",
		"Ritmo de prova controlado", "Cadência técnica", "Subidas controladas",
		"Sweet spot por potência", "Sweet spot progressivo", "Limiar controlado", "Intervalos controlados", "Intervalos moderados de estrada", "Intervalos intensos de estrada", "Intervalos VO₂max de estrada", "Intervalos curtos autorregulados", "Intervalos aeróbicos XCO",
	} {
		protocol := protocolForWorkout(name)
		if protocol.Key == "" || len(protocol.EvidenceKeys) == 0 || protocol.EvidenceScope == "" {
			t.Fatalf("protocol %q is missing evidence metadata: %#v", name, protocol)
		}
		metadata := metadataForProtocol(protocol.Key)
		if metadata.PhysiologicalObjective == "" || metadata.PracticalObjective == "" || metadata.Indication == "" || metadata.Contraindication == "" || metadata.RecommendedLevel == "" || len(metadata.Prerequisites) == 0 || metadata.HeartRateGuidance == "" || metadata.PowerGuidance == "" || metadata.CadenceGuidance == "" || metadata.StopCriteria == "" || metadata.ProgressionCriteria == "" || metadata.RegressionCriteria == "" {
			t.Fatalf("protocol %q is missing operational metadata: %#v", name, metadata)
		}
	}
	if protocolForWorkout("unknown").Key != "continuous_base" {
		t.Fatal("unknown sessions should use the safe continuous fallback protocol")
	}
	if metadataForProtocol("continuous_base").PhysiologicalObjective == "" {
		t.Fatal("safe fallback protocol should keep operational metadata")
	}
}

func TestBuildPlanUsesGradualReturnForLowRecentRegularity(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 2, TrainingStatus: "returning_after_break", Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Workouts) == 0 {
		t.Fatal("expected workouts")
	}
	for _, workout := range plan.Workouts {
		if workout.Name != "Retorno gradual" || workout.TargetRPE != 3.5 || workout.DurationMinutes > 45 {
			t.Fatalf("explicit return status must produce a conservative return workout: %#v", workout)
		}
		if workout.Structure["protocol_key"] != "return_after_break" || workout.Explanation["protocol_key"] != "return_after_break" {
			t.Fatalf("expected return protocol metadata: %#v", workout)
		}
		if workout.TargetRPE >= qualityTargetRPEThreshold {
			t.Fatalf("return plan must not contain a quality session: %#v", workout)
		}
	}
}

func TestBuildPlanDoesNotTreatRegularTrainingAsReturn(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 45}, {Weekday: 6, AvailableMinutes: 120}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 2, TrainingStatus: "regular", Discipline: "road"},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Retorno gradual" {
			t.Fatalf("regular training status must not activate gradual return: %#v", workout)
		}
	}
	longSessionFound := false
	for _, workout := range plan.Workouts {
		if workout.Name == "Pedal longo" && workout.DurationMinutes > 45 {
			longSessionFound = true
		}
	}
	if !longSessionFound {
		t.Fatalf("regular training status should preserve the normal long-session progression: %#v", plan.Workouts)
	}
}

func TestBuildPlanUsesPedalLongProtocolForLongestSlot(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 1, AvailableMinutes: 45}, {Weekday: 3, AvailableMinutes: 60}, {Weekday: 6, AvailableMinutes: 120}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	longSessions := 0
	for _, workout := range plan.Workouts {
		if workout.Name != "Pedal longo" {
			continue
		}
		longSessions++
		if workout.Structure["protocol_key"] != "long_endurance" || workout.Explanation["protocol_key"] != "long_endurance" {
			t.Fatalf("expected explicit long protocol metadata, got %#v", workout)
		}
		if workout.TargetRPE != 5.0 || workout.DurationMinutes > 120 {
			t.Fatalf("unexpected long protocol load: %#v", workout)
		}
		steps, ok := workout.Structure["steps"].([]WorkoutStep)
		if !ok || len(steps) != 3 || steps[1].Kind != "main" || steps[1].Instruction != "Volume aeróbico estável e sustentável." {
			t.Fatalf("expected continuous long structure, got %#v", workout.Structure)
		}
	}
	if longSessions != 4 {
		t.Fatalf("expected one long protocol in each cycle week, got %d: %#v", longSessions, plan.Workouts)
	}
}

func TestBuildPlanUsesHighIntensityRoadIntervalsOnAlternateEligibleCycle(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true, RotationIndex: 1,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intense := 0
	for _, workout := range plan.Workouts {
		if workout.Name != "Intervalos intensos de estrada" {
			continue
		}
		intense++
		if workout.Structure["protocol_key"] != "road_high_intensity_intervals" || workout.Explanation["protocol_key"] != "road_high_intensity_intervals" {
			t.Fatalf("expected high-intensity road metadata, got %#v", workout)
		}
		if workout.TargetRPE != 8.0 || workout.DurationMinutes > 75 {
			t.Fatalf("unexpected high-intensity road load: %#v", workout)
		}
		if workout.Explanation["evidence_keys"].([]string)[0] != "road-block-comparison-2025" {
			t.Fatalf("expected recent road evidence mapping, got %#v", workout.Explanation)
		}
	}
	if intense != 2 {
		t.Fatalf("expected high-intensity pilot in both construction weeks, got %d", intense)
	}
}

func TestBuildPlanDoesNotUseHighIntensityRoadIntervalsOutsideEligibleContext(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "performance", BaselineEligible: true, RotationIndex: 1,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos intensos de estrada" || workout.Structure["protocol_key"] == "road_high_intensity_intervals" {
			t.Fatalf("high-intensity road pilot must require advanced experience: %#v", workout)
		}
	}
}

func TestBuildPlanUsesActiveRecoveryInRecoveryWeek(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 1, AvailableMinutes: 75}, {Weekday: 3, AvailableMinutes: 100}, {Weekday: 6, AvailableMinutes: 150}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	activeRecovery := 0
	recoveryWeekStart := nextMonday(time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local)).AddDate(0, 0, 21)
	for _, workout := range plan.Workouts {
		date, parseErr := time.ParseInLocation("2006-01-02", workout.ScheduledOn, time.Local)
		if parseErr != nil {
			t.Fatalf("invalid scheduled date: %v", parseErr)
		}
		if !date.Before(recoveryWeekStart) {
			if workout.Name == "Recuperação ativa" {
				activeRecovery++
				if workout.TargetRPE != 3.5 || workout.Structure["protocol_key"] != "active_recovery" {
					t.Fatalf("unexpected active recovery workout: %#v", workout)
				}
			} else if workout.Name != "Pedal longo" {
				t.Fatalf("recovery week must not contain a quality workout: %#v", workout)
			}
		} else if workout.Name == "Recuperação ativa" {
			t.Fatalf("active recovery must be limited to the recovery week: %#v", workout)
		}
	}
	if activeRecovery != 2 {
		t.Fatalf("expected all non-long sessions in recovery week to use active recovery, got %d: %#v", activeRecovery, plan.Workouts)
	}
}

func TestBuildPlanDoesNotUseEventPaceInRecoveryWeek(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "event", BaselineEligible: true, RotationIndex: 1,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{WeeklyRides: 3, RecentTrainingWeeks: 8, Discipline: "road", EventGoal: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	recoveryWeekStart := nextMonday(time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local)).AddDate(0, 0, 21)
	for _, workout := range plan.Workouts {
		date, parseErr := time.ParseInLocation("2006-01-02", workout.ScheduledOn, time.Local)
		if parseErr != nil {
			t.Fatalf("invalid scheduled date: %v", parseErr)
		}
		if date.Before(recoveryWeekStart) {
			continue
		}
		if workout.Name == "Ritmo de prova controlado" || workout.TargetRPE > 5 {
			t.Fatalf("event pace must not be selected in recovery week: %#v", workout)
		}
	}
}

func TestBuildPlanUsesEventDateForSpecificPhase(t *testing.T) {
	nearEvent := "2026-09-12"
	farEvent := "2026-12-12"
	baseContext := Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "event", BaselineEligible: true, RotationIndex: 0,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
	}
	near := baseContext
	near.Cycling = CyclingContext{EventGoal: true, EventDate: &nearEvent}
	nearPlan, err := buildPlan(near, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("near event plan failed: %v", err)
	}
	far := baseContext
	far.Cycling = CyclingContext{EventGoal: true, EventDate: &farEvent}
	farPlan, err := buildPlan(far, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("far event plan failed: %v", err)
	}

	nearEventPace := false
	farEventPace := false
	for _, workout := range nearPlan.Workouts {
		nearEventPace = nearEventPace || workout.Name == "Ritmo de prova controlado"
	}
	for _, workout := range farPlan.Workouts {
		farEventPace = farEventPace || workout.Name == "Ritmo de prova controlado"
	}
	if !nearEventPace || farEventPace {
		t.Fatalf("event date did not control the specific phase: near=%v far=%v", nearEventPace, farEventPace)
	}
}

func TestBuildPlanProtectionOverridesActiveRecovery(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "intermediate", PrimaryGoal: "endurance",
		Availability: []AvailabilitySlot{{Weekday: 1, AvailableMinutes: 75}, {Weekday: 3, AvailableMinutes: 100}, {Weekday: 6, AvailableMinutes: 150}},
		Observed:     ObservedTrainingSummary{WindowDays: 28, CompletedSessions: 2, PainReported: true},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Recuperação ativa" || workout.Name != "Giro leve protegido" || workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("protection must override active recovery: %#v", workout)
		}
		for _, rule := range workout.Explanation["rules"].([]string) {
			if rule == "Variação de recuperação ativa aplicada na quarta semana, sem aumentar a carga planejada." {
				t.Fatalf("protected workout must not retain inactive variation explanation: %#v", workout.Explanation)
			}
		}
	}
}

func TestBuildPlanUsesXCOAerobicIntervalsForEligibleXCOContext(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{Discipline: "mtb_xco", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos aeróbicos XCO" {
			if workout.Structure["protocol_key"] != "xco_aerobic_intervals" || workout.Explanation["protocol_key"] != "xco_aerobic_intervals" {
				t.Fatalf("expected XCO protocol metadata, got %#v", workout)
			}
			steps := workout.Structure["steps"].([]WorkoutStep)
			if steps[1].Title != "Bloco aeróbico XCO 1 de 5" || steps[1].DurationMinutes != 4 || steps[2].Kind != "recovery" {
				t.Fatalf("expected conservative XCO interval structure, got %#v", steps)
			}
			return
		}
	}
	t.Fatalf("expected eligible XCO context to receive the XCO pilot, got %#v", plan.Workouts)
}

func TestBuildPlanDoesNotUseXCOProtocolOutsideXCODiscipline(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance", BaselineEligible: true,
		Availability: []AvailabilitySlot{{Weekday: 2, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 180}},
		Cycling:      CyclingContext{Discipline: "road", PreferredSessionTypes: []string{"intervals"}},
	}, time.Date(2026, time.September, 1, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, workout := range plan.Workouts {
		if workout.Name == "Intervalos aeróbicos XCO" || workout.Structure["protocol_key"] == "xco_aerobic_intervals" {
			t.Fatalf("XCO protocol must not be selected outside XCO: %#v", workout)
		}
	}
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

func TestCorrectWorkoutValidatesAndDelegates(t *testing.T) {
	const workoutID = "9a1eead7-6168-4d50-8c7c-451301e29d85"
	distance := 32.5
	store := &planStore{saved: Plan{Status: "active"}}
	service := NewService(store)

	if _, err := service.CorrectWorkout(context.Background(), "user-1", workoutID, WorkoutCorrectionInput{DistanceKM: &distance}); err != nil {
		t.Fatalf("correction failed: %v", err)
	}
	if store.correctedID != workoutID || store.correction.DistanceKM == nil || *store.correction.DistanceKM != distance {
		t.Fatalf("unexpected correction delegation: %#v", store)
	}
}

func TestCorrectWorkoutRejectsInvalidOptionalMetric(t *testing.T) {
	store := &planStore{}
	invalidDistance := 2001.0
	_, err := NewService(store).CorrectWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", WorkoutCorrectionInput{DistanceKM: &invalidDistance})
	if err != ErrInvalidCorrection || store.correctedID != "" {
		t.Fatalf("invalid correction must be rejected: %#v, %v", store, err)
	}
}

func TestMarkWorkoutMissedValidatesAndDelegates(t *testing.T) {
	const workoutID = "9a1eead7-6168-4d50-8c7c-451301e29d85"
	store := &planStore{saved: Plan{Status: "active"}}
	service := NewService(store)

	if _, err := service.MarkWorkoutMissed(context.Background(), "user-1", workoutID); err != nil {
		t.Fatalf("mark missed failed: %v", err)
	}
	if store.missedID != workoutID {
		t.Fatalf("unexpected delegated missed workout: %#v", store)
	}
}

func TestMarkWorkoutMissedRejectsInvalidWorkoutID(t *testing.T) {
	store := &planStore{}
	_, err := NewService(store).MarkWorkoutMissed(context.Background(), "user-1", "not-a-uuid")
	if err != ErrInvalidWorkoutID {
		t.Fatalf("expected invalid workout id, got %v", err)
	}
	if store.missedID != "" {
		t.Fatal("store must not receive an invalid id")
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

func TestCompleteWorkoutRejectsInvalidPostWorkoutContext(t *testing.T) {
	store := &planStore{}
	invalidRecovery := 0
	_, err := NewService(store).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3, RecoveryAfter: &invalidRecovery,
	})
	if err != ErrInvalidFeedback || store.completedID != "" {
		t.Fatalf("invalid post-workout context must be rejected: %#v, %v", store, err)
	}
}

func TestCompleteWorkoutRejectsInvalidStructuredFeedbackContext(t *testing.T) {
	store := &planStore{}
	invalidSatisfaction := 0
	_, err := NewService(store).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", CompletionInput{
		ActualRPE: 5, Difficulty: "moderate", FatigueAfter: 3,
		Satisfaction: &invalidSatisfaction, Terrain: "downhill", ExternalConditions: "storm",
	})
	if err != ErrInvalidFeedback || store.completedID != "" {
		t.Fatalf("invalid structured context must be rejected before persistence: %#v, %v", store, err)
	}
}

func TestCompleteWorkoutRejectsPartialCompletionWithoutReason(t *testing.T) {
	store := &planStore{}
	_, err := NewService(store).CompleteWorkout(context.Background(), "user-1", "9a1eead7-6168-4d50-8c7c-451301e29d85", CompletionInput{
		CompletionStatus: "partial",
		ActualRPE:        5,
		Difficulty:       "moderate",
		FatigueAfter:     3,
	})

	if err != ErrInvalidFeedback || store.completedID != "" {
		t.Fatalf("partial completion without reason must be rejected: %#v, %v", store, err)
	}
}

func TestActivitiesReturnsOnlyWhatTheStoreProvidesForTheUser(t *testing.T) {
	store := &planStore{activities: []Activity{{ID: "session-1", Name: "Giro de base", Status: "completed"}}}
	activities, err := NewService(store).Activities(context.Background(), "user-1")
	if err != nil || len(activities) != 1 || activities[0].ID != "session-1" {
		t.Fatalf("unexpected activities: %#v, error: %v", activities, err)
	}
}
