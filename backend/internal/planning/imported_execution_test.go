package planning

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func intPtr(v int) *int { return &v }

func TestAssessImportedExecutionComparesPlannedWithMeasured(t *testing.T) {
	recorded := 70
	workout := Workout{DurationMinutes: 60, Session: &WorkoutSession{Status: "completed", DurationMinutes: &recorded}}
	facts := ImportedActivityFacts{
		ActivityID: "a1", Source: "fit", StartedAt: time.Date(2026, 9, 29, 7, 43, 0, 0, time.UTC),
		MovingSeconds: 3930, AverageHeartRate: intPtr(141), AverageCadenceRPM: intPtr(86),
	}
	got := AssessImportedExecution(workout, facts)
	if got.UsedForPrescription || got.Mode != "observation" {
		t.Fatalf("the comparison must be observation only: %+v", got)
	}
	if got.MeasuredMovingMinutes != 66 || got.DurationDeltaMinutes != 6 || got.DurationCompletionPercent == nil || *got.DurationCompletionPercent != 110 {
		t.Fatalf("unexpected duration comparison: %+v", got)
	}
	if got.RecordedDurationMinutes == nil || *got.RecordedVsMeasuredDelta != 4 {
		t.Fatalf("recorded vs measured duration not compared: %+v", got)
	}
	if !slices.Equal(got.MissingData, []string{"average_power_watts"}) {
		t.Fatalf("missing data should list only the absent power: %v", got.MissingData)
	}
}

func TestAssessImportedExecutionWithoutCompletedSession(t *testing.T) {
	got := AssessImportedExecution(Workout{DurationMinutes: 0}, ImportedActivityFacts{MovingSeconds: 1800})
	if got.RecordedDurationMinutes != nil || got.DurationCompletionPercent != nil {
		t.Fatalf("without a session or planned duration there is nothing to compare: %+v", got)
	}
	if !slices.Contains(got.MissingData, "completed_session") {
		t.Fatalf("the missing session must be explicit: %v", got.MissingData)
	}
}

func TestImportedExecutionIsSerializedWithoutAuthority(t *testing.T) {
	assessment := AssessImportedExecution(Workout{DurationMinutes: 45}, ImportedActivityFacts{MovingSeconds: 2700})
	encoded, err := json.Marshal(Workout{DurationMinutes: 45, ImportedExecution: &assessment})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	block, ok := decoded["imported_execution"].(map[string]any)
	if !ok || block["used_for_prescription"] != false || block["version"] != "imported-execution-v1" {
		t.Fatalf("imported_execution not serialized as observation: %#v", decoded["imported_execution"])
	}
	encoded, _ = json.Marshal(Workout{DurationMinutes: 45})
	var plain map[string]any
	if err := json.Unmarshal(encoded, &plain); err != nil || plain["imported_execution"] != nil {
		t.Fatal("workouts without a linked activity must not carry the block")
	}
}
