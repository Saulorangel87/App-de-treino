package evolution

import "context"

type Week struct {
	WeekStart         string  `json:"week_start"`
	CompletedSessions int     `json:"completed_sessions"`
	CancelledSessions int     `json:"cancelled_sessions"`
	TotalMinutes      int     `json:"total_minutes"`
	AverageRPE        float64 `json:"average_rpe"`
}

type RecoveryPoint struct {
	RecordedOn   string `json:"recorded_on"`
	SleepMinutes int    `json:"sleep_minutes"`
	SleepQuality int    `json:"sleep_quality"`
	StressLevel  int    `json:"stress_level"`
	FatigueLevel int    `json:"fatigue_level"`
	Readiness    string `json:"readiness"`
}

type Summary struct {
	CompletedSessions int             `json:"completed_sessions"`
	CancelledSessions int             `json:"cancelled_sessions"`
	TotalMinutes      int             `json:"total_minutes"`
	AverageRPE        float64         `json:"average_rpe"`
	AverageFatigue    float64         `json:"average_fatigue"`
	CompletionRate    float64         `json:"completion_rate"`
	Weeks             []Week          `json:"weeks"`
	Recovery          []RecoveryPoint `json:"recovery"`
}

type Store interface {
	EvolutionSummaryByUserID(context.Context, string) (Summary, error)
}

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Summary(ctx context.Context, userID string) (Summary, error) {
	return s.store.EvolutionSummaryByUserID(ctx, userID)
}
