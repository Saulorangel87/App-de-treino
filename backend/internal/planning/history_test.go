package planning

import (
	"encoding/json"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"
)

func validHistoryWindow(days int) TrainingHistoryWindow {
	return TrainingHistoryWindow{
		WindowDays: days, ExpectedSessions: 4, ScheduledCompletedSessions: 3,
		CancelledSessions: 1, PerformedSessions: 3, PerformedMinutes: 120,
		SessionsWithSessionRPELoad: 3, SessionRPELoad: 600,
	}
}

func TestBuildTrainingHistorySnapshotCalculatesAdherenceAndCoverage(t *testing.T) {
	history := []TrainingHistoryWindow{
		{WindowDays: 42},
		{
			WindowDays: 7, ExpectedSessions: 4, ScheduledCompletedSessions: 3, CancelledSessions: 1,
			PerformedSessions: 3, PerformedMinutes: 90, SessionsWithSessionRPELoad: 3, SessionRPELoad: 450,
		},
		{
			WindowDays: 28, ExpectedSessions: 10, ScheduledCompletedSessions: 7, CancelledSessions: 1, MissedSessions: 2,
			PerformedSessions: 8, PerformedMinutes: 300, SessionsWithSessionRPELoad: 6,
			SessionsWithoutSessionRPELoad: 2, SessionRPELoad: 1500,
		},
	}
	now := time.Date(2026, 9, 5, 9, 30, 0, 0, time.FixedZone("BRT", -3*60*60))
	snapshot := buildTrainingHistorySnapshot(history, now)

	if snapshot.Version != trainingHistoryVersion || snapshot.Mode != trainingHistoryMode || snapshot.UsedForPrescription {
		t.Fatalf("unexpected authority: %+v", snapshot)
	}
	if snapshot.CapturedAt != "2026-09-05T12:30:00Z" || snapshot.LoadMethod != "duration_minutes_x_actual_rpe" {
		t.Fatalf("unexpected metadata: %+v", snapshot)
	}
	if got := []int{snapshot.Windows[0].WindowDays, snapshot.Windows[1].WindowDays, snapshot.Windows[2].WindowDays}; !reflect.DeepEqual(got, []int{7, 28, 42}) {
		t.Fatalf("windows not sorted: %v", got)
	}
	if snapshot.Windows[0].CompletionRatePercent == nil || *snapshot.Windows[0].CompletionRatePercent != 75 {
		t.Fatalf("7-day completion rate = %v", snapshot.Windows[0].CompletionRatePercent)
	}
	if snapshot.Windows[1].CompletionRatePercent == nil || *snapshot.Windows[1].CompletionRatePercent != 70 {
		t.Fatalf("28-day completion rate = %v", snapshot.Windows[1].CompletionRatePercent)
	}
	if snapshot.Windows[2].CompletionRatePercent != nil {
		t.Fatalf("completion rate without expected sessions should be null: %v", *snapshot.Windows[2].CompletionRatePercent)
	}
	for _, missing := range []string{"session_rpe_load_coverage_28d", "scheduled_history_42d", "performed_history_42d"} {
		if !slices.Contains(snapshot.MissingData, missing) {
			t.Fatalf("missing expected gap %q: %+v", missing, snapshot.MissingData)
		}
	}
	if len(snapshot.DataIssues) != 0 {
		t.Fatalf("unexpected data issues: %v", snapshot.DataIssues)
	}
	if !reflect.DeepEqual(snapshot.EvidenceKeys, []string{"foster-2001", "haddad-2017", "impellizzeri-2020"}) {
		t.Fatalf("unexpected evidence keys: %v", snapshot.EvidenceKeys)
	}
}

func TestBuildTrainingHistorySnapshotReportsMissingAndInconsistentWindows(t *testing.T) {
	history := []TrainingHistoryWindow{
		validHistoryWindow(7),
		validHistoryWindow(7),
		{
			WindowDays: 14, ExpectedSessions: 1, ScheduledCompletedSessions: 2,
			PerformedSessions: 1, PerformedMinutes: 30, SessionsWithSessionRPELoad: 2,
			SessionsWithoutSessionRPELoad: -1, SessionRPELoad: math.Inf(1),
		},
	}
	snapshot := buildTrainingHistorySnapshot(history, time.Unix(0, 0))
	for _, issue := range []string{"duplicate_window_7d", "unsupported_window_14d", "inconsistent_window_14d"} {
		if !slices.Contains(snapshot.DataIssues, issue) {
			t.Fatalf("missing issue %q: %v", issue, snapshot.DataIssues)
		}
	}
	for _, missing := range []string{"window_28d", "window_42d"} {
		if !slices.Contains(snapshot.MissingData, missing) {
			t.Fatalf("missing absent window %q: %v", missing, snapshot.MissingData)
		}
	}
	if snapshot.Windows[2].CompletionRatePercent != nil {
		t.Fatal("inconsistent adherence must not expose a rate outside the API contract")
	}
}

func TestBuildPlanRecordsHistoryWithoutChangingPrescription(t *testing.T) {
	input := readinessContext()
	input.TrainingHistory = []TrainingHistoryWindow{validHistoryWindow(7), validHistoryWindow(28), validHistoryWindow(42)}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	withHistory, err := buildPlan(input, now)
	if err != nil {
		t.Fatal(err)
	}
	input.TrainingHistory = nil
	withoutHistory, err := buildPlan(input, now)
	if err != nil {
		t.Fatal(err)
	}
	history, ok := withHistory.PrescriptionSnapshot["training_history"].(TrainingHistorySnapshot)
	if !ok || history.Version != trainingHistoryVersion || history.UsedForPrescription {
		t.Fatalf("history snapshot missing or authoritative: %#v", withHistory.PrescriptionSnapshot["training_history"])
	}
	delete(withHistory.PrescriptionSnapshot, "training_history")
	delete(withoutHistory.PrescriptionSnapshot, "training_history")
	if !reflect.DeepEqual(withHistory, withoutHistory) {
		t.Fatal("observational history changed the rules-v1 prescription")
	}
}

func TestBuildTrainingHistorySnapshotIsDeterministicAndDetached(t *testing.T) {
	history := []TrainingHistoryWindow{validHistoryWindow(42), validHistoryWindow(7), validHistoryWindow(28)}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	first := buildTrainingHistorySnapshot(history, now)
	second := buildTrainingHistorySnapshot(history, now)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("snapshot is not deterministic:\n%+v\n%+v", first, second)
	}
	history[0].ExpectedSessions = 99
	if first.Windows[2].ExpectedSessions != 4 {
		t.Fatal("snapshot retained mutable input")
	}
	*first.Windows[0].CompletionRatePercent = 0
	if *second.Windows[0].CompletionRatePercent != 75 {
		t.Fatal("snapshots share completion-rate pointers")
	}
}

func TestTrainingHistorySnapshotSerializesExplicitUnknownRate(t *testing.T) {
	snapshot := buildTrainingHistorySnapshot([]TrainingHistoryWindow{{WindowDays: 7}}, time.Unix(0, 0))
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	windows := decoded["windows"].([]any)
	window := windows[0].(map[string]any)
	if value, present := window["completion_rate_percent"]; !present || value != nil {
		t.Fatalf("unknown completion rate must be explicit null: %s", data)
	}
}
