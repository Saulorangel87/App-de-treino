package planning

import (
	"fmt"
	"math"
	"slices"
	"time"
)

const (
	trainingHistoryVersion = "training-history-v2"
	trainingHistoryMode    = "observation"
)

// TrainingHistoryWindow keeps adherence and performed load separate because
// one is anchored to the workout's scheduled date and the other to completion.
type TrainingHistoryWindow struct {
	WindowDays                    int      `json:"window_days"`
	ExpectedSessions              int      `json:"expected_sessions"`
	ScheduledCompletedSessions    int      `json:"scheduled_completed_sessions"`
	CancelledSessions             int      `json:"cancelled_sessions"`
	MissedSessions                int      `json:"missed_sessions"`
	OverdueInProgressSessions     int      `json:"overdue_in_progress_sessions"`
	CompletionRatePercent         *float64 `json:"completion_rate_percent"`
	PerformedSessions             int      `json:"performed_sessions"`
	PerformedMinutes              int      `json:"performed_minutes"`
	SessionsWithSessionRPELoad    int      `json:"sessions_with_session_rpe_load"`
	SessionsWithoutSessionRPELoad int      `json:"sessions_without_session_rpe_load"`
	SessionRPELoad                float64  `json:"session_rpe_load"`
	FeedbackRecords               int      `json:"feedback_records"`
	SessionsWithCompleteFeedback  int      `json:"sessions_with_complete_feedback"`
	PainReportedSessions          int      `json:"pain_reported_sessions"`
	HighFatigueSessions           int      `json:"high_fatigue_sessions"`
	AboveTargetRPESessions        int      `json:"above_target_rpe_sessions"`
	RecoveryCheckins              int      `json:"recovery_checkins"`
	CompleteRecoveryCheckins      int      `json:"complete_recovery_checkins"`
	CheckinsWithProtectiveSignal  int      `json:"checkins_with_protective_signal"`
	RecoveryNeededCheckins        int      `json:"recovery_needed_checkins"`

	// Temporal fields are repeated by the aggregate query and promoted to the
	// snapshot. They are not part of each public window.
	LatestCompletedAt               *time.Time `json:"-"`
	DaysSinceLatestCompleted        *int       `json:"-"`
	LatestSessionRPELoadAt          *time.Time `json:"-"`
	DaysSinceLatestSessionRPELoad   *int       `json:"-"`
	LatestRecoveryRecordedOn        *time.Time `json:"-"`
	DaysSinceLatestRecoveryCheckin  *int       `json:"-"`
	FutureCompletedSessionsExcluded int        `json:"-"`
	FutureRecoveryCheckinsExcluded  int        `json:"-"`
}

type TrainingHistoryTemporalQuality struct {
	AthleteTimezoneAvailable        bool    `json:"athlete_timezone_available"`
	LatestCompletedAt               *string `json:"latest_completed_at"`
	DaysSinceLatestCompleted        *int    `json:"days_since_latest_completed"`
	LatestSessionRPELoadAt          *string `json:"latest_session_rpe_load_at"`
	DaysSinceLatestSessionRPELoad   *int    `json:"days_since_latest_session_rpe_load"`
	LatestRecoveryRecordedOn        *string `json:"latest_recovery_recorded_on"`
	DaysSinceLatestRecoveryCheckin  *int    `json:"days_since_latest_recovery_checkin"`
	FutureCompletedSessionsExcluded int     `json:"future_completed_sessions_excluded"`
	FutureRecoveryCheckinsExcluded  int     `json:"future_recovery_checkins_excluded"`
	AppRecordingGapInterpretation   string  `json:"app_recording_gap_interpretation"`
}

// TrainingHistorySnapshot records measurements only. It is deliberately not
// consumed by rules-v1 until adherence and tolerance rules are reviewed.
type TrainingHistorySnapshot struct {
	Version             string                         `json:"version"`
	Mode                string                         `json:"mode"`
	CapturedAt          string                         `json:"captured_at"`
	LoadMethod          string                         `json:"load_method"`
	LoadUnit            string                         `json:"load_unit"`
	AdherenceBasis      string                         `json:"adherence_basis"`
	CompletionTimeBasis string                         `json:"completion_time_basis"`
	EvidenceKeys        []string                       `json:"evidence_keys"`
	Windows             []TrainingHistoryWindow        `json:"windows"`
	TemporalQuality     TrainingHistoryTemporalQuality `json:"temporal_quality"`
	MissingData         []string                       `json:"missing_data"`
	DataIssues          []string                       `json:"data_issues"`
	NotEvaluated        []string                       `json:"not_evaluated"`
	UsedForPrescription bool                           `json:"used_for_prescription"`
}

