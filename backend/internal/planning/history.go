package planning

import (
	"fmt"
	"math"
	"slices"
	"time"
)

const (
	trainingHistoryVersion = "training-history-v1"
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
}

// TrainingHistorySnapshot records measurements only. It is deliberately not
// consumed by rules-v1 until adherence and tolerance rules are reviewed.
type TrainingHistorySnapshot struct {
	Version             string                  `json:"version"`
	Mode                string                  `json:"mode"`
	CapturedAt          string                  `json:"captured_at"`
	LoadMethod          string                  `json:"load_method"`
	LoadUnit            string                  `json:"load_unit"`
	AdherenceBasis      string                  `json:"adherence_basis"`
	CompletionTimeBasis string                  `json:"completion_time_basis"`
	EvidenceKeys        []string                `json:"evidence_keys"`
	Windows             []TrainingHistoryWindow `json:"windows"`
	MissingData         []string                `json:"missing_data"`
	DataIssues          []string                `json:"data_issues"`
	UsedForPrescription bool                    `json:"used_for_prescription"`
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
		EvidenceKeys:        []string{"foster-2001", "haddad-2017", "impellizzeri-2020"},
		Windows:             make([]TrainingHistoryWindow, len(history)),
		MissingData:         []string{},
		DataIssues:          []string{},
	}
	copy(result.Windows, history)
	slices.SortFunc(result.Windows, func(left, right TrainingHistoryWindow) int {
		return left.WindowDays - right.WindowDays
	})

	expectedWindows := []int{7, 28, 42}
	supportedWindows := map[int]bool{7: true, 28: true, 42: true}
	seen := make(map[int]bool, len(result.Windows))
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
		if !validTrainingHistoryWindow(*window) {
			result.DataIssues = append(result.DataIssues, fmt.Sprintf("inconsistent_window_%dd", window.WindowDays))
		}
	}
	for _, days := range expectedWindows {
		if !seen[days] {
			result.MissingData = append(result.MissingData, fmt.Sprintf("window_%dd", days))
		}
	}
	return result
}

func validTrainingHistoryWindow(window TrainingHistoryWindow) bool {
	counts := []int{
		window.WindowDays, window.ExpectedSessions, window.ScheduledCompletedSessions,
		window.CancelledSessions, window.MissedSessions, window.OverdueInProgressSessions,
		window.PerformedSessions, window.PerformedMinutes, window.SessionsWithSessionRPELoad,
		window.SessionsWithoutSessionRPELoad,
	}
	if slices.ContainsFunc(counts, func(value int) bool { return value < 0 }) || window.WindowDays == 0 ||
		math.IsNaN(window.SessionRPELoad) || math.IsInf(window.SessionRPELoad, 0) || window.SessionRPELoad < 0 {
		return false
	}
	closed := window.ScheduledCompletedSessions + window.CancelledSessions + window.MissedSessions + window.OverdueInProgressSessions
	return closed == window.ExpectedSessions &&
		window.SessionsWithSessionRPELoad+window.SessionsWithoutSessionRPELoad == window.PerformedSessions &&
		window.SessionsWithSessionRPELoad <= window.PerformedMinutes
}
