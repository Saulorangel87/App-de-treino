package repository

import (
	"context"
	"errors"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CurrentAssessmentByUserID(ctx context.Context, userID string) (*athlete.Assessment, error) {
	result := &athlete.Assessment{}
	err := s.pool.QueryRow(ctx, `
		SELECT ca.id::text, ca.assessment_type, ca.completed_at, ca.duration_minutes,
			ca.target_rpe::double precision, ca.actual_rpe::double precision, ca.pain_reported,
			ca.notes, ca.eligible_for_progression
		FROM cycling_assessments ca
		JOIN athlete_profiles ap ON ap.id = ca.athlete_profile_id
		WHERE ap.user_id = $1
		ORDER BY ca.completed_at DESC
		LIMIT 1`, userID,
	).Scan(&result.ID, &result.AssessmentType, &result.CompletedAt, &result.DurationMinutes,
		&result.TargetRPE, &result.ActualRPE, &result.PainReported, &result.Notes, &result.EligibleForProgression)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) SaveSubmaxAssessment(ctx context.Context, userID string, input athlete.Assessment) (athlete.Assessment, error) {
	profileID, err := s.profileID(ctx, userID)
	if err != nil {
		return athlete.Assessment{}, err
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO cycling_assessments (athlete_profile_id, duration_minutes, target_rpe, actual_rpe, pain_reported, notes, eligible_for_progression)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, assessment_type, completed_at, duration_minutes, target_rpe::double precision,
			actual_rpe::double precision, pain_reported, notes, eligible_for_progression`,
		profileID, input.DurationMinutes, input.TargetRPE, input.ActualRPE, input.PainReported, input.Notes, input.EligibleForProgression,
	).Scan(&input.ID, &input.AssessmentType, &input.CompletedAt, &input.DurationMinutes, &input.TargetRPE,
		&input.ActualRPE, &input.PainReported, &input.Notes, &input.EligibleForProgression)
	return input, err
}