func buildTrainingHistorySnapshot(history []TrainingHistoryWindow, now time.Time) TrainingHistorySnapshot {
	result := TrainingHistorySnapshot{
		Version:             trainingHistoryVersion,
		Mode:                trainingHistoryMode,
		CapturedAt:          now.UTC().Format(time.RFC3339Nano),
		LoadMethod:          "duration_minutes_x_actual_rpe",
		LoadUnit:            "session_rpe_arbitrary_units",
		AdherenceBasis:      "closed_workouts_by_scheduled_date_database_current_date",
		CompletionTimeBasis: "rolling_completed_at_database_clock",
		EvidenceKeys:        []string{"foster-2001", "haddad-2017", "impellizzeri-2020", "bourdon-2017"},
		Windows:             make([]TrainingHistoryWindow, len(history)),
		TemporalQuality: TrainingHistoryTemporalQuality{
			AppRecordingGapInterpretation: "recorded_activity_gap_only_not_confirmed_training_cessation",
		},
		MissingData: []string{},
		DataIssues:  []string{},
		NotEvaluated: []string{
			"load_tolerance", "detraining", "fitness_change", "activities_outside_cadencia",
			"athlete_timezone", "progression_from_history",
		},
	}
	copy(result.Windows, history)
	slices.SortFunc(result.Windows, func(left, right TrainingHistoryWindow) int {
		return left.WindowDays - right.WindowDays
	})

	expectedWindows := []int{7, 28, 42}
	supportedWindows := map[int]bool{7: true, 28: true, 42: true}
	seen := make(map[int]bool, len(result.Windows))
	if len(result.Windows) > 0 {
		populateTemporalQuality(&result, result.Windows[0])
	}
	for index := range result.Windows {
		window := &result.Windows[index]
		if !supportedWindows[window.WindowDays] {
			result.DataIssues = append(result.DataIssues, fmt.Sprintf("unsupported_window_%dd", window.WindowDays))
		}
		if seen[window.WindowDays] {
			result.DataIssues = append(result.DataIssues, fmt.Sprintf("duplicate_window_%dd", window.WindowDays))
		}
		seen[window.WindowDays] = true
		if window.ExpectedSessions > 0 && window.ScheduledCompletedSessions >= 0 && window.ScheduledCompletedSessions <= window.ExpectedSessions {
			rate := float64(window.ScheduledCompletedSessions) / float64(window.ExpectedSessions) * 100
			window.CompletionRatePercent = &rate
		} else {
			window.CompletionRatePercent = nil
			if window.ExpectedSessions == 0 {
				result.MissingData = append(result.MissingData, fmt.Sprintf("scheduled_history_%dd", window.WindowDays))
			}
		}
		if window.PerformedSessions == 0 {
			result.MissingData = append(result.MissingData, fmt.Sprintf("performed_history_%dd", window.WindowDays))
		}
		if window.SessionsWithSessionRPELoad < window.PerformedSessions {
			result.MissingData = append(result.MissingData, fmt.Sprintf("session_rpe_load_coverage_%dd", window.WindowDays))
		}
		if window.FeedbackRecords < window.PerformedSessions {
			result.MissingData = append(result.MissingData, fmt.Sprintf("feedback_coverage_%dd", window.WindowDays))
		}
		if window.SessionsWithCompleteFeedback < window.FeedbackRecords {
			result.MissingData = append(result.MissingData, fmt.Sprintf("feedback_field_coverage_%dd", window.WindowDays))
		}
		if window.CompleteRecoveryCheckins < window.RecoveryCheckins {
			result.MissingData = append(result.MissingData, fmt.Sprintf("recovery_checkin_coverage_%dd", window.WindowDays))
		}
		if index > 0 && !sameTemporalMetadata(result.Windows[0], *window) {
			result.DataIssues = append(result.DataIssues, fmt.Sprintf("inconsistent_temporal_metadata_%dd", window.WindowDays))
		}
		if !validTrainingHistoryWindow(*window) {
			result.DataIssues = append(result.DataIssues, fmt.Sprintf("inconsistent_window_%dd", window.WindowDays))
		}
	}
	for _, days := range expectedWindows {
		if !seen[days] {
			result.MissingData = append(result.MissingData, fmt.Sprintf("window_%dd", days))
		}
	}
	if result.TemporalQuality.LatestCompletedAt == nil {
		result.MissingData = append(result.MissingData, "latest_completed_session")
	}
	if result.TemporalQuality.LatestSessionRPELoadAt == nil {
		result.MissingData = append(result.MissingData, "latest_session_rpe_load")
	}
	if result.TemporalQuality.LatestRecoveryRecordedOn == nil {
		result.MissingData = append(result.MissingData, "latest_recovery_checkin")
	}
	if result.TemporalQuality.FutureCompletedSessionsExcluded > 0 {
		result.DataIssues = append(result.DataIssues, "future_completed_sessions_excluded")
	}
	if result.TemporalQuality.FutureRecoveryCheckinsExcluded > 0 {
		result.DataIssues = append(result.DataIssues, "future_recovery_checkins_excluded")
	}
	return result
}

