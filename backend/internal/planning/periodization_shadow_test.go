package planning

import (
	"slices"
	"testing"
	"time"
)

func periodizationWorkout(date, name, protocolKey string, duration int, targetRPE float64) Workout {
	return Workout{
		ScheduledOn:     date,
		Name:            name,
		DurationMinutes: duration,
		TargetRPE:       targetRPE,
		Structure:       map[string]any{"protocol_key": protocolKey},
		Explanation:     map[string]any{"protocol_key": protocolKey},
		Status:          "planned",
	}
}

func coherentPeriodizationWorkouts() []Workout {
	return []Workout{
		periodizationWorkout("2026-09-14", "Giro de base", "base_endurance", 45, 4),
		periodizationWorkout("2026-09-16", "Tempo controlado", "controlled_tempo", 60, 6),
		periodizationWorkout("2026-09-19", "Endurance contínuo", "continuous_endurance", 90, 5),
		periodizationWorkout("2026-09-21", "Giro de base", "base_endurance", 50, 4),
		periodizationWorkout("2026-09-23", "Intervalos controlados", "controlled_intervals", 65, 7),
		periodizationWorkout("2026-09-26", "Endurance contínuo", "continuous_endurance", 100, 5),
		periodizationWorkout("2026-09-28", "Giro de base", "base_endurance", 55, 4),
		periodizationWorkout("2026-09-30", "Tempo controlado", "controlled_tempo", 70, 6),
		periodizationWorkout("2026-10-03", "Endurance contínuo", "continuous_endurance", 110, 5),
		periodizationWorkout("2026-10-05", "Recuperação ativa", "active_recovery", 40, 3.5),
		periodizationWorkout("2026-10-07", "Recuperação ativa", "active_recovery", 40, 3.5),
		periodizationWorkout("2026-10-10", "Endurance contínuo", "continuous_endurance", 45, 5),
	}
}

func TestPeriodizationShadowObservesCoherentFourWeekCycle(t *testing.T) {
	assessment := assessPeriodizationShadow(
		coherentPeriodizationWorkouts(),
		time.Date(2026, 9, 14, 12, 0, 0, 0, time.Local),
		time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC),
	)

	if assessment.Status != "observed" || assessment.CandidateResponse != "maintain_observed" {
		t.Fatalf("unexpected coherent periodization assessment: %+v", assessment)
	}
	if assessment.Version != periodizationShadowVersion || assessment.Mode != periodizationShadowMode || assessment.Scope != periodizationShadowScope {
		t.Fatalf("unexpected periodization shadow metadata: %+v", assessment)
	}
	if len(assessment.Weeks) != periodizationWeekCount || assessment.Weeks[0].Phase != "progression" || assessment.Weeks[3].Phase != "recovery" {
		t.Fatalf("unexpected periodization phases: %+v", assessment.Weeks)
	}
	if assessment.QualitySessions != 3 || assessment.HighIntensitySessions != 1 || assessment.AdjacentQualitySessionPairs != 0 {
		t.Fatalf("unexpected planned stimulus counts: %+v", assessment)
	}
	if assessment.MinimumDaysBetweenQualitySessions == nil || *assessment.MinimumDaysBetweenQualitySessions != 7 {
		t.Fatalf("unexpected quality spacing: %+v", assessment.MinimumDaysBetweenQualitySessions)
	}
	if assessment.Weeks[3].QualitySessions != 0 || assessment.Weeks[3].RecoverySessions != 2 || assessment.Weeks[3].LongSessions != 1 || assessment.Weeks[3].TotalPlannedMinutes >= assessment.Weeks[2].TotalPlannedMinutes {
		t.Fatalf("recovery week was not observed as lighter: %+v", assessment.Weeks[3])
	}
	if assessment.ProgressionEligible || assessment.Applied || assessment.UsedForPrescription {
		t.Fatal("periodization shadow became authoritative")
	}
	if !slices.Contains(assessment.RulesEvaluated, "prescription_isolation_gate") {
		t.Fatalf("prescription isolation was not evaluated: %+v", assessment.RulesEvaluated)
	}
}

func TestPeriodizationShadowPreservesMissingWeeks(t *testing.T) {
	assessment := assessPeriodizationShadow(
		coherentPeriodizationWorkouts()[:6],
		time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC),
	)

	if assessment.Status != "not_evaluated" || assessment.CandidateResponse != "defer_evaluation" {
		t.Fatalf("missing weeks produced an evaluated result: %+v", assessment)
	}
	for _, missing := range []string{"week_3_sessions", "week_4_sessions"} {
		if !slices.Contains(assessment.MissingData, missing) {
			t.Fatalf("missing week %q was not preserved: %+v", missing, assessment.MissingData)
		}
	}
	if assessment.UsedForPrescription {
		t.Fatal("incomplete periodization became authoritative")
	}
}

func TestPeriodizationShadowFlagsQualityInsideRecoveryWeek(t *testing.T) {
	workouts := coherentPeriodizationWorkouts()
	workouts[9] = periodizationWorkout("2026-10-05", "Tempo controlado", "controlled_tempo", 60, 6)
	assessment := assessPeriodizationShadow(
		workouts,
		time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC),
	)

	if assessment.Status != "not_evaluated" || !slices.Contains(assessment.DataIssues, "recovery_week_contains_quality") {
		t.Fatalf("recovery quality inconsistency was not flagged: %+v", assessment)
	}
	if assessment.ProgressionEligible || assessment.UsedForPrescription {
		t.Fatal("inconsistent periodization became authoritative")
	}
}

func TestBuildPlanAttachesPeriodizationShadowWithoutReplacingRulesV1(t *testing.T) {
	plan, err := buildPlan(Context{
		ProfileID:       "profile-1",
		PrimaryGoal:     "health",
		ExperienceLevel: "beginner",
		Availability: []AvailabilitySlot{
			{Weekday: 1, AvailableMinutes: 45},
			{Weekday: 3, AvailableMinutes: 60},
			{Weekday: 6, AvailableMinutes: 90},
		},
	}, time.Date(2026, 9, 14, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("buildPlan returned error: %v", err)
	}
	shadow, ok := plan.PrescriptionSnapshot["periodization_shadow"].(PeriodizationShadowAssessment)
	if !ok || shadow.Mode != "shadow" || shadow.UsedForPrescription {
		t.Fatalf("periodization shadow missing or authoritative: %#v", plan.PrescriptionSnapshot["periodization_shadow"])
	}
	if plan.PrescriptionSnapshot["engine_version"] != "rules-v1" {
		t.Fatalf("rules-v1 was replaced: %#v", plan.PrescriptionSnapshot["engine_version"])
	}
	if len(plan.Workouts) != 12 {
		t.Fatalf("periodization observation changed generated workout count: %d", len(plan.Workouts))
	}
}
