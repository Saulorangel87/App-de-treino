package planning

import (
	"encoding/json"
	"testing"
	"time"
)

var protectionPlanNow = time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)

func planWithProtection(t *testing.T, level ProtectionLevel) Plan {
	t.Helper()
	input := readinessContext()
	if level != "" {
		input.Protection = &ProtectionAssessment{Level: level, ExpiresOn: "2026-10-07", Reasons: []ReadinessReason{{Code: "test", Message: "motivo de teste"}}}
	}
	plan, err := buildPlan(input, protectionPlanNow)
	if err != nil {
		t.Fatalf("buildPlan: %v", err)
	}
	return plan
}

// roundTrip imita a gravação no banco: o explanation volta como JSON genérico.
func roundTrip(t *testing.T, explanation map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(explanation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestProtectionLevelsShapeThePlan(t *testing.T) {
	base := planWithProtection(t, "")
	none := planWithProtection(t, ProtectionNone)
	light := planWithProtection(t, ProtectionLight)
	moderate := planWithProtection(t, ProtectionModerate)
	strong := planWithProtection(t, ProtectionStrong)

	for index, workout := range base.Workouts {
		if got := none.Workouts[index]; got.Name != workout.Name || got.DurationMinutes != workout.DurationMinutes || got.TargetRPE != workout.TargetRPE {
			t.Fatalf("level none must match the unprotected plan: %+v vs %+v", got, workout)
		}
		quality := workout.TargetRPE >= 6

		l := light.Workouts[index]
		if l.Name != workout.Name || l.DurationMinutes > workout.DurationMinutes {
			t.Fatalf("light keeps the session type and never lengthens it: %+v vs %+v", l, workout)
		}
		if quality && l.TargetRPE != workout.TargetRPE-1 {
			t.Fatalf("light lowers quality RPE by 1: %+v vs %+v", l, workout)
		}

		m := moderate.Workouts[index]
		if quality && (m.Name != "Giro leve protegido" || m.TargetRPE > 3.5) {
			t.Fatalf("moderate replaces quality sessions: %+v", m)
		}
		if !quality && (m.Name != workout.Name || m.DurationMinutes > workout.DurationMinutes) {
			t.Fatalf("moderate keeps non-quality sessions, slightly shorter: %+v vs %+v", m, workout)
		}

		s := strong.Workouts[index]
		if s.Name != "Giro leve protegido" || s.TargetRPE > 3.5 || s.DurationMinutes > 45 {
			t.Fatalf("strong protects every session: %+v", s)
		}
	}
}

func TestProtectionIsRecordedInTheExplanation(t *testing.T) {
	plan := planWithProtection(t, ProtectionModerate)
	for _, workout := range plan.Workouts {
		protection, ok := workout.Explanation["protection"].(map[string]any)
		if !ok || protection["level"] != ProtectionModerate || protection["expires_on"] != "2026-10-07" {
			t.Fatalf("protection not recorded: %#v", workout.Explanation["protection"])
		}
	}
	if _, present := planWithProtection(t, "").Workouts[0].Explanation["protection"]; present {
		t.Fatal("the legacy path must not record a protection block")
	}
}

func TestReprescribeRestoresTheUnprotectedSession(t *testing.T) {
	base := planWithProtection(t, "")
	strong := planWithProtection(t, ProtectionStrong)
	input := readinessContext()
	input.Protection = &ProtectionAssessment{Level: ProtectionNone}
	for index, protected := range strong.Workouts {
		restored, ok := ReprescribeWorkout(input, protected.ScheduledOn, roundTrip(t, protected.Explanation), protectionPlanNow)
		if !ok {
			t.Fatalf("workout %d could not be rebuilt", index)
		}
		want := base.Workouts[index]
		if restored.Name != want.Name || restored.DurationMinutes != want.DurationMinutes || restored.TargetRPE != want.TargetRPE || restored.ScheduledOn != want.ScheduledOn {
			t.Fatalf("workout %d not restored: got %+v want %+v", index, restored, want)
		}
	}
}

func TestReprescribeAppliesANewProtectionLevel(t *testing.T) {
	base := planWithProtection(t, "")
	input := readinessContext()
	input.Protection = &ProtectionAssessment{Level: ProtectionStrong}
	for _, workout := range base.Workouts {
		rebuilt, ok := ReprescribeWorkout(input, workout.ScheduledOn, roundTrip(t, workout.Explanation), protectionPlanNow)
		if !ok || rebuilt.Name != "Giro leve protegido" {
			t.Fatalf("expected the protected session, got %+v (ok=%v)", rebuilt, ok)
		}
	}
}

func TestReprescribeSkipsWorkoutsItCannotRebuild(t *testing.T) {
	input := readinessContext()
	if _, ok := ReprescribeWorkout(input, "2026-09-07", map[string]any{"summary": "gerado antes do campo existir"}, protectionPlanNow); ok {
		t.Fatal("a workout without stored inputs must not be rebuilt")
	}
	explanation := roundTrip(t, planWithProtection(t, "").Workouts[0].Explanation)
	input.Availability = []AvailabilitySlot{{Weekday: 5, AvailableMinutes: 60}}
	if _, ok := ReprescribeWorkout(input, "2026-09-07", explanation, protectionPlanNow); ok {
		t.Fatal("a workout whose weekday is no longer available must not be rebuilt")
	}
	if _, ok := ReprescribeWorkout(readinessContext(), "data-invalida", explanation, protectionPlanNow); ok {
		t.Fatal("an invalid date must not be rebuilt")
	}
}

func TestProfileRestrictionsStillWinOverGraduatedProtection(t *testing.T) {
	input := readinessContext()
	input.Limitations = []LimitationContext{{Kind: "medical_condition", MedicalRestriction: true}}
	restricted, err := buildPlan(input, protectionPlanNow)
	if err != nil {
		t.Fatal(err)
	}
	input.Protection = &ProtectionAssessment{Level: ProtectionNone}
	unprotected, err := buildPlan(input, protectionPlanNow)
	if err != nil {
		t.Fatal(err)
	}
	for index, workout := range restricted.Workouts {
		if got := unprotected.Workouts[index]; got.Name != workout.Name || got.DurationMinutes != workout.DurationMinutes || got.TargetRPE != workout.TargetRPE {
			t.Fatalf("level none must never relax a profile restriction: %+v vs %+v", got, workout)
		}
		if workout.TargetRPE > 4 || workout.DurationMinutes > 45 {
			t.Fatalf("restriction lost: %+v", workout)
		}
	}
}