func populateTemporalQuality(snapshot *TrainingHistorySnapshot, source TrainingHistoryWindow) {
	snapshot.TemporalQuality.LatestCompletedAt = formatTimestamp(source.LatestCompletedAt)
	snapshot.TemporalQuality.DaysSinceLatestCompleted = copyInt(source.DaysSinceLatestCompleted)
	snapshot.TemporalQuality.LatestSessionRPELoadAt = formatTimestamp(source.LatestSessionRPELoadAt)
	snapshot.TemporalQuality.DaysSinceLatestSessionRPELoad = copyInt(source.DaysSinceLatestSessionRPELoad)
	snapshot.TemporalQuality.LatestRecoveryRecordedOn = formatDate(source.LatestRecoveryRecordedOn)
	snapshot.TemporalQuality.DaysSinceLatestRecoveryCheckin = copyInt(source.DaysSinceLatestRecoveryCheckin)
	snapshot.TemporalQuality.FutureCompletedSessionsExcluded = source.FutureCompletedSessionsExcluded
	snapshot.TemporalQuality.FutureRecoveryCheckinsExcluded = source.FutureRecoveryCheckinsExcluded
}

func formatTimestamp(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339Nano)
	return &formatted
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func copyInt(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func sameTemporalMetadata(left, right TrainingHistoryWindow) bool {
	return equalTime(left.LatestCompletedAt, right.LatestCompletedAt) &&
		equalInt(left.DaysSinceLatestCompleted, right.DaysSinceLatestCompleted) &&
		equalTime(left.LatestSessionRPELoadAt, right.LatestSessionRPELoadAt) &&
		equalInt(left.DaysSinceLatestSessionRPELoad, right.DaysSinceLatestSessionRPELoad) &&
		equalTime(left.LatestRecoveryRecordedOn, right.LatestRecoveryRecordedOn) &&
		equalInt(left.DaysSinceLatestRecoveryCheckin, right.DaysSinceLatestRecoveryCheckin) &&
		left.FutureCompletedSessionsExcluded == right.FutureCompletedSessionsExcluded &&
		left.FutureRecoveryCheckinsExcluded == right.FutureRecoveryCheckinsExcluded
}

func equalTime(left, right *time.Time) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && left.Equal(*right))
}

func equalInt(left, right *int) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func validTrainingHistoryWindow(window TrainingHistoryWindow) bool {
	counts := []int{
		window.WindowDays, window.ExpectedSessions, window.ScheduledCompletedSessions,
		window.CancelledSessions, window.MissedSessions, window.OverdueInProgressSessions,
		window.PerformedSessions, window.PerformedMinutes, window.SessionsWithSessionRPELoad,
		window.SessionsWithoutSessionRPELoad, window.FeedbackRecords, window.SessionsWithCompleteFeedback,
		window.PainReportedSessions,
		window.HighFatigueSessions, window.AboveTargetRPESessions, window.RecoveryCheckins,
		window.CompleteRecoveryCheckins, window.CheckinsWithProtectiveSignal, window.RecoveryNeededCheckins,
		window.FutureCompletedSessionsExcluded, window.FutureRecoveryCheckinsExcluded,
	}
	if slices.ContainsFunc(counts, func(value int) bool { return value < 0 }) || window.WindowDays == 0 ||
		math.IsNaN(window.SessionRPELoad) || math.IsInf(window.SessionRPELoad, 0) || window.SessionRPELoad < 0 {
		return false
	}
	closed := window.ScheduledCompletedSessions + window.CancelledSessions + window.MissedSessions + window.OverdueInProgressSessions
	return closed == window.ExpectedSessions &&
		window.SessionsWithSessionRPELoad+window.SessionsWithoutSessionRPELoad == window.PerformedSessions &&
		window.SessionsWithSessionRPELoad <= window.PerformedMinutes &&
		window.FeedbackRecords <= window.PerformedSessions &&
		window.SessionsWithCompleteFeedback <= window.FeedbackRecords &&
		window.PainReportedSessions <= window.FeedbackRecords &&
		window.HighFatigueSessions <= window.SessionsWithCompleteFeedback &&
		window.AboveTargetRPESessions <= window.PerformedSessions &&
		window.CompleteRecoveryCheckins <= window.RecoveryCheckins &&
		window.CheckinsWithProtectiveSignal <= window.CompleteRecoveryCheckins &&
		window.RecoveryNeededCheckins <= window.CheckinsWithProtectiveSignal &&
		validTemporalPair(window.LatestCompletedAt, window.DaysSinceLatestCompleted) &&
		validTemporalPair(window.LatestSessionRPELoadAt, window.DaysSinceLatestSessionRPELoad) &&
		validTemporalPair(window.LatestRecoveryRecordedOn, window.DaysSinceLatestRecoveryCheckin)
}

func validTemporalPair(timestamp *time.Time, days *int) bool {
	return (timestamp == nil && days == nil) || (timestamp != nil && days != nil && *days >= 0)
}
