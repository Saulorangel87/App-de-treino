package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ActivitiesByUserID(ctx context.Context, userID string) ([]planning.Activity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ws.id::text, w.id::text, w.name, w.objective, w.scheduled_on::text,
			ws.status, ws.started_at, ws.completed_at, ws.cancelled_at,
			ws.duration_minutes, ws.actual_rpe, ws.distance_km::double precision, ws.elevation_gain_m,
			ws.average_power_watts, ws.average_heart_rate,
			f.difficulty, f.pain_reported, f.fatigue_after, f.notes
		FROM workout_sessions ws
		JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
		JOIN workouts w ON w.id = ws.workout_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ap.user_id = $1 AND ws.status IN ('completed', 'cancelled')
		ORDER BY COALESCE(ws.completed_at, ws.cancelled_at) DESC, ws.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]planning.Activity, 0)
	for rows.Next() {
		var activity planning.Activity
		var startedAt, completedAt, cancelledAt *time.Time
		var duration, elevationGainM, averagePowerW, averageHeartRate *int
		var rpe *float64
		var distanceKM *float64
		var difficulty, notes *string
		var pain *bool
		var fatigue *int
		if err := rows.Scan(&activity.ID, &activity.WorkoutID, &activity.Name, &activity.Objective, &activity.ScheduledOn,
			&activity.Status, &startedAt, &completedAt, &cancelledAt, &duration, &rpe, &distanceKM, &elevationGainM, &averagePowerW, &averageHeartRate,
			&difficulty, &pain, &fatigue, &notes); err != nil {
			return nil, err
		}
		activity.StartedAt, activity.CompletedAt, activity.CancelledAt = startedAt, completedAt, cancelledAt
		activity.DurationMinutes, activity.ActualRPE = duration, rpe
		activity.DistanceKM, activity.ElevationGainM = distanceKM, elevationGainM
		activity.AveragePowerW, activity.AverageHeartRate = averagePowerW, averageHeartRate
		if difficulty != nil {
			activity.Feedback = &planning.Feedback{Difficulty: *difficulty, PainReported: *pain, FatigueAfter: *fatigue}
			if notes != nil {
				activity.Feedback.Notes = *notes
			}
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func (s *Store) StartWorkoutByUserID(ctx context.Context, userID, workoutID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var profileID, status string
	err = tx.QueryRow(ctx, `
		SELECT ap.id::text, w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND w.id = $2 AND tp.status = 'active'
		FOR UPDATE OF w`, userID, workoutID,
	).Scan(&profileID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrWorkoutMissing
	}
	if err != nil {
		return err
	}
	if status != "planned" {
		return planning.ErrInvalidTransition
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, status)
		VALUES ($1, $2, now(), 'in_progress')`, workoutID, profileID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE workouts SET status = 'in_progress' WHERE id = $1`, workoutID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CompleteWorkoutByUserID(ctx context.Context, userID, workoutID string, input planning.CompletionInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var workoutStatus string
	err = tx.QueryRow(ctx, `
		SELECT w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND w.id = $2 AND tp.status = 'active'
		FOR UPDATE OF w`, userID, workoutID,
	).Scan(&workoutStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrWorkoutMissing
	}
	if err != nil {
		return err
	}
	if workoutStatus != "in_progress" {
		return planning.ErrInvalidTransition
	}
	var sessionID string
	err = tx.QueryRow(ctx, `
		SELECT id::text FROM workout_sessions
		WHERE workout_id = $1 AND status = 'in_progress'
		FOR UPDATE`, workoutID,
	).Scan(&sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrInvalidTransition
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE workout_sessions
		SET status = 'completed',
			completed_at = now(),
			duration_minutes = GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (now() - started_at)) / 60))::integer,
			actual_rpe = $2,
			distance_km = $3,
			elevation_gain_m = $4,
			average_power_watts = $5,
			average_heart_rate = $6
		WHERE id = $1`, sessionID, input.ActualRPE, input.DistanceKM, input.ElevationGainM, input.AveragePowerW, input.AverageHeartRate); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO feedback (workout_session_id, difficulty, pain_reported, fatigue_after, notes)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))`,
		sessionID, input.Difficulty, input.PainReported, input.FatigueAfter, input.Notes,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE workouts SET status = 'completed' WHERE id = $1`, workoutID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CancelWorkoutByUserID(ctx context.Context, userID, workoutID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var workoutStatus string
	err = tx.QueryRow(ctx, `
		SELECT w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND w.id = $2 AND tp.status = 'active'
		FOR UPDATE OF w`, userID, workoutID,
	).Scan(&workoutStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrWorkoutMissing
	}
	if err != nil {
		return err
	}
	if workoutStatus != "in_progress" {
		return planning.ErrInvalidTransition
	}
	var sessionID string
	err = tx.QueryRow(ctx, `
		SELECT id::text FROM workout_sessions
		WHERE workout_id = $1 AND status = 'in_progress'
		FOR UPDATE`, workoutID,
	).Scan(&sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrInvalidTransition
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE workout_sessions
		SET status = 'cancelled', cancelled_at = now()
		WHERE id = $1`, sessionID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE workouts SET status = 'skipped' WHERE id = $1`, workoutID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
