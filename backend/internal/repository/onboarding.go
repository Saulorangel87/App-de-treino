package repository

import (
	"context"
	"errors"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/jackc/pgx/v5"
)

func (s *Store) OnboardingByUserID(ctx context.Context, userID string) (athlete.Onboarding, error) {
	profileID, err := s.profileID(ctx, userID)
	if err != nil {
		return athlete.Onboarding{}, err
	}
	result := athlete.Onboarding{Limitations: []athlete.Limitation{}, Goals: []athlete.Goal{}, Availability: []athlete.Availability{}}

	limitationRows, err := s.pool.Query(ctx, `
		SELECT kind, description, is_active, professional_clearance_recommended
		FROM injuries_or_limitations WHERE athlete_profile_id = $1 AND is_active = true ORDER BY created_at`, profileID)
	if err != nil {
		return athlete.Onboarding{}, err
	}
	for limitationRows.Next() {
		var item athlete.Limitation
		if err := limitationRows.Scan(&item.Kind, &item.Description, &item.IsActive, &item.ProfessionalClearanceRecommended); err != nil {
			limitationRows.Close()
			return athlete.Onboarding{}, err
		}
		result.Limitations = append(result.Limitations, item)
	}
	if err := limitationRows.Err(); err != nil {
		limitationRows.Close()
		return athlete.Onboarding{}, err
	}
	limitationRows.Close()

	goalRows, err := s.pool.Query(ctx, `
		SELECT goal_type, priority, target_date::text, COALESCE(details->>'notes', '')
		FROM goals WHERE athlete_profile_id = $1 ORDER BY priority`, profileID)
	if err != nil {
		return athlete.Onboarding{}, err
	}
	for goalRows.Next() {
		var item athlete.Goal
		if err := goalRows.Scan(&item.GoalType, &item.Priority, &item.TargetDate, &item.Details); err != nil {
			goalRows.Close()
			return athlete.Onboarding{}, err
		}
		result.Goals = append(result.Goals, item)
	}
	if err := goalRows.Err(); err != nil {
		goalRows.Close()
		return athlete.Onboarding{}, err
	}
	goalRows.Close()

	availabilityRows, err := s.pool.Query(ctx, `
		SELECT weekday, available_minutes, to_char(preferred_time, 'HH24:MI'), location
		FROM availability WHERE athlete_profile_id = $1 ORDER BY weekday`, profileID)
	if err != nil {
		return athlete.Onboarding{}, err
	}
	defer availabilityRows.Close()
	for availabilityRows.Next() {
		var item athlete.Availability
		if err := availabilityRows.Scan(&item.Weekday, &item.AvailableMinutes, &item.PreferredTime, &item.Location); err != nil {
			return athlete.Onboarding{}, err
		}
		result.Availability = append(result.Availability, item)
	}
	return result, availabilityRows.Err()
}

func (s *Store) ReplaceLimitations(ctx context.Context, userID string, limitations []athlete.Limitation) ([]athlete.Limitation, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	profileID, err := profileIDFrom(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM injuries_or_limitations WHERE athlete_profile_id = $1`, profileID); err != nil {
		return nil, err
	}
	for _, item := range limitations {
		if _, err := tx.Exec(ctx, `
			INSERT INTO injuries_or_limitations (athlete_profile_id, kind, description, is_active, professional_clearance_recommended)
			VALUES ($1, $2, $3, $4, $5)`, profileID, item.Kind, item.Description, item.IsActive, item.ProfessionalClearanceRecommended); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return limitations, nil
}

func (s *Store) ReplaceGoals(ctx context.Context, userID string, goals []athlete.Goal) ([]athlete.Goal, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	profileID, err := profileIDFrom(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM goals WHERE athlete_profile_id = $1`, profileID); err != nil {
		return nil, err
	}
	for _, item := range goals {
		if _, err := tx.Exec(ctx, `
			INSERT INTO goals (athlete_profile_id, goal_type, priority, target_date, details)
			VALUES ($1, $2, $3, $4::date, jsonb_build_object('notes', $5::text))`, profileID, item.GoalType, item.Priority, item.TargetDate, item.Details); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return goals, nil
}

func (s *Store) ReplaceAvailability(ctx context.Context, userID string, availability []athlete.Availability) ([]athlete.Availability, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	profileID, err := profileIDFrom(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM availability WHERE athlete_profile_id = $1`, profileID); err != nil {
		return nil, err
	}
	for _, item := range availability {
		if _, err := tx.Exec(ctx, `
			INSERT INTO availability (athlete_profile_id, weekday, available_minutes, preferred_time, location)
			VALUES ($1, $2, $3, $4::time, $5)`, profileID, item.Weekday, item.AvailableMinutes, item.PreferredTime, item.Location); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return availability, nil
}

func (s *Store) profileID(ctx context.Context, userID string) (string, error) {
	var profileID string
	err := s.pool.QueryRow(ctx, `SELECT id::text FROM athlete_profiles WHERE user_id = $1`, userID).Scan(&profileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", athlete.ErrProfileMissing
	}
	return profileID, err
}

func profileIDFrom(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var profileID string
	err := tx.QueryRow(ctx, `SELECT id::text FROM athlete_profiles WHERE user_id = $1`, userID).Scan(&profileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", athlete.ErrProfileMissing
	}
	return profileID, err
}
