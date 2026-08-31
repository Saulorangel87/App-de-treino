package repository

import (
	"context"
	"errors"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/jackc/pgx/v5"
)

func (s *Store) RecoveryByUserIDAndDate(ctx context.Context, userID, recordedOn string) (*athlete.RecoveryCheckin, error) {
	result := &athlete.RecoveryCheckin{}
	err := s.pool.QueryRow(ctx, `
		SELECT rd.id::text, rd.recorded_on::text, COALESCE(rd.sleep_minutes, 0), COALESCE(rd.sleep_quality, 3),
			COALESCE(rd.stress_level, 3), COALESCE(rd.fatigue_level, 3), COALESCE(rd.notes, '')
		FROM recovery_data rd
		JOIN athlete_profiles ap ON ap.id = rd.athlete_profile_id
		WHERE ap.user_id = $1 AND rd.recorded_on = $2::date`, userID, recordedOn,
	).Scan(&result.ID, &result.RecordedOn, &result.SleepMinutes, &result.SleepQuality,
		&result.StressLevel, &result.FatigueLevel, &result.Notes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result.Readiness = athlete.RecoveryReadiness(*result)
	if err := s.loadRecoveryAdaptation(ctx, userID, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) SaveRecovery(ctx context.Context, userID string, input athlete.RecoveryCheckin) (athlete.RecoveryCheckin, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return athlete.RecoveryCheckin{}, err
	}
	defer tx.Rollback(ctx)

	var profileID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM athlete_profiles WHERE user_id = $1`, userID).Scan(&profileID); errors.Is(err, pgx.ErrNoRows) {
		return athlete.RecoveryCheckin{}, athlete.ErrProfileMissing
	} else if err != nil {
		return athlete.RecoveryCheckin{}, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO recovery_data (athlete_profile_id, recorded_on, sleep_minutes, sleep_quality, stress_level, fatigue_level, notes)
		VALUES ($1, $2::date, $3, $4, $5, $6, $7)
		ON CONFLICT (athlete_profile_id, recorded_on) DO UPDATE SET
			sleep_minutes = EXCLUDED.sleep_minutes, sleep_quality = EXCLUDED.sleep_quality,
			stress_level = EXCLUDED.stress_level, fatigue_level = EXCLUDED.fatigue_level, notes = EXCLUDED.notes
		RETURNING id::text, recorded_on::text, sleep_minutes, sleep_quality, stress_level, fatigue_level, COALESCE(notes, '')`,
		profileID, input.RecordedOn, input.SleepMinutes, input.SleepQuality, input.StressLevel, input.FatigueLevel, input.Notes,
	).Scan(&input.ID, &input.RecordedOn, &input.SleepMinutes, &input.SleepQuality, &input.StressLevel, &input.FatigueLevel, &input.Notes)
	if err != nil {
		return athlete.RecoveryCheckin{}, err
	}

	if input.Readiness != "ready" {
		factor, rpeCap := 0.90, 5.0
		if input.Readiness == "recovery_needed" {
			factor, rpeCap = 0.80, 4.0
		}
		adapted := &athlete.AdaptedWorkout{}
		err = tx.QueryRow(ctx, `
			WITH candidate AS (
				SELECT w.id
				FROM workouts w
				JOIN training_plans tp ON tp.id = w.training_plan_id
				WHERE tp.athlete_profile_id = $1 AND tp.status = 'active'
					AND w.status IN ('planned', 'adapted') AND w.scheduled_on >= $2::date
					AND NOT EXISTS (
						SELECT 1 FROM workouts prior
						JOIN training_plans prior_plan ON prior_plan.id = prior.training_plan_id
						WHERE prior_plan.athlete_profile_id = $1
							AND COALESCE(prior.explanation #> '{pre_session_recovery,applied_dates}', '[]'::jsonb) ? $2
					)
				ORDER BY w.scheduled_on, w.created_at
				LIMIT 1
				FOR UPDATE
			)
			UPDATE workouts w SET
				duration_minutes = GREATEST(20, ROUND(w.duration_minutes * $3)::integer),
				target_rpe = LEAST(w.target_rpe, $4),
				status = 'adapted',
				explanation = jsonb_set(
					jsonb_set(w.explanation, '{pre_session_recovery}', jsonb_build_object(
						'recorded_on', $2, 'readiness', $5,
						'applied_dates', COALESCE(w.explanation #> '{pre_session_recovery,applied_dates}', '[]'::jsonb) || jsonb_build_array($2::text),
						'summary', CASE WHEN $5 = 'recovery_needed'
							THEN 'Carga reduzida porque o check-in indicou necessidade de recuperação.'
							ELSE 'Carga ajustada com cautela por um sinal de recuperação abaixo do habitual.' END
					), true),
					'{evidence_keys}', CASE
						WHEN COALESCE(w.explanation->'evidence_keys', '[]'::jsonb) ? 'bourdon-2017'
							THEN COALESCE(w.explanation->'evidence_keys', '[]'::jsonb)
						ELSE COALESCE(w.explanation->'evidence_keys', '[]'::jsonb) || '["bourdon-2017"]'::jsonb
					END, true)
			FROM candidate c WHERE w.id = c.id
			RETURNING w.id::text, w.scheduled_on::text, w.name, w.duration_minutes, w.target_rpe::double precision`,
			profileID, input.RecordedOn, factor, rpeCap, input.Readiness,
		).Scan(&adapted.ID, &adapted.ScheduledOn, &adapted.Name, &adapted.DurationMinutes, &adapted.TargetRPE)
		if err == nil {
			input.AdaptationApplied = true
			input.AdaptedWorkout = adapted
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return athlete.RecoveryCheckin{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return athlete.RecoveryCheckin{}, err
	}
	return input, nil
}

func (s *Store) loadRecoveryAdaptation(ctx context.Context, userID string, result *athlete.RecoveryCheckin) error {
	adapted := &athlete.AdaptedWorkout{}
	err := s.pool.QueryRow(ctx, `
		SELECT w.id::text, w.scheduled_on::text, w.name, w.duration_minutes, w.target_rpe::double precision
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1
			AND COALESCE(w.explanation #> '{pre_session_recovery,applied_dates}', '[]'::jsonb) ? $2
		ORDER BY w.scheduled_on LIMIT 1`, userID, result.RecordedOn,
	).Scan(&adapted.ID, &adapted.ScheduledOn, &adapted.Name, &adapted.DurationMinutes, &adapted.TargetRPE)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	result.AdaptationApplied = true
	result.AdaptedWorkout = adapted
	return nil
}
