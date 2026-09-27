package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ activityimport.Store = (*Store)(nil)

func (s *Store) ActivityExists(ctx context.Context, userID, fileHash string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM imported_activities WHERE user_id = $1 AND file_hash = $2)`,
		userID, fileHash,
	).Scan(&exists)
	return exists, err
}

func (s *Store) SaveActivity(ctx context.Context, userID, source, fileHash string, parsed activityimport.Parsed) (activityimport.Activity, error) {
	activity := activityimport.Activity{UserID: userID, Source: source, Parsed: parsed}
	var distanceKM *float64
	if parsed.DistanceKM > 0 {
		distanceKM = &parsed.DistanceKM
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO imported_activities (
			user_id, source, file_hash, started_at, moving_seconds, distance_km,
			elevation_gain_m, average_heart_rate, max_heart_rate,
			average_power_watts, normalized_power_watts, average_cadence_rpm
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id::text, imported_at`,
		userID, source, fileHash, parsed.StartedAt, parsed.MovingSeconds, distanceKM,
		parsed.ElevationGainM, parsed.AverageHeartRate, parsed.MaxHeartRate,
		parsed.AveragePowerW, parsed.NormalizedPowerW, parsed.AverageCadenceRPM,
	).Scan(&activity.ID, &activity.ImportedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return activityimport.Activity{}, activityimport.ErrDuplicate
	}
	if err != nil {
		return activityimport.Activity{}, err
	}
	return activity, nil
}

func (s *Store) ListActivities(ctx context.Context, userID string) ([]activityimport.Activity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, workout_id::text, source, started_at, moving_seconds, distance_km,
			elevation_gain_m, average_heart_rate, max_heart_rate,
			average_power_watts, normalized_power_watts, average_cadence_rpm, imported_at
		FROM imported_activities
		WHERE user_id = $1
		ORDER BY started_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activities := []activityimport.Activity{}
	for rows.Next() {
		var activity activityimport.Activity
		var workoutID *string
		var distanceKM *float64
		if err := rows.Scan(
			&activity.ID, &workoutID, &activity.Source, &activity.StartedAt, &activity.MovingSeconds, &distanceKM,
			&activity.ElevationGainM, &activity.AverageHeartRate, &activity.MaxHeartRate,
			&activity.AveragePowerW, &activity.NormalizedPowerW, &activity.AverageCadenceRPM, &activity.ImportedAt,
		); err != nil {
			return nil, err
		}
		activity.UserID = userID
		activity.WorkoutID = workoutID
		if distanceKM != nil {
			activity.DistanceKM = *distanceKM
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func (s *Store) DeleteActivity(ctx context.Context, userID, activityID string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM imported_activities WHERE id = $1 AND user_id = $2`, activityID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return activityimport.ErrNotFound
	}
	return nil
}

// WorkoutCandidatesOnDate mirrors the same athlete-scoping used by
// CurrentPlanByUserID: only workouts belonging to a plan the athlete can
// currently see (draft, active or completed).
func (s *Store) WorkoutCandidatesOnDate(ctx context.Context, userID string, date time.Time) ([]activityimport.WorkoutCandidate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT w.id::text, w.scheduled_on::text, w.name, w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND tp.status IN ('active', 'draft', 'completed') AND w.scheduled_on = $2::date
		ORDER BY w.created_at`, userID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := []activityimport.WorkoutCandidate{}
	for rows.Next() {
		var candidate activityimport.WorkoutCandidate
		if err := rows.Scan(&candidate.ID, &candidate.ScheduledOn, &candidate.Name, &candidate.Status); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}
