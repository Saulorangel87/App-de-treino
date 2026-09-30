package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

// recentSignalDays is how far back dated protection signals are read. The
// protection rules never look further than 14 days.
const recentSignalDays = 14

// recentSignalsByProfileID reads dated pain/fatigue records from completed
// sessions and recovery check-ins. Fatigue from a session follows the same
// data-integrity filter as the 28-day observed summary, but a reported pain
// always counts: ignoring it because a duration is missing would remove a
// protection, which is the unsafe direction.
func (s *Store) recentSignalsByProfileID(ctx context.Context, profileID string) ([]planning.RecentSignal, error) {
	signals := []planning.RecentSignal{}
	sessions, err := s.pool.Query(ctx, `
		SELECT ws.completed_at, COALESCE(f.pain_reported, false),
			CASE WHEN source_workout.explanation->'data_integrity' IS NULL
				OR source_workout.explanation->'data_integrity'->>'eligible_for_history' = 'true'
			THEN COALESCE(f.fatigue_after, 0) ELSE 0 END
		FROM workout_sessions ws
		JOIN workouts source_workout ON source_workout.id = ws.workout_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ws.athlete_profile_id = $1
		  AND ws.status = 'completed'
		  AND ws.completed_at >= now() - make_interval(days => $2)
		  AND ws.completed_at <= now()
		ORDER BY ws.completed_at DESC`, profileID, recentSignalDays)
	if err != nil {
		return nil, err
	}
	defer sessions.Close()
	for sessions.Next() {
		var completedAt time.Time
		var pain bool
		var fatigue int
		if err := sessions.Scan(&completedAt, &pain, &fatigue); err != nil {
			return nil, err
		}
		signals = append(signals, planning.RecentSignal{Date: completedAt, Source: "session", PainReported: pain, Fatigue: fatigue})
	}
	if err := sessions.Err(); err != nil {
		return nil, err
	}

	checkins, err := s.pool.Query(ctx, `
		SELECT recorded_on, COALESCE(fatigue_level, 0)
		FROM recovery_data
		WHERE athlete_profile_id = $1 AND recorded_on >= CURRENT_DATE - ($2::int - 1)
		ORDER BY recorded_on DESC`, profileID, recentSignalDays)
	if err != nil {
		return nil, err
	}
	defer checkins.Close()
	for checkins.Next() {
		var recordedOn time.Time
		var fatigue int
		if err := checkins.Scan(&recordedOn, &fatigue); err != nil {
			return nil, err
		}
		signals = append(signals, planning.RecentSignal{Date: recordedOn, Source: "checkin", Fatigue: fatigue})
	}
	if err := checkins.Err(); err != nil {
		return nil, err
	}

	reports, err := s.pool.Query(ctx, `
		SELECT reported_on FROM recovery_self_reports
		WHERE athlete_profile_id = $1 AND reported_on >= CURRENT_DATE - ($2::int - 1)
		ORDER BY reported_on DESC`, profileID, recentSignalDays)
	if err != nil {
		return nil, err
	}
	defer reports.Close()
	for reports.Next() {
		var reportedOn time.Time
		if err := reports.Scan(&reportedOn); err != nil {
			return nil, err
		}
		signals = append(signals, planning.RecentSignal{Date: reportedOn, Source: "self_report", Fatigue: 1})
	}
	return signals, reports.Err()
}

// RecordRecoverySelfReport stores today's "I am recovered" declaration.
func (s *Store) RecordRecoverySelfReport(ctx context.Context, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO recovery_self_reports (athlete_profile_id)
		SELECT id FROM athlete_profiles WHERE user_id = $1
		ON CONFLICT (athlete_profile_id, reported_on) DO NOTHING`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Either the profile does not exist or it was already reported today.
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM athlete_profiles WHERE user_id = $1)`, userID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return planning.ErrIncompleteOnboarding
		}
	}
	return nil
}

// PlannedWorkoutsForReevaluation lists the still-planned workouts of the active
// plan inside the window. Workouts in any other state are never returned.
func (s *Store) PlannedWorkoutsForReevaluation(ctx context.Context, userID, from, to string) ([]planning.Workout, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT w.id::text, w.scheduled_on::text, w.name, w.objective, w.duration_minutes,
			w.target_rpe::double precision, w.structure, w.explanation, w.status
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND tp.status = 'active' AND w.status = 'planned'
		  AND w.scheduled_on BETWEEN $2::date AND $3::date
		ORDER BY w.scheduled_on, w.created_at`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	workouts := []planning.Workout{}
	for rows.Next() {
		var workout planning.Workout
		var structure, explanation []byte
		if err := rows.Scan(&workout.ID, &workout.ScheduledOn, &workout.Name, &workout.Objective, &workout.DurationMinutes,
			&workout.TargetRPE, &structure, &explanation, &workout.Status); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(structure, &workout.Structure); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(explanation, &workout.Explanation); err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}
	return workouts, rows.Err()
}

// ApplyWorkoutRevisions rewrites planned workouts of the user's active plan in a
// single transaction. The WHERE clause repeats the guarantees of the read: only a
// workout that is still planned, in the athlete's own active plan, can change.
func (s *Store) ApplyWorkoutRevisions(ctx context.Context, userID string, revisions []planning.WorkoutRevision) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	changed := 0
	for _, revision := range revisions {
		structure, err := json.Marshal(revision.Structure)
		if err != nil {
			return 0, err
		}
		explanation, err := json.Marshal(revision.Explanation)
		if err != nil {
			return 0, err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE workouts w SET name = $3, objective = $4, duration_minutes = $5, target_rpe = $6,
				structure = $7::jsonb, explanation = $8::jsonb
			FROM training_plans tp
			JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
			WHERE w.id = $2::uuid AND w.training_plan_id = tp.id
			  AND ap.user_id = $1 AND tp.status = 'active' AND w.status = 'planned'`,
			userID, revision.WorkoutID, revision.Name, revision.Objective, revision.DurationMinutes, revision.TargetRPE, structure, explanation)
		if err != nil {
			return 0, err
		}
		changed += int(tag.RowsAffected())
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return changed, nil
}
