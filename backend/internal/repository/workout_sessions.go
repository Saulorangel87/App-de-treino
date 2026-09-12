package repository

import (
	"context"
	"encoding/json"
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
			f.completion_status, f.partial_reason, f.difficulty, f.pain_reported, f.fatigue_after, f.notes
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
		var completionStatus, partialReason, difficulty, notes *string
		var pain *bool
		var fatigue *int
		if err := rows.Scan(&activity.ID, &activity.WorkoutID, &activity.Name, &activity.Objective, &activity.ScheduledOn,
			&activity.Status, &startedAt, &completedAt, &cancelledAt, &duration, &rpe, &distanceKM, &elevationGainM, &averagePowerW, &averageHeartRate,
			&completionStatus, &partialReason, &difficulty, &pain, &fatigue, &notes); err != nil {
			return nil, err
		}
		activity.StartedAt, activity.CompletedAt, activity.CancelledAt = startedAt, completedAt, cancelledAt
		activity.DurationMinutes, activity.ActualRPE = duration, rpe
		activity.DistanceKM, activity.ElevationGainM = distanceKM, elevationGainM
		activity.AveragePowerW, activity.AverageHeartRate = averagePowerW, averageHeartRate
		if difficulty != nil {
			activity.Feedback = &planning.Feedback{CompletionStatus: *completionStatus, Difficulty: *difficulty, PainReported: *pain, FatigueAfter: *fatigue}
			if partialReason != nil {
				activity.Feedback.PartialReason = *partialReason
			}
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
	if !isStartableWorkoutStatus(status) {
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

func isStartableWorkoutStatus(status string) bool {
	return status == "planned" || status == "adapted"
}

func (s *Store) CompleteWorkoutByUserID(ctx context.Context, userID, workoutID string, input planning.CompletionInput) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var profileID string
	var plannedDurationMinutes int
	var sourceTargetRPE float64
	var status string
	err = tx.QueryRow(ctx, `
		SELECT ap.id::text, w.duration_minutes, w.target_rpe::double precision, w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND w.id = $2 AND tp.status = 'active'
		FOR UPDATE OF w`, userID, workoutID,
	).Scan(&profileID, &plannedDurationMinutes, &sourceTargetRPE, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrWorkoutMissing
	}
	if err != nil {
		return err
	}
	if status != "in_progress" {
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

	var durationMinutes int
	if err := tx.QueryRow(ctx, `
		UPDATE workout_sessions
		SET status = 'completed',
			completed_at = now(),
			duration_minutes = GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (now() - started_at)) / 60))::integer,
			actual_rpe = $2,
			distance_km = $3,
			elevation_gain_m = $4,
			average_power_watts = $5,
			average_heart_rate = $6
		WHERE id = $1
		RETURNING duration_minutes`, sessionID, input.ActualRPE, input.DistanceKM, input.ElevationGainM, input.AveragePowerW, input.AverageHeartRate).Scan(&durationMinutes); err != nil {
		return err
	}
	actualRPE := input.ActualRPE
	fatigueAfter := input.FatigueAfter
	integrityAssessedAt := time.Now()
	integrity := planning.AssessWorkoutDataIntegrity(planning.WorkoutDataIntegrityInput{
		DurationMinutes:  &durationMinutes,
		ActualRPE:        &actualRPE,
		DistanceKM:       input.DistanceKM,
		ElevationGainM:   input.ElevationGainM,
		AveragePowerW:    input.AveragePowerW,
		AverageHeartRate: input.AverageHeartRate,
		FeedbackPresent:  true,
		CompletionStatus: input.CompletionStatus,
		PartialReason:    input.PartialReason,
		Difficulty:       input.Difficulty,
		PainReported:     input.PainReported,
		FatigueAfter:     &fatigueAfter,
	}, integrityAssessedAt)
	if _, err := tx.Exec(ctx, `
		INSERT INTO feedback (workout_session_id, completion_status, partial_reason, difficulty, pain_reported, fatigue_after, notes)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, NULLIF($7, ''))`,
		sessionID, input.CompletionStatus, input.PartialReason, input.Difficulty, input.PainReported, input.FatigueAfter, input.Notes,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `SAVEPOINT rules_v2_adaptation_shadow`); err != nil {
		return err
	}
	periods, historyErr := trainingHistoryPeriodsFrom(ctx, tx, profileID)
	shadow := planning.AssessRulesV2AdaptationShadowWithIntegrity(sourceTargetRPE, input, periods, integrity, integrityAssessedAt)
	plannedVsActual := planning.AssessPlannedVsActual(planning.PlannedVsActualInput{
		PlannedDurationMinutes: plannedDurationMinutes,
		ActualDurationMinutes:  durationMinutes,
		TargetRPE:              sourceTargetRPE,
		ActualRPE:              input.ActualRPE,
		FeedbackPresent:        true,
		CompletionStatus:       input.CompletionStatus,
		PartialReason:          input.PartialReason,
		Difficulty:             input.Difficulty,
		PainReported:           input.PainReported,
		FatigueAfter:           input.FatigueAfter,
		DistanceKM:             input.DistanceKM,
		ElevationGainM:         input.ElevationGainM,
		AveragePowerW:          input.AveragePowerW,
		AverageHeartRate:       input.AverageHeartRate,
	}, integrityAssessedAt)
	shadow.PlannedVsActual = &plannedVsActual
	if historyErr != nil {
		if _, rollbackErr := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT rules_v2_adaptation_shadow`); rollbackErr != nil {
			return rollbackErr
		}
		shadow.DataIssues = append(shadow.DataIssues, "history_query_failed")
		shadow.Reasons = append(shadow.Reasons, planning.ReadinessReason{
			Code:    "history_query_failed",
			Message: "O histórico não pôde ser consultado nesta transação; a avaliação shadow ficou não avaliada.",
		})
		shadow.Status = "not_evaluated"
		shadow.CandidateResponse = "not_evaluated"
	}
	if _, err := tx.Exec(ctx, `RELEASE SAVEPOINT rules_v2_adaptation_shadow`); err != nil {
		return err
	}
	integrityJSON, err := json.Marshal(integrity)
	if err != nil {
		return err
	}
	shadowJSON, err := json.Marshal(shadow)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE workouts
		SET explanation = jsonb_set(
			jsonb_set(COALESCE(explanation, '{}'::jsonb), '{data_integrity}', $2::jsonb, true),
			'{adaptation_shadow}', $3::jsonb, true)
		WHERE id = $1`, workoutID, integrityJSON, shadowJSON); err != nil {
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

func (s *Store) MarkWorkoutMissedByUserID(ctx context.Context, userID, workoutID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var workoutStatus string
	var scheduledOnNotPast bool
	err = tx.QueryRow(ctx, `
		SELECT w.status, w.scheduled_on >= CURRENT_DATE
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND w.id = $2 AND tp.status = 'active'
		FOR UPDATE OF w`, userID, workoutID,
	).Scan(&workoutStatus, &scheduledOnNotPast)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrWorkoutMissing
	}
	if err != nil {
		return err
	}
	if workoutStatus != "planned" && workoutStatus != "adapted" {
		return planning.ErrInvalidTransition
	}
	if scheduledOnNotPast {
		return planning.ErrWorkoutNotPast
	}

	if _, err := tx.Exec(ctx, `UPDATE workouts SET status = 'skipped' WHERE id = $1`, workoutID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
