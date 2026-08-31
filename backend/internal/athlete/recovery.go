package athlete

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidRecovery = errors.New("invalid recovery data")

type AdaptedWorkout struct {
	ID              string  `json:"id"`
	ScheduledOn     string  `json:"scheduled_on"`
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	TargetRPE       float64 `json:"target_rpe"`
}

type RecoveryCheckin struct {
	ID                string          `json:"id,omitempty"`
	RecordedOn        string          `json:"recorded_on"`
	SleepMinutes      int             `json:"sleep_minutes"`
	SleepQuality      int             `json:"sleep_quality"`
	StressLevel       int             `json:"stress_level"`
	FatigueLevel      int             `json:"fatigue_level"`
	Notes             string          `json:"notes,omitempty"`
	Readiness         string          `json:"readiness"`
	AdaptationApplied bool            `json:"adaptation_applied"`
	AdaptedWorkout    *AdaptedWorkout `json:"adapted_workout,omitempty"`
}

type RecoveryStore interface {
	RecoveryByUserIDAndDate(context.Context, string, string) (*RecoveryCheckin, error)
	SaveRecovery(context.Context, string, RecoveryCheckin) (RecoveryCheckin, error)
}

type RecoveryService struct {
	store RecoveryStore
	now   func() time.Time
}

func NewRecoveryService(store RecoveryStore) *RecoveryService {
	return &RecoveryService{store: store, now: time.Now}
}

func (s *RecoveryService) Get(ctx context.Context, userID, recordedOn string) (*RecoveryCheckin, error) {
	if !validRecoveryDateFor(recordedOn, s.now()) {
		return nil, ErrInvalidRecovery
	}
	return s.store.RecoveryByUserIDAndDate(ctx, userID, recordedOn)
}

func (s *RecoveryService) Save(ctx context.Context, userID string, input RecoveryCheckin) (RecoveryCheckin, error) {
	if !validRecoveryDateFor(input.RecordedOn, s.now()) || input.SleepMinutes < 0 || input.SleepMinutes > 1440 ||
		input.SleepQuality < 1 || input.SleepQuality > 5 || input.StressLevel < 1 || input.StressLevel > 5 ||
		input.FatigueLevel < 1 || input.FatigueLevel > 5 || len(input.Notes) > 1000 {
		return RecoveryCheckin{}, ErrInvalidRecovery
	}
	input.Readiness = RecoveryReadiness(input)
	return s.store.SaveRecovery(ctx, userID, input)
}

func RecoveryReadiness(input RecoveryCheckin) string {
	poorSignals := 0
	if input.SleepMinutes < 360 || input.SleepQuality <= 2 {
		poorSignals++
	}
	if input.StressLevel >= 4 {
		poorSignals++
	}
	if input.FatigueLevel >= 4 {
		poorSignals++
	}
	if input.FatigueLevel == 5 || poorSignals >= 2 {
		return "recovery_needed"
	}
	if poorSignals == 1 {
		return "caution"
	}
	return "ready"
}

func validRecoveryDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validRecoveryDateFor(value string, now time.Time) bool {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return false
	}
	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	difference := date.Sub(today)
	return difference >= -24*time.Hour && difference <= 24*time.Hour
}
