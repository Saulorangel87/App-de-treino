package planning

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"
)

// Reevaluate rebuilds the still-planned workouts of the current week and of the
// next one with today's protection assessment, so that a recovery signal (or its
// disappearance) reaches sessions that were generated earlier. Completed,
// started, cancelled and check-in-adapted workouts are never touched. Neither are
// the next one or two sessions that the database already adapted after a
// post-workout feedback (explanation.adaptation): that short-term reduction can
// be stronger than a graduated level and must not be overwritten. Workouts
// generated before their prescription inputs were recorded are skipped too. It is
// a no-op unless protection levels are enabled, and returns how many workouts changed.
func (s *Service) Reevaluate(ctx context.Context, userID string) (int, error) {
	if !s.protectionLevels {
		return 0, nil
	}
	input, err := s.store.PlanningContextByUserID(ctx, userID)
	if errors.Is(err, ErrIncompleteOnboarding) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	now := s.now()
	input = s.withProtection(input)
	from, to := reevaluationWindow(now)
	workouts, err := s.store.PlannedWorkoutsForReevaluation(ctx, userID, from, to)
	if err != nil {
		return 0, err
	}
	var revisions []WorkoutRevision
	for _, current := range workouts {
		if current.Status != "planned" || current.Explanation["adaptation"] != nil {
			continue
		}
		rebuilt, ok := ReprescribeWorkout(input, current.ScheduledOn, current.Explanation, now)
		if !ok {
			continue
		}
		explanation := mergeExplanation(current.Explanation, rebuilt.Explanation)
		if current.Name == rebuilt.Name && current.DurationMinutes == rebuilt.DurationMinutes &&
			current.TargetRPE == rebuilt.TargetRPE && sameJSON(current.Explanation["protection"], explanation["protection"]) {
			continue
		}
		revisions = append(revisions, WorkoutRevision{
			WorkoutID: current.ID, Name: rebuilt.Name, Objective: rebuilt.Objective,
			DurationMinutes: rebuilt.DurationMinutes, TargetRPE: rebuilt.TargetRPE,
			Structure: rebuilt.Structure, Explanation: explanation,
		})
	}
	if len(revisions) == 0 {
		return 0, nil
	}
	return s.store.ApplyWorkoutRevisions(ctx, userID, revisions)
}

// reevaluateBestEffort never fails the request that triggered it: the signal was
// already saved, and the next trigger re-evaluates from the stored records.
func (s *Service) reevaluateBestEffort(ctx context.Context, userID string) {
	if _, err := s.Reevaluate(ctx, userID); err != nil {
		slog.Default().Warn("protection re-evaluation failed", "error", err)
	}
}

// ReevaluateBestEffort is the entry point for flows outside this package, such as
// the daily check-in, that change the recovery signals.
func (s *Service) ReevaluateBestEffort(ctx context.Context, userID string) {
	s.reevaluateBestEffort(ctx, userID)
}

// reevaluationWindow covers today through the Sunday of next week.
func reevaluationWindow(now time.Time) (string, string) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sinceMonday := (int(today.Weekday()) + 6) % 7
	end := today.AddDate(0, 0, 13-sinceMonday)
	return today.Format("2006-01-02"), end.Format("2006-01-02")
}

// mergeExplanation keeps keys that only the stored explanation has and lets the
// rebuilt prescription overwrite the rest. The protection block is dropped when
// the rebuilt workout has none.
func mergeExplanation(stored, rebuilt map[string]any) map[string]any {
	merged := make(map[string]any, len(stored)+len(rebuilt))
	for key, value := range stored {
		merged[key] = value
	}
	for key, value := range rebuilt {
		merged[key] = value
	}
	if _, ok := rebuilt["protection"]; !ok {
		delete(merged, "protection")
	}
	return merged
}

func sameJSON(a, b any) bool {
	encodedA, errA := json.Marshal(a)
	encodedB, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	var decodedA, decodedB any
	if json.Unmarshal(encodedA, &decodedA) != nil || json.Unmarshal(encodedB, &decodedB) != nil {
		return false
	}
	return reflect.DeepEqual(decodedA, decodedB)
}
