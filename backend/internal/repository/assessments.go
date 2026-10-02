package repository

import (
	"context"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
)

const assessmentColumns = `ca.id::text, ca.assessment_type, ca.completed_at, ca.duration_minutes,
	ca.target_rpe::double precision, ca.actual_rpe::double precision, ca.pain_reported,
	ca.notes, ca.eligible_for_progression,
	ca.average_heart_rate, ca.average_power_w, ca.distance_km::double precision,
	ca.heart_rate_first_half, ca.heart_rate_second_half`

type assessmentScanner interface{ Scan(dest ...any) error }

func scanAssessment(row assessmentScanner, result *athlete.Assessment) error {
	return row.Scan(&result.ID, &result.AssessmentType, &result.CompletedAt, &result.DurationMinutes,
		&result.TargetRPE, &result.ActualRPE, &result.PainReported, &result.Notes, &result.EligibleForProgression,
		&result.AverageHeartRate, &result.AveragePowerW, &result.DistanceKM,
		&result.HeartRateFirstHalf, &result.HeartRateSecondHalf)
}

func (s *Store) CurrentAssessmentByUserID(ctx context.Context, userID string) (*athlete.Assessment, error) {
	history, err := s.AssessmentHistoryByUserID(ctx, userID, 1)
	if err != nil || len(history) == 0 {
		return nil, err
	}
	return &history[0], nil
}

// AssessmentHistoryByUserID returns the latest assessments, newest first.
func (s *Store) AssessmentHistoryByUserID(ctx context.Context, userID string, limit int) ([]athlete.Assessment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+assessmentColumns+`
		FROM cycling_assessments ca
		JOIN athlete_profiles ap ON ap.id = ca.athlete_profile_id
		WHERE ap.user_id = $1
		ORDER BY ca.completed_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := []athlete.Assessment{}
	for rows.Next() {
		var item athlete.Assessment
		if err := scanAssessment(rows, &item); err != nil {
			return nil, err
		}
		history = append(history, item)
	}
	return history, rows.Err()
}

func (s *Store) SaveSubmaxAssessment(ctx context.Context, userID string, input athlete.Assessment) (athlete.Assessment, error) {
	profileID, err := s.profileID(ctx, userID)
	if err != nil {
		return athlete.Assessment{}, err
	}
	err = scanAssessment(s.pool.QueryRow(ctx, `
		WITH ca AS (
			INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, target_rpe, actual_rpe, pain_reported, notes,
				eligible_for_progression, average_heart_rate, average_power_w, distance_km, heart_rate_first_half, heart_rate_second_half)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING *
		)
		SELECT `+assessmentColumns+` FROM ca`,
		profileID, input.DurationMinutes, input.TargetRPE, input.ActualRPE, input.PainReported, input.Notes, input.EligibleForProgression,
		input.AverageHeartRate, input.AveragePowerW, input.DistanceKM, input.HeartRateFirstHalf, input.HeartRateSecondHalf,
	), &input)
	return input, err
}
