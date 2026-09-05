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

func validHistoryPeriods() []TrainingHistoryPeriod {
	keys := []string{"last_7d", "days_8_14", "days_15_21", "days_22_28", "days_29_35", "days_36_42"}
	periods := make([]TrainingHistoryPeriod, 0, len(keys))
	for index, key := range keys {
		periods = append(periods, TrainingHistoryPeriod{
			PeriodIndex: index, PeriodKey: key, PeriodDays: 7,
			ExpectedSessions: 2, ScheduledCompletedSessions: 1,
			CancelledSessions: 1, PerformedSessions: 1, PerformedMinutes: 30,
			SessionsWithSessionRPELoad: 1, SessionRPELoad: 120,
			FeedbackRecords: 1, SessionsWithCompleteFeedback: 1,
			RecoveryCheckins: 1, CompleteRecoveryCheckins: 1,
		})
	}
	return periods
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
	snapshot := buildTrainingHistorySnapshot(history, now, validHistoryPeriods())

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
	if !reflect.DeepEqual(snapshot.EvidenceKeys, []string{"foster-2001", "haddad-2017", "impellizzeri-2020", "bourdon-2017"}) {
		t.Fatalf("unexpected evidence keys: %v", snapshot.EvidenceKeys)
	}
}

func TestBuildTrainingHistorySnapshotRecordsTemporalQualityAndProtectiveSignals(t *testing.T) {
	latestCompleted := time.Date(2026, 9, 4, 10, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	latestLoad := time.Date(2026, 9, 3, 9, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	latestRecovery := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	one, two, zero := 1, 2, 0
	windows := make([]TrainingHistoryWindow, 0, 3)
	for _, days := range []int{7, 28, 42} {
		window := validHistoryWindow(days)
		window.FeedbackRecords = 3
		window.SessionsWithCompleteFeedback = 3
		window.PainReportedSessions = 1
		window.HighFatigueSessions = 2
		window.AboveTargetRPESessions = 1
		window.RecoveryCheckins = 3
		window.CompleteRecoveryCheckins = 3
		window.CheckinsWithProtectiveSignal = 2
		window.RecoveryNeededCheckins = 1
		window.LatestCompletedAt = &latestCompleted
		window.DaysSinceLatestCompleted = &one
		window.LatestSessionRPELoadAt = &latestLoad
		window.DaysSinceLatestSessionRPELoad = &two
		window.LatestRecoveryRecordedOn = &latestRecovery
		window.DaysSinceLatestRecoveryCheckin = &zero
		windows = append(windows, window)
	}

	snapshot := buildTrainingHistorySnapshot(windows, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC), validHistoryPeriods())
	temporal := snapshot.TemporalQuality
	if temporal.LatestCompletedAt == nil || *temporal.LatestCompletedAt != "2026-09-04T13:00:00Z" ||
		temporal.DaysSinceLatestCompleted == nil || *temporal.DaysSinceLatestCompleted != 1 {
		t.Fatalf("unexpected latest completed session: %+v", temporal)
	}
	if temporal.LatestSessionRPELoadAt == nil || *temporal.LatestSessionRPELoadAt != "2026-09-03T12:00:00Z" ||
		temporal.LatestRecoveryRecordedOn == nil || *temporal.LatestRecoveryRecordedOn != "2026-09-05" {
		t.Fatalf("unexpected temporal quality: %+v", temporal)
	}
	if temporal.AthleteTimezoneAvailable || temporal.AppRecordingGapInterpretation != "recorded_activity_gap_only_not_confirmed_training_cessation" {
		t.Fatalf("unsafe temporal interpretation: %+v", temporal)
	}
	if !slices.Contains(snapshot.NotEvaluated, "load_tolerance") || !slices.Contains(snapshot.NotEvaluated, "detraining") || snapshot.UsedForPrescription {
		t.Fatalf("history became authoritative: %+v", snapshot)
	}
	if snapshot.Windows[0].PainReportedSessions != 1 || snapshot.Windows[0].RecoveryNeededCheckins != 1 {
		t.Fatalf("protective signals not preserved: %+v", snapshot.Windows[0])
	}
	if len(snapshot.MissingData) != 0 || len(snapshot.DataIssues) != 0 {
		t.Fatalf("unexpected quality flags: missing=%v issues=%v", snapshot.MissingData, snapshot.DataIssues)
	}
}

func TestBuildTrainingHistorySnapshotKeepsNonOverlappingPeriodsObservational(t *testing.T) {
	periods := validHistoryPeriods()
	periods[0].PerformedMinutes = 60
	periods[0].SessionRPELoad = 240
	snapshot := buildTrainingHistorySnapshot([]TrainingHistoryWindow{
		validHistoryWindow(7), validHistoryWindow(28), validHistoryWindow(42),
	}, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC), periods)
	comparison := snapshot.PeriodComparison
	if comparison.Version != periodComparisonVersion || comparison.Mode != trainingHistoryMode || comparison.UsedForPrescription {
		t.Fatalf("period comparison became authoritative: %+v", comparison)
	}
	if len(comparison.Periods) != 6 || comparison.Periods[0].PeriodKey != "last_7d" || comparison.Periods[5].PeriodKey != "days_36_42" {
		t.Fatalf("periods are not complete and ordered: %+v", comparison.Periods)
	}
	if comparison.Periods[0].CompletionRatePercent == nil || *comparison.Periods[0].CompletionRatePercent != 50 {
		t.Fatalf("unexpected period completion rate: %v", comparison.Periods[0].CompletionRatePercent)
	}
	if comparison.Periods[0].PerformedMinutes != 60 || comparison.Periods[1].PerformedMinutes != 30 {
		t.Fatalf("periods were aggregated together: %+v", comparison.Periods)
	}
	if len(comparison.MissingData) != 0 || len(comparison.DataIssues) != 0 {
		t.Fatalf("unexpected period quality flags: missing=%v issues=%v", comparison.MissingData, comparison.DataIssues)
	}
}

func TestBuildTrainingHistorySnapshotFlagsFutureRecordsAndTemporalMismatch(t *testing.T) {
	seven := validHistoryWindow(7)
	seven.FutureCompletedSessionsExcluded = 1
	seven.FutureRecoveryCheckinsExcluded = 2
	twentyEight := validHistoryWindow(28)
	fortyTwo := validHistoryWindow(42)
	snapshot := buildTrainingHistorySnapshot([]TrainingHistoryWindow{seven, twentyEight, fortyTwo}, time.Unix(0, 0))
	for _, issue := range []string{
		"inconsistent_temporal_metadata_28d", "inconsistent_temporal_metadata_42d",
		"future_completed_sessions_excluded", "future_recovery_checkins_excluded",
	} {
		if !slices.Contains(snapshot.DataIssues, issue) {
			t.Fatalf("missing issue %q: %v", issue, snapshot.DataIssues)
		}
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
	first := buildTrainingHistorySnapshot(history, now, validHistoryPeriods())
	second := buildTrainingHistorySnapshot(history, now, validHistoryPeriods())
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
