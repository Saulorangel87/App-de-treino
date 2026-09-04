package evolution

import "context"

type Week struct {
	WeekStart         string  `json:"week_start"`
	CompletedSessions int     `json:"completed_sessions"`
	CancelledSessions int     `json:"cancelled_sessions"`
	TotalMinutes      int     `json:"total_minutes"`
	AverageRPE        float64 `json:"average_rpe"`
	TotalDistanceKM   float64 `json:"total_distance_km"`
	TotalElevationM   int     `json:"total_elevation_m"`
	AveragePowerW     float64 `json:"average_power_watts"`
	AverageHeartRate  float64 `json:"average_heart_rate"`
}

type RecoveryPoint struct {
	RecordedOn   string `json:"recorded_on"`
	SleepMinutes int    `json:"sleep_minutes"`
	SleepQuality int    `json:"sleep_quality"`
	StressLevel  int    `json:"stress_level"`
	FatigueLevel int    `json:"fatigue_level"`
	Readiness    string `json:"readiness"`
}

// SessionComparison exposes the planned prescription alongside the values
// recorded when a completed session was closed. It is observational only:
// the evolution screen must not turn these differences into a diagnosis.
type SessionComparison struct {
	CompletedOn          string   `json:"completed_on"`
	Name                 string   `json:"name"`
	PlannedMinutes       int      `json:"planned_minutes"`
	ActualMinutes        *int     `json:"actual_minutes,omitempty"`
	DurationDeltaMinutes *int     `json:"duration_delta_minutes,omitempty"`
	PlannedRPE           float64  `json:"planned_rpe"`
	ActualRPE            *float64 `json:"actual_rpe,omitempty"`
	RPEDelta             *float64 `json:"rpe_delta,omitempty"`
	DistanceKM           *float64 `json:"distance_km,omitempty"`
	AveragePowerW        *int     `json:"average_power_watts,omitempty"`
	AverageHeartRate     *int     `json:"average_heart_rate,omitempty"`
	FatigueAfter         *int     `json:"fatigue_after,omitempty"`
	PainReported         bool     `json:"pain_reported"`
}

type Summary struct {
	CompletedSessions int                 `json:"completed_sessions"`
	CancelledSessions int                 `json:"cancelled_sessions"`
	TotalMinutes      int                 `json:"total_minutes"`
	AverageRPE        float64             `json:"average_rpe"`
	AverageFatigue    float64             `json:"average_fatigue"`
	CompletionRate    float64             `json:"completion_rate"`
	TotalDistanceKM   float64             `json:"total_distance_km"`
	TotalElevationM   int                 `json:"total_elevation_m"`
	AveragePowerW     float64             `json:"average_power_watts"`
	AverageHeartRate  float64             `json:"average_heart_rate"`
	Weeks             []Week              `json:"weeks"`
	RecentSessions    []SessionComparison `json:"recent_sessions"`
	Recovery          []RecoveryPoint     `json:"recovery"`
}

type Store interface {
	EvolutionSummaryByUserID(context.Context, string) (Summary, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Summary(ctx context.Context, userID string) (Summary, error) {
	return s.store.EvolutionSummaryByUserID(ctx, userID)
}
