package repository

import (
	"context"
	"errors"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/jackc/pgx/v5"
)

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
			actual_rpe = $2
		WHERE id = $1`, sessionID, input.ActualRPE); err != nil {
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
