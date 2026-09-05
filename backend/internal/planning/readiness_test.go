package planning

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"
)

func readinessContext() Context {
	return Context{
		ProfileID: "profile-1", ExperienceLevel: "advanced", PrimaryGoal: "performance",
		Availability: []AvailabilitySlot{{Weekday: 1, AvailableMinutes: 90}, {Weekday: 3, AvailableMinutes: 90}, {Weekday: 6, AvailableMinutes: 120}},
		Cycling:      CyclingContext{Discipline: "road"}, BaselineEligible: true,
		Observed: ObservedTrainingSummary{
			WindowDays: 28, CompletedSessions: 2, CompletedMinutes: 90, AverageRPE: 5,
			AverageFatigue: 2, RecoveryCheckins: 2, AverageRecoveryFatigue: 2,
			DataCoverage: &ObservedDataCoverage{SessionsWithDuration: 2, SessionsWithRPE: 2, SessionsWithFeedback: 2, CompleteSessions: 2, RecoveryWithFatigue: 2},
		},
	}
}

func TestReadinessClassification(t *testing.T) {
	tests := []struct {
		name                    string
		change                  func(*Context)
		status, reason, missing string
	}{
		{"no aggregate alerts", func(*Context) {}, "stable", "no_aggregate_alerts", ""},
		{"new advanced account", func(c *Context) {
			c.Observed = ObservedTrainingSummary{WindowDays: 28, DataCoverage: &ObservedDataCoverage{}}
		}, "insufficient_data", "insufficient_observed_data", "completed_sessions"},
		{"declared history is not observed", func(c *Context) {
			c.Observed = ObservedTrainingSummary{WindowDays: 28}
			c.Cycling.WeeklyHours = 12
			c.Cycling.RecentTrainingWeeks = 40
		}, "insufficient_data", "insufficient_observed_data", "data_coverage"},
		{"unknown coverage", func(c *Context) { c.Observed.DataCoverage = nil }, "insufficient_data", "insufficient_observed_data", "data_coverage"},
		{"missing duration in one session", func(c *Context) {
			c.Observed.DataCoverage.SessionsWithDuration = 1
			c.Observed.DataCoverage.CompleteSessions = 1
		}, "caution", "partial_observed_data", "session_duration"},
		{"missing RPE in one session", func(c *Context) {
			c.Observed.DataCoverage.SessionsWithRPE = 1
			c.Observed.DataCoverage.CompleteSessions = 1
		}, "caution", "partial_observed_data", "session_rpe"},
		{"missing feedback in one session", func(c *Context) {
			c.Observed.DataCoverage.SessionsWithFeedback = 1
			c.Observed.DataCoverage.CompleteSessions = 1
		}, "caution", "partial_observed_data", "session_feedback"},
		{"missing recovery fatigue", func(c *Context) { c.Observed.DataCoverage.RecoveryWithFatigue = 1 }, "caution", "partial_observed_data", "recovery_fatigue"},
		{"only check-ins", func(c *Context) {
			c.Observed = ObservedTrainingSummary{WindowDays: 28, RecoveryCheckins: 1, AverageRecoveryFatigue: 2, DataCoverage: &ObservedDataCoverage{RecoveryWithFatigue: 1}}
		}, "caution", "partial_observed_data", "completed_sessions"},
		{"only sessions", func(c *Context) {
			c.Observed.RecoveryCheckins = 0
			c.Observed.AverageRecoveryFatigue = 0
			c.Observed.DataCoverage.RecoveryWithFatigue = 0
		}, "caution", "partial_observed_data", "recovery_checkins"},
		{"zero minute sessions with no usable check-in", func(c *Context) {
			c.Observed.CompletedMinutes = 0
			c.Observed.DataCoverage.SessionsWithDuration = 0
			c.Observed.DataCoverage.CompleteSessions = 0
			c.Observed.RecoveryCheckins = 0
			c.Observed.AverageRecoveryFatigue = 0
			c.Observed.DataCoverage.RecoveryWithFatigue = 0
		}, "insufficient_data", "insufficient_observed_data", "session_duration"},
		{"pain overrides absent data", func(c *Context) { c.Observed = ObservedTrainingSummary{WindowDays: 28, PainReported: true} }, "recovery_needed", "reported_pain", "completed_sessions"},
		{"active limitation", func(c *Context) {
			c.Limitations = []LimitationContext{{Kind: "injury", ProfessionalClearanceRecommended: true}}
		}, "recovery_needed", "active_limitation", ""},
		{"session fatigue at existing threshold", func(c *Context) { c.Observed.AverageFatigue = 4 }, "recovery_needed", "high_session_fatigue", ""},
		{"recovery fatigue at existing threshold", func(c *Context) { c.Observed.AverageRecoveryFatigue = 4 }, "recovery_needed", "high_recovery_fatigue", ""},
		{"fatigue below threshold", func(c *Context) { c.Observed.AverageFatigue = 3.99 }, "stable", "no_aggregate_alerts", ""},
		{"RPE alone is not poor tolerance", func(c *Context) { c.Observed.AverageRPE = 9 }, "stable", "no_aggregate_alerts", ""},
		{"negative count", func(c *Context) { c.Observed.CompletedSessions = -1 }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
		{"impossible coverage", func(c *Context) { c.Observed.DataCoverage.CompleteSessions = 3 }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
		{"contradictory complete count", func(c *Context) { c.Observed.DataCoverage.CompleteSessions = 0 }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
		{"unknown window", func(c *Context) { c.Observed.WindowDays = 0 }, "insufficient_data", "insufficient_observed_data", "history_window_28d"},
		{"out of range fatigue", func(c *Context) { c.Observed.AverageFatigue = 6 }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
		{"NaN", func(c *Context) { c.Observed.AverageRPE = math.NaN() }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
		{"infinity", func(c *Context) { c.Observed.AverageFatigue = math.Inf(1) }, "insufficient_data", "insufficient_observed_data", "consistent_history"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := readinessContext()
			tt.change(&input)
			result := assessReadiness(input, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))
			if result.Status != tt.status {
				t.Fatalf("status = %q, want %q: %+v", result.Status, tt.status, result)
			}
			if !slices.ContainsFunc(result.Reasons, func(reason ReadinessReason) bool { return reason.Code == tt.reason }) {
				t.Fatalf("missing reason %q: %+v", tt.reason, result)
			}
			if tt.missing != "" && !slices.Contains(result.MissingData, tt.missing) {
				t.Fatalf("missing data %q: %+v", tt.missing, result)
			}
			if result.ProgressionEligible || result.Mode != "observation" || result.ClassifierVersion != "readiness-v1" {
				t.Fatalf("unexpected prescription authority: %+v", result)
			}
			for _, pending := range []string{"adherence", "load_tolerance", "detraining", "trends_7_28_42d", "latest_sleep_stress_fatigue"} {
				if !slices.Contains(result.NotEvaluated, pending) {
					t.Fatalf("missing limitation %q", pending)
				}
			}
		})
	}
}

func TestReadinessIndependentOfExperienceAndBaseline(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	input := readinessContext()
	want := assessReadiness(input, now)
	for _, experience := range []string{"beginner", "intermediate", "advanced"} {
		input.ExperienceLevel = experience
		for _, eligible := range []bool{false, true} {
			input.BaselineEligible = eligible
			input.Cycling.WeeklyHours = 20
			if got := assessReadiness(input, now); !reflect.DeepEqual(got, want) {
				t.Fatalf("readiness depends on declared experience or assessment: %+v", got)
			}
		}
	}
}

func TestReadinessDeterministicAndDetached(t *testing.T) {
	input := readinessContext()
	now := time.Date(2026, 9, 5, 9, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	first := assessReadiness(input, now)
	second := assessReadiness(input, now)
	if !reflect.DeepEqual(first, second) || first.AssessedAt != "2026-09-05T12:00:00Z" {
		t.Fatalf("non-deterministic snapshot: %+v", first)
	}
	input.Observed.DataCoverage.CompleteSessions = 0
	if first.DataCoverage.CompleteSessions != 2 {
		t.Fatal("snapshot retained mutable input")
	}
	first.MissingData = append(first.MissingData, "test")
	if len(second.MissingData) != 0 {
		t.Fatal("calls share mutable data")
	}
}

func TestReadinessSnapshotRoundTripAndLegacyPrescriptionUnchanged(t *testing.T) {
	for _, scenario := range []string{"stable", "pain", "fatigue", "limitation", "new_account"} {
		t.Run(scenario, func(t *testing.T) {
			input := readinessContext()
			switch scenario {
			case "pain":
				input.Observed.PainReported = true
			case "fatigue":
				input.Observed.AverageFatigue = 4
			case "limitation":
				input.Limitations = []LimitationContext{{Kind: "pain"}}
			case "new_account":
				input.Observed = ObservedTrainingSummary{WindowDays: 28, DataCoverage: &ObservedDataCoverage{}}
			}
			store := &planStore{input: input}
			service := NewService(store)
			service.now = func() time.Time { return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC) }
			plan, err := service.Generate(context.Background(), "user-1")
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(store.saved.PrescriptionSnapshot)
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Readiness ReadinessAssessment `json:"readiness_assessment"`
			}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Readiness, plan.PrescriptionSnapshot["readiness_assessment"]) {
				t.Fatalf("snapshot lost fields: %s", data)
			}
			// Only the new data coverage changes; every legacy input is identical.
			// Compare the ENTIRE plan, including steps, evidence/rules and old snapshot.
			store.input.Observed.DataCoverage = nil
			legacyInputPlan, err := service.Generate(context.Background(), "user-1")
			if err != nil {
				t.Fatal(err)
			}
			delete(plan.PrescriptionSnapshot, "readiness_assessment")
			delete(legacyInputPlan.PrescriptionSnapshot, "readiness_assessment")
			if !reflect.DeepEqual(plan, legacyInputPlan) {
				t.Fatal("coverage changed legacy prescription")
			}
			if plan.PrescriptionSnapshot["engine_version"] != "rules-v1" {
				t.Fatal("prescribing engine changed")
			}
		})
	}
}
