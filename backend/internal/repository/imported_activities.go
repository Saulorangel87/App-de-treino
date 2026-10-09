package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/activityimport"
	"github.com/jackc/pgx/v5"
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
			average_power_watts, normalized_power_watts, average_cadence_rpm,
			best_20min_power_watts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id::text, imported_at`,
		userID, source, fileHash, parsed.StartedAt, parsed.MovingSeconds, distanceKM,
		parsed.ElevationGainM, parsed.AverageHeartRate, parsed.MaxHeartRate,
		parsed.AveragePowerW, parsed.NormalizedPowerW, parsed.AverageCadenceRPM,
		parsed.Best20MinPowerW,
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

// importedActivityColumns e scanImportedActivity mantêm a lista e a consulta
// individual iguais; o LEFT JOIN traz o treino vinculado só para exibição.
const importedActivityColumns = `
	ia.id::text, ia.workout_id::text, w.name, w.scheduled_on::text, ia.source, ia.started_at,
	ia.moving_seconds, ia.distance_km, ia.elevation_gain_m, ia.average_heart_rate,
	ia.max_heart_rate, ia.average_power_watts, ia.normalized_power_watts,
	ia.average_cadence_rpm, ia.best_20min_power_watts, ia.imported_at
	FROM imported_activities ia
	LEFT JOIN workouts w ON w.id = ia.workout_id`

type rowScanner interface{ Scan(dest ...any) error }

func scanImportedActivity(row rowScanner, userID string) (activityimport.Activity, error) {
	var activity activityimport.Activity
	var distanceKM *float64
	if err := row.Scan(
		&activity.ID, &activity.WorkoutID, &activity.WorkoutName, &activity.WorkoutScheduledOn,
		&activity.Source, &activity.StartedAt, &activity.MovingSeconds, &distanceKM,
		&activity.ElevationGainM, &activity.AverageHeartRate, &activity.MaxHeartRate,
		&activity.AveragePowerW, &activity.NormalizedPowerW, &activity.AverageCadenceRPM, &activity.Best20MinPowerW, &activity.ImportedAt,
	); err != nil {
		return activityimport.Activity{}, err
	}
	activity.UserID = userID
	if distanceKM != nil {
		activity.DistanceKM = *distanceKM
	}
	return activity, nil
}

func (s *Store) ListActivities(ctx context.Context, userID string) ([]activityimport.Activity, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+importedActivityColumns+`
		WHERE ia.user_id = $1
		ORDER BY ia.started_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	activities := []activityimport.Activity{}
	for rows.Next() {
		activity, err := scanImportedActivity(rows, userID)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, rows.Err()
}

func (s *Store) GetActivity(ctx context.Context, userID, activityID string) (activityimport.Activity, error) {
	activity, err := scanImportedActivity(s.pool.QueryRow(ctx, `SELECT `+importedActivityColumns+`
		WHERE ia.id = $1 AND ia.user_id = $2`, activityID, userID), userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return activityimport.Activity{}, activityimport.ErrNotFound
	}
	return activity, err
}

// LinkActivity só aceita um treino que pertença a um plano do próprio atleta,
// com o mesmo escopo usado para sugerir candidatos. A atividade é filtrada
// pelo dono na mesma instrução, então um id de outro usuário nunca é alterado.
func (s *Store) LinkActivity(ctx context.Context, userID, activityID string, workoutID *string) error {
	if workoutID == nil {
		tag, err := s.pool.Exec(ctx,
			`UPDATE imported_activities SET workout_id = NULL WHERE id = $1 AND user_id = $2`,
			activityID, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return activityimport.ErrNotFound
		}
		return nil
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE imported_activities SET workout_id = $3::uuid
		WHERE id = $1 AND user_id = $2
		  AND EXISTS (
			SELECT 1
			FROM workouts w
			JOIN training_plans tp ON tp.id = w.training_plan_id
			JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
			WHERE w.id = $3::uuid AND ap.user_id = $2 AND tp.status IN ('active', 'draft', 'completed')
		  )`, activityID, userID, *workoutID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return activityimport.ErrWorkoutNotFound
	}
	return nil
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
